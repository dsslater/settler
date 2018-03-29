

app.controller('gameController', function($scope, socket) {
  $scope.dragStart = null;

  $scope.reload = function() {
    window.location.reload();
  }

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
          room: $scope.player.room
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

  $scope.gameOver = function(board) {
    if (!$scope.game.gameStart) return false;
    var firstPlayer = null;
    for (var row = 0; row < $scope.game.dimensions[0]; row++) {
      for (var col = 0; col < $scope.game.dimensions[1]; col++) {
        var cell = board[row][col];
        if (cell.city && cell.owner != "NPC") {
          if (firstPlayer) {
            if (cell.owner != firstPlayer) return false;
          } else {
            firstPlayer = cell.owner;
          }
        }
      }
    }
    $scope.game.over = true;
    socket.disconnect();
    document.body.scrollTop = 245; 
    document.documentElement.scrollTop = 245;
    $scope.game.blockReload = false;
    return true;
  }

  $scope.win = function(board) {
    var firstPlayer = '';
    for (var row = 0; row < $scope.game.dimensions[0]; row++) {
      for (var col = 0; col < $scope.game.dimensions[1]; col++) {
        var cell = board[row][col];
        if (cell.city && cell.owner != "NPC") {
          return cell.owner == $scope.player.id;
        }
      }
    }
  }
});