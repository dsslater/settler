var mongoose = require('mongoose');

var mongoDB = 'mongodb://mongo/settler'

function robustMongoConnection() {
  mongoose.connect(mongoDB, function(err) {
    if (err) {
      setTimeout(robustMongoConnection, 2000);
    }
  });
}
robustMongoConnection();

mongoose.Promise = global.Promise;
var gamesDb = mongoose.connection;

gamesDb.on('error', console.error.bind(console, 'MongoDB connection error:'));

var Schema = mongoose.Schema;

var Point = new Schema({
  row: {
    type: Number,
    required: true,
  },
  col: {
    type: Number,
    required: true,
  },
  city: {
    type: Boolean,
    required: true,
  },
  amount: {
    type: Number,
    required: true,
  },
  owner: {
    type: String,
    required: false,
  },
  color: {
    type: String,
    required: true,
  }
});

var Field = new Schema({
  dimensions: {
    width: {
      type: Number,
      required: true,
    },
    height: {
      type: Number,
      required: true,
    },
  },
  points: [Point]
});

var GameSchema = new Schema({
  gameName: {
    type: String,
    required: true,
  },
  gamePass: {
    type: String,
  },
  room: {
    type: String,
    required: true,
  },
  started: {
    type: Boolean,
    required: true,
  },
  players: {
    type: [String],
    required: true,
  },
  readyPlayers: {
    type: [String],
    required: false,
  },
  field: {
    type: Field,
    required: true,
  }
});

var Games = mongoose.model('Games', GameSchema);

module.exports.Games = Games;
