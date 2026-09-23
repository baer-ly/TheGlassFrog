package scraper

import (
	"TheGlassFrog/Spider/internals/browser"
	config "TheGlassFrog/Spider/web"

	"context"
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/mxschmitt/playwright-go"
)

type Job struct {
	Platform config.Platform
	Username string
}

type ScanMatch struct {
	PlatformName string
	ProfileURL   string
	Username     string
}

func ScoutWorker(
	parentContext context.Context,
	engine *browser.Engine,
	proxyString string,
	jobChan <-chan Job,
	matchChan chan<- ScanMatch,
	timeoutDuration time.Duration,
) {
	for job := range jobChan {
		select {
		case <-parentContext.Done():
			return
		case <-time.After(1000*time.Millisecond + time.Duration(rand.Int64N(2000))*time.Millisecond):
		}

		_, cancel := context.WithTimeout(parentContext, timeoutDuration)

		browserContext, err := engine.CreateIsolatedContext(proxyString)
		if err != nil {
			cancel()
			continue
		}

		page, err := browserContext.NewPage()
		if err != nil {
			browserContext.Close()
			cancel()
			continue
		}

		msTimeout := float64(timeoutDuration.Milliseconds())
		page.SetDefaultTimeout(msTimeout)
		targetURL := fmt.Sprintf(job.Platform.URLTemplate, job.Username)

		err = func() error {

			if _, err := page.Goto(targetURL, playwright.PageGotoOptions{
				WaitUntil: playwright.WaitUntilStateDomcontentloaded,
				Timeout:   new(msTimeout),
			}); err != nil {
				return err
			}


			if strings.Contains(page.URL(), "/accounts/login/") || strings.Contains(page.URL(), "/login") {
				return fmt.Errorf("redirected to login wall deflection")
			}


			content, err := page.Content()
			if err != nil {
				return err
			}
			if strings.Contains(strings.ToLower(content), strings.ToLower(job.Platform.FailedText)) {
				return fmt.Errorf("user not found signature")
			}


			_, _ = page.Screenshot(playwright.PageScreenshotOptions{
				Path:     new("Spider/storage/evidence/front_page_view.png"),
				FullPage: new(true),
			})

			count, err := page.Locator(job.Platform.SuccessSelector).Count()

			if err != nil {
				return err
			}

			if count == 0 {
				return fmt.Errorf("user footprint not attached to DOM")
			}

			return nil
		}()

		if err == nil {
			matchChan <- ScanMatch{
				PlatformName: job.Platform.Name,
				ProfileURL:   targetURL,
				Username:     job.Username,
			}
		}

		page.Close()
		browserContext.Close()
		cancel()
	}
}
