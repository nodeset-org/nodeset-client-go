package db

import (
	"github.com/nodeset-org/nodeset-client-go/common"
	"github.com/rocket-pool/node-manager-core/beacon"
)

// Info about a validator that's part of a Constellation deployment
type ConstellationValidatorInfo struct {
	Pubkey beacon.ValidatorPubkey

	exitMessage *common.ExitMessage
}

// Create a new Constellation validator info
func newConstellationValidatorInfo(pubkey beacon.ValidatorPubkey) *ConstellationValidatorInfo {
	return &ConstellationValidatorInfo{
		Pubkey: pubkey,
	}
}

// Clone the StakeWise validator info
func (v *ConstellationValidatorInfo) clone() *ConstellationValidatorInfo {
	return &ConstellationValidatorInfo{
		Pubkey: v.Pubkey,
		exitMessage: &common.ExitMessage{
			Signature: v.exitMessage.Signature,
			Message: common.ExitMessageDetails{
				Epoch:          v.exitMessage.Message.Epoch,
				ValidatorIndex: v.exitMessage.Message.ValidatorIndex,
			},
		},
	}
}

// Get the exit message for the validator
func (v *ConstellationValidatorInfo) GetExitMessage() *common.ExitMessage {
	return v.exitMessage
}

// Set the exit message for the validator
func (v *ConstellationValidatorInfo) SetExitMessage(exitMessage *common.ExitMessage) {
	v.exitMessage = exitMessage
}
