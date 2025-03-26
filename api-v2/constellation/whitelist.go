package v2constellation

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"path"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common"
)

const (
	// The node isn't authorized to register with Constellation
	NodeUnauthorizedKey string = "node_unauthorized"

	// Route for requesting whitelist signature
	WhitelistPath string = "whitelist"
)

var (
	// The node isn't authorized to register with Constellation
	ErrNodeUnauthorized error = errors.New("node isn't authorized to register with Constellation")
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

func (c *V2ConstellationClient) Whitelist_Get(ctx context.Context, logger *slog.Logger, deployment string) (Whitelist_GetData, error) {
	// Send the request
	pathString := path.Join(ConstellationPrefix, deployment, WhitelistPath)
	code, response, err := common.SubmitRequest[Whitelist_GetData](c.commonClient, ctx, logger, true, http.MethodGet, nil, nil, pathString)
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

	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid or expired session
			return Whitelist_GetData{}, common.ErrInvalidSession
		}

	case http.StatusForbidden:
		switch response.Error {
		case common.InvalidPermissionsKey:
			// The user doesn't have permission to do this
			return Whitelist_GetData{}, common.ErrInvalidPermissions
		}
	}

	return Whitelist_GetData{}, fmt.Errorf("nodeset server responded to whitelist-get request with code %d: [%s]", code, response.Message)
}

func (c *V2ConstellationClient) Whitelist_Post(ctx context.Context, logger *slog.Logger, deployment string) (Whitelist_PostData, error) {
	// Send the request
	pathString := path.Join(ConstellationPrefix, deployment, WhitelistPath)
	code, response, err := common.SubmitRequest[Whitelist_PostData](c.commonClient, ctx, logger, true, http.MethodPost, nil, nil, pathString)
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

		case common.IncorrectNodeAddressKey:
			// Incorrect node address
			return Whitelist_PostData{}, common.ErrIncorrectNodeAddress
		}

	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid session
			return Whitelist_PostData{}, common.ErrInvalidSession
		}

	case http.StatusForbidden:
		switch response.Error {
		case common.InvalidPermissionsKey:
			// The user doesn't have permission to do this
			return Whitelist_PostData{}, common.ErrInvalidPermissions

		case NodeUnauthorizedKey:
			// Node isn't authorized to register with Constellation
			return Whitelist_PostData{}, ErrNodeUnauthorized
		}
	}

	return Whitelist_PostData{}, fmt.Errorf("nodeset server responded to whitelist-post request with code %d: [%s]", code, response.Message)
}
