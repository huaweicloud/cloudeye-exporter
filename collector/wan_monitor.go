package collector

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

const namespace = "SYS.WANQMonitor"

var wanMonitorInfo serversInfo

type WanMonitorInfo struct{}

type QualityMonitorTasksResponse struct {
	Tasks          []Task `json:"tasks,omitempty"`
	Count          int    `json:"count,omitempty"`
	HttpStatusCode int    `json:"-"`
}

type TaskList struct {
	Tasks []Task
}

type Task struct {
	TaskId    string   `json:"task_id"`
	TaskName  string   `json:"task_name"`
	EpID      string   `json:"enterprise_project_id"`
	Tags      []Tag    `json:"tags"`
	ProbeIds  []string `json:"probes"`
	Protocol  string   `json:"protocol"`
	Frequency int64    `json:"frequency"`
}

type QualityMonitorTasksRequest struct {
	DomainId string `json:"domain_id,omitempty"`
	Offset   int32  `json:"offset,omitempty"`
	Limit    int32  `json:"limit,omitempty"`
}

type ListProbeResponse struct {
	Probes         []Probe `json:"probes"`
	Count          int     `json:"count,omitempty"`
	HttpStatusCode int     `json:"-"`
}

type Probe struct {
	ProbeID   string     `json:"probe_id"`
	Operator  string     `json:"operator"`
	Continent string     `json:"continent"`
	Country   string     `json:"country"`
	Province  string     `json:"province"`
	City      string     `json:"city"`
	Label     ProbeLabel `json:"label"`
}

type ProbeLabel struct {
	EnUs LabelInfo `json:"en-us"`
	ZhCN LabelInfo `json:"zh-cn"`
}

type LabelInfo struct {
	Name     string `json:"name"`
	City     string `json:"city"`
	Country  string `json:"country"`
	Operator string `json:"operator"`
	Province string `json:"province"`
}

func (getter WanMonitorInfo) GetResourceInfo() (map[string]labelInfo, []model.MetricInfoList) {
	filterMetrics := make([]model.MetricInfoList, 0, 0)
	resourceInfos := map[string]labelInfo{}

	wanMonitorInfo.Lock()
	defer wanMonitorInfo.Unlock()
	if wanMonitorInfo.LabelInfo == nil || time.Now().Unix() > wanMonitorInfo.TTL {
		taskMap := getAllMonitorTask()
		probeMap := getAllProbe()
		metrics, err := listAllMetrics(namespace)
		if err != nil {
			logs.Logger.Errorf("Get all metrics error, detail: %s", err.Error())
			return nil, nil
		}
		getTaskIdDimInfo(metrics, &filterMetrics, resourceInfos, taskMap)
		getTaskIdCityDimInfo(metrics, &filterMetrics, resourceInfos, taskMap, probeMap)
		getTaskIdOperatorDimInfo(metrics, &filterMetrics, resourceInfos, taskMap, probeMap)
		getTaskIdProbeIdDimInfo(metrics, &filterMetrics, resourceInfos, taskMap, probeMap)
		getTaskIdOperatorCityDimInfo(metrics, &filterMetrics, resourceInfos, taskMap, probeMap)

		wanMonitorInfo.LabelInfo = resourceInfos
		wanMonitorInfo.FilterMetrics = filterMetrics
	}
	return wanMonitorInfo.LabelInfo, wanMonitorInfo.FilterMetrics
}

func getAllMonitorTask() map[string]Task {
	taskMap := make(map[string]Task)
	allTask := make([]Task, 0, 0)
	urlPath := fmt.Sprintf("/v2/%s/quality-monitor-tasks", conf.DomainID)
	requestDef := genDefaultReqDefWithOffsetAndLimit(urlPath, new(QualityMonitorTasksResponse))
	req := QualityMonitorTasksRequest{
		Offset: 0,
		Limit:  100,
	}

	var total = 100
	for len(allTask) < total {
		resp, err := getHcClient(getEndpoint("ces", "v2")).Sync(req, requestDef)
		taskResp, ok := resp.(*QualityMonitorTasksResponse)
		if !ok {
			logs.Logger.Errorf("Get all task convert to QualityMonitorTasksResponse failed!")
			return taskMap
		}
		if err != nil {
			logs.Logger.Errorf("Get all task error, detail: %s", err.Error())
			return taskMap
		} else {
			allTask = append(allTask, taskResp.Tasks...)
		}
		total = taskResp.Count
	}
	for _, task := range allTask {
		taskMap[task.TaskId] = task
	}
	return taskMap
}

