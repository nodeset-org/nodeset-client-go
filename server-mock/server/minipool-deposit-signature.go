package server

import (
	"encoding/json"
	"fmt"
	"math/big"
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
	salt, success := new(big.Int).SetString(request.Salt, 10)
	if !success {
		handleInputError(w, s.logger, fmt.Errorf("error decoding salt"))
		return
	}
	query := r.URL.Query()
	chainId := query.Get("chainId")
	chainIdBig, success := new(big.Int).SetString(chainId, 10)
	if !success {
		handleInputError(w, s.logger, fmt.Errorf("invalid chainId"))
		return
	}

	// Get the signature
	time, signature, err := s.manager.GetConstellationDepositSignatureAndTime(minipoolAddress, salt, chainIdBig)
	if err != nil {
		fmt.Printf("error creating signature: %w", err)
		return
	}
	data.Signature = utils.EncodeHexWithPrefix(signature)
	data.Time = time.Unix()
	s.logger.Info("Fetched minipool deposit signature")
	handleSuccess(w, s.logger, data)
}
