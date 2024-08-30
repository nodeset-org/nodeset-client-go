package db

import (
	"errors"

	"github.com/ethereum/go-ethereum/common"
)

var (
	ErrNotWhitelisted error = errors.New("node address hasn't been whitelisted on the provided NodeSet account")
)

type User struct {
	Email string

	nodes map[common.Address]*Node
	db    *Database
}

func newUser(db *Database, email string) *User {
	return &User{
		Email: email,
		nodes: map[common.Address]*Node{},
		db:    db,
	}
}

func (u *User) Clone(dbClone *Database) *User {
	userClone := newUser(dbClone, u.Email)
	for address, node := range u.nodes {
		userClone.nodes[address] = node.Clone(userClone)
	}
	return userClone
}

func (u *User) WhitelistNode(nodeAddress common.Address) *Node {
	node := u.nodes[nodeAddress]
	if node == nil {
		node = newNode(u, nodeAddress)
		u.nodes[nodeAddress] = node
	}
	return node
}

func (u *User) GetNode(nodeAddress common.Address) *Node {
	return u.nodes[nodeAddress]
}

func (u *User) GetNodes() map[common.Address]*Node {
	return u.nodes
}
