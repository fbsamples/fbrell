/**
 * Resize - Draggable panel resize handles for the editor layout.
 *
 * Supports two resize handles:
 *   #resize-v  - vertical resize between editor-pane and output-pane (within .editor-column)
 *   #resize-h  - horizontal resize between editor-column and log-column (within .main-layout)
 *
 * Stores ratios in localStorage and restores them on load.
 * Disabled on mobile (<768px) where layout is single-column.
 * Supports both mouse and touch events with pointer capture for smooth dragging.
 *
 * DOM contract:
 *   #resize-v       - vertical drag handle
 *   #resize-h       - horizontal drag handle
 *   .editor-column  - left column containing editor and output
 *   .log-column     - right column containing log panel
 *   #editor-pane    - top pane (code editor)
 *   #jsroot         - bottom pane (output/preview)
 *   .main-layout    - parent grid container
 */
var Resize = {
  /**
   * Initialize resize handles. Restores saved ratios and binds drag events.
   */
  init: function() {
    // Skip on mobile
    if (window.innerWidth < 768) return;

    Resize._initVertical();
    Resize._initHorizontal();
    Resize._restoreState();
  },

  /**
   * Set up the vertical resize handle (editor-pane / output-pane split).
   */
  _initVertical: function() {
    var handle = document.getElementById('resize-v');
    if (!handle) return;

    var editorPane = document.getElementById('editor-pane');
    var outputPane = document.getElementById('jsroot');
    if (!editorPane || !outputPane) return;

    var container = handle.parentElement;

    Resize._makeDraggable(handle, {
      onDrag: function(e, startInfo) {
        var containerRect = container.getBoundingClientRect();
        var totalHeight = containerRect.height;
        var relativeY = e.clientY - containerRect.top;

        // Clamp between 15% and 85%
        var ratio = Math.max(0.15, Math.min(0.85, relativeY / totalHeight));

        editorPane.style.flexBasis = (ratio * 100) + '%';
        outputPane.style.flexBasis = ((1 - ratio) * 100) + '%';

        return ratio;
      },
      onEnd: function(ratio) {
        localStorage.setItem('fbrell-resize-v', String(ratio));
        // Refresh CodeMirror after resize
        if (window.rellEditor) window.rellEditor.refresh();
      }
    });
  },

  /**
   * Set up the horizontal resize handle (editor-column / log-column split).
   */
  _initHorizontal: function() {
    var handle = document.getElementById('resize-h');
    if (!handle) return;

    var mainLayout = handle.closest('.main-layout');
    if (!mainLayout) return;

    var editorCol = mainLayout.querySelector('.editor-column');
    var logCol = mainLayout.querySelector('.log-column');
    if (!editorCol || !logCol) return;

    var sidebar = mainLayout.querySelector('.sidebar');

    Resize._makeDraggable(handle, {
      onDrag: function(e, startInfo) {
        var layoutRect = mainLayout.getBoundingClientRect();
        var handleWidth = 6;
        // Read sidebar width dynamically (it may be hidden/resized)
        var sidebarWidth = sidebar ? sidebar.getBoundingClientRect().width : 0;
        // Available width after sidebar and handle
        var available = layoutRect.width - sidebarWidth - handleWidth;
        // Mouse position relative to area after sidebar
        var relativeX = e.clientX - layoutRect.left - sidebarWidth;

        // Clamp editor to between 20% and 80% of available space
        var ratio = Math.max(0.20, Math.min(0.80, relativeX / available));
        var editorWidth = Math.round(ratio * available);
        var logWidth = available - editorWidth;

        mainLayout.style.gridTemplateColumns =
          sidebarWidth + 'px ' + editorWidth + 'px ' + handleWidth + 'px ' + logWidth + 'px';

        return ratio;
      },
      onEnd: function(ratio) {
        localStorage.setItem('fbrell-resize-h', String(ratio));
        if (window.rellEditor) window.rellEditor.refresh();
      }
    });
  },

  /**
   * Generic drag handler factory. Supports mouse and touch events.
   * @param {HTMLElement} handle - The drag handle element
   * @param {Object} callbacks - { onDrag(event, startInfo), onEnd(lastValue) }
   */
  _makeDraggable: function(handle, callbacks) {
    var lastValue = null;

    function onPointerDown(e) {
      // Only handle left mouse button or touch
      if (e.type === 'mousedown' && e.button !== 0) return;

      e.preventDefault();

      var startInfo = {
        startX: e.clientX || (e.touches && e.touches[0].clientX),
        startY: e.clientY || (e.touches && e.touches[0].clientY)
      };

      // Add dragging class for visual feedback
      handle.classList.add('dragging');
      document.body.style.cursor = handle.id === 'resize-h' ? 'col-resize' : 'row-resize';
      document.body.style.userSelect = 'none';

      function onPointerMove(moveEvent) {
        var ev = moveEvent;
        if (moveEvent.touches) ev = moveEvent.touches[0];
        lastValue = callbacks.onDrag(ev, startInfo);
      }

      function onPointerUp() {
        handle.classList.remove('dragging');
        document.body.style.cursor = '';
        document.body.style.userSelect = '';

        document.removeEventListener('mousemove', onPointerMove);
        document.removeEventListener('mouseup', onPointerUp);
        document.removeEventListener('touchmove', onPointerMove);
        document.removeEventListener('touchend', onPointerUp);

        if (lastValue !== null && callbacks.onEnd) {
          callbacks.onEnd(lastValue);
        }
      }

      document.addEventListener('mousemove', onPointerMove);
      document.addEventListener('mouseup', onPointerUp);
      document.addEventListener('touchmove', onPointerMove, { passive: false });
      document.addEventListener('touchend', onPointerUp);
    }

    handle.addEventListener('mousedown', onPointerDown);
    handle.addEventListener('touchstart', onPointerDown, { passive: false });
  },

  /**
   * Restore saved resize ratios from localStorage.
   */
  _restoreState: function() {
    // Restore vertical ratio
    var vRatio = parseFloat(localStorage.getItem('fbrell-resize-v'));
    if (vRatio && vRatio > 0.1 && vRatio < 0.9) {
      var editorPane = document.getElementById('editor-pane');
      var outputPane = document.getElementById('jsroot');
      if (editorPane && outputPane) {
        editorPane.style.flexBasis = (vRatio * 100) + '%';
        outputPane.style.flexBasis = ((1 - vRatio) * 100) + '%';
      }
    }

    // Restore horizontal ratio
    var hRatio = parseFloat(localStorage.getItem('fbrell-resize-h'));
    if (hRatio && hRatio > 0.2 && hRatio < 0.9) {
      var mainLayout = document.querySelector('.main-layout');
      if (mainLayout) {
        var sidebar = mainLayout.querySelector('.sidebar');
        var sidebarWidth = sidebar ? sidebar.getBoundingClientRect().width : 240;
        var handleWidth = 6;
        var available = mainLayout.getBoundingClientRect().width - sidebarWidth - handleWidth;
        var editorWidth = Math.round(hRatio * available);
        var logWidth = available - editorWidth;
        mainLayout.style.gridTemplateColumns =
          sidebarWidth + 'px ' + editorWidth + 'px ' + handleWidth + 'px ' + logWidth + 'px';
      }
    }

    // Refresh CodeMirror after restoring layout
    setTimeout(function() {
      if (window.rellEditor) window.rellEditor.refresh();
    }, 100);
  }
};
