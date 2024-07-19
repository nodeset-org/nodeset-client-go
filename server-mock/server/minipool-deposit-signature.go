package server

import (
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"

	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
	"github.com/rocket-pool/node-manager-core/utils"
)

func (s *NodeSetMockServer) minipoolDepositSignature(w http.ResponseWriter, r *http.Request) {
	data := apiv2.MinipoolDepositSignatureData{}
	request := apiv2.MinipoolDepositSignatureRequest{}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid request payload", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		handleInvalidMethod(w, s.logger)
		return
	}

	// Prep the args
	salt, success := new(big.Int).SetString(request.Salt, 10)
	if !success {
		handleInputError(w, s.logger, fmt.Errorf("error decoding salt"))
		return
	}
	chainIdBig, success := new(big.Int).SetString(request.ChainId, 10)
	if !success {
		handleInputError(w, s.logger, fmt.Errorf("invalid chainId"))
		return
	}

	// Get the signature
	time, signature, err := s.manager.GetConstellationDepositSignatureAndTime(request.MinipoolAddress, salt, chainIdBig)
	if err != nil {
		fmt.Printf("error creating signature: %w", err)
		return
	}
	data.Signature = utils.EncodeHexWithPrefix(signature)
	data.Time = time.Unix()
	s.logger.Info("Fetched minipool deposit signature")
	handleSuccess(w, s.logger, data)
}
