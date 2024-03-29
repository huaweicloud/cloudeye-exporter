package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 告警通知关闭时间
type NotificationEndTime struct {
}

func (o NotificationEndTime) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "NotificationEndTime struct{}"
	}

	return strings.Join([]string{"NotificationEndTime", string(data)}, " ")
}
