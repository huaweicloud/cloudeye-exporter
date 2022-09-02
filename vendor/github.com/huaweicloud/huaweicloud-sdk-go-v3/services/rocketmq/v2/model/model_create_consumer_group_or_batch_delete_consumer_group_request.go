package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Request Object
type CreateConsumerGroupOrBatchDeleteConsumerGroupRequest struct {

	// 实例ID。
	InstanceId string `json:"instance_id"`

	// 批量删除topic时使用，不配置则为创建接口。
	Action *string `json:"action,omitempty"`

	Body *CreateConsumerGroupOrBatchDeleteConsumerGroupReq `json:"body,omitempty"`
}

func (o CreateConsumerGroupOrBatchDeleteConsumerGroupRequest) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreateConsumerGroupOrBatchDeleteConsumerGroupRequest struct{}"
	}

	return strings.Join([]string{"CreateConsumerGroupOrBatchDeleteConsumerGroupRequest", string(data)}, " ")
}
