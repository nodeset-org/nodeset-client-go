package v0server

import (
	"fmt"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common/stakewise"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// GET api/deposit-data/meta
func (s *V0Server) depositDataMeta(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	// Get the requesting node
	args := common.ProcessApiRequest(s, w, r, nil)
	session := common.ProcessAuthHeader(s, w, r)
	if session == nil {
		return
	}
	node := common.GetNodeForSession(s, w, session)
	if node == nil {
		return
	}

	// Input validation
	network := args.Get("network")
	vaultAddress := ethcommon.HexToAddress(args.Get("vault"))
	vault := s.manager.GetStakeWiseVault(vaultAddress, network)
	if vault == nil {
		common.HandleInputError(w, s.logger, fmt.Errorf("vault with address [%s] on network [%s] not found", vaultAddress.Hex(), network))
		return
	}

	// Write the response
	data := stakewise.DepositDataMetaData{
		Version: vault.LatestDepositDataSetIndex,
	}
	common.HandleSuccess(w, s.logger, data)
}
