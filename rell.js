var
  QS     = require('querystring'),
  fs     = require('fs'),
  path   = require('path'),
  http   = require('http'),
  sys    = require('sys'),
  url    = require('url'),
  Buffer = require('buffer').Buffer;

FbOpts = {
  appId: '184484190795' // fbrell
};

DefaultConfig = {
  'appid'     : FbOpts.appId,
  'comps'     : '',
  'level'     : 'debug',
  'locale'    : 'en_US',
  'old_debug' : 6,
  'server'    : 'static.ak.connect',
  'trace'     : 1,
  'version'   : 'mu',
  'opengraph' : 'page',
  'og_url'    : 'http://fbrell.com/',
  'rte'       : 1,
  'url'       : '',
  'status'    : 1,
};

echo = function() {
  this.halt('<pre style="font-size: 15px">' + this.prettyDump() + '</pre>');
};

module.exports = require('sin/application')(__dirname)
.plug('sin/cookie')
.plug('sin/haml')
.plug('sin/jsloader')
.configure('production', function() {
  this.error(function(err) {
    sys.puts(
      'Error Report: ' +
      new Date() + "\n" +
      sys.inspect(this.request) + "\n" +
      (err.stack || err) + "\n" +
      '-------------------------------------------------------------'
    );
    this.haml('error');
  });
})
.configure(function() {
  //TODO use cb and make this a propert async configure
  this.examples = {};
  var populate = function(name, dir) {
    var E = this.examples[name] = {};
    fs.readdir(dir, function(err, categories) {
      categories.forEach(function(category) {
        fs.readdir(path.join(dir, category), function(err, examples) {
          examples.forEach(function(example) {
            var name = example.substr(0, example.length-5);
            E['/' + category + '/' + name] = {
              category : category,
              filename : path.join(dir, category, example),
              name     : name
            };
          });
        });
      });
    });
  }.bind(this);

  populate('examples', path.join(this.root, 'examples'));
  populate('examples-old', path.join(this.root, 'examples-old'));
})
.helper('makeUrl', function(path) {
  var qs = {};
  for (var key in this.config) {
    var val = this.config[key];
    if (DefaultConfig[key] != val) {
      qs[key] = val;
    }
  }
  qs = QS.stringify(qs);
  return path + (qs == '' ? '' : ('?' + qs));
})
.helper('script', function() {
  var server = this.config.server;

  var aliases = {
    sb: 'www.naitik.dev719',
    bg: 'www.brent.devrs109',
    rh: 'www.rhe.devrs106'
  };
  // alias sb to my sandbox
  if (server in aliases) {
    server = aliases[server];
  }

  var
    ssl = this.secure,
    url = 'http://' + server + '.facebook.com/';

  if (url === 'http://static.ak.connect.facebook.com/') {
    if (this.config.version === 'mu') {
      if (ssl) {
        url = 'https://connect.facebook.net/';
      } else {
        url = 'http://connect.facebook.net/';
      }
    } else if (ssl) {
      url = 'https://ssl.connect.facebook.com/';
    }
  }

  if (this.config.version === 'mu') {
    var
      special = ['snc', 'intern', 'beta', 'sandcastle', 'latest', 'dev', 'inyour'],
      comps = this.config.comps || 'all';
    if (special.some(function(s) { return server.indexOf(s) > -1;})) {
      url += 'assets.php/' + this.config.locale + '/' + comps + '.js';
    } else {
      url += this.config.locale + '/' + comps + '.js';
    }
  } else if (this.config.version === 'mid') {
    url += 'connect.php/' + this.config.locale + '/js/' + this.config.comps;
  } else {
    url += 'js/api_lib/v0.4/FeatureLoader.js.php';
  }

  return this.jsloader(
    [
      (ssl ? 'https://ssl' : 'http://www') + '.google-analytics.com/ga.js',
      '/delegator.js',
      '/jsDump-1.0.0.js',
      '/log.js',
      '/tracer.js',
      '/rell.js',
      '/codemirror/js/codemirror.js',
      url
    ],
    'Rell.init(' + JSON.stringify(this.config) + ');'
  );
})
.helper('prettyDump', function() {
  return sys.inspect({
    post: this.post,
    url: this.url,
    headers: this.headers,
    signedRequest: this.signedRequest
  });
})
.before(function() {
  this.config = {};
  [DefaultConfig, this.url.query].forEach(function(src) {
    for (var key in src) {
      this.config[key] = src[key];
    }
  }.bind(this));

  this.example_code = ''
  this.examples = this.config.version == 'mu'
    ? this.app.examples['examples']
    : this.app.examples['examples-old'];
})
.before(function() {
  var signedRequest = this.param('signed_request');
  if (signedRequest) {
    this.signedRequest = JSON.parse(
      new Buffer(
        signedRequest.split('.')[1].replace('-', '+').replace('_', '/'),
        'base64').toString('utf8'));
  }
})
.before(function(cb) {
  var urlParam = this.param('url');
  if (urlParam) {
    var
      context = this,
      urlParsed = url.parse(urlParam),
      client = http.createClient(parseInt(urlParsed.port || 80, 10), urlParsed.hostname),
      request = client.request(
        'GET',
        urlParsed.pathname + (urlParsed.search || ''),
        { host: urlParsed.hostname }
      );

    request.end();
    request.on('response', function(response) {
      var data = [];
      response.setEncoding('utf8');
      response.on('data', function(chunk) {
        data.push(chunk);
      });
      response.on('end', function() {
        context.example_code = data.join('');
        cb();
      });
    });

    return true;
  }
})
.notFound(function() {
  this.haml('not_found');
})
.helper('renderExample', function(category, name) {
  var
    title   = name + ' &middot; ' + category,
    example = this.examples['/' + category + '/' + name];
  if (!example) {
    this.pass();
    return;
  }
  if (example.code) {
    this.example_code = example.code;
    this.haml('index', { title: title });
  } else {
    fs.readFile(example.filename, this.errproof(function(data) {
      this.example_code = example.code = data;
      this.haml('index', { title: title });
    }));
  }
})
.get('/', function() {
  if (this.url.query.app_data) {
    var split = this.url.query.app_data.split('_');
    if (split.length == 3) {
      this.config.server = split.shift();
    }
    this.renderExample(split[0], split[1]);
  } else {
    this.haml('index', { title: 'Welcome' })
  }
})
.get('/help', function() {
  this.haml('help')
})
.get('/examples', function() {
  this.haml('examples')
})
.get(/^\/echo.*/, echo)
.post(/^\/echo.*/, echo)
.get('/channel', function() {
  this.halt(
    '<script src="http' + (this.secure ? 's' : '') +
    '://connect.facebook.net/en_US/all.js"></script>\n'
  );
})
.get('/test', function() {
  sys.puts(sys.inspect(this.headers));
  sys.puts(sys.inspect(this.url));
  this.haml('test')
})
.post('/', function() {
  sys.puts(sys.inspect(this.post));
  this.halt(this.render('fbml', { params: this.post }));
})
.post('/:category/:name', function(category, name) {
  this.halt(
    this.render('fbml', { params: this.post }) +
    '<h1>' +
    '<fb:prompt-permission perms="email">' +
    'TOS + Email</fb:prompt-permission>' +
    '</h1>' +

    '<h1>fb:iframe</h1>' +

    '<fb:iframe' +
    ' src="' + url.format(this.url) + '"' +
    ' frameborder=0' +
    ' scrolling=no' +
    ' style="width: 100%; height: 1000px"/>'
  );
})
.post(/^\/tab.*/, function() {
  var fbml = (
    '<fb:js-string var="myFrame">\n' +
    '  <fb:iframe width="100%" height="600" frameborder="0" src="http://fbrell.com/" />\n' +
    '</fb:js-string>\n' +
    '<div id="app-container"\n' +
    '     style="cursor: pointer;"\n' +
    '     onclick="document.getElementById(\'app-container\').setInnerFBML(myFrame)">\n' +
    '  <img height="600" width="760" src="http://fbrell.com/bliss.jpg?v=2">\n' +
    '</div>'
  );
  this.halt(
    '<div style="font-size: 15px">' +
    fbml +
    '<br>Placeholder Tab for <a href="http://fbrell.com/">fbrell.com</a>. ' +
    'Source code:' +
    '<textarea style="width: 100%; height: 150px">' + fbml + '</textarea>' +
    '<pre>' + this.prettyDump() + '</pre>' +
    '</div>'
  );
})
.get('/user/:username', function() {
  if (this.fb.user) {
    this.fb.api(
      {
        method: 'fql.query',
        query: 'select id, text from comment where xid="naitik"'
      },
      this.errproof(function(response) {
        this.halt(sys.inspect(response));
      })
    );
  }
})
.get('/:category/:name', function(category, name) {
  this.renderExample(category, name);
});
