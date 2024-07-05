package apiv2

import (
	"context"
	"fmt"
	"net/http"

	apiv1 "github.com/nodeset-org/nodeset-client-go/api-v1"
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
	code, response, err := apiv1.SubmitRequest[MinipoolAvailableData](c.NodeSetClient, ctx, true, http.MethodGet, nil, nil, MinipoolAvailablePath)
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
		case apiv1.InvalidSessionKey:
			// Invalid session
			return MinipoolAvailableData{}, apiv1.ErrInvalidSession
		}
	}

	return MinipoolAvailableData{}, fmt.Errorf("nodeset server responded to minipool available request with code %d: [%s]", code, response.Message)
}
