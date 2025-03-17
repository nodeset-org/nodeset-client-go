package v3constellation

import "errors"

const (
	// The node address cannot create more minipools
	MinipoolLimitReachedKey string = "minipool_limit_reached"

	// The node making the request isn't the user's whitelisted Constellation node
	IncorrectNodeAddressKey string = "incorrect_node_address"

	// The requesting node's owner doesn't have a node whitelisted for Constellation yet
	MissingWhitelistedNodeAddressKey string = "missing_whitelisted_node_address"

	// Nodeset.io is missing a signed exit message for a previous minipool
	MissingExitMessageKey string = "missing_exit_message"

	// A minipool with this address already exists
	AddressAlreadyRegisteredKey string = "address_already_registered"
)

var (
	// The node address cannot create more minipools
	ErrMinipoolLimitReached error = errors.New("node address cannot create more minipools")

	// The node making the request isn't the user's whitelisted Constellation node
	ErrIncorrectNodeAddress error = errors.New("requester isn't the user's whitelisted Constellation node")

	// The requesting node's owner doesn't have a node whitelisted for Constellation yet
	ErrMissingWhitelistedNodeAddress error = errors.New("requesting node's owner doesn't have a node whitelisted for Constellation yet")

	// Nodeset.io is missing a signed exit message for a previous minipool
	ErrMissingExitMessage error = errors.New("nodeset.io is missing a signed exit message for a previous minipool")

	// A minipool with this address already exists
	ErrAddressAlreadyRegistered error = errors.New("a minipool with this address already exists")
)
