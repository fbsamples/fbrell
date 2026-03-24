/**
 * Rell - Main application for Facebook Rell (FB Examples).
 *
 * Orchestrates the editor (CodeMirror), FB SDK initialization, code execution,
 * authentication actions, settings drawer, and status bar.
 *
 * Globals defined here: ScriptSoup, Rell
 * Globals consumed: CodeMirror, Log, Theme, Resize, Sidebar, FB, rellConfig, rellExample
 *
 * Load order (all scripts in HTML):
 *   1. vendor/codemirror.min.js
 *   2. js/errors.js
 *   3. js/log.js
 *   4. js/theme.js
 *   5. js/resize.js
 *   6. js/sidebar.js
 *   7. js/rell.js  <-- this file
 */

// ---------------------------------------------------------------------------
// ScriptSoup - Extract and eval <script> tags from HTML strings.
// Used by Rell.runCode() to execute inline scripts in example code.
// ---------------------------------------------------------------------------
var ScriptSoup = {
  _scriptFragment: '<script[^(>|fbml)]*>([\\S\\s]*?)<\/script>',

  /**
   * Set innerHTML on an element, stripping scripts first, then evaling them.
   * @param {HTMLElement} el - Target element
   * @param {string} html - HTML string that may contain script tags
   */
  set: function(el, html) {
    el.innerHTML = ScriptSoup.stripScripts(html);
    ScriptSoup.evalScripts(html);
  },

  /**
   * Remove all script tags from an HTML string.
   * @param {string} html
   * @returns {string} HTML without script tags
   */
  stripScripts: function(html) {
    return html.replace(new RegExp(ScriptSoup._scriptFragment, 'img'), '');
  },

  /**
   * Extract and eval all script tag contents from an HTML string.
   * @param {string} html
   */
  evalScripts: function(html) {
    var parts = html.match(new RegExp(ScriptSoup._scriptFragment, 'img')) || [];
    var matchOne = new RegExp(ScriptSoup._scriptFragment, 'im');
    for (var i = 0; i < parts.length; i++) {
      try {
        eval((parts[i].match(matchOne) || ['', ''])[1]);
      } catch (e) {
        Log.error('Error running example: ' + e, e);
      }
    }
  }
};

