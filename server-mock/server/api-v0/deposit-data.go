package v0server

import (
	"fmt"
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
	network := args.Get("network")
	vaultAddress := ethcommon.HexToAddress(args.Get("vault"))
	vault := s.manager.GetStakeWiseVault(network, vaultAddress)
	if vault == nil {
		common.HandleInputError(w, s.logger, fmt.Errorf("vault with address [%s] on network [%s] not found", vaultAddress.Hex(), network))
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

	// Handle the upload
	network := depositData[0].NetworkName
	vaultAddress := ethcommon.BytesToAddress(depositData[0].WithdrawalCredentials)
	err := s.manager.HandleDepositDataUpload(node.Address, network, vaultAddress, depositData)
	if err != nil {
		common.HandleServerError(w, s.logger, err)
		return
	}
	common.HandleSuccess(w, s.logger, struct{}{})
}
