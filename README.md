carbon-udp
=============

[![GoDoc](https://godoc.org/github.com/edadeal/carbon-udp?status.svg)](https://godoc.org/github.com/edadeal/carbon-udp)
[![Go Report Card](https://goreportcard.com/badge/github.com/edadeal/carbon-udp)](https://goreportcard.com/report/github.com/edadeal/carbon-udp)

# What is carbon-udp?

A pure Go library to make it easier to send application metrics to a Carbon server.
It uses efficient plaintext Carbon UDP protocol and provides some useful methods and mechanics.

# Installation

```go get -u github.com/edadeal/carbon-udp```

# Usage

Manually:

```golang
import (
    "time"

    "git.edadev.ru/go/carbon-udp"
)

package main

func main() {
    conn, err := carbon.Dial("localhost:2003", carbon.WithTimeout(3 * time.Second))
    if err != nil {
        panic(err)
    }

    // write given metric immediately
    conn.Write(carbon.Metric{"balance.amount", 4.2, time.Now()}, carbon.Metric{"users.count", 148, time.Now()})
    // you can also use Push method if autoflush has not been enabled for connection
}
```

With autoflush:

```golang
import (
    "time"

    "git.edadev.ru/go/carbon-udp"
)

package main

func main() {
    conn, err := carbon.Dial("localhost:2003",
        carbon.WithTimeout(3 * time.Second),
        carbon.Autoflush(5 * time.Second, carbon.AggregateSum))
    if err != nil {
        panic(err)
    }

    // client will write metrics batch every 5 seconds
    conn.Write(carbon.Metric{"posts.count", 100, time.Now()})
    conn.Write(carbon.Metric{"posts.count", 250, time.Now()})
    conn.Write(carbon.Metric{"posts.count", 2, time.Now()})
    // carbon will receive summarized value as we using AggregateSum function
    // e.g.: Metric{"posts.count", 352, 1530706327}

    // force push metric to server even if auto flush enabled
    conn.Push(carbon.Metric{"service.some.other_metric", 22, time.Now()})
}
```