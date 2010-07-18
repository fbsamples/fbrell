var
  cx = /[\u0000\u00ad\u0600-\u0604\u070f\u17b4\u17b5\u200c-\u200f\u2028-\u202f\u2060-\u206f\ufeff\ufff0-\uffff]/g,
  escapable = /[\\\"\x00-\x1f\x7f-\x9f\u00ad\u0600-\u0604\u070f\u17b4\u17b5\u200c-\u200f\u2028-\u202f\u2060-\u206f\ufeff\ufff0-\uffff]/g,
  gap,
  indent,
  meta = {    // table of character substitutions
    '\b': '\\b',
    '\t': '\\t',
    '\n': '\\n',
    '\f': '\\f',
    '\r': '\\r',
    '"' : '\\"',
    '\\': '\\\\'
  };

function quote(string) {
  escapable.lastIndex = 0;
  return escapable.test(string) ?
  '"' + string.replace(escapable, function (a) {
    var c = meta[a];
    return typeof c === 'string' ? c :
    '\\u' + ('0000' + a.charCodeAt(0).toString(16)).slice(-4);
  }) + '"' :
  '"' + string + '"';
}

function quoteKey(string) {
  // does not actually quote the key when it's not needed
  return /[^A-Za-z_]./.test(string) ? quote(string) : string;
}

function str(key, holder) {
  var
    i,          // The loop counter.
    k,          // The member key.
    v,          // The member value.
    length,
    mind = gap,
    partial,
    value = holder[key];

  if (value && typeof value == 'object' && typeof value.toJSON == 'function') {
    value = value.toJSON(key);
  }

  switch (typeof value) {
    case 'string':
    return quote(value);

    case 'number':
    return isFinite(value) ? String(value) : 'null';

    case 'boolean':
    return String(value);

    case 'object':
    if (!value) {
      return 'null';
    }
    gap += indent;
    partial = [];
    if (Object.prototype.toString.apply(value) === '[object Array]') {
      length = value.length;
      for (i = 0; i < length; i += 1) {
        partial[i] = str(i, value) || 'null';
      }

      v = partial.length === 0 ? '[]' :
      gap ? '[\n' + gap +
      partial.join(',\n' + gap) + '\n' +
      mind + ']' :
      '[' + partial.join(',') + ']';
      gap = mind;
      return v;
    }

    for (k in value) {
      if (Object.hasOwnProperty.call(value, k)) {
        v = str(k, value);
        if (v) {
          partial.push(quoteKey(k) + (gap ? ': ' : ':') + v);
        }
      }
    }

    v = partial.length === 0 ? '{}' :
    gap ? '{\n' + gap + partial.join(',\n' + gap) + '\n' +
    mind + '}' : '{' + partial.join(',') + '}';
    gap = mind;
    return v;
  }
}

module.exports = function(value, initial, space) {
  var
    i;
    gap = '';
    indent = '';
    space = space === undefined ? 2 : space;
    initial = initial === undefined ? 0 : initial;

  if (typeof space === 'number') {
    for (i = 0; i < space; i += 1) {
      indent += ' ';
    }
  } else if (typeof space === 'string') {
    indent = space;
  }

  if (typeof initial === 'number') {
    for (i = initial, initial = ''; i > 0; i -= 1) {
      initial += ' ';
    }
  }

  return str('', {'': value}).replace(/^/gm, initial).trim();
};
