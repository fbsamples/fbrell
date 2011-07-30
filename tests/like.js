var fb = require('./infra/core.js')
  , api = require('./infra/api.js')
  , soda = require('./infra/soda.js')
  , settings = require('./../settings.js')

var like_url = 'saved/f0c173b5d7308d03853db553c497a7b0'

/**
 * Tests
 */
soda.sodaTest(exports, 'like/unlike with edge events', function(browser) {
  og_url = 'http://fbrell.com/og/website/blah'
  return browser
    .and(fb.runLoggedInExample({ url: like_url }))
    .and(clickLike(browser))
    .and(fb.waitAssertTextPresent('You liked ' + og_url))
    .and(fb.assertStreamContainsLink(og_url))
    .and(clickUnlike(browser))
    .and(fb.waitAssertTextPresent('You unliked ' + og_url))
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