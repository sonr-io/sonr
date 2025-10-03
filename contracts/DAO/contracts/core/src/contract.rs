use cosmwasm_std::{
    entry_point, to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, StdResult,
    Addr, CosmosMsg, Uint128, BankMsg, WasmMsg,
};
use cw2::set_contract_version;
use cw_storage_plus::{Item, Map};

use identity_dao_shared::{
    ContractError, CoreInstantiateMsg, CoreExecuteMsg, CoreQueryMsg,
    DaoConfigResponse, TreasuryInfo, ModuleConfig, DaoStatsResponse,
    VotingConfig,
};

// Contract name and version
const CONTRACT_NAME: &str = "identity-dao-core";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

// State storage
pub const CONFIG: Item<Config> = Item::new("config");
pub const MODULES: Item<ModuleAddresses> = Item::new("modules");
pub const TREASURY: Item<TreasuryState> = Item::new("treasury");
pub const PROPOSAL_COUNT: Item<u64> = Item::new("proposal_count");
pub const EXECUTED_PROPOSALS: Map<u64, bool> = Map::new("executed_proposals");

/// Core configuration
#[derive(serde::Serialize, serde::Deserialize, Clone, Debug, PartialEq)]
pub struct Config {
    pub name: String,
    pub description: String,
    pub voting_config: VotingConfig,
    pub admin: Option<Addr>,
    pub did_integration_enabled: bool,
}

/// Module addresses
#[derive(serde::Serialize, serde::Deserialize, Clone, Debug, PartialEq)]
pub struct ModuleAddresses {
    pub voting_module: Option<Addr>,
    pub proposal_module: Option<Addr>,
    pub pre_propose_module: Option<Addr>,
}

/// Treasury state
#[derive(serde::Serialize, serde::Deserialize, Clone, Debug, PartialEq)]
pub struct TreasuryState {
    pub balance: Uint128,
    pub reserved: Uint128,
}

/// Instantiate contract
#[entry_point]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    msg: CoreInstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    let config = Config {
        name: msg.name,
        description: msg.description,
        voting_config: msg.voting_config,
        admin: msg.admin.map(|a| deps.api.addr_validate(&a)).transpose()?,
        did_integration_enabled: msg.enable_did_integration,
    };

    let modules = ModuleAddresses {
        voting_module: None,
        proposal_module: None,
        pre_propose_module: None,
    };

    let treasury = TreasuryState {
        balance: Uint128::zero(),
        reserved: Uint128::zero(),
    };

    CONFIG.save(deps.storage, &config)?;
    MODULES.save(deps.storage, &modules)?;
    TREASURY.save(deps.storage, &treasury)?;
    PROPOSAL_COUNT.save(deps.storage, &0u64)?;

    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("name", config.name)
        .add_attribute("admin", info.sender.to_string()))
}

/// Execute contract messages
#[entry_point]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: CoreExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        CoreExecuteMsg::ExecuteProposal { proposal_id } => {
            execute_proposal(deps, env, info, proposal_id)
        }
        CoreExecuteMsg::UpdateConfig { voting_config } => {
            update_config(deps, info, voting_config)
        }
        CoreExecuteMsg::UpdateModules {
            voting_module,
            proposal_module,
            pre_propose_module,
        } => update_modules(deps, info, voting_module, proposal_module, pre_propose_module),
        CoreExecuteMsg::TransferFunds { recipient, amount } => {
            transfer_funds(deps, info, recipient, amount)
        }
    }
}

/// Execute a passed proposal
fn execute_proposal(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    proposal_id: u64,
) -> Result<Response, ContractError> {
    // Verify caller is the proposal module
    let modules = MODULES.load(deps.storage)?;
    let proposal_module = modules
        .proposal_module
        .ok_or(ContractError::Unauthorized {})?;
    
    if info.sender != proposal_module {
        return Err(ContractError::Unauthorized {});
    }

    // Check if already executed
    if EXECUTED_PROPOSALS.may_load(deps.storage, proposal_id)?.unwrap_or(false) {
        return Err(ContractError::CustomError {
            msg: "Proposal already executed".to_string(),
        });
    }

    // Mark as executed
    EXECUTED_PROPOSALS.save(deps.storage, proposal_id, &true)?;

    Ok(Response::new()
        .add_attribute("method", "execute_proposal")
        .add_attribute("proposal_id", proposal_id.to_string()))
}

/// Update voting configuration
fn update_config(
    deps: DepsMut,
    info: MessageInfo,
    voting_config: VotingConfig,
) -> Result<Response, ContractError> {
    let mut config = CONFIG.load(deps.storage)?;
    
    // Check admin permission
    if let Some(admin) = &config.admin {
        if info.sender != *admin {
            return Err(ContractError::Unauthorized {});
        }
    } else {
        return Err(ContractError::Unauthorized {});
    }

    config.voting_config = voting_config;
    CONFIG.save(deps.storage, &config)?;

    Ok(Response::new()
        .add_attribute("method", "update_config")
        .add_attribute("admin", info.sender.to_string()))
}

