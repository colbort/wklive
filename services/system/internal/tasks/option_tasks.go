package tasks

import (
	"context"
	"fmt"
	"time"

	"wklive/proto/option"
	"wklive/services/system/internal/global"
	"wklive/services/system/internal/plugins/cronx"
	"wklive/services/system/models"
)

const optionTaskRPCTimeout = 5 * time.Minute

func init() {
	cronx.Register("option.ProcessContractLifecycle", "期权合约生命周期处理", runOptionProcessContractLifecycle)
	cronx.Register("option.CleanMarketSnapshots", "期权行情快照清理", runOptionCleanMarketSnapshots)
}

func runOptionProcessContractLifecycle(ctx context.Context, job *models.SysJob) error {
	return callOptionTask(ctx, job, "process contract lifecycle", global.OptionTaskCli.ProcessContractLifecycle)
}

func runOptionCleanMarketSnapshots(ctx context.Context, job *models.SysJob) error {
	return callOptionTask(ctx, job, "clean market snapshots", global.OptionTaskCli.CleanMarketSnapshots)
}

func callOptionTask(
	ctx context.Context,
	_ *models.SysJob,
	action string,
	fn func(context.Context, *option.OptionTaskReq, ...grpcCallOption) (*option.OptionTaskResp, error),
) error {
	rpcCtx, cancel := context.WithTimeout(ctx, optionTaskRPCTimeout)
	defer cancel()

	result, err := fn(rpcCtx, &option.OptionTaskReq{TenantId: 0})
	if err != nil {
		return err
	}
	if result == nil || result.Base == nil {
		return fmt.Errorf("option %s failed, empty response", action)
	}
	if result.Base.Code != 200 {
		return fmt.Errorf("option %s failed, code: %d, message: %s", action, result.Base.Code, result.Base.Msg)
	}
	return nil
}
