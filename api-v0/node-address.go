package apiv0

import (
	"context"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common/core"
)

// Registers the node with the NodeSet server. Assumes wallet validation has already been done and the actual wallet address
// is provided here; if it's not, the signature won't come from the node being registered so it will fail validation.
func (c *NodeSetClient) NodeAddress(ctx context.Context, email string, nodeWallet ethcommon.Address, signature []byte) error {
	return core.NodeAddress(c.CommonNodeSetClient, ctx, email, nodeWallet, signature, core.NodeAddressPath)
}
