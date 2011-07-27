var soda = require('soda')
  , assert = require('assert')
  , settings = require('./settings')

exports.inter = function() {
  var client = createSodaClient('interactive')
    , wrapMethods = soda.commands.slice(0)
  wrapMethods.push('session')
  wrapMethods.forEach(function(name) {
    var old = client[name]
    client[name] = function() {
      var args = Array.prototype.slice.call(arguments)
        , index = name === 'session' ? 0 : 2
      args[index] = function(er) {
        if (er) {
          console.error(name, er)
          return
        }
        var cbArgs = Array.prototype.slice.call(arguments)
        cbArgs.shift()
        console.info(name + ' callback was invoked.')
      }
      old.apply(this, args)
    }
  })
  client.session()
  return client
}

function createSodaClient(name) {
  if (process.env.SAUCE) {
    return soda.createSauceClient({
      'url': 'http://fbrell.com/',
      'username': settings.sauce.user,
      'access-key': settings.sauce.key,
      'os': process.env.SAUCE_OS || 'Windows 2003',
      'browser': process.env.SAUCE_BROWSER || 'firefox',
      'browser-version': process.env.SAUCE_BROWSER_VERSION || '3.6',
      'max-duration': 300,
      'name': name,
    })
  } else {
    return soda.createClient({
      url: 'http://local.fbrell.com:43600/',
      host: '127.0.0.1',
      port: 4444,
    })
  }
}

function sodaTest(name, test) {
  exports[name] = function(beforeExit) {
    var passed = false
    test(createSodaClient(name).chain.session().setTimeout(5000))
      .testComplete()
      .end(function(er) {
        if (er) throw er
        passed = true
      })
    beforeExit(function() {
      assert.ok(passed, name + ' passed')
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

function fbPopupLogin(opts) {
  opts = opts || {}
  return function(browser) {
    browser
      .waitForPopUp()
      .selectPopUp('Log In | Facebook')
      .and(waitAssertTextPresent('Log in to use your Facebook account with'))
      .type('id=email', opts.email || settings.facebookTestUser.email)
      .type('id=pass', opts.pass || settings.facebookTestUser.pass)
      .click('name=login')
      .deselectPopUp()
  }
}

function runLoggedInExample(opts) {
  return function(browser) {
    browser
      .open(opts.url)
      .waitForPageToLoad(2000)
      .click('css=.login-button')
      .and(fbPopupLogin(opts))
      .and(waitAssertTextPresent('connected'))
      .click('css=.run-code')
  }
}

function runExample(opts) {
  return function(browser) {
    browser
      .open(opts.url)
      .waitForPageToLoad(2000)
      .click('css=.run-code')
  }
}

function runInIFrame(selector, inIFrame) {
  return function(browser) {
    inIFrame(
      browser
        .waitForPageToLoad(1000)
        .selectFrame(selector)
    ).selectFrame('relative=top')
  }
}

var runInIFrameDialog = runInIFrame.bind(null, 'css=.fb_dialog_iframe iframe')
var runInIFramePlugin = runInIFrame.bind(null, 'css=#jsroot iframe')

sodaTest('home and examples page', function(browser) {
  return browser
    .open('/')
    .waitForPageToLoad(2000)
    .click('link=Examples')
    .waitForPageToLoad(2000)
    .and(waitAssertTextPresent('account-info'))
})

sodaTest('login and logout', function(browser) {
  return browser
    .and(runExample({ url: '/saved/79f242491065630b3dfe66cde4fa532b'}))
    .and(waitAssertTextPresent('unknown'))
    .click('css=#fb-login')
    .and(fbPopupLogin())
    .and(waitAssertTextPresent('User has logged in'))
    .and(waitAssertTextPresent('connected'))
    .click('css=#fb-logout')
    .and(waitAssertTextPresent('User has logged out'))
    .and(waitAssertTextPresent('unknown'))
})

sodaTest('user info via API', function(browser) {
  return browser
    .and(runLoggedInExample({ url: '/fb.api/user-info?autoRun' }))
    .and(waitAssertTextPresent('first_name'))
})

sodaTest('cancel feed iframe dialog', function(browser) {
  return browser
    .and(runLoggedInExample({ url: '/saved/afedcd65e0c7fe1258468b96514d2c48' }))
    .and(runInIFrameDialog(function(browser) {
      return browser
        .and(waitAssertTextPresent('Post to Your Wall'))
        .click('css=#cancel input')
    }))
    .and(waitAssertTextPresent('Did not publish to the feed'))
})

sodaTest('post via feed iframe dialog', function(browser) {
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
    .waitForPageToLoad(2000)
    .and(waitAssertTextPresent(message))
})

sodaTest('like and unlike with edge events', function(browser) {
  var url = '/saved/dfba30ac7d85862f1da8c9e2c5f20228'
  return browser
    .and(runLoggedInExample({ url: url }))
    .and(runInIFramePlugin(function(browser) {
      return browser
        .click('css=.like_button_no_like')
    }))
    .and(waitAssertTextPresent('You liked http://fbrell.com/'))
    .and(runExample({ url: url }))
    .and(runInIFramePlugin(function(browser) {
      return browser
        .click('css=.like_button_like .tombstone_cross')
    }))
    .and(waitAssertTextPresent('You unliked http://fbrell.com/'))
})
