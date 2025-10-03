use cosmwasm_schema::cw_serde;
use cosmwasm_std::{Addr, Binary, Uint128};
use crate::types::{Vote, VotingConfig, VerificationStatus};

/// Core DAO instantiate message
#[cw_serde]
pub struct CoreInstantiateMsg {
    /// Name of the DAO
    pub name: String,
    /// Description of the DAO
    pub description: String,
    /// Initial voting configuration
    pub voting_config: VotingConfig,
    /// Admin address (optional)
    pub admin: Option<String>,
    /// Enable x/did integration
    pub enable_did_integration: bool,
}

/// Core DAO execute messages
#[cw_serde]
pub enum CoreExecuteMsg {
    /// Execute a proposal
    ExecuteProposal { proposal_id: u64 },
    /// Update voting configuration
    UpdateConfig { voting_config: VotingConfig },
    /// Update module addresses
    UpdateModules {
        voting_module: Option<String>,
        proposal_module: Option<String>,
        pre_propose_module: Option<String>,
    },
    /// Transfer treasury funds
    TransferFunds {
        recipient: String,
        amount: Uint128,
    },
}

/// Voting module instantiate message
#[cw_serde]
pub struct VotingInstantiateMsg {
    /// Core DAO contract address
    pub dao_core: String,
    /// Minimum verification level required to vote
    pub min_verification_level: u8,
    /// Enable reputation-based voting weight
    pub use_reputation_weight: bool,
}

/// Voting module execute messages
#[cw_serde]
pub enum VotingExecuteMsg {
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
    RemoveVoter { did: String },
}

/// Proposal module instantiate message
#[cw_serde]
pub struct ProposalInstantiateMsg {
    /// Core DAO contract address
    pub dao_core: String,
    /// Voting module address
    pub voting_module: String,
    /// Pre-propose module address
    pub pre_propose_module: Option<String>,
    /// Allow multiple choice proposals
    pub allow_multiple_choice: bool,
}

/// Proposal module execute messages
#[cw_serde]
pub enum ProposalExecuteMsg {
    /// Create a new proposal
    Propose {
        title: String,
        description: String,
        msgs: Vec<ProposalMessage>,
    },
    /// Execute a passed proposal
    Execute { proposal_id: u64 },
    /// Close an expired proposal
    Close { proposal_id: u64 },
    /// Update proposal status
    UpdateStatus { 
        proposal_id: u64,
        status: ProposalStatusUpdate,
    },
}

/// Pre-propose module instantiate message
#[cw_serde]
pub struct PreProposeInstantiateMsg {
    /// Proposal module address
    pub proposal_module: String,
    /// Minimum verification status required
    pub min_verification_status: VerificationStatus,
    /// Deposit required for proposal
    pub deposit_amount: Uint128,
    /// Deposit denom
    pub deposit_denom: String,
}

/// Pre-propose module execute messages
#[cw_serde]
pub enum PreProposeExecuteMsg {
    /// Submit a proposal for approval
    SubmitProposal {
        title: String,
        description: String,
        msgs: Vec<ProposalMessage>,
    },
    /// Approve a pending proposal
    ApproveProposal { proposal_id: u64 },
    /// Reject a pending proposal
    RejectProposal { 
        proposal_id: u64,
        reason: String,
    },
    /// Withdraw a pending proposal
    WithdrawProposal { proposal_id: u64 },
}

/// Message to be executed by a proposal
#[cw_serde]
pub struct ProposalMessage {
    /// Contract address to execute on
    pub contract: String,
    /// Message to execute
    pub msg: Binary,
    /// Funds to send with the message
    pub funds: Vec<cosmwasm_std::Coin>,
}

/// Proposal status update
#[cw_serde]
pub enum ProposalStatusUpdate {
    /// Mark as passed
    Passed,
    /// Mark as rejected
    Rejected,
    /// Mark as executed
    Executed,
    /// Mark as failed
    ExecutionFailed { reason: String },
}