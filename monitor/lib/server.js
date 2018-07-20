/*
 *
 * For the full copyright and license information, please view the LICENSE.txt file.
 */

/* jslint node: true, esversion: 6, sub: true */
'use strict';

var express   = require('express');
var util      = require('util');
var utilex    = require('utilex');
var path      = require('path');
var http      = require('http');
var WebSocket = require('ws');

// Init the vars
var ARGS         = utilex.args();
var APP_PATH     = path.resolve(path.join(__dirname, '..'));
var NODE_PORT    = process.env.NODE_PORT    || ARGS['port']         || '3000';
var NATS_MON_URL = process.env.NATS_MON_URL || ARGS['gmonitor-url'] || 'http://192.168.1.225:6220';

// Handle errors - for preventing websocket issues
process.on('uncaughtException', function (err) {
  console.error(err);
});

// Init the server
var app    = express();
var server = http.createServer(app);
var ws     = new WebSocket.Server({server});

app.set('NODE_PORT',    NODE_PORT);    // server port
app.set('NATS_MON_URL', NATS_MON_URL); // nats monitoring url
app.set('WS',           ws);           // websocket server

app.use(require('./metrics')(app));    // metrics
app.use(require('./routes')(app));     // routes
app.use(express.static(path.join(APP_PATH, 'public'))); // static file serving

// Start the server
server.listen(NODE_PORT, function() {
  console.log('gmonitor-url Listening on %d', server.address().port);
});
