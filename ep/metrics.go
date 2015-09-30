package ep

import (
	"fmt"
	"time"

	"github.com/codahale/metrics"
)

var deliveryTimerHists = make(map[string]*metrics.Histogram)

const MetricRoot = "chinchilla"

func metricName(name string) string {
	return fmt.Sprintf("%s.%s", MetricRoot, name)
}

func epMetricName(epName string, name string) string {
	return fmt.Sprintf("%s.ep.%s.%s", MetricRoot, epName, name)
}

func RecordDeliveryTime(name string, t time.Duration) {
	if _, ok := deliveryTimerHists[name]; !ok {
		h := metrics.NewHistogram(epMetricName(name, "processing-time"), 0, 300*1000, 4)
		deliveryTimerHists[name] = h
	}

	h := deliveryTimerHists[name]

	h.RecordValue(int64(t.Nanoseconds() / 1000000))
}
