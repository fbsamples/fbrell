var fb = require('./infra/core.js')
  , soda = require('./infra/soda.js')

soda.sodaTest(exports, 'user info via API', function(browser) {
  return browser
    .and(fb.runLoggedInExample({ url: '/fb.api/user-info?autoRun' }))
    .and(fb.waitAssertTextPresent('first_name'))
})