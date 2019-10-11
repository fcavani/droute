# Warning

**Under development don't use its!**

# Droute

[![Build Status](https://travis-ci.org/fcavani/droute.svg?branch=master)](https://travis-ci.org/fcavani/droute) [![GoDoc](https://godoc.org/github.com/fcavani/droute?status.svg)](https://godoc.org/github.com/fcavani/droute)
[![Go Report Card](https://goreportcard.com/badge/github.com/fcavani/droute)](https://goreportcard.com/report/github.com/fcavani/droute)

Droute is a proxy server and a http router. The proxy part
receive the requests and redirect it based on the domain and path of the
request. The client part, where is the logic, register itself with the proxy
sending to it the domain, path and host. Is supported multiple hosts for the same
domain and path, you can do the load balancing politics of our choice. The proxy supports
retries, load balance and circuit brake. You can add more middlewares if we need
it.

## Install

Execute the above command:

```
go get github.com/fcavani/droute
```

And change the folder to the droute folder.

## Dep

Start sync yours vendor folder.

```
go get -u github.com/golang/dep/cmd/dep
dep ensure
```

## Server

I have made a sample server in main.go, just compile it.

```
go build github.com/fcavani/droute
```

## Client

Client is simple, it's like the httprouter. See the client/client_test.go.

## TODO

Client can only add new routes. Need to implement remove and list routes.
httprouter don't support it, and I don't know how I will do it.
May be I will make a new router, I don't know...
