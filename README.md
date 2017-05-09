[![Go Report Card](https://goreportcard.com/badge/github.com/appscode/g2)](https://goreportcard.com/report/github.com/appscode/g2)

[Website](https://appscode.com) • [Slack](https://slack.appscode.com) • [Forum](https://discuss.appscode.com) • [Twitter](https://twitter.com/AppsCodeHQ)

G2
==========

G2 is a server, worker and client implementation of [Gearman](http://gearman.org/) in [Go Programming Language](http://golang.org).

The client package is used for sending jobs to the Gearman job server and getting responses from the server.

	"github.com/appscode/g2/client"

The worker package will help developers in developing Gearman worker service easily.

	"github.com/appscode/g2/worker"
	    
The gearadmin package implements a client for the [gearman admin protocol](http://gearman.org/protocol/).

    "github.com/appscode/g2/gearadmin"

[![GoDoc](https://godoc.org/github.com/appscode/g2?status.png)](https://godoc.org/github.com/appscode/g2)

Install
=======

Install the client package:

> $ go get github.com/appscode/g2/client

Install the worker package:

> $ go get github.com/appscode/g2/worker

Both of them:

> $ go get github.com/appscode/g2

Usage
=====
## Server
	how to start gearmand?

	./gearmand --addr="0.0.0.0:4730"

	how to not use leveldb as storage?

	./gearmand --storage-dir= --addr="0.0.0.0:4730"

how to track stats:

	http://localhost:3000/debug/stats

how to list workers by "cando" ?

	http://localhost:3000/worker/function

how to list all workers ?

	http://localhost:3000/worker

how to query job status ?

	http://localhost:3000/job/jobhandle

how to list all jobs ?

	http://localhost:3000/job

how to change monitor address ?

	export GEARMAND_MONITOR_ADDR=:4567

## Worker

```go
import (
	"log"
	"strings"
	"os"
	"time"
	"github.com/appscode/g2/worker"
	"github.com/mikespook/golib/signal"
)

func ToUpper(job worker.Job) ([]byte, error) {
	log.Printf("ToUpper: Data=[%s]\n", job.Data())
	data := []byte(strings.ToUpper(string(job.Data())))
	time.Sleep(10 * time.Second)
	return data, nil
}

func main(){

	log.Println("Worker starting...")
	defer log.Println("Worker shutdown")

	w := worker.New(worker.Unlimited)
	defer w.Close()

	//error handling
	w.ErrorHandler = func(e error) {
		log.Println(e)
	}

	w.AddServer("tcp4","127.0.0.1:4730")

	// Use worker.Unlimited if you want no timeout
	w.AddFunc("ToUpper", ToUpper, worker.Unlimited)
	// This will give a timeout of 5 seconds
	w.AddFunc("ToUpperTimeOut5", ToUpper, 5)

	if err := w.Ready(); err != nil {
		log.Fatal(err)
		return
	}

	go w.Work()

	//Ctrl-C to exit
	signal.Bind(os.Interrupt, func() uint { return signal.BreakExit })
	signal.Wait()
}
```

## Client

```go
import (
	"log"
	"os"
	"github.com/appscode/g2/client"
	"github.com/appscode/g2/pkg/runtime"
	"github.com/mikespook/golib/signal"
)

func main(){

	c, err := client.New("tcp4", "127.0.0.1:4730")
	if err != nil {
		log.Fatalln(err)
	}
	defer c.Close()

	//error handling
	c.ErrorHandler = func(e error) {
		log.Println(e)
	}

	echo := []byte("Hello world")
	echoMsg, err := c.Echo(echo)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println("Echo:", string(echoMsg))

	jobHandler := func(resp *client.Response) {
		log.Printf("Response: %s", resp.Data)
	}

	c.Do("ToUpper", echo, runtime.JobNormal, jobHandler)
	c.Do("ToUpperTimeOut5", echo, runtime.JobNormal, jobHandler)

	//Ctrl-C to exit
	signal.Bind(os.Interrupt, func() uint { return signal.BreakExit })
	signal.Wait()
}
```

## Gearman Admin Client
Package gearadmin provides simple bindings to the gearman admin protocol: http://gearman.org/protocol/. Here's an example program that outputs the status of all worker queues in gearman:

```go
package main

import (
	"fmt"
	"github.com/appscode/g2/gearadmin"
	"net"
)

func main() {
	c, err := net.Dial("tcp", "localhost:4730")
	if err != nil {
		panic(err)
	}
	defer c.Close()
	admin := gearadmin.NewGearmanAdmin(c)
	status, _ := admin.Status()
	fmt.Printf("%#v\n", status)
}
```

Build Instructions
==================
```sh
# dev build
./hack/make.py

# Install/Update dependency (needs glide)
glide slow

# Build Docker image
./hack/docker/setup.sh

# Push Docker image (https://hub.docker.com/r/appscode/gearmand/)
./hack/docker/setup.sh push

# Deploy to Kubernetes (one time setup operation)
kubectl run gearmand --image=appscode/gearmand:<tag> --replica=1

# Deploy new image
kubectl set image deployment/gearmand tc=appscode/gearmand:<tag>
```

Acknowledgement
===============
 * Client and Worker package forked from https://github.com/mikespook/gearman-go
 * Server package forked from https://github.com/ngaut/gearmand
 * Gearadmin client forked from https://github.com/Clever/gearadmin
 * Gearman project (http://gearman.org/protocol/)

License
==================================
Apache 2.0. See [LICENSE](LICENSE).

- Copyright (C) 2016-2017 by AppsCode Inc.
- Copyright (C) 2016 by Clever.com (portions of gearadmin client)
- Copyright (c) 2014 [ngaut](https://github.com/ngaut) (portions of gearmand)
- Copyright (C) 2011 by Xing Xing (portions of client and worker)
