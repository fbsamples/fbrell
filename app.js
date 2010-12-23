var async = require('async')
  , crypto = require('crypto')
  , dotaccess = require('dotaccess')
  , express = require('express')
  , fs = require('fs')
  , knox = require('knox')
  , nurl = require('nurl')
  , path = require('path')
  , settings = require('./settings')
  , util = require('util')
  , walker = require('walker')
  , assetManager = require('connect-assetmanager')
  , assetHandler = require('connect-assetmanager-handlers')

var s3 = knox.createClient(settings.amazon)

var DefaultConfig = {
  appid: settings.facebook.id,
  level: 'debug',
  locale: 'en_US',
  server: '',
  trace: 1,
  version: 'mu',
  rte: 1,
  status: 1,
  autoRun: true,
}

var examples = function() {
  // caches
  var _contentCache = {}
    , _listCache = {}

  // normalize directory paths
  function normalizeDir(target) {
    target = path.normalize(target)
    if (target[0] !== '/') path.normalize(path.join(process.cwd(), target))
    if (target[target.length-1] !== '/') target += '/'
    return target
  }
  return {
    get: function(root, name, cb) { // get a specific file
      root = normalizeDir(root)
      var fullname = path.join(root, path.normalize('/' + name))
        , data = _contentCache[fullname]
      if (data) return process.nextTick(cb.bind(null, null, data))
      fs.readFile(fullname, 'utf8', function(er, data) {
        if (er) return cb(er)
        cb(null, _contentCache[fullname] = data)
      })
    },
    list: function(root, cb) { // get a listing of directory
      root = normalizeDir(root)
      var data = _listCache[root]
      if (data) return process.nextTick(cb.bind(null, null, data))
      data = {}
      walker(root)
        .on('file', function(file) {
          dotaccess.set(data, file.substr(root.length).split('/'), true)
        })
        .on('end', function() { cb(null, _listCache[root] = data) })
    },
  }
}()

// generate a url, maintaining the non default query params
function makeUrl(config, path) {
  var url = nurl.parse(path)
  for (var key in config) {
    if (key in { urls: 1, examplesRoot: 1, autoRun: 1 }) continue //FIXME
    var val = config[key]
    if (DefaultConfig[key] != val) url = url.setQueryParam(key, val)
  }
  return url.href
}

// generate the connect js sdk script url
function getConnectScriptUrl(version, locale, server, ssl) {
  server = {
    sb: 'www.naitik.dev1315',
    bg: 'www.brent.devrs109',
    rh: 'www.rhe.devrs106',
  }[server] || server || 'static.ak.connect'
  var url = 'http' + (ssl ? 's' : '') + '://' + server + '.facebook.com/'

  if (server === 'static.ak.connect') {
    if (version === 'mu') {
      url = 'http' + (ssl ? 's' : '') + '://connect.facebook.net/'
    } else if (ssl) {
      url = 'https://ssl.connect.facebook.com/'
    }
  }

  if (version === 'mu') {
    if (url.indexOf('//connect.facebook.net/') < 0) url += 'assets.php/'
    url += locale + '/all.js'
  } else if (version === 'mid') {
    url += 'connect.php/' + locale + '/js/'
  } else {
    url += 'js/api_lib/v0.4/FeatureLoader.js.php'
  }

  return url
}

function loadExample(req, res, next) {
  var pathname = req.params[0]
    , filename = pathname + '.html'
  examples.get(
    req.rellConfig.examplesRoot,
    filename,
    function(er, exampleCode) {
      req.exampleCode = exampleCode
      next()
    })
}

function appDataMiddleware(req, res, next) {
  var url = nurl.parse(req.url)
  if (url.hasQueryParam('app_data')) {
    var parts = url.getQueryParam('app_data').split('_')
    req.url = url
      .setQueryParam('server', parts.shift())
      .setPathname(parts.join('/'))
      .toString()
  }
  next()
}

var assets = function() {
  var groups = {
    main: {
      dataType: 'javascript',
      files: [
        'delegator.js',
        'jsDump-1.0.0.js',
        'log.js',
        'tracer.js',
        'rell.js',
        'codemirror/js/codemirror.js',
      ],
    },
    'main-css': {
      dataType: 'css',
      files: [
        'rell.css',
      ]
    },
    'codemirror-js': {
      dataType: 'javascript',
      files: [
        'codemirror/js/codemirror.js',
        'codemirror/js/stringstream.js',
        'codemirror/js/util.js',
        'codemirror/js/editor.js',
        'codemirror/js/select.js',
        'codemirror/js/undo.js',
        'codemirror/js/tokenize.js',
        'codemirror/js/parsecss.js',
        'codemirror/js/parsexml.js',
        'codemirror/js/tokenizejavascript.js',
        'codemirror/js/parsehtmlmixed.js',
        'codemirror/js/parsejavascript.js',
      ],
    },
    'codemirror-css': {
      dataType: 'css',
      files: [
        'codemirror/css/xmlcolors.css',
        'codemirror/css/jscolors.css',
        'codemirror/css/csscolors.css',
        'custom-codemirror.css',
      ]
    },
  }

  return {
    middleware: function(options) {
      for (var groupName in groups) {
        var group = groups[groupName]
        group.route = new RegExp(
          '\/bundle\/' + group.dataType + '\/' + groupName + '\/[0-9]+$')
        group.path = __dirname + '/public/'
        group.debug = options.debug
        group.stale = !options.debug
        if (!options.debug && group.dataType == 'javascript')
          group.postManipulate = {'^': [ assetHandler.uglifyJsOptimize ]}
      }
      _manager = assetManager(groups)
      return _manager
    },
    url: function(groupName) {
      var group = groups[groupName]
      if (!group) throw new Error('Group "' + groupName + '" not found!')
      return '/bundle/' + group.dataType + '/' + groupName + '/' +
        (_manager.cacheTimestamps[groupName] || Date.now());
    },
  }
}()

