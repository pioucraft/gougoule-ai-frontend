package functions

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// Import strings for TrimSpace

func SimpleWebSearch(query string) (string, error) {

	escapedQuery := url.QueryEscape(query)
	// Construct the URL using the standard query parameter format for DuckDuckGo HTML
	searchURL := fmt.Sprintf("https://html.duckduckgo.com/html/?q=%s", escapedQuery)

	// Make the HTTP GET request with a human-like User-Agent
	client := &http.Client{}
	req, err := http.NewRequest("GET", searchURL, nil)
	if err != nil {
		return "", fmt.Errorf("failed to create HTTP request: %w", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 (KHTML, like Gecko) Version/18.3.1 Safari/605.1.15")

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL %s: %w", searchURL, err)
	}
	// Ensure the response body is closed
	defer resp.Body.Close()

	// Check if the request was successful
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusAccepted {
		// Optionally, read the body even on error to provide more context
		return "", fmt.Errorf("received non-200/202 status code %d from %s", resp.StatusCode, searchURL)
	}
	// Corrected variable name from res.Body to resp.Body
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %w", err)
	}

	// Find elements with class 'result__snippet' and print the first 5
	count := 0
	answer := ""
	doc.Find(".result__snippet").EachWithBreak(func(i int, s *goquery.Selection) bool {
		if count >= 5 {
			return false // Stop after finding 5
		}
		// Get the text content of the snippet and trim leading/trailing whitespace
		snippetText := strings.TrimSpace(s.Text())

		answer += fmt.Sprintf("%d: %s\n", i+1, snippetText)
		count++
		return true // Continue searching
	})

	return answer, nil
}
