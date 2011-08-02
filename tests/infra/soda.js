var assert = require('assert')
  , soda = require('soda')

/**
 * Browser
 * Use this to interact with the browser on the command line.
 */
exports.browser = function() {
  var client = createSodaClient('interactive')
    , wrapMethods = soda.commands.slice(0)
  wrapMethods.push('session')
  wrapMethods.forEach(function(name) {
    var old = client[name]
    client[name] = function() {
      var args = Array.prototype.slice.call(arguments)
        , index = name === 'session' ? 0 : 2
      args[index] = function(er) {
        if (er) {
          console.error(name, er)
          return
        }
        var cbArgs = Array.prototype.slice.call(arguments)
        cbArgs.shift()
        console.info(name + ' callback was invoked.')
      }
      old.apply(this, args)
    }
  })
  client.session()
  return client
}

/**
 * SODA
 */
exports.makeTest = function(exports, name, test) {
  exports[name] = function(beforeExit) {
    var passed = false
    test(createSodaClient(name).chain.session())
      .testComplete()
      .end(function(er) {
        if (er) throw er
        passed = true
      })
    beforeExit(function() {
      assert.ok(passed, name + ' passed')
    })
  }
}

// private
function createSodaClient(name) {
  if (process.env.SAUCE) {
    return soda.createSauceClient({
      'url': 'http://fbrell.com/',
      'username': settings.sauce.user,
      'access-key': settings.sauce.key,
      'os': process.env.SAUCE_OS || 'Windows 2003',
      'browser': process.env.SAUCE_BROWSER || 'firefox',
      'browser-version': process.env.SAUCE_BROWSER_VERSION || '3.6',
      'max-duration': 300,
      'name': name,
    })
  } else {
    return soda.createClient({
      url: 'http://fbrell.com/',
      host: '127.0.0.1',
      port: 4444,
    })
  }
}