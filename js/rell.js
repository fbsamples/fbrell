var Log = require('./log')
  , Tracer = require('./tracer')

// stolen from prototypejs
// used to set innerHTML and execute any contained <scripts>
var ScriptSoup ={
  _scriptFragment: '<script[^(>|fbml)]*>([\\S\\s]*?)<\/script>',
  set: function(el, html) {
    el.innerHTML = ScriptSoup.stripScripts(html)
    ScriptSoup.evalScripts(html)
  },

  stripScripts: function(html) {
    return html.replace(new RegExp(ScriptSoup._scriptFragment, 'img'), '')
  },

  evalScripts: function(html) {
    var
      parts = html.match(new RegExp(ScriptSoup._scriptFragment, 'img')) || [],
      matchOne = new RegExp(ScriptSoup._scriptFragment, 'im')
    for (var i=0, l=parts.length; i<l; i++) {
      try {
        eval((parts[i].match(matchOne) || ['', ''])[1])
      } catch(e) {
        Log.error('Error running example: ' + e, e)
      }
    }
  }
}

function $(id) {
  return document.getElementById(id)
}

var Rell = {
  /**
   * go go go
   */
  init: function(config, example) {
    Rell.config = config
    Rell.config.autoRun = example ? example.autoRun : false
    Log.init($('log'), Rell.config.level)
    Log.debug('Configuration', Rell.config);
    (Rell['init_' + Rell.config.version] || Rell.init_old)()
    $('rell-login').onclick = Rell.login
    $('rell-disconnect').onclick = Rell.disconnect
    $('rell-logout').onclick = Rell.logout
    $('rell-run-code').onclick = Rell.runCode
    $('rell-log-clear').onclick = Log.clear
    $('rell-view-mode').onchange = Rell.onURLSelectorChange
    if (config.isEmployee)
      $('rell-env').onchange = Rell.onURLSelectorChange
    Rell.setCurrentViewMode()
  },

  /**
   * name is magical
   */
  init_old: function() {
    Log.debug('FB_RequireFeatures & FB.Facebook.init')

    FB_RequireFeatures(['XFBML'], function() {
      Log.debug('FB_RequireFeatures callback invoked')

      // NOTE: replace built in logging with custom logging
      FB.FBDebug.dump = function(obj, name) {
        Log.debug(name, obj)
      }
      FB.FBDebug.writeLine = function(line) {
        Log.debug(line)
      }

      FB.FBDebug.isEnabled = true
      FB.FBDebug.logLevel = 6

      var xd_receiver = '/public/xd_receiver.html'
      if (document.location.protocol == 'https:') {
        xd_receiver = '/public/xd_receiver_ssl.html'
      }
      FB.Facebook.init(Rell.config.appID, xd_receiver)
      // sigh
      window.setInterval(function() {
        var
          result = FB.Connect._singleton._status.result,
          status = 'unknown'
        if (result == 1) {
          status = 'connected'
        } else if (result == 3) {
          status = 'notConnected'
        }
        var el = $('auth-status')
        el.className = status
        el.innerHTML = status
      }, 500)

      Rell.autoRunCode()
    })
  },

  /**
   * name is magical
   */
  init_mu: function() {
    if (!window.FB) {
      Log.error('SDK failed to load.')
      return
    }

    FB.Event.subscribe('fb.log', Log.info.bind('fb.log'))
    FB.Event.subscribe('auth.login', function(response) {
      Log.info('auth.login event', response)
    })
    FB.Event.subscribe('auth.statusChange', Rell.onStatusChange)

    if (Rell.config.trace) {
      Tracer.instrument('FB', FB)
    }

    var options = {
      appId : Rell.config.appID,
      cookie: true,
      status: Rell.config.status,
      frictionlessRequests: Rell.config.frictionlessRequests
    }

    if (Rell.config.channel) {
      options.channelUrl = Rell.config.channelURL
    }

    FB.init(options)
    if (top != self) {
      FB.Canvas.setAutoGrow(true)
    }

    if (!Rell.config.status) {
      Rell.autoRunCode()
    } else {
      FB.getLoginStatus(function() { Rell.autoRunCode() })
      FB.getLoginStatus(Rell.onStatusChange)
    }
  },

  onStatusChange: function(response) {
    var el = $('auth-status')
    el.className = response.status
    el.innerHTML = response.status
  },

  autoRunCode: function() {
    if (Rell.config.autoRun) Rell.runCode()
  },

  /**
   * Run's the code in the textarea.
   */
  runCode: function() {
    var root = $('jsroot')
    ScriptSoup.set(root, Rell.getCode())
    if (Rell.config.version == 'mu') {
      FB.XFBML.parse(root)
    } else {
      FB.XFBML.Host.parseDomTree(root)
    }
  },

  getCode: function() {
    return $('jscode').value
  },

  login: function() {
    if (Rell.config.version == 'mu') {
      FB.login(Log.debug.bind('FB.login callback'))
    } else {
      FB.Connect.requireSession(Log.debug.bind('requireSession callback'))
    }
  },

  logout: function() {
    if (Rell.config.version == 'mu') {
      FB.logout(Log.debug.bind('FB.logout callback'))
    } else {
      FB.Connect.logout(Log.debug.bind('FB.Connect.logout callback'))
    }
  },

  disconnect: function() {
    if (Rell.config.version == 'mu') {
      FB.api({ method: 'Auth.revokeAuthorization' }, Log.debug.bind('revokeAuthorization callback'))
    } else {
      FB.Facebook.apiClient.revokeAuthorization(null, Log.debug.bind('revokeAuthorization callback'))
    }
  },

  onURLSelectorChange: function(e) {
    top.location = this[this.selectedIndex].getAttribute('data-url')
  },

  setCurrentViewMode: function() {
    var select = $('rell-view-mode')
    if (window.name.indexOf('canvas') > -1) {
      select.value = 'canvas' // context.Canvas
    } else if (window.name.indexOf('app_runner') > -1) {
      select.value = 'page-tab' // context.PageTab
    } else if (self === top) {
      select.value = 'website' // context.Website
    }
  }
}

if (typeof module !== 'undefined') module.exports = Rell
