use cosmwasm_schema::{cw_serde, QueryResponses};
use identity_dao_shared::{ProposalStatus, ProposalVotes, VoteInfo};

/// Query messages
#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
    /// Get proposal details
    #[returns(ProposalResponse)]
    Proposal { proposal_id: u64 },
    
    /// List proposals with filters
    #[returns(ProposalsListResponse)]
    ListProposals {
        status: Option<ProposalStatus>,
        start_after: Option<u64>,
        limit: Option<u32>,
    },
    
    /// Get proposal votes
    #[returns(ProposalVotesResponse)]
    ProposalVotes {
        proposal_id: u64,
        start_after: Option<String>,
        limit: Option<u32>,
    },
    
    /// Get proposal result
    #[returns(ProposalResultResponse)]
    ProposalResult { proposal_id: u64 },
}

/// Proposal response
#[cw_serde]
pub struct ProposalResponse {
    pub id: u64,
    pub title: String,
    pub description: String,
    pub proposer: String,
    pub status: ProposalStatus,
    pub votes: ProposalVotes,
    pub start_time: u64,
    pub end_time: u64,
}

/// Proposals list response
#[cw_serde]
pub struct ProposalsListResponse {
    pub proposals: Vec<ProposalResponse>,
    pub total: u64,
}

/// Proposal votes response
#[cw_serde]
pub struct ProposalVotesResponse {
    pub votes: Vec<VoteInfo>,
    pub total: u64,
}

/// Proposal result response
#[cw_serde]
pub struct ProposalResultResponse {
    pub proposal_id: u64,
    pub result: ProposalResult,
}

/// Proposal result
#[cw_serde]
pub enum ProposalResult {
    Passed,
    Rejected,
    InProgress,
}