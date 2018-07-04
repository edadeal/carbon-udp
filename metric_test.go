package carbon

import (
	"fmt"
	"strconv"
	"testing"
	"time"
)

func TestMetricString(t *testing.T) {
	now := time.Now()
	m := Metric{"my.service.value", 4.2, now}

	expected := fmt.Sprintf("%s %v %d", m.Name, m.Value, now.Unix())
	if m.String() != expected {
		t.Fatalf("Expected %s, got %s", expected, m.String())
	}
}

func TestMetricAggregationKey(t *testing.T) {
	now := time.Now()
	m := Metric{"my.service.value", 4.2, now}

	expected := m.Name + strconv.FormatInt(now.Unix(), 10)
	if m.aggregationKey() != expected {
		t.Fatalf("Expected %s, got %s", expected, m.aggregationKey())
	}
}
