var soda = require('soda')
  , assert = require('assert')
  , settings = require('./../settings.js')

exports.browser = function() {
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

exports.sodaTest = function(exports, name, test) {
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
      url: 'http://fbrell.com/',
      host: '127.0.0.1',
      port: 4444,
    })
  }
}

waitAssertTextPresent = exports.waitAssertTextPresent = function(text) {
  return function(browser) {
    browser
      .waitForTextPresent(text)
      .assertTextPresent(text)
  }
}

fbPopupLogin = exports.fbPopupLogin = function(opts) {
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

exports.runLoggedInExample = function(opts) {
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

exports.runExample = function(opts) {
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

exports.runInIFrameDialog = runInIFrame.bind(null,
                                             'css=.fb_dialog_iframe iframe')
exports.runInIFramePlugin = runInIFrame.bind(null, 'css=#jsroot iframe')
