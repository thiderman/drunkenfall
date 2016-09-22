import Vue from 'vue'
import Router from 'vue-router'
import Resource from 'vue-resource'
import Cookie from 'vue-cookie'
// import DateTimePicker from 'vue-datetime-picker'
import App from './App'

import TournamentList from './components/TournamentList.vue'
import Tournament from './components/Tournament.vue'
import New from './components/New.vue'
import Match from './components/Match.vue'
import ScoreScreen from './components/ScoreScreen.vue'
import NextScreen from './components/NextScreen.vue'
import User from './models/User.js'
import Facebook from './components/Facebook.vue'
import FacebookFinalize from './components/FacebookFinalize.vue'

// install router
Vue.use(Router)
Vue.use(Resource)
Vue.use(Cookie)
// Vue.use(DateTimePicker)

// routing
var router = new Router({
  'hashbang': false,
  'history': true
})

router.map({
  '/facebook/': {
    component: Facebook,
    name: 'facebook',
  },
  '/facebook/finalize': {
    component: FacebookFinalize
  },
  '/towerfall/': {
    component: TournamentList
  },
  '/towerfall/new/': {
    name: 'new',
    component: New
  },
  '/towerfall/:tournament/': {
    name: 'tournament',
    component: Tournament
  },
  '/towerfall/:tournament/scores/': {
    name: 'scores',
    component: ScoreScreen
  },
  '/towerfall/:tournament/next/': {
    name: 'next',
    component: NextScreen
  },
  '/towerfall/:tournament/:kind/:match/': {
    name: 'match',
    component: Match
  }
})

router.beforeEach(function () {
  window.scrollTo(0, 0)
  router.app.connect()

  // Reset any pulsating lights
  document.getElementsByTagName("body")[0].className = ""

  // Always set up the user model from cookies
  router.app.$set('user', User.fromCookies(router.app.$cookie))
})

// As long as we only have Drunken TowerFall on drunkenfall.com, we should
// always redirect to the towerfall app right away.
router.redirect({
  '/': '/towerfall/'
})

router.start(App, 'app')
