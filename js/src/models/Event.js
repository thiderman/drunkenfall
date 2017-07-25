import moment from 'moment'
import template from 'string-template'
import Player from './Player.js'
import _ from 'lodash'

export default class Event {
  static fromObject (obj) {
    let e = new Event()
    Object.assign(e, obj)

    e.time = moment(e.time)
    e.hasPlayer = false
    if (_.has(e.items, 'person')) {
      e.player = Player.fromObject({"person": e.items.person})
      e.hasPlayer = true
    }
    return e
  }

  get print () {
    return template(this.message, this.items)
  }

}
