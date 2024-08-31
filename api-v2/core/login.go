package v2core

import (
	"context"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common/core"
)

// Logs into the NodeSet server, starting a new session
func (c *V2CoreClient) Login(ctx context.Context, nonce string, address ethcommon.Address, signature []byte) (core.LoginData, error) {
	return core.Login(c.commonClient, ctx, nonce, address, signature, CorePrefix+core.LoginPath)
}
