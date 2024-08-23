package db

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

type Node struct {
	Address    common.Address
	Validators map[string][]*Validator
}

func newNode(address common.Address) *Node {
	return &Node{
		Address:    address,
		Validators: map[string][]*Validator{},
	}
}

func (n *Node) AddDepositData(depositData beacon.ExtendedDepositData, deployment string, vaultAddress common.Address) {
	validatorsForDeployment, exists := n.Validators[deployment]
	if !exists {
		validatorsForDeployment = []*Validator{}
		n.Validators[deployment] = validatorsForDeployment
	}

	pubkey := beacon.ValidatorPubkey(depositData.PublicKey)
	for _, validator := range validatorsForDeployment {
		if validator.Pubkey == pubkey {
			// Already present
			return
		}
	}

	validator := newValidator(depositData, vaultAddress)
	validatorsForDeployment = append(validatorsForDeployment, validator)
	n.Validators[deployment] = validatorsForDeployment
}

func (n *Node) Clone() *Node {
	clone := newNode(n.Address)
	for deployment, validatorsForDeployment := range n.Validators {
		cloneSlice := make([]*Validator, len(validatorsForDeployment))
		for i, validator := range validatorsForDeployment {
			cloneSlice[i] = validator.Clone()
		}
		clone.Validators[deployment] = cloneSlice
	}
	return clone
}
