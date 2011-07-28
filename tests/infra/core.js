var assert = require('assert')
  , settings = require('./../../settings.js')


/**
 * UI LIBS
 *
 * Generically useful functions for UI testing
 */
var waitAssertTextPresent = exports.waitAssertTextPresent = function(text) {
  return function(browser) {
    browser
      .waitForTextPresent(text)
      .assertTextPresent(text)
  }
}


/**
 * FB UI LIBS
 *
 * Generically useful functions for fb-ui testing
 */
var fbPopupLogin = exports.fbPopupLogin = function(opts) {
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


/**
 * FBRELL
 *
 * Generically useful interactions with fbrell
 */
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

exports.runInIFrameDialog = runInIFrame.bind(
  null, 'css=.fb_dialog_iframe iframe')
exports.runInIFramePlugin = runInIFrame.bind(
  null, 'css=#jsroot iframe')
