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

var Rell = {
  /**
   * go go go
   */
  init: function() {
    var example = window.rellExample

    window.location.hash = ''
    window.rellConfig.autoRun = example ? example.autoRun : false
    Log.init($('#log')[0], window.rellConfig.level)
    Log.debug('Configuration', window.rellConfig);

    FB.Event.subscribe('fb.log', Log.info.bind('fb.log'))
    FB.Event.subscribe('auth.login', function(response) {
      Log.info('auth.login event', response)
    })
    FB.Event.subscribe('auth.statusChange', Rell.onStatusChange)

    if (!window.rellConfig.init) {
      return;
    }

    var options = {
      appId: window.rellConfig.appID,
      version: window.rellConfig.version,
      cookie: true,
      status: window.rellConfig.status,
      frictionlessRequests: window.rellConfig.frictionlessRequests
    }

    FB.init(options)
    if (top != self) {
      FB.Canvas.setAutoGrow(true)
    }

    if (!window.rellConfig.status) {
      Rell.autoRunCode()
    } else {
      FB.getLoginStatus(function() { Rell.autoRunCode() })
      FB.getLoginStatus(Rell.onStatusChange)
    }

    $('#rell-login').click(Rell.login)
    $('#rell-disconnect').click(Rell.disconnect)
    $('#rell-logout').click(Rell.logout)
    $('#rell-run-code').click(Rell.runCode)
    $('#rell-log-clear').click(Rell.clearLog)
    Rell.setCurrentViewMode()
    if (example && !example.autoRun) {
      Rell.setupAutoRunPopover()
    }
  },

  onStatusChange: function(response) {
    var status = response.status
    $('#auth-status').removeClass().addClass(status).html(status)
  },

  autoRunCode: function() {
    if (window.rellConfig.autoRun) Rell.runCode()
  },

  /**
   * Run's the code in the textarea.
   */
  runCode: function() {
    Log.info('Executed example')
    var root = $('#jsroot')[0]
    ScriptSoup.set(root, Rell.getCode())
    FB.XFBML.parse(root)
  },

  getCode: function() {
    return $('#jscode').val()
  },

  login: function() {
    FB.login(Log.debug.bind('FB.login callback'))
  },

  logout: function() {
    FB.logout(Log.debug.bind('FB.logout callback'))
  },

  disconnect: function() {
    FB.api('/me/permissions', 'DELETE', Log.debug.bind('revokeAuthorization callback'))
  },

  setCurrentViewMode: function() {
    var select = $('#rell-view-mode')
    if (window.name.indexOf('canvas') > -1) {
      select.val('canvas') // context.Canvas
    } else if (window.name.indexOf('app_runner') > -1) {
      select.val('page-tab') // context.PageTab
    } else if (self === top) {
      select.val('website') // context.Website
    }
  },

  setupAutoRunPopover: function() {
    var el = $('#rell-run-code')
    el.popover('show')
    el.hover(function() { el.popover('hide') })
  },

  clearLog: function() {
    Log.clear()
    return false
  }
}

window.fbAsyncInit = Rell.init

// Settings panel: live URL preview and update button
$(function() {
  var defaults = {
    appid: '342526215814610',
    server: '',
    version: 'v25.0',
    locale: 'en_US',
    level: 'debug',
    'view-mode': 'website',
    init: 'true',
    status: 'true',
    frictionlessRequests: 'true'
  }

  function buildSettingsUrl() {
    var params = {}
    $('.rell-setting').each(function() {
      var $el = $(this)
      var name = $el.attr('name')
      var val
      if ($el.is(':checkbox')) {
        val = $el.is(':checked') ? 'true' : 'false'
      } else {
        val = $el.val()
      }
      if (val !== '' && val !== defaults[name]) {
        params[name] = val
      }
    })
    var path = window.location.pathname
    var qs = $.param(params)
    return qs ? path + '?' + qs : path
  }

  function updatePreview() {
    var url = buildSettingsUrl()
    $('#rell-url-preview').html('<small>' + url + '</small>')
  }

  $('.rell-setting').on('change input', updatePreview)

  $('#rell-settings-update').on('click', function() {
    window.location = buildSettingsUrl()
  })
})
