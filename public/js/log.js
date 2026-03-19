/**
 * Log - Color-coded, filterable log panel for FB API responses and app events.
 *
 * Features:
 *   - Color-coded entries by level (error=red, info=blue, debug=green)
 *   - Level badges ([ERROR], [INFO], [DEBUG])
 *   - Filter tabs (All, Errors, Info, Debug)
 *   - Relative timestamps that auto-update every 30 seconds
 *   - JSON pretty-printing with syntax highlighting
 *   - Copy single entry or all entries to clipboard
 *   - Clear functionality
 *   - Error count badge
 *   - Auto-scroll with scroll-lock detection
 *   - Integration with ErrorExplanations for FB API errors
 *
 * DOM contract:
 *   #log                          - log entries container (.log-entries)
 *   #rell-log-clear               - clear button
 *   #rell-log-copy                - copy all button
 *   #log-error-count              - error count badge
 *   .log-filter[data-filter]      - filter buttons (all, error, info, debug)
 */
var Log = (function() {

  // Internal state
  var _root = null;       // The #log container element
  var _level = 'debug';   // Minimum level to display
  var _entries = [];       // All log entry data for copy-all
  var _errorCount = 0;     // Running error count

  // Level priority for filtering
  var _levels = { error: 0, info: 1, debug: 2 };

  /**
   * Format a relative time string from a timestamp.
   * @param {number} timestamp - Unix timestamp in milliseconds
   * @returns {string} Human-readable relative time
   */
  function relativeTime(timestamp) {
    var diff = Math.floor((Date.now() - timestamp) / 1000);
    if (diff < 5) return 'just now';
    if (diff < 60) return diff + 's ago';
    if (diff < 3600) return Math.floor(diff / 60) + 'm ago';
    if (diff < 86400) return Math.floor(diff / 3600) + 'h ago';
    return Math.floor(diff / 86400) + 'd ago';
  }

  /**
   * Apply syntax highlighting to a JSON string.
   * Wraps keys, strings, numbers, booleans, and null in spans.
   * @param {string} json - Pre-formatted JSON string
   * @returns {string} HTML with syntax-highlighted spans
   */
  function highlightJson(json) {
    return json
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/("(\\u[a-fA-F0-9]{4}|\\[^u]|[^\\"])*"(\s*:)?)/g, function(match) {
        var cls = 'json-string';
        if (/:$/.test(match)) {
          cls = 'json-key';
        }
        return '<span class="' + cls + '">' + match + '</span>';
      })
      .replace(/\b(true|false)\b/g, '<span class="json-boolean">$1</span>')
      .replace(/\bnull\b/g, '<span class="json-null">null</span>')
      .replace(/\b(-?\d+\.?\d*([eE][+-]?\d+)?)\b/g, '<span class="json-number">$1</span>');
  }

  /**
   * Format arguments for display. Objects become pretty JSON; primitives become strings.
   * @param {Array} args - Arguments to format
   * @returns {string} HTML representation
   */
  function formatArgs(args) {
    var parts = [];
    for (var i = 0; i < args.length; i++) {
      var arg = args[i];
      if (arg === undefined) continue;
      if (arg === null) {
        parts.push('<pre class="log-json"><span class="json-null">null</span></pre>');
      } else if (typeof arg === 'object') {
        try {
          var json = JSON.stringify(arg, null, 2);
          parts.push('<pre class="log-json">' + highlightJson(json) + '</pre>');
        } catch (e) {
          parts.push('<pre class="log-json">' + String(arg) + '</pre>');
        }
      } else {
        parts.push('<pre class="log-json">' + String(arg) + '</pre>');
      }
    }
    return parts.join('');
  }

  /**
   * Check whether the log container is scrolled to the bottom (within 50px tolerance).
   * @returns {boolean}
   */
  function isScrolledToBottom() {
    if (!_root) return true;
    return _root.scrollHeight - _root.scrollTop - _root.clientHeight < 50;
  }

  /**
   * Scroll the log container to the bottom.
   */
  function scrollToBottom() {
    if (_root) {
      _root.scrollTop = _root.scrollHeight;
    }
  }

  /**
   * Update the error count badge.
   */
  function updateErrorBadge() {
    var badge = document.getElementById('log-error-count');
    if (badge) {
      badge.textContent = _errorCount;
      badge.style.display = _errorCount > 0 ? '' : 'none';
    }
  }

  /**
   * Create and append a log entry to the DOM.
   * @param {string} level - "error", "info", or "debug"
   * @param {string} header - Title text for the entry
   * @param {Array} args - Additional data to display in the body
   */
  function addEntry(level, header, args) {
    if (!_root) return;
    if (_levels[level] > _levels[_level]) return;

    var now = Date.now();
    var shouldScroll = isScrolledToBottom();

    // Track for copy-all
    _entries.push({ level: level, header: header, args: args, time: now });

    // Update error count
    if (level === 'error') {
      _errorCount++;
      updateErrorBadge();
    }

    // Build entry element
    var entry = document.createElement('div');
    entry.className = 'log-entry log-' + level;
    entry.setAttribute('data-level', level);
    entry.setAttribute('data-time', now);

    // Header row
    var entryHeader = document.createElement('div');
    entryHeader.className = 'log-entry-header';

    var badge = document.createElement('span');
    badge.className = 'log-level-badge log-badge-' + level;
    badge.textContent = level.toUpperCase();

    var title = document.createElement('span');
    title.className = 'log-entry-title';
    title.textContent = header;

    var time = document.createElement('span');
    time.className = 'log-entry-time';
    time.textContent = relativeTime(now);

    var toggle = document.createElement('span');
    toggle.className = 'log-toggle';
    toggle.textContent = '\u25B6';

    entryHeader.appendChild(badge);
    entryHeader.appendChild(title);
    entryHeader.appendChild(time);
    entryHeader.appendChild(toggle);

    // Body
    var body = document.createElement('div');
    body.className = 'log-entry-body';

    // Check for error explanations on error-level entries
    if (level === 'error' && args.length > 0 && typeof ErrorExplanations !== 'undefined') {
      for (var i = 0; i < args.length; i++) {
        var explanation = ErrorExplanations.lookup(args[i]);
        if (explanation) {
          var explDiv = document.createElement('div');
          explDiv.className = 'error-explanation';
          explDiv.innerHTML =
            '<div class="error-explanation-message">' + explanation.message + '</div>' +
            '<div class="error-explanation-suggestion">' + explanation.suggestion + '</div>' +
            '<a class="error-explanation-link" href="' + explanation.docs + '" target="_blank">View docs \u2192</a>';
          body.appendChild(explDiv);
          break;
        }
      }
    }

    // Formatted arguments
    if (args.length > 0) {
      var argsHtml = formatArgs(args);
      if (argsHtml) {
        var argsContainer = document.createElement('div');
        argsContainer.innerHTML = argsHtml;
        while (argsContainer.firstChild) {
          body.appendChild(argsContainer.firstChild);
        }
      }
    }

    entry.appendChild(entryHeader);
    entry.appendChild(body);
    _root.appendChild(entry);

    // Auto-scroll if user was at bottom
    if (shouldScroll) scrollToBottom();
  }

  /**
   * Create a logging function for a given level.
   * The returned function also has a .bind(title) method that returns a callback
   * suitable for FB SDK event handlers.
   * @param {string} level
   * @returns {Function}
   */
  function createLevelFn(level) {
    var fn = function(header) {
      var args = [];
      for (var i = 1; i < arguments.length; i++) {
        args.push(arguments[i]);
      }
      addEntry(level, header, args);
    };

    /**
     * Returns a callback function that, when called, logs with the given title.
     * Usage: FB.Event.subscribe('fb.log', Log.info.bind('fb.log'));
     * @param {string} title
     * @returns {Function}
     */
    fn.bind = function(title) {
      return function() {
        var args = [];
        for (var i = 0; i < arguments.length; i++) {
          args.push(arguments[i]);
        }
        addEntry(level, title, args);
      };
    };

    return fn;
  }

  // Update relative timestamps every 30 seconds
  setInterval(function() {
    if (!_root) return;
    var timeEls = _root.querySelectorAll('.log-entry-time');
    timeEls.forEach(function(el) {
      var entry = el.closest('.log-entry');
      if (entry) {
        var t = parseInt(entry.getAttribute('data-time'), 10);
        if (t) el.textContent = relativeTime(t);
      }
    });
  }, 30000);

  // Public API
  return {
    /**
     * Initialize the log system.
     * @param {HTMLElement} root - The #log container element
     * @param {string} levelName - Minimum log level to display ("error", "info", "debug")
     */
    init: function(root, levelName) {
      _root = root;
      if (levelName && _levels[levelName] !== undefined) {
        _level = levelName;
      }

      if (!_root) return;

      // Event delegation for toggle and copy on entries
      _root.addEventListener('click', function(e) {
        // Toggle entry body visibility
        var header = e.target.closest('.log-entry-header');
        if (header) {
          var entryEl = header.closest('.log-entry');
          if (entryEl) {
            var bodyDiv = entryEl.querySelector('.log-entry-body');
            var toggleSpan = header.querySelector('.log-toggle');
            if (bodyDiv) {
              var isOpen = bodyDiv.classList.contains('open');
              bodyDiv.classList.toggle('open');
              if (toggleSpan) toggleSpan.textContent = isOpen ? '\u25B6' : '\u25BC';
            }
          }
        }
      });

      // Copy all button
      var copyAllBtn = document.getElementById('rell-log-copy');
      if (copyAllBtn) {
        copyAllBtn.addEventListener('click', function() {
          var text = _entries.map(function(e) {
            var argStr = e.args.map(function(a) {
              if (a && typeof a === 'object') {
                try { return JSON.stringify(a, null, 2); } catch (ex) { return String(a); }
              }
              return String(a);
            }).join('\n');
            return '[' + e.level.toUpperCase() + '] ' + e.header + (argStr ? '\n' + argStr : '');
          }).join('\n\n');
          if (navigator.clipboard) {
            navigator.clipboard.writeText(text);
          }
        });
      }
    },

    info: createLevelFn('info'),
    error: createLevelFn('error'),
    debug: createLevelFn('debug'),

    /**
     * Clear all log entries.
     */
    clear: function() {
      _entries = [];
      _errorCount = 0;
      updateErrorBadge();
      if (_root) _root.innerHTML = '';
    }
  };
})();
