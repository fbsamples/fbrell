fb = require('./infra.js')

fb.sodaTest(exports, 'user info via API', function(browser) {
  return browser
    .and(fb.runLoggedInExample({ url: '/fb.api/user-info?autoRun' }))
    .and(fb.waitAssertTextPresent('first_name'))
})