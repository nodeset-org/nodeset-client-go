package v0server

import (
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

// Handler for api/deposit-data
func (s *V0Server) handleDepositData(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getDepositData(w, r)
	case http.MethodPost:
		s.uploadDepositData(w, r)
	default:
		common.HandleInvalidMethod(w, s.logger)
	}
}

// GET api/deposit-data
func (s *V0Server) getDepositData(w http.ResponseWriter, r *http.Request) {
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
	vault := deployment.GetVault(vaultAddress)
	if vault == nil {
		common.HandleInvalidVault(w, s.logger, network, vaultAddress)
		return
	}

	// Write the data
	data := stakewise.DepositDataData{
		Version:     vault.LatestDepositDataSetIndex,
		DepositData: vault.LatestDepositDataSet,
	}
	common.HandleSuccess(w, s.logger, data)
}

// POST api/deposit-data
func (s *V0Server) uploadDepositData(w http.ResponseWriter, r *http.Request) {
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

	// Input validation
	db := s.manager.GetDatabase()
	network := depositData[0].NetworkName
	deployment := db.StakeWise.GetDeployment(network)
	if deployment == nil {
		common.HandleInvalidDeployment(w, s.logger, network)
		return
	}
	vaultAddress := ethcommon.BytesToAddress(depositData[0].WithdrawalCredentials)
	vault := deployment.GetVault(vaultAddress)
	if vault == nil {
		common.HandleInvalidVault(w, s.logger, network, vaultAddress)
		return
	}

	// Handle the upload
	err := vault.HandleDepositDataUpload(node, depositData)
	if err != nil {
		common.HandleServerError(w, s.logger, err)
		return
	}
	common.HandleSuccess(w, s.logger, struct{}{})
}
