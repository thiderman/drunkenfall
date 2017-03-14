# Drunkenfall

Tournament management for video game based drinking games! Written in
[go](https://golang.org/) and [vue.js](https://vuejs.org/).

## Supported games

### TowerFall
*Get a shot when you lose points (i.e. accidentally kill your player)*

* Supports 8-32 players, with a backfilling runner-up system making it possible
  to run a tournament with a number of players that is not divisable by 4.
* Lets players choose their preferred archer color and handles conflicts if
  two players with the same color are put in the same match.
* Controlled via a tablet-ready judging interface that mimics the looks of the
  score screen in the game.

## Installation

```
go get -u github.com/thiderman/drunkenfall
cd $GOPATH/src/github.com/thiderman/drunkenfall
make
```

### Development environment

In separate terminals, run each of:

```
make start-drunkenfall
make start-npm
make start-nginx
```

## License

Licensed under the MIT license.
