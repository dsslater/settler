var io = null
var Games = require('./db').Games;

var COLORS = ['red', 'green', 'blue', 'orange', 'purple', 'yellow', 'grey', 'pink'];
var NUM_CITIES = 10;
var GROWTH_TIME = 20000; //ms
var CITY_BONUS = 5;
var CITY_JITTER_AMOUNT = 10;
var CITY_BASE_AMOUNT = 40;

module.exports = function(ioInit) {
  io = ioInit;
  return {
    removePlayer: removePlayer,
    createGame: createGame,
    joinGame: joinGame,
    addArmies: addArmies,
    moveArmies: moveArmies,
    addReadyPlayer: addReadyPlayer,
    armyGrowth: armyGrowth,
    GROWTH_TIME: GROWTH_TIME,
  }
}

var armyGrowthLoop = null;

var cityGrowthLoop = null;


function setupGrowth() {
  if (!armyGrowthLoop) {
    console.log("Creating growth threads.");
    armyGrowthLoop = setInterval(armyGrowth, GROWTH_TIME);
    cityGrowthLoop = setInterval(function() {
      armyGrowth(true);
    }, GROWTH_TIME/5);
  }
}

function tearDownGrowth() {
  Games.find({started: true}, function(err, games) {
    if (err)
      return console.error(err);

    if (games.length == 0) {
      // race condition
      console.log("Removing growth threads.");
      clearInterval(armyGrowthLoop);
      clearInterval(cityGrowthLoop);
      armyGrowthLoop = null;
      cityGrowthLoop = null;
    }
  });
}

function getColor(game, player) {
  /* Return the color for the provided player in the provided game. */
  if (player == 'NPC') return 'white';

  var colorIndex = game.readyPlayers.findIndex(function(elem) {
    return elem == player;
  });

  return COLORS[colorIndex];
}

function getCities(height, width) {
  /* Return an array of NUM_CITIES random points. */
  var cities = []
  for (var i = 0; i < NUM_CITIES; i++) {
    var point = {
      row: Math.floor(Math.random() * height),
      col: Math.floor(Math.random() * width)
    };
    cities.push(point);
  }
  return cities;
}

function initializeBoard(game, height, width) {
  /* Initialize board with zeros and cities. */
  var cities = getCities(height, width);
  game.points = [];
  for (var row = 0; row < height; row++) {
    for (var col = 0; col < width; col++) {
      var city = Boolean(cities.find(function(elem) {
        return elem.row == row && elem.col == col;
      }));

      var amount = 0;
      
      if (city) {
        // Amount to defend cities.
        var jitter = Math.floor(Math.random() * CITY_JITTER_AMOUNT);
        amount = jitter + CITY_BASE_AMOUNT;
      }
      
      var point = {
        row: row,
        col: col,
        amount: amount,
        owner: 'NPC',
        city: city,
        color: getColor('', 'NPC'),
      };
      game.points.push(point);
    }
  }
}

function createGame(socket, gameName, gamePass, height, width) {
  /* Create a collection for a game with the provided gameName and gamePass. */
  Games.find({gameName: gameName}, function(err, games){
    if (err) {
      socket.emit('unkownGameCreationError');
      return console.error(err);
    }
    if (games.length >= 1) {
      // gameName has already been used.
      socket.emit('gameNameUsedError');
      return console.error('Game name ' + gameName + ' already in use.');
    }

    // Create a game object. 
    var game = {
      gameName: gameName,
      gamePass: gamePass,
      started: false,
      room: socket.id,
      players: [socket.id],
      height: height, 
      width: width,
    };

    initializeBoard(game, height, width);
    
    // Write game to db.
    Games.create(game, function(err, gameInst){
      if (err) {
        socket.emit('unkownGameCreationError');
        console.error(err);
        return
      }

      var room = gameInst.room;

      // Join room for all future messages.
      socket.join(room);

      var gameInformation = {
        room: room, 
        id: socket.id,
        dimensions: [height, width],
        points: gameInst.points,
        numPlayers: 0
      };
      socket.emit('gameReady', gameInformation);
    
      var playerInformation = {
        players: gameInst.players,
        readyPlayers: gameInst.readyPlayers
      };
      messageRoom(game.room, 'playerUpdate', playerInformation);
      console.log('Game ' + gameName + ' created.');
      setupGrowth();
    });
  });  
}