func getAllProbe() map[string]Probe {
	probeMap := make(map[string]Probe)
	allProbe := make([]Probe, 0, 0)
	urlPath := fmt.Sprintf("/v2/%s/quality-monitor-probes", conf.DomainID)
	requestDef := genDefaultReqDefWithOffsetAndLimit(urlPath, new(ListProbeResponse))
	req := QualityMonitorTasksRequest{
		Offset: 0,
		Limit:  100,
	}

	var total = 100
	for len(allProbe) < total {
		resp, err := getHcClient(getEndpoint("ces", "v2")).Sync(req, requestDef)
		probes, ok := resp.(*ListProbeResponse)
		if !ok {
			logs.Logger.Errorf("Get all probe convert to ListProbeResponse failed!")
			return probeMap
		}
		if err != nil {
			logs.Logger.Errorf("Get all probe error, detail: %s", err.Error())
			return probeMap
		} else {
			allProbe = append(allProbe, probes.Probes...)
		}
		total = probes.Count
	}
	for _, probe := range allProbe {
		probeMap[probe.ProbeID] = probe
	}
	return probeMap
}

func getTaskIdDimInfo(metricList []model.MetricInfoList, filterMetric *[]model.MetricInfoList, resourceInfo map[string]labelInfo, taskMap map[string]Task) {
	sysConfigMap := getMetricConfigMap("SYS.WANQMonitor")
	metrics := sysConfigMap["task_id"]
	sort.Strings(metrics)
	for i := range metricList {
		isTaskIdDim := len(metricList[i].Dimensions) == 1 && metricList[i].Dimensions[0].Name == "task_id"
		if isTaskIdDim && ContainsInArray(metrics, metricList[i].MetricName) {
			*filterMetric = append(*filterMetric, metricList[i])
			task := taskMap[metricList[i].Dimensions[0].Value]
			frequency := strconv.FormatInt(task.Frequency, 10)
			info := labelInfo{
				Name:  []string{"taskName", "epId", "protocol", "frequency"},
				Value: []string{task.TaskName, task.EpID, task.Protocol, frequency},
			}
			tagMap := fmtTags(task.Tags)
			keys, values := getTags(tagMap)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfo[GetResourceKeyFromMetricInfo(metricList[i])] = info
		}
	}
}

func getTaskIdCityDimInfo(metricList []model.MetricInfoList, filterMetric *[]model.MetricInfoList, resourceInfo map[string]labelInfo, taskMap map[string]Task, probeMap map[string]Probe) {
	sysConfigMap := getMetricConfigMap("SYS.WANQMonitor")
	metrics := sysConfigMap["task_id,city"]
	sort.Strings(metrics)
	for i := range metricList {
		dimName := getDimsNameKey(metricList[i].Dimensions)
		// 由于接口返回的维度顺序为字母序排序此处需要将原始维度名和接口返回维度名均排序后再比较相同
		if DimNameEquals("task_id,city", dimName) && ContainsInArray(metrics, metricList[i].MetricName) {
			*filterMetric = append(*filterMetric, metricList[i])

			var taskIdValue, cityValue string
			for _, dimension := range metricList[i].Dimensions {
				if dimension.Name == "task_id" {
					taskIdValue = dimension.Value
				} else {
					cityValue = dimension.Value
				}
			}

			task := taskMap[taskIdValue]
			var cityEnName, cityCnName string
			for _, probeId := range task.ProbeIds {
				probe := probeMap[probeId]
				if probe.City == cityValue {
					cityCnName = probe.Label.ZhCN.City
					cityEnName = probe.Label.EnUs.City
				}
			}
			frequency := strconv.FormatInt(task.Frequency, 10)
			info := labelInfo{
				Name:  []string{"taskName", "epId", "cityEnName", "cityCnName", "protocol", "frequency"},
				Value: []string{task.TaskName, task.EpID, cityEnName, cityCnName, task.Protocol, frequency},
			}
			tagMap := fmtTags(task.Tags)
			keys, values := getTags(tagMap)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfo[GetResourceKeyFromMetricInfo(metricList[i])] = info
		}

	}
}

func getTaskIdOperatorDimInfo(metricList []model.MetricInfoList, filterMetric *[]model.MetricInfoList, resourceInfo map[string]labelInfo, taskMap map[string]Task, probeMap map[string]Probe) {
	sysConfigMap := getMetricConfigMap("SYS.WANQMonitor")
	metrics := sysConfigMap["task_id,operator"]
	sort.Strings(metrics)
	for i := range metricList {
		dimName := getDimsNameKey(metricList[i].Dimensions)
		// 由于接口返回的维度顺序为字母序排序此处需要将原始维度名和接口返回维度名均排序后再比较相同
		if DimNameEquals("task_id,operator", dimName) && ContainsInArray(metrics, metricList[i].MetricName) {
			*filterMetric = append(*filterMetric, metricList[i])

			var taskIdValue, operatorValue string
			for _, dimension := range metricList[i].Dimensions {
				if dimension.Name == "task_id" {
					taskIdValue = dimension.Value
				} else {
					operatorValue = dimension.Value
				}
			}

			task := taskMap[taskIdValue]
			frequency := strconv.FormatInt(task.Frequency, 10)
			var operatorEnName, operatorCnName string
			for _, probeId := range task.ProbeIds {
				probe := probeMap[probeId]
				if probe.Operator == operatorValue {
					operatorCnName = probe.Label.ZhCN.Operator
					operatorEnName = probe.Label.EnUs.Operator
				}
			}
			info := labelInfo{
				Name:  []string{"taskName", "epId", "operatorEnName", "operatorCnName", "protocol", "frequency"},
				Value: []string{task.TaskName, task.EpID, operatorEnName, operatorCnName, task.Protocol, frequency},
			}
			tagMap := fmtTags(task.Tags)
			keys, values := getTags(tagMap)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfo[GetResourceKeyFromMetricInfo(metricList[i])] = info
		}
	}
}

