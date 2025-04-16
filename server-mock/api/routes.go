package api

const (
	// API routes
	DevPath string = "dev/"

	// Admin routes
	AdminAddConstellationDeploymentPath       string = "add-constellation-deployment"
	AdminAddStakeWiseDeploymentPath           string = "add-stakewise-deployment"
	AdminAddStakeWiseVaultPath                string = "add-stakewise-vault"
	AdminSnapshotPath                         string = "snapshot"
	AdminRevertPath                           string = "revert"
	AdminCycleSetPath                         string = "cycle-set"
	AdminAddUserPath                          string = "add-user"
	AdminWhitelistNodePath                    string = "whitelist-node"
	AdminRegisterNodePath                     string = "register-node"
	AdminAddVaultPath                         string = "add-vault"
	AdminSetConstellationPrivateKeyPath       string = "constellation/private-key"
	AdminIncrementWhitelistNoncePath          string = "constellation/increment-whitelist-nonce"
	AdminIncrementSuperNodeNoncePath          string = "constellation/increment-supernode-nonce"
	AdminSetEncryptionKeyPath                 string = "set-encryption-key"
	AdminConstellationSetValidatorForMinipool string = "constellation/set-validator-for-minipool"
)
