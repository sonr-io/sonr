package module

import (
	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	modulev1 "github.com/sonr-io/sonr/api/svc/v1"
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
					RpcMethod: "DomainVerification",
					Use:       "domain-verification [domain]",
					Short:     "Query domain verification status",
					Long: "Query the verification status and details for a specific domain.\n" +
						"Shows the verification token, status, and expiration time.\n\n" +
						"Example:\n" +
						"  snrd query svc domain-verification example.com",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "domain"},
					},
				},
				{
					RpcMethod: "Service",
					Use:       "service [service-id]",
					Short:     "Query service information by ID",
					Long: "Query detailed information about a registered service including domain binding,\n" +
						"permissions, and capability information.\n\n" +
						"Example:\n" +
						"  snrd query svc service my-app",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "service_id"},
					},
				},
				{
					RpcMethod: "ServicesByOwner",
					Use:       "services-by-owner [owner]",
					Short:     "Query all services owned by an address",
					Long: "Query all services registered by a specific owner address.\n\n" +
						"Example:\n" +
						"  snrd query svc services-by-owner idx1abc123...",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "owner"},
					},
				},
				{
					RpcMethod: "ServicesByDomain",
					Use:       "services-by-domain [domain]",
					Short:     "Query services bound to a specific domain",
					Long: "Query services that are bound to a specific verified domain.\n\n" +
						"Example:\n" +
						"  snrd query svc services-by-domain example.com",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "domain"},
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
					RpcMethod: "InitiateDomainVerification",
					Use:       "initiate-domain-verification [domain]",
					Short:     "Initiate domain verification by generating a DNS TXT record token",
					Long: "Initiate domain verification by generating a unique verification token that must be added as a DNS TXT record.\n" +
						"The generated token must be added to your domain's DNS records as:\n" +
						"sonr-verification=<token>\n\n" +
						"Example:\n" +
						"  snrd tx svc initiate-domain-verification example.com --from alice",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "domain"},
					},
				},
				{
					RpcMethod: "VerifyDomain",
					Use:       "verify-domain [domain]",
					Short:     "Verify domain ownership by checking DNS TXT records",
					Long: "Verify domain ownership by performing a DNS lookup for the verification token.\n" +
						"This command checks if the required DNS TXT record has been properly configured.\n\n" +
						"Example:\n" +
						"  snrd tx svc verify-domain example.com --from alice",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "domain"},
					},
				},
				{
					RpcMethod: "RegisterService",
					Use:       "register-service [service-id] [domain] [permissions...]",
					Short:     "Register a new service with the specified domain and permissions",
					Long: "Register a new service that will be bound to a verified domain with specific permissions.\n" +
						"The domain must be verified before service registration.\n\n" +
						"Available permissions:\n" +
						"  - dwn:read, dwn:write, dwn:delete (DWN operations)\n" +
						"  - identity:read, identity:write (Identity operations)\n" +
						"  - vault:access (Vault access)\n" +
						"  - service:register, service:update, service:delete (Service management)\n\n" +
						"Example:\n" +
						"  snrd tx svc register-service my-app example.com dwn:read,dwn:write,identity:read --from alice\n" +
						"  snrd tx svc register-service vault-service vault.example.com vault:access,identity:read --ucan-delegation-chain=\"<jwt-token>\" --from alice",
					PositionalArgs: []*autocliv1.PositionalArgDescriptor{
						{ProtoField: "service_id"},
						{ProtoField: "domain"},
						{ProtoField: "requested_permissions", Varargs: true},
					},
					FlagOptions: map[string]*autocliv1.FlagOptions{
						"ucan_delegation_chain": {
							Name:         "ucan-delegation-chain",
							Shorthand:    "u",
							Usage:        "UCAN delegation chain (JWT format) for authorization",
							DefaultValue: "",
						},
					},
				},
			},
		},
	}
}
