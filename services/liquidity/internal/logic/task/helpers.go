package tasklogic

import (
	"fmt"

	"wklive/common/helper"
	"wklive/proto/liquidity"
)

func validateTask(in *liquidity.LiquidityTaskReq) error {
	if in == nil {
		return fmt.Errorf("task request is required")
	}
	if in.ConfigId < 0 || in.ProviderId < 0 || in.BatchSize < 0 {
		return fmt.Errorf("task filters cannot be negative")
	}
	return nil
}

func taskDependencyUnavailable(name string) *liquidity.LiquidityTaskResp {
	return &liquidity.LiquidityTaskResp{
		Base:        helper.ErrResp(503, name+" executor is not configured"),
		FailedCount: 1,
	}
}
