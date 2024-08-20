package stakewise

import "fmt"

const (
	// The vault doesn't correspond to a StakeWise vault recognized by the service
	InvalidVaultKey string = "invalid_vault"
)

var (
	// The vault doesn't correspond to a StakeWise vault recognized by the service
	ErrInvalidVault error = fmt.Errorf("the provided vault doesn't correspond to a StakeWise vault recognized by the service")
)
