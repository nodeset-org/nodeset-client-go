package admin

import (
	"fmt"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Whitelist a new node with a user account
func (s *AdminServer) whitelistNode(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	// Input validation
	query := r.URL.Query()
	email := query.Get("email")
	if email == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing email query parameter"))
		return
	}
	addressString := query.Get("address")
	if addressString == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing address query parameter"))
		return
	}
	address := ethcommon.HexToAddress(addressString)

	// Whitelist the node
	err := s.manager.WhitelistNodeAccount(email, address)
	if err != nil {
		common.HandleServerError(w, s.logger, err)
		return
	}
	s.logger.Info("Whitelisted new node account", "email", email, "address", address.Hex())
	common.HandleSuccess(w, s.logger, "")
}
