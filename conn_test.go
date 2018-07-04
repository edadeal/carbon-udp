package carbon

import (
	"fmt"
	"net"
	"testing"
	"time"
)

func TestDial(t *testing.T) {
	_, err := Dial("localhost:2003")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
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

	if conn.dialTimeout != timeout {
		t.Fatalf("Expecting timeout to be %v, got %v", timeout, conn.dialTimeout)
	}
}

func TestDialWithFlush(t *testing.T) {
	flushEvery := 2 * time.Second

	conn, err := Dial("localhost:2003", Autoflush(flushEvery, AggregateSum))
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if conn.flushEnabled != true {
		t.Fatalf("Expecting flush to be eanbled")
	}

	testTimer := time.NewTimer(3 * time.Second)
	defer testTimer.Stop()

	for {
		select {
		case <-testTimer.C:
			goto finish
		default:
			conn.Write(Metric{"my.service.value", 4.2, time.Now()})
			time.Sleep(500 * time.Millisecond)
		}
	}

finish:
	return
}

func TestCloseConnWithFlush(t *testing.T) {
	conn, err := Dial("localhost:2003", Autoflush(3*time.Second, AggregateSum))
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if err := conn.Close(); err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	if _, ok := <-conn.flushChan; ok {
		t.Fatalf("Expecting flush channel to be closed error")
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

	err = conn.Write(Metric{"my.service.value", 4.2, time.Now()})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestConnPush(t *testing.T) {
	conn, err := Dial("localhost:2003")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	err = conn.Push(Metric{"my.service.value", 4.2, time.Now()})
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}
}

func TestConnPushNetError(t *testing.T) {
	conn, err := Dial("localhost:2003")
	if err != nil {
		t.Fatalf("Unexpected error: %s", err)
	}

	wait := make(chan bool)
	go func() {
		lis, err := net.ListenPacket("udp", "localhost:45641")
		if err != nil {
			t.Fatalf("Unexpected error: %s", err)
		}
		<-wait
		lis.Close()
	}()

	conn.nc, _ = net.Dial("udp", "localhost:45641")
	close(wait)
	time.Sleep(2 * time.Second)

	err = conn.Push(Metric{"my.service.value", 4.2, time.Now()})
	// returns err == nil, need to investigate
}
