package service

import (
	"context"
	"net/url"
	"sync"
	"sync/atomic"
	"time"

	"github.com/WelcomerTeam/Welcomer/welcomer-core"
	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
)

type Pool interface {
	Get() (string, error)
	Return(string)
}

type HardcodedPool struct {
	url string
}

func NewHardcodedPool(url string) Pool {
	pool := &HardcodedPool{url}

	return pool
}

func (h *HardcodedPool) Get() (string, error) {
	return h.url, nil
}

func (h *HardcodedPool) Return(url string) {
	return
}

type URLPool struct {
	mu   *sync.Mutex
	cond *sync.Cond
	urls []string
}

func NewURLPool(urls []string) Pool {
	mu := sync.Mutex{}

	pool := &URLPool{
		urls: urls,
		mu:   &mu,
		cond: sync.NewCond(&mu),
	}

	return pool
}

func (p *URLPool) Get() (string, error) {
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

func (is *ImageService) ScreenshotFromHTML(ctx context.Context, htmlString string) ([]byte, time.Duration, error) {
	cdpurl, err := is.URLPool.Get()
	if err != nil {
		return nil, time.Duration(0), err
	}

	defer is.URLPool.Return(cdpurl)

	ctx, _ = chromedp.NewRemoteAllocator(ctx, cdpurl)
	ctx, _ = chromedp.NewContext(ctx)

	var buf []byte

	start := time.Now()

	if is.Options.Debug {
		println(htmlString)
	}

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

	return buf, time.Since(start), err
}
