'use strict';

var express = require('express');
var app = express();
var server = require('http').Server(app);
var io = require('socket.io')(server);
var game = require('./gameLogic')(io);

var PORT_NUMBER = 80;   // Port number to use

io.on('connection', function (socket) {

  socket.on('disconnect', function() {
    game.removePlayer(socket.id);
  });
  
  socket.on('createGame', function (data) {
    var gameName = data.gameName;
    var gamePass = data.gamePass;
    game.createGame(gameName, gamePass, socket);
  });

  socket.on('joinGame', function (data) {
    var gameName = data.gameName;
    var gamePass = data.gamePass;
    game.joinGame(gameName, gamePass, socket);
  });

  socket.on('addArmies', function(data) {
    var room = data.room;
    var row = data.row;
    var col = data.col;
    var amount = data.amount;
    game.addArmies(socket.id, room, row, col, amount);
  });

  socket.on('moveArmies', function(data) {
    var room = data.room;
    var startRow = data.start_row;
    var startCol = data.start_col;
    var endRow = data.end_row;
    var endCol = data.end_col;
    game.moveArmies(socket.id, room, startRow, startCol, endRow, endCol);
  });

  socket.on('iAmReady', function(data) {
    game.setupPlayer(socket.id, data.room);
  });
});


app.use(express.static('public'));
server.listen(PORT_NUMBER, function() {
  console.log('server up and running at port ' + PORT_NUMBER);
});
