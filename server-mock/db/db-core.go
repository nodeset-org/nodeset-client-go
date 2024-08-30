package db

import (
	"fmt"
	"log/slog"

	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/server-mock/auth"
)

type Database_Core struct {
	// Collection of users
	users []*User

	// Collection of sessions
	Sessions []*Session

	// Internal fields
	logger *slog.Logger
	db     *Database
}

// Create a new core database
func newDatabase_Core(db *Database, logger *slog.Logger) *Database_Core {
	return &Database_Core{
		users:    []*User{},
		Sessions: []*Session{},
		logger:   logger,
		db:       db,
	}
}

// Clone the database
func (d *Database_Core) Clone(dbClone *Database) *Database_Core {
	clone := newDatabase_Core(dbClone, d.logger)
	for _, user := range d.users {
		clone.users = append(clone.users, user.Clone(dbClone))
	}
	for _, session := range d.Sessions {
		clone.Sessions = append(clone.Sessions, session.Clone())
	}
	return clone
}

// =========================
// === Website Emulation ===
// =========================

// Adds a user to the database
func (d *Database_Core) AddUser(email string) (*User, error) {
	for _, user := range d.users {
		if user.Email == email {
			return nil, fmt.Errorf("user with email [%s] already exists", email)
		}
	}

	user := newUser(d.db, email)
	d.users = append(d.users, user)
	return user, nil
}

// Gets a user by their email. Returns nil if not found
func (d *Database_Core) GetUser(email string) *User {
	for _, user := range d.users {
		if user.Email == email {
			return user
		}
	}
	return nil
}

// Gets all users
func (d *Database_Core) GetUsers() []*User {
	return d.users
}

// ============
// === Core ===
// ============

// Get a node by address - returns true if registered, false if not registered and just whitelisted
func (d *Database_Core) GetNode(address ethcommon.Address) (*Node, bool) {
	for _, user := range d.users {
		node := user.GetNode(address)
		if node != nil {
			return node, node.isRegistered
		}
	}
	return nil, false
}

// Creates a new session
func (d *Database_Core) CreateSession() *Session {
	session := newSession()
	d.Sessions = append(d.Sessions, session)
	return session
}

// Gets a session by its nonce
func (d *Database_Core) GetSessionByNonce(nonce string) *Session {
	for _, session := range d.Sessions {
		if session.Nonce == nonce {
			return session
		}
	}
	return nil
}

// Gets a session by its token
func (d *Database_Core) GetSessionByToken(token string) *Session {
	for _, session := range d.Sessions {
		if session.Token == token {
			return session
		}
	}
	return nil
}

// Attempts to log an existing session in with the provided node address and nonce
func (d *Database_Core) Login(nodeAddress ethcommon.Address, nonce string, signature []byte) error {
	return d.loginImpl(nodeAddress, nonce, signature, false)
}

// Attempts to log an existing session in with the provided node address and nonce, skipping the signature verification for testing purposes
func (d *Database_Core) LoginWithoutSignature(nodeAddress ethcommon.Address, nonce string) error {
	return d.loginImpl(nodeAddress, nonce, nil, true)
}

// Implementation for login
func (d *Database_Core) loginImpl(nodeAddress ethcommon.Address, nonce string, signature []byte, skipVerification bool) error {
	// Get the session
	session := d.GetSessionByNonce(nonce)
	if session == nil {
		return fmt.Errorf("no session with provided nonce")
	}

	if session.IsLoggedIn {
		return fmt.Errorf("session already logged in")
	}

	// Verify the signature
	if !skipVerification {
		err := auth.VerifyLoginSignature(nonce, nodeAddress, signature)
		if err != nil {
			return err
		}
	}

	// Find the user account for the node
	for _, user := range d.users {
		node := user.GetNode(nodeAddress)
		if node == nil {
			continue
		}
		if node.isRegistered {
			session.login(nodeAddress)
			return nil
		}
	}

	return ErrUnregisteredNode
}
