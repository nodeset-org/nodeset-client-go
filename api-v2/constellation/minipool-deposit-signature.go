package v2constellation

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"math/big"
	"net/http"

	"github.com/goccy/go-json"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common"
)

const (

	// Route for requesting the signature to create a minipool
	MinipoolDepositSignaturePath string = "minipool/deposit-signature"
)

// Request to generate signature for SuperNodeAccount.createMinipool()
type MinipoolDepositSignatureRequest struct {
	// the EIP55-compliant hex string representing the address of the
	// minipool that will be created (the address to generate the signature for)
	MinipoolAddress ethcommon.Address `json:"minipoolAddress"`

	// a hex string (lower-case, no 0x prefix) representing the 32-byte salt used
	// to create minipoolAddress during CREATE2 calculation
	Salt string `json:"salt"`
}

// Response to a create minipool signature request
type MinipoolDepositSignatureData struct {
	// The signature for SuperNodeAccount.createMinipool()
	Signature string `json:"signature"`
}

func (c *V2ConstellationClient) MinipoolDepositSignature(ctx context.Context, logger *slog.Logger, deployment string, minipoolAddress ethcommon.Address, salt *big.Int) (MinipoolDepositSignatureData, error) {
	// Create the request body
	request := MinipoolDepositSignatureRequest{
		MinipoolAddress: minipoolAddress,
		Salt:            salt.Text(16),
	}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return MinipoolDepositSignatureData{}, fmt.Errorf("error marshalling minipool deposit signature request: %w", err)
	}
	common.SafeDebugLog(logger, "Prepared minipool deposit signature body",
		"body", request,
	)

	// Send the request
	path := ConstellationPrefix + deployment + "/" + MinipoolDepositSignaturePath
	code, response, err := common.SubmitRequest[MinipoolDepositSignatureData](c.commonClient, ctx, logger, true, http.MethodPost, bytes.NewBuffer(jsonData), nil, path)
	if err != nil {
		return MinipoolDepositSignatureData{}, fmt.Errorf("error requesting minipool deposit signature: %w", err)
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		// Successfully generated minipool deposit signature
		return response.Data, nil

	case http.StatusBadRequest:
		switch response.Error {
		case common.InvalidDeploymentKey:
			// Invalid deployment
			return MinipoolDepositSignatureData{}, common.ErrInvalidDeployment

		case common.MissingWhitelistedNodeAddressKey:
			// Node address not whitelisted for deployment
			return MinipoolDepositSignatureData{}, common.ErrMissingWhitelistedNodeAddress

		case common.IncorrectNodeAddressKey:
			// Incorrect node address
			return MinipoolDepositSignatureData{}, common.ErrIncorrectNodeAddress
		}

	case http.StatusUnauthorized:
		switch response.Error {
		case common.InvalidSessionKey:
			// Invalid session
			return MinipoolDepositSignatureData{}, common.ErrInvalidSession
		}

	case http.StatusForbidden:
		switch response.Error {
		case common.MinipoolLimitReachedKey:
			// Address has been given access to Constellation, but cannot create any more minipools.
			return MinipoolDepositSignatureData{}, common.ErrMinipoolLimitReached

		case common.MissingExitMessageKey:
			// Nodeset.io is missing a signed exit message for a previous minipool
			return MinipoolDepositSignatureData{}, common.ErrMissingExitMessage

		case common.AddressAlreadyRegisteredKey:
			// A minipool with this address already exists
			return MinipoolDepositSignatureData{}, common.ErrAddressAlreadyRegistered

		case common.InvalidPermissionsKey:
			// The user doesn't have permission to do this
			return MinipoolDepositSignatureData{}, common.ErrInvalidPermissions
		}
	}

	return MinipoolDepositSignatureData{}, fmt.Errorf("nodeset server responded to minipool deposit signature request with code %d: [%s]", code, response.Message)
}
