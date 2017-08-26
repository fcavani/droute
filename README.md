# Droute

Droute is a proxy server and a http router for microservices. The proxy part
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

## Govendor

Start sync yours vendor folder.

```
go get github.com/kardianos/govendor
govendor sync
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
