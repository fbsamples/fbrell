rell
====

An interactive environment for exploring the Facebook Connect JavaScript SDK.
The SDK is also available on [GitHub](http://github.com/facebook/connect-js).

[Try it out](http://fbrell.com/xfbml/fb:login-button).

Getting Started
---------------

**This is broken. In the process of moving to npm to handle dependencies.**

Make sure you have [nodejs][nodejs] installed, and then:

    git clone git://github.com/nshah/rell.git
    cd rell
    git submodule update --init
    sin_port=8080 node rell.js

Then go to:

    http://localhost:8080/xfbml/fb:login-button?apikey=bf73baf19273fc2d40b7466309a3d592

Note, the API key in code (file `rell.js`) is tied to the `fbrell.com` domain.
The above specifies another one of **my** applications which has the URL set to
`http://localhost:8080/`.

[nodejs]: http://nodejs.org/
