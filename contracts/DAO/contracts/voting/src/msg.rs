use cosmwasm_schema::cw_serde;
use identity_dao_shared::Vote;

/// Instantiate message
#[cw_serde]
pub struct InstantiateMsg {
    pub dao_core: String,
    pub min_verification_level: u8,
    pub use_reputation_weight: bool,
}

/// Execute messages
#[cw_serde]
pub enum ExecuteMsg {
    /// Cast a vote
    Vote {
        proposal_id: u64,
        vote: Vote,
    },
    /// Update voter registration
    UpdateVoter {
        did: String,
        address: String,
    },
    /// Remove voter
    RemoveVoter { 
        did: String 
    },
}