package apiv2

import "errors"

const (
	// Value of the auth response header if the login token has expired
	invalidSessionKey string = "invalid_session"

	// The signature provided can't be verified
	invalidSignatureKey string = "invalid_signature"

	// The request didn't have the correct fields or the fields were malformed
	malformedInputKey string = "malformed_input"

	// The provided network was invalid
	invalidNetworkKey string = "invalid_network"

	// The user is not authorized
	userNotAuthorizedKey string = "user_not_authorized"

	// The node address cannot create more minipools
	minipoolLimitReachedKey string = "minipool_limit_reached"

	// The node address is missing the exit message
	missingExitMessageKey string = "missing_exit_message"
)

var (
	// The session token is invalid, probably expired
	ErrInvalidSession error = errors.New("session token is invalid")

	// The provided signature could not be verified
	ErrInvalidSignature error = errors.New("the provided signature could not be verified")

	// The request didn't have the correct fields or the fields were malformed
	ErrMalformedInput error = errors.New("the request didn't have the correct fields or the fields were malformed")

	// The provided network was invalid
	ErrInvalidNetwork error = errors.New("the provided network was invalid")

	// The user is not authorize to whitelist for Constellation
	ErrNotAuthorized error = errors.New("user account owning the requesting node is not on the internal NodeSet service whitelist for Constellation")

	// The node address cannot create more minipools
	ErrMinipoolLimitReached error = errors.New("node address cannot create more minipools")

	// The node address is missing the exit message for latest minipool
	ErrMissingExitMessage error = errors.New("node address is missing the exit message for latest minipool")
)
