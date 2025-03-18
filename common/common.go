package common

import (
	"errors"
)

// Routes
const (
	// Path for the deployments route
	DeploymentsPath string = "deployments"
)

// Error keys
const (
	// Value of the auth response header if the login token has expired
	InvalidSessionKey string = "invalid_session"

	// The signature provided can't be verified
	InvalidSignatureKey string = "invalid_signature"

	// The request didn't have the correct fields or the fields were malformed
	MalformedInputKey string = "malformed_input"

	// The provided deployment doesn't correspond to a deployment recognized by the service
	InvalidDeploymentKey string = "invalid_deployment"

	// The requester doesn't own the provided validator
	InvalidValidatorOwnerKey string = "invalid_validator_owner"

	// The exit message provided was invalid
	InvalidExitMessageKey string = "invalid_exit_message"

	// The user doesn't have permission to do this
	InvalidPermissionsKey string = "invalid_permissions"

	// The vault doesn't correspond to a StakeWise vault recognized by the service
	InvalidVaultKey string = "invalid_vault"

	// The vault does not have enough ETH deposits in it to support the number of validators being registered.
	InsufficientVaultBalanceKey string = "insufficient_vault_balance"
)

// Errors
var (
	// The session token is invalid, probably expired
	ErrInvalidSession error = errors.New("session token is invalid")

	// The provided signature could not be verified
	ErrInvalidSignature error = errors.New("the provided signature could not be verified")

	// The request didn't have the correct fields or the fields were malformed
	ErrMalformedInput error = errors.New("the request didn't have the correct fields or the fields were malformed")

	// The provided deployment doesn't correspond to a deployment recognized by the service
	ErrInvalidDeployment error = errors.New("the provided deployment doesn't correspond to a deployment recognized by the service")

	// The requester doesn't own the provided validator
	ErrInvalidValidatorOwner error = errors.New("this node doesn't own one of the provided validators")

	// The exit message provided was invalid
	ErrInvalidExitMessage error = errors.New("the provided exit message was invalid")

	// The user doesn't have permission to do this
	ErrInvalidPermissions error = errors.New("the user doesn't have permission from the nodeset.io service to do this yet")

	// The vault doesn't correspond to a StakeWise vault recognized by the service
	ErrInvalidVault error = errors.New("the provided vault doesn't correspond to a StakeWise vault recognized by the service")

	// The vault does not have enough ETH deposits in it to support the number of validators being registered.
	ErrInsufficientVaultBalance error = errors.New("the vault does not have enough ETH deposits in it to support the number of validators being registered")
)

// =======================
// === Deployment Data ===
// =======================

// A deployment of the service
type Deployment struct {
	// The Ethereum chain ID of the deployment
	ChainID string `json:"chainId"`

	// The name of the deployment, used as a key in subsequent requests
	Name string `json:"name"`
}

// Standard response data for a list of service deployments
type DeploymentsData struct {
	Deployments []Deployment `json:"deployments"`
}

// =================
// === Exit Data ===
// =================

// Details of an exit message
type ExitMessageDetails struct {
	Epoch          string `json:"epoch"`
	ValidatorIndex string `json:"validator_index"`
}

// Voluntary exit message
type ExitMessage struct {
	Message   ExitMessageDetails `json:"message"`
	Signature string             `json:"signature"`
}

// Data for a pubkey's voluntary exit message
type ExitData struct {
	Pubkey      string      `json:"pubkey"`
	ExitMessage ExitMessage `json:"exit_message"`
}

// Data for a pubkey's voluntary exit message
type EncryptedExitData struct {
	Pubkey      string `json:"pubkey"`
	ExitMessage string `json:"exit_message"`
}
