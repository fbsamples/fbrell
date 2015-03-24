rell [![Build Status](https://secure.travis-ci.org/daaku/rell.svg)](https://travis-ci.org/daaku/rell)
====

Facebook Read Eval Log Loop is an interactive environment for exploring the
Facebook Connect JavaScript SDK. The SDK is available
[here](https://developers.facebook.com/docs/reference/javascript/).

[Try it out](https://www.fbrell.com/examples/).

Development Environment
-----------------------

You'll need [Go](https://golang.org/) to work on rell. Once you have it:

```sh
go get github.com/daaku/rell
rell
```

Then go to [http://localhost:43600/](http://localhost:43600/). Look at the help
text from `rell -h` to see what other options are available. You'll need your
own [Facebook Application](https://developers.facebook.com/) and
a [Parse Application](https://parse.com/) to make some of the features
available.

Heroku
------

The application can be run on Heroku:

```sh
heroku create -s cedar
heroku config:add BUILDPACK_URL=http://github.com/kr/heroku-buildpack-inline.git
heroku config:set RELL_EMPCHECK_APP_ID=...
heroku config:set RELL_EMPCHECK_APP_SECRET=...
heroku config:set RELL_FB_APP_ID=...
heroku config:set RELL_FB_APP_NS=...
heroku config:set RELL_FB_APP_SECRET=...
heroku config:set RELL_PARSE_APP_ID=...
heroku config:set RELL_PARSE_REST_API_KEY=...
```
