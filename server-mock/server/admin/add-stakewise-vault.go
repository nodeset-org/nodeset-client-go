package admin

import (
	"fmt"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Add a StakeWise vault to the server
func (s *AdminServer) addStakeWiseVault(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	// Input validation
	query := r.URL.Query()
	network := query.Get("network")
	if network == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing network query parameter"))
		return
	}
	addressString := query.Get("address")
	if addressString == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing address query parameter"))
		return
	}
	address := ethcommon.HexToAddress(addressString)

	// Create a new deposit data set
	err := s.manager.AddStakeWiseVault(address, network)
	if err != nil {
		common.HandleServerError(w, s.logger, err)
		return
	}
	s.logger.Info("Added new stakewise vault", "address", address.Hex(), "network", network)
	common.HandleSuccess(w, s.logger, "")
}
