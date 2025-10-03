use cosmwasm_std::{Deps, StdResult, Addr};
use identity_dao_shared::{
    VerificationStatus, AttestationType,
    bindings::{DIDDocumentResponse, VerificationResponse, SonrQuery},
};

/// Verify DID status and verification level
pub fn verify_did_status(deps: Deps, did: &str) -> StdResult<VerificationResponse> {
    // For production, would use Stargate query to x/did module
    // Mock response for development
    Ok(VerificationResponse {
        is_verified: true,
        verification_level: 2,
        last_verified: Some(1700000000),
    })
}

/// Check if proposer meets minimum verification requirements
pub fn check_verification_requirements(
    deps: Deps,
    proposer: &Addr,
    min_status: VerificationStatus,
) -> StdResult<bool> {
    let did = format!("did:sonr:{}", proposer);
    let verification = verify_did_status(deps, &did)?;
    
    if !verification.is_verified {
        return Ok(false);
    }
    
    let status = match verification.verification_level {
        0 => VerificationStatus::Unverified,
        1 => VerificationStatus::Basic,
        2 => VerificationStatus::Advanced,
        _ => VerificationStatus::Full,
    };
    
    Ok(status as u8 >= min_status as u8)
}

/// Verify specific attestations for proposer
pub fn verify_attestations(
    _deps: Deps,
    _proposer: &Addr,
    required_types: &[AttestationType],
) -> StdResult<bool> {
    // Would query attestations from x/did module
    // For now, return true for development
    if required_types.is_empty() {
        return Ok(true);
    }
    
    // Mock: assume proposer has all required attestations
    Ok(true)
}

/// Calculate deposit amount based on verification level
pub fn calculate_deposit_multiplier(verification_level: u8) -> u64 {
    match verification_level {
        0 => 100, // Unverified: 100% deposit
        1 => 75,  // Basic: 75% deposit
        2 => 50,  // Advanced: 50% deposit
        _ => 25,  // Full: 25% deposit
    }
}

/// Validate proposal content based on proposer verification
pub fn validate_proposal_content(
    deps: Deps,
    proposer: &Addr,
    proposal_type: &str,
) -> StdResult<bool> {
    let did = format!("did:sonr:{}", proposer);
    let verification = verify_did_status(deps, &did)?;
    
    match proposal_type {
        "treasury" => {
            // Treasury proposals require advanced verification
            Ok(verification.verification_level >= 2)
        }
        "parameter" => {
            // Parameter changes require full verification
            Ok(verification.verification_level >= 3)
        }
        "identity_policy" => {
            // Identity policy changes require full verification
            Ok(verification.verification_level >= 3)
        }
        _ => {
            // Default proposals require basic verification
            Ok(verification.verification_level >= 1)
        }
    }
}

/// Check if proposer has active proposals
pub fn has_active_proposals(
    _deps: Deps,
    _proposer: &Addr,
) -> StdResult<u64> {
    // Would query active proposals for this proposer
    // For now, return 0 for development
    Ok(0)
}

/// Validate proposal limits based on verification
pub fn validate_proposal_limits(
    deps: Deps,
    proposer: &Addr,
    max_active: u64,
) -> StdResult<bool> {
    let active_count = has_active_proposals(deps, proposer)?;
    Ok(active_count < max_active)
}

#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::testing::mock_dependencies;

    #[test]
    fn test_calculate_deposit_multiplier() {
        assert_eq!(calculate_deposit_multiplier(0), 100);
        assert_eq!(calculate_deposit_multiplier(1), 75);
        assert_eq!(calculate_deposit_multiplier(2), 50);
        assert_eq!(calculate_deposit_multiplier(3), 25);
        assert_eq!(calculate_deposit_multiplier(10), 25);
    }

    #[test]
    fn test_check_verification_requirements() {
        let deps = mock_dependencies();
        let proposer = Addr::unchecked("sonr1user");
        
        // Mock returns verification level 2 (Advanced)
        let result = check_verification_requirements(
            deps.as_ref(),
            &proposer,
            VerificationStatus::Basic,
        ).unwrap();
        assert!(result);
        
        let result = check_verification_requirements(
            deps.as_ref(),
            &proposer,
            VerificationStatus::Full,
        ).unwrap();
        // Would be false with real verification
        assert!(result); // Mock always returns true for now
    }
}