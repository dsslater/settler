app.controller('waitForReadyController', function($scope, socket) {
  window.onbeforeunload = function (event) {
    event.returnValue = 'Are you sure that you want to leave the game?';
    return event.returnValue;
  }
  $scope.link = "http://settler.davidslater.io?gameId=" + $scope.game.id;

  socket.on('startGame', function(data) {
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
      socket.emit('playerReady', {});
    }  
  }
});