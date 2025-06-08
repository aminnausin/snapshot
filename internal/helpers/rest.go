package helpers

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

func RunRestQuery(client *http.Client, path string, queryParams map[string]string) ([]byte, error) {
	// Make a request to the REST API
	// :param client: HTTP client
	// :param path: API path to query
	// :param params: Query parameters to be passed to the API
	// :return: deserialized REST JSON output

	baseURL := "https://api.github.com"
	fullURL := fmt.Sprintf("%s/%s", strings.TrimRight(baseURL, "/"), strings.TrimLeft(path, "/"))

	// Attempt query 60 times until the response is ready (not a 202 response) or too many responses were received
	for range 60 {
		// Raw http request -> needs to be recreated on every loop
		rawRequest, err := http.NewRequest("GET", fullURL, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to build request: %w", err)
		}

		// Setup request with headers, parameters and URI encoding
		rawRequest.Header.Set("Accept", "application/vnd.github+json")
		q := rawRequest.URL.Query()
		for key, value := range queryParams {
			q.Add(key, value)
		}
		rawRequest.URL.RawQuery = q.Encode()

		resp, err := client.Do(rawRequest)
		// If error, wait 2 seconds and then resend the request
		if err != nil {
			log.Printf("HTTP request failed: %v", err)
			time.Sleep(2 * time.Second)
			continue
		}

		defer resp.Body.Close()

		// If 202, wait 2 seconds and then resend the request
		if resp.StatusCode == http.StatusAccepted {
			log.Printf("GitHub returned 202 for %s. Waiting and retrying...", path)
			time.Sleep(2 * time.Second)
			continue
		}

		if resp.StatusCode != http.StatusOK {
			body, _ := io.ReadAll(resp.Body)
			return nil, fmt.Errorf("unexpected status %d from %s: %s", resp.StatusCode, path, body)
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		return body, nil
	}

	return nil, fmt.Errorf("too many 202 responses from GitHub for %s", path)
}

func (t *TransportWithToken) RoundTrip(req *http.Request) (*http.Response, error) {
	req.Header.Set("Authorization", "Bearer "+t.Token)
	req.Header.Set("Accept", "application/vnd.github+json")
	return t.Transport.RoundTrip(req)
}
