package server

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"
	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/rocket-pool/node-manager-core/utils"
)

func (s *NodeSetMockServer) minipoolDepositSignature(w http.ResponseWriter, r *http.Request) {
	data := apiv2.MinipoolDepositSignatureData{}

	db := db.NewDatabase(s.logger)
	privateKey, err := crypto.ToECDSA(db.ConstellationAdminPrivateKey)
	if err != nil {
		fmt.Printf("error converting private key: %w", err)
		return
	}
	message := []byte("TODO: minipoolDepositSignature messages")
	signature, err := createSignature(message, privateKey)
	if err != nil {
		fmt.Printf("error creating signature: %w", err)
		return
	}
	data.Signature = utils.EncodeHexWithPrefix(signature)

	handleSuccess(w, s.logger, data)

	s.logger.Info("Fetched minipool deposit signature")
}
