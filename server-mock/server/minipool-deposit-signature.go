package server

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"
	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/rocket-pool/node-manager-core/utils"
)

func (s *NodeSetMockServer) minipoolDepositSignature(w http.ResponseWriter, r *http.Request) {
	var request struct {
		MinipoolAddress string `json:"minipoolAddress"`
		Salt            string `json:"salt"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		handleInvalidMethod(w, s.logger)
		return
	}

	data := apiv2.MinipoolDepositSignatureData{}

	db := db.NewDatabase(s.logger)
	privateKey, err := crypto.ToECDSA(db.ConstellationAdminPrivateKey)
	if err != nil {
		fmt.Printf("error converting private key: %w", err)
		return
	}

	adminAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	message := []byte(request.MinipoolAddress + request.Salt + adminAddress)

	signature, err := createSignature(message, privateKey)
	if err != nil {
		fmt.Printf("error creating signature: %w", err)
		return
	}
	data.Signature = utils.EncodeHexWithPrefix(signature)

	handleSuccess(w, s.logger, data)

	s.logger.Info("Fetched minipool deposit signature")
}
