var soda = require('soda')
  , assert = require('assert')
  , settings = require('../settings')
  , app = require('../app')
  , seleniumLauncher = require('selenium-launcher')

var selenium

// TODO: before/after assume starting selenium is slower than starting the
// webserver. this is probably not ideal.
exports['before'] = function(done) {
  app.listen(0)
  seleniumLauncher(function(er, s) {
    if (er) throw er
    selenium = s
    done()
  })
}

exports['after'] = function(done) {
  app.close()
  selenium.on('exit', function() { done() })
  selenium.kill()
}

var makeTest = function(name, test) {
  exports[name] = function(done) {
    var passed = false
    test(createSodaClient(name).chain.session().setTimeout(5000))
      .testComplete()
      .end(done)
  }
}

function createSodaClient(name) {
  if (process.env.SAUCE) {
    return soda.createSauceClient({
      'url': process.env.SAUCE_URL,
      'username': process.env.SAUCE_USER,
      'access-key': process.env.SAUCE_KEY,
      'os': process.env.SAUCE_OS || 'Windows 2003',
      'browser': process.env.SAUCE_BROWSER || 'firefox',
      'browser-version': process.env.SAUCE_BROWSER_VERSION || '3.6',
      'max-duration': 90,
      'name': name,
    })
  } else {
    return soda.createClient({
      url: 'http://localhost:' + app.address().port + '/',
      host: selenium.host,
      port: selenium.port,
    })
  }
}

function waitAssertTextPresent(text) {
  return function(browser) {
    browser
      .waitForTextPresent(text)
      .assertTextPresent(text)
  }
}

function waitAssertElementPresent(selector) {
  return function(browser) {
    browser
      .waitForElementPresent(selector)
      .assertElementPresent(selector)
  }
}

function waitClickElement(selector) {
  return function(browser) {
    browser
      .and(waitAssertElementPresent(selector))
      .click(selector)
  }
}

function fbLogin(opts) {
  opts = opts || {}
  return function(browser) {
    browser
      .and(waitAssertTextPresent('Log in to use your Facebook account with'))
      .type('id=email', opts.email || settings.facebookTestUser.email)
      .type('id=pass', opts.pass || settings.facebookTestUser.pass)
      .click('name=login')
  }
}

function fbAuthorizeApp() {
  return function(browser) {
    browser
      .and(waitClickElement('css=#grant_clicked'))
  }
}

function fbLoginAndAuthorizeApp(opts) {
  opts = opts || {}
  return function(browser) {
    browser
      .waitForPopUp()
      .selectPopUp('Log In | Facebook')
      .and(fbLogin(opts))
      .and(fbAuthorizeApp())
      .deselectPopUp()
  }
}

function runLoggedInExample(opts) {
  return function(browser) {
    browser
      .open(opts.url)
      .waitForPageToLoad(5000)
      .click('css=#rell-login')
      .and(fbLoginAndAuthorizeApp(opts))
      .and(waitAssertTextPresent('connected'))
      .click('css=#rell-run-code')
  }
}

function runExample(opts) {
  return function(browser) {
    browser
      .open(opts.url)
      .waitForPageToLoad(5000)
      .click('css=#rell-run-code')
  }
}

function runInIFrame(selector, inIFrame) {
  return function(browser) {
    inIFrame(
      browser
        .selectFrame(selector)
    ).selectFrame('relative=top')
  }
}

var runInIFrameDialog = runInIFrame.bind(null, 'css=.fb_dialog_iframe iframe')
var runInIFramePlugin = runInIFrame.bind(null, 'css=#jsroot iframe')

makeTest('can load an example', function(browser) {
  return browser
    .and(runExample({ url: '/saved/9aaec52757371abdba348300ce9ac20a'}))
    .and(waitAssertTextPresent('The answer is 42'))
})

makeTest('login and logout', function(browser) {
  return browser
    .and(runExample({ url: '/saved/79f242491065630b3dfe66cde4fa532b'}))
    .and(waitAssertTextPresent('unknown'))
    .click('css=#fb-login')
    .and(fbLoginAndAuthorizeApp())
    .and(waitAssertTextPresent('User has logged in'))
    .and(waitAssertTextPresent('connected'))
    .click('css=#fb-logout')
    .and(waitAssertTextPresent('User has logged out'))
    .and(waitAssertTextPresent('unknown'))
})

makeTest('home and examples page', function(browser) {
  return browser
    .open('/')
    .waitForPageToLoad(5000)
    .click('link=Examples')
    .waitForPageToLoad(5000)
    .and(waitAssertTextPresent('account-info'))
})

makeTest('user info via API', function(browser) {
  return browser
    .and(runLoggedInExample({ url: '/fb.api/user-info?autoRun' }))
    .and(waitAssertTextPresent('first_name'))
})

makeTest('cancel feed iframe dialog', function(browser) {
  return browser
    .and(runLoggedInExample({ url: '/saved/afedcd65e0c7fe1258468b96514d2c48' }))
    .and(runInIFrameDialog(function(browser) {
      return browser
        .and(waitAssertTextPresent('Post to Your Wall'))
        .click('css=#cancel input')
    }))
    .and(waitAssertTextPresent('Did not publish to the feed'))
})

makeTest('post via feed iframe dialog', function(browser) {
  var message = 'Test run at ' + Date.now()
  return browser
    .and(runLoggedInExample({ url: '/saved/afedcd65e0c7fe1258468b96514d2c48' }))
    .and(runInIFrameDialog(function(browser) {
      return browser
        .and(waitAssertTextPresent('Post to Your Wall'))
        .type('id=feedform_user_message', message)
        .click('css=#publish input')
        .waitForPageToLoad(1000)
    }))
    .and(waitAssertTextPresent('Successfully published to the feed'))
    .open('http://www.facebook.com/profile.php')
    .waitForPageToLoad(5000)
    .and(waitAssertTextPresent(message))
})

makeTest('like and unlike with edge events', function(browser) {
  var url = '/saved/dfba30ac7d85862f1da8c9e2c5f20228'
  return browser
    .and(runLoggedInExample({ url: url }))
    .and(runInIFramePlugin(function(browser) {
      return browser
        .waitForPageToLoad(5000)
        .and(waitClickElement('css=.like_button_no_like'))
    }))
    .and(waitAssertTextPresent('You liked http://fbrell.com/'))
    .and(runExample({ url: url }))
    .and(runInIFramePlugin(function(browser) {
      return browser
        .and(waitClickElement('css=.like_button_like .tombstone_cross'))
    }))
    .and(waitAssertTextPresent('You unliked http://fbrell.com/'))
})
