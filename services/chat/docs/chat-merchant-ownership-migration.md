# 客服商户主数据迁移

## 目标

客服商户主数据由 System 的 `sys_chat_merchant` 迁移到 Chat 的
`t_chat_merchant`。迁移后 System 只负责平台管理员认证和 RBAC，客服商户
及其主账号、配置由 Chat 在本地事务中维护。

## 上线顺序

1. 备份 System 的 `sys_chat_merchant` 和 Chat 的 `t_chat_user`、
   `t_chat_merchant_info`。
2. 在 Chat 数据库执行 `migrations/20260726_add_chat_merchant.sql`。
3. 从 System 导出仍需保留的 `sys_chat_merchant`，导入 Chat
   `t_chat_merchant`。必须保留原 `id`，因为它已被 Chat 表作为
   `merchant_id` 使用。
4. 核对每个 `t_chat_merchant.id` 都存在唯一的商户主账号
   (`t_chat_user.user_type = 1 AND is_owner = 1`) 和唯一的
   `t_chat_merchant_info`。
5. 按下面的兼容发布阶段发布服务。
6. 验证平台端商户列表、详情、新增、更新、禁用，以及商户管理员登录。

## 兼容发布阶段

不能把最终代码直接按任意顺序滚动发布。旧 `admin-api` 依赖 System 商户
RPC，旧 System 又依赖 Chat 的 `SyncChatMerchantUser`，因此必须拆成下面
四个阶段：

1. **Chat 兼容版本**：增加 `Platform.PlatformChatMerchant*` RPC 和新表读写，同时
   暂时保留 `SyncChatMerchantUser`。
2. **切换 admin-api**：发布直接调用 Chat `Platform.PlatformChatMerchant*` 的版本；
   确认所有旧 `admin-api` 实例均已退出。
3. **移除 System 依赖**：发布删除客服商户 RPC、Chat client 和
   `sys_chat_merchant` 运行时读写的 System；确认所有旧 System 实例均已退出。
4. **Chat 最终版本**：删除 `SyncChatMerchantUser` 服务及兼容逻辑，即本次
   整改后的最终代码。

如果无法构建第一阶段的兼容版本，只能安排维护窗口，同时停止写入并一次性
切换 Chat、admin-api 和 System；不能采用普通滚动发布。

## 导入字段

将旧表字段按名称映射到新表：

```sql
INSERT INTO t_chat_merchant (
  id, merchant_code, merchant_name, enabled, expire_time,
  contact_name, contact_phone, contact_email, remark,
  create_by, create_times, update_by, update_times
) VALUES (...);
```

System 和 Chat 使用不同数据库时，应通过受控导出/导入任务执行，不能写
跨库 SQL。导入完成后执行：

```sql
SELECT COUNT(*) FROM t_chat_merchant;

SELECT m.id, m.merchant_code
FROM t_chat_merchant m
LEFT JOIN t_chat_user u
  ON u.merchant_id = m.id AND u.user_type = 1 AND u.is_owner = 1
LEFT JOIN t_chat_merchant_info i ON i.merchant_id = m.id
WHERE u.id IS NULL OR i.id IS NULL;
```

第二条查询必须返回空集。确认迁移和回滚窗口结束前，不删除
`sys_chat_merchant`，仅停止其运行时读写。
