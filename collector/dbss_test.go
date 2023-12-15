package collector

import (
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/rms/v1/model"
	"github.com/stretchr/testify/assert"
)

func TestDBssGetResourceInfo(t *testing.T) {
	sysConfig := map[string][]string{"audit_id": {"cpu_util"}}
	patches := gomonkey.ApplyFuncReturn(getMetricConfigMap, sysConfig)
	patches.ApplyFuncReturn(listResources, mockRmsResource(), nil)
	defer patches.Reset()

	var dbssgetter DBSSInfo
	labels, metrics := dbssgetter.GetResourceInfo()
	assert.Equal(t, 1, len(labels))
	assert.Equal(t, 1, len(metrics))
}

func mockRmsResource() []model.ResourceEntity {
	id := "0001-0001-000001"
	name := "resource1"
	epId := "0"
	resp := []model.ResourceEntity{
		{Id: &id, Name: &name, EpId: &epId},
	}
	return resp
}
