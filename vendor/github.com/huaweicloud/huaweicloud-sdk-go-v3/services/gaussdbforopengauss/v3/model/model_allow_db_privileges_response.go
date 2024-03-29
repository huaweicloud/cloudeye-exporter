package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type AllowDbPrivilegesResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o AllowDbPrivilegesResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "AllowDbPrivilegesResponse struct{}"
	}

	return strings.Join([]string{"AllowDbPrivilegesResponse", string(data)}, " ")
}
