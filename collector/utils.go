package collector

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/def"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/impl"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/core/sdkerr"
	"github.com/huaweicloud/huaweicloud-sdk-go-v3/services/ces/v1/model"
	iam "github.com/huaweicloud/huaweicloud-sdk-go-v3/services/iam/v3"

	"github.com/huaweicloud/cloudeye-exporter/logs"
)

const (
	MinimumResourceInfoSyncInterval = 10
	MaxNamespacesCount              = 1000
	MaxEpsCount                     = 10000

	AuthModePermanentAkSk = "aksk"
	AuthModeEcsAgency     = "ecsagency"
	AuthModeOidcToken     = "oidc"
	GlobalServiceType     = "GlobalService"
	RegionServiceType     = "RegionService"
)

var tagRegexp *regexp.Regexp

func init() {
	var err error
	tagRegexp, err = regexp.Compile("^([a-z]|[A-Z]){1}([a-z]|[A-Z]|_)*$")
	if err != nil {
		logs.Logger.Error("init tag regexp error: %s", err.Error())
	}
}

type MetricInfoListWithTTL struct {
	TTL int64
	model.MetricInfoList
}

type serversInfo struct {
	TTL           int64
	LabelInfo     map[string]labelInfo
	FilterMetrics []model.MetricInfoList
	ExtendInfo    map[string]interface{}
	sync.Mutex
}

type newServersInfo struct {
	TTL           int64
	LabelInfo     map[string]labelInfo
	FilterMetrics []MetricInfoListWithTTL
	sync.Mutex
}

func (o newServersInfo) GetFilteredMetrics() []model.MetricInfoList {
	var resultMetricList []model.MetricInfoList
	for index := range o.FilterMetrics {
		resultMetricList = append(resultMetricList, o.FilterMetrics[index].MetricInfoList)
	}
	return resultMetricList
}

type labelInfo struct {
	Name  []string
	Value []string
}

type RmsInfo struct {
	Id   string
	Name string
	EpId string
	Tags map[string]string
}

func GetResourceKeyFromDimensions(dimensions []model.MetricsDimension) string {
	sort.Slice(dimensions, func(i, j int) bool {
		return dimensions[i].Name < dimensions[j].Name
	})
	dimValuesList := make([]string, 0, len(dimensions))
	for _, dim := range dimensions {
		dimValuesList = append(dimValuesList, dim.Value)
	}
	return strings.Join(dimValuesList, ".")
}

func GetResourceKeyFromMetricInfo(metric model.MetricInfoList) string {
	sort.Slice(metric.Dimensions, func(i, j int) bool {
		return metric.Dimensions[i].Name < metric.Dimensions[j].Name
	})
	dimValuesList := make([]string, 0, len(metric.Dimensions))
	for _, dim := range metric.Dimensions {
		dimValuesList = append(dimValuesList, dim.Value)
	}
	return strings.Join(dimValuesList, ".")
}

func GetResourceKeyFromMetricData(metric model.BatchMetricData) string {
	// DMS实例其他维度不需要适配资源标签，只匹配实例信息
	if *metric.Namespace == "SYS.DMS" {
		return getDmsResourceKey(metric)
	}
	if *metric.Namespace == "AGT.ECS" || *metric.Namespace == "SERVICE.BMS" {
		return getServerResourceKey(metric)
	}
	if *metric.Namespace == "SYS.MRS" {
		return getMrsResourceKey(metric)
	}
	sort.Slice(*metric.Dimensions, func(i, j int) bool {
		return (*metric.Dimensions)[i].Name < (*metric.Dimensions)[j].Name
	})
	dimValuesList := make([]string, 0, len(*metric.Dimensions))
	for _, dim := range *metric.Dimensions {
		dimValuesList = append(dimValuesList, dim.Value)
	}
	return strings.Join(dimValuesList, ".")
}

func getServerResourceKey(metric model.BatchMetricData) string {
	for _, dim := range *metric.Dimensions {
		if dim.Name == "instance_id" {
			return dim.Value
		}
	}
	return ""
}

