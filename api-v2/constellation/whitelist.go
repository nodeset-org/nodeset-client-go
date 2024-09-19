package v2constellation

import (
	"context"
	"fmt"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common"
)

const (
	// Route for requesting whitelist signature
	WhitelistPath string = "whitelist"
)

// Response to a whitelist GET request
type Whitelist_GetData struct {
	// Whether the user has a whitelisted node
	Whitelisted bool `json:"whitelisted"`

	// The address of the whitelisted node for the user account
	Address ethcommon.Address `json:"address,omitempty"`
}

// Response to a whitelist POST request
type Whitelist_PostData struct {
	// The signature for Whitelist.addOperator()
	Signature string `json:"signature"`
}

func (c *V2ConstellationClient) Whitelist_Get(ctx context.Context, deployment string) (Whitelist_GetData, error) {
	// Send the request
	path := ConstellationPrefix + deployment + "/" + WhitelistPath
	code, response, err := common.SubmitRequest[Whitelist_GetData](c.commonClient, ctx, true, http.MethodGet, nil, nil, path)
	if err != nil {
		return Whitelist_GetData{}, fmt.Errorf("error requesting whitelist signature: %w", err)
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		// Success
		return response.Data, nil

	case http.StatusBadRequest:
		switch response.Error {
		case common.InvalidDeploymentKey:
			// Invalid deployment
			return Whitelist_GetData{}, common.ErrInvalidDeployment
		}
	}

	return Whitelist_GetData{}, fmt.Errorf("nodeset server responded to whitelist-get request with code %d: [%s]", code, response.Message)
}

func (c *V2ConstellationClient) Whitelist_Post(ctx context.Context, deployment string) (Whitelist_PostData, error) {
	// Send the request
	path := ConstellationPrefix + deployment + "/" + WhitelistPath
	code, response, err := common.SubmitRequest[Whitelist_PostData](c.commonClient, ctx, true, http.MethodPost, nil, nil, path)
	if err != nil {
		return Whitelist_PostData{}, fmt.Errorf("error requesting whitelist signature: %w", err)
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		// Successs
		return response.Data, nil

	case http.StatusBadRequest:
		switch response.Error {
		case common.InvalidDeploymentKey:
			// Invalid deployment
			return Whitelist_PostData{}, common.ErrInvalidDeployment

		case IncorrectNodeAddressKey:
			// Incorrect node address
			return Whitelist_PostData{}, ErrIncorrectNodeAddress
		}

	case http.StatusUnauthorized:
		switch response.Error {
		case UserNotAuthorizedKey:
			// User not authorized to whitelist for Constellation
			return Whitelist_PostData{}, ErrNotAuthorized

		case common.InvalidSessionKey:
			// Invalid session
			return Whitelist_PostData{}, common.ErrInvalidSession
		}
	}

	return Whitelist_PostData{}, fmt.Errorf("nodeset server responded to whitelist-post request with code %d: [%s]", code, response.Message)
}
