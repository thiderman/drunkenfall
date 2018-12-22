# ![logo](https://avatars1.githubusercontent.com/u/22790142?s=30&v=4) Drunkenfall

Tournament management for video game based drinking games! Written in
[Go](https://golang.org/) and [Vue.js](https://vuejs.org/).

## Table of Contents

* [Supported games](#supported-games)
    * [TowerFall](#towerfall)
* [Installation](#installation)
* [Developing](#developing)
    * [Getting the code](#getting-the-code)
    * [Running](#running)
    * [Linting](#linting)
    * [Testing](#testing)
* [License](#license)

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

Drunkenfall uses [dep](https://github.com/golang/dep/) to manage its dependencies.
Please ensure it is installed on your system before proceeding with the following
instructions.

```sh
$ git clone git@github.com:drunkenfall/drunkenfall.git $(go env GOPATH)/src/github.com/drunkenfall/drunkenfall

Cloning into '.../src/github.com/drunkenfall/drunkenfall'...
remote: Enumerating objects: 108, done.
remote: Counting objects: 100% (108/108), done.
remote: Compressing objects: 100% (73/73), done.
remote: Total 6899 (delta 54), reused 71 (delta 35), pack-reused 6791
Receiving objects: 100% (6899/6899), 20.12 MiB | 11.57 MiB/s, done.
Resolving deltas: 100% (5171/5171), done.

$ cd $(go env GOPATH)/src/github.com/drunkenfall/drunkenfall
$ dep ensure -v

Root project is "github.com/drunkenfall/drunkenfall"
 3 transitively valid internal packages
 19 external packages imported from 16 projects
(0)   ✓ select (root)
(1)     ? attempt github.com/StefanSchroeder/Golang-Roman with 1 pkgs; at least 1 versions to try
(1)         try github.com/StefanSchroeder/Golang-Roman@master
(1)     ✓ select github.com/StefanSchroeder/Golang-Roman@master w/1 pkgs
...
(34/36) Wrote golang.org/x/crypto@master
(35/36) Wrote golang.org/x/net@master
(36/36) Wrote github.com/magefile/mage@v1.8.0
```

## Developing

The following dependencies/utilities are expected to be availabe on your system:

* [mage](https://magefile.org)
* [Docker Compose](https://docs.docker.com/compose/)
* [Caddy](https://caddyserver.com/)
* [GolangCI-Lint](https://github.com/golangci/golangci-lint)

Once you've installed `mage` you can run `mage proxy:install` in order to get
Caddy installed.

### Getting the code

If you want to develop on Drunkenfall you should start with forking the
repository. Once that's done you can follow the [Installation instructions](#installation)
on how to get it up and running. However, adjust the `git clone` command like
so:

```sh
git clone git@github.com:<YOUR GITHUB USERNAME>/drunkenfall.git $(go env GOPATH)/src/github.com/drunkenfall/drunkenfall
```

This will ensure your fork is used but checked out in the right part of your
GOPATH, so you don't get any import path issues. It's advisable to then add
`git@github.com/drunkenfall/drunkenfall` as a remote to your clone in order
to be able to regularly update your copy of the repository. Github [provides
documentation](https://help.github.com/articles/configuring-a-remote-for-a-fork/)
on how to do so.

### Running

First run:

```sh
export DF_DB_PATH=$(go env GOPATH)/src/github.com/drunkenfall/drunkenfall/data/test.db
mkdir $(go env GOPATH)/src/github.com/drunkenfall/drunkenfall/data && touch $DF_DB_PATH
```

Then, in separate terminals, run each of:

```sh
mage run:postgres
mage run:drunkenfall
mage run:npm
mage run:proxy
```

### Linting

[GolangCI-Lint](https://github.com/golangci/golangci-lint) is used for linting
purposes. Once it is installed run the following

```sh
mage code:lint
```

This will use the [`.golangici-pedantic.yaml`](.golangci-pedantic.yaml) which
encodes all the linters we expect to pass during CI.

In order to help you out during development it's advisable to hookup
`golangci-lint run` with the `--fast` argument to your editor. This won't run
all the linters but should catch most issues while not blocking your editor.

### Testing

Drunkenfall has a suite of tests that you can run:

```sh
mage code:test
```

## License

Licensed under the MIT license.
