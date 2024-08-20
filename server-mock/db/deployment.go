package db

import (
	"math/big"

	"github.com/ethereum/go-ethereum/common"
)

type Deployment struct {
	DeploymentID     string
	WhitelistAddress common.Address
	SuperNodeAddress common.Address
	ChainID          *big.Int
}

func (d *Deployment) Clone() *Deployment {
	return &Deployment{
		DeploymentID:     d.DeploymentID,
		WhitelistAddress: d.WhitelistAddress,
		SuperNodeAddress: d.SuperNodeAddress,
		ChainID:          new(big.Int).Set(d.ChainID),
	}
}
