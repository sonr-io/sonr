package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	modulev1 "github.com/sonr-io/sonr/api/dwn/v1"
)

// AutoCLIOptions implements the autocli.HasAutoCLIConfig interface.
func (a AppModule) AutoCLIOptions() *autocliv1.ModuleOptions {
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
					RpcMethod: "Records",
					Use:       "records [target]",
					Short:     "Query DWN records for a target",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "target"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"protocol":       {Usage: "Filter by protocol URI"},
						"schema":         {Usage: "Filter by schema URI"},
						"parent_id":      {Usage: "Filter by parent record ID"},
						"published_only": {Usage: "Filter to show only published records"},
					},
				},
				{
					RpcMethod: "Record",
					Use:       "record [target] [record-id]",
					Short:     "Query a specific DWN record",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "target"},
						{ProtoField: "record_id"},
					},
				},
				{
					RpcMethod: "Protocols",
					Use:       "protocols [target]",
					Short:     "Query DWN protocols for a target",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "target"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"published_only": {Usage: "Filter to show only published protocols"},
					},
				},
				{
					RpcMethod: "Protocol",
					Use:       "protocol [target] [protocol-uri]",
					Short:     "Query a specific DWN protocol",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "target"},
						{ProtoField: "protocol_uri"},
					},
				},
				{
					RpcMethod: "Permissions",
					Use:       "permissions [target]",
					Short:     "Query DWN permissions for a target",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "target"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"grantor":         {Usage: "Filter by grantor DID"},
						"grantee":         {Usage: "Filter by grantee DID"},
						"interface_name":  {Usage: "Filter by interface name"},
						"method":          {Usage: "Filter by method name"},
						"include_revoked": {Usage: "Include revoked permissions in results"},
					},
				},
				{
					RpcMethod: "Vault",
					Use:       "vault [vault-id]",
					Short:     "Query a specific vault",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "vault_id"},
					},
				},
				{
					RpcMethod: "Vaults",
					Use:       "vaults",
					Short:     "Query vaults by owner",
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"owner": {Usage: "Filter by owner address (defaults to sender)"},
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
					RpcMethod: "RecordsWrite",
					Use:       "records-write [target] [data]",
					Short:     "Creates or updates a record in the DWN",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "target"},
						{ProtoField: "data"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"authorization": {Usage: "Authorization JWT or signature"},
						"attestation":   {Usage: "Attestation signature"},
						"encryption":    {Usage: "Encryption details"},
						"protocol":      {Usage: "Protocol URI this record conforms to"},
						"protocol_path": {Usage: "Protocol path"},
						"schema":        {Usage: "Schema URI for data validation"},
						"parent_id":     {Usage: "Parent record ID for threading"},
						"published":     {Usage: "Mark record as published"},
					},
				},
				{
					RpcMethod: "RecordsDelete",
					Use:       "records-delete [target] [record-id]",
					Short:     "Deletes a record from the DWN",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "target"},
						{ProtoField: "record_id"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"authorization": {Usage: "Authorization JWT or signature"},
						"prune":         {Usage: "Prune all descendant records"},
					},
				},
				{
					RpcMethod: "ProtocolsConfigure",
					Use:       "protocols-configure [target] [protocol-uri] [definition]",
					Short:     "Configures a protocol in the DWN",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "target"},
						{ProtoField: "protocol_uri"},
						{ProtoField: "definition"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"authorization": {Usage: "Authorization JWT or signature"},
						"published":     {Usage: "Mark protocol as published"},
					},
				},
				{
					RpcMethod: "PermissionsGrant",
					Use:       "permissions-grant [target] [grantee]",
					Short:     "Grants permissions in the DWN",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "target"},
						{ProtoField: "grantee"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"authorization": {Usage: "Authorization JWT or signature"},
					},
				},
				{
					RpcMethod: "PermissionsRevoke",
					Use:       "permissions-revoke [permission-id]",
					Short:     "Revokes permissions in the DWN",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "permission_id"},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"authorization": {Usage: "Authorization JWT or signature"},
					},
				},
			},
		},
	}
}
