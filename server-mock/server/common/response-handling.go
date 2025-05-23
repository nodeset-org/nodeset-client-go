package common

import (
	"fmt"
	"log/slog"
	"net/http"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/goccy/go-json"
	v2constellation "github.com/nodeset-org/nodeset-client-go/api-v2/constellation"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/nodeset-org/nodeset-client-go/common/core"
	"github.com/rocket-pool/node-manager-core/log"
)

// Handle routes called with an invalid method
func HandleInvalidMethod(w http.ResponseWriter, logger *slog.Logger) {
	writeResponse(w, logger, http.StatusMethodNotAllowed, []byte{})
}

// Handles an error related to parsing the input parameters of a request
func HandleInputError(w http.ResponseWriter, logger *slog.Logger, err error) {
	msg := err.Error()
	bytes := formatError(msg, "")
	writeResponse(w, logger, http.StatusBadRequest, bytes)
}

// Write an error if the auth header couldn't be decoded
func HandleAuthHeaderError(w http.ResponseWriter, logger *slog.Logger, err error) {
	msg := err.Error()
	bytes := formatError(msg, "")
	writeResponse(w, logger, http.StatusUnauthorized, bytes)
}

// Write an error if the auth header is missing
func HandleMissingAuthHeader(w http.ResponseWriter, logger *slog.Logger) {
	msg := "No Authorization header found"
	bytes := formatError(msg, "")
	writeResponse(w, logger, http.StatusUnauthorized, bytes)
}

// Write an error if the session provided in the auth header is not valid
func HandleInvalidSessionError(w http.ResponseWriter, logger *slog.Logger, err error) {
	msg := err.Error()
	bytes := formatError(msg, common.InvalidSessionKey)
	writeResponse(w, logger, http.StatusUnauthorized, bytes)
}

// Write an error if the node providing the request isn't registered
func HandleUnregisteredNode(w http.ResponseWriter, logger *slog.Logger, address ethcommon.Address) {
	msg := fmt.Sprintf("No user found with authorized address %s", address.Hex())
	bytes := formatError(msg, core.UnregisteredAddressKey)
	writeResponse(w, logger, http.StatusUnauthorized, bytes)
}

// Write an error if the node providing the request is already registered
func HandleNodeNotInWhitelist(w http.ResponseWriter, logger *slog.Logger, address ethcommon.Address) {
	msg := fmt.Sprintf("Address %s is not whitelisted", address.Hex())
	bytes := formatError(msg, core.AddressMissingWhitelistKey)
	writeResponse(w, logger, http.StatusBadRequest, bytes)
}

// Write an error if the node providing the request is already registered
func HandleAlreadyRegisteredNode(w http.ResponseWriter, logger *slog.Logger, address ethcommon.Address) {
	msg := fmt.Sprintf("Address %s already registered", address.Hex())
	bytes := formatError(msg, core.AddressAlreadyAuthorizedKey)
	writeResponse(w, logger, http.StatusBadRequest, bytes)
}

// Handles an invalid deployment
func HandleInvalidDeployment(w http.ResponseWriter, logger *slog.Logger, deployment string) {
	msg := fmt.Sprintf("Invalid or unknown deployment: %s", deployment)
	bytes := formatError(msg, common.InvalidDeploymentKey)
	writeResponse(w, logger, http.StatusBadRequest, bytes)
}

// Handles an invalid StakeWise vault
func HandleInvalidVault(w http.ResponseWriter, logger *slog.Logger, deployment string, vault ethcommon.Address) {
	msg := fmt.Sprintf("vault with address [%s] on deployment [%s] not found", vault.Hex(), deployment)
	bytes := formatError(msg, common.InvalidVaultKey)
	writeResponse(w, logger, http.StatusBadRequest, bytes)
}

// Handles a signed exit upload with an already existing message
func HandleExitAlreadyExists(w http.ResponseWriter, logger *slog.Logger) {
	msg := "at least one signed exit message already exists"
	bytes := formatError(msg, v2constellation.ExitMessageExistsKey)
	writeResponse(w, logger, http.StatusBadRequest, bytes)
}

// Write an error if the auth header couldn't be decoded
func HandleServerError(w http.ResponseWriter, logger *slog.Logger, err error) {
	msg := err.Error()
	bytes := formatError(msg, "")
	writeResponse(w, logger, http.StatusInternalServerError, bytes)
}

// The request completed successfully
func HandleSuccess[DataType any](w http.ResponseWriter, logger *slog.Logger, data DataType) {
	response := common.NodeSetResponse[DataType]{
		OK:      true,
		Message: "Success",
		Error:   "",
		Data:    data,
	}

	// Serialize the response
	bytes, err := json.Marshal(response)
	if err != nil {
		HandleServerError(w, logger, fmt.Errorf("error serializing response: %w", err))
	}
	// Write it
	logger.Debug("Response body", slog.String(log.BodyKey, string(bytes)))
	writeResponse(w, logger, http.StatusOK, bytes)
}

// Writes a response to an HTTP request back to the client and logs it
func writeResponse(w http.ResponseWriter, logger *slog.Logger, statusCode int, message []byte) {
	// Prep the log attributes
	codeMsg := fmt.Sprintf("%d %s", statusCode, http.StatusText(statusCode))
	attrs := []any{
		slog.String(log.CodeKey, codeMsg),
	}

	// Log the response
	logMsg := "Responded with:"
	switch statusCode {
	case http.StatusOK:
		logger.Info(logMsg, attrs...)
	case http.StatusInternalServerError:
		logger.Error(logMsg, attrs...)
	default:
		logger.Warn(logMsg, attrs...)
	}

	// Write it to the client
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_, writeErr := w.Write(message)
	if writeErr != nil {
		logger.Error("Error writing response", "error", writeErr)
	}
}

// JSONifies an error for responding to requests
func formatError(message string, errorKey string) []byte {
	msg := common.NodeSetResponse[struct{}]{
		OK:      false,
		Message: message,
		Error:   errorKey,
		Data:    struct{}{},
	}

	bytes, _ := json.Marshal(msg)
	return bytes
}
