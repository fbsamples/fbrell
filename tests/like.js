var fb = require('./infra/core.js')
  , api = require('./infra/api.js')
  , soda = require('./infra/soda.js')
  , settings = require('./../settings.js')

var edge_event_url = '/saved/dfba30ac7d85862f1da8c9e2c5f20228'

/**
 * Tests
 */
soda.sodaTest(exports, 'like/unlike with edge events', function(browser) {
  return browser
    .and(fb.runLoggedInExample({ url: edge_event_url}))
    .and(clickLike(browser))
    .and(fb.waitAssertTextPresent('You liked http://fbrell.com/'))
    .and(clickUnlike(browser))
    .and(fb.waitAssertTextPresent('You unliked http://fbrell.com/'))
})

/**
 * Utility
 */
function clickLike(browser) {
  return fb.runInIFramePlugin(function(browser) {
    return browser.click('css=.like_button_no_like')
  })
}

function clickUnlike(browser) {
  return fb.runInIFramePlugin(function(browser) {
    return browser.click('css=.like_button_like .tombstone_cross')
  })
}