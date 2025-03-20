package v3server_stakewise

import (
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"

	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Handler for api/v3/modules/stakewise/{deployment}/{vault}/validators/meta
func (s *V3StakeWiseServer) handleValidatorsMeta(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.getValidatorsMeta(w, r)
	default:
		common.HandleInvalidMethod(w, s.logger)
	}
}

// Returns information about the requesting user's node account with respect to the number of validators the user has deployed and can deploy on this vault.
// Response: "data": {
//     "active": number,
//     "max": number,
//     "available": number
// }

// GET api/v3/modules/stakewise/{deployment}/{vault}/validators/meta
func (s *V3StakeWiseServer) getValidatorsMeta(w http.ResponseWriter, r *http.Request) {
	// TODO: HN
	if r.Method != http.MethodGet {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	_, pathArgs := common.ProcessApiRequest(s, w, r, nil)
	session := common.ProcessAuthHeader(s, w, r)
	if session == nil {
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

	data := stakewise.VaultsMetaData{
		Active:    deployment.ActiveValidators,
		Max:       deployment.MaxValidators,
		Available: deployment.AvailableValidators,
	}
	common.HandleSuccess(w, s.logger, data)
}
