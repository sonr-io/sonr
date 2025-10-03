use cosmwasm_std::StdError;
use thiserror::Error;

/// Common errors for Identity DAO contracts
#[derive(Error, Debug, PartialEq)]
pub enum ContractError {
    #[error("{0}")]
    Std(#[from] StdError),

    #[error("Unauthorized")]
    Unauthorized {},

    #[error("Invalid DID: {did}")]
    InvalidDID { did: String },

    #[error("DID not verified")]
    DIDNotVerified {},

    #[error("Insufficient voting power")]
    InsufficientVotingPower {},

    #[error("Proposal not found")]
    ProposalNotFound {},

    #[error("Voting period ended")]
    VotingPeriodEnded {},

    #[error("Voting period not ended")]
    VotingPeriodNotEnded {},

    #[error("Already voted")]
    AlreadyVoted {},

    #[error("Invalid threshold")]
    InvalidThreshold {},

    #[error("No attestation found for DID: {did}")]
    NoAttestation { did: String },

    #[error("Custom error: {msg}")]
    CustomError { msg: String },

    #[error("Invalid IBC channel")]
    InvalidIbcChannel {},

    #[error("Invalid IBC packet: {error}")]
    InvalidIbcPacket { error: String },
}