func getServerResourceKeyFromMetricInfo(metric model.MetricInfoList) string {
	for _, dim := range metric.Dimensions {
		if dim.Name == "instance_id" {
			return dim.Value
		}
	}
	return ""
}

func getDmsResourceKey(metric model.BatchMetricData) string {
	for _, dim := range *metric.Dimensions {
		if dim.Name == "kafka_instance_id" || dim.Name == "rabbitmq_instance_id" || dim.Name == "reliablemq_instance_id" {
			return dim.Value
		}
	}
	return ""
}

func getMrsResourceKey(metric model.BatchMetricData) string {
	for _, dim := range *metric.Dimensions {
		if dim.Name == "cluster_id" {
			return dim.Value
		}
	}
	return ""
}

func getEndpoint(server, version string) string {
	if endpoint, ok := endpointConfig[server]; ok {
		return endpoint
	}
	return fmt.Sprintf("https://%s/%s", strings.Replace(host, "iam", server, 1), version)
}

// 标签只允许大写字母，小写字母和下划线，过滤tags中有效的tag
func getTags(tags map[string]string) ([]string, []string) {
	var keys, values []string
	for key, value := range tags {
		valid := tagRegexp.MatchString(key)
		if !valid {
			continue
		}
		keys = append(keys, key)
		values = append(values, value)
	}
	return keys, values
}

type Tag struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

func fmtTags(tagInfo interface{}) map[string]string {
	bytes, err := json.Marshal(tagInfo)
	if err != nil {
		return nil
	}
	var tags []Tag
	err = json.Unmarshal(bytes, &tags)
	if err != nil {
		return nil
	}
	tagMap := make(map[string]string)
	for _, tag := range tags {
		tagMap[tag.Key] = tag.Value
	}
	return tagMap
}

type ResourceBaseInfo struct {
	ID   string
	Name string
	EpId string
	Tags map[string]string
}

func getDimsNameKey(dims []model.MetricsDimension) string {
	dimsNamesList := make([]string, 0, len(dims))
	for _, dim := range dims {
		dimsNamesList = append(dimsNamesList, dim.Name)
	}
	return strings.Join(dimsNamesList, ",")
}

func getDimsValueKey(dims []model.MetricsDimension) string {
	dimsValuesList := make([]string, 0, len(dims))
	for _, dim := range dims {
		dimsValuesList = append(dimsValuesList, dim.Value)
	}
	return strings.Join(dimsValuesList, ",")
}

func buildSingleDimensionMetrics(metricNames []string, namespace, dimName, dimValue string) []model.MetricInfoList {
	filterMetrics := make([]model.MetricInfoList, len(metricNames))
	for index := range metricNames {
		filterMetrics[index] = model.MetricInfoList{
			Namespace:  namespace,
			MetricName: metricNames[index],
			Dimensions: []model.MetricsDimension{
				{
					Name:  dimName,
					Value: dimValue,
				},
			},
		}
	}
	return filterMetrics
}

func buildDimensionMetrics(metricNames []string, namespace string, dimensions []model.MetricsDimension) []model.MetricInfoList {
	filterMetrics := make([]model.MetricInfoList, len(metricNames))
	for index := range metricNames {
		filterMetrics[index] = model.MetricInfoList{
			Namespace:  namespace,
			MetricName: metricNames[index],
			Dimensions: dimensions,
		}
	}
	return filterMetrics
}

func getHcClient(endpoint string) *core.HcHttpClient {
	return core.NewHcHttpClient(impl.NewDefaultHttpClient(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify))).
		WithCredential(authCredentialMap[conf.AuthMode](RegionServiceType)).
		WithEndpoints([]string{endpoint})
}

func genDefaultReqDefWithOffsetAndLimit(path string, response interface{}) *def.HttpRequestDef {
	reqDefBuilder := def.NewHttpRequestDefBuilder().WithMethod(http.MethodGet).WithPath(path).
		WithResponse(response).WithContentType("application/json")

	reqDefBuilder.WithRequestField(def.NewFieldDef().WithName("Offset").WithJsonTag("offset").WithLocationType(def.Query))
	reqDefBuilder.WithRequestField(def.NewFieldDef().WithName("Limit").WithJsonTag("limit").WithLocationType(def.Query))
	return reqDefBuilder.Build()
}

