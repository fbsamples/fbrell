var fb = require('./infra/core.js')
  , soda = require('./infra/soda.js')

soda.makeTest(exports, 'login and logout', function(browser) {
  return browser
    .and(fb.runExample({ url: '/saved/79f242491065630b3dfe66cde4fa532b'}))
    .and(fb.waitAssertTextPresent('unknown'))
    .click('css=#fb-login')
    .and(fb.popupLogin())
    .and(fb.waitAssertTextPresent('User has logged in'))
    .and(fb.waitAssertTextPresent('connected'))
    .click('css=#fb-logout')
    .and(fb.waitAssertTextPresent('User has logged out'))
    .and(fb.waitAssertTextPresent('unknown'))
})