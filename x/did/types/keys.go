package types

import (
	"cosmossdk.io/collections"

	ormv1alpha1 "cosmossdk.io/api/cosmos/orm/v1alpha1"
)

// ParamsKey saves the current module params.
var ParamsKey = collections.NewPrefix(0)

const (
	ModuleName = "did"

	StoreKey = ModuleName

	QuerierRoute = ModuleName
)

// Event types and attribute keys
const (
	// Event types
	EventTypeDIDCreated                = "did_created"
	EventTypeDIDUpdated                = "did_updated"
	EventTypeDIDDeactivated            = "did_deactivated"
	EventTypeVerificationMethodAdded   = "verification_method_added"
	EventTypeVerificationMethodRemoved = "verification_method_removed"
	EventTypeServiceAdded              = "service_added"
	EventTypeServiceRemoved            = "service_removed"
	EventTypeCredentialIssued          = "credential_issued"
	EventTypeCredentialRevoked         = "credential_revoked"
	EventTypeExternalWalletLinked      = "external_wallet_linked"

	// Attribute keys
	AttributeKeyDID                = "did"
	AttributeKeyController         = "controller"
	AttributeKeyVersion            = "version"
	AttributeKeyVerificationMethod = "verification_method"
	AttributeKeyService            = "service"
	AttributeKeyCredential         = "credential"
	AttributeKeyIssuer             = "issuer"
	AttributeKeySubject            = "subject"
)

var ORMModuleSchema = ormv1alpha1.ModuleSchemaDescriptor{
	SchemaFile: []*ormv1alpha1.ModuleSchemaDescriptor_FileEntry{
		{Id: 1, ProtoFileName: "did/v1/state.proto"},
	},
	Prefix: []byte{0},
}
