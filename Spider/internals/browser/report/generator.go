package report

import (
	"encoding/base64"
	"fmt"
	"html/template"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

type TargetMatch struct {
	PlatformName string
	ProfileURL   string
	Downloaded   []string
}
type ReportPayload struct {
	Username string
	Matches  []TargetMatch
}

var templateFuncs = template.FuncMap{
	"embedImage": embedImage,
	"baseName":   filepath.Base,
}

func embedImage(path string) template.URL {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	mimeType := http.DetectContentType(data)
	if !strings.HasPrefix(mimeType, "image/") {
		return ""
	}
	encoded := base64.StdEncoding.EncodeToString(data)
	return template.URL(fmt.Sprintf("data:%s;base64,%s", mimeType, encoded))
}

const htmlTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <title>Streamlined OSINT Audit Run</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; background: #0f172a; color: #e2e8f0; margin: 40px; }
        h1 { color: #38bdf8; border-bottom: 2px solid #334155; padding-bottom: 10px; }
        .card { background: #1e293b; border-radius: 8px; padding: 20px; margin-bottom: 20px; border: 1px solid #334155; }
        .title { font-size: 20px; color: #f43f5e; margin: 0 0 10px 0; }
        .url { color: #38bdf8; text-decoration: none; }
        .gallery { display: flex; flex-wrap: wrap; gap: 12px; margin-top: 10px; }
        .thumb { width: 140px; margin: 0; }
        .thumb img { width: 100%; height: 140px; object-fit: cover; border-radius: 6px; border: 1px solid #334155; display: block; }
        .thumb figcaption { font-family: monospace; color: #94a3b8; font-size: 12px; margin-top: 4px; word-break: break-all; }
    </style>
</head>
<body>
    <h1>High-Performance Pipeline OSINT Audit Report</h1>
    <p>Target Profile Query Matrix Variable: <strong>@{{.Username}}</strong></p>
    {{range .Matches}}
    <div class="card">
        <div class="title">{{.PlatformName}} Endpoint Match</div>
        <p>Location URL: <a class="url" href="{{.ProfileURL}}" target="_blank">{{.ProfileURL}}</a></p>
        <p><strong>Harvested Media Source Artifacts:</strong></p>
        {{if .Downloaded}}
        <div class="gallery">
            {{range .Downloaded}}
            <figure class="thumb">
                <a href="{{embedImage .}}" target="_blank">
                    <img src="{{embedImage .}}" alt="{{baseName .}}" loading="lazy">
                </a>
                <figcaption>{{baseName .}}</figcaption>
            </figure>
            {{end}}
        </div>
        {{else}}
        <p style="color: #64748b; font-style: italic;">No static visual image file handles preserved from destination host framework DOM.</p>
        {{end}}
    </div>
    {{end}}
</body>
</html>`

func ExportHTMLReport(outputPath string, payload ReportPayload) error {
	tmpl, err := template.New("ReportPayload").Funcs(templateFuncs).Parse(htmlTemplate)
	if err != nil {
		return err
	}
	file, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer file.Close()
	return tmpl.Execute(file, payload)
}
