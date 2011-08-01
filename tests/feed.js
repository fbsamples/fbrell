var fb = require('./infra/core.js')
  , soda = require('./infra/soda.js')

soda.makeTest(exports, 'cancel feed iframe dialog', function(browser) {
  return browser
    .and(fb.runLoggedInExample({
      url: '/saved/afedcd65e0c7fe1258468b96514d2c48' }))
    .and(fb.runInIFrameDialog(function(browser) {
      return browser
        .and(fb.waitAssertTextPresent('Post to Your Wall'))
        .click('css=#cancel input')
    }))
    .and(fb.waitAssertTextPresent('Did not publish to the feed'))
})

// TODO(jubishop): I think this test is broken
soda.makeTest(exports, 'post via feed iframe dialog', function(browser) {
  var message = 'Test run at ' + Date.now()
  return browser
    .and(fb.runLoggedInExample({
      url: '/saved/afedcd65e0c7fe1258468b96514d2c48' }))
    .and(fb.runInIFrameDialog(function(browser) {
      return browser
        .and(fb.waitAssertTextPresent('Post to Your Wall'))
        .type('id=feedform_user_message', message)
        .click('css=#publish input')
        .waitForPageToLoad(1000)
    }))
    .and(fb.waitAssertTextPresent('Successfully published to the feed'))
    .open('http://www.facebook.com/profile.php')
    .waitForPageToLoad(2000)
    .and(fb.waitAssertTextPresent(message))
})
