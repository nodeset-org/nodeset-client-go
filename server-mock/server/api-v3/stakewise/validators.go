package v3server_stakewise

import (
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	v3stakewise "github.com/nodeset-org/nodeset-client-go/api-v3/stakewise"
	clientcommon "github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Handler for api/v3/modules/stakewise/{deployment}/{vault}/validators
func (s *V3StakeWiseServer) handleValidators(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getValidators(w, r)
	case http.MethodPost:
		s.postValidators(w, r)
	case http.MethodPatch:
		s.patchValidators(w, r)
	default:
		common.HandleInvalidMethod(w, s.logger)
	}
}

// POST api/v3/modules/stakewise/{deployment}/{vault}/validators
func (s *V3StakeWiseServer) postValidators(w http.ResponseWriter, r *http.Request) {
	// TODO: HN
}

// GET api/v3/modules/stakewise/{deployment}/{vault}/validators
func (s *V3StakeWiseServer) getValidators(w http.ResponseWriter, r *http.Request) {
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

	// Find the validator
	validatorStatuses := []stakewise.ValidatorStatus{}
	validators := vault.GetStakeWiseValidatorsForNode(node)
	for _, validator := range validators {
		validatorStatuses = append(validatorStatuses, stakewise.ValidatorStatus{
			Pubkey:              validator.Pubkey,
			Status:              validator.GetStatus(),
			ExitMessageUploaded: validator.ExitMessageUploaded,
		})
	}

	// Write the response
	data := stakewise.ValidatorsData{
		Validators: validatorStatuses,
	}
	common.HandleSuccess(w, s.logger, data)
}

// PATCH api/v3/modules/stakewise/{deployment}/{vault}/validators
func (s *V3StakeWiseServer) patchValidators(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	var body v3stakewise.Validators_PatchBody
	_, pathArgs := common.ProcessApiRequest(s, w, r, &body)
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

	// Handle the upload
	castedExitData := make([]clientcommon.EncryptedExitData, len(body.ExitData))
	for i, data := range body.ExitData {
		castedExitData[i] = clientcommon.EncryptedExitData{
			Pubkey:      data.Pubkey,
			ExitMessage: data.ExitMessage,
		}
	}
	err := vault.HandleEncryptedSignedExitUpload(node, castedExitData)
	if err != nil {
		common.HandleServerError(w, s.logger, err)
		return
	}
	common.HandleSuccess(w, s.logger, struct{}{})
}
