package apiv2

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
	MessageKey string = "message"
	NonceKey   string = "nonce"
)

const (
	// API version to use
	apiVersion string = "v2"

	// Header to include when sending messages that have been logged in
	authHeader string = "Authorization"

	// Format for the authorization header
	authHeaderFormat string = "Bearer %s"
)

// Client for interacting with the NodeSet server
type NodeSetClient struct {
	baseUrl      string
	sessionToken string
	client       *http.Client
}

// Creates a new NodeSet client
func NewNodeSetClient(baseUrl string, timeout time.Duration) *NodeSetClient {
	return &NodeSetClient{
		baseUrl: fmt.Sprintf("%s/%s", baseUrl, apiVersion),
		client: &http.Client{
			Timeout: timeout,
		},
	}
}

// Set the session token for the client after logging in
func (c *NodeSetClient) SetSessionToken(sessionToken string) {
	c.sessionToken = sessionToken
}

// Send a request to the server and read the response
func SubmitRequest[DataType any](c *NodeSetClient, ctx context.Context, requireAuth bool, method string, body io.Reader, queryParams map[string]string, subroutes ...string) (int, NodeSetResponse[DataType], error) {
	var defaultVal NodeSetResponse[DataType]

	// Make the request
	path, err := url.JoinPath(c.baseUrl, subroutes...)
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
		if c.sessionToken == "" {
			return 0, defaultVal, ErrInvalidSession
		}
		request.Header.Set(authHeader, fmt.Sprintf(authHeaderFormat, c.sessionToken))
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	// Upload it to the server
	resp, err := c.client.Do(request)
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
