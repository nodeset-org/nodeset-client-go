package apiv0

import "errors"

const (
	// The provided network was invalid
	InvalidNetworkKey string = "invalid_network"
)

var (
	// The provided network was invalid
	ErrInvalidNetwork error = errors.New("the provided network was invalid")
)
