use cosmwasm_std::{
    entry_point, to_json_binary, Binary, Deps, DepsMut, Env, MessageInfo, Response, 
    StdResult, Addr, Uint128, QueryRequest, StargateResponse,
    IbcBasicResponse, IbcChannelCloseMsg, IbcChannelConnectMsg, IbcChannelOpenMsg,
    IbcChannelOpenResponse, IbcPacketAckMsg, IbcPacketReceiveMsg, IbcPacketTimeoutMsg,
    IbcReceiveResponse, Never, from_json,
};
use cw2::set_contract_version;
use cw_storage_plus::{Item, Map};

use identity_dao_shared::{
    ContractError, VotingInstantiateMsg, VotingExecuteMsg, VotingQueryMsg,
    VotingPowerResponse, TotalPowerResponse, VoterInfoResponse, VotersListResponse,
    VoteResponse, VoteInfo, Vote, IdentityVoter, VerificationStatus,
    bindings::{SonrQuery, DIDDocumentResponse, VerificationResponse},
};

// Contract name and version
const CONTRACT_NAME: &str = "identity-dao-voting";
const CONTRACT_VERSION: &str = env!("CARGO_PKG_VERSION");

/// Voting module configuration
#[derive(serde::Serialize, serde::Deserialize, Clone, Debug, PartialEq)]
pub struct Config {
    pub dao_core: Addr,
    pub min_verification_level: u8,
    pub use_reputation_weight: bool,
}

/// Pending verification from IBC query
#[derive(serde::Serialize, serde::Deserialize, Clone, Debug)]
pub struct PendingVerification {
    pub is_verified: bool,
    pub verification_level: u8,
    pub timestamp: u64,
}

// State storage
pub const CONFIG: Item<Config> = Item::new("config");
pub const VOTERS: Map<&str, IdentityVoter> = Map::new("voters");
pub const VOTES: Map<(u64, &str), VoteInfo> = Map::new("votes");
pub const TOTAL_POWER: Item<Uint128> = Item::new("total_power");
pub const PROPOSAL_VOTERS: Map<u64, Vec<String>> = Map::new("proposal_voters");
pub const IBC_CHANNEL: Item<String> = Item::new("ibc_channel");
pub const PENDING_VERIFICATIONS: Map<&str, PendingVerification> = Map::new("pending_verifications");

/// Instantiate contract
#[entry_point]
pub fn instantiate(
    deps: DepsMut,
    _env: Env,
    _info: MessageInfo,
    msg: VotingInstantiateMsg,
) -> Result<Response, ContractError> {
    set_contract_version(deps.storage, CONTRACT_NAME, CONTRACT_VERSION)?;

    let config = Config {
        dao_core: deps.api.addr_validate(&msg.dao_core)?,
        min_verification_level: msg.min_verification_level,
        use_reputation_weight: msg.use_reputation_weight,
    };

    CONFIG.save(deps.storage, &config)?;
    TOTAL_POWER.save(deps.storage, &Uint128::zero())?;

    Ok(Response::new()
        .add_attribute("method", "instantiate")
        .add_attribute("dao_core", config.dao_core.to_string()))
}

/// Execute contract messages
#[entry_point]
pub fn execute(
    deps: DepsMut,
    env: Env,
    info: MessageInfo,
    msg: VotingExecuteMsg,
) -> Result<Response, ContractError> {
    match msg {
        VotingExecuteMsg::Vote { proposal_id, vote } => {
            execute_vote(deps, env, info, proposal_id, vote)
        }
        VotingExecuteMsg::UpdateVoter { did, address } => {
            execute_update_voter(deps, env, info, did, address)
        }
        VotingExecuteMsg::RemoveVoter { did } => {
            execute_remove_voter(deps, info, did)
        }
    }
}

