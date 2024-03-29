package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type RestoreDisasterResponse struct {
	DisasterRecovery *DisasterRecoveryId `json:"disaster_recovery,omitempty"`
	HttpStatusCode   int                 `json:"-"`
}

func (o RestoreDisasterResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "RestoreDisasterResponse struct{}"
	}

	return strings.Join([]string{"RestoreDisasterResponse", string(data)}, " ")
}
