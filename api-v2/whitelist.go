package apiv2

import (
	"context"
	"fmt"
	"net/http"
)

const (
	// Route for requesting whitelist signature
	whitelistPath string = "modules/constellation/whitelist"
)

// Response to a whitelist request
type WhitelistData struct {
	// The signature for Whitelist.addOperator()
	Signature string `json:"signature"`
}

func (c *NodeSetClient) Whitelist(ctx context.Context) (WhitelistData, error) {
	code, response, err := SubmitRequest[WhitelistData](c, ctx, true, http.MethodGet, nil, nil, whitelistPath)
	if err != nil {
		return WhitelistData{}, fmt.Errorf("error requesting whitelist signature: %w", err)
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		// Node successfully registered
		return response.Data, nil

	case http.StatusUnauthorized:
		switch response.Error {
		case userNotAuthorizedKey:
			// User not authorized to whitelist for Constellation
			return WhitelistData{}, ErrNotAuthorized

		case invalidSessionKey:
			// Invalid session
			return WhitelistData{}, ErrInvalidSession
		}
	}

	return WhitelistData{}, fmt.Errorf("nodeset server responded to whitelist request with code %d: [%s]", code, response.Message)
}
