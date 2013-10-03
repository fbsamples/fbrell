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
