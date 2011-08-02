var assert = require('assert')
  , request = require('request')
  , qs = require('querystring')
  , settings = require('./../../settings.js')

var graphUrl = 'https://graph.facebook.com/'
var fbUrl = 'https://www.facebook.com/'
var rellCode = null

/**
 * Generic API Wrappers
 */
var callWithAccessToken = exports.callWithAccessToken = function(url, cb) {
  requestAccessToken(function(er, accessToken) {
    if (er) cb(er)
    else {
      fullUrl = graphUrl + url + accessToken
      request({ uri: fullUrl }, function(er, response, body) {
        passResponseToCallback(er, response, body, cb)
      })
    }
  })
}

var requestAccessToken = exports.requestAccessToken = function(cb) {
  var url = graphUrl + 'oauth/access_token?' +
            'client_id=' + settings.facebook.id +
            '&client_secret=' + settings.facebook.secret +
            '&grant_type=client_credentials'

  request({ uri: url }, function(er, response, body) {
    passResponseToCallback(er, response, body, cb)
  })
}

var requestUserAccessToken = exports.requestUserAccessToken =
function() {
  var url = fbUrl + 'dialog/oauth?' + 
            'client_id=' + settings.facebook.id + 
            '&redirect_uri=http://fbrell.com'

  return function(browser) {
    browser.openWindow(url, 'login')
           .waitForPopUp('login', 10000)
           .selectWindow('login')
           .waitForTitle('Welcome — Facebook Read Eval Log Loop')
           .assertTitle('Welcome — Facebook Read Eval Log Loop')
           .waitForCondition('rellCode = browser.url', 10000)
           .openWindow(rellCode, 'newwindow')
  }
}

/**
 * Test User Wrappers
 */
exports.createTestUser = function(cb) {
  url = settings.facebook.id + '/accounts/test-users?' +
        'installed=true' + 
        '&name=TestUser' +
        '&permissions=read_stream' +
        '&method=post&'
  callWithAccessToken(url, function(er, body) {
    assert.ok(!er, 'Successfully create a test user')
    cb(er, JSON.parse(body))
  })
}

// private
function passResponseToCallback(er, response, body, cb) {
  if (!er && response.statusCode == 200) cb(null, body)
  else cb(er, '')
}