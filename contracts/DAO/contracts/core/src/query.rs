use cosmwasm_schema::{cw_serde, QueryResponses};
use cosmwasm_std::{Addr, Uint128};
use identity_dao_shared::{VotingConfig, TreasuryInfo, ModuleConfig};

/// Query messages
#[cw_serde]
#[derive(QueryResponses)]
pub enum QueryMsg {
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

/// DAO configuration response
#[cw_serde]
pub struct DaoConfigResponse {
    pub name: String,
    pub description: String,
    pub voting_config: VotingConfig,
    pub admin: Option<Addr>,
    pub did_integration_enabled: bool,
}

/// DAO statistics response
#[cw_serde]
pub struct DaoStatsResponse {
    pub total_proposals: u64,
    pub active_proposals: u64,
    pub total_voters: u64,
    pub treasury_balance: Uint128,
}