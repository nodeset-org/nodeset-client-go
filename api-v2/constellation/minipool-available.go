package v2constellation

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nodeset-org/nodeset-client-go/common"
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

func (c *V2ConstellationClient) MinipoolAvailable(ctx context.Context, deployment string) (MinipoolAvailableData, error) {
	// Send the request
	path := ConstellationPrefix + deployment + "/" + MinipoolAvailablePath
	code, response, err := common.SubmitRequest[MinipoolAvailableData](c.commonClient, ctx, true, http.MethodGet, nil, nil, path)
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
		case common.InvalidSessionKey:
			// Invalid session
			return MinipoolAvailableData{}, common.ErrInvalidSession
		}
	}

	return MinipoolAvailableData{}, fmt.Errorf("nodeset server responded to minipool available request with code %d: [%s]", code, response.Message)
}
