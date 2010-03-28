require.paths.unshift('./sub/haml.js/lib');
require.paths.unshift('./sub/sin');
require.paths.unshift('./sub/underscore');

var
  QS   = require('querystring'),
  _    = require('underscore')._,
  fs   = require('fs'),
  path = require('path'),
  sys  = require('sys');


DefaultConfig = {
  'apikey'    : 'ef8f112b63adfc86f5430a1b566f4dc1',
  'comps'     : '',
  'level'     : 'debug',
  'locale'    : 'en_US',
  'old_debug' : 6,
  'server'    : 'static.ak.connect',
  'trace'     : 1,
  'version'   : 'mu'
};

FbOpts = {
  apiKey: 'ef8f112b63adfc86f5430a1b566f4dc1',
  secret: 'fa16a3b5c96463dff7ef78d783b3025a'
};

require('sin')()
.plug(
  require('sin/cookie')(),
  require('sin/facebook')(FbOpts),
  require('sin/haml')(),
  require('sin/jsloader')()
)
.configure('development', function() {
  this.plug(
    require('sin/reloader')(),
    require('sin/logger')(),
    require('sin/static')()
  );
})
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
  var populate = _.bind(function(dir) {
    var E = this.examples[dir] = {};
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

  populate('examples');
  populate('examples-old');
})
.helper('url', function(path) {
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
    ssl = this.request.headers['x-forwarded-proto'] === 'https',
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
    var special = ['snc', 'intern', 'beta', 'sandcastle', 'latest', 'dev'];
    if (_.any(special, _.bind(function(s) { return server.indexOf(s) > -1;}))) {
      url += 'assets.php/' + this.config.locale + '/all.js';
    } else {
      url += this.config.locale + '/all.js';
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
  this.title = 'FB Read Eval Log Loop'
  this.config = _.extend({}, DefaultConfig, this.uri.query);
  this.example_code = ''
  this.examples = this.config.version == 'mu'
    ? this.app.examples['examples']
    : this.app.examples['examples-old'];
})
.notFound(function() {
  this.haml('not_found');
})
.get('^/([^\/]+)/([^\/]+)$', function(category, name) {
  var example = this.examples['/' + category + '/' + name];
  if (!example) {
    this.pass();
    return;
  }
  fs.readFile(example.filename, this.errproof(function(data) {
    this.example_code = data;
    this.haml('index');
  }));
})
.get('^/$', function() {
  sys.puts(sys.inspect(this.fb().user()));
  this.haml('index')
})
.get('^/help$', function() {
  this.haml('help')
})
.get('^/examples$', function() {
  this.haml('examples')
})
.get('^/test$', function() {
  this.haml('test')
})
.run();