/// Execute vote on proposal
fn execute_vote(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    proposal_id: u64,
    vote: Vote,
) -> Result<Response, ContractError> {
    // Get voter by address
    let voter = VOTERS
        .range(deps.storage, None, None, cosmwasm_std::Order::Ascending)
        .find(|item| {
            if let Ok((_, v)) = item {
                v.address == info.sender
            } else {
                false
            }
        })
        .ok_or(ContractError::Unauthorized {})?
        .map_err(|_| ContractError::Unauthorized {})?;

    let voter_did = voter.0;
    let voter_info = voter.1;

    // Check if already voted
    if VOTES.may_load(deps.storage, (proposal_id, &voter_did))?.is_some() {
        return Err(ContractError::AlreadyVoted {});
    }

    // Create vote info
    let vote_info = VoteInfo {
        proposal_id,
        voter: voter_did.clone(),
        vote: vote.clone(),
        voting_power: voter_info.voting_power,
    };

    // Save vote
    VOTES.save(deps.storage, (proposal_id, &voter_did), &vote_info)?;

    // Update proposal voters list
    let mut voters = PROPOSAL_VOTERS.may_load(deps.storage, proposal_id)?.unwrap_or_default();
    voters.push(voter_did.clone());
    PROPOSAL_VOTERS.save(deps.storage, proposal_id, &voters)?;

    Ok(Response::new()
        .add_attribute("method", "vote")
        .add_attribute("proposal_id", proposal_id.to_string())
        .add_attribute("voter", voter_did)
        .add_attribute("vote", format!("{:?}", vote))
        .add_attribute("voting_power", voter_info.voting_power.to_string()))
}

/// Update or register a voter
fn execute_update_voter(
    deps: DepsMut,
    _env: Env,
    info: MessageInfo,
    did: String,
    address: String,
) -> Result<Response, ContractError> {
    let config = CONFIG.load(deps.storage)?;
    
    // Only DAO core can update voters
    if info.sender != config.dao_core {
        return Err(ContractError::Unauthorized {});
    }

    let voter_addr = deps.api.addr_validate(&address)?;

    // Query DID verification status using stargate
    let verification = query_did_verification(deps.as_ref(), &did)?;
    
    if !verification.is_verified {
        return Err(ContractError::DIDNotVerified {});
    }

    if verification.verification_level < config.min_verification_level {
        return Err(ContractError::InsufficientVotingPower {});
    }

    // Calculate voting power based on verification level and reputation
    let base_power = Uint128::from(100u128);
    let level_multiplier = verification.verification_level as u128;
    let voting_power = if config.use_reputation_weight {
        base_power * Uint128::from(level_multiplier)
    } else {
        base_power
    };

    // Create or update voter
    let voter = IdentityVoter {
        did: did.clone(),
        address: voter_addr,
        voting_power,
        verification_level: verification.verification_level,
        reputation_score: 0, // Would be fetched from reputation system
    };

    // Update total power
    let mut total_power = TOTAL_POWER.load(deps.storage)?;
    if let Some(existing) = VOTERS.may_load(deps.storage, &did)? {
        total_power = total_power - existing.voting_power + voting_power;
    } else {
        total_power += voting_power;
    }
    TOTAL_POWER.save(deps.storage, &total_power)?;

    VOTERS.save(deps.storage, &did, &voter)?;

    Ok(Response::new()
        .add_attribute("method", "update_voter")
        .add_attribute("did", did)
        .add_attribute("address", address)
        .add_attribute("voting_power", voting_power.to_string()))
}

/// Remove a voter
fn execute_remove_voter(
    deps: DepsMut,
    info: MessageInfo,
    did: String,
) -> Result<Response, ContractError> {
    let config = CONFIG.load(deps.storage)?;
    
    // Only DAO core can remove voters
    if info.sender != config.dao_core {
        return Err(ContractError::Unauthorized {});
    }

    // Remove voter and update total power
    if let Some(voter) = VOTERS.may_load(deps.storage, &did)? {
        let mut total_power = TOTAL_POWER.load(deps.storage)?;
        total_power -= voter.voting_power;
        TOTAL_POWER.save(deps.storage, &total_power)?;
        VOTERS.remove(deps.storage, &did);
    }

    Ok(Response::new()
        .add_attribute("method", "remove_voter")
        .add_attribute("did", did))
}

/// Query DID verification status using stargate
fn query_did_verification(deps: Deps, did: &str) -> StdResult<VerificationResponse> {
    // For now, return mock data - would use actual stargate query
    // let query = identity_dao_shared::bindings::stargate::query_did_verification(did);
    // let response: StargateResponse = deps.querier.query(&QueryRequest::Stargate {
    //     path: query.path,
    //     data: query.data.into(),
    // })?;
    
    // Mock response for testing
    Ok(VerificationResponse {
        is_verified: true,
        verification_level: 2,
        last_verified: Some(1700000000),
    })
}

