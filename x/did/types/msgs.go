package types

import (
	"fmt"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

var (
	_ sdk.Msg = &MsgUpdateParams{}
	_ sdk.Msg = &MsgCreateDID{}
	_ sdk.Msg = &MsgUpdateDID{}
	_ sdk.Msg = &MsgDeactivateDID{}
	_ sdk.Msg = &MsgAddVerificationMethod{}
	_ sdk.Msg = &MsgRemoveVerificationMethod{}
	_ sdk.Msg = &MsgAddService{}
	_ sdk.Msg = &MsgRemoveService{}
	_ sdk.Msg = &MsgIssueVerifiableCredential{}
	_ sdk.Msg = &MsgRevokeVerifiableCredential{}
	_ sdk.Msg = &MsgLinkExternalWallet{}
	_ sdk.Msg = &MsgRegisterWebAuthnCredential{}
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

// GetSignBytes implements the LegacyMsg interface.
func (msg MsgUpdateParams) GetSignBytes() []byte {
	return sdk.MustSortJSON(AminoCdc.MustMarshalJSON(&msg))
}

// GetSigners returns the expected signers for a MsgUpdateParams message.
func (msg *MsgUpdateParams) GetSigners() []sdk.AccAddress {
	addr, _ := sdk.AccAddressFromBech32(msg.Authority)
	return []sdk.AccAddress{addr}
}

// ValidateBasic does a sanity check on the provided data.
func (msg *MsgUpdateParams) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Authority); err != nil {
		return errors.Wrap(ErrInvalidAuthorityAddress, err.Error())
	}

	return msg.Params.Validate()
}

// Validate validates the message.
func (msg *MsgUpdateParams) Validate() error {
	return msg.Params.Validate()
}

// ValidateBasic does a sanity check on MsgCreateDID.
func (msg *MsgCreateDID) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Controller); err != nil {
		return errors.Wrap(ErrInvalidControllerAddress, err.Error())
	}

	if msg.DidDocument.Id == "" {
		return ErrEmptyDIDDocumentID
	}

	return nil
}

// ValidateBasic does a sanity check on MsgUpdateDID.
func (msg *MsgUpdateDID) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Controller); err != nil {
		return errors.Wrap(ErrInvalidControllerAddress, err.Error())
	}

	if msg.Did == "" {
		return ErrEmptyDID
	}

	if msg.DidDocument.Id == "" {
		return ErrEmptyDIDDocumentID
	}

	if msg.Did != msg.DidDocument.Id {
		return ErrDIDMismatch
	}

	return nil
}

// ValidateBasic does a sanity check on MsgDeactivateDID.
func (msg *MsgDeactivateDID) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Controller); err != nil {
		return errors.Wrap(ErrInvalidControllerAddress, err.Error())
	}

	if msg.Did == "" {
		return ErrEmptyDID
	}

	return nil
}

// ValidateBasic does a sanity check on MsgAddVerificationMethod.
func (msg *MsgAddVerificationMethod) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Controller); err != nil {
		return errors.Wrap(ErrInvalidControllerAddress, err.Error())
	}

	if msg.Did == "" {
		return ErrEmptyDID
	}

	if msg.VerificationMethod.Id == "" {
		return ErrEmptyVerificationMethodID
	}

	if msg.VerificationMethod.VerificationMethodKind == "" {
		return ErrEmptyVerificationMethodKind
	}

	return nil
}

// ValidateBasic does a sanity check on MsgRemoveVerificationMethod.
func (msg *MsgRemoveVerificationMethod) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Controller); err != nil {
		return errors.Wrap(ErrInvalidControllerAddress, err.Error())
	}

	if msg.Did == "" {
		return ErrEmptyDID
	}

	if msg.VerificationMethodId == "" {
		return ErrEmptyVerificationMethodID
	}

	return nil
}

// ValidateBasic does a sanity check on MsgAddService.
func (msg *MsgAddService) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Controller); err != nil {
		return errors.Wrap(ErrInvalidControllerAddress, err.Error())
	}

	if msg.Did == "" {
		return ErrEmptyDID
	}

	if msg.Service.Id == "" {
		return ErrEmptyServiceID
	}

	if msg.Service.ServiceKind == "" {
		return ErrEmptyServiceKind
	}

	return nil
}

