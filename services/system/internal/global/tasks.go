package global

import (
	"wklive/proto/itick"
	"wklive/proto/option"
	"wklive/proto/staking"
	"wklive/proto/trade"
)

var (
	ItickTaskCli   itick.ItickTaskClient
	OptionTaskCli  option.OptionTaskClient
	StakingTaskCli staking.StakingTaskClient
	TradeTaskCli   trade.TradeTaskClient
)
