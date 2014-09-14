FROM daaku/arch
MAINTAINER Naitik Shah "n@daaku.org"

# build system
RUN pacman --sync --noconfirm \
  ca-certificates \
  docker \
  git \
  go \
  iptables \
  mercurial \
  nodejs

# browserify
RUN npm install -g browserify@1.17.x uglify-js@1.3.x

# add local source
ENV GOPATH /gopath:/gopath/src/github.com/daaku/rell/Godeps/_workspace
ADD . /gopath/src/github.com/daaku/rell/

# add static resources & examples
ADD public/ /gopath/bin/usr/share/rell/public/
ADD examples/db/mu/ /gopath/bin/usr/share/rell/examples/mu/

# build js
WORKDIR /gopath/src/github.com/daaku/rell/js
RUN browserify -e rell.js --exports require > /gopath/bin/usr/share/rell/browserify.js

# build go
RUN go install github.com/daaku/rell

# build image
ADD Dockerfile.runtime /gopath/bin/Dockerfile
CMD docker build -t daaku/rell /gopath/bin
