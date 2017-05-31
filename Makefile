api/vmregistry.pb.go: proto/vmregistry.proto
	mkdir -p api
	cd proto && protoc -I/usr/local/include -I. \
	 -I${GOPATH}/src \
	 -I${GOPATH}/src/github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis \
	 --go_out=Mgoogle/api/annotations.proto=github.com/grpc-ecosystem/grpc-gateway/third_party/googleapis/google/api,plugins=grpc:../api \
	 vmregistry.proto

clean:
	rm -rf api/*.pb.go *.pem

testkeys:
	openssl req -x509 \
	  -newkey rsa:2048 \
		-keyout key.pem \
		-out cert.pem \
		-days 365 -nodes -subj "/CN=localhost"
	openssl req -x509 \
	  -newkey rsa:2048 \
		-keyout clientcakey.pem \
		-out clientcacert.pem \
		-days 365 -nodes -subj "/CN=vmregistry ca"
	openssl genrsa -out clientkey.pem 2048
	openssl req -new -sha256 -key clientkey.pem -out client.csr -subj "/CN=vmregistry client"
	openssl x509 -req -days 365 -in client.csr -CA clientcacert.pem -CAkey clientcakey.pem -set_serial 01 -out clientcert.pem
	rm -f client.csr

run:
	go run main.go \
	  -listen=127.0.0.1:8000 \
		-logtostderr \
		-libvirt-uri=$(LIBVIRT_CONN) \
		-credstore-address=127.0.0.1:8008 \
		-server-cert=cert.pem \
		-server-key=key.pem
		

.PHONY: clean testkeys run
