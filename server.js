#!/usr/bin/env node

var cluster = require('cluster')
  , path = require('path')
  , devMode = path.existsSync('/System')

var app = cluster(__dirname + '/app')
if (devMode) app.use(cluster.reload(__dirname))

app
  .use(cluster.logger(__dirname + '/var/logs'))
  .use(cluster.pidfiles(__dirname + '/var/pids'))
  .use(cluster.cli())
  .listen(43600)