/// Update module addresses
fn update_modules(
    deps: DepsMut,
    info: MessageInfo,
    voting_module: Option<String>,
    proposal_module: Option<String>,
    pre_propose_module: Option<String>,
) -> Result<Response, ContractError> {
    let config = CONFIG.load(deps.storage)?;
    
    // Check admin permission
    if let Some(admin) = &config.admin {
        if info.sender != *admin {
            return Err(ContractError::Unauthorized {});
        }
    } else {
        return Err(ContractError::Unauthorized {});
    }

    let mut modules = MODULES.load(deps.storage)?;

    if let Some(addr) = voting_module {
        modules.voting_module = Some(deps.api.addr_validate(&addr)?);
    }
    if let Some(addr) = proposal_module {
        modules.proposal_module = Some(deps.api.addr_validate(&addr)?);
    }
    if let Some(addr) = pre_propose_module {
        modules.pre_propose_module = Some(deps.api.addr_validate(&addr)?);
    }

    MODULES.save(deps.storage, &modules)?;

    Ok(Response::new()
        .add_attribute("method", "update_modules")
        .add_attribute("admin", info.sender.to_string()))
}

/// Transfer funds from treasury
fn transfer_funds(
    deps: DepsMut,
    info: MessageInfo,
    recipient: String,
    amount: Uint128,
) -> Result<Response, ContractError> {
    // Verify caller is the core module itself (called via proposal execution)
    let modules = MODULES.load(deps.storage)?;
    let proposal_module = modules
        .proposal_module
        .ok_or(ContractError::Unauthorized {})?;
    
    if info.sender != proposal_module {
        return Err(ContractError::Unauthorized {});
    }

    let mut treasury = TREASURY.load(deps.storage)?;
    
    if treasury.balance < amount {
        return Err(ContractError::CustomError {
            msg: "Insufficient treasury balance".to_string(),
        });
    }

    treasury.balance -= amount;
    TREASURY.save(deps.storage, &treasury)?;

    let recipient_addr = deps.api.addr_validate(&recipient)?;
    
    Ok(Response::new()
        .add_message(BankMsg::Send {
            to_address: recipient_addr.to_string(),
            amount: vec![cosmwasm_std::Coin {
                denom: "usnr".to_string(),
                amount,
            }],
        })
        .add_attribute("method", "transfer_funds")
        .add_attribute("recipient", recipient)
        .add_attribute("amount", amount.to_string()))
}

/// Query contract state
#[entry_point]
pub fn query(deps: Deps, _env: Env, msg: CoreQueryMsg) -> StdResult<Binary> {
    match msg {
        CoreQueryMsg::Config {} => to_json_binary(&query_config(deps)?),
        CoreQueryMsg::Treasury {} => to_json_binary(&query_treasury(deps)?),
        CoreQueryMsg::Modules {} => to_json_binary(&query_modules(deps)?),
        CoreQueryMsg::Stats {} => to_json_binary(&query_stats(deps)?),
    }
}

/// Query DAO configuration
fn query_config(deps: Deps) -> StdResult<DaoConfigResponse> {
    let config = CONFIG.load(deps.storage)?;
    Ok(DaoConfigResponse {
        name: config.name,
        description: config.description,
        voting_config: config.voting_config,
        admin: config.admin,
        did_integration_enabled: config.did_integration_enabled,
    })
}

/// Query treasury information
fn query_treasury(deps: Deps) -> StdResult<TreasuryInfo> {
    let treasury = TREASURY.load(deps.storage)?;
    let addr = deps.api.addr_validate(deps.api.addr_humanize(&deps.api.addr_canonicalize(
        &deps.querier.query_bonded_denom()?.to_string()
    )?)?.as_str())?;
    
    Ok(TreasuryInfo {
        address: addr,
        balance: treasury.balance,
        reserved: treasury.reserved,
    })
}

/// Query module addresses
fn query_modules(deps: Deps) -> StdResult<ModuleConfig> {
    let modules = MODULES.load(deps.storage)?;
    let dao_core = deps.api.addr_validate(
        &deps.querier.query_bonded_denom()?.to_string()
    )?;
    
    Ok(ModuleConfig {
        dao_core,
        voting_module: modules.voting_module.unwrap_or(Addr::unchecked("")),
        proposal_module: modules.proposal_module.unwrap_or(Addr::unchecked("")),
        pre_propose_module: modules.pre_propose_module.unwrap_or(Addr::unchecked("")),
        did_integration_enabled: CONFIG.load(deps.storage)?.did_integration_enabled,
    })
}

/// Query DAO statistics
fn query_stats(deps: Deps) -> StdResult<DaoStatsResponse> {
    let proposal_count = PROPOSAL_COUNT.load(deps.storage)?;
    let treasury = TREASURY.load(deps.storage)?;
    
    Ok(DaoStatsResponse {
        total_proposals: proposal_count,
        active_proposals: 0, // Would query from proposal module
        total_voters: 0, // Would query from voting module
        treasury_balance: treasury.balance,
    })
}