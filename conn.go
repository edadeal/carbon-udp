package carbon

import (
	"bytes"
	"log"
	"net"
	"strings"
	"time"
)

// Conn holds connection to Carbon and
// provides usefull methods.
type Conn struct {
	nc          net.Conn
	dialTimeout time.Duration
	verbose     bool
	prefix      string
}

// WithTimeout controls time to connect to Carbon server.
func WithTimeout(timeout time.Duration) func(*Conn) {
	return func(c *Conn) {
		c.dialTimeout = timeout
	}
}

// Verbose sets verbose output.
func Verbose(c *Conn) { c.verbose = true }

func (c *Conn) log(format string, args ...interface{}) {
	if c.verbose {
		log.Printf(format, args...)
	}
}

// WithPrefix sets global metrics key prefix.
func WithPrefix(prefixes ...string) func(*Conn) {
	return func(c *Conn) {
		c.prefix = makePrefix(prefixes...)
	}
}

func makePrefix(prefixes ...string) (prefix string) {
	for _, p := range prefixes {
		if p == "" {
			continue
		}
		if strings.HasSuffix(p, ".") {
			p = p[:len(p)-1]
		}
		prefix += strings.Replace(p, ".", "_", -1) + "."
	}
	return
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
	c.log("Closing UDP connection")
	return c.nc.Close()
}

// Write writes metrics to Carbon.
func (c *Conn) Write(metrics ...Metric) error {
	buf := bytes.NewBuffer(nil)
	for _, metric := range metrics {
		buf.WriteString(c.prefix + metric.String() + "\n")
	}

	c.log("Writing %d bytes to carbon through UDP", buf.Len())
	_, err := c.nc.Write(buf.Bytes())
	return err
}

// NewAggregation rules
func (c *Conn) NewAggregation(flushInterval time.Duration, fn AggregationFunc, prefixes ...string) chan<- Metric {
	c.log("Creating new aggregation channel with interval %s", flushInterval)
	fc := make(chan Metric)

	prefix := makePrefix(prefixes...)

	go func() {
		metrics := make(map[string][]Metric)
		flushTicker := time.NewTicker(flushInterval)

		defer flushTicker.Stop()

		c.log("Starting flush cicle (every %v)", flushInterval)
		for {
			select {
			case metric := <-fc:
				key := metric.aggregationKey()
				c.log("Metric received through channel: %+v. Key: %+v", metric, key)
				metric.Name = prefix + metric.Name
				metrics[key] = append(metrics[key], metric)

			case <-flushTicker.C:
				c.log("Aggregating and flattening metrics")
				fm := make([]Metric, 0, len(metrics))
				for k := range metrics {
					if len(metrics[k]) > 0 {
						fm = append(fm, fn(metrics[k]))
					}
				}

				c.log("Pushing %d aggregated metrics to carbon: %+v", len(fm), fm)
				if err := c.Write(fm...); err != nil {
					c.log("An error occured while pushing metrics in flush loop: %s", err)
				}

				c.log("Cleaning up metrics batch")
				metrics = make(map[string][]Metric)
			}
		}
	}()

	return fc
}
