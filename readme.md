rell
====

Facebook Read Eval Log Loop is an interactive environment for exploring the
Facebook Connect JavaScript SDK. The SDK is available
[here](https://developers.facebook.com/docs/reference/javascript/).

[Try it out](https://www.fbrell.com/examples/).

Development Environment
-----------------------

You'll need these to make modifications to rell:

- [Node](http://nodejs.org/) tested with version 0.8.x.
- [Go](http://golang.org/) tested with version 1.1.x.
- [Git](http://gitscm.com/) tested with version 1.7.x.

Install the main command which will automatically pull Go
dependencies, then use npm to fetch JavaScript dependencies.

```sh
go get -u github.com/daaku/rell
cd $(go list -f '{{.Dir}}' github.com/daaku/rell)/js
npm install
rell -h
```

Docker
------

Deployment is done via [Docker](https://www.docker.com/). Included is a
`Dockerfile` that builds the runtime container image. To run it:

```sh
docker build --tag make-rell . &&
docker run \
  --rm \
  --tty \
  --volume /var/run/docker.sock:/var/run/docker.sock \
  make-rell
```

Running with Docker
-------------------

First run the
[redis container](https://github.com/daaku/dockerfiles/tree/master/redis):

```sh
docker run --name=rell-redis daaku/redis
```

Put your configuration in a file, for example:

```sh
cat > config <<EOF
FBAPP_ID=<app-id>
FBAPP_NAMESPACE=<canvas-namespace>
FBAPP_SECRET=<app-secret>
RELL_BROWSERIFY_OVERRIDE=/usr/share/rell/browserify.js
RELL_EXAMPLES_NEW=/usr/share/rell/examples/mu
RELL_STATIC_DISK_PATH=/usr/share/rell/public
EOF
```

Then start the `rell` container:

```sh
docker run --link rell-redis:redis --env-file=config daaku/rell
```
