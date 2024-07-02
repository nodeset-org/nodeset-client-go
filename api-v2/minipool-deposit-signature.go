package apiv2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/utils"
)

const (
	// Route for requesting the signature to create a minipool
	minipoolDepositSignaturePath string = "modules/constellation/minipool/deposit-signature"
)

// Request to generate signature for SuperNodeAccount.createMinipool()
type MinipoolDepositSignatureRequest struct {
	// the EIP55-compliant hex string representing the address of the
	// minipool that will be created (the address to generate the signature for)
	Address string `json:"address"`
	// a hex string (lower-case, no 0x prefix) representing the 32-byte salt used
	// to create minipoolAddress during CREATE2 calculation
	Salt string `json:"salt"`
}

// Response to a create minipool signature request
type MinipoolDepositSignatureData struct {
	// The signature for SuperNodeAccount.createMinipool()
	Signature string `json:"signature"`
}

func (c *NodeSetClient) MinipoolDepositSignature(ctx context.Context, address common.Address, salt []byte) (MinipoolDepositSignatureData, error) {
	request := MinipoolDepositSignatureRequest{
		Address: address.Hex(),
		Salt:    utils.EncodeHexWithPrefix(salt),
	}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return MinipoolDepositSignatureData{}, fmt.Errorf("error marshalling minipool deposit signature request: %w", err)
	}
	code, response, err := SubmitRequest[MinipoolDepositSignatureData](c, ctx, true, http.MethodPost, bytes.NewBuffer(jsonData), nil, minipoolDepositSignaturePath)
	if err != nil {
		return MinipoolDepositSignatureData{}, fmt.Errorf("error requesting minipool deposit signature: %w", err)
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		// Successfully generated minipool deposit signature
		return response.Data, nil

	case http.StatusForbidden:
		switch response.Error {
		case minipoolLimitReachedKey:
			// Address has been given access to Constellation, but cannot create any more minipools.
			return MinipoolDepositSignatureData{}, ErrMinipoolLimitReached
		case missingExitMessageKey:
			// Address has been given access to Constellation, but the NodeSet service does not have
			// a signed exit message stored for the minipool that user account previously created.
			return MinipoolDepositSignatureData{}, ErrMissingExitMessage
		}

	case http.StatusUnauthorized:
		switch response.Error {
		case addressNotAuthorizedKey:
			// Address not authorized to whitelist for Constellation
			return MinipoolDepositSignatureData{}, ErrNotAuthorized
		case invalidSessionKey:
			// Invalid session
			return MinipoolDepositSignatureData{}, ErrInvalidSession
		}
	}

	return MinipoolDepositSignatureData{}, fmt.Errorf("nodeset server responded to minipool deposit signature request with code %d: [%s]", code, response.Message)
}
