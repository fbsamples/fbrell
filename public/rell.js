// stolen from prototypejs
// used to set innerHTML and execute any contained <scripts>
var ScriptSoup ={
  _scriptFragment: '<script[^(>|fbml)]*>([\\S\\s]*?)<\/script>',
  set: function(el, html) {
    el.innerHTML = ScriptSoup.stripScripts(html);
    ScriptSoup.evalScripts(html);
  },

  stripScripts: function(html) {
    return html.replace(new RegExp(ScriptSoup._scriptFragment, 'img'), '');
  },

  evalScripts: function(html) {
    var
      parts = html.match(new RegExp(ScriptSoup._scriptFragment, 'img')) || [],
      matchOne = new RegExp(ScriptSoup._scriptFragment, 'im');
    for (var i=0, l=parts.length; i<l; i++) {
      try {
        eval((parts[i].match(matchOne) || ['', ''])[1]);
      } catch(e) {
        Log.error('Error running example: ' + e, e);
      }
    }
  }
};

Rell = {
  /**
   * go go go
   */
  init: function(config) {
    Rell.config = config;
    Log.init(document.getElementById('log'), Rell.config.level);
    Log.debug('Configuration', Rell.config);
    (Rell['init_' + Rell.config.version] || Rell.init_old)();

    Rell._editor = CodeMirror.fromTextArea('jscode', {
      height: '390px',
      content: Rell.getCode(),
      parserfile: [
        'parsexml.js',
        'parsecss.js',
        'tokenizejavascript.js',
        'parsejavascript.js',
        'parsehtmlmixed.js'
      ],
      stylesheet: [
        '/codemirror/css/xmlcolors.css',
        '/codemirror/css/jscolors.css',
        '/codemirror/css/csscolors.css',
        '/custom-codemirror.css'
      ],
      path: '/codemirror/js/',
      autoMatchParens: true,
      lineNumbers: true
    });
  },

  /**
   * name is magical
   */
  init_old: function() {
    Log.debug('FB_RequireFeatures & FB.Facebook.init');

    FB_RequireFeatures(['XFBML'], function() {
      Log.debug('FB_RequireFeatures callback invoked');

      // NOTE: replace built in logging with custom logging
      FB.FBDebug.dump = function(obj, name) {
        Log.debug(name, obj);
      };
      FB.FBDebug.writeLine = function(line) {
        Log.debug(line);
      };

      FB.FBDebug.isEnabled = true;
      FB.FBDebug.logLevel = Rell.config.old_debug;

      FB.Facebook.init(Rell.config.apikey, '/xd_receiver.html');
      // sigh
      window.setInterval(function() {
        var
          result = FB.Connect._singleton._status.result,
          status = 'unknown';
        if (result == 1) {
          status = 'connected';
        } else if (result == 3) {
          status = 'notConnected';
        }
        var el = document.getElementById('auth-status');
        el.className = status;
        el.innerHTML = status;
      }, 500);

      Rell.runCode();
    });
  },

  /**
   * name is magical
   */
  init_mu: function() {
    FB.Event.subscribe('fb.log', Log.info.bind('fb.log'));
    FB.Event.subscribe('auth.statusChange', function(response) {
      var el = document.getElementById('auth-status');
      el.className = response.status;
      el.innerHTML = response.status;
    });

    if (Rell.config.trace && Rell.config.trace !== '0') {
      Tracer.instrument('FB', FB);
    }

    FB.init({
      apiKey: Rell.config.apikey,
      cookie: true,
      status: true
    });
    Rell.runCode();
  },

  /**
   * Run's the code in the textarea.
   */
  runCode: function() {
    var root = document.getElementById('jsroot');
    ScriptSoup.set(root, Rell.getCode());
    if (Rell.config.version == 'mu') {
      FB.XFBML.parse(root);
    } else {
      FB.XFBML.Host.parseDomTree(root);
    }
  },

  getCode: function() {
    if (Rell._editor && Rell._editor.editor) {
      return Rell._editor.getCode();
    } else {
      return document.getElementById('jscode').value;
    }
  },

  login: function() {
    if (Rell.config.version == 'mu') {
      FB.login(Log.debug.bind('FB.login callback'));
    } else {
      FB.Connect.requireSession(Log.debug.bind('requireSession callback'));
    }
  },

  logout: function() {
    if (Rell.config.version == 'mu') {
      FB.logout(Log.debug.bind('FB.logout callback'));
    } else {
      FB.Connect.logout(Log.debug.bind('FB.Connect.logout callback'));
    }
  },

  disconnect: function() {
    if (Rell.config.version == 'mu') {
      FB.api({ method: 'Auth.revokeAuthorization' }, Log.debug.bind('revokeAuthorization callback'));
    } else {
      FB.Facebook.apiClient.revokeAuthorization(null, Log.debug.bind('revokeAuthorization callback'));
    }
  },

  addHiddenInput: function(form, name, value) {
    var el = document.createElement('input');
    el.type = 'hidden';
    el.name = name;
    el.value = value;
    form.appendChild(el);
  }
};