function joinGame(socket, gameName, gamePass) {
  /* Attempt to join the provided gameName. */
  Games.find({gameName: gameName}, function(err, games){
    if (err) {
      socket.emit('unknownGameJoinError');
      return console.error(err);
    } else if (games.length < 1) {
      socket.emit('gameNotFound');
      return console.log('Game ' + gameName + ' not found.')
    } else if (games.length > 1) {
      socket.emit('unknownGameJoinError');
      return console.error('DBERROR: Too many games with this id.');
    } 
    var game = games[0];
    if (game.gamePass != gamePass) {
      socket.emit('wrongPassword');
      return console.log('Game ' + gameName + ' was given the wrong password.');
    } else if (game.started) {
      socket.emit('gameStarted');
      return console.log('Game ' + gameName + ' has already started.');
    }

    game.players.push(socket.id);
    game.save(function(err, prod, numAffected) {
      socket.join(game.room);
      console.log('Joining ' + gameName);

      var gameInformation = {
        room: game.room,
        id: socket.id,
        dimensions: [game.height, game.width],
        points: game.points,
      };
      socket.emit('gameReady', gameInformation);

      var playerInformation = {
        players: game.players,
        readyPlayers: game.readyPlayers
      };
      messageRoom(game.room, 'playerUpdate', playerInformation);
    });

  });
}

function lessThanEqual(lh, rh) {
  return lh <= rh;
}

function greaterThanEqual(lh, rh) {
  return lh >= rh;
}

function addition(a, b) {
  return a + b;
}

function subtraction(a, b) {
  return a - b;
}

function getPath(game, player, startRow, startCol, endRow, endCol) {
  /* Return a list of the points in the order of the movement. Returns null if 
     the move is not legitimate. */
  var path = [];
  var rowOp = startRow > endRow ? subtraction : addition;
  var rowComp = startRow > endRow ? greaterThanEqual : lessThanEqual;
  var colOp = startCol > endCol ? subtraction : addition;
  var colComp = startCol > endCol ? greaterThanEqual : lessThanEqual;

  for (var row = startRow; rowComp(row, endRow); row = rowOp(row,  1)) {
    for (var col = startCol; colComp(col, endCol); col = colOp(col, 1)) {
      var point = game.points.find(function(elem) {
        return elem.row == row && elem.col == col;
      });

      /* If the player does not own all of the points in the movement besides 
         the last one return null. */
      if (point.owner != player && (col != endCol || row != endRow)) 
        return null;
      
      path.push(point);
    }
  }
  return path;
}

function moveArmies(player, room, startRow, startCol, endRow, endCol) {
  /* Move all armies along the path from the start point to the end point. */
  Games.find({'room': room}, function(err, games){
    if (err) {
      return console.error(err);
    } else if (games.length < 1) {
      return console.log('Game not found for moving Armies.');
    }

    var game = games[0];
    var path = getPath(game, player, startRow, startCol, endRow, endCol);
    // If the path is not legitimate do not move any armies.
    if (!path) 
      return;

    // Sum the number of armies on the path and set points on the path to 1.
    var total = 0;
    path.forEach(function(point){
      if (point.owner == player) {
        total += Math.max(0, (point.amount - 1));
        if (point.amount > 1)
          point.amount = 1;
      }
    });

    game.save(function(err, prod, numAffected) {
      if (err) {
        console.log(err);
      } else {
        if (total > 0) {
          // Add all armies from the path to the last point on the path.
          var target = path[path.length -1];
          addArmies(player, room, target.row, target.col, total);
        }
        messageRoom(room, 'update', path);
      }
    });
  });
}

function messageRoom(room, event, payload={}) {
  io.to(room).emit(event, payload);
}

function updatePoints(players, game, points, amounts, citySetup=false, callback=null) {
  /* Update an array of points with the provided information. */

  var changedPoints = [];

  for (var i = 0; i < points.length; i++) {
    var player = players[i];
    var amount = amounts[i];
    var point = game.points.find(function(elem) {
      return elem.row == points[i].row && elem.col == points[i].col;
    });

    if (point.owner == 'NPC' && !point.city) {
      // Update a non-player point that does not need to be captured.
      point.owner = player;
      point.amount = amount;
      point.color = getColor(game, player);
    } else if (point.owner == player) {
      // If player already owns this point add the amount to the point.
      point.amount += amount;
    } else {
      /* If the player does not own the point subtract the amoint from the 
         point and, if the amount is greater than the amount in the point, 
         change ownership. */
      var remainder = point.amount - amount;
      if (remainder >= 0) {
        point.amount = remainder;
      } else {
        point.amount = remainder * -1;
        point.color = getColor(game, player);
        point.owner = player;
      }
    }

    if (citySetup)
      point.city = true;

    changedPoints.push(point);
  }

  game.save(function(err, prod, numAffected) {
    messageRoom(game.room, 'update', changedPoints);
    if (callback)
      callback();
  });
}

