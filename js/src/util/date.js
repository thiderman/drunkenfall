function isGoZeroDateOrFalsy (m) { // moment date
  if (m && m.isSame) {
    return m.isSame("0001-01-01T00:00:00Z")
  } else {
    // console.warn("moment was invalid:", m)
    return true
  }
}

export { isGoZeroDateOrFalsy }
