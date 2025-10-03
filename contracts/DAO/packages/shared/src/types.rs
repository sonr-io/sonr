use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Addr, Timestamp, Uint128};

/// Represents a DID holder with voting power
#[cw_serde]
pub struct IdentityVoter {
    /// DID of the voter
    pub did: String,
    /// Address associated with the DID
    pub address: Addr,
    /// Voting power based on identity attributes
    pub voting_power: Uint128,
    /// Verification level (0-100)
    pub verification_level: u8,
    /// Reputation score
    pub reputation_score: u64,
}

/// Identity verification status
#[cw_serde]
pub enum VerificationStatus {
    /// Not verified
    Unverified,
    /// Basic verification completed
    Basic,
    /// Advanced verification with KYC
    Advanced,
    /// Full verification with attestations
    Full,
}

/// Proposal status in the DAO
#[cw_serde]
pub enum ProposalStatus {
    /// Pending approval from pre-propose module
    Pending,
    /// Open for voting
    Open,
    /// Voting period ended, waiting execution
    Passed,
    /// Proposal rejected
    Rejected,
    /// Proposal executed
    Executed,
    /// Proposal execution failed
    ExecutionFailed,
}

/// Vote option
#[cw_serde]
pub enum Vote {
    Yes,
    No,
    Abstain,
    NoWithVeto,
}

/// Voting configuration
#[cw_serde]
pub struct VotingConfig {
    /// Minimum percentage of yes votes required
    pub threshold: Decimal,
    /// Minimum voter turnout percentage
    pub quorum: Decimal,
    /// Voting duration in seconds
    pub voting_period: u64,
    /// Proposal deposit amount
    pub proposal_deposit: Uint128,
}

/// Identity attestation
#[cw_serde]
pub struct IdentityAttestation {
    /// DID being attested
    pub did: String,
    /// Attester's DID
    pub attester_did: String,
    /// Type of attestation
    pub attestation_type: AttestationType,
    /// Attestation data
    pub data: String,
    /// Timestamp of attestation
    pub timestamp: Timestamp,
    /// Expiration time
    pub expires_at: Option<Timestamp>,
}

/// Types of attestations
#[cw_serde]
pub enum AttestationType {
    /// Identity verification
    Identity,
    /// Skill or credential
    Credential,
    /// Reputation endorsement
    Reputation,
    /// Custom attestation
    Custom(String),
}

/// DAO treasury info
#[cw_serde]
pub struct TreasuryInfo {
    /// Treasury address
    pub address: Addr,
    /// Available balance
    pub balance: Uint128,
    /// Reserved funds for proposals
    pub reserved: Uint128,
}

/// Module configuration
#[cw_serde]
pub struct ModuleConfig {
    /// Core DAO contract address
    pub dao_core: Addr,
    /// Voting module address
    pub voting_module: Addr,
    /// Proposal module address
    pub proposal_module: Addr,
    /// Pre-propose module address
    pub pre_propose_module: Addr,
    /// x/did module integration enabled
    pub did_integration_enabled: bool,
}

use cosmwasm_std::Decimal;