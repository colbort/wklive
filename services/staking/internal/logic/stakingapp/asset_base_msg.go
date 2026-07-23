package stakingapplogic

import "wklive/proto/common"

func assetBaseMsg(resp interface{ GetBase() *common.RespBase }) string {
	if resp == nil || resp.GetBase() == nil {
		return ""
	}
	return resp.GetBase().GetMsg()
}
