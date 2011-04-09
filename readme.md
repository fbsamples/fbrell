fbrell
======

Facebook Read Eval Log Loop is an interactive environment for exploring the
Facebook Connect JavaScript SDK. The SDK is available
[here](http://github.com/facebook/connect-js).

[Try it out](http://fbrell.com/xfbml/fb:login-button).

Getting Started
---------------

Make sure you have [Node](http://nodejs.org/) (v0.4.x) and
[npm](https://github.com/isaacs/npm) installed, and then:

    git clone git://github.com/nshah/rell.git
    cd rell
    git submodule update --init
    cp settings.js.sample settings.js
    vim settings.js # put in your settings
    npm install
    node server.js

Then go to:

    http://localhost:3000/xfbml/fb:login-button
