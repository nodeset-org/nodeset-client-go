package v3core

import (
	"context"
	"fmt"
	"log/slog"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common/core"
)

// Logs into the NodeSet server, starting a new session
func (c *V3CoreClient) Login(ctx context.Context, logger *slog.Logger, nonce string, address ethcommon.Address, signer func([]byte) ([]byte, error)) (core.LoginData, error) {
	// Create the signature
	message := fmt.Sprintf(core.LoginMessageFormat, nonce, address)
	signature, err := signer([]byte(message))
	if err != nil {
		return core.LoginData{}, fmt.Errorf("error signing login message: %w", err)
	}
	return core.Login(c.commonClient, ctx, logger, nonce, address, signature, CorePrefix+core.LoginPath)
}
