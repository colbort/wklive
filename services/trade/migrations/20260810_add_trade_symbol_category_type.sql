-- 将交易对绑定到 Market 的六种 category，区分行情分类与交易产品形态。
-- product_type 仍然表示现货/衍生品/秒合约，contract_type 仍然表示永续/交割。

ALTER TABLE `t_trade_symbol`
  ADD COLUMN `category_type` TINYINT NOT NULL DEFAULT 0
    COMMENT 'Market分类：1外汇 2加密货币 3股票 4期货 5指数 6基金，0表示历史数据未绑定'
    AFTER `tenant_id`;

ALTER TABLE `t_trade_symbol`
  ADD KEY `idx_tenant_category_product_status`
    (`tenant_id`, `category_type`, `product_type`, `status`);
