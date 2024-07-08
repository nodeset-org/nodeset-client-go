package server

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/crypto"
	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	nsutil "github.com/nodeset-org/nodeset-client-go/utils"
	"github.com/rocket-pool/node-manager-core/utils"
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

	adminAddress := crypto.PubkeyToAddress(db.ConstellationAdminPrivateKey.PublicKey)

	message := append(node.Address[:], adminAddress[:]...)

	signature, err := nsutil.CreateSignature(message, db.ConstellationAdminPrivateKey)
	if err != nil {
		fmt.Printf("error creating signature: %w", err)
		return
	}
	data.Signature = utils.EncodeHexWithPrefix(signature)
	handleSuccess(w, s.logger, data)

	s.logger.Info("Fetched Constellation whitelist")
}
