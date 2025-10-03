use cosmwasm_std::{
    entry_point, to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, 
    StdResult, Addr, Uint128, CosmosMsg, WasmMsg, Order,
};
use cw2::set_contract_version;
use cw_storage_plus::{Item, Map};

use identity_dao_shared::{
    ContractError, ProposalInstantiateMsg, ProposalExecuteMsg, ProposalQueryMsg,
    ProposalResponse, ProposalsListResponse, ProposalVotesResponse, ProposalResultResponse,
    ProposalStatus, ProposalMessage, ProposalResult, ProposalVotes, VoteInfo,
    ProposalStatusUpdate,
};

// Contract name and version
const CONTRACT_NAME: &str = "identity-dao-proposals";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

// State storage
pub const CONFIG: Item<Config> = Item::new("config");
pub const PROPOSAL_COUNT: Item<u64> = Item::new("proposal_count");
pub const PROPOSALS: Map<u64, Proposal> = Map::new("proposals");
pub const PROPOSAL_VOTES: Map<u64, ProposalVotes> = Map::new("proposal_votes");

/// Proposal module configuration
#[derive(serde::Serialize, serde::Deserialize, Clone, Debug, PartialEq)]
pub struct Config {
    pub dao_core: Addr,
    pub voting_module: Addr,
    pub pre_propose_module: Option<Addr>,
    pub allow_multiple_choice: bool,
    pub voting_period: u64,
    pub execution_delay: u64,
}

/// Proposal data
#[derive(serde::Serialize, serde::Deserialize, Clone, Debug, PartialEq)]
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

/// Instantiate contract
#[entry_point]
pub fn instantiate(
    deps: DepsMut,
    env: Env,
    _info: MessageInfo,
    msg: ProposalInstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    let config = Config {
        dao_core: deps.api.addr_validate(&msg.dao_core)?,
        voting_module: deps.api.addr_validate(&msg.voting_module)?,
        pre_propose_module: msg.pre_propose_module
            .map(|a| deps.api.addr_validate(&a))
            .transpose()?,
        allow_multiple_choice: msg.allow_multiple_choice,
        voting_period: 7 * 24 * 60 * 60, // 7 days
        execution_delay: 24 * 60 * 60,   // 1 day
    };

    CONFIG.save(deps.storage, &config)?;
    PROPOSAL_COUNT.save(deps.storage, &0u64)?;

    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("dao_core", config.dao_core.to_string())
        .add_attribute("voting_module", config.voting_module.to_string()))
}

/// Execute contract messages
#[entry_point]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: ProposalExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        ProposalExecuteMsg::Propose { title, description, msgs } => {
            execute_propose(deps, env, info, title, description, msgs)
        }
        ProposalExecuteMsg::Execute { proposal_id } => {
            execute_proposal(deps, env, info, proposal_id)
        }
        ProposalExecuteMsg::Close { proposal_id } => {
            execute_close(deps, env, info, proposal_id)
        }
        ProposalExecuteMsg::UpdateStatus { proposal_id, status } => {
            execute_update_status(deps, env, info, proposal_id, status)
        }
    }
}

/// Create a new proposal
fn execute_propose(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    title: String,
    description: String,
    msgs: Vec<ProposalMessage>,
) -> Result<Response, ContractError> {
    let config = CONFIG.load(deps.storage)?;
    
    // Check if pre-propose module is set and sender is authorized
    if let Some(pre_propose) = &config.pre_propose_module {
        if info.sender != *pre_propose {
            return Err(ContractError::Unauthorized {});
        }
    }

    // Increment proposal count
    let proposal_id = PROPOSAL_COUNT.load(deps.storage)? + 1;
    PROPOSAL_COUNT.save(deps.storage, &proposal_id)?;

    // Create proposal
    let proposal = Proposal {
        id: proposal_id,
        title,
        description,
        proposer: info.sender.clone(),
        messages: msgs,
        status: ProposalStatus::Open,
        start_time: env.block.time.seconds(),
        end_time: env.block.time.seconds() + config.voting_period,
        execution_time: None,
    };

    // Initialize vote counts
    let votes = ProposalVotes {
        yes: Uint128::zero(),
        no: Uint128::zero(),
        abstain: Uint128::zero(),
        no_with_veto: Uint128::zero(),
    };

    PROPOSALS.save(deps.storage, proposal_id, &proposal)?;
    PROPOSAL_VOTES.save(deps.storage, proposal_id, &votes)?;

    Ok(Response::new()
        .add_attribute("method", "propose")
        .add_attribute("proposal_id", proposal_id.to_string())
        .add_attribute("proposer", info.sender.to_string())
        .add_attribute("title", proposal.title))
}

