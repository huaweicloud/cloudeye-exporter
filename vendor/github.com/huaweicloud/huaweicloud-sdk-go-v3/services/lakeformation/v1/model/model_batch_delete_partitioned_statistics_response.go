package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type BatchDeletePartitionedStatisticsResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o BatchDeletePartitionedStatisticsResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "BatchDeletePartitionedStatisticsResponse struct{}"
	}

	return strings.Join([]string{"BatchDeletePartitionedStatisticsResponse", string(data)}, " ")
}
