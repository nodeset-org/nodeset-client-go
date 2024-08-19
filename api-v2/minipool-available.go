package apiv2

import (
	"context"
	"fmt"
	"net/http"

	apiv0 "github.com/nodeset-org/nodeset-client-go/api-v0"
)

const (
	// Route for requesting minipool available count
	MinipoolAvailablePath string = "minipool/available"
)

// Response to a minipool available count request
type MinipoolAvailableData struct {
	// The number of new Constellation minipools the node is allowed to create
	Count int `json:"count"`
}

func (c *NodeSetClient) MinipoolAvailable(ctx context.Context) (MinipoolAvailableData, error) {
	code, response, err := apiv0.SubmitRequest[MinipoolAvailableData](c.NodeSetClient, ctx, true, http.MethodGet, nil, nil, c.routes.MinipoolAvailable)
	if err != nil {
		return MinipoolAvailableData{}, fmt.Errorf("error requesting minipool available count: %w", err)
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		// Successfully retrieved minipool available count
		return response.Data, nil

	case http.StatusForbidden:
		switch response.Error {
		case apiv0.InvalidSessionKey:
			// Invalid session
			return MinipoolAvailableData{}, apiv0.ErrInvalidSession
		}
	}

	return MinipoolAvailableData{}, fmt.Errorf("nodeset server responded to minipool available request with code %d: [%s]", code, response.Message)
}
