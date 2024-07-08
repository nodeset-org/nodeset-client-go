package server

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"
	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	nsutil "github.com/nodeset-org/nodeset-client-go/utils"
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

	adminAddress := crypto.PubkeyToAddress(db.ConstellationAdminPrivateKey.PublicKey).Hex()

	minipoolAddressBytes, err := hex.DecodeString(request.MinipoolAddress)
	if err != nil {
		log.Fatal(err)
		return
	}

	saltBytes, err := hex.DecodeString(request.Salt)
	if err != nil {
		log.Fatal(err)
		return
	}

	adminAddressBytes, err := hex.DecodeString(adminAddress)
	if err != nil {
		log.Fatal(err)
		return
	}

	message := append(minipoolAddressBytes, saltBytes...)
	message = append(message, adminAddressBytes...)

	signature, err := nsutil.CreateSignature(message, db.ConstellationAdminPrivateKey)
	if err != nil {
		fmt.Printf("error creating signature: %w", err)
		return
	}
	data.Signature = utils.EncodeHexWithPrefix(signature)

	handleSuccess(w, s.logger, data)

	s.logger.Info("Fetched minipool deposit signature")
}
