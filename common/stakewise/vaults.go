package stakewise

type ValidatorsMetaData struct {
	// Validators that the user has registered for this vault.
	// This includes validators that:
	// - Are active on the Beacon Chain (e.g., pending and active, *not* exited or slashed)
	// - Are included in deposit events on the Beacon deposit contract
	// - Have already had a previous submission signed using the current Beacon deposit root according to the Beacon deposit contract's get_deposit_root() function
	Registered uint `json:"registered"`

	// The maximum number of active validators that the current user is allowed to have for this vault
	Max uint `json:"max"`

	// The number of validators the user is still permitted to create and upload to this vault, according to the registered rules above
	Available uint `json:"available"`
}
