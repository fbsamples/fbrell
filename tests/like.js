fb = require('./infra.js')

fb.sodaTest(exports, 'like and unlike with edge events', function(browser) {
  var url = '/saved/dfba30ac7d85862f1da8c9e2c5f20228'
  return browser
    .and(fb.runLoggedInExample({ url: url }))
    .and(fb.runInIFramePlugin(function(browser) {
      return browser.click('css=.like_button_no_like')
    }))
    .and(fb.waitAssertTextPresent('You liked http://fbrell.com/'))
    .and(fb.runExample({ url: url }))
    .and(fb.runInIFramePlugin(function(browser) {
      return browser.click('css=.like_button_like .tombstone_cross')
    }))
    .and(fb.waitAssertTextPresent('You unliked http://fbrell.com/'))
})
