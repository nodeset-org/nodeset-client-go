package admin

import (
	"fmt"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
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
	_ = common.ProcessApiRequest(s, w, r, &request)

	// Decode the key
	privateKey, err := crypto.HexToECDSA(request.PrivateKey)
	if err != nil {
		common.HandleInputError(w, s.logger, fmt.Errorf("invalid private key"))
		return
	}

	// Set the key
	s.manager.SetConstellationAdminPrivateKey(privateKey)
	pubkey := ethcommon.BytesToAddress(crypto.FromECDSAPub(&privateKey.PublicKey))
	s.logger.Info("Set Constellation private key", "address", pubkey.Hex())
	common.HandleSuccess(w, s.logger, "")
}
