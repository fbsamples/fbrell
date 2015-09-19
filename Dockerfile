FROM golang
ENV GOPATH /go
COPY . /go/src/github.com/daaku/rell
ENV GO15VENDOREXPERIMENT=1
RUN go install github.com/daaku/rell
COPY Dockerfile.run /go/bin/Dockerfile
COPY public /go/bin/public
COPY examples/db /go/bin/examples
WORKDIR /go/bin
CMD tar -cf - .
