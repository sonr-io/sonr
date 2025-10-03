use cosmwasm_schema::cw_serde;
use cosmwasm_std::Uint128;
use identity_dao_shared::VotingConfig;

/// Instantiate message
#[cw_serde]
pub struct InstantiateMsg {
    pub name: String,
    pub description: String,
    pub voting_config: VotingConfig,
    pub admin: Option<String>,
    pub enable_did_integration: bool,
}

/// Execute messages
#[cw_serde]
pub enum ExecuteMsg {
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