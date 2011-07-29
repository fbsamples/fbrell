var fb = require('./infra/core.js')
  , api = require('./infra/api.js')
  , soda = require('./infra/soda.js')
  , settings = require('./../settings.js')

var like_url = 'saved/13859253c3801058354b280b6109b8bf'

/**
 * Tests
 */
soda.sodaTest(exports, 'like/unlike with edge events', function(browser) {
  return browser
    .and(fb.runLoggedInExample({ url: like_url }))
    .and(clickUnlike(browser)) // ensure we're unliked to start.
    .and(clickLike(browser))
    .and(fb.waitAssertTextPresent('You liked http://fbrell.com/'))
    .and(fb.assertStreamContainsLink(og_url))
    .and(clickUnlike(browser))
    .and(fb.waitAssertTextPresent('You unliked http://fbrell.com/'))
    .and(fb.assertStreamDoesNotContainLink(og_url))
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