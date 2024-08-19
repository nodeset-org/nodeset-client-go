package db

import (
	ethcommon "github.com/ethereum/go-ethereum/common"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

type Validator struct {
	Pubkey              beacon.ValidatorPubkey
	VaultAddress        ethcommon.Address
	DepositData         beacon.ExtendedDepositData
	SignedExit          common.ExitMessage
	ExitMessageUploaded bool
	DepositDataUsed     bool
	MarkedActive        bool
}

func newValidator(depositData beacon.ExtendedDepositData, vaultAddress ethcommon.Address) *Validator {
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

func (v *Validator) SetExitMessage(exitMessage common.ExitMessage) {
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
