app.controller('dialogController', function($scope, $mdDialog, socket, player, game, dialog) {
  $scope.player = player;
  $scope.game = game;
  $scope.dialog = dialog;
  $scope.player.gameName = '';
  $scope.player.gamePass = '';
  $scope.game.size = "20x20";

  $scope.gameAction = function() {
    var event = $scope.dialog.create ? 'createGame' : 'joinGame';
    var height = $scope.game.size.split('x')[0];
    var width = $scope.game.size.split('x')[1];
    socket.emit(event, {
      gameName: $scope.player.gameName, 
      gamePass: $scope.player.gamePass,
      height: height,
      width: width,
    });
  }

  $scope.closeDialog = function() {
    $mdDialog.hide();
  }
});