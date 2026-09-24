package scraper

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/mxschmitt/playwright-go"
)

var httpClient = &http.Client{
	Timeout: 15 * time.Second,
}

func ExtractImages(page playwright.Page, outputDir string) ([]string, error) {
	images, err := page.Locator("img").All()
	if err != nil {
		return nil, err
	}

	baseURL, err := url.Parse(page.URL())
	if err != nil {
		return nil, fmt.Errorf("could not parse page URL: %w", err)
	}

	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return nil, fmt.Errorf("could not create output dir: %w", err)
	}

	seen := make(map[string]bool)
	var savePaths []string

	for i, img := range images {
		src, err := img.GetAttribute("src")
		if err != nil || src == "" {
			continue
		}
		if strings.HasPrefix(src, "data:") {
			continue // inline base64 images aren't fetchable via HTTP
		}

		parsedSrc, err := url.Parse(src)
		if err != nil {
			continue
		}
		resolved := baseURL.ResolveReference(parsedSrc)

		if seen[resolved.String()] {
			continue // same image referenced more than once on the page
		}
		seen[resolved.String()] = true

		filePath, err := downloadImage(resolved, outputDir, i)
		if err != nil {
			continue // one bad image shouldn't abort the rest
		}
		savePaths = append(savePaths, filePath)
	}

	return savePaths, nil
}

func downloadImage(imageURL *url.URL, outputDir string, index int) (string, error) {
	req, err := http.NewRequest(http.MethodGet, imageURL.String(), nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("unexpected status %d for %s", resp.StatusCode, imageURL)
	}

	ext := extensionFor(resp.Header.Get("Content-Type"), imageURL.Path)
	destPath := filepath.Join(outputDir, fmt.Sprintf("image_%d%s", index, ext))

	out, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		os.Remove(destPath) // don't leave a truncated file behind
		return "", err
	}

	return destPath, nil
}

func extensionFor(contentType, urlPath string) string {
	switch {
	case strings.Contains(contentType, "png"):
		return ".png"
	case strings.Contains(contentType, "webp"):
		return ".webp"
	case strings.Contains(contentType, "gif"):
		return ".gif"
	case strings.Contains(contentType, "svg"):
		return ".svg"
	case strings.Contains(contentType, "jpeg"):
		return ".jpg"
	}
	if ext := filepath.Ext(urlPath); ext != "" && len(ext) <= 5 {
		return ext
	}
	return ".jpg"
}
