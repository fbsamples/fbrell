var
  QS   = require('querystring'),
  _    = require('underscore')._,
  fs   = require('fs'),
  path = require('path'),
  sys  = require('sys'),
  url  = require('url');

FbOpts = {
  appId: '184484190795', // fbrell
  secret: 'fa16a3b5c96463dff7ef78d783b3025a'
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
};

module.exports = require('sin/application')(__dirname)
.plug('sin/cookie')
.plug('sin/facebook', FbOpts)
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
  var populate = _.bind(function(name, dir) {
    var E = this.examples[name] = {};
    fs.readdir(dir, function(err, categories) {
      _.each(categories, function(category) {
        fs.readdir(path.join(dir, category), function(err, examples) {
          _.each(examples, function(example) {
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
  }, this);

  populate('examples', path.join(this.root, 'examples'));
  populate('examples-old', path.join(this.root, 'examples-old'));
})
.helper('makeUrl', function(path) {
  var qs = {};
  _.each(this.config, function(val, key) {
    if (DefaultConfig[key] != val) {
      qs[key] = val;
    }
  });
  qs = QS.stringify(qs);
  return path + (qs == '' ? '' : ('?' + qs));
})
.helper('script', function() {
  var server = this.config.server;

  // alias sb to my sandbox
  if (server === 'sb') {
    server = 'www.naitik.dev575.snc1';
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
    if (_.any(special, _.bind(function(s) { return server.indexOf(s) > -1;}))) {
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
.before(function() {
  this.config = _.extend({}, DefaultConfig, this.url.query);
  this.example_code = ''
  this.examples = this.config.version == 'mu'
    ? this.app.examples['examples']
    : this.app.examples['examples-old'];
})
.notFound(function() {
  this.haml('not_found');
})
.get('/', function() {
  this.haml('index', { title: 'Welcome' })
})
.get('/help', function() {
  this.haml('help')
})
.get('/examples', function() {
  this.haml('examples')
})
.get('/echo', function() {
  this.halt(
    JSON.stringify({
      post: this.post,
      url: this.url
    }),
    { 'content-type': 'text/plain' }
  );
})
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
});
