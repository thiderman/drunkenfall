/* eslint-env browser */

import Vue from 'vue'
import Vuex from 'vuex'
import Router from 'vue-router'
import Resource from 'vue-resource'
import Cookie from 'vue-cookie'
import Vue2Filters from 'vue2-filters'
import App from './App'

import _ from 'lodash'

import TournamentList from './components/TournamentList.vue'
import TournamentView from './components/Tournament.vue'
import New from './components/New.vue'
import Join from './components/Join.vue'
import Participants from './components/Participants.vue'
import Runnerups from './components/Runnerups.vue'
import Edit from './components/Edit.vue'
import Match from './components/Match.vue'
import PostMatch from './components/PostMatch.vue'
import Log from './components/Log.vue'
import ScoreScreen from './components/ScoreScreen.vue'
import NextScreen from './components/NextScreen.vue'
import Facebook from './components/Facebook.vue'
import FacebookFinalize from './components/FacebookFinalize.vue'
import Credits from './components/Credits.vue'

import User from './models/User.js'
import {Credits as CreditsModel} from './models/Credits.js'
import Tournament from './models/Tournament.js'

// install packages
Vue.use(Vuex)
Vue.use(Router)
Vue.use(Resource)
Vue.use(Cookie)
Vue.use(Vue2Filters)

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
      component: TournamentView
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
      path: '/towerfall/:tournament/runnerups/',
      name: 'runnerups',
      component: Runnerups
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
      path: '/towerfall/:tournament/log/',
      name: 'log',
      component: Log
    },
    {
      path: '/towerfall/:tournament/credits/',
      name: 'credits',
      component: Credits
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

    // Always set up the user model from cookies
    router.app.$store.commit('setUser', User.fromCookies(router.app.$cookie))
  }, 50)

  // Reset any pulsating lights
  document.getElementsByTagName("body")[0].className = ""
  next()
})

const store = new Vuex.Store({ // eslint-disable-line
  state: {
    tournaments: [],
    user: new User(),
    credits: {}
  },
  mutations: {
    updateAll (state, tournaments) {
      state.tournaments = _.map(tournaments, Tournament.fromObject)
    },
    setUser (state, user) {
      state.user = user
    },
    setCredits (state, credits) {
      state.credits = CreditsModel.fromObject(credits)
    },
  },
  getters: {
    getTournament: (state, getters) => (id) => {
      return state.tournaments.find(t => t.id === id)
    }
  }
})

var Root = Vue.extend(App)
new Root({ // eslint-disable-line
  router: router,
  store: store,
}).$mount("#app")
