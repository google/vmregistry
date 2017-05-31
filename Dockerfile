FROM alpine:latest

RUN set -ex; \
  apk add --no-cache --no-progress --virtual .build-deps git libvirt-dev gcc musl-dev openssh bash go; \
  env GOPATH=/go go get -v github.com/google/credstore; \
  install -t /bin /go/bin/credstore; \
  rm -rf /go; \
  apk --no-progress del .build-deps; \
  apk add --no-cache --no-progress libvirt-client

EXPOSE 8000
CMD ["/bin/vmregistry", "-listen=0.0.0.0:8000", "-logtostderr"]
