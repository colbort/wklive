# BTCUSDT 交割合约启用材料

## 已核实事实

- Symbol ID：7；
- Symbol：`BTCUSDT-20260925`；
- 租户：1；
- 当前状态：2（停用）；
- 交割时间：2026-09-25 16:00:00 Asia/Hong_Kong；
- 停止开仓：2026-09-25 15:00:00；
- 停止撮合：2026-09-25 15:30:00；
- DELIVERY 锁价窗口：30 秒；
- 结算算法/公式版本：`delivery-v1`；
- 合约、逐仓杠杆和风险档位配置完整；
- 当前 Order、Position、Position History 均为 0；
- 三源 DELIVERY 输出新鲜；
- 因前四组生产门禁未通过，产品不得启用。

## 发布前检查

1. `contract-readiness` 除“交割合约未启用”外必须全部 PASS；
2. 重新确认交割时间仍在未来且与业务日历一致；
3. Order、Position、Reservation、Settlement Instruction 和 Delivery Batch 无未完成事实；
4. 三源 DELIVERY、INDEX、MARK、FUNDING 均新鲜；
5. 保险基金达到审批水位；
6. 自动强平和全仓开关仍保持关闭；
7. 获取独立产品启用发布单、操作人和复核人。

满足后才通过 `POST /admin/trade/symbols/update` 将 Symbol 7 状态改为启用。请求必须从
详情接口读回完整现值，仅修改 `status`，避免覆盖时间、精度或交易方向配置。

## 启用后验收

- 再次运行 `contract-readiness`，要求零 FAIL、输出 `READY`；
- 验证行情、下单门禁、到期状态机和取消/释放边界；
- 保存 Admin 操作日志、前后详情、发布单和终检输出；
- `READY` 不自动打开自动强平或全仓开关。

