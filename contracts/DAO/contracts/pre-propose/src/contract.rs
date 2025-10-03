use cosmwasm_std::{
    entry_point, to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, 
    StdResult, Addr, Uint128, CosmosMsg, WasmMsg, BankMsg, Coin,
};
use cw2::set_contract_version;
use cw_storage_plus::{Item, Map};

use identity_dao_shared::{
    ContractError, PreProposeInstantiateMsg, PreProposeExecuteMsg, PreProposeQueryMsg,
    PendingProposalsResponse, PendingProposal, DepositInfoResponse, PreProposeConfigResponse,
    VerificationStatus, ProposalMessage,
};
use crate::verification::{verify_did_status, check_verification_requirements};

// Contract name and version
const CONTRACT_NAME: &str = "identity-dao-pre-propose";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

// State storage
pub const CONFIG: Item<Config> = Item::new("config");
pub const PENDING_COUNT: Item<u64> = Item::new("pending_count");
pub const PENDING_PROPOSALS: Map<u64, PendingProposalData> = Map::new("pending_proposals");
pub const DEPOSITS: Map<&str, DepositData> = Map::new("deposits");

/// Pre-propose module configuration
#[derive(serde::Serialize, serde::Deserialize, Clone, Debug, PartialEq)]
pub struct Config {
    pub proposal_module: Addr,
    pub min_verification_status: VerificationStatus,
    pub deposit_amount: Uint128,
    pub deposit_denom: String,
    pub refund_failed_proposals: bool,
}

/// Pending proposal data
#[derive(serde::Serialize, serde::Deserialize, Clone, Debug, PartialEq)]
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
#[derive(serde::Serialize, serde::Deserialize, Clone, Debug, PartialEq)]
pub struct DepositData {
    pub depositor: Addr,
    pub amount: Uint128,
    pub refundable: bool,
    pub proposal_id: Option<u64>,
}

/// Instantiate contract
#[entry_point]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: PreProposeInstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    let config = Config {
        proposal_module: deps.api.addr_validate(&msg.proposal_module)?,
        min_verification_status: msg.min_verification_status,
        deposit_amount: msg.deposit_amount,
        deposit_denom: msg.deposit_denom,
        refund_failed_proposals: true,
    };

    CONFIG.save(deps.storage, &config)?;
    PENDING_COUNT.save(deps.storage, &0u64)?;

    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("proposal_module", config.proposal_module.to_string())
        .add_attribute("deposit_amount", config.deposit_amount.to_string()))
}

/// Execute contract messages
#[entry_point]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: PreProposeExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        PreProposeExecuteMsg::SubmitProposal { title, description, msgs } => {
            execute_submit_proposal(deps, env, info, title, description, msgs)
        }
        PreProposeExecuteMsg::ApproveProposal { proposal_id } => {
            execute_approve_proposal(deps, env, info, proposal_id)
        }
        PreProposeExecuteMsg::RejectProposal { proposal_id, reason } => {
            execute_reject_proposal(deps, env, info, proposal_id, reason)
        }
        PreProposeExecuteMsg::WithdrawProposal { proposal_id } => {
            execute_withdraw_proposal(deps, env, info, proposal_id)
        }
    }
}

