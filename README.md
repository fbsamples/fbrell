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
Apple association file is independent of Android's association statement.

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

## iOS Universal Links

The iOS callback is `https://www.fbrell.com/universal-links/callback`, with no
trailing slash. It returns a fixed iOS-specific fallback notice, using the same
credential-independent GET/HEAD behavior and response protections as the
Android callback. Seeing this notice is not proof of app delivery or login
completion. It does not provide a mock authorization flow or a token exchange.

The Apple association is served publicly at
`https://www.fbrell.com/.well-known/apple-app-site-association` with
`Content-Type: application/json` and no redirect. Its exact callback component
is `/universal-links/callback`; the older `/oauth/response/*` component is
preserved. Android's callback and `assetlinks.json` remain independent.

### App configuration checklist

1. In the developer dashboard for **iOS Hackbook**, Facebook app ID
   `421415891237674`, add the exact callback above to **Valid OAuth Redirect
   URIs** in Facebook Login settings. Preserve existing entries. Android
   Hackbook's app ID `479829488705076` is a different Facebook application;
   its redirect registration does not register the iOS callback.
2. In the test app, add `applinks:www.fbrell.com` to Associated Domains and
   use this exact HTTPS URI in the experimental authorization request and
   callback receiver. Keep the normal custom-scheme login as a separate control.
3. Verify the signed test app's `application-identifier` matches the AASA
   value `T84QZS65DQ.com.facebook.internal.sdk.samples.Hackbook.localDevelopment`.
   This is an Apple app-ID prefix plus bundle ID, not the Facebook app ID or
   necessarily the development team ID. Verify the provisioning profile permits
   Associated Domains; change the association if the installed signer changes.
4. After deployment, verify the public AASA and callback responses, then verify
   actual OS association and delivery. A registered redirect and an HTTP 200
   alone do not prove universal-link behavior in an authentication session.

The host must remain `www.fbrell.com`; do not substitute `fbrell.com`, a wildcard,
query string, or a trailing-slash variant. Developer-dashboard settings and
Apple associated-domain entitlements are separate configuration layers. Native
protocol or first-party redirect checks may need independent server-side work;
adding the URI does not demonstrate that those checks accept the experiment.

The same upstream query-logging caveat described for Android applies to iOS.
Do not send live authorization traffic through unreviewed logging infrastructure.

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