// ValidateBasic does a sanity check on MsgRemoveService.
func (msg *MsgRemoveService) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Controller); err != nil {
		return errors.Wrap(ErrInvalidControllerAddress, err.Error())
	}

	if msg.Did == "" {
		return ErrEmptyDID
	}

	if msg.ServiceId == "" {
		return ErrEmptyServiceID
	}

	return nil
}

// ValidateBasic does a sanity check on MsgIssueVerifiableCredential.
func (msg *MsgIssueVerifiableCredential) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Issuer); err != nil {
		return errors.Wrap(ErrInvalidIssuerAddress, err.Error())
	}

	if msg.Credential.Id == "" {
		return ErrEmptyCredentialID
	}

	if msg.Credential.Issuer == "" {
		return ErrEmptyCredentialIssuer
	}

	return nil
}

// ValidateBasic does a sanity check on MsgRevokeVerifiableCredential.
func (msg *MsgRevokeVerifiableCredential) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Issuer); err != nil {
		return errors.Wrap(ErrInvalidIssuerAddress, err.Error())
	}

	if msg.CredentialId == "" {
		return ErrEmptyCredentialID
	}

	return nil
}

// ValidateBasic does a sanity check on MsgLinkExternalWallet.
func (msg *MsgLinkExternalWallet) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Controller); err != nil {
		return errors.Wrap(ErrInvalidControllerAddress, err.Error())
	}

	if msg.Did == "" {
		return ErrEmptyDID
	}

	if msg.WalletAddress == "" {
		return errors.Wrap(ErrInvalidWalletVerification, "wallet address cannot be empty")
	}

	if msg.WalletChainId == "" {
		return errors.Wrap(ErrInvalidWalletVerification, "chain ID cannot be empty")
	}

	if msg.WalletType == "" {
		return errors.Wrap(ErrInvalidWalletVerification, "wallet type cannot be empty")
	}

	// Validate wallet type
	walletType := WalletType(msg.WalletType)
	if err := walletType.Validate(); err != nil {
		return err
	}

	if len(msg.OwnershipProof) == 0 {
		return errors.Wrap(ErrInvalidWalletVerification, "ownership proof cannot be empty")
	}

	if len(msg.Challenge) == 0 {
		return errors.Wrap(ErrInvalidWalletVerification, "challenge cannot be empty")
	}

	if msg.VerificationMethodId == "" {
		return ErrEmptyVerificationMethodID
	}

	// Validate blockchain account ID format
	accountID, err := ParseBlockchainAccountID(fmt.Sprintf("%s:%s:%s",
		walletType.GetNamespace(), msg.WalletChainId, msg.WalletAddress))
	if err != nil {
		return err
	}

	if err := accountID.Validate(); err != nil {
		return err
	}

	return nil
}

// ValidateBasic does a sanity check on MsgRegisterWebAuthnCredential.
func (msg *MsgRegisterWebAuthnCredential) ValidateBasic() error {
	if _, err := sdk.AccAddressFromBech32(msg.Controller); err != nil {
		return errors.Wrap(ErrInvalidControllerAddress, err.Error())
	}

	if msg.Username == "" {
		return errors.Wrap(ErrInvalidWebAuthnCredential, "username cannot be empty")
	}

	if msg.WebauthnCredential.CredentialId == "" {
		return errors.Wrap(ErrInvalidWebAuthnCredential, "credential ID cannot be empty")
	}

	if msg.WebauthnCredential.Origin == "" {
		return errors.Wrap(ErrInvalidWebAuthnCredential, "origin cannot be empty")
	}

	if len(msg.WebauthnCredential.PublicKey) == 0 {
		return errors.Wrap(ErrInvalidWebAuthnCredential, "public key cannot be empty")
	}

	if msg.VerificationMethodId == "" {
		return ErrEmptyVerificationMethodID
	}

	return nil
}