func getDefaultString(value *string) string {
	if value != nil {
		return *value
	}
	return ""
}

func fmtResourceProperties(properties map[string]interface{}, value interface{}) error {
	bytes, err := json.Marshal(properties)
	if err != nil {
		return err
	}

	return json.Unmarshal(bytes, value)
}

func getResourcesBaseInfoFromRMS(provider, resourceType string, optionalRegionID ...string) ([]ResourceBaseInfo, error) {
	resp, err := listResources(provider, resourceType, optionalRegionID...)
	if err != nil {
		logs.Logger.Errorf("Failed to list resource of %s.%s, error: %s", provider, resourceType, err.Error())
		return nil, err
	}
	services := make([]ResourceBaseInfo, len(resp))
	for index, resource := range resp {
		services[index].ID = *resource.Id
		services[index].Name = *resource.Name
		services[index].EpId = *resource.EpId
		services[index].Tags = resource.Tags
	}
	return services, nil
}

func GetResourceInfoExpirationTime() time.Duration {
	intervalMinutes := CloudConf.Global.ResourceSyncIntervalMinutes
	if intervalMinutes <= MinimumResourceInfoSyncInterval {
		return MinimumResourceInfoSyncInterval * time.Minute
	}
	return time.Duration(intervalMinutes) * time.Minute
}

func GetMetricInfoExpirationTime() time.Duration {
	expirationDays := CloudConf.Global.MetricInfoExpirationDays
	return time.Duration(expirationDays*24) * time.Hour
}

// ContainsInArray 判断字符串是否包含在数组中,由于sort.SearchStrings使用二分查找法,需要传入按字母序排序后的数组
func ContainsInArray(sortedArray []string, target string) bool {
	index := sort.SearchStrings(sortedArray, target)
	if index < len(sortedArray) && sortedArray[index] == target {
		return true
	}
	return false
}

func DimNameEquals(originalDimName, targetDimName string) bool {
	if originalDimName == targetDimName {
		return true
	}
	if strings.Contains(originalDimName, ",") && strings.Contains(targetDimName, ",") {
		originalDimNameArray := strings.Split(originalDimName, ",")
		sort.Strings(originalDimNameArray)
		sortedDimName := strings.Join(originalDimNameArray, ",")
		targetDimNameArray := strings.Split(targetDimName, ",")
		sort.Strings(targetDimNameArray)
		sortedTargetDimName := strings.Join(targetDimNameArray, ",")
		return sortedDimName == sortedTargetDimName
	}
	return false
}

func GetIAMClient() *iam.IamClient {
	return iam.NewIamClient(
		iam.IamClientBuilder().
			WithEndpoint(getEndpoint("iam", "v3")).
			WithCredential(authCredentialMap[conf.AuthMode](GlobalServiceType)).
			WithHttpConfig(GetHttpConfig().WithIgnoreSSLVerification(CloudConf.Global.IgnoreSSLVerify)).
			Build())
}

func strSliceContains(ss []string, s string) bool {
	for _, v := range ss {
		if v == s {
			return true
		}
	}
	return false
}

func isErrorTypeForTooManyRequests(err error) bool {
	serviceRespError, ok := err.(*sdkerr.ServiceResponseError)
	if ok {
		return serviceRespError.StatusCode == TooManyRequestsErrorCode
	}
	return false
}

// 判断指标是否在白名单中
func IsMetricInfoInWhiteList(metricInfo model.MetricInfoList) bool {
	configMap := getMetricConfigMap(metricInfo.Namespace)
	// 如果白名单中无对应命名空间，代表全量查询该命名空间下的指标数据
	if configMap == nil {
		return true
	}
	var currentDimNameArray []string
	for _, dimension := range metricInfo.Dimensions {
		currentDimNameArray = append(currentDimNameArray, dimension.Name)
	}
	currentDimNames := strings.Join(currentDimNameArray, ",")
	var metricNames []string
	// 从白名单中获取当前维度所有指标列表
	for dimNames := range configMap {
		if DimNameEquals(currentDimNames, dimNames) {
			metricNames = configMap[dimNames]
			break
		}
	}
	// 判断当前指标是否在指标列表中
	for _, metricName := range metricNames {
		if metricName == metricInfo.MetricName {
			return true
		}
	}
	return false
}

