// Package pdf renders HTML to PDF using a Gotenberg service.
package pdf

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{baseURL: baseURL, http: &http.Client{}}
}

// RenderHTML sends html to Gotenberg's Chromium route and returns the PDF bytes.
func (c *Client) RenderHTML(ctx context.Context, html []byte) ([]byte, error) {
	var body bytes.Buffer
	w := multipart.NewWriter(&body)

	part, err := w.CreateFormFile("files", "index.html")
	if err != nil {
		return nil, err
	}
	if _, err := part.Write(html); err != nil {
		return nil, err
	}
	// Sensible print defaults for a one-page invoice.
	_ = w.WriteField("marginTop", "0.5")
	_ = w.WriteField("marginBottom", "0.5")
	_ = w.WriteField("marginLeft", "0.5")
	_ = w.WriteField("marginRight", "0.5")
	_ = w.WriteField("printBackground", "true")
	if err := w.Close(); err != nil {
		return nil, err
	}

	url := c.baseURL + "/forms/chromium/convert/html"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	out, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("gotenberg %d: %s", resp.StatusCode, string(out))
	}
	return out, nil
}
