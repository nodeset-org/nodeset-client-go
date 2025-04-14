package db

import (
	"log/slog"

	"filippo.io/age"
)

// Mock database for storing nodeset.io info
type Database struct {
	Core          *Database_Core
	Eth           *Database_Ethereum
	Constellation *Database_Constellation
	StakeWise     *Database_StakeWise

	// Age identity for the secret key used to encrypt the exit data
	secretEncryptionIdentity *age.X25519Identity

	// Logger
	logger *slog.Logger
}

// Creates a new database
func NewDatabase(logger *slog.Logger) *Database {
	db := &Database{
		logger: logger,
	}
	db.Core = newDatabase_Core(db, logger)
	db.Eth = newDatabase_Ethereum(db, logger)
	db.Constellation = newDatabase_Constellation(db, logger)
	db.StakeWise = newDatabase_StakeWise(db, logger)
	return db
}

// Clones the database
func (d *Database) Clone() *Database {
	dbClone := &Database{
		logger: d.logger,
	}
	dbClone.Core = d.Core.clone(dbClone)
	dbClone.Eth = d.Eth.clone(dbClone)
	dbClone.Constellation = d.Constellation.clone(dbClone)
	dbClone.StakeWise = d.StakeWise.clone(dbClone)
	dbClone.secretEncryptionIdentity = d.secretEncryptionIdentity
	return dbClone
}

// Get the secret encryption identity
func (d *Database) GetSecretEncryptionIdentity() *age.X25519Identity {
	return d.secretEncryptionIdentity
}

// Set the secret encryption identity
func (d *Database) SetSecretEncryptionIdentity(identity *age.X25519Identity) {
	d.secretEncryptionIdentity = identity
}
