#!/usr/bin/env node

var cluster = require('cluster')
  , path = require('path')
  , clusterLoggly = require('cluster-loggly')
  , settings = require('./settings')
process.env.NODE_ENV = path.existsSync('/System') ? 'development' : 'production'

cluster(__dirname + '/app')
  .use(cluster.pidfiles(__dirname + '/var/pids'))
  .use(cluster.cli())
  .set('socket path', __dirname + '/var/sockets')
  .set('workers', 1)
  .in('development')
    .use(cluster.logger(__dirname + '/var/logs'))
    .use(cluster.reload(__dirname))
  .in('production')
    .use(clusterLoggly(settings.loggly))
  .in('all').listen(43600)
