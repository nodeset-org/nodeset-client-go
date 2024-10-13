package admin

import (
	"fmt"
	"net/http"

	"filippo.io/age"
	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Set the Constellation admin private key
func (s *AdminServer) setNodeSetEncryptionKey(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	// Input validation
	query := r.URL.Query()
	key := query.Get("key")
	if key == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing key query parameter"))
		return
	}

	id, err := age.ParseX25519Identity(key)
	if err != nil {
		common.HandleInputError(w, s.logger, fmt.Errorf("invalid key"))
		return
	}

	// Set the key
	db := s.manager.GetDatabase()
	db.SetSecretEncryptionIdentity(id)
	pubkey := id.Recipient().String()
	s.logger.Info("Set nodeset encryption key", "pubkey", pubkey)
	common.HandleSuccess(w, s.logger, "")
}
