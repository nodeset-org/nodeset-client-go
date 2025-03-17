package v3core

import (
	"context"
	"log/slog"

	"github.com/nodeset-org/nodeset-client-go/common/core"
)

// Get a nonce from the NodeSet server for a new session
func (c *V3CoreClient) Nonce(ctx context.Context, logger *slog.Logger) (core.NonceData, error) {
	return core.Nonce(c.commonClient, ctx, logger, CorePrefix+core.NoncePath)
}
