package apiv2

import (
	"context"
	"fmt"
	"net/http"
)

const (
	// Route for requesting minipool available count
	minipoolAvailablePath string = "modules/constellation/minipool/available"
)

// Response to a minipool available count request
type MinipoolAvailableData struct {
	// The number of new Constellation minipools the node is allowed to create
	Count int `json:"count"`
}

func (c *NodeSetClient) MinipoolAvailable(ctx context.Context) (MinipoolAvailableData, error) {
	code, response, err := SubmitRequest[MinipoolAvailableData](c, ctx, true, http.MethodGet, nil, nil, minipoolAvailablePath)
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
		case invalidSessionKey:
			// Invalid session
			return MinipoolAvailableData{}, ErrInvalidSession
		}
	}

	return MinipoolAvailableData{}, fmt.Errorf("nodeset server responded to minipool available request with code %d: [%s]", code, response.Message)
}
