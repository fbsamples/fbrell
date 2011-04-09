#!/usr/bin/env node

var cluster = require('cluster')

cluster(__dirname + '/app')
  .use(cluster.logger(__dirname + '/var/logs'))
  .use(cluster.pidfiles(__dirname + '/var/pids'))
  .use(cluster.cli())
  .in('development')
    .set('workers', 1)
    .use(cluster.reload(__dirname))
  .in('all').listen(43600)
