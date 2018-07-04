package carbon

import (
	"bytes"
	"log"
	"net"
	"time"
)

// Conn holds connection to Carbon and
// provides usefull methods.
type Conn struct {
	nc           net.Conn
	dialTimeout  time.Duration
	flushEnabled bool
	flushChan    chan Metric
	verbose      bool
}

// WithTimeout controls time to connect to Carbon server.
func WithTimeout(timeout time.Duration) func(*Conn) {
	return func(c *Conn) {
		c.dialTimeout = timeout
	}
}

// Verbose sets verbose output.
func Verbose(c *Conn) { c.verbose = true }

// Autoflush enables automatic metrics flush every given duration.
// Metrics would be aggregated using given function.
func Autoflush(every time.Duration, agg AggregationFunc) func(*Conn) {
	return func(c *Conn) {
		c.log("Enabling autoflush")
		c.flushEnabled = true
		c.flushChan = make(chan Metric)

		go func() {
			metrics := make(map[string][]Metric)
			flushTicker := time.NewTicker(every)

			c.log("Starting flush cicle (every %v)", every)
			for {
				select {
				case metric := <-c.flushChan:
					key := metric.aggregationKey()
					c.log("Metric received through channel: %+v. Key: %+v", metric, key)
					metrics[key] = append(metrics[key], metric)

				case <-flushTicker.C:
					c.log("Aggregating and flattening metrics")
					fm := make([]Metric, 0, len(metrics))
					for k := range metrics {
						if len(metrics[k]) > 0 {
							fm = append(fm, agg(metrics[k]))
						}
					}

					c.log("Pushing %d aggregated metrics to carbon: %+v", len(fm), fm)
					if err := c.Push(fm...); err != nil {
						c.log("An error occured while pushing metrics in flush loop: %s", err)
					}

					c.log("Cleaning up metrics batch")
					metrics = make(map[string][]Metric)
				}
			}
		}()
	}
}

// Dial returns new Carbon connection.
func Dial(address string, opts ...func(*Conn)) (*Conn, error) {
	var conn Conn
	for _, opt := range opts {
		opt(&conn)
	}

	c, err := net.DialTimeout("udp", address, conn.dialTimeout)
	if err != nil {
		return nil, err
	}
	conn.nc = c

	return &conn, nil
}

// Close implements Closer interface and
// closes underlying Carbon connection.
func (c *Conn) Close() error {
	if c.flushEnabled {
		c.log("Closing underlying flush channel")
		close(c.flushChan)
		c.log("Flush channel closed")
	}

	c.log("Closing UDP connection")
	return c.nc.Close()
}

// Write is a multipurpose method.
// If FlushInterval option has been used it accumulates all metrics for later flush.
// By default it writes metrics right away.
func (c *Conn) Write(metrics ...Metric) error {
	if c.flushEnabled {
		for _, metric := range metrics {
			c.log("Sending metric to flush channel: %+v", metric)
			c.flushChan <- metric
		}
		return nil
	}
	return c.Push(metrics...)
}

// Push writes metrics to Carbon right away even if FlushInterval option has been used.
func (c *Conn) Push(metrics ...Metric) error {
	buf := bytes.NewBuffer(nil)
	for _, metric := range metrics {
		buf.WriteString(metric.String() + "\n")
	}

	c.log("Writing %d bytes to carbon through UDP", buf.Len())
	_, err := c.nc.Write(buf.Bytes())
	return err
}

func (c *Conn) log(format string, args ...interface{}) {
	if c.verbose {
		log.Printf(format, args...)
	}
}
