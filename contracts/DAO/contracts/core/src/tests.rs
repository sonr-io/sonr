#[cfg(test)]
mod tests {
    use super::*;
    use cosmwasm_std::testing::{mock_dependencies, mock_env, mock_info};
    use cosmwasm_std::{coins, from_json, Addr, Coin, Uint128};
    use identity_dao_shared::{
        CoreInstantiateMsg, CoreExecuteMsg, CoreQueryMsg, DaoConfigResponse,
        ModuleConfig, ProposalExecuteMsg, TreasuryInfo, VotingConfig,
        DaoStatsResponse,
    };

    fn setup_contract() -> (cosmwasm_std::DepsMut<'_>, Env, MessageInfo) {
        let mut deps = mock_dependencies();
        let env = mock_env();
        let info = mock_info("creator", &[]);
        
        let msg = CoreInstantiateMsg {
            name: "Test Identity DAO".to_string(),
            description: "A test DAO for identity management".to_string(),
            voting_config: VotingConfig {
                threshold: cosmwasm_std::Decimal::percent(51),
                quorum: cosmwasm_std::Decimal::percent(10),
                voting_period: 86400, // 24 hours
                proposal_deposit: Uint128::from(1000000u128),
            },
            admin: Some("admin".to_string()),
            enable_did_integration: true,
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
        
        let msg = CoreInstantiateMsg {
            name: "Identity DAO".to_string(),
            description: "Decentralized Identity DAO".to_string(),
            voting_config: VotingConfig {
                threshold: cosmwasm_std::Decimal::percent(60),
                quorum: cosmwasm_std::Decimal::percent(20),
                voting_period: 604800, // 7 days
                proposal_deposit: Uint128::from(5000000u128),
            },
            admin: Some("admin".to_string()),
            enable_did_integration: true,
        };
        
        let res = instantiate(deps.as_mut(), env, info, msg.clone()).unwrap();
        assert_eq!(res.attributes.len(), 3);
        assert_eq!(res.attributes[0].value, "instantiate");
        assert_eq!(res.attributes[1].value, msg.name);
        
        // Verify state was saved correctly
        let config = CONFIG.load(&deps.storage).unwrap();
        assert_eq!(config.name, "Identity DAO");
        assert_eq!(config.description, "Decentralized Identity DAO");
        assert_eq!(config.voting_config.threshold, cosmwasm_std::Decimal::percent(60));
        assert_eq!(config.did_integration_enabled, true);
        
        let modules = MODULES.load(&deps.storage).unwrap();
        assert_eq!(modules.voting_module, None);
        assert_eq!(modules.proposal_module, None);
        assert_eq!(modules.pre_propose_module, None);
        
        let treasury = TREASURY.load(&deps.storage).unwrap();
        assert_eq!(treasury.balance, Uint128::zero());
        assert_eq!(treasury.reserved, Uint128::zero());
        
        let proposal_count = PROPOSAL_COUNT.load(&deps.storage).unwrap();
        assert_eq!(proposal_count, 0);
    }

    #[test]
    fn test_register_voting_module() {
        let (mut deps, env, _) = setup_contract();
        let admin_info = mock_info("admin", &[]);
        
        let msg = CoreExecuteMsg::RegisterModule {
            module_type: "voting".to_string(),
            module_address: "voting_module_addr".to_string(),
        };
        
        let res = execute(deps.branch(), env, admin_info, msg).unwrap();
        assert_eq!(res.attributes[0].value, "register_module");
        assert_eq!(res.attributes[1].value, "voting");
        
        let modules = MODULES.load(&deps.storage).unwrap();
        assert_eq!(modules.voting_module, Some(Addr::unchecked("voting_module_addr")));
    }

    #[test]
    fn test_register_proposal_module() {
        let (mut deps, env, _) = setup_contract();
        let admin_info = mock_info("admin", &[]);
        
        let msg = CoreExecuteMsg::RegisterModule {
            module_type: "proposal".to_string(),
            module_address: "proposal_module_addr".to_string(),
        };
        
        let res = execute(deps.branch(), env, admin_info, msg).unwrap();
        assert_eq!(res.attributes[0].value, "register_module");
        assert_eq!(res.attributes[1].value, "proposal");
        
        let modules = MODULES.load(&deps.storage).unwrap();
        assert_eq!(modules.proposal_module, Some(Addr::unchecked("proposal_module_addr")));
    }

    #[test]
    fn test_register_pre_propose_module() {
        let (mut deps, env, _) = setup_contract();
        let admin_info = mock_info("admin", &[]);
        
        let msg = CoreExecuteMsg::RegisterModule {
            module_type: "pre_propose".to_string(),
            module_address: "pre_propose_module_addr".to_string(),
        };
        
        let res = execute(deps.branch(), env, admin_info, msg).unwrap();
        assert_eq!(res.attributes[0].value, "register_module");
        assert_eq!(res.attributes[1].value, "pre_propose");
        
        let modules = MODULES.load(&deps.storage).unwrap();
        assert_eq!(modules.pre_propose_module, Some(Addr::unchecked("pre_propose_module_addr")));
    }

    #[test]
    fn test_unauthorized_register_module() {
        let (mut deps, env, _) = setup_contract();
        let unauthorized_info = mock_info("unauthorized", &[]);
        
        let msg = CoreExecuteMsg::RegisterModule {
            module_type: "voting".to_string(),
            module_address: "voting_module_addr".to_string(),
        };
        
        let err = execute(deps.branch(), env, unauthorized_info, msg).unwrap_err();
        assert!(matches!(err, ContractError::Unauthorized {}));
    }

    #[test]
    fn test_execute_proposal() {
        let (mut deps, env, _) = setup_contract();
        
        // Register voting module first
        let admin_info = mock_info("admin", &[]);
        let register_msg = CoreExecuteMsg::RegisterModule {
            module_type: "voting".to_string(),
            module_address: "voting_module_addr".to_string(),
        };
        execute(deps.branch(), env.clone(), admin_info, register_msg).unwrap();
        
        // Execute proposal from voting module
        let voting_info = mock_info("voting_module_addr", &[]);
        let msg = CoreExecuteMsg::ExecuteProposal { proposal_id: 1 };
        
        let res = execute(deps.branch(), env, voting_info, msg).unwrap();
        assert_eq!(res.attributes[0].value, "execute_proposal");
        assert_eq!(res.attributes[1].value, "1");
        
        // Verify proposal was marked as executed
        let executed = EXECUTED_PROPOSALS.load(&deps.storage, 1).unwrap();
        assert_eq!(executed, true);
    }

    #[test]
    fn test_unauthorized_execute_proposal() {
        let (mut deps, env, _) = setup_contract();
        let unauthorized_info = mock_info("unauthorized", &[]);
        
        let msg = CoreExecuteMsg::ExecuteProposal { proposal_id: 1 };
        
        let err = execute(deps.branch(), env, unauthorized_info, msg).unwrap_err();
        assert!(matches!(err, ContractError::Unauthorized {}));
    }

    #[test]
    fn test_update_config() {
        let (mut deps, env, _) = setup_contract();
        let admin_info = mock_info("admin", &[]);
        
        let msg = CoreExecuteMsg::UpdateConfig {
            name: Some("Updated DAO".to_string()),
            description: Some("Updated description".to_string()),
            voting_config: None,
            admin: Some("new_admin".to_string()),
        };
        
        let res = execute(deps.branch(), env, admin_info, msg).unwrap();
        assert_eq!(res.attributes[0].value, "update_config");
        
        let config = CONFIG.load(&deps.storage).unwrap();
        assert_eq!(config.name, "Updated DAO");
        assert_eq!(config.description, "Updated description");
        assert_eq!(config.admin, Some(Addr::unchecked("new_admin")));
    }

    #[test]
    fn test_deposit_to_treasury() {
        let (mut deps, env, _) = setup_contract();
        let depositor_info = mock_info("depositor", &coins(1000000, "usnr"));
        
        let msg = CoreExecuteMsg::DepositToTreasury {};
        
        let res = execute(deps.branch(), env, depositor_info, msg).unwrap();
        assert_eq!(res.attributes[0].value, "deposit_treasury");
        assert_eq!(res.attributes[1].value, "1000000");
        
        let treasury = TREASURY.load(&deps.storage).unwrap();
        assert_eq!(treasury.balance, Uint128::from(1000000u128));
    }

    #[test]
    fn test_withdraw_from_treasury() {
        let (mut deps, env, _) = setup_contract();
        
        // First deposit some funds
        let depositor_info = mock_info("depositor", &coins(2000000, "usnr"));
        let deposit_msg = CoreExecuteMsg::DepositToTreasury {};
        execute(deps.branch(), env.clone(), depositor_info, deposit_msg).unwrap();
        
        // Register voting module
        let admin_info = mock_info("admin", &[]);
        let register_msg = CoreExecuteMsg::RegisterModule {
            module_type: "voting".to_string(),
            module_address: "voting_module_addr".to_string(),
        };
        execute(deps.branch(), env.clone(), admin_info, register_msg).unwrap();
        
        // Withdraw from treasury (through voting module)
        let voting_info = mock_info("voting_module_addr", &[]);
        let msg = CoreExecuteMsg::WithdrawFromTreasury {
            recipient: "recipient".to_string(),
            amount: Uint128::from(1000000u128),
            denom: "usnr".to_string(),
        };
        
        let res = execute(deps.branch(), env, voting_info, msg).unwrap();
        assert_eq!(res.attributes[0].value, "withdraw_treasury");
        assert_eq!(res.attributes[1].value, "1000000");
        
        // Check bank message was created
        assert_eq!(res.messages.len(), 1);
        
        let treasury = TREASURY.load(&deps.storage).unwrap();
        assert_eq!(treasury.balance, Uint128::from(1000000u128));
    }

    #[test]
    fn test_query_config() {
        let (deps, _, _) = setup_contract();
        
        let res = query(deps.as_ref(), mock_env(), CoreQueryMsg::GetConfig {}).unwrap();
        let config_response: DaoConfigResponse = from_json(&res).unwrap();
        
        assert_eq!(config_response.name, "Test Identity DAO");
        assert_eq!(config_response.description, "A test DAO for identity management");
        assert_eq!(config_response.did_integration_enabled, true);
    }

    #[test]
    fn test_query_modules() {
        let (mut deps, env, _) = setup_contract();
        
        // Register all modules
        let admin_info = mock_info("admin", &[]);
        
        execute(deps.branch(), env.clone(), admin_info.clone(), CoreExecuteMsg::RegisterModule {
            module_type: "voting".to_string(),
            module_address: "voting_addr".to_string(),
        }).unwrap();
        
        execute(deps.branch(), env.clone(), admin_info.clone(), CoreExecuteMsg::RegisterModule {
            module_type: "proposal".to_string(),
            module_address: "proposal_addr".to_string(),
        }).unwrap();
        
        execute(deps.branch(), env.clone(), admin_info, CoreExecuteMsg::RegisterModule {
            module_type: "pre_propose".to_string(),
            module_address: "pre_propose_addr".to_string(),
        }).unwrap();
        
        let res = query(deps.as_ref(), mock_env(), CoreQueryMsg::GetModules {}).unwrap();
        let modules_response: ModuleConfig = from_json(&res).unwrap();
        
        assert_eq!(modules_response.voting_module, Some("voting_addr".to_string()));
        assert_eq!(modules_response.proposal_module, Some("proposal_addr".to_string()));
        assert_eq!(modules_response.pre_propose_module, Some("pre_propose_addr".to_string()));
    }

    #[test]
    fn test_query_treasury() {
        let (mut deps, env, _) = setup_contract();
        
        // Deposit funds
        let depositor_info = mock_info("depositor", &coins(5000000, "usnr"));
        let deposit_msg = CoreExecuteMsg::DepositToTreasury {};
        execute(deps.branch(), env, depositor_info, deposit_msg).unwrap();
        
        let res = query(deps.as_ref(), mock_env(), CoreQueryMsg::GetTreasury {}).unwrap();
        let treasury_response: TreasuryInfo = from_json(&res).unwrap();
        
        assert_eq!(treasury_response.balance, Uint128::from(5000000u128));
        assert_eq!(treasury_response.reserved, Uint128::zero());
    }

    #[test]
    fn test_query_dao_stats() {
        let (mut deps, env, _) = setup_contract();
        
        // Register voting module and execute some proposals
        let admin_info = mock_info("admin", &[]);
        let register_msg = CoreExecuteMsg::RegisterModule {
            module_type: "voting".to_string(),
            module_address: "voting_module_addr".to_string(),
        };
        execute(deps.branch(), env.clone(), admin_info, register_msg).unwrap();
        
        let voting_info = mock_info("voting_module_addr", &[]);
        
        // Execute multiple proposals
        for i in 1..=3 {
            let msg = CoreExecuteMsg::ExecuteProposal { proposal_id: i };
            execute(deps.branch(), env.clone(), voting_info.clone(), msg).unwrap();
        }
        
        let res = query(deps.as_ref(), mock_env(), CoreQueryMsg::GetStats {}).unwrap();
        let stats_response: DaoStatsResponse = from_json(&res).unwrap();
        
        assert_eq!(stats_response.total_proposals, 3);
        assert_eq!(stats_response.executed_proposals, 3);
        assert_eq!(stats_response.treasury_balance, Uint128::zero());
    }

    #[test]
    fn test_is_proposal_executed() {
        let (mut deps, env, _) = setup_contract();
        
        // Register voting module
        let admin_info = mock_info("admin", &[]);
        let register_msg = CoreExecuteMsg::RegisterModule {
            module_type: "voting".to_string(),
            module_address: "voting_module_addr".to_string(),
        };
        execute(deps.branch(), env.clone(), admin_info, register_msg).unwrap();
        
        // Execute proposal
        let voting_info = mock_info("voting_module_addr", &[]);
        let msg = CoreExecuteMsg::ExecuteProposal { proposal_id: 42 };
        execute(deps.branch(), env, voting_info, msg).unwrap();
        
        // Query if proposal is executed
        let res = query(deps.as_ref(), mock_env(), CoreQueryMsg::IsProposalExecuted { 
            proposal_id: 42 
        }).unwrap();
        let is_executed: bool = from_json(&res).unwrap();
        assert_eq!(is_executed, true);
        
        // Query non-executed proposal
        let res = query(deps.as_ref(), mock_env(), CoreQueryMsg::IsProposalExecuted { 
            proposal_id: 99 
        }).unwrap();
        let is_executed: bool = from_json(&res).unwrap();
        assert_eq!(is_executed, false);
    }

    #[test]
    fn test_execute_proposal_messages() {
        let (mut deps, env, _) = setup_contract();
        
        // Register voting module
        let admin_info = mock_info("admin", &[]);
        let register_msg = CoreExecuteMsg::RegisterModule {
            module_type: "voting".to_string(),
            module_address: "voting_module_addr".to_string(),
        };
        execute(deps.branch(), env.clone(), admin_info, register_msg).unwrap();
        
        // Execute proposal with messages
        let voting_info = mock_info("voting_module_addr", &[]);
        let bank_msg = CosmosMsg::Bank(BankMsg::Send {
            to_address: "recipient".to_string(),
            amount: coins(100000, "usnr"),
        });
        
        let msg = CoreExecuteMsg::ExecuteProposalWithMessages {
            proposal_id: 1,
            messages: vec![bank_msg.clone()],
        };
        
        let res = execute(deps.branch(), env, voting_info, msg).unwrap();
        assert_eq!(res.messages.len(), 1);
        assert_eq!(res.messages[0].msg, bank_msg);
    }

    #[test]
    fn test_double_execution_prevention() {
        let (mut deps, env, _) = setup_contract();
        
        // Register voting module
        let admin_info = mock_info("admin", &[]);
        let register_msg = CoreExecuteMsg::RegisterModule {
            module_type: "voting".to_string(),
            module_address: "voting_module_addr".to_string(),
        };
        execute(deps.branch(), env.clone(), admin_info, register_msg).unwrap();
        
        // Execute proposal first time
        let voting_info = mock_info("voting_module_addr", &[]);
        let msg = CoreExecuteMsg::ExecuteProposal { proposal_id: 1 };
        execute(deps.branch(), env.clone(), voting_info.clone(), msg.clone()).unwrap();
        
        // Try to execute same proposal again
        let err = execute(deps.branch(), env, voting_info, msg).unwrap_err();
        assert!(matches!(err, ContractError::ProposalAlreadyExecuted { .. }));
    }
}