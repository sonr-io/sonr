use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{Addr, CustomQuery, Uint128};

/// Custom query for x/did module integration via Stargate
#[cw_serde]
#[derive(QueryResponses)]
pub enum SonrQuery {
    /// Query DID document by DID
    #[returns(DIDDocumentResponse)]
    GetDIDDocument { did: String },
    
    /// Query if DID is verified
    #[returns(VerificationResponse)]
    IsDIDVerified { did: String },
    
    /// Query DID by address
    #[returns(DIDByAddressResponse)]
    GetDIDByAddress { address: String },
    
    /// Query all DIDs with pagination
    #[returns(DIDsResponse)]
    ListDIDs {
        start_after: Option<String>,
        limit: Option<u32>,
    },
    
    /// Query WebAuthn credentials for DID
    #[returns(WebAuthnCredentialsResponse)]
    GetWebAuthnCredentials { did: String },
}

impl CustomQuery for SonrQuery {}

/// DID Document response
#[cw_serde]
pub struct DIDDocumentResponse {
    pub did: String,
    pub controller: String,
    pub verification_methods: Vec<VerificationMethod>,
    pub authentication: Vec<String>,
    pub assertion_method: Vec<String>,
    pub capability_invocation: Vec<String>,
    pub capability_delegation: Vec<String>,
    pub service: Vec<Service>,
}

/// Verification method in DID document
#[cw_serde]
pub struct VerificationMethod {
    pub id: String,
    pub controller: String,
    pub method_type: String,
    pub public_key: String,
}

/// Service endpoint in DID document
#[cw_serde]
pub struct Service {
    pub id: String,
    pub service_type: String,
    pub service_endpoint: String,
}

/// DID verification response
#[cw_serde]
pub struct VerificationResponse {
    pub is_verified: bool,
    pub verification_level: u8,
    pub last_verified: Option<u64>,
}

/// DID by address response
#[cw_serde]
pub struct DIDByAddressResponse {
    pub did: Option<String>,
    pub address: String,
}

/// List of DIDs response
#[cw_serde]
pub struct DIDsResponse {
    pub dids: Vec<DIDInfo>,
    pub total: u64,
}

/// Basic DID information
#[cw_serde]
pub struct DIDInfo {
    pub did: String,
    pub controller: Addr,
    pub created_at: u64,
    pub updated_at: u64,
}

/// WebAuthn credentials response
#[cw_serde]
pub struct WebAuthnCredentialsResponse {
    pub credentials: Vec<WebAuthnCredential>,
}

/// WebAuthn credential
#[cw_serde]
pub struct WebAuthnCredential {
    pub credential_id: String,
    pub public_key: String,
    pub attestation_type: String,
    pub user_verified: bool,
}

/// Stargate query wrapper for x/did module
#[cw_serde]
pub struct StargateQuery {
    /// Path to the module query endpoint
    pub path: String,
    /// Protobuf encoded query data
    pub data: Vec<u8>,
}

/// Helper to create stargate queries for x/did module
pub mod stargate {
    use super::*;
    
    /// Query path for x/did module
    pub const DID_MODULE_PATH: &str = "/sonr.did.v1.Query";
    
    /// Create a stargate query for DID document
    pub fn query_did_document(did: &str) -> StargateQuery {
        StargateQuery {
            path: format!("{}/DIDDocument", DID_MODULE_PATH),
            data: encode_did_query(did),
        }
    }
    
    /// Create a stargate query for DID verification
    pub fn query_did_verification(did: &str) -> StargateQuery {
        StargateQuery {
            path: format!("{}/VerifyDID", DID_MODULE_PATH),
            data: encode_did_query(did),
        }
    }
    
    // Helper to encode DID query (simplified - actual implementation would use prost)
    fn encode_did_query(did: &str) -> Vec<u8> {
        // This would use prost to encode the protobuf message
        // For now, returning a placeholder
        did.as_bytes().to_vec()
    }
}