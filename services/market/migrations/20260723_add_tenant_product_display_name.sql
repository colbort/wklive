ALTER TABLE `t_market_tenant_product`
    ADD COLUMN `display_name` VARCHAR(128) NOT NULL DEFAULT ''
        COMMENT '租户自定义展示名称，为空时使用产品展示名称'
        AFTER `app_visible`;
