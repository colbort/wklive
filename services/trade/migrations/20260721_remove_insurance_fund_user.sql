ALTER TABLE `t_contract_insurance_fund_account`
  DROP CHECK `chk_insurance_fund_account`,
  DROP COLUMN `fund_user_id`,
  DROP COLUMN `wallet_type`,
  ADD CONSTRAINT `chk_insurance_fund_account` CHECK (`adl_enabled` IN (1,2) AND `status` IN (1,2));
