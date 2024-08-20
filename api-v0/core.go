package apiv0

import (
	"context"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common/core"
)

// Logs into the NodeSet server, starting a new session
func (c *NodeSetClient) Login(ctx context.Context, nonce string, address ethcommon.Address, signature []byte) (core.LoginData, error) {
	return core.Login(c.CommonNodeSetClient, ctx, nonce, address, signature, core.LoginPath)
}

// Registers the node with the NodeSet server. Assumes wallet validation has already been done and the actual wallet address
// is provided here; if it's not, the signature won't come from the node being registered so it will fail validation.
func (c *NodeSetClient) NodeAddress(ctx context.Context, email string, nodeWallet ethcommon.Address, signature []byte) error {
	return core.NodeAddress(c.CommonNodeSetClient, ctx, email, nodeWallet, signature, core.NodeAddressPath)
}

// Get a nonce from the NodeSet server for a new session
func (c *NodeSetClient) Nonce(ctx context.Context) (core.NonceData, error) {
	return core.Nonce(c.CommonNodeSetClient, ctx, core.NoncePath)
}
