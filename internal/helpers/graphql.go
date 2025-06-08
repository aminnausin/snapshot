package helpers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/hasura/go-graphql-client"
)

// StringPtrOrNil handles an empty string for use with a graphql query.
// It returns nil if the given string is empty. Otherwise, it returns the string.
func StringPtrOrNil(v graphql.String) *string {
	if v == "" {
		return nil
	}
	s := string(v)
	return &s
}

// RunQuery sends a graphql query that is defined by a struct with the hasura graphql client.
// Variables are passed into the query by the client.
// It returns an error if present and directly puts the result in to the query object sent as a parameter.
func RunQuery(client *graphql.Client, query any, variables map[string]any) error {
	if variables == nil {
		variables = make(map[string]any)
	}

	err := client.Query(context.Background(), query, variables)
	if err != nil {
		log.Fatalf("Failed GraphQL query: %s\n%s", err, variables)
		return err
	}

	return nil
}

// RunRawQuery sends a raw graphql query from a query string using a base http client.
// It encodes the query string into json and sends it to the Github graphql endpoint.
// It returns the response, decoded from json into a map of string value pairs.
func RunRawQuery(httpClient *http.Client, query string) (map[string]any, error) {
	// Encode query string into json
	payload := map[string]string{"query": query}
	body, _ := json.Marshal(payload)

	// Build query from json payload
	req, err := http.NewRequest("POST", "https://api.github.com/graphql", bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	// Send request
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	// Decode json response into map
	var jsonResp map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&jsonResp); err != nil {
		return nil, fmt.Errorf("failed to decode JSON: %w", err)
	}

	// Check for api errors
	if errs, ok := jsonResp["errors"]; ok {
		return nil, fmt.Errorf("graphql error: %v", errs)
	}

	// Check for returned data
	data, ok := jsonResp["data"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("missing data field in response")
	}

	return data, nil
}
