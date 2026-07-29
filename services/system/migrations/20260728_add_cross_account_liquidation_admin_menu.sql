-- 全仓账户强平列表、详情与受控人工重试入口。
-- 重试只恢复 MANUAL_REVIEW Saga，不能修改金额或直接标记完成。
INSERT IGNORE INTO `sys_menu`
(`id`, `parent_id`, `app_scope`, `name`, `menu_type`, `method`, `path`, `perms`, `component`, `icon`, `sort`)
VALUES
(1195, 1000, 1, '全仓账户强平', 2, 'GET', '/trade/account-liquidations', 'trade:account-liquidation:list', 'trade/account-liquidations', 'WarningFilled', 1195),
(1196, 1195, 1, '全仓账户强平详情', 3, 'GET', '/trade/account-liquidations/detail', 'trade:account-liquidation:detail', '', '', 1196),
(1197, 1195, 1, '重试全仓账户强平', 3, 'POST', '/trade/account-liquidations/retry', 'trade:account-liquidation:retry', '', '', 1197);

INSERT IGNORE INTO `sys_role_menu` (`tenant_id`, `role_id`, `menu_id`)
VALUES
(0, 1, 1195),(0, 1, 1196),(0, 1, 1197),
(0, 2, 1195),(0, 2, 1196),(0, 2, 1197);
