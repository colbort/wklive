ALTER TABLE `t_itick_authoritative_snapshot`
  DROP INDEX `uk_authority_product_revision`,
  DROP INDEX `idx_product_time`,
  ADD UNIQUE KEY `uk_authority_product_revision` (`authority`,`snapshot_kind`,`category_code`,`market`,`symbol`,`source_timestamp`,`revision`),
  ADD KEY `idx_product_time` (`authority`,`snapshot_kind`,`category_code`,`market`,`symbol`,`source_timestamp`,`revision`);
