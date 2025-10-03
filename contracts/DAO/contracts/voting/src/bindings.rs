use cosmwasm_std::{Deps, StdResult, QueryRequest};
use identity_dao_shared::bindings::{
    SonrQuery, DIDDocumentResponse, VerificationResponse, DIDByAddressResponse,
    WebAuthnCredentialsResponse, stargate,
};

/// Query DID document from x/did module
pub fn query_did_document(deps: Deps, did: &str) -> StdResult<DIDDocumentResponse> {
    // Create stargate query for DID document
    let query = stargate::query_did_document(did);
    
    // For production, would use actual stargate query:
    // deps.querier.query(&QueryRequest::Stargate {
    //     path: query.path,
    //     data: query.data.into(),
    // })
    
    // Mock response for development
    Ok(DIDDocumentResponse {
        did: did.to_string(),
        controller: "sonr1abc...".to_string(),
        verification_methods: vec![],
        authentication: vec![],
        assertion_method: vec![],
        capability_invocation: vec![],
        capability_delegation: vec![],
        service: vec![],
    })
}

/// Query DID verification status
pub fn query_did_verification(deps: Deps, did: &str) -> StdResult<VerificationResponse> {
    // Create stargate query for verification
    let query = stargate::query_did_verification(did);
    
    // For production, would use actual stargate query:
    // deps.querier.query(&QueryRequest::Stargate {
    //     path: query.path,
    //     data: query.data.into(),
    // })
    
    // Mock response for development
    Ok(VerificationResponse {
        is_verified: true,
        verification_level: 2,
        last_verified: Some(1700000000),
    })
}

/// Query DID by address
pub fn query_did_by_address(deps: Deps, address: &str) -> StdResult<DIDByAddressResponse> {
    // For production, would use actual custom query
    // deps.querier.query(&QueryRequest::Custom(
    //     SonrQuery::GetDIDByAddress { 
    //         address: address.to_string() 
    //     }
    // ))
    
    // Mock response for development
    Ok(DIDByAddressResponse {
        did: Some(format!("did:sonr:{}", address)),
        address: address.to_string(),
    })
}

/// Query WebAuthn credentials for a DID
pub fn query_webauthn_credentials(deps: Deps, did: &str) -> StdResult<WebAuthnCredentialsResponse> {
    // For production, would use actual custom query
    // deps.querier.query(&QueryRequest::Custom(
    //     SonrQuery::GetWebAuthnCredentials { 
    //         did: did.to_string() 
    //     }
    // ))
    
    // Mock response for development
    Ok(WebAuthnCredentialsResponse {
        credentials: vec![],
    })
}

/// Calculate voting power based on DID attributes
pub fn calculate_voting_power(
    verification_level: u8,
    reputation_score: u64,
    use_reputation: bool,
) -> u128 {
    let base_power = 100u128;
    let level_multiplier = verification_level as u128;
    
    if use_reputation {
        let reputation_multiplier = (reputation_score / 100).max(1) as u128;
        base_power * level_multiplier * reputation_multiplier
    } else {
        base_power * level_multiplier
    }
}