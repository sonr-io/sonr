package types

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/address"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
)

var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgRecordsWrite{}
	_ sdk.Msg = &MsgRecordsDelete{}
	_ sdk.Msg = &MsgProtocolsConfigure{}
	_ sdk.Msg = &MsgPermissionsGrant{}
	_ sdk.Msg = &MsgPermissionsRevoke{}
	_ sdk.Msg = &MsgRotateVaultKeys{}
)

// NewMsgUpdateParams creates new instance of MsgUpdateParams
func NewMsgUpdateParams(
	sender sdk.Address,
	params Params,
) *MsgUpdateParams {
	return &MsgUpdateParams{
		Authority: sender.String(),
		Params:    params,
	}
}

// Route returns the name of the module
func (msg MsgUpdateParams) Route() string { return ModuleName }

// Type returns the the action
func (msg MsgUpdateParams) Type() string { return "update_params" }

// GetSignBytes implements the Msg interface.
func (m MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(ModuleCdc.MustMarshalJSON(&m))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (m *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// Validate does a sanity check on the provided data.
func (m *MsgUpdateParams) Validate() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errors.Wrap(err, "invalid authority address")
	}

	return m.Params.Validate()
}

// MsgRecordsWrite implementation
func (m *MsgRecordsWrite) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Author)
	return []sdk.AccAddress{addr}
}

func (m *MsgRecordsWrite) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Author); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid author address: %s", err)
	}
	if m.Target == "" {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "target DID cannot be empty")
	}
	if m.Descriptor_ == nil {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "descriptor cannot be nil")
	}
	if len(m.Data) == 0 {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "data cannot be empty")
	}
	return nil
}

// MsgRecordsDelete implementation
func (m *MsgRecordsDelete) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Author)
	return []sdk.AccAddress{addr}
}

func (m *MsgRecordsDelete) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Author); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid author address: %s", err)
	}
	if m.Target == "" {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "target DID cannot be empty")
	}
	if m.RecordId == "" {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "record ID cannot be empty")
	}
	if m.Descriptor_ == nil {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "descriptor cannot be nil")
	}
	return nil
}

// MsgProtocolsConfigure implementation
func (m *MsgProtocolsConfigure) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Author)
	return []sdk.AccAddress{addr}
}

func (m *MsgProtocolsConfigure) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Author); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid author address: %s", err)
	}
	if m.Target == "" {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "target DID cannot be empty")
	}
	if m.ProtocolUri == "" {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "protocol URI cannot be empty")
	}
	if m.Descriptor_ == nil {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "descriptor cannot be nil")
	}
	if len(m.Definition) == 0 {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "definition cannot be empty")
	}
	return nil
}

// MsgPermissionsGrant implementation
func (m *MsgPermissionsGrant) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Grantor)
	return []sdk.AccAddress{addr}
}

func (m *MsgPermissionsGrant) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Grantor); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid grantor address: %s", err)
	}
	if m.Grantee == "" {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "grantee DID cannot be empty")
	}
	if m.Target == "" {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "target DID cannot be empty")
	}
	if m.InterfaceName == "" {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "interface name cannot be empty")
	}
	if m.Method == "" {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "method cannot be empty")
	}
	if m.Descriptor_ == nil {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "descriptor cannot be nil")
	}
	return nil
}

// MsgPermissionsRevoke implementation

// GetSigners returns the expected signers for a MsgPermissionsRevoke message.
func (m *MsgPermissionsRevoke) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Grantor)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data
func (m *MsgPermissionsRevoke) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Grantor); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid grantor address: %s", err)
	}
	if m.PermissionId == "" {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "permission ID cannot be empty")
	}
	if m.Descriptor_ == nil {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "descriptor cannot be nil")
	}
	return nil
}

// GetSigners returns the expected signers for a MsgRotateVaultKeys message
func (m *MsgRotateVaultKeys) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(m.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data
func (m *MsgRotateVaultKeys) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(m.Authority); err != nil {
		return errors.Wrapf(sdkerrors.ErrInvalidAddress, "invalid authority address: %s", err)
	}
	if m.Reason == "" {
		return errors.Wrap(sdkerrors.ErrInvalidRequest, "reason cannot be empty")
	}
	return nil
}

// GetModuleAddress returns the dwn module account address
func GetModuleAddress() sdk.AccAddress {
	return address.Module(ModuleName)
}
