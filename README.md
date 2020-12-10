# HTTP-Proxy
Super simple HTTPProxy that handles GET, POST, and CONNECT requests.

This implementation is coded with a resume-on-fail attitude, to avoid any outages in about 100 lines of code!

Basic DNS caching is implemented to help speed up connections.

Statically compiled to make the docker container light weight.

# Usage
To run the proxy, listening on the default port :8080 use
```
http-proxy
```

Else if you want to specify a port use
```
http-proxy --listen :2000
```

Or listen on a specific port and host:
```
http-proxy --listen 1.2.3.4:2000
```

The corresponding pre-built container can be pulled here:
```
docker pull pschou/http-proxy:0.1
```
