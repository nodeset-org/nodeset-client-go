package db

import (
	apiv0 "github.com/nodeset-org/nodeset-client-go/api-v0"
	v2stakewise "github.com/nodeset-org/nodeset-client-go/api-v2/stakewise"
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

// Info about a validator that's part of a StakeWise vault
type StakeWiseValidatorInfo struct {
	Pubkey              beacon.ValidatorPubkey
	DepositData         beacon.ExtendedDepositData
	SignedExit          common.ExitMessage
	ExitMessageUploaded bool
	DepositDataUsed     bool
	MarkedActive        bool
}

// Create a new StakeWise validator info
func newStakeWiseValidatorInfo(depositData beacon.ExtendedDepositData) *StakeWiseValidatorInfo {
	return &StakeWiseValidatorInfo{
		Pubkey:      beacon.ValidatorPubkey(depositData.PublicKey),
		DepositData: depositData,
	}
}

// Clone the StakeWise validator info
func (v *StakeWiseValidatorInfo) clone() *StakeWiseValidatorInfo {
	return &StakeWiseValidatorInfo{
		Pubkey:              v.Pubkey,
		DepositData:         v.DepositData,
		SignedExit:          v.SignedExit,
		ExitMessageUploaded: v.ExitMessageUploaded,
		DepositDataUsed:     v.DepositDataUsed,
		MarkedActive:        v.MarkedActive,
	}
}

// Mark the deposit data as used
func (v *StakeWiseValidatorInfo) UseDepositData() {
	v.DepositDataUsed = true
}

// Mark the validator as active
func (v *StakeWiseValidatorInfo) MarkActive() {
	v.MarkedActive = true
}

// Set the signed exit message for the validator
func (v *StakeWiseValidatorInfo) SetExitMessage(exitMessage common.ExitMessage) {
	// Normally this is where validation would occur
	v.SignedExit = exitMessage
	v.ExitMessageUploaded = true
}

func (v *StakeWiseValidatorInfo) GetStatusV0() apiv0.StakeWiseStatus {
	if v.MarkedActive {
		return apiv0.StakeWiseStatus_Registered
	}
	if v.DepositDataUsed {
		return apiv0.StakeWiseStatus_Uploaded
	}
	return apiv0.StakeWiseStatus_Pending
}

func (v *StakeWiseValidatorInfo) GetStatusV2() v2stakewise.StakeWiseStatus {
	if v.MarkedActive {
		return v2stakewise.StakeWiseStatus_Registered
	}
	if v.DepositDataUsed {
		return v2stakewise.StakeWiseStatus_Uploaded
	}
	return v2stakewise.StakeWiseStatus_Pending
}
