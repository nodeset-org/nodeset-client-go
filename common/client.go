package common

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/goccy/go-json"
)

const (
	// Header to include when sending messages that have been logged in
	AuthHeader string = "Authorization"

	// Format for the authorization header
	AuthHeaderFormat string = "Bearer %s"
)

// Client for interacting with the NodeSet server
type CommonNodeSetClient struct {
	baseUrl      string
	sessionToken string
	httpClient   *http.Client
}

// Creates a new NodeSet client
// baseUrl: The base URL to use for the client, for example [https://nodeset.io/api]
func NewCommonNodeSetClient(baseUrl string, timeout time.Duration) *CommonNodeSetClient {
	return &CommonNodeSetClient{
		baseUrl: baseUrl,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

// Set the session token for the client after logging in
func (c *CommonNodeSetClient) SetSessionToken(token string) {
	c.sessionToken = token
}

// =============================
// === HTTP Request Handling ===
// =============================

// All responses from the NodeSet API will have this format
// `message` may or may not be populated (but should always be populated if `ok` is false)
// `data` should be populated if `ok` is true, and will be omitted if `ok` is false
type NodeSetResponse[DataType any] struct {
	OK      bool     `json:"ok"`
	Message string   `json:"message,omitempty"`
	Data    DataType `json:"data,omitempty"`
	Error   string   `json:"error,omitempty"`
}

// Send a request to the server and read the response
// NOTE: this is better suited to be a method of c but Go doesn't allow for generic methods yet
func SubmitRequest[DataType any](c *CommonNodeSetClient, ctx context.Context, logger *slog.Logger, requireAuth bool, method string, body io.Reader, queryParams map[string]string, subroutes ...string) (int, NodeSetResponse[DataType], error) {
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
	SafeDebugLog(logger, "Submitting request to NodeSet server",
		"method", method,
		"path", path,
		"query", request.URL.RawQuery,
	)

	// Set the headers
	if requireAuth {
		if c.sessionToken == "" {
			return 0, defaultVal, ErrInvalidSession
		}
		request.Header.Set(AuthHeader, fmt.Sprintf(AuthHeaderFormat, c.sessionToken))
	}
	request.Header.Set("Content-Type", "application/json; charset=UTF-8")

	// Upload it to the server
	resp, err := c.httpClient.Do(request)
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
	SafeDebugLog(logger, "Received response from NodeSet server",
		"status", resp.Status,
		"response", response,
	)
	return resp.StatusCode, response, nil
}
