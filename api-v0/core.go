package apiv0

import (
	"context"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common/core"
	"github.com/rocket-pool/node-manager-core/utils"
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
func (c *NodeSetClient) Login(ctx context.Context, nonce string, address ethcommon.Address, signature []byte) (core.LoginData, error) {
	return core.Login(c.CommonNodeSetClient, ctx, nonce, address, signature, core.LoginPath)
}

// Registers the node with the NodeSet server. Assumes wallet validation has already been done and the actual wallet address
// is provided here; if it's not, the signature won't come from the node being registered so it will fail validation.
func (c *NodeSetClient) NodeAddress(ctx context.Context, email string, nodeWallet ethcommon.Address, signature []byte) error {
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
