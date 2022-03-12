# rater

A rate limiter implementation based on Bucket in Go.

[![Go Report Card](https://goreportcard.com/badge/github.com/go-the-way/rater)](https://goreportcard.com/report/github.com/go-the-way/rater)
[![GoDoc](https://pkg.go.dev/badge/github.com/go-the-way/rater?status.svg)](https://pkg.go.dev/github.com/go-the-way/rater?tab=doc)

## quickstart
```go
package main

import (
	"fmt"
	"time"
	
	r "github.com/go-the-way/rater"
)

func main() {
	l := r.NewLimiter(r.CacheBucket(1, 3, 1, time.Second, r.DefaultGenerator(), nil))
	fmt.Println(l.Try()) // outputs: {} true
	fmt.Println(l.Try()) // outputs: <nil> false
}
```