package admin

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nodeset-org/nodeset-client-go/server-mock/api"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Set the Constellation admin private key
func (s *AdminServer) setConstellationAdminPrivateKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	// Get the login request
	var request api.AdminSetConstellationPrivateKeyRequest
	_, _ = common.ProcessApiRequest(s, w, r, &request)

	// Decode the key
	privateKey, err := crypto.HexToECDSA(request.PrivateKey)
	if err != nil {
		common.HandleInputError(w, s.logger, fmt.Errorf("invalid private key"))
		return
	}

	// Set the key
	db := s.manager.GetDatabase()
	deployment := db.Constellation.GetDeployment(request.Deployment)
	if deployment == nil {
		common.HandleInvalidDeployment(w, s.logger, request.Deployment)
		return
	}
	deployment.SetAdminPrivateKey(privateKey)
	pubkey := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	s.logger.Info("Set Constellation private key", "address", pubkey)
	common.HandleSuccess(w, s.logger, "")
}
