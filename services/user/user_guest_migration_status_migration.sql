ALTER TABLE `t_user`
  ADD COLUMN `guest_migrated_origin` varchar(255) NOT NULL DEFAULT ''
    COMMENT '游客迁移成功的目标域名Origin'
    AFTER `source_origin`,
  ADD COLUMN `guest_migrated_time` bigint NOT NULL DEFAULT 0
    COMMENT '游客迁移成功时间'
    AFTER `guest_migrated_origin`,
  DROP INDEX `idx_tenant_guest_origin_activity`,
  ADD KEY `idx_tenant_guest_origin_activity`
    (`tenant_id`, `is_guest`, `source_origin`, `deleted`, `guest_migrated_time`, `last_login_time`);
