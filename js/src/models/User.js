export const PRODUCER = 100
export const COMMENTATOR = 50
export const JUDGE = 30
export const PLAYER = 10

export default class User {
  static fromCookies (cookieHandler) {
    let u = new User()

    // Store the public cookie data from the backend
    let data = {
      authenticated: cookieHandler.get('session') && true || false,
      userlevel: cookieHandler.get('userlevel') || 0,
    }

    Object.assign(u, data)
    return u
  }

  isProducer () {
    return this.userlevel >= PRODUCER
  }
  isCommentator () {
    return this.userlevel >= COMMENTATOR
  }
  isJudge () {
    return this.userlevel >= JUDGE
  }
  isPlayer () {
    return this.userlevel >= PLAYER
  }
};
