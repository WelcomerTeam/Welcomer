package service

import (
	"context"
	"sync/atomic"
	"time"

	"net/url"
	"sync"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type URLPool struct {
	mu   sync.Mutex
	cond *sync.Cond
	urls []string
}

func NewURLPool(urls []string) *URLPool {
	pool := &URLPool{urls: urls}
	pool.cond = sync.NewCond(&pool.mu)
	return pool
}

func (p *URLPool) Get(ctx context.Context) (string, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for len(p.urls) == 0 {
		p.cond.Wait()
	}

	url := p.urls[0]
	p.urls = p.urls[1:]

	return url, nil
}

func (p *URLPool) Return(url string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.urls = append(p.urls, url)
	p.cond.Signal()
}

func waitNetwork(timeout time.Duration) chromedp.ActionFunc {
	return chromedp.ActionFunc(func(ctx context.Context) error {
		requestsWaiting := atomic.Int32{}
		requestsUpdate := make(chan struct{}, 1)

		timeout := time.NewTicker(timeout)
		defer timeout.Stop()

		chromedp.ListenTarget(ctx, func(ev any) {
			switch ev.(type) {
			case *network.EventLoadingFailed:
				requestsWaiting.Add(-1)
				requestsUpdate <- struct{}{}
			case *network.EventLoadingFinished:
				requestsWaiting.Add(-1)
				requestsUpdate <- struct{}{}
			case *network.EventRequestWillBeSent:
				requestsWaiting.Add(1)
				requestsUpdate <- struct{}{}
			}
		})

		time.Sleep(time.Millisecond * 50) // allow some time for requests to start

		select {
		case requestsUpdate <- struct{}{}:
		default:
		}

		for {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-requestsUpdate:
				if requestsWaiting.Load() == 0 {
					return nil
				}
			case <-timeout.C:
				welcomer.Logger.Warn().Msg("network timeout")

				return nil
			}
		}
	})
}

func ScreenshotFromHTML(ctx context.Context, pool *URLPool, htmlString string) ([]byte, error) {
	cdpurl, err := pool.Get(ctx)
	if err != nil {
		return nil, err
	}

	defer pool.Return(cdpurl)

	ctx, _ = chromedp.NewRemoteAllocator(ctx, cdpurl)
	ctx, _ = chromedp.NewContext(ctx)

	var buf []byte

	start := time.Now()
	err = chromedp.Run(ctx,
		chromedp.Tasks{
			chromedp.ActionFunc(func(ctx context.Context) error {
				return emulation.SetDefaultBackgroundColorOverride().
					WithColor(&cdp.RGBA{0, 0, 0, 0}).
					Do(ctx)
			}),

			chromedp.Navigate("data:text/html;charset=utf-8," + url.PathEscape(htmlString)),

			waitNetwork(time.Second * 2),
			chromedp.Evaluate(`document.fonts.ready`, nil),

			chromedp.WaitReady("#canvas", chromedp.ByID),
			chromedp.Screenshot("div#canvas", &buf, chromedp.NodeReady),
		},
	)

	println("Screenshot took", time.Since(start).String(), cdpurl)

	return buf, err
}
