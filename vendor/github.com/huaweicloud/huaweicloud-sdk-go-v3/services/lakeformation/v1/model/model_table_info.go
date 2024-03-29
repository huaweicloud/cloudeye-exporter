package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// table input when grant policy
type TableInfo struct {
	Columns *ColumnInfo `json:"columns,omitempty"`

	// table name
	Name string `json:"name"`
}

func (o TableInfo) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "TableInfo struct{}"
	}

	return strings.Join([]string{"TableInfo", string(data)}, " ")
}