/// Query contract state
#[entry_point]
pub fn query(deps: Deps, env: Env, msg: VotingQueryMsg) -> StdResult<Binary> {
    match msg {
        VotingQueryMsg::VotingPower { did } => {
            to_json_binary(&query_voting_power(deps, env, did)?)
        }
        VotingQueryMsg::TotalPower { height } => {
            to_json_binary(&query_total_power(deps, env, height)?)
        }
        VotingQueryMsg::VoterInfo { did } => {
            to_json_binary(&query_voter_info(deps, did)?)
        }
        VotingQueryMsg::ListVoters { start_after, limit } => {
            to_json_binary(&query_list_voters(deps, start_after, limit)?)
        }
        VotingQueryMsg::Vote { proposal_id, voter } => {
            to_json_binary(&query_vote(deps, proposal_id, voter)?)
        }
    }
}

/// Query voting power for a DID
fn query_voting_power(deps: Deps, env: Env, did: String) -> StdResult<VotingPowerResponse> {
    let voter = VOTERS.may_load(deps.storage, &did)?;
    let power = voter.map(|v| v.voting_power).unwrap_or(Uint128::zero());
    
    Ok(VotingPowerResponse {
        power,
        height: env.block.height,
    })
}

/// Query total voting power
fn query_total_power(deps: Deps, env: Env, _height: Option<u64>) -> StdResult<TotalPowerResponse> {
    let power = TOTAL_POWER.load(deps.storage)?;
    
    Ok(TotalPowerResponse {
        power,
        height: env.block.height,
    })
}

/// Query voter information
fn query_voter_info(deps: Deps, did: String) -> StdResult<VoterInfoResponse> {
    let voter = VOTERS.load(deps.storage, &did)?;
    
    // Count proposals voted on
    let proposals_voted = VOTES
        .prefix(&did)
        .range(deps.storage, None, None, cosmwasm_std::Order::Ascending)
        .count() as u64;
    
    Ok(VoterInfoResponse {
        voter,
        proposals_voted,
    })
}

/// List all voters with pagination
fn query_list_voters(
    deps: Deps,
    start_after: Option<String>,
    limit: Option<u32>,
) -> StdResult<VotersListResponse> {
    let limit = limit.unwrap_or(30).min(100) as usize;
    let start = start_after.map(|s| cosmwasm_std::Bound::exclusive(s.as_str()));
    
    let voters: Vec<IdentityVoter> = VOTERS
        .range(deps.storage, start, None, cosmwasm_std::Order::Ascending)
        .take(limit)
        .map(|item| item.map(|(_, v)| v))
        .collect::<StdResult<Vec<_>>>()?;
    
    let total = VOTERS
        .range(deps.storage, None, None, cosmwasm_std::Order::Ascending)
        .count() as u64;
    
    Ok(VotersListResponse { voters, total })
}

/// Query vote on a proposal
fn query_vote(deps: Deps, proposal_id: u64, voter: String) -> StdResult<VoteResponse> {
    let vote = VOTES.may_load(deps.storage, (proposal_id, &voter))?;
    Ok(VoteResponse { vote })
}

// ====== IBC Entry Points ======

/// IBC channel handshake - Step 1: Channel open try
#[entry_point]
pub fn ibc_channel_open(
    _deps: DepsMut,
    _env: Env,
    msg: IbcChannelOpenMsg,
) -> Result<IbcChannelOpenResponse, ContractError> {
    // Validate the channel is being opened for the correct port
    if msg.channel().port_id != "wasm.identity_dao_voting" {
        return Err(ContractError::InvalidIbcChannel {});
    }

    // Accept channel if version is correct
    if msg.channel().version != "identity-dao-1" {
        return Ok(IbcChannelOpenResponse {
            version: "identity-dao-1".to_string(),
        });
    }

    Ok(IbcChannelOpenResponse { version: msg.channel().version.clone() })
}

