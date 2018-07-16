app.controller('gameController', function($scope, $mdDialog, socket) {
  $scope.dragStart = null;

  $scope.clearMoveArmies = function() {
    $scope.dragStart = null;
  };

  $scope.moveArmies = function(row, col) {
    if (!$scope.dragStart) return;
    if (row != $scope.dragStart[0] || col != $scope.dragStart[1]) {
      if (row == $scope.dragStart[0] || col == $scope.dragStart[1]) {
        var payload = {
          start_row: $scope.dragStart[0], 
          end_row: row,
          start_col: $scope.dragStart[1],
          end_col: col, 
          gameId: $scope.game.id
        };
        socket.emit('moveArmies', payload);
        $scope.clearMoveArmies();
      }
    }
  };

  $scope.mouseDown = function(row, col) {
    $scope.dragStart = [row, col];
  };

  $scope.fogOfWar = function(row, col, verbose=false) {
    if ($scope.game.over) 
      return false;
    var cell = $scope.game.board[row][col];
    if (cell.owner == $scope.player.id) return false;

    var found = false;

    cell.neighbors.forEach(function(neighbor) {
      if (neighbor.owner == $scope.player.id) {
        found = true;
      }
    });
    return !found;
  };

  function endGame() {
    $scope.game.over = true;
    socket.disconnect();
    window.onbeforeunload = null;
    $mdDialog.show({
      controller: 'endDialogController',
      locals: {
        game: $scope.game,
        player: $scope.player,
      },
      templateUrl: 'endDialog.html',
      parent: angular.element(document.body),
      clickOutsideToClose: true,
      fullscreen: false
    });
  }

  $scope.$watch('game.board', function(data) {
    // When the board is changed, check if the current player's game is over.
    if (!$scope.game.gameStart) return;
    if ($scope.game.over) return;
    var firstPlayer = null;
    var cityFoundForThisPlayer = false;
    var multipleActivePlayers = false;
    for (var row = 0; row < $scope.game.dimensions[0]; row++) {
      for (var col = 0; col < $scope.game.dimensions[1]; col++) {
        var cell = $scope.game.board[row][col];
        if (cell.city && cell.owner != "NPC") {
          if (cell.owner == $scope.player.id)
            cityFoundForThisPlayer = true

          if (!firstPlayer) {
            firstPlayer = cell.owner
          } else {
            if (cell.owner != firstPlayer)
              multipleActivePlayers = true;
          }
        }
      }
    }
    if (!cityFoundForThisPlayer || !multipleActivePlayers) {
      endGame();
    }
  }, true);

  $scope.$watch('game.players', function(data) {
    // When the board is changed, check if the current player's game is over.
    if (!$scope.game.gameStart) return;
    if ($scope.game.over) return;
    
    if ($scope.game.players.length <= 1) {
      endGame();
    }
  }, true);

  
});