package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type DisassociateElbResponse struct {

	// 任务ID
	JobId          *string `json:"job_id,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o DisassociateElbResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DisassociateElbResponse struct{}"
	}

	return strings.Join([]string{"DisassociateElbResponse", string(data)}, " ")
}
