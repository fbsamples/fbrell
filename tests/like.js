var fb = require('./infra/core.js')
  , api = require('./infra/api.js')
  , soda = require('./infra/soda.js')
  , settings = require('./../settings.js')

api.createTestUser(function(er, test_user) {
  soda.sodaTest(exports, 'like and unlike with edge events', function(browser) {
    var url = '/saved/dfba30ac7d85862f1da8c9e2c5f20228'
    return browser
      .and(fb.runLoggedInExample({
        url: url, email: test_user.email, pass: test_user.password }))
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
})
