package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// 资源池
type WorkloadQueueReq struct {
	WorkloadQueue *WorkloadQueue `json:"workload_queue"`
}

func (o WorkloadQueueReq) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "WorkloadQueueReq struct{}"
	}

	return strings.Join([]string{"WorkloadQueueReq", string(data)}, " ")
}
