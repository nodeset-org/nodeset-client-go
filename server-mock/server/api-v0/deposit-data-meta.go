package v0server

import (
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// GET api/deposit-data/meta
func (s *V0Server) depositDataMeta(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	// Get the requesting node
	args, _ := common.ProcessApiRequest(s, w, r, nil)
	session := common.ProcessAuthHeader(s, w, r)
	if session == nil {
		return
	}
	node := common.GetNodeForSession(s, w, session)
	if node == nil {
		return
	}

	// Input validation
	db := s.manager.GetDatabase()
	network := args.Get("network")
	deployment := db.StakeWise.GetDeployment(network)
	if deployment == nil {
		common.HandleInvalidDeployment(w, s.logger, network)
		return
	}
	vaultAddress := ethcommon.HexToAddress(args.Get("vault"))
	vault := deployment.GetStakeWiseVault(vaultAddress)
	if vault == nil {
		common.HandleInvalidVault(w, s.logger, network, vaultAddress)
		return
	}

	// Write the response
	data := stakewise.DepositDataMetaData{
		Version: vault.LatestDepositDataSetIndex,
	}
	common.HandleSuccess(w, s.logger, data)
}