/// Execute a passed proposal
fn execute_proposal(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    proposal_id: u64,
) -> Result<Response, ContractError> {
    let mut proposal = PROPOSALS.load(deps.storage, proposal_id)?;
    
    // Check if proposal has passed
    if proposal.status != ProposalStatus::Passed {
        return Err(ContractError::CustomError {
            msg: "Proposal has not passed".to_string(),
        });
    }

    // Check execution delay
    if let Some(exec_time) = proposal.execution_time {
        if env.block.time.seconds() < exec_time {
            return Err(ContractError::CustomError {
                msg: "Execution delay not met".to_string(),
            });
        }
    }

    // Update status
    proposal.status = ProposalStatus::Executed;
    PROPOSALS.save(deps.storage, proposal_id, &proposal)?;

    // Execute messages
    let mut messages = vec![];
    for msg in proposal.messages {
        messages.push(CosmosMsg::Wasm(WasmMsg::Execute {
            contract_addr: msg.contract,
            msg: msg.msg,
            funds: msg.funds,
        }));
    }

    // Notify core module
    let core_msg = WasmMsg::Execute {
        contract_addr: CONFIG.load(deps.storage)?.dao_core.to_string(),
        msg: to_json_binary(&identity_dao_shared::CoreExecuteMsg::ExecuteProposal { 
            proposal_id 
        })?,
        funds: vec![],
    };
    messages.push(CosmosMsg::Wasm(core_msg));

    Ok(Response::new()
        .add_messages(messages)
        .add_attribute("method", "execute_proposal")
        .add_attribute("proposal_id", proposal_id.to_string())
        .add_attribute("executor", info.sender.to_string()))
}

/// Close an expired proposal
fn execute_close(
    deps: DepsMut,
    env: Env,
    _info: MessageInfo,
    proposal_id: u64,
) -> Result<Response, ContractError> {
    let mut proposal = PROPOSALS.load(deps.storage, proposal_id)?;
    
    // Check if voting period has ended
    if env.block.time.seconds() < proposal.end_time {
        return Err(ContractError::VotingPeriodNotEnded {});
    }

    // Check if proposal is still open
    if proposal.status != ProposalStatus::Open {
        return Err(ContractError::CustomError {
            msg: "Proposal is not open".to_string(),
        });
    }

    // Determine result based on votes
    let votes = PROPOSAL_VOTES.load(deps.storage, proposal_id)?;
    let total_votes = votes.yes + votes.no + votes.abstain + votes.no_with_veto;
    
    // Get voting config from core
    let config = CONFIG.load(deps.storage)?;
    
    // Simple majority for now (would query from core for actual config)
    let threshold = Uint128::from(50u128);
    let quorum = Uint128::from(33u128);
    
    let participation = if total_votes.is_zero() {
        Uint128::zero()
    } else {
        total_votes * Uint128::from(100u128) / total_votes // Would get total power from voting module
    };

    if participation < quorum {
        proposal.status = ProposalStatus::Rejected;
    } else if votes.yes * Uint128::from(100u128) > total_votes * threshold {
        proposal.status = ProposalStatus::Passed;
        proposal.execution_time = Some(env.block.time.seconds() + config.execution_delay);
    } else {
        proposal.status = ProposalStatus::Rejected;
    }

    PROPOSALS.save(deps.storage, proposal_id, &proposal)?;

    Ok(Response::new()
        .add_attribute("method", "close")
        .add_attribute("proposal_id", proposal_id.to_string())
        .add_attribute("status", format!("{:?}", proposal.status)))
}

