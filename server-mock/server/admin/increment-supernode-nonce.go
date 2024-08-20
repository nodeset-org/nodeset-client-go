package admin

import (
	"fmt"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Increment the SuperNodeAccount nonce for a node
func (s *AdminServer) incrementSuperNodeNonce(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	// Input validation
	query := r.URL.Query()
	addressString := query.Get("address")
	if addressString == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing address query parameter"))
		return
	}
	address := ethcommon.HexToAddress(addressString)

	// Whitelist the node
	s.manager.IncrementSuperNodeNonce(address)
	s.logger.Info("SuperNode nonce incremented", "address", address.Hex())
	common.HandleSuccess(w, s.logger, "")
}
