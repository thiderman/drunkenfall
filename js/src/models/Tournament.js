import { isGoZeroDateOrFalsy } from '../util/date.js'
import moment from 'moment'
import Player from './Player.js'
import Person from './Person.js'
import Match from './Match.js'
import _ from 'lodash'

export default class Tournament {
  static fromObject (obj) {
    let t = new Tournament()
    Object.assign(t, obj)

    t.opened = moment(t.opened)
    t.scheduled = moment(t.scheduled)
    t.started = moment(t.started)
    t.ended = moment(t.ended)

    t.tryouts = _.map(t.tryouts, Match.fromObject)
    t.semis = _.map(t.semis, Match.fromObject)
    t.final = Match.fromObject(t.final)

    t.players = _.map(t.players, Player.fromObject)
    t.runnerups = _.map(t.runnerups, Person.fromObject)

    return t
  }

  get isStarted () {
    // tournament is started if 'started' is defined and NOT equal to go's zero date
    return !isGoZeroDateOrFalsy(this.started)
  }

  get isEnded () {
    // tournament is ended if 'ended' is defined and NOT equal to go's zero date
    return !isGoZeroDateOrFalsy(this.ended)
  }

  get isTest () {
    // tournament is ended if 'ended' is defined and NOT equal to go's zero date
    return !this.name.startsWith('DrunkenFall')
  }

  get canStart () {
    return !this.isStarted
  }

  get isRunning () {
    return this.isStarted && !this.isEnded
  }

  get canShuffle () {
    // We can only shuffle before the first match has been started.
    let match = Match.fromObject(this.tryouts[0])
    return !match.isStarted
  }

}
