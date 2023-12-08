package types

import (
	"fmt"
	math "math"

	"github.com/babylonchain/babylon/btcstaking"
	bbn "github.com/babylonchain/babylon/types"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// ensure that these message types implement the sdk.Msg interface
var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgCreateBTCValidator{}
	_ sdk.Msg = &MsgCreateBTCDelegation{}
	_ sdk.Msg = &MsgAddCovenantSigs{}
	_ sdk.Msg = &MsgBTCUndelegate{}
)

func (m *MsgCreateBTCValidator) ValidateBasic() error {
	if m.Commission == nil {
		return fmt.Errorf("empty commission")
	}
	if m.Description == nil {
		return fmt.Errorf("empty description")
	}
	if _, err := m.Description.EnsureLength(); err != nil {
		return err
	}
	if m.BabylonPk == nil {
		return fmt.Errorf("empty Babylon public key")
	}
	if m.BtcPk == nil {
		return fmt.Errorf("empty BTC public key")
	}
	if m.Pop == nil {
		return fmt.Errorf("empty proof of possession")
	}
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		return err
	}
	if err := m.Pop.ValidateBasic(); err != nil {
		return err
	}

	return nil
}

func (m *MsgCreateBTCDelegation) ValidateBasic() error {
	if m.BabylonPk == nil {
		return fmt.Errorf("empty Babylon public key")
	}
	if m.Pop == nil {
		return fmt.Errorf("empty proof of possession")
	}
	if m.BtcPk == nil {
		return fmt.Errorf("empty delegator BTC public key")
	}
	if m.StakingTx == nil {
		return fmt.Errorf("empty staking tx info")
	}
	if m.SlashingTx == nil {
		return fmt.Errorf("empty slashing tx")
	}
	if m.DelegatorSlashingSig == nil {
		return fmt.Errorf("empty delegator signature")
	}
	if _, err := sdk.AccAddressFromBech32(m.Signer); err != nil {
		return err
	}

	// Check staking time is at most uint16
	if m.StakingTime > math.MaxUint16 {
		return ErrInvalidStakingTx.Wrapf("invalid lock time: %d, max: %d", m.StakingTime, math.MaxUint16)
	}
	// Ensure list of validator BTC PKs is not empty
	if len(m.ValBtcPkList) == 0 {
		return ErrEmptyValidatorList
	}
	// Ensure list of validator BTC PKs is not duplicated
	if ExistsDup(m.ValBtcPkList) {
		return ErrDuplicatedValidator
	}

	// staking tx should be correctly formatted
	if err := m.StakingTx.ValidateBasic(); err != nil {
		return err
	}
	if err := m.Pop.ValidateBasic(); err != nil {
		return err
	}

	// verifications about on-demand unbonding
	if m.UnbondingTx == nil {
		return fmt.Errorf("empty unbonding tx")
	}
	if m.UnbondingSlashingTx == nil {
		return fmt.Errorf("empty slashing tx")
	}
	if m.DelegatorUnbondingSlashingSig == nil {
		return fmt.Errorf("empty delegator signature")
	}
	unbondingTxMsg, err := bbn.NewBTCTxFromBytes(m.UnbondingTx)
	if err != nil {
		return err
	}
	if err := btcstaking.IsSimpleTransfer(unbondingTxMsg); err != nil {
		return err
	}

	// Check unbonding time is lower than max uint16
	if uint64(m.UnbondingTime) > math.MaxUint16 {
		return ErrInvalidUnbondingTx.Wrapf("unbonding time %d must be lower than %d", m.UnbondingTime, math.MaxUint16)
	}

	return nil
}

func (m *MsgAddCovenantSigs) ValidateBasic() error {
	if m.Pk == nil {
		return fmt.Errorf("empty BTC covenant public key")
	}
	if m.SlashingTxSigs == nil {
		return fmt.Errorf("empty covenant signatures on slashing tx")
	}
	if len(m.StakingTxHash) != chainhash.MaxHashStringSize {
		return fmt.Errorf("staking tx hash is not %d", chainhash.MaxHashStringSize)
	}

	// verifications about on-demand unbonding
	if m.UnbondingTxSig == nil {
		return fmt.Errorf("empty covenant signature")
	}
	if m.SlashingUnbondingTxSigs == nil {
		return fmt.Errorf("empty covenant signature")
	}

	return nil
}

func (m *MsgBTCUndelegate) ValidateBasic() error {
	if len(m.StakingTxHash) != chainhash.MaxHashStringSize {
		return fmt.Errorf("staking tx hash is not %d", chainhash.MaxHashStringSize)
	}

	if m.UnbondingTxSig == nil {
		return fmt.Errorf("empty signature from the delegator")
	}

	return nil
}
