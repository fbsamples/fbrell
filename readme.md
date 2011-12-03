fbrell
======

Facebook Read Eval Log Loop is an interactive environment for exploring the
Facebook Connect JavaScript SDK. The SDK is available
[here](http://github.com/facebook/connect-js).

[Try it out](http://www.fbrell.com/xfbml/fb:login-button).

Getting Started
---------------

Make sure you have [Node](http://nodejs.org/) (v0.6.x) and
[npm](https://github.com/isaacs/npm) (v1.1.x) installed, and then:

    git clone git://github.com/nshah/rell.git
    cd rell
    cp settings.js.sample settings.js
    vim settings.js # put in your settings
    npm install
    ./app.js

Then go to:

    http://localhost:43600/xfbml/fb:login-button


Running Selenium Tests
----------------------

Download the [Selenium Server](http://seleniumhq.org/download/) and run it,
maybe using something like:

    java -jar selenium-server-standalone-2.0.0.jar

Then run the tests using [expresso](http://visionmedia.github.com/expresso/)
which should have been installed by the earlier `npm install` command:

    ./node_modules/.bin/expresso tests.js

If you signup for [Sauce Labs](https://saucelabs.com/) and enter your
credentials into `settings.js`, then you can also run the tests in the cloud
for free without needing to have Selenium running locally:

    SAUCE=1 ./node_modules/.bin/expresso tests.js
