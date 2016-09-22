export default class Player {
  static fromObject (obj) {
    let p = new Player()
    Object.assign(p, obj)
    return p
  }

  // TODO(thiderman): Fix so that every avatar is set per tournament
  get avatar () {
    if (this.person.avatar_url) {
      return this.person.avatar_url
    }
    return "https://graph.facebook.com/" + this.person.facebook_id + "/picture?width=9999"
  }

  get color () {
    if (this.person.color_preference) {
      return this.person.color_preference[0]
    }
    return ""
  }

  set color (value) { }

  get displayName () {
    return this.person.nick
  }

  get firstName () {
    return this.person.name.split(" ")[0]
  }
}
