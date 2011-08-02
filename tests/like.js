var fb = require('./infra/core.js')
  , api = require('./infra/api.js')
  , soda = require('./infra/soda.js')
  , settings = require('./../settings.js')

var likeUrl = 'saved/f0c173b5d7308d03853db553c497a7b0'

/**
 * Tests
 */
soda.makeTest(exports, 'like/unlike with edge events', function(browser) {
  var ogUrl = 'http://fbrell.com/og/website/blah'
  return browser
    .and(fb.runLoggedInExample({ url: likeUrl }))
    .and(clickLike(browser))
    .and(fb.waitAssertTextPresent('You liked ' + ogUrl))
    .and(fb.assertLikesPage('blah'))
    .and(fb.assertStreamContainsLink(ogUrl))
    .and(clickUnlike(browser))
    .and(fb.waitAssertTextPresent('You unliked ' + ogUrl))
    .and(fb.assertStreamDoesNotContainLink(ogUrl))
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