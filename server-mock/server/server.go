package server

import (
	"context"
	"crypto/ecdsa"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"sync"

	"github.com/ethereum/go-ethereum/accounts"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gorilla/mux"
	apiv1 "github.com/nodeset-org/nodeset-client-go/api-v1"
	apiv2 "github.com/nodeset-org/nodeset-client-go/api-v2"
	"github.com/nodeset-org/nodeset-client-go/server-mock/api"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
	"github.com/nodeset-org/nodeset-client-go/server-mock/db"
	"github.com/nodeset-org/nodeset-client-go/server-mock/manager"
	"github.com/rocket-pool/node-manager-core/log"
)

type NodeSetMockServer struct {
	logger  *slog.Logger
	ip      string
	port    uint16
	socket  net.Listener
	server  http.Server
	router  *mux.Router
	manager *manager.NodeSetMockManager
}

func NewNodeSetMockServer(logger *slog.Logger, ip string, port uint16) (*NodeSetMockServer, error) {
	// Create the router
	router := mux.NewRouter()

	// Create the manager
	server := &NodeSetMockServer{
		logger: logger,
		ip:     ip,
		port:   port,
		router: router,
		server: http.Server{
			Handler: router,
		},
		manager: manager.NewNodeSetMockManager(logger),
	}

	// Register each route
	apiRouter := router.PathPrefix("/api").Subrouter()
	server.registerApiRoutes(apiRouter)
	adminRouter := router.PathPrefix("/admin").Subrouter()
	server.registerAdminRoutes(adminRouter)
	return server, nil
}

// Starts listening for incoming HTTP requests
func (s *NodeSetMockServer) Start(wg *sync.WaitGroup) error {
	// Create the socket
	socket, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.ip, s.port))
	if err != nil {
		return fmt.Errorf("error creating socket: %w", err)
	}
	s.socket = socket

	// Get the port if random
	if s.port == 0 {
		s.port = uint16(socket.Addr().(*net.TCPAddr).Port)
	}

	// Start listening
	wg.Add(1)
	go func() {
		err := s.server.Serve(socket)
		if !errors.Is(err, http.ErrServerClosed) {
			s.logger.Error("error while listening for HTTP requests", log.Err(err))
		}
		wg.Done()
	}()

	return nil
}

// Stops the HTTP listener
func (s *NodeSetMockServer) Stop() error {
	err := s.server.Shutdown(context.Background())
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("error stopping listener: %w", err)
	}
	return nil
}

// Get the port the server is listening on
func (s *NodeSetMockServer) GetPort() uint16 {
	return s.port
}

// Get the mock manager for direct access
func (s *NodeSetMockServer) GetManager() *manager.NodeSetMockManager {
	return s.manager
}

// API routes
func (s *NodeSetMockServer) registerApiRoutes(apiRouter *mux.Router) {
	// deposit-data/meta
	depositDataMeta := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.depositDataMeta(w, r)
		default:
			handleInvalidMethod(w, s.logger)
		}
	}
	// v1
	apiRouter.HandleFunc("/"+apiv1.DepositDataMetaPath, depositDataMeta)
	apiRouter.HandleFunc("/"+api.DevPath+apiv1.DepositDataMetaPath, depositDataMeta)
	// v2
	apiRouter.HandleFunc("/"+apiv2.StakeWisePath+apiv1.DepositDataMetaPath, depositDataMeta)

	// deposit-data
	depositData := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.getDepositData(w, r)
		case http.MethodPost:
			s.uploadDepositData(w, r)
		default:
			handleInvalidMethod(w, s.logger)
		}
	}
	// v1
	apiRouter.HandleFunc("/"+apiv1.DepositDataPath, depositData)
	apiRouter.HandleFunc("/"+api.DevPath+apiv1.DepositDataPath, depositData)
	// v2
	apiRouter.HandleFunc("/"+apiv2.StakeWisePath+apiv1.DepositDataPath, depositData)

	// validators
	validators := func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			s.getValidators(w, r)
		case http.MethodPatch:
			s.uploadSignedExits(w, r)
		default:
			handleInvalidMethod(w, s.logger)
		}
	}
	// v1
	apiRouter.HandleFunc("/"+apiv1.ValidatorsPath, validators)
	apiRouter.HandleFunc("/"+api.DevPath+apiv1.ValidatorsPath, validators)
	// v2
	apiRouter.HandleFunc("/"+apiv2.StakeWisePath+apiv1.ValidatorsPath, validators)

	// node-address
	// v1
	apiRouter.HandleFunc("/"+apiv1.NodeAddressPath, s.registerNode)
	apiRouter.HandleFunc("/"+api.DevPath+apiv1.NodeAddressPath, s.registerNode)
	// v2
	apiRouter.HandleFunc("/"+apiv2.CorePath+apiv1.NodeAddressPath, s.registerNode)

	// nonce
	// v1
	apiRouter.HandleFunc("/"+apiv1.NoncePath, s.getNonce)
	apiRouter.HandleFunc("/"+api.DevPath+apiv1.NoncePath, s.getNonce)
	// v2
	apiRouter.HandleFunc("/"+apiv2.CorePath+apiv1.NoncePath, s.getNonce)

	// login
	// v1
	apiRouter.HandleFunc("/"+apiv1.LoginPath, s.login)
	apiRouter.HandleFunc("/"+api.DevPath+apiv1.LoginPath, s.login)
	// v2
	apiRouter.HandleFunc("/"+apiv2.CorePath+apiv1.LoginPath, s.login)

	// constellation
	// v1 - constellation introduced in later version
	// v2
	apiRouter.HandleFunc("/"+apiv2.ConstellationPath+apiv2.WhitelistPath, s.getWhitelist)
	apiRouter.HandleFunc("/"+apiv2.ConstellationPath+apiv2.MinipoolAvailablePath, s.getMinipoolAvailable)
	apiRouter.HandleFunc("/"+apiv2.ConstellationPath+apiv2.MinipoolDepositSignaturePath, s.minipoolDepositSignature)

}

