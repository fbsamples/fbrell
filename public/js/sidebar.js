/**
 * Sidebar - Example navigation with collapsible categories, search, and mobile toggle.
 *
 * The sidebar HTML is rendered server-side. This script adds interactive behavior:
 *   - Collapsible category sections
 *   - Live search/filter of examples
 *   - Active-state highlighting based on current URL
 *   - Mobile sidebar toggle with backdrop overlay
 *   - Persisted collapsed state via localStorage
 *   - Swipe gesture to open/close sidebar on mobile
 *
 * DOM contract:
 *   #sidebar                    - sidebar container
 *   #sidebar-search             - text input for filtering
 *   #sidebar-toggle             - mobile hamburger button (in header)
 *   #sidebar-overlay            - backdrop overlay for mobile sidebar
 *   .sidebar-category           - each category section
 *   .sidebar-category-header    - clickable header to collapse/expand
 *   .sidebar-category-name      - text name of the category (for persistence)
 *   .sidebar-category-items     - list of items in the category
 *   .sidebar-item               - individual example link (anchor tag with href)
 *   .sidebar-toggle             - expand/collapse chevron inside header
 */
var Sidebar = {
  /**
   * Initialize all sidebar interactions.
   */
  init: function() {
    var sidebar = document.getElementById('sidebar');
    if (!sidebar) return;

    // Category toggle - click header to collapse/expand
    var headers = sidebar.querySelectorAll('.sidebar-category-header');
    headers.forEach(function(header) {
      header.addEventListener('click', function() {
        var category = this.closest('.sidebar-category');
        category.classList.toggle('collapsed');
        var toggle = this.querySelector('.sidebar-toggle');
        if (toggle) {
          toggle.textContent = category.classList.contains('collapsed') ? '\u25B6' : '\u25BC';
        }
        Sidebar.saveState();
      });
    });

    // Search/filter
    var search = document.getElementById('sidebar-search');
    if (search) {
      search.addEventListener('input', function() {
        var query = this.value.toLowerCase();
        var items = sidebar.querySelectorAll('.sidebar-item');
        var categories = sidebar.querySelectorAll('.sidebar-category');

        items.forEach(function(item) {
          var match = !query || item.textContent.toLowerCase().indexOf(query) !== -1;
          item.style.display = match ? '' : 'none';
        });

        // Hide categories with no visible items
        categories.forEach(function(cat) {
          var allItems = cat.querySelectorAll('.sidebar-item');
          var hasVisible = false;
          allItems.forEach(function(item) { if (item.style.display !== 'none') hasVisible = true; });
          cat.style.display = hasVisible || !query ? '' : 'none';
          // Expand categories when searching so results are visible
          if (query) cat.classList.remove('collapsed');
        });
      });
    }

    // Highlight the active example based on current path
    var current = window.location.pathname;
    var activeLink = sidebar.querySelector('.sidebar-item[href="' + current + '"]');
    if (activeLink) {
      activeLink.classList.add('active');
      // Expand the parent category so the active item is visible
      var parentCategory = activeLink.closest('.sidebar-category');
      if (parentCategory) parentCategory.classList.remove('collapsed');
    }

    // Mobile toggle with overlay
    var toggle = document.getElementById('sidebar-toggle');
    var overlay = document.getElementById('sidebar-overlay');

    if (toggle) {
      toggle.addEventListener('click', function() {
        sidebar.classList.toggle('open');
        if (overlay) overlay.classList.toggle('open');
      });
    }

    // Close sidebar when overlay is tapped
    if (overlay) {
      overlay.addEventListener('click', function() {
        sidebar.classList.remove('open');
        overlay.classList.remove('open');
      });
    }

    // Swipe gesture to open/close sidebar on mobile (≤1024px)
    Sidebar._initSwipe(sidebar, overlay);

    // Restore previously collapsed state
    Sidebar.restoreState();
  },

  /**
   * Initialize swipe gesture for sidebar open/close.
   * Swipe right from left edge opens; swipe left while open closes.
   */
  _initSwipe: function(sidebar, overlay) {
    var touchStartX = 0;
    var touchStartY = 0;

    document.addEventListener('touchstart', function(e) {
      if (window.innerWidth > 1024) return;
      touchStartX = e.touches[0].clientX;
      touchStartY = e.touches[0].clientY;
    }, { passive: true });

    document.addEventListener('touchend', function(e) {
      if (window.innerWidth > 1024) return;

      var touchEndX = e.changedTouches[0].clientX;
      var touchEndY = e.changedTouches[0].clientY;
      var deltaX = touchEndX - touchStartX;
      var deltaY = touchEndY - touchStartY;

      // Ignore if vertical swipe is dominant
      if (Math.abs(deltaY) > Math.abs(deltaX)) return;

      // Swipe right from left edge to open
      if (touchStartX < 30 && deltaX > 60 && !sidebar.classList.contains('open')) {
        sidebar.classList.add('open');
        if (overlay) overlay.classList.add('open');
      }

      // Swipe left to close
      if (deltaX < -60 && sidebar.classList.contains('open')) {
        sidebar.classList.remove('open');
        if (overlay) overlay.classList.remove('open');
      }
    }, { passive: true });
  },

  /**
   * Save the list of collapsed category names to localStorage.
   */
  saveState: function() {
    var collapsed = [];
    document.querySelectorAll('.sidebar-category.collapsed').forEach(function(cat) {
      var name = cat.querySelector('.sidebar-category-name');
      if (name) collapsed.push(name.textContent);
    });
    localStorage.setItem('fbrell-sidebar-collapsed', JSON.stringify(collapsed));
  },

  /**
   * Restore collapsed state from localStorage.
   */
  restoreState: function() {
    try {
      var collapsed = JSON.parse(localStorage.getItem('fbrell-sidebar-collapsed') || '[]');
      document.querySelectorAll('.sidebar-category').forEach(function(cat) {
        var name = cat.querySelector('.sidebar-category-name');
        if (name && collapsed.indexOf(name.textContent) !== -1) {
          cat.classList.add('collapsed');
          var toggle = cat.querySelector('.sidebar-toggle');
          if (toggle) toggle.textContent = '\u25B6';
        }
      });
    } catch (e) {
      // Silently ignore corrupted localStorage data
    }
  }
};
