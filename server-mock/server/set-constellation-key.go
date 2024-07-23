package server

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/nodeset-org/nodeset-client-go/server-mock/api"
)

func (s *NodeSetMockServer) setConstellationAdminPrivateKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		handleInvalidMethod(w, s.logger)
		return
	}

	// Get the login request
	var request api.AdminSetConstellationPrivateKeyRequest
	_ = s.processApiRequest(w, r, &request)
	session := s.processAuthHeader(w, r)
	if session == nil {
		return
	}

	// Decode the key
	privateKey, err := crypto.HexToECDSA(request.PrivateKey)
	if err != nil {
		handleInputError(w, s.logger, fmt.Errorf("invalid private key"))
		return
	}

	// Set the key
	s.manager.SetConstellationAdminPrivateKey(privateKey)
	pubkey := common.BytesToAddress(crypto.FromECDSAPub(&privateKey.PublicKey))
	s.logger.Info("Set Constellation private key", "address", pubkey.Hex())
	handleSuccess(w, s.logger, "")
}