// Admin routes
func (s *NodeSetMockServer) registerAdminRoutes(adminRouter *mux.Router) {
	adminRouter.HandleFunc("/"+api.AdminSnapshotPath, s.snapshot)
	adminRouter.HandleFunc("/"+api.AdminRevertPath, s.revert)
	adminRouter.HandleFunc("/"+api.AdminCycleSetPath, s.cycleSet)
	adminRouter.HandleFunc("/"+api.AdminAddUserPath, s.addUser)
	adminRouter.HandleFunc("/"+api.AdminWhitelistNodePath, s.whitelistNode)
	adminRouter.HandleFunc("/"+api.AdminAddVaultPath, s.addStakeWiseVault)
}

// =============
// === Utils ===
// =============

func (s *NodeSetMockServer) processApiRequest(w http.ResponseWriter, r *http.Request, requestBody any) url.Values {
	args := r.URL.Query()
	s.logger.Info("New request", slog.String(log.MethodKey, r.Method), slog.String(log.PathKey, r.URL.Path))
	s.logger.Debug("Request params:", slog.String(log.QueryKey, r.URL.RawQuery))

	if requestBody != nil {
		// Read the body
		bodyBytes, err := io.ReadAll(r.Body)
		if err != nil {
			handleInputError(w, s.logger, fmt.Errorf("error reading request body: %w", err))
			return nil
		}
		s.logger.Debug("Request body:", slog.String(log.BodyKey, string(bodyBytes)))

		// Deserialize the body
		err = json.Unmarshal(bodyBytes, &requestBody)
		if err != nil {
			handleInputError(w, s.logger, fmt.Errorf("error deserializing request body: %w", err))
			return nil
		}
	}

	return args
}

func (s *NodeSetMockServer) processAuthHeader(w http.ResponseWriter, r *http.Request) *db.Session {
	// Get the auth header
	session, err := s.manager.VerifyRequest(r)
	if err != nil {
		if errors.Is(err, manager.ErrInvalidSession) {
			handleInvalidSessionError(w, s.logger, err)
			return nil
		}
		if errors.Is(err, auth.ErrAuthHeader) {
			handleAuthHeaderError(w, s.logger, err)
			return nil
		}
		if errors.Is(err, auth.ErrMissingAuthHeader) {
			handleMissingAuthHeader(w, s.logger)
			return nil
		}

		// Catch-all
		handleServerError(w, s.logger, err)
		return nil
	}

	return session
}

func (s *NodeSetMockServer) getNodeForSession(w http.ResponseWriter, session *db.Session) *db.Node {
	// Get the node
	node, isRegistered := s.manager.GetNode(session.NodeAddress)
	if node == nil || !isRegistered {
		handleUnregisteredNode(w, s.logger, session.NodeAddress)
		return nil
	}

	// Make sure it's logged in
	if !session.IsLoggedIn {
		handleInvalidSessionError(w, s.logger, fmt.Errorf("session is not logged in"))
		return nil
	}
	return node
}

// https://github.com/nodeset-org/nodeset-svc-mock/blob/4f7aeb08967ec1234ba510db24ec26977851a6ee/auth/auth.go#L117
func createSignature(message []byte, privateKey *ecdsa.PrivateKey) ([]byte, error) {
	messageHash := accounts.TextHash(message)
	signature, err := crypto.Sign(messageHash, privateKey)
	if err != nil {
		return nil, fmt.Errorf("error signing message: %w", err)
	}

	// Fix the ECDSA 'v' (see https://medium.com/mycrypto/the-magic-of-digital-signatures-on-ethereum-98fe184dc9c7#:~:text=The%20version%20number,2%E2%80%9D%20was%20introduced)
	signature[crypto.RecoveryIDOffset] += 27
	return signature, nil
}
