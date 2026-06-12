package logic

import "github.com/jinzhu/copier"

func copyNonZero(to any, from any) {
	_ = copier.CopyWithOption(to, from, copier.Option{IgnoreEmpty: true})
}
