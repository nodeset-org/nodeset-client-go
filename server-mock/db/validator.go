package db

import (
	"github.com/ethereum/go-ethereum/common"
	apiv1 "github.com/nodeset-org/nodeset-client-go/api-v1"
	"github.com/rocket-pool/node-manager-core/beacon"
)

type Validator struct {
	Pubkey              beacon.ValidatorPubkey
	VaultAddress        common.Address
	DepositData         beacon.ExtendedDepositData
	SignedExit          apiv1.ExitMessage
	ExitMessageUploaded bool
	DepositDataUsed     bool
	MarkedActive        bool
}

func newValidator(depositData beacon.ExtendedDepositData, vaultAddress common.Address) *Validator {
	return &Validator{
		Pubkey:       beacon.ValidatorPubkey(depositData.PublicKey),
		VaultAddress: vaultAddress,
		DepositData:  depositData,
	}
}

func (v *Validator) UseDepositData() {
	v.DepositDataUsed = true
}

func (v *Validator) MarkActive() {
	v.MarkedActive = true
}

func (v *Validator) SetExitMessage(exitMessage apiv1.ExitMessage) {
	// Normally this is where validation would occur
	v.SignedExit = exitMessage
	v.ExitMessageUploaded = true
}

func (v *Validator) Clone() *Validator {
	return &Validator{
		Pubkey:              v.Pubkey,
		VaultAddress:        v.VaultAddress,
		DepositData:         v.DepositData,
		SignedExit:          v.SignedExit,
		ExitMessageUploaded: v.ExitMessageUploaded,
		DepositDataUsed:     v.DepositDataUsed,
		MarkedActive:        v.MarkedActive,
	}
}
