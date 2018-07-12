app.controller('startDialogController', function($scope, $mdDialog, socket, game, dialog) {
  $scope.game = game;
  $scope.dialog = dialog;
  $scope.game.gamePass = '';
  $scope.game.size = "20x20";

  $scope.gameAction = function() {
    var event = $scope.dialog.create ? 'createGame' : 'joinGame';
    var height = $scope.game.size.split('x')[0];
    var width = $scope.game.size.split('x')[1];
    socket.emit(event, {
      gameId: $scope.game.gameId, 
      gamePass: $scope.game.gamePass,
      height: parseInt(height),
      width: parseInt(width),
    });
  }

  $scope.closeDialog = function() {
    $mdDialog.hide();
  }
});