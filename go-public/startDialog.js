app.controller('startDialogController', function($scope, $mdDialog, socket, player, game, dialog) {
  $scope.player = player;
  $scope.game = game;
  $scope.dialog = dialog;
  $scope.player.gamePass = '';
  $scope.game.size = "20x20";

  $scope.gameAction = function() {
    var event = $scope.dialog.create ? 'createGame' : 'joinGame';
    var height = $scope.game.size.split('x')[0];
    var width = $scope.game.size.split('x')[1];
    socket.emit(event, {
      gameId: $scope.game.gameId, 
      gamePass: $scope.player.gamePass,
      height: parseInt(height),
      width: parseInt(width),
    });
  }

  $scope.closeDialog = function() {
    $mdDialog.hide();
  }
});