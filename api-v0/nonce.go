package apiv0

import (
	"context"

	"github.com/nodeset-org/nodeset-client-go/common/core"
)

// Get a nonce from the NodeSet server for a new session
func (c *NodeSetClient) Nonce(ctx context.Context) (core.NonceData, error) {
	return core.Nonce(c.CommonNodeSetClient, ctx, core.NoncePath)
}
