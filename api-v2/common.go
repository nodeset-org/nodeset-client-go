package apiv2

import "errors"

const (
	// The user is not authorized
	UserNotAuthorizedKey string = "user_not_authorized"

	// The node address cannot create more minipools
	MinipoolLimitReachedKey string = "minipool_limit_reached"

	// The node address is missing the exit message
	MissingExitMessageKey string = "missing_exit_message"
)

var (
	// The user is not authorize to whitelist for Constellation
	ErrNotAuthorized error = errors.New("user account owning the requesting node is not on the internal NodeSet service whitelist for Constellation")

	// The node address cannot create more minipools
	ErrMinipoolLimitReached error = errors.New("node address cannot create more minipools")

	// The node address is missing the exit message for latest minipool
	ErrMissingExitMessage error = errors.New("node address is missing the exit message for latest minipool")
)
