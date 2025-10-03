use cosmwasm_std::{Addr, Uint128};
use cw_storage_plus::{Item, Map};
use identity_dao_shared::VotingConfig;
use serde::{Deserialize, Serialize};

/// Core configuration
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub struct Config {
    pub name: String,
    pub description: String,
    pub voting_config: VotingConfig,
    pub admin: Option<Addr>,
    pub did_integration_enabled: bool,
}

/// Module addresses
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub struct ModuleAddresses {
    pub voting_module: Option<Addr>,
    pub proposal_module: Option<Addr>,
    pub pre_propose_module: Option<Addr>,
}

/// Treasury state
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub struct TreasuryState {
    pub balance: Uint128,
    pub reserved: Uint128,
}

/// State storage items
pub const CONFIG: Item<Config> = Item::new("config");
pub const MODULES: Item<ModuleAddresses> = Item::new("modules");
pub const TREASURY: Item<TreasuryState> = Item::new("treasury");
pub const PROPOSAL_COUNT: Item<u64> = Item::new("proposal_count");
pub const EXECUTED_PROPOSALS: Map<u64, bool> = Map::new("executed_proposals");