/// IBC channel handshake - Step 2: Channel connected
#[entry_point]
pub fn ibc_channel_connect(
    deps: DepsMut,
    _env: Env,
    msg: IbcChannelConnectMsg,
) -> Result<IbcBasicResponse, ContractError> {
    // Store the channel ID for future use
    let channel_id = msg.channel().endpoint.channel_id.clone();
    
    // Store IBC channel info
    IBC_CHANNEL.save(deps.storage, &channel_id)?;

    Ok(IbcBasicResponse::new()
        .add_attribute("method", "ibc_channel_connect")
        .add_attribute("channel_id", channel_id))
}

/// IBC channel close handler
#[entry_point]
pub fn ibc_channel_close(
    deps: DepsMut,
    _env: Env,
    msg: IbcChannelCloseMsg,
) -> Result<IbcBasicResponse, ContractError> {
    let channel_id = msg.channel().endpoint.channel_id.clone();
    
    // Remove stored channel
    IBC_CHANNEL.remove(deps.storage);

    Ok(IbcBasicResponse::new()
        .add_attribute("method", "ibc_channel_close")
        .add_attribute("channel_id", channel_id))
}

/// IBC packet receive handler - Process DID verification responses
#[entry_point]
pub fn ibc_packet_receive(
    deps: DepsMut,
    env: Env,
    msg: IbcPacketReceiveMsg,
) -> Result<IbcReceiveResponse, Never> {
    // Parse the packet data
    let packet_data: IbcDIDQueryResponse = from_json(&msg.packet.data)
        .map_err(|err| ContractError::InvalidIbcPacket { error: err.to_string() })
        .unwrap_or_else(|_| IbcDIDQueryResponse {
            did: String::new(),
            is_verified: false,
            verification_level: 0,
            error: Some("Failed to parse packet".to_string()),
        });

    // Process DID verification response
    if let Some(error) = packet_data.error {
        // Log error but don't fail the IBC transaction
        return Ok(IbcReceiveResponse::new()
            .add_attribute("method", "ibc_packet_receive")
            .add_attribute("error", error)
            .set_ack(to_json_binary(&IbcAcknowledgement { success: false }).unwrap()));
    }

    // Store the verification result for pending voter update
    if !packet_data.did.is_empty() {
        PENDING_VERIFICATIONS.save(
            deps.storage,
            &packet_data.did,
            &PendingVerification {
                is_verified: packet_data.is_verified,
                verification_level: packet_data.verification_level,
                timestamp: env.block.time.seconds(),
            },
        ).ok();
    }

    Ok(IbcReceiveResponse::new()
        .add_attribute("method", "ibc_packet_receive")
        .add_attribute("did", packet_data.did)
        .add_attribute("verified", packet_data.is_verified.to_string())
        .set_ack(to_json_binary(&IbcAcknowledgement { success: true }).unwrap()))
}

/// IBC packet acknowledgement handler
#[entry_point]
pub fn ibc_packet_ack(
    _deps: DepsMut,
    _env: Env,
    msg: IbcPacketAckMsg,
) -> Result<IbcBasicResponse, ContractError> {
    // Parse acknowledgement
    let ack: IbcAcknowledgement = from_json(&msg.acknowledgement.data)
        .map_err(|err| ContractError::InvalidIbcPacket { error: err.to_string() })?;

    Ok(IbcBasicResponse::new()
        .add_attribute("method", "ibc_packet_ack")
        .add_attribute("success", ack.success.to_string()))
}

/// IBC packet timeout handler
#[entry_point]
pub fn ibc_packet_timeout(
    _deps: DepsMut,
    _env: Env,
    _msg: IbcPacketTimeoutMsg,
) -> Result<IbcBasicResponse, ContractError> {
    // Handle timeout - could retry or mark verification as failed
    Ok(IbcBasicResponse::new()
        .add_attribute("method", "ibc_packet_timeout"))
}

// ====== IBC Helper Types ======

/// IBC DID query response packet
#[derive(serde::Serialize, serde::Deserialize)]
struct IbcDIDQueryResponse {
    pub did: String,
    pub is_verified: bool,
    pub verification_level: u8,
    pub error: Option<String>,
}

/// IBC acknowledgement
#[derive(serde::Serialize, serde::Deserialize)]
struct IbcAcknowledgement {
    pub success: bool,
}