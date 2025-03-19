package v3server_stakewise

import (
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

type DepositData_PostBody struct {
	Validators []ExtendedDepositData `json:"validators"`
}

// Extended deposit data beyond what is required in an actual deposit message to Beacon, emulating what the deposit CLI produces
type ExtendedDepositData struct {
	PublicKey             beacon.ByteArray `json:"pubkey"`
	WithdrawalCredentials beacon.ByteArray `json:"withdrawalCredentials"`
	Amount                uint64           `json:"amount"`
	Signature             beacon.ByteArray `json:"signature"`
	DepositMessageRoot    beacon.ByteArray `json:"depositMessageRoot"`
	DepositDataRoot       beacon.ByteArray `json:"depositDataRoot"`
	ForkVersion           beacon.ByteArray `json:"forkVersion"`
	NetworkName           string           `json:"networkName"`
}

// Handler for api/v3/modules/stakewise/{deployment}/{vault}/deposit-data
func (s *V3StakeWiseServer) handleDepositData(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getDepositData(w, r)
	case http.MethodPost:
		s.uploadDepositData(w, r)
	default:
		common.HandleInvalidMethod(w, s.logger)
	}
}

// GET api/v3/modules/stakewise/{deployment}/{vault}/deposit-data
func (s *V3StakeWiseServer) getDepositData(w http.ResponseWriter, r *http.Request) {
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
	db := s.manager.GetDatabase()
	deploymentID := pathArgs["deployment"]
	deployment := db.StakeWise.GetDeployment(deploymentID)
	if deployment == nil {
		common.HandleInvalidDeployment(w, s.logger, deploymentID)
		return
	}
	vaultAddress := ethcommon.HexToAddress(pathArgs["vault"])
	vault := deployment.GetVault(vaultAddress)
	if vault == nil {
		common.HandleInvalidVault(w, s.logger, deploymentID, vaultAddress)
		return
	}

	// Write the data
	data := stakewise.DepositDataData{
		Version:     vault.LatestDepositDataSetIndex,
		DepositData: make([]beacon.ExtendedDepositData, len(vault.LatestDepositDataSet)),
	}
	for i, deposit := range vault.LatestDepositDataSet {
		data.DepositData[i] = beacon.ExtendedDepositData(deposit)
	}

	common.HandleSuccess(w, s.logger, data)
}

// POST api/v3/modules/stakewise/{deployment}/{vault}/deposit-data
func (s *V3StakeWiseServer) uploadDepositData(w http.ResponseWriter, r *http.Request) {
	// Get the params
	var body DepositData_PostBody
	_, pathArgs := common.ProcessApiRequest(s, w, r, &body)
	session := common.ProcessAuthHeader(s, w, r)
	if session == nil {
		return
	}

	// Get the requesting node
	node := common.GetNodeForSession(s, w, session)
	if node == nil {
		return
	}

	// Input validation
	db := s.manager.GetDatabase()
	deploymentID := pathArgs["deployment"]
	deployment := db.StakeWise.GetDeployment(deploymentID)
	if deployment == nil {
		common.HandleInvalidDeployment(w, s.logger, deploymentID)
		return
	}
	vaultAddress := ethcommon.HexToAddress(pathArgs["vault"])
	vault := deployment.GetVault(vaultAddress)
	if vault == nil {
		common.HandleInvalidVault(w, s.logger, deploymentID, vaultAddress)
		return
	}

	// Handle the request
	castedDepositData := make([]beacon.ExtendedDepositData, len(body.Validators))
	for i, deposit := range body.Validators {
		castedDepositData[i] = beacon.ExtendedDepositData(deposit)
	}
	err := vault.HandleDepositDataUpload(node, castedDepositData)
	if err != nil {
		common.HandleServerError(w, s.logger, err)
		return
	}
	common.HandleSuccess(w, s.logger, struct{}{})
}
