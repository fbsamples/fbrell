fb = require('./infra.js')

fb.sodaTest(exports, 'home and examples page', function(browser) {
  return browser
    .open('/')
    .waitForPageToLoad(2000)
    .click('link=Examples')
    .waitForPageToLoad(2000)
    .and(fb.waitAssertTextPresent('account-info'))
})