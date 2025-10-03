use cosmwasm_schema::cw_serde;
use identity_dao_shared::{ProposalMessage, ProposalStatusUpdate};

/// Instantiate message
#[cw_serde]
pub struct InstantiateMsg {
    pub dao_core: String,
    pub voting_module: String,
    pub pre_propose_module: Option<String>,
    pub allow_multiple_choice: bool,
}

/// Execute messages
#[cw_serde]
pub enum ExecuteMsg {
    /// Create a new proposal
    Propose {
        title: String,
        description: String,
        msgs: Vec<ProposalMessage>,
    },
    /// Execute a passed proposal
    Execute { 
        proposal_id: u64 
    },
    /// Close an expired proposal
    Close { 
        proposal_id: u64 
    },
    /// Update proposal status
    UpdateStatus { 
        proposal_id: u64,
        status: ProposalStatusUpdate,
    },
}