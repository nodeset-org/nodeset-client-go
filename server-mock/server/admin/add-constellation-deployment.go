package admin

import (
	"fmt"
	"math/big"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Add a Constellation deployment to the service
func (s *AdminServer) addConstellationDeployment(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	// Input validation
	query := r.URL.Query()
	id := query.Get("id")
	if id == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing id query parameter"))
		return
	}
	chainIDString := query.Get("chain")
	if chainIDString == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing chain query parameter"))
		return
	}
	chainID, success := new(big.Int).SetString(chainIDString, 10)
	if !success {
		common.HandleInputError(w, s.logger, fmt.Errorf("invalid chain id"))
		return
	}
	whitelistString := query.Get("whitelist")
	if whitelistString == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing whitelist query parameter"))
		return
	}
	whitelist := ethcommon.HexToAddress(whitelistString)
	superNodeString := query.Get("supernode")
	if superNodeString == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing supernode query parameter"))
		return
	}
	superNode := ethcommon.HexToAddress(superNodeString)

	// Add a new deployment
	db := s.manager.GetDatabase()
	db.Constellation.AddDeployment(id, chainID, whitelist, superNode)
	s.logger.Info("Added Constellation deployment",
		"id", id,
		"chain", chainIDString,
		"whitelist", whitelistString,
		"supernode", superNodeString,
	)
	common.HandleSuccess(w, s.logger, "")
}
