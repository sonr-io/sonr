use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::Uint128;
use identity_dao_shared::{IdentityVoter, VoteInfo};

/// Query messages
#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    /// Get voting power for a DID
    #[returns(VotingPowerResponse)]
    VotingPower { did: String },
    
    /// Get total voting power
    #[returns(TotalPowerResponse)]
    TotalPower { height: Option<u64> },
    
    /// Get voter info
    #[returns(VoterInfoResponse)]
    VoterInfo { did: String },
    
    /// List all voters with pagination
    #[returns(VotersListResponse)]
    ListVoters {
        start_after: Option<String>,
        limit: Option<u32>,
    },
    
    /// Get vote on a proposal
    #[returns(VoteResponse)]
    Vote { 
        proposal_id: u64,
        voter: String,
    },
}

/// Voting power response
#[cw_serde]
pub struct VotingPowerResponse {
    pub power: Uint128,
    pub height: u64,
}

/// Total power response
#[cw_serde]
pub struct TotalPowerResponse {
    pub power: Uint128,
    pub height: u64,
}

/// Voter info response
#[cw_serde]
pub struct VoterInfoResponse {
    pub voter: IdentityVoter,
    pub proposals_voted: u64,
}

/// Voters list response
#[cw_serde]
pub struct VotersListResponse {
    pub voters: Vec<IdentityVoter>,
    pub total: u64,
}

/// Vote response
#[cw_serde]
pub struct VoteResponse {
    pub vote: Option<VoteInfo>,
}