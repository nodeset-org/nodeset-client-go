package common

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	// Header to include when sending messages that have been logged in
	AuthHeader string = "Authorization"

	// Format for the authorization header
	AuthHeaderFormat string = "Bearer %s"
)

// Client for interacting with the NodeSet server
type CommonNodeSetClient struct {
	BaseUrl      string
	SessionToken string
	HttpClient   *http.Client
}

// Creates a new NodeSet client
// baseUrl: The base URL to use for the client, for example [https://nodeset.io/api]
func NewCommonNodeSetClient(baseUrl string, timeout time.Duration) *CommonNodeSetClient {
	return &CommonNodeSetClient{
		BaseUrl: baseUrl, // v1 doesn't have a version in the subroute so just use the base URL
		HttpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Set the session token for the client after logging in
func (c *CommonNodeSetClient) SetSessionToken(token string) {
	c.SessionToken = token
}

// Send a request to the server and read the response
func SubmitRequest[DataType any](c *CommonNodeSetClient, ctx context.Context, requireAuth bool, method string, body io.Reader, queryParams map[string]string, subroutes ...string) (int, NodeSetResponse[DataType], error) {
	var defaultVal NodeSetResponse[DataType]

	// Make the request
	path, err := url.JoinPath(c.BaseUrl, subroutes...)
	if err != nil {
		return 0, defaultVal, fmt.Errorf("error joining path [%v]: %w", subroutes, err)
	}
	request, err := http.NewRequestWithContext(ctx, method, path, body)
	if err != nil {
		return 0, defaultVal, fmt.Errorf("error generating request to [%s]: %w", path, err)
	}
	query := request.URL.Query()
	for name, value := range queryParams {
		query.Add(name, value)
	}
	request.URL.RawQuery = query.Encode()

	// Set the headers
	if requireAuth {
		if c.SessionToken == "" {
			return 0, defaultVal, ErrInvalidSession
		}
		request.Header.Set(AuthHeader, fmt.Sprintf(AuthHeaderFormat, c.SessionToken))
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	// Upload it to the server
	resp, err := c.HttpClient.Do(request)
	if err != nil {
		return 0, defaultVal, fmt.Errorf("error submitting request to nodeset server: %w", err)
	}

	// Read the body
	defer resp.Body.Close()
	bytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, defaultVal, fmt.Errorf("nodeset server responded to request with code %s but reading the response body failed: %w", resp.Status, err)
	}

	// Unmarshal the response
	var response NodeSetResponse[DataType]
	err = json.Unmarshal(bytes, &response)
	if err != nil {
		return 0, defaultVal, fmt.Errorf("nodeset server responded to request with code %s and unmarshalling the response failed: [%w]... original body: [%s]", resp.Status, err, string(bytes))
	}

	// Debug log
	return resp.StatusCode, response, nil
}
