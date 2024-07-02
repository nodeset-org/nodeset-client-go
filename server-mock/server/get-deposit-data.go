package server

import (
	"fmt"
	"net/http"

	"github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/server-mock/api"
)

func (s *NodeSetMockServer) getDepositData(w http.ResponseWriter, r *http.Request) {
	// Get the requesting node
	args := s.processApiRequest(w, r, nil)
	session := s.processAuthHeader(w, r)
	if session == nil {
		return
	}
	node := s.getNodeForSession(w, session)
	if node == nil {
		return
	}

	// Input validation
	network := args.Get("network")
	vaultAddress := common.HexToAddress(args.Get("vault"))
	vault := s.manager.GetStakeWiseVault(vaultAddress, network)
	if vault == nil {
		handleInputError(w, s.logger, fmt.Errorf("vault with address [%s] on network [%s] not found", vaultAddress.Hex(), network))
		return
	}

	// Write the data
	data := api.DepositDataData{
		Version:     vault.LatestDepositDataSetIndex,
		DepositData: vault.LatestDepositDataSet,
	}
	handleSuccess(w, s.logger, data)
}
