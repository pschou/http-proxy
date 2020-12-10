ARG ARCH="amd64"
ARG OS="linux"
FROM scratch
LABEL description="Very simple reliable HTTP Proxy, built in golang" owner="dockerfile@paulschou.com"

EXPOSE      8080
ADD ./LICENSE /LICENSE
ADD ./http-proxy "/http-proxy"
ENTRYPOINT  [ "/http-proxy" ]
