package carbon

import (
	"fmt"
	"strconv"
	"time"
)

// Metric represents single client metric record.
type Metric struct {
	Name  string
	Value float64
	Time  time.Time
}

// String implements Stringer interface
func (m Metric) String() string {
	return fmt.Sprintf("%s %v %d", m.Name, m.Value, m.Time.Unix())
}

func (m Metric) aggregationKey() string {
	return m.Name + strconv.FormatInt(m.Time.Unix(), 10)
}
