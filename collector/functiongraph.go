package collector

import (
	"encoding/json"
	"fmt"
	"time"

	cesmodel "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

var functionGraphInfo serversInfo

type functionInfo struct {
	ResourceBaseInfo
	FunctionProperties
}
type FunctionProperties struct {
	FuncName string `json:"func_name"`
	Package  string `json:"package"`
}

type FunctionGraphInfo struct{}

func (getter FunctionGraphInfo) GetResourceInfo() (map[string]labelInfo, []cesmodel.MetricInfoList) {
	resourceInfos := map[string]labelInfo{}
	filterMetrics := make([]cesmodel.MetricInfoList, 0)
	functionGraphInfo.Lock()
	defer functionGraphInfo.Unlock()
	if functionGraphInfo.LabelInfo == nil || time.Now().Unix() > functionGraphInfo.TTL {
		if metricNames, ok := getMetricConfigMap("SYS.FunctionGraph")["package-functionname"]; ok {
			if functions, err := getAllFunctionGraphFromRMS(); err == nil {
				for _, function := range functions {
					metrics := buildSingleDimensionMetrics(metricNames, "SYS.FunctionGraph", "package-functionname", fmt.Sprintf("%s-%s", function.Package, function.FuncName))
					filterMetrics = append(filterMetrics, metrics...)
					info := labelInfo{
						Name:  []string{"epId", "package", "function_name"},
						Value: []string{function.EpId, function.Package, function.FuncName},
					}
					keys, values := getTags(function.Tags)
					info.Name = append(info.Name, keys...)
					info.Value = append(info.Value, values...)
					resourceInfos[GetResourceKeyFromMetricInfo(metrics[0])] = info
				}
			}
		}
		functionGraphInfo.LabelInfo = resourceInfos
		functionGraphInfo.FilterMetrics = filterMetrics
		functionGraphInfo.TTL = time.Now().Add(TTL).Unix()
	}
	return functionGraphInfo.LabelInfo, functionGraphInfo.FilterMetrics
}

func getAllFunctionGraphFromRMS() ([]functionInfo, error) {
	resp, err := listResources("fgs", "functions")
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of fgs.functions, error: %s", err.Error())
		return nil, err
	}
	functions := make([]functionInfo, 0, len(resp))
	for _, resource := range resp {
		properties, err := fmtFunctionProperties(resource.Properties)
		if err != nil {
			continue
		}
		functions = append(functions, functionInfo{
			ResourceBaseInfo:   ResourceBaseInfo{*resource.Id, *resource.Name, *resource.EpId, resource.Tags},
			FunctionProperties: *properties,
		})
	}
	return functions, nil
}

func fmtFunctionProperties(properties map[string]interface{}) (*FunctionProperties, error) {
	bytes, err := json.Marshal(properties)
	if err != nil {
		logs.Logger.Errorf("Marshal function properties error: %s", err.Error())
		return nil, err
	}
	var functionProperties FunctionProperties
	err = json.Unmarshal(bytes, &functionProperties)
	if err != nil {
		logs.Logger.Errorf("Unmarshal to FunctionProperties error: %s", err.Error())
		return nil, err
	}

	return &functionProperties, nil
}
