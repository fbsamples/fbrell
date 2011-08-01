var assert = require('assert')
  , request = require('request')
  , settings = require('./../../settings.js')

var graph_url = 'https://graph.facebook.com/'
  , access_token = null

/**
 * Generic API Wrappers
 */
var callWithAccessToken = exports.callWithAccessToken = function(url, cb) {
  requestAccessToken(function(er, access_token) {
    if (er) cb(er)
    else {
      url = graph_url + url + access_token
      request({ uri: url }, function(er, response, body) {
        passResponseToCallback(er, response, body, cb)
      })
    }
  })
}

var requestAccessToken = exports.requestAccessToken = function(cb) {
  if (access_token) cb(null, access_token)
  else {
    var url = graph_url + 'oauth/access_token?' +
              'client_id=' + settings.facebook.id +
              '&client_secret=' + settings.facebook.secret +
              '&grant_type=client_credentials'

    request({ uri: url }, function(er, response, body) {
      passResponseToCallback(er, response, body, cb)
    })
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