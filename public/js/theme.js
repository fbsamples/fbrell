/**
 * Theme - Dark/light theme toggle system.
 *
 * Respects OS preference via prefers-color-scheme, saves user choice to
 * localStorage, and updates the CodeMirror editor theme when toggled.
 *
 * DOM contract:
 *   #theme-toggle  - button that toggles between dark and light
 *   html[data-theme] - attribute set to "dark" or "light"
 */
var Theme = {
  /**
   * Initialize the theme system. Reads saved preference or falls back to OS setting.
   */
  init: function() {
    var saved = localStorage.getItem('fbrell-theme');
    var prefersDark = window.matchMedia('(prefers-color-scheme: dark)').matches;
    var theme = saved || (prefersDark ? 'dark' : 'light');
    Theme.set(theme);

    var toggle = document.getElementById('theme-toggle');
    if (toggle) {
      toggle.addEventListener('click', function() {
        var current = document.documentElement.getAttribute('data-theme');
        Theme.set(current === 'dark' ? 'light' : 'dark');
      });
    }

    // Listen for OS theme changes (only applies if user hasn't explicitly chosen)
    window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', function(e) {
      if (!localStorage.getItem('fbrell-theme')) {
        Theme.set(e.matches ? 'dark' : 'light');
      }
    });
  },

  /**
   * Apply a theme.
   * @param {string} theme - "dark" or "light"
   */
  set: function(theme) {
    document.documentElement.setAttribute('data-theme', theme);
    localStorage.setItem('fbrell-theme', theme);

    var toggle = document.getElementById('theme-toggle');
    if (toggle) {
      toggle.textContent = theme === 'dark' ? '\u2600\uFE0F' : '\uD83C\uDF19';
      toggle.title = theme === 'dark' ? 'Switch to light mode' : 'Switch to dark mode';
    }

    // CodeMirror theme is handled by CSS custom properties — refresh to repaint
    if (window.rellEditor) {
      window.rellEditor.refresh();
    }
  }
};
