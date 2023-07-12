package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// function input when grant policy
type FunctionInfo struct {

	// function name
	Name string `json:"name"`
}

func (o FunctionInfo) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "FunctionInfo struct{}"
	}

	return strings.Join([]string{"FunctionInfo", string(data)}, " ")
}
