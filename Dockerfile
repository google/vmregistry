FROM alpine:latest

RUN set -ex; \
  apk add --no-cache --no-progress --virtual .build-deps git libvirt-dev gcc musl-dev bash go; \
  env GOPATH=/go go get -v github.com/google/vmregistry; \
  env GOPATH=/go go get -v github.com/google/vmregistry/cmd/vmregistry-cli; \
  install -t /bin /go/bin/vmregistry /go/bin/vmregistry-cli; \
  rm -rf /go; \
  apk --no-progress del .build-deps; \
  apk add --no-cache --no-progress libvirt-client

EXPOSE 8000
CMD ["/bin/vmregistry", "-listen=0.0.0.0:8000", "-logtostderr"]
