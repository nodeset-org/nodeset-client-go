package apiv2

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	apiv1 "github.com/nodeset-org/nodeset-client-go/api-v1"
	"github.com/rocket-pool/node-manager-core/utils"
)

const (
	// Route for requesting the signature to create a minipool
	MinipoolDepositSignaturePath string = "minipool/deposit-signature"
)

// Request to generate signature for SuperNodeAccount.createMinipool()
type MinipoolDepositSignatureRequest struct {
	// the EIP55-compliant hex string representing the address of the
	// minipool that will be created (the address to generate the signature for)
	Address string `json:"address"`

	// a hex string (lower-case, no 0x prefix) representing the 32-byte salt used
	// to create minipoolAddress during CREATE2 calculation
	Salt string `json:"salt"`

	// the chain ID of the network the minipool will be created on
	ChainId string `json:"chainId"`
}

// Response to a create minipool signature request
type MinipoolDepositSignatureData struct {
	// The signature for SuperNodeAccount.createMinipool()
	Signature string `json:"signature"`
	Time      int64  `json:"time"`
}

func (c *NodeSetClient) MinipoolDepositSignature(ctx context.Context, address common.Address, salt []byte, chainId *big.Int) (MinipoolDepositSignatureData, error) {
	request := MinipoolDepositSignatureRequest{
		Address: address.Hex(),
		Salt:    utils.EncodeHexWithPrefix(salt),
		ChainId: chainId.String(),
	}
	jsonData, err := json.Marshal(request)
	if err != nil {
		return MinipoolDepositSignatureData{}, fmt.Errorf("error marshalling minipool deposit signature request: %w", err)
	}

	code, response, err := apiv1.SubmitRequest[MinipoolDepositSignatureData](c.NodeSetClient, ctx, true, http.MethodPost, bytes.NewBuffer(jsonData), nil, c.routes.MinipoolDepositSignature)
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
		case MinipoolLimitReachedKey:
			// Address has been given access to Constellation, but cannot create any more minipools.
			return MinipoolDepositSignatureData{}, ErrMinipoolLimitReached

		case MissingExitMessageKey:
			// Address has been given access to Constellation, but the NodeSet service does not have
			// a signed exit message stored for the minipool that user account previously created.
			return MinipoolDepositSignatureData{}, ErrMissingExitMessage
		}

	case http.StatusUnauthorized:
		switch response.Error {
		case UserNotAuthorizedKey:
			// Address not authorized to get minipool deposit signature
			return MinipoolDepositSignatureData{}, ErrNotAuthorized

		case apiv1.InvalidSessionKey:
			// Invalid session
			return MinipoolDepositSignatureData{}, apiv1.ErrInvalidSession
		}
	}

	return MinipoolDepositSignatureData{}, fmt.Errorf("nodeset server responded to minipool deposit signature request with code %d: [%s]", code, response.Message)
}
