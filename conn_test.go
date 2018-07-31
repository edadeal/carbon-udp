package carbon

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestDial(t *testing.T) {
	conn, err := Dial("localhost:2003")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	conn.Close()
}

func TestDialConnectionError(t *testing.T) {
	_, err := Dial("localhost:100500")
	if err == nil {
		t.Fatalf("Expecting error, got nil")
	}
}

func TestDialWithTimeout(t *testing.T) {
	timeout := 2 * time.Second

	conn, err := Dial("localhost:2003", WithTimeout(timeout))
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
	conn.Close()

	if conn.dialTimeout != timeout {
		t.Fatalf("Expecting timeout to be %v, got %v", timeout, conn.dialTimeout)
	}
}

func TestDialWithPrefix(t *testing.T) {
	cases := []struct {
		given, expected string
	}{
		{"", ""},
		{"test", "test."},
		{"medium.", "medium."},
	}

	for _, c := range cases {
		conn, err := Dial("localhost:2003", WithPrefix(c.given))
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		conn.Close()

		if conn.prefix != c.expected {
			t.Fatalf("Expecting prefix to be %v, got %v", c.expected, conn.prefix)
		}
	}
}

func TestMakePrefix(t *testing.T) {
	cases := []struct {
		given    []string
		expected string
	}{
		{nil, ""},
		{
			[]string{"test", "nest"},
			"test.nest.",
		},
		{
			[]string{"test", "nest", "que.st", "de...st"},
			"test.nest.que_st.de___st.",
		},
	}

	for _, c := range cases {
		prefix := makePrefix(c.given...)
		if prefix != c.expected {
			t.Fatalf("Expecting prefix to be %v, got %v", c.expected, prefix)
		}
	}
}

func TestDialTimeoutOpt(t *testing.T) {
	timeout := 1 * time.Second
	opt := WithTimeout(timeout)

	if fmt.Sprintf("%T", opt) != "func(*carbon.Conn)" {
		t.Fatalf("Expecting opt to be of type `func(*carbon.Conn)`, got %T", opt)
	}

	var conn Conn
	opt(&conn)

	if conn.dialTimeout != timeout {
		t.Fatalf("Expecting timeout to be %v, got %v", timeout, conn.dialTimeout)
	}
}

func TestConnWrite(t *testing.T) {
	conn, err := Dial("localhost:2003")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	defer conn.Close()

	err = conn.Write(Metric{"my.service.value", 4.2, time.Now()})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestNewAggregation(t *testing.T) {
	now := time.Now()

	go func() {
		conn, err := Dial("localhost:2003")
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		defer conn.Close()

		ch := conn.NewAggregation(2*time.Second, AggregateFirst)
		ch <- Metric{"test.metric", 1, now}
	}()

	pc, err := net.ListenPacket("udp", ":2003")
	if err != nil {
		t.Fatal(err)
	}

	defer pc.Close()

	buf := make([]byte, 24)
	_, _, err = pc.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	expected := fmt.Sprintf("test.metric 1 %d", now.Unix())
	if string(buf) != expected {
		t.Fatalf("Expecting metric to be '%v', got '%v'", expected, string(buf))
	}
}

func TestNewAggregationWithPrefix(t *testing.T) {
	now := time.Now()

	go func() {
		conn, err := Dial("localhost:2003")
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		defer conn.Close()

		ch := conn.NewAggregation(2*time.Second, AggregateFirst, "medium")
		ch <- Metric{"test.metric", 1, now}
	}()

	pc, err := net.ListenPacket("udp", ":2003")
	if err != nil {
		t.Fatal(err)
	}

	defer pc.Close()

	expected := fmt.Sprintf("medium.test.metric 1 %d", now.Unix())

	buf := make([]byte, len(expected))
	_, _, err = pc.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	if string(buf) != expected {
		t.Fatalf("Expecting metric to be '%v', got '%v'", expected, string(buf))
	}
}

func TestNewAggregationWithChainedPrefix(t *testing.T) {
	now := time.Now()
	host := "my-hostname-local"

	go func() {
		conn, err := Dial("localhost:2003", WithPrefix(host))
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		defer conn.Close()

		ch := conn.NewAggregation(2*time.Second, AggregateFirst, "medium", "tmp")
		ch <- Metric{"test.metric", 1, now}
	}()

	pc, err := net.ListenPacket("udp", ":2003")
	if err != nil {
		t.Fatal(err)
	}

	defer pc.Close()

	expected := fmt.Sprintf("my-hostname-local.medium.tmp.test.metric 1 %d", now.Unix())

	buf := make([]byte, len(expected))
	_, _, err = pc.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	if string(buf) != expected {
		t.Fatalf("Expecting metric to be '%v', got '%v'", expected, string(buf))
	}
}

func TestAggregateAndFlush(t *testing.T) {
	now := time.Now()

	go func() {
		conn, err := Dial("localhost:2003")
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		defer conn.Close()

		metrics := map[string][]Metric{
			"test.metric": []Metric{
				Metric{"test.metric", 1, now},
				Metric{"test.metric", 2, now},
			},
		}
		aggregateAndFlush(conn, metrics, AggregateSum)
	}()

	pc, err := net.ListenPacket("udp", ":2003")
	if err != nil {
		t.Fatal(err)
	}

	defer pc.Close()

	buf := make([]byte, 24)
	_, _, err = pc.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	expected := fmt.Sprintf("test.metric 3 %d", now.Unix())
	if string(buf) != expected {
		t.Fatalf("Expecting metric to be '%v', got '%v'", expected, string(buf))
	}
}

func TestFlushOnChannelClose(t *testing.T) {
	now := time.Now()

	go func() {
		conn, err := Dial("localhost:2003")
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		defer conn.Close()

		ch := conn.NewAggregation(1*time.Minute, AggregateSum)
		ch <- Metric{"test.metric", 1, now}
		ch <- Metric{"test.metric", 2, now}

		// close before autoflush occured
		time.Sleep(2 * time.Second)
		close(ch)
		time.Sleep(2 * time.Second)
	}()

	pc, err := net.ListenPacket("udp", ":2003")
	if err != nil {
		t.Fatal(err)
	}

	defer pc.Close()

	buf := make([]byte, 24)
	_, _, err = pc.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	expected := fmt.Sprintf("test.metric 3 %d", now.Unix())
	if string(buf) != expected {
		t.Fatalf("Expecting metric to be '%v', got '%v'", expected, string(buf))
	}
}

func TestFlushOnConnClose(t *testing.T) {
	now := time.Now()

	go func() {
		conn, err := Dial("localhost:2003")
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}

		defer conn.Close()

		ch := conn.NewAggregation(1*time.Minute, AggregateSum)
		ch <- Metric{"test.metric", 4, now}
		ch <- Metric{"test.metric", 5, now}

		// close before autoflush occured
		time.Sleep(2 * time.Second)
	}()

	pc, err := net.ListenPacket("udp", ":2003")
	if err != nil {
		t.Fatal(err)
	}

	defer pc.Close()

	buf := make([]byte, 24)
	_, _, err = pc.ReadFrom(buf)
	if err != nil {
		t.Fatal(err)
	}

	expected := fmt.Sprintf("test.metric 9 %d", now.Unix())
	if string(buf) != expected {
		t.Fatalf("Expecting metric to be '%v', got '%v'", expected, string(buf))
	}
}