function addArmies(player, room, row, col, amount) {
  /* Add armies from the provided player to the point at row, col. */
  Games.find({'room': room}, function(err, games){
    if (err) {
      return console.error(err);
    } else if (games.length < 1) {
      return console.log('Could not find game to add armies to.')
    }
    updatePoints([player], games[0], [{row: row, col: col}], [amount]);
  });
}

function setupPlayerCities(game) {
  /* Select a random starting position for each player. */
  var citiesAdded = [];
  var amounts = [];
  game.readyPlayers.forEach(function(player) {
    while (true) {
      var row = Math.floor(Math.random() * game.height)
      var col = Math.floor(Math.random() * game.width)

      var point = game.points.find(function(elem) {
        return elem.row == row && elem.col == col;
      });

      if (!point.city) {
        console.log('Setting starting point at: ', row, col, " for ", player);
        citiesAdded.push({row: row, col: col});
        amounts.push(1);
        break;
      }
    }
  });

  updatePoints(game.readyPlayers, game, citiesAdded, amounts, true, function() {
    var playerInformation = {
      players: game.players,
      readyPlayers: game.readyPlayers,
    };

    messageRoom(game.room, 'playerUpdate', playerInformation);
    if (game.started) 
      messageRoom(game.room, 'gameStart');
  });

}

function addReadyPlayer(player, room) {
  /* Mark given player as ready and if all players are ready start the game. */
  Games.find({'room': room}, function(err, games){
    if (err) {
      return console.error(err);
    } else if (games.length < 1) {
      return console.log('Could not find game to mark player as ready.');
    }
    var game = games[0];

    game.readyPlayers.push(player);

    // If all players are ready, start the game.
    if (game.readyPlayers.length == game.players.length && game.players.length > 1) {
      game.started = true;
      setupPlayerCities(game);
    } else {
      game.save(function(err, prod, numAffected) {
        var playerInformation = {
          players: game.players,
          readyPlayers: game.readyPlayers,
        };
        messageRoom(game.room, 'playerUpdate', playerInformation);
      });
    }
  });
}

function removePlayer(player) {
  /* Remove a player from a game. If they are the last player, delete the 
     game.*/
  Games.find({'players': {$in: [player]}}, function(err, games){
    if (err) {
      return console.error(err);
    } else if (games.length < 1) {
      return console.log('No game found for socket ' + player + '.');
    }

    var game = games[0];
    var playerIndex = game.players.findIndex(function(elem) {
      return elem == player;
    });

    game.players.splice(playerIndex, 1);

    if (!game.started) {
      /* If they indicated they were ready, they must also be removed the from 
         the readyPlayer array. */
      var readyPlayerIndex = game.readyPlayers.findIndex(function(elem) {
        return elem == player;
      });
      if (readyPlayerIndex >= 0) {
        game.readyPlayers.splice(readyPlayerIndex, 1);
      }
    }

    if (game.players.length == 0) {
      // This is the last player in the game to disconnect.
      Games.findByIdAndRemove(game.id, function(err) {
        if (err) {
          console.log(err);
        } else {
          console.log("Removed game " + game.gameName + '.');
          tearDownGrowth();
        }
      });
    } else {
      game.save(function(err, prod, numAffected) {
        messageRoom(game.room, 'playerUpdate', {players: game.players,
                                              readyPlayers: game.readyPlayers});
      });
    }
  });
}

function armyGrowth(cityGrowth=false) {
  /* Grow the armies for all active games. */
  var query = {started: true};
  var update = null;
  var options = null
  if (cityGrowth) {
    update = {$inc: {"points.$[point].amount": 1}};
    options = {arrayFilters: [{$and: [{"point.owner": {"$ne": "NPC"}}, 
                                      {"point.city": true}]}]};
  } else {
    update = {$inc: {"points.$[point].amount": 1}}
    options = {arrayFilters: [{"point.owner": {"$ne": "NPC"}}]};
  }

  Games.updateMany(query, update, options, function(err, affected) {
    if (err) return console.error(err);
    Games.find({started: true}, function(err, games) {
      if (err) return console.error(err);
      games.forEach(function(game) {
        messageRoom(game.room, 'update', game.points);
      });
    });
  });
}

