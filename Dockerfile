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

# add local source
ENV GOPATH /gopath
ADD . /gopath/src/github.com/daaku/rell/

# add static resources & examples
ADD public/ /gopath/bin/usr/share/rell/public/
ADD examples/db/mu/ /gopath/bin/usr/share/rell/examples/mu/

# build js
WORKDIR /gopath/src/github.com/daaku/rell/js
RUN npm install
RUN ./node_modules/.bin/browserify -e rell.js --exports require > /gopath/bin/usr/share/rell/browserify.js

# build go
RUN go get -v github.com/daaku/rell

# build image
ADD Dockerfile.runtime /gopath/bin/Dockerfile
CMD docker build -t daaku/rell /gopath/bin
