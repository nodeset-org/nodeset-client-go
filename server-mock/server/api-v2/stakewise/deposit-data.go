package v2server_stakewise

import (
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
	"github.com/nodeset-org/nodeset-client-go/server-mock/internal/test"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

// Handler for api/v2/modules/stakewise/{deployment}/{vault}/deposit-data
func (s *V2StakeWiseServer) handleDepositData(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getDepositData(w, r)
	case http.MethodPost:
		s.uploadDepositData(w, r)
	default:
		common.HandleInvalidMethod(w, s.logger)
	}
}

// GET api/v2/modules/stakewise/{deployment}/{vault}/deposit-data
func (s *V2StakeWiseServer) getDepositData(w http.ResponseWriter, r *http.Request) {
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

	// Write the data
	data := stakewise.DepositDataData{
		Version:     vault.LatestDepositDataSetIndex,
		DepositData: vault.LatestDepositDataSet,
	}
	common.HandleSuccess(w, s.logger, data)
}

// POST api/v2/modules/stakewise/{deployment}/{vault}/deposit-data
func (s *V2StakeWiseServer) uploadDepositData(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	var depositData []beacon.ExtendedDepositData
	_, _ = common.ProcessApiRequest(s, w, r, &depositData)
	session := common.ProcessAuthHeader(s, w, r)
	if session == nil {
		return
	}
	node := common.GetNodeForSession(s, w, session)
	if node == nil {
		return
	}

	// Handle the upload
	err := s.manager.HandleDepositDataUpload(node.Address, depositData)
	if err != nil {
		common.HandleServerError(w, s.logger, err)
		return
	}
	common.HandleSuccess(w, s.logger, struct{}{})
}
