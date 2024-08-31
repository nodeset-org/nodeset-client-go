package admin

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Add a StakeWise deployment to the service
func (s *AdminServer) addStakeWiseDeployment(w http.ResponseWriter, r *http.Request) {
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

	// Add a new deployment
	db := s.manager.GetDatabase()
	db.StakeWise.AddDeployment(id, chainID)
	s.logger.Info("Added StakeWise deployment",
		"id", id,
		"chain", chainIDString,
	)
	common.HandleSuccess(w, s.logger, "")
}
