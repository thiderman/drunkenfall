import Vue from 'vue'
import Router from 'vue-router'
import Resource from 'vue-resource'
import Cookie from 'vue-cookie'
import App from './App'

import vPikaday from '../vendor/vue-pikaday'

import TournamentList from './components/TournamentList.vue'
import Tournament from './components/Tournament.vue'
import New from './components/New.vue'
import Join from './components/Join.vue'
import Participants from './components/Participants.vue'
import Edit from './components/Edit.vue'
import Match from './components/Match.vue'
import PostMatch from './components/PostMatch.vue'
import ScoreScreen from './components/ScoreScreen.vue'
import NextScreen from './components/NextScreen.vue'
import User from './models/User.js'
import Facebook from './components/Facebook.vue'
import FacebookFinalize from './components/FacebookFinalize.vue'
import Credits from './components/Credits.vue'

// install router
Vue.use(Router)
Vue.use(Resource)
Vue.use(Cookie)
Vue.use(vPikaday)

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
  '/towerfall/:tournament/join/': {
    name: 'join',
    component: Join
  },
  '/towerfall/:tournament/participants/': {
    name: 'participants',
    component: Participants
  },
  '/towerfall/:tournament/edit/': {
    name: 'edit',
    component: Edit
  },
  '/towerfall/:tournament/scores/': {
    name: 'scores',
    component: ScoreScreen
  },
  '/towerfall/:tournament/next/': {
    name: 'next',
    component: NextScreen
  },
  '/towerfall/:tournament/charts/': {
    name: 'charts',
    component: PostMatch
  },
  '/towerfall/:tournament/credits/': {
    name: 'credits',
    component: Credits
  },
  '/towerfall/:tournament/:kind/:match/': {
    name: 'match',
    component: Match
  },
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