/// Submit a proposal for approval
fn execute_submit_proposal(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    title: String,
    description: String,
    msgs: Vec<ProposalMessage>,
) -> Result<Response, ContractError> {
    let config = CONFIG.load(deps.storage)?;
    
    // Verify DID status
    let did = format!("did:sonr:{}", info.sender);
    let verification = verify_did_status(deps.as_ref(), &did)?;
    
    if !verification.is_verified {
        return Err(ContractError::DIDNotVerified {});
    }
    
    // Check minimum verification level
    let status = match verification.verification_level {
        0 => VerificationStatus::Unverified,
        1 => VerificationStatus::Basic,
        2 => VerificationStatus::Advanced,
        _ => VerificationStatus::Full,
    };
    
    if (status as u8) < (config.min_verification_status as u8) {
        return Err(ContractError::CustomError {
            msg: "Insufficient verification level".to_string(),
        });
    }
    
    // Check deposit
    let deposit_paid = info
        .funds
        .iter()
        .find(|c| c.denom == config.deposit_denom)
        .map(|c| c.amount)
        .unwrap_or(Uint128::zero());
    
    if deposit_paid < config.deposit_amount {
        return Err(ContractError::CustomError {
            msg: "Insufficient deposit".to_string(),
        });
    }
    
    // Create pending proposal
    let proposal_id = PENDING_COUNT.load(deps.storage)? + 1;
    PENDING_COUNT.save(deps.storage, &proposal_id)?;
    
    let pending = PendingProposalData {
        id: proposal_id,
        proposer: info.sender.clone(),
        title,
        description,
        messages: msgs,
        submitted_at: env.block.time.seconds(),
        deposit_amount: deposit_paid,
    };
    
    PENDING_PROPOSALS.save(deps.storage, proposal_id, &pending)?;
    
    // Save deposit info
    let deposit = DepositData {
        depositor: info.sender.clone(),
        amount: deposit_paid,
        refundable: config.refund_failed_proposals,
        proposal_id: Some(proposal_id),
    };
    
    DEPOSITS.save(deps.storage, info.sender.as_str(), &deposit)?;
    
    Ok(Response::new()
        .add_attribute("method", "submit_proposal")
        .add_attribute("proposal_id", proposal_id.to_string())
        .add_attribute("proposer", info.sender.to_string())
        .add_attribute("title", pending.title))
}

/// Approve a pending proposal
fn execute_approve_proposal(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    proposal_id: u64,
) -> Result<Response, ContractError> {
    let config = CONFIG.load(deps.storage)?;
    
    // Only admin or governance can approve
    // For now, allow anyone (would check permissions in production)
    
    let pending = PENDING_PROPOSALS.load(deps.storage, proposal_id)?;
    
    // Remove from pending
    PENDING_PROPOSALS.remove(deps.storage, proposal_id);
    
    // Create proposal in proposal module
    let propose_msg = WasmMsg::Execute {
        contract_addr: config.proposal_module.to_string(),
        msg: to_json_binary(&identity_dao_shared::ProposalExecuteMsg::Propose {
            title: pending.title,
            description: pending.description,
            msgs: pending.messages,
        })?,
        funds: vec![],
    };
    
    Ok(Response::new()
        .add_message(propose_msg)
        .add_attribute("method", "approve_proposal")
        .add_attribute("proposal_id", proposal_id.to_string())
        .add_attribute("approver", info.sender.to_string()))
}

/// Reject a pending proposal
fn execute_reject_proposal(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    proposal_id: u64,
    reason: String,
) -> Result<Response, ContractError> {
    let config = CONFIG.load(deps.storage)?;
    
    // Only admin or governance can reject
    // For now, allow anyone (would check permissions in production)
    
    let pending = PENDING_PROPOSALS.load(deps.storage, proposal_id)?;
    
    // Remove from pending
    PENDING_PROPOSALS.remove(deps.storage, proposal_id);
    
    // Refund deposit if configured
    let mut messages = vec![];
    if config.refund_failed_proposals {
        if let Some(mut deposit) = DEPOSITS.may_load(deps.storage, pending.proposer.as_str())? {
            deposit.refundable = false;
            DEPOSITS.save(deps.storage, pending.proposer.as_str(), &deposit)?;
            
            messages.push(CosmosMsg::Bank(BankMsg::Send {
                to_address: pending.proposer.to_string(),
                amount: vec![Coin {
                    denom: config.deposit_denom,
                    amount: pending.deposit_amount,
                }],
            }));
        }
    }
    
    Ok(Response::new()
        .add_messages(messages)
        .add_attribute("method", "reject_proposal")
        .add_attribute("proposal_id", proposal_id.to_string())
        .add_attribute("rejector", info.sender.to_string())
        .add_attribute("reason", reason))
}

