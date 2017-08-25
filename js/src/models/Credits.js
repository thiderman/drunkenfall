import _ from 'lodash'
import Person from './Person.js'

export class Credits {
  static fromObject (obj) {
    let c = new Credits()
    Object.assign(c, obj)

    c.executive = Person.fromObject(c.executive)
    c.players = _.map(c.players, Person.fromObject)
    c.producers = _.map(c.producers, Person.fromObject)

    return c
  }
}
