var async = require('async')
  , crypto = require('crypto')
  , dotaccess = require('dotaccess')
  , express = require('express')
  , fs = require('fs')
  , knox = require('knox')
  , nurl = require('nurl')
  , path = require('path')
  , settings = require('./settings')
  , uglify = require('uglify-js')
  , util = require('util')
  , walker = require('walker')

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
  timestamp: Date.now(),
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
    if (key in { sdkUrl: 1, examplesRoot: 1, autoRun: 1 }) continue //FIXME
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

function cachedBundleHandler(contentType, files) {
  var cached
  return function(req, res, next) {
    var ttl = 315360000
    res.writeHead(200, {
      'Content-Type': contentType,
      'Cache-Control': 'public, max-age=' + ttl,
      'Expires': new Date(Date.now() + (ttl * 1000)).toUTCString(),
    })
    if (cached) {
      cached.forEach(res.write.bind(res))
      return res.end()
    }
    async.map(files, fs.readFile, function(er, results) {
      if (er) throw er
      if (contentType == 'text/javascript') {
        results = [
          uglify.uglify.gen_code(
            uglify.uglify.ast_squeeze(
              uglify.uglify.ast_mangle(
                uglify.parser.parse(results.map(
                  function(d) { return d.toString('utf8') }).join("\n")))))];
      }
      cached = results
      results.forEach(res.write.bind(res))
      res.end()
    })
  }
}

var app = module.exports = express.createServer(
  express.bodyDecoder(),
  express.methodOverride(),
  express.gzip(),
  express.staticProvider(__dirname + '/public')
)
app.configure(function() {
  app.set('view engine', 'jade')
  app.set('views', __dirname + '/views')
})
app.configure('development', function() {
  app.use(express.errorHandler({ dumpExceptions: true, showStack: true }))
})
app.configure('production', function() {
  app.use(express.errorHandler())
})
app.all('*', function(req, res, next) {
  var config = {};
  [DefaultConfig, req.query].forEach(function(src) {
    for (var key in src) {
      config[key] = src[key]
    }
  })
  config.sdkUrl = getConnectScriptUrl(
    config.version, config.locale, config.server,
    req.headers['x-forwarded-proto'] === 'https')
  config.examplesRoot = config.version == 'mu' ? 'examples' : 'examples-old'
  req.rellConfig = config

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
    rellConfig: req.rellConfig,
    makeUrl: makeUrl.bind(null, req.rellConfig),
  }})
})
app.all('/*', loadExample, function(req, res, next) {
  if (!req.exampleCode) return next()
  res.render('index', { locals: {
    title: req.params[0].replace('/', ' &middot; '),
    exampleCode: req.exampleCode,
    rellConfig: req.rellConfig,
    makeUrl: makeUrl.bind(null, req.rellConfig),
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
      rellConfig: req.rellConfig,
    },
  })
})
app.get('/examples', function(req, res, next) {
  examples.list(req.rellConfig.examplesRoot, function(er, data) {
    if (er) return next(er)
    res.render('examples', { locals: {
      examples: data,
      makeUrl: makeUrl.bind(null, req.rellConfig),
    }})
  })
})
app.all('/echo*', function(req, res, next) {
  res.writeHead(200, { 'Content-Type': 'text/plain' })
  res.end(util.inspect({
    method: req.method,
    url: req.url,
    pathname: nurl.parse(req.url).pathname,
    query: req.query,
    body: req.body,
    signedRequest: req.signedRequest,
    headers: req.headers,
    rawBody: req.rawBody,
  }))
})
app.post('/saved', function(req, res, next) {
  var exampleCode = req.body.code
  if (exampleCode.length > 10240)
    return res.send('Maximum allowed size is 10 kilybytes.', 413)

  var id = crypto.createHash('md5').update(exampleCode).digest('hex')
  s3
    .put('/' + id, {
      'Content-Length': exampleCode.length,
      'Content-Type': 'text/plain',
      'x-amz-acl': 'private',
    })
    .on('response', function(sres) {
      if (200 == sres.statusCode) return res.redirect('/saved/' + id)
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
        console.log('s3 get failed: ' + util.inspect(sres))
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
            rellConfig: req.rellConfig,
            makeUrl: makeUrl.bind(null, req.rellConfig),
          }})
        })
        .setEncoding('utf8')
    })
    .end()
})
app.get('/bundle/js/main/:timestamp', cachedBundleHandler('text/javascript', [
  __dirname + '/public/delegator.js',
  __dirname + '/public/jsDump-1.0.0.js',
  __dirname + '/public/log.js',
  __dirname + '/public/tracer.js',
  __dirname + '/public/rell.js',
  __dirname + '/public/codemirror/js/codemirror.js',
]))
app.get('/bundle/js/codemirror/:timestamp', cachedBundleHandler('text/javascript', [
  __dirname + '/public/codemirror/js/codemirror.js',
  __dirname + '/public/codemirror/js/stringstream.js',
  __dirname + '/public/codemirror/js/util.js',
  __dirname + '/public/codemirror/js/editor.js',
  __dirname + '/public/codemirror/js/select.js',
  __dirname + '/public/codemirror/js/undo.js',
  __dirname + '/public/codemirror/js/tokenize.js',
  __dirname + '/public/codemirror/js/parsecss.js',
  __dirname + '/public/codemirror/js/parsexml.js',
  __dirname + '/public/codemirror/js/tokenizejavascript.js',
  __dirname + '/public/codemirror/js/parsehtmlmixed.js',
  __dirname + '/public/codemirror/js/parsejavascript.js',
]))
app.get('/bundle/css/codemirror/:timestamp', cachedBundleHandler('text/css', [
  __dirname + '/public/codemirror/css/xmlcolors.css',
  __dirname + '/public/codemirror/css/jscolors.css',
  __dirname + '/public/codemirror/css/csscolors.css',
  __dirname + '/public/custom-codemirror.css',
]))
