export default class Person {
  static fromObject (obj) {
    let p = new Person()
    Object.assign(p, obj)
    return p
  }

  get avatar () {
    if (this.avatar_url) {
      return this.avatar_url
    }
    return "https://graph.facebook.com/" + this.facebook_id + "/picture?width=9999"
  }

  get color () {
    if (!this.color_preference) {
      return "white"
    }
    return this.color_preference[0]
  }

  get displayName () {
    return this.nick
  }

  get firstName () {
    return this.name.split(" ")[0]
  }
}
