import { isGoZeroDateOrFalsy } from '../util/date.js'
import moment from 'moment'
import _ from 'lodash'
import Player from './Player.js'
// import Tournament from './Tournament.js'

// import store from '../core/store.js'a

export default class Match {
  static fromObject (obj, t) {
    let m = new Match()
    Object.assign(m, obj)

    m.started = moment(m.started)
    m.ended = moment(m.ended)
    m.scheduled = moment(m.scheduled)
    m.players = _.map(m.players, Player.fromObject)

    m.endScore = m.length

    // TODO(thiderman): There used to be this weird bug where some
    // matches are created without a length. This circumvents this in
    // a semi-ugly way. The bug is fixed, but workarounds are forever. <3
    if (m.endScore === 0) {
      m.endScore = m.kind === "final" ? 20 : 10
    }

    if (t !== undefined) {
      m.tournament = t
      m.tournament_id = t.id
    }

    return m
  }

  // TODO(thiderman): This could somehow not be migrated to here from
  // Match.vue. When moved, the request turns from a POST into a GET
  // and the backend rightfully denies it. A thing for later, I guess.
  // commit ($control, payload) {
  //   this.api.commit(this.id, payload).then((res) => {
  //     console.log("Round committed.")
  //     _.each($control.players, (p) => { p.reset() })
  //   }, (res) => {
  //     console.error('error when setting score', res)
  //   })
  // }

  get id () {
    return {
      id: this.tournament_id,
      index: this.index,
    }
  }

  get endScore () { return this._end }
  set endScore (value) { this._end = value }

  get title () {
    if (this.kind === "final") {
      return "Final"
    }
    let kind = this.kind + "s"
    return _.capitalize(this.kind) + " " + this.relativeIndex + " / " + this.tournament[kind].length
  }

  get relativeIndex () {
    if (this.kind === "semi") {
      return this.index - this.tournament.playoffs.length + 1
    }
    return this.index + 1
  }

  get isStarted () {
    // match is started if 'started' is defined and NOT equal to go's zero date
    return !isGoZeroDateOrFalsy(this.started)
  }

  get isEnded () {
    // match is ended if 'ended' is defined and NOT equal to go's zero date
    return !isGoZeroDateOrFalsy(this.ended)
  }

  get isScheduled () {
    return !isGoZeroDateOrFalsy(this.scheduled)
  }

  get canStart () {
    return !this.isStarted
  }

  get canEnd () {
    // can't end if already ended
    if (this.isEnded) {
      return false
    }

    // can end if at least one player has enough kills (ie >= end)
    return _.some(this.players, (player) => { return player.kills >= this.endScore })
  }

  get canReset () {
    return this.isRunning && this.commits.length > 0
  }

  get isRunning () {
    return this.isStarted && !this.isEnded
  }

  get chartData () {
    var out = []
    for (var i = 0; i < this.players.length; i++) {
      out.push([0])
      _.forEach(this.commits, function (commit) {
        let pastScore = _.last(out[i])
        let roundScore = _.sum(commit.kills[i])
        out[i].push(pastScore + roundScore)
      })
    }
    return out
  }

  get levelTitle () {
    if (this.level === "twilight") {
      return "Twilight Spire"
    } else if (this.level === "kingscourt") {
      return "King's Court"
    } else if (this.level === "frostfang") {
      return "Frostfang Keep"
    } else if (this.level === "sunken") {
      return "Sunken City"
    } else if (this.level === "amaranth") {
      return "The Amaranth"
    }
    return this.level.charAt(0).toUpperCase() + this.level.slice(1)
  }
};
