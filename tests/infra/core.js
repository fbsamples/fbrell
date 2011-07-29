var assert = require('assert')
  , settings = require('./../../settings.js')

var wall_url = 'http://www.facebook.com/profile.php?sk=wall'

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

var waitAssertLinkPresent = exports.waitAssertLinkPresent = function(url) {
  return function(browser) {
    browser
      .waitForElementPresent(linkXPath(url))
      .assertElementPresent(linkXPath(url))
  }
}

function linkXPath(url) {
  return "xpath=//a[contains(@href,'" + url + "')]"
}

/**
 * FB UI LIBS
 *
 * Generically useful functions for fb-ui testing
 */
var assertStreamContainsLink = exports.assertStreamContainsLink =
function(url) {
  return function(browser) {
    browser
      .openWindow(wall_url, 'wall')
      .waitForPopUp('wall', 1000)
      .selectWindow('wall')
      .and(waitAssertLinkPresent(url))
      .close()
      .selectWindow()
  }
}

var assertStreamDoesNotContainLink = exports.assertStreamDoesNotContainLink =
function(url) {
  return function(browser) {
    browser
      .openWindow(wall_url, 'wall')
      .waitForPopUp('wall', 1000)
      .selectWindow('wall')
      .waitForPageToLoad(4000)
      .assertElementNotPresent(linkXPath(url))
      .close()
      .selectWindow()
  }
}

var popupLogin = exports.popupLogin = function(opts) {
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
      .and(popupLogin(opts))
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
