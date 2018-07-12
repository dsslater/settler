app.controller('waitForReadyController', function($scope, socket) {
  window.onbeforeunload = function (event) {
    event.returnValue = 'Are you sure that you want to leave the game?';
    return event.returnValue;
  }
  $scope.link = "http://35.202.116.91?gameId=" + $scope.game.gameId;
  
  socket.on('playerUpdate', function(data) {
    $scope.$apply(function(){
      $scope.game.readyPlayers = data.readyPlayers ? data.readyPlayers : [];
      $scope.game.players = data.players;
    });
  });

  socket.on('gameStart', function(data) {
    $scope.$apply(function(){
      $scope.player.armies = 0;
      $scope.game.gameStart = true;
    });
  });

  socket.on('playerLeft', function(data) {
    $scope.$apply(function(){
      $scope.game.numPlayers -= 1;
      var playerIndex = $scope.game.readyPlayers.findIndex(function(elem){
        return elem == data.player;
      });

      if (playerIndex >= 0) $scope.game.readyPlayers.splice(playerIndex, 1);
    });
  });

  $scope.playerReady = function() {
    if ($scope.game.numPlayers < 2) {
      alert("There must be at least two players to play.");
    } else {
      $scope.player.ready = true;
      socket.emit('iAmReady', {gameId: $scope.game.gameId});
    }  
  }
});