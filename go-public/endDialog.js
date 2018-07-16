app.controller('endDialogController', function($scope, socket, game, player) {
  $scope.game = game;
  $scope.player = player;

  $scope.reload = function() {
    newUrl = window.location.href.substring(0, window.location.href.indexOf('?'));
    window.location = newUrl;
  }

  $scope.win = function() {
    if ($scope.game.players.length <= 1)
      return true;
    
    for (var row = 0; row < $scope.game.dimensions[0]; row++) {
      for (var col = 0; col < $scope.game.dimensions[1]; col++) {
        var cell = $scope.game.board[row][col];
        if (cell.city && cell.owner != "NPC") {
          return cell.owner == $scope.player.id;
        }
      }
    }
  }
});