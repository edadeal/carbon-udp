package carbon

// AggregationFunc is a type for custom flush aggregation funcs
type AggregationFunc func(mm []Metric) Metric

// AggregateSum will sum metrics
func AggregateSum(mm []Metric) Metric {
	return Metric{
		Name: mm[0].Name,
		Value: func() (sum float64) {
			for i := range mm {
				sum += mm[i].Value
			}
			return
		}(),
		Time: mm[len(mm)-1].Time,
	}
}

// AggregateAvg will calculate average value of metrics
func AggregateAvg(mm []Metric) Metric {
	return Metric{
		Name: mm[0].Name,
		Value: func() (avg float64) {
			for i := range mm {
				avg += mm[i].Value
			}
			avg /= float64(len(mm))
			return
		}(),
		Time: mm[len(mm)-1].Time,
	}
}

// AggregateMin will store min value of metrics
func AggregateMin(mm []Metric) (m Metric) {
	m = mm[0]
	for i := range mm {
		if mm[i].Value < m.Value {
			m = mm[i]
		}
	}
	return
}

// AggregateMax will store max value of metrics
func AggregateMax(mm []Metric) (m Metric) {
	m = mm[0]
	for i := range mm {
		if mm[i].Value > m.Value {
			m = mm[i]
		}
	}
	return
}

// AggregateFirst will store first value of metrics
func AggregateFirst(mm []Metric) Metric {
	return Metric{
		Name:  mm[0].Name,
		Value: mm[0].Value,
		Time:  mm[0].Time,
	}
}

// AggregateLast will store first value of metrics
func AggregateLast(mm []Metric) Metric {
	return Metric{
		Name:  mm[len(mm)-1].Name,
		Value: mm[len(mm)-1].Value,
		Time:  mm[len(mm)-1].Time,
	}
}
