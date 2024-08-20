package v2server_stakewise

import (
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// GET api/v2/modules/stakewise/{deployment}/{vault}/deposit-data/meta
func (s *V2StakeWiseServer) depositDataMeta(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	// Get the requesting node
	_, pathArgs := common.ProcessApiRequest(s, w, r, nil)
	session := common.ProcessAuthHeader(s, w, r)
	if session == nil {
		return
	}
	node := common.GetNodeForSession(s, w, session)
	if node == nil {
		return
	}

	// Input validation
	deployment := pathArgs["deployment"]
	if deployment != test.Network { // TEMP
		common.HandleInvalidDeployment(w, s.logger, deployment)
		return
	}
	vaultAddress := ethcommon.HexToAddress(pathArgs["vault"])
	vault := s.manager.GetStakeWiseVault(vaultAddress, deployment)
	if vault == nil {
		common.HandleInvalidVault(w, s.logger, deployment, vaultAddress)
		return
	}

	// Write the response
	data := stakewise.DepositDataMetaData{
		Version: vault.LatestDepositDataSetIndex,
	}
	common.HandleSuccess(w, s.logger, data)
}