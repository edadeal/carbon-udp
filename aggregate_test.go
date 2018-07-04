package carbon

import (
	"reflect"
	"testing"
	"time"
)

func TestAggregation(t *testing.T) {
	first := time.Now()
	last := time.Now()

	cases := []struct {
		metrics  []Metric
		fn       AggregationFunc
		expected Metric
	}{
		// 0
		{
			[]Metric{
				{"test.agg", 10, first},
				{"test.agg", 10, last},
			},
			AggregateSum,
			Metric{"test.agg", 20, last},
		},

		// 1
		{
			[]Metric{
				{"test.agg", 20, first},
				{"test.agg", 10, last},
			},
			AggregateAvg,
			Metric{"test.agg", 15, last},
		},

		// 2
		{
			[]Metric{
				{"test.agg", 1, first},
				{"test.agg", 100, last},
			},
			AggregateMin,
			Metric{"test.agg", 1, first},
		},

		// 3
		{
			[]Metric{
				{"test.agg", 1, first},
				{"test.agg", 100, last},
			},
			AggregateMax,
			Metric{"test.agg", 100, last},
		},

		// 4
		{
			[]Metric{
				{"test.agg", 100, first},
				{"test.agg", 99, last},
			},
			AggregateFirst,
			Metric{"test.agg", 100, first},
		},

		// 5
		{
			[]Metric{
				{"test.agg", 99, first},
				{"test.agg", 100, last},
			},
			AggregateLast,
			Metric{"test.agg", 100, last},
		},
	}

	for i := range cases {
		r := cases[i].fn(cases[i].metrics)
		if !reflect.DeepEqual(r, cases[i].expected) {
			t.Fatalf("[case %d] Expecting %+v, got %+v", i, cases[i].expected, r)
		}
	}
}
