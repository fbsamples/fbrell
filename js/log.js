var jsDump = require('jsDump')
  , Delegator = require('delegator')

var Log = {
  levels: ['error', 'info', 'debug'],
  root: null,
  count: 0,

  impl: function(level) {
    return function() {
      Log.write(level, Array.prototype.slice.apply(arguments))
    }
  },

  write: function(level, args) {
    var
      hd = args.shift(),
      bd = Log.dumpArray(args)

    Log.writeHTML(level, hd, bd)
  },

  dumpArray: function(args) {
    var bd = ''

    for (var i=0, l=args.length; i<l; i++) {
      if (bd) {
        bd += '<hr>'
      }
      bd += jsDump.parse(args[i])
    }

    return bd
  },

  writeHTML: function(level, hd, bd) {
    if (level > Log.level) {
      return
    }

    var entry = document.createElement('div')
    entry.className = 'log-entry log-' + Log.levels[level]
    entry.innerHTML = Log.genBare(hd, bd)
    Log.root.insertBefore(entry, Log.root.firstChild)
  },

  genBare: function(hd, bd) {
    return (
      '<div class="hd">' +
        '<span class="toggle">&#9658;</span> ' +
        '<span class="count">' + (++Log.count) + '</span> ' +
        hd +
      '</div>' +
      (bd ? '<div class="bd" style="display: none;">' + bd + '</div>' : '')
    )
  },

  genHTML: function(hd, bd) {
    return '<div class="log-entry">' + Log.genBare(hd, bd) + '</div>'
  },

  clear: function() {
    Log.root.innerHTML = ''
    Log.count = 0
  },

  getLevel: function(name) {
    for (var i=0, l=Log.levels.length; i<l; i++) {
      if (name == Log.levels[i]) {
        return i
      }
    }
    return l // max level
  },

  init: function(root, levelName) {
    jsDump.HTML = true
    Log.level = Log.getLevel(levelName)
    Log.root = root
    root.style.height = (
      (window.innerHeight || document.documentElement.clientHeight)
      + 'px'
    )
    for (var i=0, l=Log.levels.length; i<l; i++) {
      var name = Log.levels[i]
      Log[name] = Log.impl(i)
      Log[name].bind = function(title) {
        var self = this
        return function() {
          var args = Array.prototype.slice.apply(arguments)
          args.unshift(title)
          self.apply(null, args)
        }
      }
    }

    Delegator.listen('.log-entry .toggle', 'click', function() {
      try {
        var style = this.parentNode.nextSibling.style
        if (style.display == 'none') {
          style.display = 'block'
          this.innerHTML = '&#9660;'
        } else {
          style.display = 'none'
          this.innerHTML = '&#9658;'
        }
      } catch(e) {
        // ignore, the body is probably missing
      }
    })
  },

  flashTrace: function(title, obj) {
    Log.info(decodeURIComponent(title), decodeURIComponent(obj))
  }
}

if (typeof module !== 'undefined') module.exports = Log
