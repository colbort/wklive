UPDATE sys_user u
JOIN (
  SELECT DISTINCT ur.user_id
  FROM sys_user_role ur
  JOIN sys_role_menu rm ON rm.role_id = ur.role_id
  WHERE rm.menu_id = 1080
) affected ON affected.user_id = u.id
SET u.perms_ver = u.perms_ver + 1;
