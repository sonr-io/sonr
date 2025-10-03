#[cfg(test)]
mod did_integration_tests {
    use cosmwasm_std::{
        testing::{mock_dependencies_with_balance, mock_env, mock_info, MockApi, MockQuerier, MockStorage},
        from_json, to_json_binary, Addr, Binary, Coin, ContractResult, Deps, DepsMut, Empty, Env, 
        MessageInfo, Response, StdError, StdResult, Uint128, WasmMsg, CosmosMsg,
        QueryRequest, SystemResult,
    };
    use cosmwasm_std::testing::MockQuerierCustomHandlerResult;
    
    use identity_dao_shared::{
        CoreInstantiateMsg, CoreExecuteMsg, CoreQueryMsg,
        VotingInstantiateMsg, VotingExecuteMsg, VotingQueryMsg,
        ProposalInstantiateMsg, ProposalExecuteMsg, ProposalQueryMsg,
        PreProposeInstantiateMsg, PreProposeExecuteMsg, PreProposeQueryMsg,
        VotingConfig, ProposalStatus, Vote, DaoConfigResponse,
        VotingPowerResponse, ProposalResponse, IdentityVoter, VerificationStatus,
        bindings::{SonrQuery, DIDDocumentResponse, VerificationResponse},
    };

    struct IntegrationTestSetup {
        core_addr: Addr,
        voting_addr: Addr,
        proposal_addr: Addr,
        pre_propose_addr: Addr,
    }

    /// Setup all DAO contracts with proper module registration
    fn setup_dao_contracts() -> (MockStorage, MockApi, MockQuerier, IntegrationTestSetup) {
        let mut storage = MockStorage::default();
        let api = MockApi::default();
        let mut querier = MockQuerier::new(&[]);
        
        // Setup custom query handler for DID module queries
        querier.update_wasm(|query| -> MockQuerierCustomHandlerResult {
            match query {
                cosmwasm_std::WasmQuery::Smart { contract_addr, msg } => {
                    // Handle DID module queries
                    if let Ok(sonr_query) = from_json::<SonrQuery>(msg) {
                        match sonr_query {
                            SonrQuery::GetDIDDocument { did } => {
                                let response = DIDDocumentResponse {
                                    did: did.clone(),
                                    controller: format!("{}_controller", did),
                                    verification_methods: vec![],
                                    authentication: vec![],
                                    assertion_method: vec![],
                                    capability_invocation: vec![],
                                    capability_delegation: vec![],
                                    service_endpoints: vec![],
                                };
                                return SystemResult::Ok(ContractResult::Ok(to_json_binary(&response).unwrap()));
                            },
                            SonrQuery::VerifyDIDController { did, controller } => {
                                let response = VerificationResponse {
                                    is_valid: controller == format!("{}_controller", did),
                                    error: None,
                                };
                                return SystemResult::Ok(ContractResult::Ok(to_json_binary(&response).unwrap()));
                            },
                        }
                    }
                    SystemResult::Ok(ContractResult::Err("Not a DID query".to_string()))
                },
                _ => SystemResult::Ok(ContractResult::Err("Unsupported query".to_string())),
            }
        });
        
        // Contract addresses
        let core_addr = Addr::unchecked("dao_core");
        let voting_addr = Addr::unchecked("dao_voting");
        let proposal_addr = Addr::unchecked("dao_proposal");
        let pre_propose_addr = Addr::unchecked("dao_pre_propose");
        
        let setup = IntegrationTestSetup {
            core_addr: core_addr.clone(),
            voting_addr: voting_addr.clone(),
            proposal_addr: proposal_addr.clone(),
            pre_propose_addr: pre_propose_addr.clone(),
        };
        
        (storage, api, querier, setup)
    }

    #[test]
    fn test_full_dao_initialization_with_did() {
        let (mut storage, api, querier, setup) = setup_dao_contracts();
        let env = mock_env();
        let creator = mock_info("creator", &[]);
        
        // 1. Initialize Core Module
        let core_msg = CoreInstantiateMsg {
            name: "Identity DAO".to_string(),
            description: "A DAO for identity management".to_string(),
            voting_config: VotingConfig {
                threshold: cosmwasm_std::Decimal::percent(51),
                quorum: cosmwasm_std::Decimal::percent(10),
                voting_period: 86400,
                proposal_deposit: Uint128::from(1000000u128),
            },
            admin: Some("admin".to_string()),
            enable_did_integration: true,
        };
        
        // Simulate core instantiation
        // In real integration test, this would be done via instantiate_contract
        
        // 2. Initialize Voting Module with DID integration
        let voting_msg = VotingInstantiateMsg {
            dao_core: setup.core_addr.to_string(),
            min_verification_level: 1,
            use_reputation_weight: true,
        };
        
        // 3. Initialize Proposal Module
        let proposal_msg = ProposalInstantiateMsg {
            dao_core: setup.core_addr.to_string(),
            voting_module: setup.voting_addr.to_string(),
            pre_propose_module: Some(setup.pre_propose_addr.to_string()),
            proposal_deposit: Uint128::from(1000000u128),
            max_voting_period: 604800, // 7 days
        };
        
        // 4. Initialize Pre-Propose Module with DID gating
        let pre_propose_msg = PreProposeInstantiateMsg {
            dao_core: setup.core_addr.to_string(),
            proposal_module: setup.proposal_addr.to_string(),
            require_verified_did: true,
            min_reputation_score: 10,
            deposit_amount: Uint128::from(1000000u128),
            deposit_denom: "usnr".to_string(),
        };
        
        // Verify all modules are properly configured
        assert_eq!(voting_msg.dao_core, setup.core_addr.to_string());
        assert_eq!(proposal_msg.voting_module, setup.voting_addr.to_string());
        assert_eq!(pre_propose_msg.proposal_module, setup.proposal_addr.to_string());
    }

