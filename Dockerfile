FROM daaku/arch
RUN pacman --noconfirm --sync go
ENV GOPATH /go
COPY . /go/src/github.com/daaku/rell
RUN go install github.com/daaku/rell
COPY Dockerfile.run /go/bin/Dockerfile
COPY public /go/bin/public
COPY examples/db /go/bin/examples
WORKDIR /go/bin
CMD tar -cf - .
