package main

import (
	"TheGlassFrog/Spider/internals/browser"
	"TheGlassFrog/Spider/internals/browser/report"
	"TheGlassFrog/Spider/internals/scraper"
	config "TheGlassFrog/Spider/web"
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/mxschmitt/playwright-go"
)

func main() {
	usernameFlag := flag.String("u", "", "The target username indicator to check (Required)")
	platformsFlag := flag.String("p", "", "Comma-separated selective search filters")
	scoutWorkersFlag := flag.Int("w", 3, "Total concurrent validation scout pipelines")
	flag.Parse()

	if *usernameFlag == "" {
		fmt.Println("Usage: go run cmd/osint/main.go -u <username> [-p <platforms>] [-w <workers>]")
		os.Exit(1)
	}

	// 1. Initialize Visual Live Logging Framework
	logger := report.NewLiveLogger()
	logger.LogPrint(report.ColorBlue, "*", "Initializing OSINT Engine configurations...")

	cfg, err := config.LoadConfig("Spider/web/profiles.json")
	if err != nil {
		log.Fatalf("System error loading operational file: %v", err)
	}

	engine, err := browser.InitEngine(true)
	if err != nil {
		log.Fatalf("Fatal system browser driver crash: %v", err)
	}
	defer engine.Close()

	rotator, _ := browser.NewProxyRotator("proxies.txt")
	ctx := context.Background()

	var targetedPlatforms []config.Platform
	if *platformsFlag != "" {
		filterMap := make(map[string]bool)
		for input := range strings.SplitSeq(*platformsFlag, ",") {
			filterMap[strings.ToLower(strings.TrimSpace(input))] = true
		}
		for _, p := range cfg.Platforms {
			if filterMap[strings.ToLower(p.Name)] {
				targetedPlatforms = append(targetedPlatforms, p)
			}
		}
	} else {
		targetedPlatforms = cfg.Platforms
	}

	jobs := make(chan scraper.Job, len(targetedPlatforms))
	matches := make(chan scraper.ScanMatch, len(targetedPlatforms))

	var reportDataMutex sync.Mutex
	var finalReportMatches []report.TargetMatch

	// 2. Start the Live Spinner UI block
	logger.StartSpinner(fmt.Sprintf("Scanning infrastructure maps for username target: @%s...", *usernameFlag))

	var scoutWg sync.WaitGroup
	numScouts := min(*scoutWorkersFlag, len(targetedPlatforms))

	for w := 1; w <= numScouts; w++ {
		scoutWg.Go(func() {
			scraper.ScoutWorker(ctx, engine, rotator.Rotate(), jobs, matches, 12*time.Second)
		})
	}

	var forensicWg sync.WaitGroup
	numForensicWorkers := 2

	for f := 1; f <= numForensicWorkers; f++ {
		forensicWg.Go(func() {
			for match := range matches {
				// Log a beautiful color-coded discovery alert line without destroying the ongoing spinner thread!
				logger.LogPrint(report.ColorGreen, "+", fmt.Sprintf("MATCH CONFIRMED: [%s] -> %s", match.PlatformName, match.ProfileURL))

				bCtx, err := engine.CreateIsolatedContext("")
				if err != nil {
					continue
				}
				page, _ := bCtx.NewPage()

				var downloadedFiles []string
				if _, err := page.Goto(match.ProfileURL, playwright.PageGotoOptions{WaitUntil: playwright.WaitUntilStateDomcontentloaded}); err == nil {
					outputDir := "Spider/storage/" + match.PlatformName
					downloadedFiles, _ = scraper.ExtractImages(page, outputDir)
				}
				page.Close()
				bCtx.Close()

				reportDataMutex.Lock()
				finalReportMatches = append(finalReportMatches, report.TargetMatch{
					PlatformName: match.PlatformName,
					ProfileURL:   match.ProfileURL,
					Downloaded:   downloadedFiles,
				})
				reportDataMutex.Unlock()
			}
		})
	}

	for _, plat := range targetedPlatforms {
		jobs <- scraper.Job{Platform: plat, Username: *usernameFlag}
	}
	close(jobs)

	scoutWg.Wait()
	close(matches)
	forensicWg.Wait()

	// 3. Stop the spinner execution cleanly before compiling outputs
	logger.StopSpinner()

	payload := report.ReportPayload{Username: *usernameFlag, Matches: finalReportMatches}
	if err := report.ExportHTMLReport("report.html", payload); err != nil {
		logger.LogPrint(report.ColorRed, "!", fmt.Sprintf("Failed to write report: %v", err))
	}

	logger.LogPrint(report.ColorBlue, "Done", "Pipeline tasks completed successfully. Final audit saved to: report.html")
}
