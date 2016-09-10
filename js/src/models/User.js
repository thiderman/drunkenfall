export default class User {
  static fromCookies (cookieHandler) {
    let u = new User()

    // Store the public cookie data from the backend
    let data = {
      authenticated: cookieHandler.get('session') && true || false,
      userlevel: cookieHandler.get('userlevel') || 0,
    }
    console.log(data.userlevel)

    Object.assign(u, data)
    return u
  }

  level (lvl) {
    if (this.userlevel >= lvl) {
      return true
    }
    return false
  }
};
