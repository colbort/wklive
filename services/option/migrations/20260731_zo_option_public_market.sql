-- OPT-P2-006：公开期权盘口查询索引。
-- 可重复执行；不修改业务数据。

SET @option_public_book_index_exists = (
  SELECT COUNT(1)
  FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 't_option_order'
    AND index_name = 'idx_option_public_book'
);
SET @option_public_book_index_sql = IF(
  @option_public_book_index_exists = 0,
  'ALTER TABLE `t_option_order` ADD INDEX `idx_option_public_book` (`tenant_id`,`contract_id`,`status`,`side`,`price`,`id`)',
  'SELECT 1'
);
PREPARE option_public_book_index_stmt FROM @option_public_book_index_sql;
EXECUTE option_public_book_index_stmt;
DEALLOCATE PREPARE option_public_book_index_stmt;

SET @option_public_chain_index_exists = (
  SELECT COUNT(1)
  FROM information_schema.statistics
  WHERE table_schema = DATABASE()
    AND table_name = 't_option_contract'
    AND index_name = 'idx_option_public_chain'
);
SET @option_public_chain_index_sql = IF(
  @option_public_chain_index_exists = 0,
  'ALTER TABLE `t_option_contract` ADD INDEX `idx_option_public_chain` (`tenant_id`,`underlying_symbol`,`expire_time`,`status`,`is_deleted`,`strike_price`,`option_type`,`id`)',
  'SELECT 1'
);
PREPARE option_public_chain_index_stmt FROM @option_public_chain_index_sql;
EXECUTE option_public_chain_index_stmt;
DEALLOCATE PREPARE option_public_chain_index_stmt;
