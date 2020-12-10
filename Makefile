PROG_NAME := "http-proxy"
IMAGE_NAME := "pschou/http-proxy"
VERSION := "0.1"


build:
	CGO_ENABLED=0 go build -o http-proxy main.go

docker:
	docker build -f Dockerfile --tag ${IMAGE_NAME}:${VERSION} .
	docker push ${IMAGE_NAME}:${VERSION}; \
	docker save -o pschou_${PROG_NAME}.tar ${IMAGE_NAME}:${VERSION}

