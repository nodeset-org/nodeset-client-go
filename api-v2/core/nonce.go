package v2core

import (
	"context"

	"github.com/nodeset-org/nodeset-client-go/common/core"
)

// Get a nonce from the NodeSet server for a new session
func (c *V2CoreClient) Nonce(ctx context.Context) (core.NonceData, error) {
	return core.Nonce(c.commonClient, ctx, CorePrefix+core.NoncePath)
}
