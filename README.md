# Droute

Droute is a proxy server and a http router for microservices. The proxy part
receive the requests and redirect it based on the domain and path of the
request. The client part, where is the logic, register itself with the proxy
sending to it the domain, path and host. Is supported multiple hosts for the same
domain and path, you can do the load balancing politics of our choice. The proxy supports
retries, load balance and circuit brake. You can add more middleware if we need
it.

## Govendor

Before start sync our vendor folder.

```
go get github.com/kardianos/govendor
govendor sync
```

## Server

I have made a sample server in main.go, just compile it.

```
go build main.go
```

## Client

Client is simples, it's like the httprouter. See the client/client_test.go.

## TODO

Client can only add new routes. Need to implement remove and list routes.
httprouter don't support it, and I don't know how I will do it for this router.
May be I will make a new router, I don't know...
