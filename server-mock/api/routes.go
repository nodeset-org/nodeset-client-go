package api

const (
	// API routes
	DevPath               string = "dev"
	DepositDataMetaPath   string = "deposit-data/meta"
	DepositDataPath       string = "deposit-data"
	ValidatorsPath        string = "validators"
	NoncePath             string = "nonce"
	LoginPath             string = "login"
	RegisterPath          string = "node-address"
	V2StakewiseModulePath string = "v2/modules/stakewise"
	V2CorePath            string = "v2/core"

	// Admin routes
	AdminSnapshotPath      string = "snapshot"
	AdminRevertPath        string = "revert"
	AdminCycleSetPath      string = "cycle-set"
	AdminAddUserPath       string = "add-user"
	AdminWhitelistNodePath string = "whitelist-node"
	AdminRegisterNodePath  string = "register-node"
	AdminAddVaultPath      string = "add-vault"
)
