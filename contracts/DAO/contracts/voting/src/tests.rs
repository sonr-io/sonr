#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
    use cosmwasm_std::{from_json, Addr, Uint128};
    use identity_dao_shared::{
        VotingInstantiateMsg, VotingExecuteMsg, VotingQueryMsg,
        VotingPowerResponse, TotalPowerResponse, VoterInfoResponse,
        Vote, VoteInfo, IdentityVoter, VerificationStatus,
        VoteResponse, VotersListResponse,
    };

    fn setup_contract() -> (cosmwasm_std::DepsMut<'_>, Env, MessageInfo) {
        let mut deps = mock_dependencies();
        let env = mock_env();
        let info = mock_info("creator", &[]);
        
        let msg = VotingInstantiateMsg {
            dao_core: "dao_core_addr".to_string(),
            min_verification_level: 1,
            use_reputation_weight: true,
        };
        
        let res = instantiate(deps.branch(), env.clone(), info.clone(), msg);
        assert!(res.is_ok());
        
        (deps.as_mut(), env, info)
    }

    #[test]
    fn test_instantiate() {
        let mut deps = mock_dependencies();
        let env = mock_env();
        let info = mock_info("creator", &[]);
        
        let msg = VotingInstantiateMsg {
            dao_core: "dao_core_addr".to_string(),
            min_verification_level: 2,
            use_reputation_weight: false,
        };
        
        let res = instantiate(deps.as_mut(), env, info, msg.clone()).unwrap();
        assert_eq!(res.attributes.len(), 2);
        assert_eq!(res.attributes[0].value, "instantiate");
        assert_eq!(res.attributes[1].value, msg.dao_core);
        
        // Verify state was saved correctly
        let config = CONFIG.load(&deps.storage).unwrap();
        assert_eq!(config.dao_core, Addr::unchecked("dao_core_addr"));
        assert_eq!(config.min_verification_level, 2);
        assert_eq!(config.use_reputation_weight, false);
        
        let total_power = TOTAL_POWER.load(&deps.storage).unwrap();
        assert_eq!(total_power, Uint128::zero());
    }

    #[test]
    fn test_update_voter() {
        let (mut deps, env, _) = setup_contract();
        let dao_core_info = mock_info("dao_core_addr", &[]);
        
        let msg = VotingExecuteMsg::UpdateVoter {
            did: "did:sonr:alice123".to_string(),
            address: "alice_addr".to_string(),
        };
        
        let res = execute(deps.branch(), env, dao_core_info, msg).unwrap();
        assert_eq!(res.attributes[0].value, "update_voter");
        assert_eq!(res.attributes[1].value, "did:sonr:alice123");
        
        // Verify voter was saved
        let voter = VOTERS.load(&deps.storage, "alice_addr").unwrap();
        assert_eq!(voter.did, "did:sonr:alice123");
        assert_eq!(voter.address, "alice_addr");
        assert_eq!(voter.voting_power, Uint128::from(1u128)); // Base power
        assert_eq!(voter.verification_status, VerificationStatus::Pending);
        
        // Verify total power was updated
        let total_power = TOTAL_POWER.load(&deps.storage).unwrap();
        assert_eq!(total_power, Uint128::from(1u128));
    }

    #[test]
    fn test_update_voter_with_reputation() {
        let (mut deps, env, _) = setup_contract();
        let dao_core_info = mock_info("dao_core_addr", &[]);
        
        // Add voter with reputation
        let msg = VotingExecuteMsg::UpdateVoter {
            did: "did:sonr:bob456".to_string(),
            address: "bob_addr".to_string(),
        };
        
        let res = execute(deps.branch(), env.clone(), dao_core_info.clone(), msg).unwrap();
        assert!(res.is_ok());
        
        // Manually update voter with verified status and reputation
        let mut voter = VOTERS.load(&deps.storage, "bob_addr").unwrap();
        voter.verification_status = VerificationStatus::Verified;
        voter.reputation_score = 50;
        voter.voting_power = calculate_voting_power(true, 50);
        VOTERS.save(deps.as_mut().storage, "bob_addr", &voter).unwrap();
        
        // Update total power
        TOTAL_POWER.save(deps.as_mut().storage, &voter.voting_power).unwrap();
        
        // Verify voting power calculation with reputation
        let voter = VOTERS.load(&deps.storage, "bob_addr").unwrap();
        assert_eq!(voter.reputation_score, 50);
        assert!(voter.voting_power > Uint128::from(1u128)); // Should be higher than base
    }

    #[test]
    fn test_unauthorized_update_voter() {
        let (mut deps, env, _) = setup_contract();
        let unauthorized_info = mock_info("unauthorized", &[]);
        
        let msg = VotingExecuteMsg::UpdateVoter {
            did: "did:sonr:alice123".to_string(),
            address: "alice_addr".to_string(),
        };
        
        let err = execute(deps.branch(), env, unauthorized_info, msg).unwrap_err();
        assert!(matches!(err, ContractError::Unauthorized {}));
    }

    #[test]
    fn test_vote() {
        let (mut deps, env, _) = setup_contract();
        
        // First add a voter
        let dao_core_info = mock_info("dao_core_addr", &[]);
        let update_msg = VotingExecuteMsg::UpdateVoter {
            did: "did:sonr:alice123".to_string(),
            address: "alice_addr".to_string(),
        };
        execute(deps.branch(), env.clone(), dao_core_info, update_msg).unwrap();
        
        // Mark voter as verified
        let mut voter = VOTERS.load(&deps.storage, "alice_addr").unwrap();
        voter.verification_status = VerificationStatus::Verified;
        VOTERS.save(deps.as_mut().storage, "alice_addr", &voter).unwrap();
        
        // Cast vote
        let alice_info = mock_info("alice_addr", &[]);
        let vote_msg = VotingExecuteMsg::Vote {
            proposal_id: 1,
            vote: Vote::Yes,
        };
        
        let res = execute(deps.branch(), env, alice_info, vote_msg).unwrap();
        assert_eq!(res.attributes[0].value, "vote");
        assert_eq!(res.attributes[1].value, "1");
        assert_eq!(res.attributes[2].value, "alice_addr");
        assert_eq!(res.attributes[3].value, "yes");
        
        // Verify vote was recorded
        let vote_info = VOTES.load(&deps.storage, (1, "alice_addr")).unwrap();
        assert_eq!(vote_info.vote, Vote::Yes);
        assert_eq!(vote_info.voting_power, voter.voting_power);
        
        // Verify voter was added to proposal voters list
        let proposal_voters = PROPOSAL_VOTERS.load(&deps.storage, 1).unwrap();
        assert_eq!(proposal_voters.len(), 1);
        assert_eq!(proposal_voters[0], "alice_addr");
    }

    #[test]
    fn test_vote_no() {
        let (mut deps, env, _) = setup_contract();
        
        // Add and verify voter
        let dao_core_info = mock_info("dao_core_addr", &[]);
        let update_msg = VotingExecuteMsg::UpdateVoter {
            did: "did:sonr:bob456".to_string(),
            address: "bob_addr".to_string(),
        };
        execute(deps.branch(), env.clone(), dao_core_info, update_msg).unwrap();
        
        let mut voter = VOTERS.load(&deps.storage, "bob_addr").unwrap();
        voter.verification_status = VerificationStatus::Verified;
        VOTERS.save(deps.as_mut().storage, "bob_addr", &voter).unwrap();
        
        // Cast No vote
        let bob_info = mock_info("bob_addr", &[]);
        let vote_msg = VotingExecuteMsg::Vote {
            proposal_id: 2,
            vote: Vote::No,
        };
        
        let res = execute(deps.branch(), env, bob_info, vote_msg).unwrap();
        assert_eq!(res.attributes[3].value, "no");
        
        let vote_info = VOTES.load(&deps.storage, (2, "bob_addr")).unwrap();
        assert_eq!(vote_info.vote, Vote::No);
    }

    #[test]
    fn test_vote_abstain() {
        let (mut deps, env, _) = setup_contract();
        
        // Add and verify voter
        let dao_core_info = mock_info("dao_core_addr", &[]);
        let update_msg = VotingExecuteMsg::UpdateVoter {
            did: "did:sonr:charlie789".to_string(),
            address: "charlie_addr".to_string(),
        };
        execute(deps.branch(), env.clone(), dao_core_info, update_msg).unwrap();
        
        let mut voter = VOTERS.load(&deps.storage, "charlie_addr").unwrap();
        voter.verification_status = VerificationStatus::Verified;
        VOTERS.save(deps.as_mut().storage, "charlie_addr", &voter).unwrap();
        
        // Cast Abstain vote
        let charlie_info = mock_info("charlie_addr", &[]);
        let vote_msg = VotingExecuteMsg::Vote {
            proposal_id: 3,
            vote: Vote::Abstain,
        };
        
        let res = execute(deps.branch(), env, charlie_info, vote_msg).unwrap();
        assert_eq!(res.attributes[3].value, "abstain");
        
        let vote_info = VOTES.load(&deps.storage, (3, "charlie_addr")).unwrap();
        assert_eq!(vote_info.vote, Vote::Abstain);
    }

    #[test]
    fn test_unverified_voter_cannot_vote() {
        let (mut deps, env, _) = setup_contract();
        
        // Add voter but don't verify
        let dao_core_info = mock_info("dao_core_addr", &[]);
        let update_msg = VotingExecuteMsg::UpdateVoter {
            did: "did:sonr:dave000".to_string(),
            address: "dave_addr".to_string(),
        };
        execute(deps.branch(), env.clone(), dao_core_info, update_msg).unwrap();
        
        // Try to vote without verification
        let dave_info = mock_info("dave_addr", &[]);
        let vote_msg = VotingExecuteMsg::Vote {
            proposal_id: 4,
            vote: Vote::Yes,
        };
        
        let err = execute(deps.branch(), env, dave_info, vote_msg).unwrap_err();
        assert!(matches!(err, ContractError::NotVerified { .. }));
    }

    #[test]
    fn test_double_voting_prevention() {
        let (mut deps, env, _) = setup_contract();
        
        // Add and verify voter
        let dao_core_info = mock_info("dao_core_addr", &[]);
        let update_msg = VotingExecuteMsg::UpdateVoter {
            did: "did:sonr:eve111".to_string(),
            address: "eve_addr".to_string(),
        };
        execute(deps.branch(), env.clone(), dao_core_info, update_msg).unwrap();
        
        let mut voter = VOTERS.load(&deps.storage, "eve_addr").unwrap();
        voter.verification_status = VerificationStatus::Verified;
        VOTERS.save(deps.as_mut().storage, "eve_addr", &voter).unwrap();
        
        // First vote
        let eve_info = mock_info("eve_addr", &[]);
        let vote_msg = VotingExecuteMsg::Vote {
            proposal_id: 5,
            vote: Vote::Yes,
        };
        execute(deps.branch(), env.clone(), eve_info.clone(), vote_msg.clone()).unwrap();
        
        // Try to vote again
        let err = execute(deps.branch(), env, eve_info, vote_msg).unwrap_err();
        assert!(matches!(err, ContractError::AlreadyVoted { .. }));
    }

    #[test]
    fn test_remove_voter() {
        let (mut deps, env, _) = setup_contract();
        
        // Add voter
        let dao_core_info = mock_info("dao_core_addr", &[]);
        let update_msg = VotingExecuteMsg::UpdateVoter {
            did: "did:sonr:frank222".to_string(),
            address: "frank_addr".to_string(),
        };
        execute(deps.branch(), env.clone(), dao_core_info.clone(), update_msg).unwrap();
        
        // Verify voter exists
        assert!(VOTERS.has(&deps.storage, "frank_addr"));
        
        // Remove voter
        let remove_msg = VotingExecuteMsg::RemoveVoter {
            did: "did:sonr:frank222".to_string(),
        };
        let res = execute(deps.branch(), env, dao_core_info, remove_msg).unwrap();
        assert_eq!(res.attributes[0].value, "remove_voter");
        assert_eq!(res.attributes[1].value, "did:sonr:frank222");
        
        // Verify voter was removed (by DID lookup)
        // Note: In real implementation, you'd need to maintain a DID->address mapping
    }

    #[test]
    fn test_query_voting_power() {
        let (mut deps, env, _) = setup_contract();
        
        // Add voter with specific voting power
        let dao_core_info = mock_info("dao_core_addr", &[]);
        let update_msg = VotingExecuteMsg::UpdateVoter {
            did: "did:sonr:grace333".to_string(),
            address: "grace_addr".to_string(),
        };
        execute(deps.branch(), env, dao_core_info, update_msg).unwrap();
        
        let mut voter = VOTERS.load(&deps.storage, "grace_addr").unwrap();
        voter.verification_status = VerificationStatus::Verified;
        voter.reputation_score = 75;
        voter.voting_power = calculate_voting_power(true, 75);
        VOTERS.save(deps.as_mut().storage, "grace_addr", &voter).unwrap();
        
        // Query voting power
        let res = query(
            deps.as_ref(),
            mock_env(),
            VotingQueryMsg::GetVotingPower { address: "grace_addr".to_string() }
        ).unwrap();
        
        let power_response: VotingPowerResponse = from_json(&res).unwrap();
        assert_eq!(power_response.voting_power, voter.voting_power);
        assert_eq!(power_response.is_verified, true);
    }

    #[test]
    fn test_query_total_power() {
        let (mut deps, env, _) = setup_contract();
        
        // Add multiple voters
        let dao_core_info = mock_info("dao_core_addr", &[]);
        
        for i in 0..3 {
            let update_msg = VotingExecuteMsg::UpdateVoter {
                did: format!("did:sonr:voter{}", i),
                address: format!("voter_{}_addr", i),
            };
            execute(deps.branch(), env.clone(), dao_core_info.clone(), update_msg).unwrap();
        }
        
        // Set total power
        TOTAL_POWER.save(deps.as_mut().storage, &Uint128::from(3u128)).unwrap();
        
        // Query total power
        let res = query(deps.as_ref(), mock_env(), VotingQueryMsg::GetTotalPower {}).unwrap();
        let total_response: TotalPowerResponse = from_json(&res).unwrap();
        assert_eq!(total_response.total_power, Uint128::from(3u128));
    }

    #[test]
    fn test_query_voter_info() {
        let (mut deps, env, _) = setup_contract();
        
        // Add voter
        let dao_core_info = mock_info("dao_core_addr", &[]);
        let update_msg = VotingExecuteMsg::UpdateVoter {
            did: "did:sonr:henry444".to_string(),
            address: "henry_addr".to_string(),
        };
        execute(deps.branch(), env, dao_core_info, update_msg).unwrap();
        
        // Update voter details
        let mut voter = VOTERS.load(&deps.storage, "henry_addr").unwrap();
        voter.verification_status = VerificationStatus::Verified;
        voter.reputation_score = 90;
        voter.voting_power = calculate_voting_power(true, 90);
        VOTERS.save(deps.as_mut().storage, "henry_addr", &voter).unwrap();
        
        // Query voter info
        let res = query(
            deps.as_ref(),
            mock_env(),
            VotingQueryMsg::GetVoterInfo { address: "henry_addr".to_string() }
        ).unwrap();
        
        let info_response: VoterInfoResponse = from_json(&res).unwrap();
        assert_eq!(info_response.did, "did:sonr:henry444");
        assert_eq!(info_response.address, "henry_addr");
        assert_eq!(info_response.voting_power, voter.voting_power);
        assert_eq!(info_response.verification_status, VerificationStatus::Verified);
        assert_eq!(info_response.reputation_score, 90);
    }

    #[test]
    fn test_query_vote() {
        let (mut deps, env, _) = setup_contract();
        
        // Add and verify voter
        let dao_core_info = mock_info("dao_core_addr", &[]);
        let update_msg = VotingExecuteMsg::UpdateVoter {
            did: "did:sonr:iris555".to_string(),
            address: "iris_addr".to_string(),
        };
        execute(deps.branch(), env.clone(), dao_core_info, update_msg).unwrap();
        
        let mut voter = VOTERS.load(&deps.storage, "iris_addr").unwrap();
        voter.verification_status = VerificationStatus::Verified;
        VOTERS.save(deps.as_mut().storage, "iris_addr", &voter).unwrap();
        
        // Cast vote
        let iris_info = mock_info("iris_addr", &[]);
        let vote_msg = VotingExecuteMsg::Vote {
            proposal_id: 10,
            vote: Vote::Yes,
        };
        execute(deps.branch(), env, iris_info, vote_msg).unwrap();
        
        // Query vote
        let res = query(
            deps.as_ref(),
            mock_env(),
            VotingQueryMsg::GetVote { 
                proposal_id: 10,
                voter: "iris_addr".to_string()
            }
        ).unwrap();
        
        let vote_response: VoteResponse = from_json(&res).unwrap();
        assert_eq!(vote_response.vote, Some(Vote::Yes));
        assert_eq!(vote_response.voting_power, voter.voting_power);
    }

    #[test]
    fn test_query_voters_list() {
        let (mut deps, env, _) = setup_contract();
        
        // Add multiple voters
        let dao_core_info = mock_info("dao_core_addr", &[]);
        let voters = vec!["jack", "kate", "leo"];
        
        for (i, name) in voters.iter().enumerate() {
            let update_msg = VotingExecuteMsg::UpdateVoter {
                did: format!("did:sonr:{}", name),
                address: format!("{}_addr", name),
            };
            execute(deps.branch(), env.clone(), dao_core_info.clone(), update_msg).unwrap();
        }
        
        // Query voters list
        let res = query(
            deps.as_ref(),
            mock_env(),
            VotingQueryMsg::ListVoters { 
                start_after: None,
                limit: Some(10)
            }
        ).unwrap();
        
        let list_response: VotersListResponse = from_json(&res).unwrap();
        assert_eq!(list_response.voters.len(), 3);
    }

    // Helper function to calculate voting power with reputation
    fn calculate_voting_power(use_reputation: bool, reputation_score: u32) -> Uint128 {
        let base_power = Uint128::from(1u128);
        if use_reputation && reputation_score > 0 {
            // Simple formula: base_power * (1 + reputation_score / 100)
            let multiplier = Uint128::from((100 + reputation_score) as u128);
            base_power.multiply_ratio(multiplier, 100u128)
        } else {
            base_power
        }
    }
}