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
	"github.com/chromedp/chromedp/remote"
)

type Pool interface {
	Get() (string, error)
	Return(url string)
}

type HardcodedPool struct {
	url string
}

func NewHardcodedPool(url string) *HardcodedPool {
	pool := &HardcodedPool{url}

	return pool
}

func (h *HardcodedPool) Get() (string, error) {
	return h.url, nil
}

func (h *HardcodedPool) Return(_ string) {
}

type URLPool struct {
	mu   *sync.Mutex
	cond *sync.Cond
	urls []string
}

func NewURLPool(urls []string) *URLPool {
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

func waitNetwork(timeout time.Duration) chromedp.Action[chromedp.Void] {
	return chromedp.Func(func(ctx context.Context, _ *chromedp.Target) error {
		requestsWaiting := atomic.Int32{}
		requestsUpdate := make(chan struct{}, 1)
		eventsCtx, cancelEvents := context.WithCancel(ctx)
		defer cancelEvents()

		requestEvents := chromedp.Events(eventsCtx, network.RequestWillBeSent)
		finishedEvents := chromedp.Events(eventsCtx, network.LoadingFinished)
		failedEvents := chromedp.Events(eventsCtx, network.LoadingFailed)

		notify := func(delta int32) {
			requestsWaiting.Add(delta)
			select {
			case requestsUpdate <- struct{}{}:
			default:
			}
		}

		go func() {
			for _, err := range requestEvents {
				if err != nil {
					return
				}
				notify(1)
			}
		}()
		go func() {
			for _, err := range finishedEvents {
				if err != nil {
					return
				}
				notify(-1)
			}
		}()
		go func() {
			for _, err := range failedEvents {
				if err != nil {
					return
				}
				notify(-1)
			}
		}()

		if _, err := chromedp.Call(ctx, network.Enable, network.EnableParams{}); err != nil {
			return err
		}

		timeoutTimer := time.NewTimer(timeout)
		defer timeoutTimer.Stop()

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
			case <-timeoutTimer.C:
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

	allocatorCtx, cancelAllocator := remote.NewAllocator(ctx, cdpurl)
	defer cancelAllocator()
	ctx, cancelBrowser := chromedp.NewContext(allocatorCtx)
	defer cancelBrowser()

	start := time.Now()

	if is.Options.Debug {
		println(htmlString)
	}

	err = chromedp.Do(ctx,
		chromedp.Func(func(ctx context.Context, _ *chromedp.Target) error {
			_, err := chromedp.Call(ctx, emulation.SetDefaultBackgroundColorOverride,
				emulation.SetDefaultBackgroundColorOverrideParams{Color: &cdp.RGBA{}})
			return err
		}),
		chromedp.Func(func(ctx context.Context, _ *chromedp.Target) error {
			_, err := chromedp.Call(ctx, emulation.SetDeviceMetricsOverride,
				emulation.SetDeviceMetricsOverrideParams{
					DeviceScaleFactor: 3,
				})
			return err
		}),
		chromedp.Navigate("data:text/html;charset=utf-8,"+url.PathEscape(htmlString)),
		waitNetwork(2*time.Second),
		chromedp.Evaluate[chromedp.Void](`document.fonts.ready`, chromedp.EvalAwaitPromise),
		chromedp.WaitReady(chromedp.CSS("#canvas")),
	)
	if err != nil {
		return nil, time.Since(start), err
	}

	buf, err := chromedp.Run(ctx, chromedp.Screenshot(chromedp.CSS("div#canvas")))

	return buf, time.Since(start), err
}
