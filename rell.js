require.paths.unshift('./sub/haml-js/lib');
require.paths.unshift('./sub/sin');
require.paths.unshift('./sub/underscore');

var
  QS          = require('querystring'),
  _           = require('underscore')._,
  fs          = require('fs'),
  path        = require('path'),
  sin         = require('sin'),
  sinHaml     = require('sin/haml'),
  sinJsLoader = require('sin/jsloader'),
  sinLogger   = require('sin/logger'),
  sinReloader = require('sin/reloader'),
  sinStatic   = require('sin/static'),
  sys         = require('sys');


DefaultConfig = {
  'apikey'    : 'ef8f112b63adfc86f5430a1b566f4dc1',
  'build_dev' : false,
  'comps'     : '',
  'level'     : 'debug',
  'locale'    : 'en_US',
  'old_debug' : 6,
  'server'    : 'static.ak.connect',
  'trace'     : 1,
  'version'   : 'mu'
};

sin()
.plug(sinHaml(), sinJsLoader())
.configure('development', function() {
  this.plug(sinReloader(), sinStatic(), sinLogger());
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

  var url = 'http://' + server + '.facebook.com/';

  if (this.config.version === 'mu') {
    if (url === 'http://static.ak.connect.facebook.com/') {
      url = 'http://static.ak.fbcdn.net/';
    }

    var special = ['snc', 'intern', 'beta', 'sandcastle', 'latest', 'dev'];
    if (_.any(special, _.bind(server, 'indexOf'))) {
      if (this.config.build_dev) {
        url += 'connect/en_US/core.js';
      } else {
        url += 'assets.php/' + this.config.locale + '/all.js';
      }
    } else {
      if (this.config.build_dev) {
        url += 'connect/' + this.config.locale + '/core.debug.js';
      } else {
        url += 'connect/' + this.config.locale + '/core.js';
      }
    }
  } else if (this.config.version === 'mid') {
    url += 'connect.php/' + this.config.locale + '/js/' + this.config.comps;
  } else {
    url += 'js/api_lib/v0.4/FeatureLoader.js.php';
  }

  return this.jsloader(
    [
      'http://origin.daaku.org/js-delegator/delegator.js',
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
  this.haml('index')
})
.get('^/help$', function() {
  this.haml('help')
})
.get('^/examples$', function() {
  this.haml('examples')
})
.run();
