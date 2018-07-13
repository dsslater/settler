app.controller('endDialogController', function($scope, socket, game, player) {
  $scope.game = game;
  $scope.player = player;

  $scope.reload = function() {
    newUrl = window.location.href.substring(0, window.location.href.indexOf('?'));
    console.log("Reloading: " + newUrl);
    window.location = newUrl;
    // window.location.reload();
    console.log("URL: " + window.location.href)
  }

  $scope.win = function() {
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