var app = module.exports = express.createServer(
  express.bodyDecoder(),
  express.methodOverride(),
  express.staticProvider(__dirname + '/public'),
  appDataMiddleware
)
app.configure(function() {
  app.set('view engine', 'jade')
  app.set('views', __dirname + '/views')
})
app.configure('development', function() {
  app.use(assets.middleware({ debug: true }))
  app.use(express.errorHandler({ dumpExceptions: true, showStack: true }))
})
app.configure('production', function() {
  app.use(assets.middleware({ debug: false }))
  app.use(express.errorHandler())
})
app.dynamicHelpers({
  rellConfig: function(req, res) { return req.rellConfig },
  makeUrl: function(req, res) { return req.makeUrl },
})
app.all('*', function(req, res, next) {
  var config = {};
  [DefaultConfig, req.query].forEach(function(src) {
    for (var key in src) {
      config[key] = src[key]
    }
  })
  config.urls = {
    sdk: getConnectScriptUrl(
      config.version, config.locale, config.server,
      req.headers['x-forwarded-proto'] === 'https'),
    main: assets.url('main'),
    mainCss: assets.url('main-css'),
    codemirrorJs: assets.url('codemirror-js'),
    codemirrorCss: assets.url('codemirror-css'),
  }
  config.examplesRoot = config.version == 'mu' ? 'examples' : 'examples-old'
  req.rellConfig = config
  req.makeUrl = makeUrl.bind(null, config)

  var signedRequest = req.body && req.body.signed_request
  if (signedRequest) {
    req.signedRequest = JSON.parse(
      new Buffer(
        signedRequest.split('.')[1].replace('-', '+').replace('_', '/'),
        'base64').toString('utf8'))
  }
  next()
})
app.all('/', function(req, res, next) {
  res.render('index', { locals: {
    title: 'Welcome',
    exampleCode: '',
  }})
})
app.all('/*', loadExample, function(req, res, next) {
  if (!req.exampleCode) return next()
  res.render('index', { locals: {
    title: req.params[0].replace('/', ' &middot; '),
    exampleCode: req.exampleCode,
  }})
})
app.all('/raw/*', loadExample, function(req, res, next) {
  if (!req.exampleCode) return next()
  res.send(req.exampleCode)
})
app.all('/simple/*', loadExample, function(req, res, next) {
  if (!req.exampleCode) return next()
  res.render('simple', {
    layout: false,
    locals: {
      title: req.params[0].replace('/', ' &middot; '),
      exampleCode: req.exampleCode,
    },
  })
})
app.get('/examples', function(req, res, next) {
  examples.list(req.rellConfig.examplesRoot, function(er, data) {
    if (er) return next(er)
    res.render('examples', { locals: {
      examples: data,
    }})
  })
})
app.all('/echo*', function(req, res, next) {
  res.send(util.inspect({
    method: req.method,
    url: req.url,
    pathname: nurl.parse(req.url).pathname,
    query: req.query,
    body: req.body,
    signedRequest: req.signedRequest,
    headers: req.headers,
    rawBody: req.rawBody,
  }), { 'Content-Type': 'text/plain' }, 200)
})
app.post('/saved', function(req, res, next) {
  var exampleCode = req.body.code
  if (exampleCode.length > 10240)
    return res.send('Maximum allowed size is 10 kilobytes.', 413)

  var id = crypto.createHash('md5').update(exampleCode).digest('hex')
  s3
    .put('/' + id, {
      'Content-Length': exampleCode.length,
      'Content-Type': 'text/plain',
      'x-amz-acl': 'private',
    })
    .on('response', function(sres) {
      if (200 == sres.statusCode)
        return res.redirect(req.makeUrl('/saved/' + id))
      console.error('s3 put failed: ' + util.inspect(sres))
      res.render('error')
    })
    .end(exampleCode)
})
app.all('/saved/:id', function(req, res, next) {
  s3
    .get('/' + req.params.id)
    .on('response', function(sres) {
      if (200 != sres.statusCode) {
        if (sres.statusCode != 404)
          console.error('s3 get failed: ' + util.inspect(sres))
        return next()
      }

      var exampleCode = ''
      sres
        .on('data', function(chunk) { exampleCode += chunk })
        .on('end', function() {
          req.rellConfig.autoRun = false
          res.render('index', { locals: {
            title: 'Stored Example',
            exampleCode: exampleCode,
          }})
        })
        .setEncoding('utf8')
    })
    .end()
})
