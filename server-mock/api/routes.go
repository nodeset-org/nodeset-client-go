package api

const (
	// API routes
	DevPath                   string = "dev"
	DepositDataMetaPath       string = "deposit-data/meta"
	DepositDataPath           string = "deposit-data"
	ValidatorsPath            string = "validators"
	NoncePath                 string = "nonce"
	LoginPath                 string = "login"
	MinipoolPath              string = "minipool"
	RegisterPath              string = "node-address"
	V2StakewiseModulePath     string = "v2/modules/stakewise"
	V2ConstellationModulePath string = "v2/modules/constellation"
	V2CorePath                string = "v2/core"
	AvailablePath             string = "available"
	DepositSignaturePath      string = "deposit-signature"
	WhitelistPath             string = "whitelist"
	// Admin routes
	AdminSnapshotPath      string = "snapshot"
	AdminRevertPath        string = "revert"
	AdminCycleSetPath      string = "cycle-set"
	AdminAddUserPath       string = "add-user"
	AdminWhitelistNodePath string = "whitelist-node"
	AdminRegisterNodePath  string = "register-node"
	AdminAddVaultPath      string = "add-vault"
)