/// Withdraw a pending proposal
fn execute_withdraw_proposal(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    proposal_id: u64,
) -> Result<Response, ContractError> {
    let config = CONFIG.load(deps.storage)?;
    let pending = PENDING_PROPOSALS.load(deps.storage, proposal_id)?;
    
    // Only proposer can withdraw
    if info.sender != pending.proposer {
        return Err(ContractError::Unauthorized {});
    }
    
    // Remove from pending
    PENDING_PROPOSALS.remove(deps.storage, proposal_id);
    
    // Refund deposit
    let mut messages = vec![];
    if let Some(mut deposit) = DEPOSITS.may_load(deps.storage, info.sender.as_str())? {
        deposit.refundable = false;
        deposit.proposal_id = None;
        DEPOSITS.save(deps.storage, info.sender.as_str(), &deposit)?;
        
        messages.push(CosmosMsg::Bank(BankMsg::Send {
            to_address: info.sender.to_string(),
            amount: vec![Coin {
                denom: config.deposit_denom,
                amount: pending.deposit_amount,
            }],
        }));
    }
    
    Ok(Response::new()
        .add_messages(messages)
        .add_attribute("method", "withdraw_proposal")
        .add_attribute("proposal_id", proposal_id.to_string())
        .add_attribute("proposer", info.sender.to_string()))
}

// Verification functions are now imported from the verification module

/// Query contract state
#[entry_point]
pub fn query(deps: Deps, _env: Env, msg: PreProposeQueryMsg) -> StdResult<Binary> {
    match msg {
        PreProposeQueryMsg::PendingProposals { start_after, limit } => {
            to_json_binary(&query_pending_proposals(deps, start_after, limit)?)
        }
        PreProposeQueryMsg::DepositInfo { proposer } => {
            to_json_binary(&query_deposit_info(deps, proposer)?)
        }
        PreProposeQueryMsg::Config {} => {
            to_json_binary(&query_config(deps)?)
        }
    }
}

/// Query pending proposals
fn query_pending_proposals(
    deps: Deps,
    start_after: Option<u64>,
    limit: Option<u32>,
) -> StdResult<PendingProposalsResponse> {
    let limit = limit.unwrap_or(30).min(100) as usize;
    let start = start_after.map(|id| cosmwasm_std::Bound::exclusive(id));
    
    let proposals: Vec<PendingProposal> = PENDING_PROPOSALS
        .range(deps.storage, start, None, cosmwasm_std::Order::Ascending)
        .take(limit)
        .map(|item| {
            let (_, data) = item?;
            Ok(PendingProposal {
                id: data.id,
                proposer: data.proposer.to_string(),
                title: data.title,
                submitted_at: data.submitted_at,
            })
        })
        .collect::<StdResult<Vec<_>>>()?;
    
    let total = PENDING_COUNT.load(deps.storage)?;
    
    Ok(PendingProposalsResponse { proposals, total })
}

/// Query deposit info
fn query_deposit_info(deps: Deps, proposer: String) -> StdResult<DepositInfoResponse> {
    let deposit = DEPOSITS.may_load(deps.storage, &proposer)?;
    
    if let Some(deposit) = deposit {
        Ok(DepositInfoResponse {
            depositor: deposit.depositor.to_string(),
            amount: deposit.amount,
            refundable: deposit.refundable,
        })
    } else {
        Ok(DepositInfoResponse {
            depositor: proposer,
            amount: Uint128::zero(),
            refundable: false,
        })
    }
}

/// Query module config
fn query_config(deps: Deps) -> StdResult<PreProposeConfigResponse> {
    let config = CONFIG.load(deps.storage)?;
    
    Ok(PreProposeConfigResponse {
        proposal_module: config.proposal_module,
        min_verification_status: format!("{:?}", config.min_verification_status),
        deposit_amount: config.deposit_amount,
        deposit_denom: config.deposit_denom,
    })
}