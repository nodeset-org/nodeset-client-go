package v2constellation

import "errors"

const (
	// The user is not authorized
	UserNotAuthorizedKey string = "user_not_authorized"

	// The node address cannot create more minipools
	MinipoolLimitReachedKey string = "minipool_limit_reached"

	// The requesting node's user is missing an exit message for one of its validators
	ValidatorRequiresExitMessageKey string = "validator_requires_exit_message"

	// The node making the request isn't the user's whitelisted Constellation node
	IncorrectNodeAddressKey string = "incorrect_node_address"

	// The requesting node's owner doesn't have a node whitelisted for Constellation yet
	MissingWhitelistedNodeAddressKey string = "missing_whitelisted_node_address"
)

var (
	// The user is not authorize to whitelist for Constellation
	ErrNotAuthorized error = errors.New("user account owning the requesting node is not on the internal NodeSet service whitelist for Constellation")

	// The node address cannot create more minipools
	ErrMinipoolLimitReached error = errors.New("node address cannot create more minipools")

	// The node address is missing the exit message for latest minipool
	ErrValidatorRequiresExitMessage error = errors.New("requesting node's user is missing an exit message for one of their minipools")

	// The node making the request isn't the user's whitelisted Constellation node
	ErrIncorrectNodeAddress error = errors.New("requester isn't the user's whitelisted Constellation node")

	// The requesting node's owner doesn't have a node whitelisted for Constellation yet
	ErrMissingWhitelistedNodeAddress error = errors.New("requesting node's owner doesn't have a node whitelisted for Constellation yet")
)
