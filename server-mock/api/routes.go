package api

const (
	// API routes
	DevPath string = "dev/"

	// Admin routes
	AdminSetDeploymentPath              string = "set-deployment"
	AdminSnapshotPath                   string = "snapshot"
	AdminRevertPath                     string = "revert"
	AdminCycleSetPath                   string = "cycle-set"
	AdminAddUserPath                    string = "add-user"
	AdminWhitelistNodePath              string = "whitelist-node"
	AdminRegisterNodePath               string = "register-node"
	AdminAddVaultPath                   string = "add-vault"
	AdminSetConstellationPrivateKeyPath string = "constellation/private-key"
	AdminIncrementWhitelistNoncePath    string = "constellation/increment-whitelist-nonce"
	AdminIncrementSuperNodeNoncePath    string = "constellation/increment-supernode-nonce"
)