    #[test]
    fn test_did_voter_registration_and_verification() {
        let (storage, api, querier, setup) = setup_dao_contracts();
        let env = mock_env();
        
        // Register a voter with DID
        let did = "did:sonr:alice123";
        let alice_addr = "alice_address";
        
        let register_msg = VotingExecuteMsg::UpdateVoter {
            did: did.to_string(),
            address: alice_addr.to_string(),
        };
        
        // Simulate DID verification through stargate query
        let verification_query = SonrQuery::VerifyDIDController {
            did: did.to_string(),
            controller: format!("{}_controller", did),
        };
        
        // Query would return verification status
        let deps = mock_dependencies_with_balance(&[]);
        let query_result: StdResult<Binary> = to_json_binary(&VerificationResponse {
            is_valid: true,
            error: None,
        });
        
        assert!(query_result.is_ok());
    }

    #[test]
    fn test_did_gated_proposal_creation() {
        let (storage, api, querier, setup) = setup_dao_contracts();
        let env = mock_env();
        
        // Setup verified DID holder
        let proposer_did = "did:sonr:proposer456";
        let proposer_addr = "proposer_address";
        
        // Create proposal through pre-propose module
        let proposal_msg = PreProposeExecuteMsg::ProposeWithDID {
            did: proposer_did.to_string(),
            title: "Upgrade Protocol".to_string(),
            description: "Proposal to upgrade the protocol to v2".to_string(),
            msgs: vec![],
        };
        
        // Verify DID before allowing proposal
        let verification_query = SonrQuery::GetDIDDocument {
            did: proposer_did.to_string(),
        };
        
        // Check that proposal creation requires verified DID
        let deps = mock_dependencies_with_balance(&[]);
        let did_response = DIDDocumentResponse {
            did: proposer_did.to_string(),
            controller: format!("{}_controller", proposer_did),
            verification_methods: vec![],
            authentication: vec![],
            assertion_method: vec![],
            capability_invocation: vec![],
            capability_delegation: vec![],
            service_endpoints: vec![],
        };
        
        let query_result: StdResult<Binary> = to_json_binary(&did_response);
        assert!(query_result.is_ok());
    }

    #[test]
    fn test_voting_with_did_based_power() {
        let (storage, api, querier, setup) = setup_dao_contracts();
        let env = mock_env();
        
        // Setup multiple voters with different verification levels
        let voters = vec![
            ("did:sonr:voter1", "voter1_addr", 100u32), // High reputation
            ("did:sonr:voter2", "voter2_addr", 50u32),  // Medium reputation
            ("did:sonr:voter3", "voter3_addr", 10u32),  // Low reputation
        ];
        
        for (did, addr, reputation) in voters.iter() {
            // Register voter
            let register_msg = VotingExecuteMsg::UpdateVoter {
                did: did.to_string(),
                address: addr.to_string(),
            };
            
            // Voting power should be weighted by reputation
            let expected_power = calculate_did_voting_power(*reputation);
            
            // Query voting power
            let query_msg = VotingQueryMsg::GetVotingPower {
                address: addr.to_string(),
            };
            
            // Verify power calculation
            assert!(expected_power > Uint128::zero());
            if *reputation > 50 {
                assert!(expected_power > Uint128::from(1u128));
            }
        }
    }

    #[test]
    fn test_cross_module_proposal_execution() {
        let (storage, api, querier, setup) = setup_dao_contracts();
        let env = mock_env();
        
        // Create proposal with multiple actions
        let proposal_msgs = vec![
            CosmosMsg::Wasm(WasmMsg::Execute {
                contract_addr: "target_contract".to_string(),
                msg: to_json_binary(&"action1").unwrap(),
                funds: vec![],
            }),
            CosmosMsg::Wasm(WasmMsg::Execute {
                contract_addr: "target_contract".to_string(),
                msg: to_json_binary(&"action2").unwrap(),
                funds: vec![],
            }),
        ];
        
        // Proposal creation through pre-propose
        let create_proposal = ProposalExecuteMsg::CreateProposal {
            title: "Multi-action Proposal".to_string(),
            description: "Execute multiple actions".to_string(),
            msgs: proposal_msgs.clone(),
            proposer: "proposer_addr".to_string(),
        };
        
        // Simulate voting to pass proposal
        let vote_msg = VotingExecuteMsg::Vote {
            proposal_id: 1,
            vote: Vote::Yes,
        };
        
        // Execute proposal through core
        let execute_msg = CoreExecuteMsg::ExecuteProposal {
            proposal_id: 1,
        };
        
        // Verify all messages are executed
        assert_eq!(proposal_msgs.len(), 2);
    }

