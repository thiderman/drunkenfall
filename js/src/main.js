/* eslint-env browser */

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
import Log from './components/Log.vue'
import ScoreScreen from './components/ScoreScreen.vue'
import NextScreen from './components/NextScreen.vue'
// import User from './models/User.js'
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
  mode: 'history',
  routes: [
    // As long as we only have Drunken TowerFall on drunkenfall.com, we should
    // always redirect to the towerfall app right away.
    {
      path: '/',
      redirect: '/towerfall/',
    },
    {
      path: '/facebook/',
      component: Facebook,
      name: 'facebook',
    },
    {
      path: '/facebook/finalize',
      component: FacebookFinalize
    },
    {
      path: '/towerfall/',
      component: TournamentList
    },
    {
      path: '/towerfall/new/',
      name: 'new',
      component: New
    },
    {
      path: '/towerfall/:tournament/',
      name: 'tournament',
      component: Tournament
    },
    {
      path: '/towerfall/:tournament/join/',
      name: 'join',
      component: Join
    },
    {
      path: '/towerfall/:tournament/participants/',
      name: 'participants',
      component: Participants
    },
    {
      path: '/towerfall/:tournament/edit/',
      name: 'edit',
      component: Edit
    },
    {
      path: '/towerfall/:tournament/scores/',
      name: 'scores',
      component: ScoreScreen
    },
    {
      path: '/towerfall/:tournament/next/',
      name: 'next',
      component: NextScreen
    },
    {
      path: '/towerfall/:tournament/charts/',
      name: 'charts',
      component: PostMatch
    },
    {
      path: '/towerfall/:tournament/:kind/:match/',
      name: 'match',
      component: Match
    },
  ],
})

router.beforeEach((to, from, next) => {
  window.scrollTo(0, 0)

  // Why in the world does this need a short timeout? The connect
  // isn't set otherwise.
  setTimeout(function () {
    router.app.connect()
  }, 50)

  // Reset any pulsating lights
  document.getElementsByTagName("body")[0].className = ""

  // Always set up the user model from cookies
  // router.app.$set(router.app.$data, 'user', User.fromCookies(router.app.$cookie))

  next()
})

var Root = Vue.extend(App)
new Root({ // eslint-disable-line
  router: router
}).$mount("#app")
