var app = angular.module('Settler', ['ngMaterial', 'ngMessages']);

app.factory('socket', [function() {
  var events = {} // map event names to registered functions
  var socket = new WebSocket("ws://35.202.116.91/game");
  socket.onmessage = function (evt) { 
    message = JSON.parse(evt.data);
    if (message['event'] in events {
      handler = message['event']
      data = message['data'];
      handler(data);
    }
  }

  return {
    on: function(eventName, callback){
      if (callback == null) {
        delete events[eventName];
      } else {
        events[eventName] = callback;
      }
    },
    emit: function(eventName, data) {
      wrapper = {};
      wrapper['event'] = eventName;
      wrapper['data'] = data;
      socket.send(JSON.stringify(wrapper));
    },
    disconnect: function() {
      socket.close();
    }
  };
}]);

app.controller('mainController', function($scope, $mdDialog, socket) {
  $scope.dialog = {}
  $scope.player = {};
  $scope.game = {};
  $scope.player.gameReady = false;
  $scope.player.armies = 0;
  $scope.game.gameStart = false;
  $scope.game.readyPlayers = [];
  $scope.game.over = false;
  $scope.player.win = false;

  $scope.showDialog = function(ev, create) {
    $scope.dialog.create = create;
    $mdDialog.show({
      controller: 'startDialogController',
      locals: {
        player: $scope.player,
        game: $scope.game,
        dialog: $scope.dialog,
      },
      templateUrl: 'startDialog.html',
      parent: angular.element(document.body),
      targetEvent: ev,
      clickOutsideToClose:true,
      fullscreen: false
    });
  };

  socket.on('gameReady', function(data){
    $scope.$apply(function(){
      $scope.game.dimensions = data.dimensions;
      $scope.game.board = [];
      // Create board
      for (var row = 0; row < $scope.game.dimensions[0]; row++) {
        $scope.game.board.push(new Array($scope.game.dimensions[1]))
      }
      // Populate Board
      data.points.forEach(function(point) {
        $scope.game.board[point.row][point.col] = {
          amount: point.amount, 
          color: point.color, 
          owner: point.owner, 
          city: point.city,
          row: point.row,
          col: point.col,
        };
      });
      // Set up neighbors
      for (var row = 0; row < $scope.game.dimensions[0]; row++) {
        for (var col = 0; col < $scope.game.dimensions[1]; col++) {
          var cell = $scope.game.board[row][col];
          var top = Math.max(0, row - 1);
          var bottom = Math.min($scope.game.dimensions[0] - 1, row + 1);
          var left = Math.max(0, col - 1);
          var right = Math.min($scope.game.dimensions[1] - 1, col + 1);
          cell.neighbors = new Set([
            $scope.game.board[top][col],
            $scope.game.board[row][left],
            $scope.game.board[row][right],
            $scope.game.board[bottom][col],
          ].filter(neighbor => neighbor != cell));
        }
      }
      
      $scope.player.room = data.room;
      $scope.player.id = data.id;
      $scope.game.players = [$scope.player.id];
      $scope.player.gameReady = true;
    });
  });

  socket.on('unkownGameCreationError', function(data) {
    alert('Sorry, we were unable to create ' + $scope.player.gameName
          + '. Please try creating a new game.');
  });

  socket.on('wrongPassword', function(data) {
    alert('Sorry, this is not the password for ' + $scope.player.gameName)
  });

  socket.on('gameNameUsedError', function(data) {
    alert('Sorry, ' + $scope.player.gameName + ' is already in use, '
          + 'please try another.');
  });

  socket.on('unknownGameJoinError', function(data) {
    alert('Sorry, we were unable to connect to ' + $scope.player.gameName
          + '. Please try creating a new game.');
  });

  socket.on('gameNotFound', function(data) {
    alert('Sorry, we were unable to find ' + $scope.player.gameName
          + '. Please try creating a new game.');
  });

  socket.on('gameStarted', function(data) {
    alert('Sorry, ' + $scope.player.gameName + ' has already started.');
  });

  socket.on('update', function(datas){
    $scope.$apply(function(){
      datas.forEach(function(data) {
        var cell = $scope.game.board[data.row][data.col];
        cell.amount = data.amount;
        cell.color = data.color;
        cell.owner = data.owner;
        cell.city = data.city;
      });
    });
  });
});