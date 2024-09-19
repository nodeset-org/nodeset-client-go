package apiv0

import (
	"context"
	"fmt"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common/core"
	"github.com/rocket-pool/node-manager-core/utils"
)

const (
	// Format for signing node address messages
	NodeAddressMessageFormat string = `{"email":"%s","node_address":"%s"}`
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

// Logs into the NodeSet server, starting a new session
func (c *NodeSetClient) Login(ctx context.Context, nonce string, address ethcommon.Address, signer func([]byte) ([]byte, error)) (core.LoginData, error) {
	// Create the signature
	message := fmt.Sprintf(core.LoginMessageFormat, nonce, address)
	signature, err := signer([]byte(message))
	if err != nil {
		return core.LoginData{}, fmt.Errorf("error signing login message: %w", err)
	}
	return core.Login(c.CommonNodeSetClient, ctx, nonce, address, signature, core.LoginPath)
}

// Registers the node with the NodeSet server. Assumes wallet validation has already been done and the actual wallet address
// is provided here; if it's not, the signature won't come from the node being registered so it will fail validation.
func (c *NodeSetClient) NodeAddress(ctx context.Context, email string, nodeWallet ethcommon.Address, signer func([]byte) ([]byte, error)) error {
	// Create the signature
	message := fmt.Sprintf(NodeAddressMessageFormat, email, nodeWallet)
	signature, err := signer([]byte(message))
	if err != nil {
		return fmt.Errorf("error signing node address message: %w", err)
	}

	// Create the request body
	signatureString := utils.EncodeHexWithPrefix(signature)
	request := NodeAddressRequest{
		Email:       email,
		NodeAddress: nodeWallet.Hex(),
		Signature:   signatureString,
	}

	// Send the request
	return core.NodeAddress(c.CommonNodeSetClient, ctx, email, nodeWallet, signature, core.NodeAddressPath, request)
}

// Get a nonce from the NodeSet server for a new session
func (c *NodeSetClient) Nonce(ctx context.Context) (core.NonceData, error) {
	return core.Nonce(c.CommonNodeSetClient, ctx, core.NoncePath)
}
