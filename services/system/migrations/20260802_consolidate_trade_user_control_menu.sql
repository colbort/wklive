-- Trade 用户控制只保留一个可见入口，其余记录保留为按钮/API权限。
UPDATE sys_menu
SET name='用户交易控制', menu_type=2, method='GET', path='/trade/user-trade-limit',
    perms='trade:user-trade-limit:detail', component='trade/risk-controls', icon='Operation',
    parent_id=1000, sort=1080
WHERE id=1080;

UPDATE sys_menu
SET parent_id=1080, menu_type=3, component='', icon=''
WHERE id IN (1081,1090,1091,1100,1101,1102,1103,1110,1120,1121);

UPDATE sys_menu SET name='获取用户交易对限制' WHERE id=1090;
UPDATE sys_menu SET name='获取旧用户交易配置' WHERE id=1100;
UPDATE sys_menu SET name='设置旧用户交易配置' WHERE id=1101;
UPDATE sys_menu SET name='获取用户合约偏好' WHERE id=1102;
UPDATE sys_menu SET name='设置用户合约偏好' WHERE id=1103;
UPDATE sys_menu SET name='获取用户杠杆偏好' WHERE id=1120;
UPDATE sys_menu SET name='设置用户杠杆偏好' WHERE id=1121;
