package admin

import (
	"fmt"
	"net/http"

	"github.com/nodeset-org/nodeset-client-go/server-mock/server/common"
)

// Make a snapshot of the current server state
func (s *AdminServer) snapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		common.HandleInvalidMethod(w, s.logger)
		return
	}

	snapshotName := r.URL.Query().Get("name")
	if snapshotName == "" {
		common.HandleInputError(w, s.logger, fmt.Errorf("missing snapshot name"))
		return
	}
	s.manager.TakeSnapshot(snapshotName)
	common.HandleSuccess(w, s.logger, "")
}
