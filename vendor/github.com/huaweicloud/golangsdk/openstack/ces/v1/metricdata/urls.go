package metricdata

import "github.com/huaweicloud/golangsdk"

// batch query metric data url
func batchQueryMetricDataURL(c *golangsdk.ServiceClient) string {
	return c.ServiceURL("batch-query-metric-data")
}

func addMetricDataURL(c *golangsdk.ServiceClient) string {
	return c.ServiceURL("metric-data")
}

func getEventDataURL(c *golangsdk.ServiceClient) string {
	return c.ServiceURL("event-data")
}

func getURL(c *golangsdk.ServiceClient) string {
	return c.ServiceURL("metric-data")
}
