package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	modulev1 "github.com/sonr-io/sonr/api/did/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (am AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
	return &autocliv1.ModuleOptions{
		Query: &autocliv1.ServiceCommandDescriptor{
			Service: modulev1.Query_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "Params",
					Use:       "params",
					Short:     "Query the current consensus parameters",
				},
				{
					RpcMethod: "ResolveDID",
					Use:       "resolve [did]",
					Short:     "Resolve a DID to its document",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "did"},
					},
				},
				{
					RpcMethod: "GetDIDDocument",
					Use:       "document [did]",
					Short:     "Get a W3C DID document by DID",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "did"},
					},
				},
				{
					RpcMethod: "ListDIDDocuments",
					Use:       "documents",
					Short:     "List all W3C DID documents",
				},
				{
					RpcMethod: "GetDIDDocumentsByController",
					Use:       "documents-by-controller [controller]",
					Short:     "Get W3C DID documents by controller",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "controller"},
					},
				},
				{
					RpcMethod: "GetVerificationMethod",
					Use:       "verification-method [did] [method-id]",
					Short:     "Get a verification method from a DID document",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "did"},
						{ProtoField: "method_id"},
					},
				},
				{
					RpcMethod: "GetService",
					Use:       "service [did] [service-id]",
					Short:     "Get a service endpoint from a DID document",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "did"},
						{ProtoField: "service_id"},
					},
				},
				{
					RpcMethod: "GetVerifiableCredential",
					Use:       "credential [credential-id]",
					Short:     "Get a W3C verifiable credential",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "credential_id"},
					},
				},
				{
					RpcMethod: "ListVerifiableCredentials",
					Use:       "credentials",
					Short:     "List all W3C verifiable credentials",
				},
				{
					RpcMethod: "GetCredentialsByDID",
					Use:       "credentials-by-did [did]",
					Short:     "Get all credentials (verifiable and WebAuthn) associated with a DID",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "did"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"include_verifiable": {
							Usage: "Include verifiable credentials (default: true)",
						},
						"include_webauthn": {
							Usage: "Include WebAuthn credentials (default: true)",
						},
						"include_revoked": {
							Usage: "Include revoked credentials (default: false)",
						},
					},
				},
			},
		},
		Tx: &autocliv1.ServiceCommandDescriptor{
			Service: modulev1.Msg_ServiceDesc.ServiceName,
			RpcCommandOptions: []*autocliv1.RpcCommandOptions{
				{
					RpcMethod: "UpdateParams",
					Skip:      false, // set to true if authority gated
				},
				{
					RpcMethod: "CreateDID",
					Use:       "create-did [did-document]",
					Short:     "Create a new DID document",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "did_document"},
					},
				},
				{
					RpcMethod: "UpdateDID",
					Use:       "update-did [did] [did-document]",
					Short:     "Update an existing DID document",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "did"},
						{ProtoField: "did_document"},
					},
				},
				{
					RpcMethod: "DeactivateDID",
					Use:       "deactivate-did [did]",
					Short:     "Deactivate a DID document",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "did"},
					},
				},
				{
					RpcMethod: "AddVerificationMethod",
					Use:       "add-verification-method [did] [verification-method]",
					Short:     "Add a verification method to a DID document",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "did"},
						{ProtoField: "verification_method"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"relationships": {Usage: "Verification relationships (comma-separated)"},
					},
				},
				{
					RpcMethod: "RemoveVerificationMethod",
					Use:       "remove-verification-method [did] [verification-method-id]",
					Short:     "Remove a verification method from a DID document",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "did"},
						{ProtoField: "verification_method_id"},
					},
				},
				{
					RpcMethod: "AddService",
					Use:       "add-service [did] [service]",
					Short:     "Add a service endpoint to a DID document",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "did"},
						{ProtoField: "service"},
					},
				},
				{
					RpcMethod: "RemoveService",
					Use:       "remove-service [did] [service-id]",
					Short:     "Remove a service endpoint from a DID document",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "did"},
						{ProtoField: "service_id"},
					},
				},
				{
					RpcMethod: "IssueVerifiableCredential",
					Use:       "issue-credential [credential]",
					Short:     "Issue a W3C verifiable credential",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "credential"},
					},
				},
				{
					RpcMethod: "RevokeVerifiableCredential",
					Use:       "revoke-credential [credential-id]",
					Short:     "Revoke a W3C verifiable credential",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "credential_id"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"revocation_reason": {Usage: "Reason for credential revocation"},
					},
				},
			},
		},
	}
}
