package metrics

import "github.com/huaweicloud/golangsdk"

func getMetricsURL(c *golangsdk.ServiceClient) string {
	return c.ServiceURL("metrics")
}
