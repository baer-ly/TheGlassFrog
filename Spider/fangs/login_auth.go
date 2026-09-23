package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/mxschmitt/playwright-go"
)

func main() {
	pw, err := playwright.Run()
	if err != nil {
		log.Fatalf("Could not start Playwright engine: %v", err)
	}
	defer func(pw *playwright.Playwright) {
		err := pw.Stop()
		if err != nil {

		}
	}(pw)

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(false),
	})
	if err != nil {
		log.Fatalf("Could not launch Chromium browser: %v", err)
	}
	defer browser.Close()

	context, err := browser.NewContext()
	if err != nil {
		log.Fatalf("Could not create browser context: %v", err)
	}
	defer context.Close()

	page, err := context.NewPage()
	if err != nil {
		log.Fatalf("Could not open page: %v", err)
	}

	targetURL := "https://instagram.com"
	fmt.Printf("[*] Opening authentication window: %s\n", targetURL)
	if _, err = page.Goto(targetURL); err != nil {
		log.Fatalf("Navigation error: %v", err)
	}

	fmt.Println("\n[!] ACTION REQUIRED:")
	fmt.Println("    1. Log into your Instagram account manually in the browser window.")
	fmt.Println("    2. Navigate past any 2FA or security checkpoints.")
	fmt.Println("    3. Once you see your home feed, type 'done' here in your terminal and press enter.")

	// FIX: Use bufio.NewScanner to read text input securely and prevent premature skipping
	reader := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("\nType 'done' when ready: ")
		if reader.Scan() {
			text := strings.TrimSpace(strings.ToLower(reader.Text()))
			if text == "done" {
				break
			}
		}
	}

	// Extract the active session tokens
	cookies, err := context.Cookies()
	if err != nil {
		log.Fatalf("Failed to extract active session tokens: %v", err)
	}

	// Safety check: Don't write the file if Instagram blocked the session and returned 0 tokens
	if len(cookies) == 0 {
		log.Fatalf("[-] ERROR: 0 cookies captured. You must be fully logged into Instagram's home feed before typing 'done'.")
	}

	cookieJSON, err := json.MarshalIndent(cookies, "", "  ")
	if err != nil {
		log.Fatalf("Failed to format cookie JSON: %v", err)
	}

	outputFile := "cookies.json"
	err = os.WriteFile(outputFile, cookieJSON, 0600)
	if err != nil {
		log.Fatalf("Failed to save auth state file to disk: %v", err)
	}

	fmt.Printf("\n[+] SUCCESS: %d active authentication cookies saved securely to -> %s\n", len(cookies), outputFile)
}
