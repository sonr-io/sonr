## Overview
Query endpoints for the DEX module provide read-only access to decentralized exchange state and operations.

## Request: {{.RequestType.Name}}

| Field ID | Name | Type | Description |
| -------- | ---- | ---- | ----------- | {{range .RequestType.Fields}}
| {{.Number}} | {{.Name}} | {{if eq .Label.String "LABEL_REPEATED"}}[]{{end}}{{.Type}} | {{fieldcomments .Message .}} | {{end}}

## Response: {{.ResponseType.Name}}

| Field ID | Name | Type | Description |
| -------- | ---- | ---- | ----------- | {{range .ResponseType.Fields}}
| {{.Number}} | {{.Name}} | {{if eq .Label.String "LABEL_REPEATED"}}[]{{end}}{{.Type}} | {{fieldcomments .Message .}} | {{end}}

## Query Details

{{if eq .MethodDescriptorProto.Name "Params"}}
- Returns the current module parameters
- Includes fee configurations and operational settings
- No authentication required for parameter queries
{{else if eq .MethodDescriptorProto.Name "Account"}}
- Retrieves a specific DEX account by DID and connection
- Returns ICA account details and status
- Shows linked features and capabilities
{{else if eq .MethodDescriptorProto.Name "Accounts"}}
- Lists all DEX accounts for a given DID
- Supports pagination for large result sets
- Returns accounts across all connected chains
{{else if eq .MethodDescriptorProto.Name "Balance"}}
- Queries balance on remote chain through ICA
- Can filter by specific denomination
- Returns real-time balance information
{{else if eq .MethodDescriptorProto.Name "Pool"}}
- Retrieves detailed pool information
- Shows current liquidity and swap fees
- Returns asset composition and total shares
{{else if eq .MethodDescriptorProto.Name "Orders"}}
- Lists orders for a specific DID and connection
- Supports filtering by order status
- Returns paginated order history
{{else if eq .MethodDescriptorProto.Name "History"}}
- Provides transaction history across all DEX operations
- Supports filtering by connection and operation type
- Returns detailed transaction records with timestamps
{{end}}

## Usage Examples

```bash
# Query module parameters
snrd query dex params

# Get specific DEX account
snrd query dex account did:sonr:123 connection-0

# Check balance on remote chain
snrd query dex balance did:sonr:123 connection-0 --denom uatom

# Query pool information
snrd query dex pool connection-0 pool-1

# List orders
snrd query dex orders did:sonr:123 connection-0
```