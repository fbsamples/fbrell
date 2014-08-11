FROM base/archlinux
MAINTAINER Naitik Shah "n@daaku.org"

RUN pacman --sync --refresh --sysupgrade --noconfirm
RUN pacman --sync --noconfirm \
  ca-certificates \
  docker \
  git \
  go \
  iptables \
  mercurial \
  nodejs

ENV GOPATH /gopath
ENV GO_LDFLAGS "-X github.com/daaku/rell/context/viewcontext.version docker-1"
ADD . /gopath/src/github.com/daaku/rell/

# add static resources & examples
ADD public/ /gopath/bin/usr/share/rell/public/
ADD examples/db/mu/ /gopath/bin/usr/share/rell/examples/mu/

# build js
WORKDIR /gopath/src/github.com/daaku/rell/js
RUN npm install
RUN ./node_modules/.bin/browserify -e rell.js --exports require > /gopath/bin/usr/share/rell/browserify.js

# build go
RUN go get -v -d github.com/daaku/rell
RUN go build -ldflags="$GO_LDFLAGS" -o=/gopath/bin/rell github.com/daaku/rell

# build image
ADD Dockerfile.runtime /gopath/bin/Dockerfile
CMD docker build -t daaku/rell /gopath/bin
