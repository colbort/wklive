-- Staking rewards are paid from STAKING_REWARD; early redemption fees are
-- credited to FEE_REVENUE. Accounts are tenant/coin specific and are created
-- and funded explicitly through Asset admin operations before product enable.
ALTER TABLE t_asset_platform_account
  MODIFY COLUMN account_type VARCHAR(32) NOT NULL
  COMMENT 'INSURANCE_FUND/FUNDING_DIFFERENCE/FEE_REVENUE/OPTION_BACKSTOP/STAKING_REWARD';
