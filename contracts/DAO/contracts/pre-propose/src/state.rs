use cosmwasm_std::{Addr, Uint128};
use cw_storage_plus::{Item, Map};
use identity_dao_shared::{VerificationStatus, ProposalMessage};
use serde::{Deserialize, Serialize};

/// Pre-propose module configuration
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub struct Config {
    pub proposal_module: Addr,
    pub min_verification_status: VerificationStatus,
    pub deposit_amount: Uint128,
    pub deposit_denom: String,
    pub refund_failed_proposals: bool,
}

/// Pending proposal data
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub struct PendingProposalData {
    pub id: u64,
    pub proposer: Addr,
    pub title: String,
    pub description: String,
    pub messages: Vec<ProposalMessage>,
    pub submitted_at: u64,
    pub deposit_amount: Uint128,
}

/// Deposit data
#[derive(Serialize, Deserialize, Clone, Debug, PartialEq)]
pub struct DepositData {
    pub depositor: Addr,
    pub amount: Uint128,
    pub refundable: bool,
    pub proposal_id: Option<u64>,
}

/// State storage items
pub const CONFIG: Item<Config> = Item::new("config");
pub const PENDING_COUNT: Item<u64> = Item::new("pending_count");
pub const PENDING_PROPOSALS: Map<u64, PendingProposalData> = Map::new("pending_proposals");
pub const DEPOSITS: Map<&str, DepositData> = Map::new("deposits");