// ---------------------------------------------------------------------------
// Rell - Main application object
// ---------------------------------------------------------------------------
var Rell = {
  editor: null, // CodeMirror instance

  /**
   * Initialize UI components. Called on DOMContentLoaded — does NOT require FB SDK.
   */
  initUI: function() {
    var example = window.rellExample;

    // Initialize log
    Log.init(document.getElementById('log'), window.rellConfig ? window.rellConfig.level : 'debug');

    // Initialize CodeMirror editor
    Rell.initEditor();

    // Initialize theme
    Theme.init();

    // Initialize resize handles
    Resize.init();

    // Initialize sidebar
    Sidebar.init();

    // Initialize mobile tab switching
    Rell.initMobileTabs();

    // Bind click handlers (vanilla JS)
    Rell.bindClick('rell-run-code', Rell.runCode);
    Rell.bindClick('rell-log-clear', Rell.clearLog);
    Rell.bindClick('rell-disconnect', Rell.disconnect);
    Rell.bindClick('fb-login-custom', Rell.loginToggle);

    // Initialize settings drawer
    Rell.initSettings();

    // When custom login is disabled, switch to XFBML plugin mode
    if (window.rellConfig && !window.rellConfig.customLogin) {
      var authControls = document.querySelector('.auth-controls');
      if (authControls) authControls.classList.add('plugin-login-active');
    }

    // Keyboard shortcut: Cmd/Ctrl+Enter to run code
    document.addEventListener('keydown', function(e) {
      if ((e.metaKey || e.ctrlKey) && e.key === 'Enter') {
        e.preventDefault();
        Rell.runCode();
      }
    });

    // Auto-run hint on the run button
    if (example && !example.autoRun) {
      var runBtn = document.getElementById('rell-run-code');
      if (runBtn) runBtn.title = 'Click to Run \u2014 this example does not run automatically';
    }
  },

  _sdkInitialized: false,

  /**
   * Initialize FB SDK integration. Called via window.fbAsyncInit when SDK loads.
   */
  initSDK: function() {
    if (Rell._sdkInitialized) return;
    Rell._sdkInitialized = true;
    var example = window.rellExample;
    if (window.rellConfig) {
      window.rellConfig.autoRun = example ? example.autoRun : false;
    }
    Log.debug('Configuration', window.rellConfig);

    // Subscribe to FB SDK events
    FB.Event.subscribe('fb.log', Log.info.bind('fb.log'));
    FB.Event.subscribe('auth.login', function(response) {
      Log.info('auth.login event', response);
    });
    FB.Event.subscribe('auth.statusChange', Rell.onStatusChange);

    if (!window.rellConfig || !window.rellConfig.init) return;

    var options = {
      appId: window.rellConfig.appID,
      version: window.rellConfig.version,
      cookie: true,
      xfbml: true,
      status: window.rellConfig.status,
      frictionlessRequests: window.rellConfig.frictionlessRequests
    };

    // NOTE: Do NOT strip the URL hash before FB.init(). On mobile Safari,
    // FB.login() uses a full-page redirect and the access token is returned
    // in the hash fragment. The SDK must see it during init.
    FB.init(options);
    if (top !== self) FB.Canvas.setAutoGrow(true);

    if (window.location.hash && window.location.hash.indexOf('access_token') !== -1) {
      // Mobile redirect return — force fresh status check, then clean hash
      FB.getLoginStatus(function(response) {
        Rell.onStatusChange(response);
        history.replaceState(null, '', window.location.pathname + window.location.search);
        Rell.autoRunCode();
      }, true); // true = force roundtrip, don't use cached status
    } else {
      if (window.location.hash) {
        history.replaceState(null, '', window.location.pathname + window.location.search);
      }
      if (!window.rellConfig.status) {
        Rell.autoRunCode();
      } else {
        FB.getLoginStatus(function() { Rell.autoRunCode(); });
        FB.getLoginStatus(Rell.onStatusChange);
      }
    }

    Rell.updateStatusBar();
  },

  /**
   * Bind a click handler to an element by ID, preventing default.
   * @param {string} id - Element ID
   * @param {Function} handler - Click handler
   */
  bindClick: function(id, handler) {
    var el = document.getElementById(id);
    if (el) {
      el.addEventListener('click', function(e) {
        e.preventDefault();
        handler();
      });
    }
  },

  /**
   * Initialize CodeMirror on the #jscode textarea.
   * Configures htmlmixed mode, line numbers, bracket matching, code folding,
   * and keyboard shortcuts.
   */
  initEditor: function() {
    var textarea = document.getElementById('jscode');
    if (!textarea || typeof CodeMirror === 'undefined') return;

    Rell.editor = CodeMirror.fromTextArea(textarea, {
      mode: 'htmlmixed',
      lineNumbers: true,
      matchBrackets: true,
      autoCloseBrackets: true,
      foldGutter: true,
      gutters: ['CodeMirror-linenumbers', 'CodeMirror-foldgutter'],
      indentUnit: 2,
      tabSize: 2,
      indentWithTabs: false,
      lineWrapping: true,
      viewportMargin: window.innerWidth < 768 ? 10 : Infinity,
      extraKeys: {
        'Cmd-Enter': function() { Rell.runCode(); },
        'Ctrl-Enter': function() { Rell.runCode(); },
        'Ctrl-/': 'toggleComment',
        'Cmd-/': 'toggleComment'
      }
    });

    // Expose globally for theme and resize modules
    window.rellEditor = Rell.editor;
  },

  /**
   * Handle FB auth status changes. Updates the auth badge in the toolbar.
   * @param {Object} response - FB auth status response
   */
  _authStatus: 'unknown',

  onStatusChange: function(response) {
    var status = response.status;
    Rell._authStatus = status;
    var el = document.getElementById('auth-status');
    if (el) {
      el.className = 'auth-badge auth-' + status;
      el.textContent = status;
    }
    // Update custom login button
    Rell.updateCustomLoginButtons(status);
  },

  updateCustomLoginButtons: function(status) {
    var isLoggedIn = status === 'connected';
    var btn = document.getElementById('fb-login-custom');
    if (btn) {
      btn.textContent = isLoggedIn ? 'Log out' : 'Log in with Facebook';
      btn.classList.toggle('logged-in', isLoggedIn);
    }
  },

  loginToggle: function() {
    if (typeof FB === 'undefined') { Log.error('FB SDK not loaded'); return; }
    if (Rell._authStatus === 'connected') {
      FB.logout(function(response) {
        Log.debug('FB.logout callback', response);
        Rell.onStatusChange(response);
      });
    } else {
      FB.login(function(response) {
        Log.debug('FB.login callback', response);
        // On mobile Safari, the cookie-based auth.statusChange event
        // may not fire due to ITP blocking third-party cookies.
        // The callback still receives the auth data.
        if (response.authResponse) {
          Rell.onStatusChange(response);
        }
      });
    }
  },

  /**
   * Run code automatically if autoRun is enabled.
   */
  autoRunCode: function() {
    if (window.rellConfig.autoRun) Rell.runCode();
  },

  /**
   * Execute the code from the editor in the output pane (#jsroot).
   * Updates the run button state: running -> success/error with visual feedback.
   */
  runCode: function() {
    var runBtn = document.getElementById('rell-run-code');
    var root = document.getElementById('jsroot');
    if (!root) return;

    // Clear empty state placeholder
    var emptyState = root.querySelector('.empty-state');
    if (emptyState) emptyState.remove();

    // Run button state: running
    if (runBtn) {
      runBtn.classList.add('running');
      runBtn.classList.remove('success', 'error');
    }

    Log.info('Executed example');

    try {
      ScriptSoup.set(root, Rell.getCode());
      if (typeof FB !== 'undefined' && FB.XFBML) {
        FB.XFBML.parse(root);
      }

      // Success flash
      if (runBtn) {
        runBtn.classList.remove('running');
        runBtn.classList.add('success');
        setTimeout(function() { runBtn.classList.remove('success'); }, 1000);
      }
    } catch (e) {
      Log.error('Execution error: ' + e.message, e);
      if (runBtn) {
        runBtn.classList.remove('running');
        runBtn.classList.add('error');
        setTimeout(function() { runBtn.classList.remove('error'); }, 1500);
      }
    }
  },

  /**
   * Get the current code from the editor (CodeMirror or textarea fallback).
   * @returns {string} The code string
   */
  getCode: function() {
    if (Rell.editor) return Rell.editor.getValue();
    var textarea = document.getElementById('jscode');
    return textarea ? textarea.value : '';
  },

  /**
   * Trigger FB login dialog.
   */
  login: function() {
    FB.login(function(response) {
      Log.debug('FB.login callback', response);
      if (response.authResponse) {
        Rell.onStatusChange(response);
      }
    });
  },

  /**
   * Log the user out of Facebook.
   */
  logout: function() {
    FB.logout(function(response) {
      Log.debug('FB.logout callback', response);
      Rell.onStatusChange(response);
    });
  },

  /**
   * Revoke the app's authorization (disconnect).
   */
  disconnect: function() {
    if (typeof FB === 'undefined') { Log.error('FB SDK not loaded'); return; }
    FB.api('/me/permissions', 'DELETE', Log.debug.bind('revokeAuthorization callback'));
  },

  /**
   * Update the status bar with SDK version and app ID.
   */
  updateStatusBar: function() {
    var sdkStatus = document.getElementById('sdk-status');
    if (sdkStatus) {
      sdkStatus.innerHTML = '<span class="status-dot status-dot-success"></span> SDK Ready';
    }
    var versionEl = document.getElementById('sdk-version');
    if (versionEl && window.rellConfig.version) {
      versionEl.textContent = window.rellConfig.version;
    }
    var appIdEl = document.getElementById('app-id-display');
    if (appIdEl && window.rellConfig.appID) {
      appIdEl.textContent = 'App: ' + window.rellConfig.appID;
    }
  },

  /**
   * Initialize the settings drawer: open/close, URL preview, and update button.
   */
  initSettings: function() {
    var overlay = document.getElementById('settings-overlay');
    var drawer = document.getElementById('settings-drawer');
    var toggleBtn = document.getElementById('settings-toggle');
    var closeBtn = document.getElementById('settings-close');

    function openSettings() {
      if (drawer) drawer.classList.add('open');
      if (overlay) overlay.classList.add('open');
    }
    function closeSettings() {
      if (drawer) drawer.classList.remove('open');
      if (overlay) overlay.classList.remove('open');
    }

    if (toggleBtn) toggleBtn.addEventListener('click', openSettings);
    if (closeBtn) closeBtn.addEventListener('click', closeSettings);
    if (overlay) overlay.addEventListener('click', closeSettings);

    // Escape key closes settings
    document.addEventListener('keydown', function(e) {
      if (e.key === 'Escape') closeSettings();
    });

    // Default setting values - used to determine which params to include in URL
    var defaults = {
      appid: '342526215814610',
      server: '',
      version: 'v25.0',
      locale: 'en_US',
      level: 'debug',
      'view-mode': 'website',
      init: 'true',
      status: 'true',
      frictionlessRequests: 'true',
      customLogin: 'true'
    };

    /**
     * Build a URL from current settings, only including non-default values.
     * @returns {string} URL path with query string
     */
    function buildSettingsUrl() {
      var params = [];
      var settings = document.querySelectorAll('.rell-setting');
      settings.forEach(function(el) {
        var name = el.name;
        var val;
        if (el.type === 'checkbox') {
          val = el.checked ? 'true' : 'false';
        } else {
          val = el.value;
        }
        if (val !== '' && val !== defaults[name]) {
          params.push(encodeURIComponent(name) + '=' + encodeURIComponent(val));
        }
      });
      var path = window.location.pathname;
      return params.length ? path + '?' + params.join('&') : path;
    }

    /**
     * Update the URL preview text.
     */
    function updatePreview() {
      var preview = document.getElementById('rell-url-preview');
      if (preview) preview.textContent = buildSettingsUrl();
    }

    // Watch for setting changes to update the URL preview
    document.querySelectorAll('.rell-setting').forEach(function(el) {
      el.addEventListener('change', updatePreview);
      el.addEventListener('input', updatePreview);
    });

    // Update button navigates to the built URL
    var updateBtn = document.getElementById('rell-settings-update');
    if (updateBtn) {
      updateBtn.addEventListener('click', function() {
        window.location = buildSettingsUrl();
      });
    }
  },

  /**
   * Initialize mobile tab switching between Editor, Output, and Log panels.
   * Only active on viewports ≤768px.
   */
  initMobileTabs: function() {
    var tabBar = document.querySelector('.mobile-tab-bar');
    if (!tabBar) return;

    var tabs = tabBar.querySelectorAll('.mobile-tab');
    var editorColumn = document.querySelector('.editor-column');
    var editorToolbar = document.querySelector('.editor-toolbar');
    var editorPane = document.getElementById('editor-pane');
    var outputPane = document.querySelector('.output-pane');
    var logColumn = document.querySelector('.log-column');

    function isMobile() {
      return window.innerWidth < 768;
    }

    function switchTab(tabName) {
      if (!isMobile()) return;

      // Update active tab
      tabs.forEach(function(t) {
        t.classList.toggle('active', t.getAttribute('data-tab') === tabName);
      });

      if (tabName === 'editor') {
        // Show editor toolbar + editor pane, hide output + log
        if (editorColumn) editorColumn.classList.remove('mobile-hidden');
        if (editorToolbar) editorToolbar.style.display = '';
        if (editorPane) editorPane.style.display = '';
        if (outputPane) outputPane.style.display = 'none';
        if (logColumn) {
          logColumn.classList.add('mobile-hidden');
          logColumn.classList.remove('mobile-full');
        }
      } else if (tabName === 'output') {
        // Show output pane full height, hide editor pane + toolbar + log
        if (editorColumn) editorColumn.classList.remove('mobile-hidden');
        if (editorToolbar) editorToolbar.style.display = 'none';
        if (editorPane) editorPane.style.display = 'none';
        if (outputPane) outputPane.style.display = '';
        if (logColumn) {
          logColumn.classList.add('mobile-hidden');
          logColumn.classList.remove('mobile-full');
        }
        // Re-run the code so XFBML plugins render at correct dimensions —
        // they were originally parsed while the output pane was hidden.
        Rell.runCode();
      } else if (tabName === 'log') {
        // Show log column full height, hide editor column
        if (editorColumn) editorColumn.classList.add('mobile-hidden');
        if (logColumn) {
          logColumn.classList.remove('mobile-hidden');
          logColumn.classList.add('mobile-full');
        }
      }
    }

    tabs.forEach(function(tab) {
      tab.addEventListener('click', function() {
        switchTab(this.getAttribute('data-tab'));
      });
    });

    // Apply initial tab state on mobile
    if (isMobile()) {
      switchTab('editor');
    }

    // Reset visibility when resizing back to desktop
    window.addEventListener('resize', function() {
      if (!isMobile()) {
        if (editorColumn) editorColumn.classList.remove('mobile-hidden');
        if (logColumn) {
          logColumn.classList.remove('mobile-hidden', 'mobile-full');
        }
        if (editorToolbar) editorToolbar.style.display = '';
        if (editorPane) editorPane.style.display = '';
        if (outputPane) outputPane.style.display = '';
      }
    });
  },

  /**
   * Clear the log panel. Bound to #rell-log-clear.
   * @returns {boolean} false (for legacy onclick compatibility)
   */
  clearLog: function() {
    Log.clear();
    return false;
  }
};

// Initialize UI immediately (no SDK dependency)
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', Rell.initUI);
} else {
  Rell.initUI();
}

// FB SDK calls this when loaded
window.fbAsyncInit = Rell.initSDK;

// If the SDK already loaded before this script ran, call initSDK now
if (typeof FB !== 'undefined') {
  Rell.initSDK();
}
