var Tracer = {
  level: -1,
  bdCache: '',
  cache: [],

  exclude: {
    'FB.Event.subscribers': 1,
    'FB.UIServer._popupMonitor': 1,
    'FB.md5sum': 1,
    'FB.QS.decode': 1,
    'FB.QS.encode': 1,
    'FB.copy': 1,
    'FB.guid': 1,
    'FB.Canvas.setSize': 1,
    'FB.Canvas._computeContentSize': 1,
    'FB.Canvas._getBodyProp': 1,
    'FB.Dom.getStyle': 1,
    'FB.Array.forEach': 1,
    'FB.String.format': 1,
    'FB.log': 1
  },

  mixins: {
    'FB.EventProvider': 1
  },

  instrument: function(prefix, obj, instanceMethod) {
    if (prefix == 'FB.CLASSES') {
      return;
    }

    for (var name in obj) {
      if (obj.hasOwnProperty(name)) {
        var
          val = obj[name],
          fullname = prefix + '.' + name;
        if (typeof val == 'function') {
          if (instanceMethod || !val.prototype.bind) {
            obj[name] = Tracer.wrap({
              func           : val,
              instanceMethod : instanceMethod,
              name           : name,
              prefix         : prefix,
              scope          : obj
            });
          } else {
            Tracer.instrument(fullname, val.prototype, true);
          }
        } else if (typeof val == 'object') {
          Tracer.instrument(fullname, val, (fullname in Tracer.mixins));
        }
      }
    }
  },

  wrap: function(conf) {
    var name = conf.prefix + '.' + conf.name;

    // things that are excluded do not get wrapped
    if (conf.func._tracerMark || name in Tracer.exclude) {
      return conf.func;
    }

    var wrapped = function() {
      Tracer.level++;
      Tracer.lastLevel = Tracer.level;
      if (!Tracer.cache[Tracer.level]) {
        Tracer.cache[Tracer.level] = [];
      }

      var
        args = Array.prototype.slice.apply(arguments),
        argsHTML = Log.dumpArray(args),
        returnValue = conf.func.apply(conf.instanceMethod ? this : conf.scope, args);

      if (returnValue) {
        argsHTML += '<hr><h3>Return Value</h3>' + jsDump.parse(returnValue);
      } else {
        argsHTML += '<hr><h3>No Return Value</h3>';
      }

      if (Tracer.lastLevel == Tracer.level) {
        Tracer.cache[Tracer.level].push(Log.genHTML(name, argsHTML));
      } else {
        Tracer.lastLevel = Tracer.level;
        var children = Tracer.cache[Tracer.level+1] || [];
        Tracer.cache[Tracer.level+1] = [];
        Tracer.cache[Tracer.level].push(
          Log.genHTML(name, argsHTML + children.join('')));
      }

      if (Tracer.level == 0 && Tracer.cache[0][0]) {
        var entry = document.createElement('div');
        entry.className = 'log-entry log-trace';
        entry.innerHTML = Tracer.cache[0][0];
        Log.root.insertBefore(entry, Log.root.firstChild);
        Tracer.cache[0] = [];
      }

      Tracer.level--;

      return returnValue;
    };
    wrapped._tracerMark = true;
    return wrapped;
  }
};
