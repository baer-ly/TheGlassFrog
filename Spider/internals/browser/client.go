package browser

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/mxschmitt/playwright-go"
)

type Engine struct {
	PW      *playwright.Playwright
	Browser playwright.Browser
	Cookies []playwright.OptionalCookie
}

func InitEngine(headless bool) (*Engine, error) {
	pw, err := playwright.Run()
	if err != nil {
		return nil, fmt.Errorf("could not launch playwright: %v", err)
	}
	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: new(headless),
		Args: []string{
			"--disable-blink-features=AutomationControlled",
			"--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36",
		},
	})
	if err != nil {
		err := pw.Stop()
		return nil, fmt.Errorf("could not launch Chromium: %v", err)
	}
	engine := &Engine{PW: pw, Browser: browser}

	if cookieData, err := os.ReadFile("cookies.json"); err == nil {
		var parseCookies []playwright.OptionalCookie
		if err := json.Unmarshal(cookieData, &parseCookies); err == nil {
			engine.Cookies = parseCookies
		}
	}
	return engine, nil
}

func (engine *Engine) CreateIsolatedContext(proxyServer string) (playwright.BrowserContext, error) {
	opts := playwright.BrowserNewContextOptions{}
	if proxyServer != "" {
		opts.Proxy = &playwright.Proxy{Server: proxyServer}
	}
	context, err := engine.Browser.NewContext(opts)
	if err != nil {
		return nil, fmt.Errorf("could not create context: %v", err)
	}

	if len(engine.Cookies) > 0 {
		if err := context.AddCookies(engine.Cookies); err != nil {
			err := context.Close()
			return nil, fmt.Errorf("could not add cookies: %v", err)
		}
	}
	return context, nil
}

func (engine *Engine) Close() {
	if engine.PW != nil {
		err := engine.PW.Stop()
		if err != nil {
			return
		}
	}
	if engine.Browser != nil {
		err := engine.Browser.Close()
		if err != nil {
			return
		}
	}
}
