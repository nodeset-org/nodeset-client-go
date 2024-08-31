package api

type AdminSetConstellationPrivateKeyRequest struct {
	// ID of the deployment to set the private key for
	Deployment string `json:"deploymentID"`

	// Private key in 0x-prefixed hex format
	PrivateKey string `json:"privateKey"`
}

type AdminSetManualSignatureTimestampRequest struct {
	// Unix timestamp in seconds
	Timestamp int64 `json:"timestamp"`
}

type AdminSetAvailableConstellationMinipoolCountRequest struct {
	// Number of available minipools
	Count int `json:"count"`

	// User email address
	UserEmail string `json:"user"`
}
