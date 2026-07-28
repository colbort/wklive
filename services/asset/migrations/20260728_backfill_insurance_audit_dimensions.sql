-- Backfill insurance-fund rows written before the enum/string mapping included
-- BIZ_TYPE_INSURANCE_FUND and its cover/reversal scenes.
START TRANSACTION;

UPDATE t_asset_platform_flow AS flow
JOIN t_asset_insurance_cover AS cover
  ON cover.tenant_id = flow.tenant_id
 AND cover.platform_account_id = flow.platform_account_id
 AND cover.liquidation_id = flow.biz_id
 AND cover.liquidation_no = flow.biz_no
SET flow.biz_type = 'insurance_fund',
    flow.scene_type = 'insurance_fund_cover'
WHERE flow.account_type = 'INSURANCE_FUND'
  AND flow.op_type = 2
  AND (flow.biz_type = '' OR flow.scene_type = '');

UPDATE t_asset_platform_flow AS flow
JOIN t_asset_insurance_cover AS cover
  ON cover.tenant_id = flow.tenant_id
 AND cover.platform_account_id = flow.platform_account_id
 AND cover.liquidation_id = flow.biz_id
 AND cover.covered_amount = flow.amount
SET flow.biz_type = 'insurance_fund',
    flow.scene_type = 'insurance_fund_reversal'
WHERE flow.account_type = 'INSURANCE_FUND'
  AND flow.op_type = 1
  AND cover.status = 2
  AND flow.biz_no <> cover.liquidation_no
  AND (flow.biz_type = '' OR flow.scene_type = '');

UPDATE t_asset_idempotent AS idem
JOIN t_asset_insurance_cover AS cover
  ON cover.tenant_id = idem.tenant_id
 AND cover.liquidation_no = idem.biz_no
SET idem.biz_type = 'insurance_fund',
    idem.scene_type = 'insurance_fund_cover'
WHERE idem.biz_type = ''
  AND idem.scene_type = '';

UPDATE t_asset_idempotent AS idem
JOIN t_asset_platform_flow AS flow
  ON flow.tenant_id = idem.tenant_id
 AND flow.biz_no = idem.biz_no
 AND flow.biz_type = 'insurance_fund'
 AND flow.scene_type = 'insurance_fund_reversal'
SET idem.biz_type = 'insurance_fund',
    idem.scene_type = 'insurance_fund_reversal'
WHERE idem.biz_type = ''
  AND idem.scene_type = '';

COMMIT;
