use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{Addr, Uint128};
use crate::types::{
    IdentityVoter, ProposalStatus, Vote, VotingConfig, 
    IdentityAttestation, TreasuryInfo, ModuleConfig
};

/// Core DAO query messages
#[cw_serde]
#[derive(QueryResponses)]
pub enum CoreQueryMsg {
    /// Get DAO configuration
    #[returns(DaoConfigResponse)]
    Config {},
    
    /// Get treasury information
    #[returns(TreasuryInfo)]
    Treasury {},
    
    /// Get module addresses
    #[returns(ModuleConfig)]
    Modules {},
    
    /// Get DAO stats
    #[returns(DaoStatsResponse)]
    Stats {},
}

/// Voting module query messages
#[cw_serde]
#[derive(QueryResponses)]
pub enum VotingQueryMsg {
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

/// Proposal module query messages
#[cw_serde]
#[derive(QueryResponses)]
pub enum ProposalQueryMsg {
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

/// Pre-propose module query messages
#[cw_serde]
#[derive(QueryResponses)]
pub enum PreProposeQueryMsg {
    /// Get pending proposals
    #[returns(PendingProposalsResponse)]
    PendingProposals {
        start_after: Option<u64>,
        limit: Option<u32>,
    },
    
    /// Get deposit info
    #[returns(DepositInfoResponse)]
    DepositInfo { proposer: String },
    
    /// Get module config
    #[returns(PreProposeConfigResponse)]
    Config {},
}

// Response types

#[cw_serde]
pub struct DaoConfigResponse {
    pub name: String,
    pub description: String,
    pub voting_config: VotingConfig,
    pub admin: Option<Addr>,
    pub did_integration_enabled: bool,
}

#[cw_serde]
pub struct DaoStatsResponse {
    pub total_proposals: u64,
    pub active_proposals: u64,
    pub total_voters: u64,
    pub treasury_balance: Uint128,
}

#[cw_serde]
pub struct VotingPowerResponse {
    pub power: Uint128,
    pub height: u64,
}

#[cw_serde]
pub struct TotalPowerResponse {
    pub power: Uint128,
    pub height: u64,
}

#[cw_serde]
pub struct VoterInfoResponse {
    pub voter: IdentityVoter,
    pub proposals_voted: u64,
}

#[cw_serde]
pub struct VotersListResponse {
    pub voters: Vec<IdentityVoter>,
    pub total: u64,
}

#[cw_serde]
pub struct VoteResponse {
    pub vote: Option<VoteInfo>,
}

#[cw_serde]
pub struct VoteInfo {
    pub proposal_id: u64,
    pub voter: String,
    pub vote: Vote,
    pub voting_power: Uint128,
}

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

#[cw_serde]
pub struct ProposalVotes {
    pub yes: Uint128,
    pub no: Uint128,
    pub abstain: Uint128,
    pub no_with_veto: Uint128,
}

#[cw_serde]
pub struct ProposalsListResponse {
    pub proposals: Vec<ProposalResponse>,
    pub total: u64,
}

#[cw_serde]
pub struct ProposalVotesResponse {
    pub votes: Vec<VoteInfo>,
    pub total: u64,
}

#[cw_serde]
pub struct ProposalResultResponse {
    pub proposal_id: u64,
    pub result: ProposalResult,
}

#[cw_serde]
pub enum ProposalResult {
    Passed,
    Rejected,
    InProgress,
}

#[cw_serde]
pub struct PendingProposalsResponse {
    pub proposals: Vec<PendingProposal>,
    pub total: u64,
}

#[cw_serde]
pub struct PendingProposal {
    pub id: u64,
    pub proposer: String,
    pub title: String,
    pub submitted_at: u64,
}

#[cw_serde]
pub struct DepositInfoResponse {
    pub depositor: String,
    pub amount: Uint128,
    pub refundable: bool,
}

#[cw_serde]
pub struct PreProposeConfigResponse {
    pub proposal_module: Addr,
    pub min_verification_status: String,
    pub deposit_amount: Uint128,
    pub deposit_denom: String,
}