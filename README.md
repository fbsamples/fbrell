# fbrell

[![Build Status](https://secure.travis-ci.org/fbsamples/fbrell.svg)](https://travis-ci.org/fbsamples/fbrell)

Facebook Read Eval Log Loop is an interactive environment for exploring the
Facebook JavaScript SDK. The SDK is available
[here](https://developers.facebook.com/docs/reference/javascript/).

[Try it out](https://www.fbrell.com/examples/).

## Development Environment

You'll need [Go](https://golang.org/) to work on rell. Once you have it:

```bash
git clone git@github.com:fbsamples/fbrell.git
cd fbrell
go get ./... # this will install your dependencies
go build # This will create a fbrell executable file
./fbrell
```

Then, navigate to http://localhost:43600/

Huzzah!

Then go to [http://localhost:43600/](http://localhost:43600/). Look at the help
text from `rell -h` to see what other options are available. You'll need your
own [Facebook Application](https://developers.facebook.com/).

## Android App Links

Rell serves Android Digital Asset Links at `/.well-known/assetlinks.json`.
The file is `public/.well-known/assetlinks.json`; it contains the association
for the Hackbook debug APK used in this experiment. Serving this file alone
does not establish a verified App Link.

When changing APK signers, update the association using the following shape and
the signing certificate SHA-256 fingerprint from the APK that will actually be
installed:

```json
[
  {
    "relation": ["delegate_permission/common.handle_all_urls"],
    "target": {
      "namespace": "android_app",
      "package_name": "com.facebook.hackbook.internal.current",
      "sha256_cert_fingerprints": ["REPLACE_WITH_INSTALLED_APK_SIGNER_SHA256"]
    }
  }
]
```

Do not deploy the placeholder fingerprint. Use the colon-separated SHA-256
certificate fingerprint reported by `apksigner verify --print-certs <apk>`
(or format its SHA-256 hex digest into colon-separated byte pairs). Do not use
an APK file checksum, a Facebook key hash, or a certificate from a different
build. Recheck the package name if the test app changes its application ID.

The Hackbook App Link intent filter must use HTTPS, the exact deployed hostname
(for example, `www.fbrell.com`), and the intended test callback path. Callback
path restrictions belong in the Android intent filter, not this basic
association statement. `fbrell.com` and `www.fbrell.com` are different hosts;
one must not redirect the association-file request to the other.

After deployment, confirm an unauthenticated GET of the association URL returns
HTTP 200, `Content-Type: application/json`, the completed statement, and no
redirect. Then verify domain association and link delivery on the device;
a successful HTTP request does not itself prove Android verification. The
existing Apple association file is independent and should remain unchanged.

### HTTPS callback fallback

`/app-links/callback` returns a fixed, public text notice if the browser reaches
the server instead of handing the link to Android. It does not process or echo
OAuth parameters, use the application error logger, or introduce a redirect.
GET and HEAD are supported; other methods return 405. The response is not proof
of app delivery or login completion.

The callback response uses `Cache-Control: no-store`, `Referrer-Policy:
no-referrer`, and a restrictive content security policy. Browser URL fragments
are not sent to the server. Upstream router/proxy request logging is separate
and may still include query parameters; review that before sending live
authorization traffic. The callback handler only avoids the application's
credential-bearing error-log path.

## Heroku

The application can be run on Heroku:

```sh
heroku create -s cedar
heroku config:add BUILDPACK_URL=http://github.com/kr/heroku-buildpack-inline.git
heroku config:set RELL_EMPCHECK_APP_ID=...
heroku config:set RELL_EMPCHECK_APP_SECRET=...
heroku config:set RELL_FB_APP_ID=...
heroku config:set RELL_FB_APP_NS=...
heroku config:set RELL_FB_APP_SECRET=...
```
