package core

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common"
)

const (
	// Route for registering a node address with the NodeSet server
	NodeAddressPath string = "node-address"

	// The node address has already been confirmed on a NodeSet account
	AddressAlreadyAuthorizedKey string = "address_already_authorized"

	// The node address hasn't been whitelisted on the provided NodeSet account
	AddressMissingWhitelistKey string = "address_missing_whitelist"
)

var (
	// The node address has already been confirmed on a NodeSet account
	ErrAlreadyRegistered error = errors.New("node has already been registered with the NodeSet server")

	// The node address hasn't been whitelisted on the provided NodeSet account
	ErrNotWhitelisted error = errors.New("node address hasn't been whitelisted on the provided NodeSet account")
)

// Request to register a node with the NodeSet server
type NodeAddressRequest struct {
	// The email address of the NodeSet account
	Email string `json:"email"`

	// The node's wallet address
	NodeAddress string `json:"node_address"`

	// Signature of the request
	Signature string `json:"signature"` // Must be 0x-prefixed hex encoded
}

// Registers the node with the NodeSet server. Assumes wallet validation has already been done and the actual wallet address
// is provided here; if it's not, the signature won't come from the node being registered so it will fail validation.
func NodeAddress(c *common.CommonNodeSetClient, ctx context.Context, email string, nodeWallet ethcommon.Address, signature []byte, nodeAddressPath string, request any) error {
	// Create the request body
	jsonData, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("error marshalling registration request: %w", err)
	}

	// Submit the request
	code, response, err := common.SubmitRequest[struct{}](c, ctx, false, http.MethodPost, bytes.NewBuffer(jsonData), nil, nodeAddressPath)
	if err != nil {
		return fmt.Errorf("error registering node: %w", err)
	}

	// Handle response based on return code
	switch code {
	case http.StatusOK:
		// Node successfully registered
		return nil

	case http.StatusBadRequest:
		switch response.Error {
		case AddressAlreadyAuthorizedKey:
			// Already registered
			return ErrAlreadyRegistered

		case AddressMissingWhitelistKey:
			// Not whitelisted in the user account
			return ErrNotWhitelisted

		case common.InvalidSignatureKey:
			// Invalid signature
			return common.ErrInvalidSignature

		case common.MalformedInputKey:
			// Malformed input
			return common.ErrMalformedInput
		}
	}
	return fmt.Errorf("nodeset server responded to node-address request with code %d: [%s]", code, response.Message)
}
