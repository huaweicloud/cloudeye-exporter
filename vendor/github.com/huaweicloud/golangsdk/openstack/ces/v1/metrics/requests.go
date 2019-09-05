package metrics

import (
	"github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/pagination"
)

// ListOptsBuilder allows extensions to add additional parameters to the
// List request.
type ListOptsBuilder interface {
	ToMetricsListMap() (string, error)
}

//ListOpts allows the filtering and sorting of paginated collections through the API.
type ListOpts struct {
	// Specifies the namespace.
	Namespace string `q:"namespace"`

	// The value ranges from 1 to 1000, and is 1000 by default.
	// This parameter is used to limit the number of query results.
	Limit *int `q:"limit"`

	// Specifies the metric name.
	MetricName string `q:"metric_name"`

	// Specifies the metric dimension.
	// A maximum of three dimensions are supported, and the dimensions are numbered from 0 in dim.
	Dim0 string `q:"dim.0"`
	Dim1 string `q:"dim.1"`
	Dim2 string `q:"dim.2"`

	// Specifies the paging start value.
	Start string `q:"start"`

	// Specifies the sorting order of query results.
	Order string `q:"order"`
}

// ToMetricsListMap formats a ListOpts into a query string.
func (opts ListOpts) ToMetricsListMap() (string, error) {
	s, err := golangsdk.BuildQueryString(opts)
	if err != nil {
		return "", err
	}
	return s.String(), err
}

//Get the Metric List
func List(client *golangsdk.ServiceClient, opts ListOptsBuilder) pagination.Pager {
	url := getMetricsURL(client)
	if opts != nil {
		query, err := opts.ToMetricsListMap()
		if err != nil {
			return pagination.Pager{Err: err}
		}
		url += query
	}
	return pagination.NewPager(client, url, func(r pagination.PageResult) pagination.Page {
		return MetricsPage{pagination.LinkedPageBase{PageResult: r}}
	})
}