func getTaskIdProbeIdDimInfo(metricList []model.MetricInfoList, filterMetric *[]model.MetricInfoList, resourceInfo map[string]labelInfo, taskMap map[string]Task, probeMap map[string]Probe) {
	sysConfigMap := getMetricConfigMap("SYS.WANQMonitor")
	metrics := sysConfigMap["task_id,probe_id"]
	sort.Strings(metrics)
	for i := range metricList {
		dimName := getDimsNameKey(metricList[i].Dimensions)
		// 由于接口返回的维度顺序为字母序排序此处需要将原始维度名和接口返回维度名均排序后再比较相同
		if DimNameEquals("task_id,probe_id", dimName) && ContainsInArray(metrics, metricList[i].MetricName) {
			*filterMetric = append(*filterMetric, metricList[i])

			var taskIdValue, probeValue string
			for _, dimension := range metricList[i].Dimensions {
				if dimension.Name == "task_id" {
					taskIdValue = dimension.Value
				} else {
					probeValue = dimension.Value
				}
			}

			task := taskMap[taskIdValue]
			frequency := strconv.FormatInt(task.Frequency, 10)
			var probeEnName, probeCnName string
			for _, probeId := range task.ProbeIds {
				probe := probeMap[probeId]
				if probe.ProbeID == probeValue {
					probeCnName = probe.Label.ZhCN.Name
					probeEnName = probe.Label.EnUs.Name
				}
			}
			info := labelInfo{
				Name:  []string{"taskName", "epId", "probeCnName", "probeEnName", "protocol", "frequency"},
				Value: []string{task.TaskName, task.EpID, probeCnName, probeEnName, task.Protocol, frequency},
			}
			tagMap := fmtTags(task.Tags)
			keys, values := getTags(tagMap)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfo[GetResourceKeyFromMetricInfo(metricList[i])] = info
		}
	}
}

func getTaskIdOperatorCityDimInfo(metricList []model.MetricInfoList, filterMetric *[]model.MetricInfoList, resourceInfo map[string]labelInfo, taskMap map[string]Task, probeMap map[string]Probe) {
	sysConfigMap := getMetricConfigMap("SYS.WANQMonitor")
	metrics := sysConfigMap["task_id,operator,city"]
	sort.Strings(metrics)
	for i := range metricList {
		dimName := getDimsNameKey(metricList[i].Dimensions)
		// 由于接口返回的维度顺序为字母序排序此处需要将原始维度名和接口返回维度名均排序后再比较相同
		if DimNameEquals("task_id,operator,city", dimName) && ContainsInArray(metrics, metricList[i].MetricName) {
			*filterMetric = append(*filterMetric, metricList[i])

			var taskIdValue, operatorValue, cityValue string
			for _, dimension := range metricList[i].Dimensions {
				if dimension.Name == "task_id" {
					taskIdValue = dimension.Value
				}
				if dimension.Name == "city" {
					cityValue = dimension.Value
				}
				if dimension.Name == "operator" {
					operatorValue = dimension.Value
				}
			}

			task := taskMap[taskIdValue]
			frequency := strconv.FormatInt(task.Frequency, 10)
			var operatorEnName, operatorCnName, cityEnName, cityCnName string
			for _, probeId := range task.ProbeIds {
				probe := probeMap[probeId]
				if probe.City == cityValue {
					cityCnName = probe.Label.ZhCN.City
					cityEnName = probe.Label.EnUs.City
				}
				if probe.Operator == operatorValue {
					operatorCnName = probe.Label.ZhCN.Operator
					operatorEnName = probe.Label.EnUs.Operator
				}
			}
			info := labelInfo{
				Name:  []string{"taskName", "epId", "cityCnName", "cityEnName", "operatorEnName", "operatorCnName", "protocol", "frequency"},
				Value: []string{task.TaskName, task.EpID, cityCnName, cityEnName, operatorEnName, operatorCnName, task.Protocol, frequency},
			}
			tagMap := fmtTags(task.Tags)
			keys, values := getTags(tagMap)
			info.Name = append(info.Name, keys...)
			info.Value = append(info.Value, values...)
			resourceInfo[GetResourceKeyFromMetricInfo(metricList[i])] = info
		}
	}
}
