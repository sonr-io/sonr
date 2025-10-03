use cosmwasm_std::Addr;
use cw_storage_plus::{Item, Map};
use identity_dao_shared::{ProposalStatus, ProposalMessage, ProposalVotes};
use serde::{Deserialize, Serialize};

/// Proposal module configuration
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub struct Config {
    pub dao_core: Addr,
    pub voting_module: Addr,
    pub pre_propose_module: Option<Addr>,
    pub allow_multiple_choice: bool,
    pub voting_period: u64,
    pub execution_delay: u64,
}

/// Proposal data
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub struct Proposal {
    pub id: u64,
    pub title: String,
    pub description: String,
    pub proposer: Addr,
    pub messages: Vec<ProposalMessage>,
    pub status: ProposalStatus,
    pub start_time: u64,
    pub end_time: u64,
    pub execution_time: Option<u64>,
}

/// State storage items
pub const CONFIG: Item<Config> = Item::new("config");
pub const PROPOSAL_COUNT: Item<u64> = Item::new("proposal_count");
pub const PROPOSALS: Map<u64, Proposal> = Map::new("proposals");
pub const PROPOSAL_VOTES: Map<u64, ProposalVotes> = Map::new("proposal_votes");