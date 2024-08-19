package stakewise

import (
	"context"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	apiv0 "github.com/nodeset-org/nodeset-client-go/api-v0"
)

const (
	// Route for requesting deployments
	DeploymentsPath string = "deployments"
)

// Response to a whitelist request
type WhitelistData struct {
	// The signature for Whitelist.addOperator()
	Signature string `json:"signature"`
	Time      int64  `json:"time"`
}

func (c *StakewiseV2Module) Whitelist(ctx context.Context, chainId *big.Int, whitelistAddress common.Address) (WhitelistData, error) {
	args := map[string]string{
		"chainId":          chainId.String(),
		"whitelistAddress": whitelistAddress.Hex(),
	}
	code, response, err := apiv0.SubmitRequest[WhitelistData](c.NodeSetClient, ctx, true, http.MethodGet, nil, args, c.routes.Whitelist)
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
		case UserNotAuthorizedKey:
			// User not authorized to whitelist for Constellation
			return WhitelistData{}, ErrNotAuthorized

		case apiv0.InvalidSessionKey:
			// Invalid session
			return WhitelistData{}, apiv0.ErrInvalidSession
		}
	}

	return WhitelistData{}, fmt.Errorf("nodeset server responded to whitelist request with code %d: [%s]", code, response.Message)
}
