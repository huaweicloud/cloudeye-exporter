package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type DeleteAggregationAuthorizationResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteAggregationAuthorizationResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteAggregationAuthorizationResponse struct{}"
	}

	return strings.Join([]string{"DeleteAggregationAuthorizationResponse", string(data)}, " ")
}
