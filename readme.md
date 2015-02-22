rell
====

Facebook Read Eval Log Loop is an interactive environment for exploring the
Facebook Connect JavaScript SDK. The SDK is available
[here](https://developers.facebook.com/docs/reference/javascript/).

[Try it out](https://www.fbrell.com/examples/).

Development Environment
-----------------------

You'll need [Go](http://golang.org/) to work on rell. Once you have it:

```sh
go get -u github.com/daaku/rell
```

Put your configuration in the default location:

```sh
mkdir -p ~/.config/rell
cat > ~/.config/rell/config <<EOF
fbapp_id=<app-id>
fbapp_namespace=<canvas-namespace>
fbapp_secret=<app-secret>
EOF
```

Then start the server:

```sh
rell
```
