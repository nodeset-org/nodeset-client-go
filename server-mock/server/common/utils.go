package common

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"

	"github.com/gorilla/mux"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/rocket-pool/node-manager-core/log"
)

// ==============
// === Errors ===
// ==============

var (
	ErrInvalidSession error = errors.New("session token is invalid")
)

// Logs the request and returns the query args and path args
func ProcessApiRequest(serverImpl IServerImpl, w http.ResponseWriter, r *http.Request, requestBody any) (url.Values, map[string]string) {
	args := r.URL.Query()
	logger := serverImpl.GetLogger()
	logger.Info("New request", slog.String(log.MethodKey, r.Method), slog.String(log.PathKey, r.URL.Path))
	logger.Debug("Request params:", slog.String(log.QueryKey, r.URL.RawQuery))

	if requestBody != nil {
		// Read the body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			HandleInputError(w, logger, fmt.Errorf("error reading request body: %w", err))
			return nil, nil
		}
		logger.Debug("Request body:", slog.String(log.BodyKey, string(bodyBytes)))

		// Deserialize the body
		err = json.Unmarshal(bodyBytes, &requestBody)
		if err != nil {
			HandleInputError(w, logger, fmt.Errorf("error deserializing request body: %w", err))
			return nil, nil
		}
	}

	return args, mux.Vars(r)
}

// Makes sure the request has a valid auth header and returns the session it belongs to
func ProcessAuthHeader(serverImpl IServerImpl, w http.ResponseWriter, r *http.Request) *db.Session {
	// Get the auth header
	mgr := serverImpl.GetManager()
	logger := serverImpl.GetLogger()

	// Get the session token
	token, err := auth.GetSessionTokenFromRequest(r)
	if err != nil {
		if errors.Is(err, auth.ErrAuthHeader) {
			HandleAuthHeaderError(w, logger, err)
			return nil
		}
		if errors.Is(err, auth.ErrMissingAuthHeader) {
			HandleMissingAuthHeader(w, logger)
			return nil
		}

		// Catch-all
		HandleServerError(w, logger, err)
		return nil
	}

	// Get the session
	db := mgr.GetDatabase()
	session := db.Core.GetSessionByToken(token)
	if session == nil {
		HandleInvalidSessionError(w, logger, err)
		return nil
	}
	return session
}

// Gets the node for the session, making sure it's registered and logged in
func GetNodeForSession(serverImpl IServerImpl, w http.ResponseWriter, session *db.Session) *db.Node {
	// Get the node
	mgr := serverImpl.GetManager()
	db := mgr.GetDatabase()
	logger := serverImpl.GetLogger()
	node, isRegistered := db.Core.GetNode(session.NodeAddress)
	if node == nil || !isRegistered {
		HandleUnregisteredNode(w, logger, session.NodeAddress)
		return nil
	}

	// Make sure it's logged in
	if !session.IsLoggedIn {
		HandleInvalidSessionError(w, logger, fmt.Errorf("session is not logged in"))
		return nil
	}
	return node
}
