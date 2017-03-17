# Drunkenfall

[![Build Status](http://jenkins.dragonisle.se/job/Drunkenfall/job/drunkenfall/job/master/badge/icon)](http://jenkins.dragonisle.se/job/Drunkenfall/job/drunkenfall/job/master/)

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
go get -u github.com/drunkenfall/drunkenfall
cd $GOPATH/src/github.com/drunkenfall/drunkenfall
make
```

### Development environment

In separate terminals, run each of:

```
make drunkenfall-start
make npm-start
make nginx-start
```

## License

Licensed under the MIT license.
