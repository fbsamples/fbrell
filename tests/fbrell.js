var fb = require('./infra/core.js')
  , soda = require('./infra/soda.js')

soda.sodaTest(exports, 'home and examples page', function(browser) {
  return browser
    .open('/')
    .waitForPageToLoad(2000)
    .click('link=Examples')
    .waitForPageToLoad(2000)
    .and(fb.waitAssertTextPresent('account-info'))
})