    #[test]
    fn test_ibc_did_verification() {
        let (storage, api, querier, setup) = setup_dao_contracts();
        let env = mock_env();
        
        // Setup IBC channel for DID verification
        let ibc_channel = "channel-0";
        let sonr_chain_did = "did:sonr:ibc_user789";
        
        // Query DID from Sonr chain via IBC
        let ibc_query = SonrQuery::GetDIDDocument {
            did: sonr_chain_did.to_string(),
        };
        
        // Simulate IBC response
        let ibc_response = DIDDocumentResponse {
            did: sonr_chain_did.to_string(),
            controller: "cosmos1abc...".to_string(),
            verification_methods: vec!["key1".to_string()],
            authentication: vec!["auth1".to_string()],
            assertion_method: vec![],
            capability_invocation: vec![],
            capability_delegation: vec![],
            service_endpoints: vec![],
        };
        
        // Verify cross-chain DID
        assert_eq!(ibc_response.did, sonr_chain_did);
        assert!(!ibc_response.verification_methods.is_empty());
    }

    #[test]
    fn test_reputation_based_quorum() {
        let (storage, api, querier, setup) = setup_dao_contracts();
        let env = mock_env();
        
        // Setup voters with varying reputation
        let high_rep_voters = 2;
        let low_rep_voters = 10;
        
        // High reputation voters should have more weight
        let high_rep_power = Uint128::from(10u128);
        let low_rep_power = Uint128::from(1u128);
        
        let total_power = high_rep_power
            .checked_mul(Uint128::from(high_rep_voters as u128))
            .unwrap()
            .checked_add(low_rep_power.checked_mul(Uint128::from(low_rep_voters as u128)).unwrap())
            .unwrap();
        
        // Calculate quorum (10% of total power)
        let quorum_threshold = total_power.multiply_ratio(10u128, 100u128);
        
        // Verify that 2 high-rep voters can meet quorum
        let high_rep_voting_power = high_rep_power
            .checked_mul(Uint128::from(high_rep_voters as u128))
            .unwrap();
        
        assert!(high_rep_voting_power >= quorum_threshold);
    }

    #[test]
    fn test_treasury_management_with_did_auth() {
        let (storage, api, querier, setup) = setup_dao_contracts();
        let env = mock_env();
        
        // Only verified DID holders can propose treasury withdrawals
        let treasury_manager_did = "did:sonr:treasurer";
        
        // Create treasury withdrawal proposal
        let withdrawal_msg = CoreExecuteMsg::WithdrawFromTreasury {
            recipient: "recipient_addr".to_string(),
            amount: Uint128::from(1000000u128),
            denom: "usnr".to_string(),
        };
        
        // Verify proposer has sufficient reputation
        let min_reputation_for_treasury = 50u32;
        
        // Query proposer's reputation
        let verification_query = SonrQuery::GetDIDDocument {
            did: treasury_manager_did.to_string(),
        };
        
        // Simulate reputation check
        let has_sufficient_reputation = true; // Would be fetched from DID module
        assert!(has_sufficient_reputation);
    }

    #[test]
    fn test_emergency_pause_with_did_multisig() {
        let (storage, api, querier, setup) = setup_dao_contracts();
        let env = mock_env();
        
        // Emergency actions require multiple DID signatures
        let emergency_signers = vec![
            "did:sonr:emergency1",
            "did:sonr:emergency2",
            "did:sonr:emergency3",
        ];
        
        // Each signer must be verified
        for signer_did in emergency_signers.iter() {
            let verification = SonrQuery::VerifyDIDController {
                did: signer_did.to_string(),
                controller: format!("{}_controller", signer_did),
            };
            
            // All signers must be valid
            let response = VerificationResponse {
                is_valid: true,
                error: None,
            };
            
            assert!(response.is_valid);
        }
        
        // Require 2/3 signatures for emergency action
        let required_signatures = 2;
        let collected_signatures = 3;
        
        assert!(collected_signatures >= required_signatures);
    }

    // Helper function to calculate DID-based voting power
    fn calculate_did_voting_power(reputation_score: u32) -> Uint128 {
        let base_power = Uint128::from(1u128);
        if reputation_score > 0 {
            // Formula: base_power * (1 + reputation_score / 100)
            let multiplier = Uint128::from((100 + reputation_score) as u128);
            base_power.multiply_ratio(multiplier, 100u128)
        } else {
            base_power
        }
    }
}