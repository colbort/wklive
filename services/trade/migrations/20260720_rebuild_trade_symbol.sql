-- 开发阶段破坏性迁移：无兼容和数据回填要求时，直接按 trade.sql 重建交易标的相关表。
SET FOREIGN_KEY_CHECKS = 0;
DROP TABLE IF EXISTS `t_trade_symbol_seconds`;
DROP TABLE IF EXISTS `t_trade_symbol_contract`;
DROP TABLE IF EXISTS `t_trade_symbol_spot`;
DROP TABLE IF EXISTS `t_trade_symbol`;
SET FOREIGN_KEY_CHECKS = 1;

-- 表定义以 services/trade/trade.sql 为唯一来源。
