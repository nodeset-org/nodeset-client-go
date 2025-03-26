package stakewise

type ValidatorsMetaData struct {
	// validators that the user has for this vault that are active on the Beacon Chain (e.g., pending and active, *not* exited or slashed).
	Active uint `json:"active"`

	// validators that the current user is allowed to have for this vault
	Max uint `json:"max"`

	// number of validators the user is still permitted to create and upload to this vault
	Available uint `json:"available"`
}
