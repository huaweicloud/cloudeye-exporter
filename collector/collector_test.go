package collector

import (
	"sync"
	"testing"
	"time"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap/zapcore"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

func TestReplaceName(t *testing.T) {
	assert.Equal(t, "sys_ecs", replaceName("SYS.ECS"))
}

func TestGetLabel(t *testing.T) {
	ns := "SYS.ECS"
	avg := 23.5
	unit := "%"
	metric := model.BatchMetricData{
		Namespace:  &ns,
		MetricName: "cpu_utils",
		Dimensions: &[]model.MetricsDimension{
			{Name: "instance_id", Value: "0001-0001-0000001"},
		},
		Datapoints: []model.DatapointForBatchMetric{
			{Average: &avg, Timestamp: time.Now().Unix() * 1000},
		},
		Unit: &unit,
	}
	info := map[string]labelInfo{
		"0001-0001-0000001": {Name: []string{"name"}, Value: []string{"host01"}},
	}
	label := getLabel(metric, info)
	assert.Equal(t, 2, len(label.Name))
	assert.Equal(t, 2, len(label.Value))
}

func TestGetDimValue(t *testing.T) {
	conf.AccessKey = "test_ak"
	conf.SecretKey = "test_sk"
	patches := getPatches()
	defer patches.Reset()
	logs.InitLog("")
	sysECSNamespace := "SYS.ECS"
	agtECSNamespace := "AGT.ECS"
	sysECSMetricData := model.BatchMetricData{
		Namespace: &sysECSNamespace,
		Dimensions: &[]model.MetricsDimension{
			{Name: "instance_id", Value: "0001-0001-0000001"},
		},
	}
	agtECSMetricData := model.BatchMetricData{
		Namespace: &agtECSNamespace,
		Dimensions: &[]model.MetricsDimension{
			{Name: "instance_id", Value: "0001-0001-0000001"},
		},
	}
	assert.Equal(t, "0001-0001-0000001", getDimValue(sysECSMetricData, "instance_id", "0001-0001-0000001"))
	assert.Equal(t, "0001-0001-0000001", getDimValue(agtECSMetricData, "instance_id", "0001-0001-0000001"))
	assert.Equal(t, "000000000000000", getDimValue(agtECSMetricData, "disk", "000000000000000"))
	agentDimensions.Store("000000000000000", AgentDimensionsValue{
		originValue: "vda",
	})
	assert.Equal(t, "vda", getDimValue(agtECSMetricData, "disk", "000000000000000"))
}

func TestTransMetric(t *testing.T) {
	metricInfoList := model.MetricInfoList{
		Namespace: "SYS.ECS",
	}
	assert.Equal(t, "SYS.ECS", transMetric(metricInfoList).Namespace)
}

func TestGetLatestData(t *testing.T) {
	datapoint, err := getLatestData([]model.DatapointForBatchMetric{})
	assert.Equal(t, true, datapoint == nil)
	assert.Equal(t, "data not found", err.Error())
	avg1, avg2 := 23.5, 63.52
	data := []model.DatapointForBatchMetric{
		{Average: &avg1, Timestamp: time.Now().Unix()*1000 - 1000*60},
		{Average: &avg2, Timestamp: time.Now().Unix() * 1000},
	}
	datapoint, err = getLatestData(data)
	assert.Equal(t, true, err == nil)
	assert.Equal(t, avg2, *datapoint.Average)
}

func TestIsMetricLabelConflict(t *testing.T) {
	var label = labelInfo{
		Name:  []string{"disk", "instance_id", "unit"},
		Value: []string{"0eb113c0-0111-4f51-9d44-f2157fcc138c", "2e84018fc8b4484b94e89aae212fe615", "%"},
	}
	proMap := PrometheusMetricMap{
		RWMutex:   sync.RWMutex{},
		MetricMap: make(map[string]bool),
	}
	proMap.MetricMap["huaweicloud_agt_ecs_cpu_usage{instance_id=9234ad9f-87a5-49a9-b6ec-11ecde1d943b,mount_point=2e84018fc8b4484b94e89aae212fe615,unit=%}"] = true
	result1 := isMetricLabelConflict("huaweicloud_agt_ecs_cpu_usage", label, &proMap)
	assert.False(t, result1)

	proMap.MetricMap["huaweicloud_agt_ecs_cpu_usage{instance_id=9234ad9f-87a5-49a9-b6ec-11ecde1d943b,disk=2e84018fc8b4484b94e89aae212fe615,unit=%}"] = true
	result2 := isMetricLabelConflict("huaweicloud_agt_ecs_cpu_usage", label, &proMap)
	assert.True(t, result2)
}

func TestIsAgentMetric(t *testing.T) {
	resultECS := isAgentMetric("AGT.ECS")
	assert.True(t, resultECS)

	resultBMS := isAgentMetric("SERVICE.BMS")
	assert.True(t, resultBMS)
}

func TestSetProData1(t *testing.T) {
	proMap := PrometheusMetricMap{
		RWMutex:   sync.RWMutex{},
		MetricMap: make(map[string]bool),
	}
	proMap.MetricMap["agt_ecs_cpu_usage{instance_id=9234ad9f-87a5-49a9-b6ec-11ecde1d943b,mount_point=2e84018fc8b4484b94e89aae212fe615,unit=%}"] = true

	unit := "%"
	namespace := "AGT.ECS"
	avg := 0.33
	data := model.BatchMetricData{
		Unit: &unit,
		Datapoints: []model.DatapointForBatchMetric{
			{
				Average: &avg,
			},
		},
		Namespace:  &namespace,
		MetricName: "cpu_usage",
		Dimensions: &[]model.MetricsDimension{
			{
				Name:  "instance_id",
				Value: "9234ad9f-87a5-49a9-b6ec-11ecde1d943b",
			},
			{
				Name:  "mount_point",
				Value: "2e84018fc8b4484b94e89aae212fe615",
			},
		},
	}
	metricDataArray := []model.BatchMetricData{data}
	var exporter BaseHuaweiCloudExporter
	resourceInfo := make(map[string]labelInfo, 0)
	var label = labelInfo{
		Name:  []string{"mount_point", "instance_id", "unit"},
		Value: []string{"2e84018fc8b4484b94e89aae212fe615", "0eb113c0-0111-4f51-9d44-f2157fcc138c", "%"},
	}
	resourceInfo["instance_id"] = label
	confLoader := logs.ConfLoader{}
	patches := gomonkey.ApplyMethodFunc(confLoader, "LoadFile", func(fPath string, cfg interface{}) error {
		cfgTmp, ok := cfg.(*map[string][]logs.Config)
		assert.True(t, ok)
		cfgPointer := make(map[string][]logs.Config)
		cfgPointer["business"] = []logs.Config{
			{
				Level: zapcore.InfoLevel,
			},
		}
		*cfgTmp = cfgPointer
		return nil
	})
	defer patches.Reset()
	logs.InitLog("")
	exporter.Collect(nil)
	exporter.setProData(nil, nil, metricDataArray, resourceInfo, &proMap)
	assert.Equal(t, 3, len(label.Name))
}

func TestGetEvsInfoForBMS(t *testing.T) {
	testInstanceID := "test-instance-id"
	testDiskName := "vda"
	testEvsID := "evs-12345"

	// 初始化日志和 patches
	patches := getPatches()
	logs.InitLog("")
	// mock loadAgentDimensions 避免调用 CES client
	patches.ApplyFunc(loadAgentDimensions, func(instanceID string) error { return nil })
	defer patches.Reset()

	// 清理 agentDimensions 中测试添加的 key
	defer func() {
		agentDimensions.Delete(testDiskName)
	}()

	// 测试用例1: 正常获取 evsId
	t.Run("normal case: get evsId successfully", func(t *testing.T) {
		// 重置 bmsInfo
		bmsInfo = serversInfo{}
		bmsInfo.ExtendInfo = map[string]interface{}{
			testInstanceID: map[string]string{
				testDiskName: testEvsID,
			},
		}

		// 预先存储 agentDimensions，避免调用 loadAgentDimensions
		agentDimensions.Store(testDiskName, testDiskName)

		namespace := "SERVICE.BMS"
		metric := model.BatchMetricData{
			Namespace: &namespace,
			Dimensions: &[]model.MetricsDimension{
				{Name: "instance_id", Value: testInstanceID},
				{Name: "disk", Value: testDiskName},
			},
		}

		label := labelInfo{}
		getEvsInfoForBMS(metric, &label)

		assert.Contains(t, label.Name, "evsId")
		evsIdx := -1
		for i, name := range label.Name {
			if name == "evsId" {
				evsIdx = i
				break
			}
		}
		assert.Equal(t, testEvsID, label.Value[evsIdx])
	})

	// 测试用例2: diskName 为空
	t.Run("diskName is empty", func(t *testing.T) {
		bmsInfo = serversInfo{}
		bmsInfo.ExtendInfo = map[string]interface{}{
			testInstanceID: map[string]string{
				testDiskName: testEvsID,
			},
		}

		namespace := "SERVICE.BMS"
		metric := model.BatchMetricData{
			Namespace: &namespace,
			Dimensions: &[]model.MetricsDimension{
				{Name: "instance_id", Value: testInstanceID},
				{Name: "disk", Value: ""},
			},
		}

		label := labelInfo{}
		getEvsInfoForBMS(metric, &label)

		assert.NotContains(t, label.Name, "evsId")
	})

	// 测试用例3: instanceID 不在 ExtendInfo 中
	t.Run("instanceID not found in ExtendInfo", func(t *testing.T) {
		bmsInfo = serversInfo{}
		bmsInfo.ExtendInfo = map[string]interface{}{
			"other-instance": map[string]string{
				testDiskName: testEvsID,
			},
		}

		agentDimensions.Store(testDiskName, testDiskName)

		namespace := "SERVICE.BMS"
		metric := model.BatchMetricData{
			Namespace: &namespace,
			Dimensions: &[]model.MetricsDimension{
				{Name: "instance_id", Value: testInstanceID},
				{Name: "disk", Value: testDiskName},
			},
		}

		label := labelInfo{}
		getEvsInfoForBMS(metric, &label)

		// instanceID 不存在时，直接返回，不添加 evsId
		assert.NotContains(t, label.Name, "evsId")
	})

	// 测试用例4: extendInfoMap 转换为 map[string]string 失败
	t.Run("extendInfoMap convert failed", func(t *testing.T) {
		bmsInfo = serversInfo{}
		bmsInfo.ExtendInfo = map[string]interface{}{
			testInstanceID: "not a map", // 错误的类型
		}

		agentDimensions.Store(testDiskName, testDiskName)

		namespace := "SERVICE.BMS"
		metric := model.BatchMetricData{
			Namespace: &namespace,
			Dimensions: &[]model.MetricsDimension{
				{Name: "instance_id", Value: testInstanceID},
				{Name: "disk", Value: testDiskName},
			},
		}

		label := labelInfo{}
		getEvsInfoForBMS(metric, &label)

		// 类型转换失败，不添加 evsId
		assert.NotContains(t, label.Name, "evsId")
	})

	// 测试用例5: diskName 不在 map 中
	t.Run("diskName not found in map", func(t *testing.T) {
		bmsInfo = serversInfo{}
		bmsInfo.ExtendInfo = map[string]interface{}{
			testInstanceID: map[string]string{
				"other-disk": testEvsID,
			},
		}

		agentDimensions.Store(testDiskName, testDiskName)

		namespace := "SERVICE.BMS"
		metric := model.BatchMetricData{
			Namespace: &namespace,
			Dimensions: &[]model.MetricsDimension{
				{Name: "instance_id", Value: testInstanceID},
				{Name: "disk", Value: testDiskName},
			},
		}

		label := labelInfo{}
		getEvsInfoForBMS(metric, &label)

		// diskName 不存在时，添加 evsId 但值为空
		assert.Contains(t, label.Name, "evsId")
		evsIdx := -1
		for i, name := range label.Name {
			if name == "evsId" {
				evsIdx = i
				break
			}
		}
		assert.Equal(t, "", label.Value[evsIdx])
	})
}
