/**
 * Implements event delegation for most common events.
 *
 * Listeners are bound to a given selector and event type. Selectors
 * are very simple, and only support tag, id and className. The
 * handler will be invoked with the matching node as the "this"
 * object. In addition, the Event object will be passed as the only
 * argument to the function.
 *
 * Since basic selectors are supported, delegation is implictly
 * performed at the document body level and there is no need to
 * specify a element as the delegator.
 *
 *
 * @author Naitik Shah <n@daaku.org>
 * @link http://github.com/nshah/js-delegator
 */
var Delegator = {
  /**
   * Map's event types to subscribers lists.
   *
   * @access private
   */
  subscribers: {},

  /**
   * Bind an event handler to the given selectors/type.
   *
   * The selector can only contain tagName, id or className based
   * rules. Comma separated multiple rules are also supported. Here's
   * a complex example making use of all supported features:
   *
   *   #id1 div.classOne, div.classTwo .classThree
   *
   * This should work with click, mousedown, mouseup, mousemove,
   * mouseover, mouseout, keydown, keypress, keyup, blur, focus.
   *
   * @access public
   * @param selector  {String}   CSS selector
   * @param type      {String}   the event type
   * @param handler   {Function} the event handler
   */
  listen: function(selector, type, handler) {
    // for IE focus/blur support
    if (document.attachEvent) {
      if (type == 'focus') {
        type = 'focusin';
      } else if (type == 'blur') {
        type = 'focusout';
      }
    }

    // setup the listener if needed
    if (!(type in Delegator.subscribers)) {
      document.addEventListener
        ? document.addEventListener(type, Delegator.handler(type), true)
        : document.attachEvent('on' + type, Delegator.handler(type));
      Delegator.subscribers[type] = [];
    }

    // flatten potential multiple rules
    var rules = selector.split(/\s*,\s*/);
    for (var i=0, l=rules.length; i<l; i++) {
      Delegator.subscribers[type].push({ rule: rules[i], handler: handler });
    }
  },

  /**
   * Generate a handler for a given event type.
   *
   * @access private
   * @param type {String} the event type
   * @returns {Function}
   */
  handler: function(type) {
    return function(event) {
      Delegator.dispatch(type, event || window.event);
    };
  },

  /**
   * Dispatch a given event.
   *
   * @access private
   * @param type  {String} the event type
   * @param event {Event}  the event object
   */
  dispatch: function(type, event) {
    var
      node        = event.target || event.srcElement,
      subscribers = Delegator.subscribers[type],
      num         = subscribers.length,
      machine     = [];

    while (node) {
      // a permission error can be thrown here. we silently ignore it
      try {
        var domData = {
          id        : node.id,
          className : node.className,
          tagName   : node.tagName
        };
      } catch(e) { return; }

      for (var i=0; i<num; i++) {
        // load and compile the subscriber rule if necessary
        var sub = subscribers[i];
        if (!sub.compiled) {
          sub.compiled = Delegator.compile(sub.rule);
        }

        // default the current position in the rules
        // start at the last entry in the compiled list since we are working
        // upwards from the given element
        var state = typeof machine[i] === 'undefined'
          ? machine[i] = { index: (sub.compiled.length - 1) }
          : machine[i];

        // if the state is less than 0, we're already done with this subscriber
        if (state.index < 0) {
          continue;
        }

        // based on the state, this is the expected rule
        var rule = sub.compiled[state.index];

        // check if the expected rule matches the current node
        if ((!rule.id || rule.id == domData.id) &&
            (!rule.tagName || rule.tagName == domData.tagName) &&
            (
              rule.className.length === 0 ||
              Delegator.matchClasses(domData.className, rule.className))) {

          // we just consumed a rule entry
          --state.index;

          // set the node if this is the first match, it becomes the 'this'
          // variable for the handler
          if (!state.node) {
            state.node = node;
          }

          // complete match, this handler is a match and good to go
          if (state.index == -1) {
            sub.handler.call(state.node, event);
          }
        }
      }

      node = node.parentNode;
    }
  },

  /**
   * Cache regexp's used to match classes.
   *
   * @access private
   */
  regexpCache: {},
  /**
   * Checks if all the expected classes are in the given class string.
   *
   * @access private
   * @param givenString  {String} the className value from a node
   * @param expectedList {Array}  the expected list of classes
   * @returns {Boolean} true if all expected classes are present
   */
  matchClasses: function(givenString, expectedList) {
    for (var i=0, l=expectedList.length; i<l; i++) {
      var str = '(?:^|\\s+)' + expectedList[i] + '(?:\\s+|$)';
      if (!Delegator.regexpCache[str]) {
        Delegator.regexpCache[str] = new RegExp(str);
      }
      if (!Delegator.regexpCache[str].test(givenString)) {
        return false;
      }
    }
    return true;
  },

  /**
   * Takes something like:
   *   span#myId.myClass.myOtherClass
   * And returns:
   *   { tagName: 'span', id: 'myId', className: ['myClass', 'myOtherClass'] }
   *
   * @access private
   * @param s {String} a CSS selector
   * @returns {Object} described above
   */
  compile: function(s) {
    var
      parts = Delegator.split(s),
      rules = [],
      rule;

    for (var i=0, l=parts.length; i<l; i++) {
      var part = parts[i];

      if (!rule) {
        rule  = {
          tagName   : null,
          id        : null,
          className : []
        };
      }

      switch (part) {
        case '':
          break;
        case '#':
          rule.id = parts[++i];
          break;
        case '.':
          rule.className.push(parts[++i]);
          break;
        case ' ':
          rules.push(rule);
          rule = null;
          break;
        default:
          rule.tagName = part.toUpperCase();
      }
    }

    // last rule
    if (rule) {
      rules.push(rule);
    }

    return rules;
  },

  /**
   * A load time test to determine if capture groups work as expected.
   *
   * @access private
   */
  _compliantCG: /()??/.exec("")[1] === undefined,
  /**
   * Split a CSS selector into its parts. This function exists because
   * RegExp splits do not work the same in all browsers.
   *
   * @access private
   * @param selector {String} the CSS selector
   * @returns {Array} the components of the selector
   */
  split: function(selector) {
    // all credit goes to:
    //   http://stevenlevithan.com/assets/misc/split.js
    //
    // this function exists because the line below doesnt work in IE as
    // expected. sigh.
    //return selector.split(/(#|\.| )/g);

    var
      separator     = /(#|\.| )/g,
      cg_separator  = /^(#|\.| )$(?!\s)/g,
      output        = [],
      lastLastIndex = 0,
      match, lastIndex, lastLength;

    // NOTE: assignment and condition
    while ((match = separator.exec(selector))) {
      // 'separator.lastIndex' is not reliable cross-browser
      lastIndex = match.index + match[0].length;

      if (lastIndex > lastLastIndex) {
        output.push(selector.slice(lastLastIndex, match.index));

        // fix browsers whose 'exec' methods don't consistently return
        // 'undefined' for nonparticipating capturing groups
        if (!Delegator._compliantCG && match.length > 1) {
          match[0].replace(cg_separator, function () {
            for (var i = 1; i < arguments.length - 2; i++) {
              if (arguments[i] === undefined) {
                match[i] = undefined;
              }
            }
          });
        }

        if (match.length > 1 && match.index < selector.length) {
          Array.prototype.push.apply(output, match.slice(1));
        }

        lastLength = match[0].length;
        lastLastIndex = lastIndex;
      }

      if (separator.lastIndex === match.index) {
        separator.lastIndex++; // avoid an infinite loop
      }
    }

    if (lastLastIndex === selector.length) {
      if (!separator.test("") || lastLength) {
        output.push("");
      }
    } else {
      output.push(selector.slice(lastLastIndex));
    }

    return output;
  }
};
