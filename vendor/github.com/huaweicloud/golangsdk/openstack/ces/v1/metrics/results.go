package metrics

import (
	"bytes"
	"strconv"

	"github.com/huaweicloud/golangsdk"
	"github.com/huaweicloud/golangsdk/pagination"
)

type Metrics struct {
	Metrics  []Metric `json:"metrics"`
	MetaData MetaData `json:"meta_data"`
}

type MetaData struct {
	Count  int    `json:"count"`
	Marker string `json:"marker"`
	Total  int    `json:"total"`
}

type Metric struct {
	// Specifies the metric namespace.
	Namespace string `json:"namespace"`

	// Specifies the metric name, such as cpu_util.
	MetricName string `json:"metric_name"`

	// Specifies the metric unit.
	Unit string `json:"unit"`

	//Specifies the list of dimensions.
	Dimensions []Dimension `json:"dimensions"`
}

type Dimension struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type ListResult struct {
	golangsdk.Result
}

//Extract is a function that accepts a result and extracts metrics.
func ExtractMetrics(r pagination.Page) (Metrics, error) {
	var s Metrics
	err := r.(MetricsPage).ExtractInto(&s)
	return s, err
}

//Extract is a function that all accepts a result and extracts metrics.
func ExtractAllPagesMetrics(r pagination.Page) (Metrics, error) {
	var s Metrics
	s.Metrics = make([]Metric, 0)
	err := r.(MetricsPage).ExtractInto(&s)
	if len(s.Metrics) == 0 {
		return s, nil
	}
	s.MetaData.Count = len(s.Metrics)
	s.MetaData.Total = len(s.Metrics)
	var buf bytes.Buffer
	buf.WriteString(s.Metrics[len(s.Metrics)-1].Namespace)
	buf.WriteString(".")
	buf.WriteString(s.Metrics[len(s.Metrics)-1].MetricName)
	for _, dimension := range s.Metrics[len(s.Metrics)-1].Dimensions {
		buf.WriteString(".")
		buf.WriteString(dimension.Name)
		buf.WriteString(":")
		buf.WriteString(dimension.Value)
	}
	s.MetaData.Marker = buf.String()
	return s, err
}

// MetricsPage is the page returned by a pager when traversing over a
// collection of metrics.
type MetricsPage struct {
	pagination.LinkedPageBase
}

// NextPageURL is invoked when a paginated collection of metrics has reached
// the end of a page and the pager seeks to traverse over a new one. In order
// to do this, it needs to construct the next page's URL.
func (r MetricsPage) NextPageURL() (string, error) {
	metrics, err := ExtractMetrics(r)
	if err != nil {
		return "", err
	}

	if len(metrics.Metrics) < 1 {
		return "", nil
	}

	limit := r.URL.Query().Get("limit")
	num, _ := strconv.Atoi(limit)
	if num > len(metrics.Metrics) {
		return "", nil
	}

	metricslen := len(metrics.Metrics) - 1

	var buf bytes.Buffer
	buf.WriteString(metrics.Metrics[metricslen].Namespace)
	buf.WriteString(".")
	buf.WriteString(metrics.Metrics[metricslen].MetricName)
	for _, dimension := range metrics.Metrics[metricslen].Dimensions {
		buf.WriteString(".")
		buf.WriteString(dimension.Name)
		buf.WriteString(":")
		buf.WriteString(dimension.Value)
	}
	return r.WrapNextPageURL(buf.String())
}

// IsEmpty checks whether a NetworkPage struct is empty.
func (r MetricsPage) IsEmpty() (bool, error) {
	s, err := ExtractMetrics(r)
	return len(s.Metrics) == 0, err
}

/*
ExtractNextURL is an internal function useful for packages of collection
resources that are paginated in a certain way.

It attempts to extract the "start" URL from slice of Link structs, or
"" if no such URL is present.
*/
func (r MetricsPage) WrapNextPageURL(start string) (string, error) {
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		return "", nil
	}
	uq := r.URL.Query()
	uq.Set("start", start)
	r.URL.RawQuery = uq.Encode()
	return r.URL.String(), nil
}