func getEpIdRequestPart() []string {
	if CloudConf.Global.EpIds == "" {
		return []string{"all_granted_eps"}
	}
	epIds := strings.Split(CloudConf.Global.EpIds, ",")
	return epIds
}

func appendNameValuePairToLabelInfo(labelInfo labelInfo, names []string, values []string) labelInfo {
	if len(names) != len(values) {
		return labelInfo
	}
	for index, name := range names {
		if isContainsInStringArr(name, labelInfo.Name) {
			logs.Logger.Infof("Label name already exists: %s", name)
			continue
		}
		labelInfo.Name = append(labelInfo.Name, name)
		labelInfo.Value = append(labelInfo.Value, values[index])
	}
	return labelInfo
}

func getMetricKeyFromMetricInfo(metric model.MetricInfoList) string {
	resourceKey := GetResourceKeyFromMetricInfo(metric)
	return fmt.Sprintf("%s.%s", resourceKey, metric.MetricName)
}

func mergeMetricsWithCache(newMetrics []model.MetricInfoList, oldMetrics []MetricInfoListWithTTL) []MetricInfoListWithTTL {
	mergedMetricMap := map[string]MetricInfoListWithTTL{}
	for _, oldMetric := range oldMetrics {
		metricKey := getMetricKeyFromMetricInfo(oldMetric.MetricInfoList)
		mergedMetricMap[metricKey] = oldMetric
	}

	var newMetricInfoList []MetricInfoListWithTTL
	for _, metric := range newMetrics {
		newMetricKey := getMetricKeyFromMetricInfo(metric)
		metricItem := MetricInfoListWithTTL{
			TTL:            time.Now().Add(GetMetricInfoExpirationTime()).Unix(),
			MetricInfoList: metric,
		}
		newMetricInfoList = append(newMetricInfoList, metricItem)
		mergedMetricMap[newMetricKey] = metricItem
	}

	// 小于需要清理的阈值，直接返回merge后的全量指标列表
	if len(mergedMetricMap) < CloudConf.Global.MetricInfoCleanThreshold {
		var resultMetrics []MetricInfoListWithTTL
		for _, tmpCacheMetric := range mergedMetricMap {
			resultMetrics = append(resultMetrics, tmpCacheMetric)
		}
		return resultMetrics
	}
	// 大于需要清理的阈值，小于该阈值的2倍（超过该二倍值需要使用新指标直接覆盖缓存），执行清理任务，并返回清理后的指标列表
	if len(mergedMetricMap) < 2*CloudConf.Global.MetricInfoCleanThreshold {
		return cleanMergedMetrics(mergedMetricMap)
	}
	// 大于2倍阈值，直接返回新的指标列表
	return newMetricInfoList
}

func cleanMergedMetrics(mergedMetricMap map[string]MetricInfoListWithTTL) []MetricInfoListWithTTL {
	var cleanedMetrics []MetricInfoListWithTTL
	for _, metric := range mergedMetricMap {
		if time.Now().Unix() < metric.TTL {
			cleanedMetrics = append(cleanedMetrics, metric)
		}
	}
	return cleanedMetrics
}

func GetDimensionsValueByName(dimensions []model.MetricsDimension, dimName string) string {
	if dimName == "" || len(dimensions) == 0 {
		return ""
	}
	for _, dimension := range dimensions {
		if dimension.Name == dimName {
			return dimension.Value
		}
	}
	return ""
}

func ContainDimensionName(dimensions []model.MetricsDimension, dimName string) bool {
	if dimName == "" || len(dimensions) == 0 {
		return false
	}
	for _, dimension := range dimensions {
		if dimension.Name == dimName {
			return true
		}
	}
	return false
}
