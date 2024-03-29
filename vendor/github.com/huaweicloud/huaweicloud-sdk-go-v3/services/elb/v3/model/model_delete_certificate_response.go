package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type DeleteCertificateResponse struct {
	HttpStatusCode int `json:"-"`
}

func (o DeleteCertificateResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "DeleteCertificateResponse struct{}"
	}

	return strings.Join([]string{"DeleteCertificateResponse", string(data)}, " ")
}
