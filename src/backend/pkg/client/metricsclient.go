package client

import (
	"fmt"
	"strings"

	resty "github.com/go-resty/resty/v2"
)

func (selector *CumulativeMetricsResourceSelector) getUrl() string {
	if len(selector.Pods) > 0 {
		return fmt.Sprintf("/namespaces/%s/pods/%s", selector.Namespace, strings.Join(selector.Pods, ","))
	} else if selector.Node != "" {
		return fmt.Sprintf("/nodes/%s", selector.Node)
	} else {
		return ""
	}
}

// RestfulClient 리턴
func NewCumulativeMetricsClient(metricsScraperUrl string, context string) *CumulativeMetricsClient {
	return &CumulativeMetricsClient{
		Get: func(selector CumulativeMetricsResourceSelector) ([]CumulativeMetricUnit, error) {
			result := []CumulativeMetricUnit{}

			_, err := resty.New().R().
				SetHeader("Content-Type", "application/json").
				SetResult(&result).
				Get(fmt.Sprintf("%s/api/v1/clusters/%s%s", metricsScraperUrl, context, selector.getUrl()))
			if err != nil {
				return nil, err
			}

			return result, nil
		},
	}
}
