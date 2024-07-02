package apiv2

import (
	"context"
	"fmt"
	"net/http"
)

// Response to a minipool available count request
type MinipoolAvailableData struct {
	// The signature for Whitelist.addOperator()
	Count int `json:"count"`
}

const (
	// Route for requesting minipool available count
	minipoolAvailablePath string = "modules/constellation/minipool/available"
)

func (c *NodeSetClient) MinipoolAvailable(ctx context.Context) (MinipoolAvailableData, error) {
	code, response, err := SubmitRequest[MinipoolAvailableData](c, ctx, true, http.MethodGet, nil, nil, minipoolAvailablePath)
	if err != nil {
		return MinipoolAvailableData{}, fmt.Errorf("error requesting whitelist signature: %w", err)
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		// Node successfully registered
		return response.Data, nil

	case http.StatusBadRequest:
		switch response.Error {
		case invalidSessionKey:
			// Invalid session
			return MinipoolAvailableData{}, ErrInvalidSession
		}
	}

	return MinipoolAvailableData{}, fmt.Errorf("nodeset server responded to minipool available request with code %d: [%s]", code, response.Message)
}
