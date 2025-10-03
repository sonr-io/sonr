## Overview
The DEX module enables decentralized exchange operations across multiple chains via IBC (Inter-Blockchain Communication).

## Request: {{.RequestType.Name}}

| Field ID | Name | Type | Description |
| -------- | ---- | ---- | ----------- | {{range .RequestType.Fields}}
| {{.Number}} | {{.Name}} | {{if eq .Label.String "LABEL_REPEATED"}}[]{{end}}{{.Type}} | {{fieldcomments .Message .}} | {{end}}

## Response: {{.ResponseType.Name}}

| Field ID | Name | Type | Description |
| -------- | ---- | ---- | ----------- | {{range .ResponseType.Fields}}
| {{.Number}} | {{.Name}} | {{if eq .Label.String "LABEL_REPEATED"}}[]{{end}}{{.Type}} | {{fieldcomments .Message .}} | {{end}}

## Implementation Notes

{{if eq .MethodDescriptorProto.Name "RegisterDEXAccount"}}
- Creates an ICA (Interchain Account) for DEX operations on remote chains
- Establishes secure cross-chain communication channels
- Supports configurable feature sets for different DEX capabilities
{{else if eq .MethodDescriptorProto.Name "ExecuteSwap"}}
- Performs atomic token swaps on remote DEX platforms
- Includes slippage protection via minimum output amount
- Supports custom routing for optimal price execution
- UCAN token authorization ensures secure delegation
{{else if eq .MethodDescriptorProto.Name "ProvideLiquidity"}}
- Adds liquidity to AMM (Automated Market Maker) pools
- Returns LP (Liquidity Provider) tokens as proof of deposit
- Protects against slippage during liquidity provision
{{else if eq .MethodDescriptorProto.Name "RemoveLiquidity"}}
- Withdraws liquidity from pools using LP tokens
- Ensures minimum asset amounts to protect against losses
- Supports partial and full liquidity removal
{{else if eq .MethodDescriptorProto.Name "CreateLimitOrder"}}
- Places limit orders on remote orderbook DEXs
- Supports time-based expiration for orders
- Automatically executes when price conditions are met
{{else if eq .MethodDescriptorProto.Name "CancelOrder"}}
- Cancels pending limit orders before execution
- Returns unfilled assets to user's account
- Requires proper authorization via UCAN token
{{end}}

## Security Considerations

- All operations require UCAN authorization tokens for delegation
- IBC packet timeouts prevent stuck transactions
- Connection IDs must reference valid, established IBC channels
- DID-based authentication ensures identity verification