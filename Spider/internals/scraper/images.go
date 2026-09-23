package scraper

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"

	"github.com/mxschmitt/playwright-go"
)

func ExtractImages(page playwright.Page, outputDir string) ([]string, error) {
	images, err := page.Locator("img").All()
	if err != nil {
		return nil, err
	}
	var savePaths []string
	baseURL, _ := url.Parse(page.URL())
	_ = os.MkdirAll(outputDir, 0755)

	for i, img := range images {
		src, err := img.GetAttribute("src")
		if err != nil || src == "" {
			continue
		}
		parsedSrc, err := url.Parse(src)
		if err != nil {
			continue
		}
		absolutePath := baseURL.ResolveReference(parsedSrc).String()
		filePath := filepath.Join(outputDir, fmt.Sprintf("image_%d.jpg", i))

		if err := downloadFile(absolutePath, filePath); err == nil {
			savePaths = append(savePaths, filePath)
		}
	}
	return savePaths, nil
}
func downloadFile(urlStr string, destPath string) error {
	resp, err := http.Get(urlStr)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	out, err := os.Create(destPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, resp.Body)
	return err
}
