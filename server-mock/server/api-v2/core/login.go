package v2server_core

import (
	"errors"
	"fmt"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common/core"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
	"github.com/rocket-pool/node-manager-core/utils"
)

// POST api/v2/core/login
func (s *V2CoreServer) login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	// Get the login request
	var request core.LoginRequest
	_, _ = common.ProcessApiRequest(s, w, r, &request)
	session := common.ProcessAuthHeader(s, w, r)
	if session == nil {
		return
	}

	// Log it in
	address := ethcommon.HexToAddress(request.Address)
	signature, err := utils.DecodeHex(request.Signature)
	if err != nil {
		common.HandleInputError(w, s.logger, fmt.Errorf("invalid signature"))
		return
	}
	database := s.manager.GetDatabase()
	err = database.Core.Login(address, request.Nonce, signature)
	if err != nil {
		if errors.Is(err, db.ErrUnregisteredNode) {
			common.HandleUnregisteredNode(w, s.logger, address)
			return
		}
		common.HandleServerError(w, s.logger, err)
		return
	}

	// Respond
	data := core.LoginData{
		Token: session.Token,
	}
	common.HandleSuccess(w, s.logger, data)
	s.logger.Info("Logged into session", "nonce", request.Nonce, "address", address.Hex())
}