/// Update proposal status
fn execute_update_status(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    proposal_id: u64,
    status: ProposalStatusUpdate,
) -> Result<Response, ContractError> {
    let config = CONFIG.load(deps.storage)?;
    
    // Only voting module can update status
    if info.sender != config.voting_module {
        return Err(ContractError::Unauthorized {});
    }

    let mut proposal = PROPOSALS.load(deps.storage, proposal_id)?;
    
    proposal.status = match status {
        ProposalStatusUpdate::Passed => ProposalStatus::Passed,
        ProposalStatusUpdate::Rejected => ProposalStatus::Rejected,
        ProposalStatusUpdate::Executed => ProposalStatus::Executed,
        ProposalStatusUpdate::ExecutionFailed { .. } => ProposalStatus::ExecutionFailed,
    };

    PROPOSALS.save(deps.storage, proposal_id, &proposal)?;

    Ok(Response::new()
        .add_attribute("method", "update_status")
        .add_attribute("proposal_id", proposal_id.to_string())
        .add_attribute("status", format!("{:?}", proposal.status)))
}

/// Query contract state
#[entry_point]
pub fn query(deps: Deps, _env: Env, msg: ProposalQueryMsg) -> StdResult<Binary> {
    match msg {
        ProposalQueryMsg::Proposal { proposal_id } => {
            to_json_binary(&query_proposal(deps, proposal_id)?)
        }
        ProposalQueryMsg::ListProposals { status, start_after, limit } => {
            to_json_binary(&query_list_proposals(deps, status, start_after, limit)?)
        }
        ProposalQueryMsg::ProposalVotes { proposal_id, start_after, limit } => {
            to_json_binary(&query_proposal_votes(deps, proposal_id, start_after, limit)?)
        }
        ProposalQueryMsg::ProposalResult { proposal_id } => {
            to_json_binary(&query_proposal_result(deps, proposal_id)?)
        }
    }
}

/// Query proposal details
fn query_proposal(deps: Deps, proposal_id: u64) -> StdResult<ProposalResponse> {
    let proposal = PROPOSALS.load(deps.storage, proposal_id)?;
    let votes = PROPOSAL_VOTES.load(deps.storage, proposal_id)?;
    
    Ok(ProposalResponse {
        id: proposal.id,
        title: proposal.title,
        description: proposal.description,
        proposer: proposal.proposer.to_string(),
        status: proposal.status,
        votes,
        start_time: proposal.start_time,
        end_time: proposal.end_time,
    })
}

/// List proposals with filters
fn query_list_proposals(
    deps: Deps,
    status: Option<ProposalStatus>,
    start_after: Option<u64>,
    limit: Option<u32>,
) -> StdResult<ProposalsListResponse> {
    let limit = limit.unwrap_or(30).min(100) as usize;
    let start = start_after.map(|id| cosmwasm_std::Bound::exclusive(id));
    
    let proposals: Vec<ProposalResponse> = PROPOSALS
        .range(deps.storage, start, None, Order::Ascending)
        .filter(|item| {
            if let Ok((_, proposal)) = item {
                status.is_none() || status == Some(proposal.status.clone())
            } else {
                false
            }
        })
        .take(limit)
        .map(|item| {
            let (id, proposal) = item?;
            let votes = PROPOSAL_VOTES.load(deps.storage, id)?;
            Ok(ProposalResponse {
                id: proposal.id,
                title: proposal.title,
                description: proposal.description,
                proposer: proposal.proposer.to_string(),
                status: proposal.status,
                votes,
                start_time: proposal.start_time,
                end_time: proposal.end_time,
            })
        })
        .collect::<StdResult<Vec<_>>>()?;
    
    let total = PROPOSAL_COUNT.load(deps.storage)?;
    
    Ok(ProposalsListResponse { proposals, total })
}

/// Query proposal votes (placeholder - would integrate with voting module)
fn query_proposal_votes(
    _deps: Deps,
    proposal_id: u64,
    _start_after: Option<String>,
    _limit: Option<u32>,
) -> StdResult<ProposalVotesResponse> {
    Ok(ProposalVotesResponse {
        votes: vec![],
        total: 0,
    })
}

/// Query proposal result
fn query_proposal_result(deps: Deps, proposal_id: u64) -> StdResult<ProposalResultResponse> {
    let proposal = PROPOSALS.load(deps.storage, proposal_id)?;
    
    let result = match proposal.status {
        ProposalStatus::Passed | ProposalStatus::Executed => ProposalResult::Passed,
        ProposalStatus::Rejected | ProposalStatus::ExecutionFailed => ProposalResult::Rejected,
        _ => ProposalResult::InProgress,
    };
    
    Ok(ProposalResultResponse {
        proposal_id,
        result,
    })
}