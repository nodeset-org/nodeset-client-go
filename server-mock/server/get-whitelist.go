package server

import (
	"fmt"
	"net/http"

	"github.com/rocket-pool/node-manager-core/utils"

	"github.com/ethereum/go-ethereum/crypto"
	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
)

func (s *NodeSetMockServer) getWhitelist(w http.ResponseWriter, r *http.Request) {
	data := apiv2.WhitelistData{}

	// Get the requesting node
	session := s.processAuthHeader(w, r)
	if session == nil {
		return
	}
	node := s.getNodeForSession(w, session)
	if node == nil {
		return
	}

	db := db.NewDatabase(s.logger)
	privateKey, err := crypto.ToECDSA(db.ConstellationAdminPrivateKey)
	if err != nil {
		fmt.Printf("error converting private key: %w", err)
		return
	}

	adminAddress := crypto.PubkeyToAddress(privateKey.PublicKey).Hex()
	message := []byte(node.Address.Hex() + adminAddress)
	signature, err := createSignature(message, privateKey)
	if err != nil {
		fmt.Printf("error creating signature: %w", err)
		return
	}
	data.Signature = utils.EncodeHexWithPrefix(signature)
	handleSuccess(w, s.logger, data)

	s.logger.Info("Fetched Constellation whitelist")
}
