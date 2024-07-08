package server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
	"github.com/rocket-pool/node-manager-core/utils"
)

func (s *NodeSetMockServer) minipoolDepositSignature(w http.ResponseWriter, r *http.Request) {
	data := apiv2.MinipoolDepositSignatureData{}
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

	// Prep the args
	minipoolAddress := common.HexToAddress(request.MinipoolAddress)
	salt, err := hex.DecodeString(request.Salt)
	if err != nil {
		handleInputError(w, s.logger, fmt.Errorf("error decoding salt: %w", err))
		return
	}

	// Get the signature
	signature, err := s.manager.GetConstellationDepositSignature(minipoolAddress, salt)
	if err != nil {
		fmt.Printf("error creating signature: %w", err)
		return
	}
	data.Signature = utils.EncodeHexWithPrefix(signature)
	s.logger.Info("Fetched minipool deposit signature")
	handleSuccess(w, s.logger, data)
}
