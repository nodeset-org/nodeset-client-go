package server

import (
	"fmt"
	"math/big"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
	"github.com/rocket-pool/node-manager-core/utils"
)

func (s *NodeSetMockServer) getWhitelistSignature(w http.ResponseWriter, r *http.Request) {
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

	query := r.URL.Query()
	chainId := query.Get("chainId")
	chainIdBig, success := new(big.Int).SetString(chainId, 10)
	if !success {
		handleInputError(w, s.logger, fmt.Errorf("invalid chainId"))
		return
	}
	whitelistAddressString := query.Get("whitelistAddress")
	if whitelistAddressString == "" {
		handleInputError(w, s.logger, fmt.Errorf("missing whitelistAddress"))
		return
	}
	whitelistAddress := common.HexToAddress(whitelistAddressString)

	// Get the signature
	time, signature, err := s.manager.GetConstellationWhitelistSignatureAndTime(node.Address, chainIdBig, whitelistAddress)
	if err != nil {
		handleServerError(w, s.logger, fmt.Errorf("error creating signature: %w", err))
		return
	}
	data.Signature = utils.EncodeHexWithPrefix(signature)
	data.Time = time.Unix()
	s.logger.Info("Fetched Constellation whitelist")
	handleSuccess(w, s.logger, data)
}
