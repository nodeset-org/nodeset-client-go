package admin

import (
	"fmt"
	"math/big"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Set the deployment values for the service
func (s *AdminServer) setDeployment(w http.ResponseWriter, r *http.Request) {
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

	// Create a new deployment
	deployment := &db.Deployment{
		DeploymentID:     id,
		WhitelistAddress: whitelist,
		SuperNodeAddress: superNode,
		ChainID:          chainID,
	}
	s.manager.SetDeployment(deployment)
	s.logger.Info("Set deployment info",
		"id", id,
		"whitelist", whitelistString,
		"supernode", superNodeString,
		"chain", chainIDString,
	)
	common.HandleSuccess(w, s.logger, "")
}
