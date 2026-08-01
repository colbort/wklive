-- Fixture for TestReconcileFullFundsMySQL. Load into a fresh isolated database
-- after Asset/Option schemas and 20260731_zt_option_daily_reconciliation_run.sql.
SET @prior=1785412800000;
SET @day=1785499200000;

INSERT INTO t_user_asset
(tenant_id,user_id,wallet_type,coin,total_amount,available_amount,frozen_amount,locked_amount,enabled,version,create_times,update_times)
VALUES(99,1001,5,'USDT',105,105,0,0,1,2,@prior,@day);

INSERT INTO t_asset_flow
(id,flow_no,tenant_id,user_id,wallet_type,coin,change_type,biz_type,scene_type,biz_id,biz_no,op_type,
 change_amount,before_total_amount,after_total_amount,before_available_amount,after_available_amount,
 before_frozen_amount,after_frozen_amount,before_locked_amount,after_locked_amount,balance_snapshot_version,
 remark,create_times,update_times)
VALUES
(1,'scope2-prior',99,1001,5,'USDT','transfer_in','transfer','transfer_in',0,'scope2-prior-biz',1,
 100,0,100,0,100,0,0,0,0,0,'',@prior,@prior),
(2,'scope2-option',99,1001,5,'USDT','trade_match','option','trade_match',1,'scope2-instruction',1,
 5,100,105,100,105,0,0,0,0,0,'',@day,@day);

INSERT INTO t_option_account
(tenant_id,user_id,account_id,margin_coin,balance,available_balance,frozen_balance,status,create_times,update_times)
VALUES(99,1001,0,'USDT',105,105,0,1,@prior DIV 1000,@day DIV 1000);

INSERT INTO t_option_asset_instruction
(id,tenant_id,instruction_no,biz_no,user_id,account_id,action,coin,amount,step_no,status,asset_flow_no,
 reconciliation_status,reconciled_at,create_times,update_times)
VALUES(1,99,'scope2-instruction','scope2-business',1001,0,4,'USDT',5,1,3,'scope2-option',2,
 @day DIV 1000,@day DIV 1000,@day DIV 1000);

INSERT INTO t_option_bill
(tenant_id,user_id,account_id,biz_no,ref_type,ref_id,coin,change_amount,balance_before,balance_after,remark,create_times)
VALUES(99,1001,0,'scope2-instruction',2,1,'USDT',5,100,105,
 'Asset authoritative flow scope2-option',@day DIV 1000);

INSERT INTO t_asset_platform_account
(id,tenant_id,account_type,coin,available_amount,frozen_amount,status,version,create_times,update_times)
VALUES(1,99,'FEE_REVENUE','USDT',2,0,1,2,@prior,@day);

INSERT INTO t_asset_platform_flow
(id,tenant_id,platform_account_id,account_type,coin,op_type,amount,before_available,after_available,
 biz_type,scene_type,biz_id,biz_no,remark,create_times)
VALUES
(1,99,1,'FEE_REVENUE','USDT',1,1,0,1,'admin','platform_manual_adjust',0,
 'scope2-platform-prior','',@prior),
(2,99,1,'FEE_REVENUE','USDT',1,1,1,2,'option','trade_fee',1,
 'scope2-platform-day','',@day);
