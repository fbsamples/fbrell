rell
====

An interactive environment for exploring the Facebook Connect JavaScript SDK.
The SDK is also available on [GitHub](http://github.com/facebook/connect-js).

[Try it out](http://fbrell.com/xfbml/fb:login-button).

Getting Started
---------------

    git clone git://github.com/nshah/rell.git
    cd rell
    git submodule init
    git submodule update
    gem install sinatra haml json shotgun
    shotgun -P 8080

Then go to:

    http://localhost:8080/xfbml/fb:login-button?apikey=bf73baf19273fc2d40b7466309a3d592

Note, the API key in code (file `rell.rb`) is tied to the `fbrell.com` domain.
The above specifies another one of **my** applications which has the URL set to
`http://localhost:8080/`.
