-- 永续/交割合约生产职责角色。
-- 这些角色由 db-init 管理，仅授予履行职责所需的后台菜单；生产审批仍需真实人员留痕。

INSERT INTO sys_role
  (tenant_id, app_scope, name, code, enabled, remark, create_times, update_times)
VALUES
  (0, 1, '合约生产值班', 'contract_oncall', 1,
   '接收合约行情、Outbox 和对账告警并执行一线处置',
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (0, 1, '保险基金操作员', 'insurance_fund_operator', 1,
   '执行已审批的保险基金账户配置和幂等调账',
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (0, 1, '灾备操作员', 'disaster_recovery_operator', 1,
   '执行灾备演练并查看任务、登录和操作审计',
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (0, 1, '交割发布操作员', 'delivery_release_operator', 1,
   '在批准窗口内执行交割合约配置并观察交割结果',
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (0, 1, '生产发布复核员', 'production_reviewer', 1,
   '只读复核行情、资金、交割、对账和审计事实',
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000),
  (0, 1, '生产发布审批员', 'production_approver', 1,
   '只读核验生产材料并承担最终系统审批身份',
   UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000, UNIX_TIMESTAMP(CURRENT_TIMESTAMP(3)) * 1000)
ON DUPLICATE KEY UPDATE
  name = VALUES(name),
  enabled = 1,
  remark = VALUES(remark),
  update_times = VALUES(update_times);

-- 角色权限由该迁移收敛，避免后续重复执行时残留超出职责的权限。
DELETE rm
FROM sys_role_menu rm
JOIN sys_role r ON r.id = rm.role_id
WHERE r.tenant_id = 0
  AND r.app_scope = 1
  AND r.code IN (
    'contract_oncall',
    'insurance_fund_operator',
    'disaster_recovery_operator',
    'delivery_release_operator',
    'production_reviewer',
    'production_approver'
  );

-- 合约生产值班：三类 WS 告警对应的读取权限，以及关键运行事实和审计日志。
INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT 0, r.id, m.id
FROM sys_role r
JOIN sys_menu m ON m.id IN (
  400, 480, 490,
  1000, 1160, 1161, 1170, 1193, 1195, 1196,
  10000, 10500, 10700, 10900
)
WHERE r.tenant_id = 0 AND r.app_scope = 1 AND r.code = 'contract_oncall';

-- 保险基金操作员：只允许保险基金账户、平台账户调账及相关资金/风险事实。
INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT 0, r.id, m.id
FROM sys_role r
JOIN sys_menu m ON m.id IN (
  500, 530,
  1000, 1170, 1181, 1182, 1184, 1185, 1186, 1193
)
WHERE r.tenant_id = 0 AND r.app_scope = 1 AND r.code = 'insurance_fund_operator';

-- 灾备操作员：后台仅授予只读任务及审计权限；主机、备份存储和 KMS 权限由基础设施侧另行控制。
INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT 0, r.id, m.id
FROM sys_role r
JOIN sys_menu m ON m.id IN (10000, 10500, 10600, 10700, 10800, 10900)
WHERE r.tenant_id = 0 AND r.app_scope = 1 AND r.code = 'disaster_recovery_operator';

-- 交割发布操作员：可修改合约交易对配置，其余为发布观察所需只读权限。
INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT 0, r.id, m.id
FROM sys_role r
JOIN sys_menu m ON m.id IN (
  400, 480, 490,
  1000, 1010, 1011, 1015, 1160, 1161, 1193
)
WHERE r.tenant_id = 0 AND r.app_scope = 1 AND r.code = 'delivery_release_operator';

-- 生产复核和审批账号均为只读，不能自行注资或启用交割合约。
INSERT IGNORE INTO sys_role_menu (tenant_id, role_id, menu_id)
SELECT 0, r.id, m.id
FROM sys_role r
JOIN sys_menu m ON m.id IN (
  400, 480, 490,
  500, 530,
  1000, 1010, 1011, 1160, 1161, 1170, 1181, 1184, 1193, 1195, 1196,
  10000, 10500, 10600, 10700, 10800, 10900
)
WHERE r.tenant_id = 0
  AND r.app_scope = 1
  AND r.code IN ('production_reviewer', 'production_approver');

