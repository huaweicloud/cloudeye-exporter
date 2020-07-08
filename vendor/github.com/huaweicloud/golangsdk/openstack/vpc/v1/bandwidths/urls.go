package bandwidths

import (
	"github.com/huaweicloud/golangsdk"
)

func GetURL(c *golangsdk.ServiceClient, bandwidthId string) string {
	return c.ServiceURL("bandwidths", bandwidthId)
}

func ListURL(c *golangsdk.ServiceClient) string {
	return c.ServiceURL("bandwidths")
}

func UpdateURL(c *golangsdk.ServiceClient, bandwidthId string) string {
	return c.ServiceURL("bandwidths", bandwidthId)
}
