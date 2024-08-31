package v2constellation

import (
	"context"
	"fmt"
	"net/http"

	"github.com/nodeset-org/nodeset-client-go/common"
)

const (
	// Route for requesting whitelist signature
	WhitelistPath string = "whitelist"
)

// Response to a whitelist request
type WhitelistData struct {
	// The signature for Whitelist.addOperator()
	Signature string `json:"signature"`
}

func (c *V2ConstellationClient) Whitelist(ctx context.Context, deployment string) (WhitelistData, error) {
	// Send the request
	path := ConstellationPrefix + deployment + "/" + WhitelistPath
	code, response, err := common.SubmitRequest[WhitelistData](c.commonClient, ctx, true, http.MethodGet, nil, nil, path)
	if err != nil {
		return WhitelistData{}, fmt.Errorf("error requesting whitelist signature: %w", err)
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		// Node successfully registered
		return response.Data, nil

	case http.StatusBadRequest:
		switch response.Error {
		case common.InvalidDeploymentKey:
			// Invalid deployment
			return WhitelistData{}, common.ErrInvalidDeployment
		}

	case http.StatusUnauthorized:
		switch response.Error {
		case UserNotAuthorizedKey:
			// User not authorized to whitelist for Constellation
			return WhitelistData{}, ErrNotAuthorized

		case common.InvalidSessionKey:
			// Invalid session
			return WhitelistData{}, common.ErrInvalidSession
		}
	}

	return WhitelistData{}, fmt.Errorf("nodeset server responded to whitelist request with code %d: [%s]", code, response.Message)
}
