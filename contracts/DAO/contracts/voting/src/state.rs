use cosmwasm_std::{Addr, Uint128};
use cw_storage_plus::{Item, Map};
use identity_dao_shared::{IdentityVoter, VoteInfo};
use serde::{Deserialize, Serialize};

/// Voting module configuration
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub struct Config {
    pub dao_core: Addr,
    pub min_verification_level: u8,
    pub use_reputation_weight: bool,
}

/// State storage items
pub const CONFIG: Item<Config> = Item::new("config");
pub const VOTERS: Map<&str, IdentityVoter> = Map::new("voters");
pub const VOTES: Map<(u64, &str), VoteInfo> = Map::new("votes");
pub const TOTAL_POWER: Item<Uint128> = Item::new("total_power");
pub const PROPOSAL_VOTERS: Map<u64, Vec<String>> = Map::new("proposal_voters");