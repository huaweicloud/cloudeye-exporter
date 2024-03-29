package model

import (
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/utils"

	"strings"
)

// Response Object
type CreateConfigurationResponse struct {

	// 参数模板ID。
	Id *string `json:"id,omitempty"`

	// 参数模板名称。
	Name *string `json:"name,omitempty"`

	// 参数模板描述。
	Description *string `json:"description,omitempty"`

	// 数据库版本。
	DatastoreVersion *string `json:"datastore_version,omitempty"`

	// 数据库类型。
	DatastoreName *string `json:"datastore_name,omitempty"`

	// 创建时间，格式为\"yyyy-MM-ddTHH:mm:ssZ\"。其中，T指某个时间的开始；Z指时区偏移量，例如北京时间偏移显示为+0800。
	Created *string `json:"created,omitempty"`

	// 更新时间，格式为\"yyyy-MM-ddTHH:mm:ssZ\"。其中，T指某个时间的开始；Z指时区偏移量，例如北京时间偏移显示为+0800。
	Updated        *string `json:"updated,omitempty"`
	HttpStatusCode int     `json:"-"`
}

func (o CreateConfigurationResponse) String() string {
	data, err := utils.Marshal(o)
	if err != nil {
		return "CreateConfigurationResponse struct{}"
	}

	return strings.Join([]string{"CreateConfigurationResponse", string(data)}, " ")
}
