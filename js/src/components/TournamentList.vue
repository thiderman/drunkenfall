<template>
  <div>
    <header>
      <div class="content">
        <div class="title">Drunken TowerFall</div>
      </div>
      <div class="links">
        <router-link to="{name: 'new'}" v-if="user.level(levels.producer)">New Tournament</router-link>
        <a @click="clear" v-if="user.level(levels.producer)">Clear tests</a>
        <router-link to="{name: 'facebook'}" v-if="!user.authenticated">Facebook</router-link>
        <a href="/api/facebook/login" v-if="user.level(levels.producer) && user.authenticated">Re-facebook</a>
      </div>
      <div class="clear"></div>
    </header>

    <div class="tournaments" :class="{ loading: !tournament }">
      <div v-for="tournament in tournaments"
        :tournament="tournament.id" track-by="id">

        <router-link to="{ name: 'tournament', params: { tournament: tournament.id }}"
          :class="{ test: tournament.isTest, current: !tournament.isStest && !tournament.isStarted}">
          {{tournament.name}}
        </router-link>
      </div>
    </div>

    <h1 v-if="tournaments.length === 0">
      Tournaments will appear here soon! &lt;3
    </h1>
  </div>
</template>

<script>
import _ from "lodash"
import Tournament from "../models/Tournament.js"
import * as levels from "../models/Level.js"

export default {
  name: 'TournamentList',

  data () {
    return {
      tournaments: [],
      user: this.$root.user,
      levels: levels,
    }
  },
  methods: {
    clear (event) {
      event.preventDefault()
      return this.$http.get('/api/towerfall/tournament/clear/').then(function (res) {
        this.$set('tournaments', _.map(res.data.tournaments, Tournament.fromObject))
      }, function (res) {
        console.error('error when clearing tournaments', res)
        return { tournaments: [] }
      })
    }
  },
  route: {
    data ({ to }) {
      return this.$http.get('/api/towerfall/tournament/').then(function (res) {
        let data = JSON.parse(res.data)
        return {
          tournaments: _.map(data.tournaments, Tournament.fromObject)
        }
      }, function (res) {
        console.error('error when getting tournaments', res)
        return { tournaments: [] }
      })
    }
  }
}
</script>

<style lang="scss" scoped>

h1 {
  text-align: center;
  text-shadow: 3px 3px 3px rgba(0,0,0,0.7);
  margin: 2em;
  font-size: 3em;
}

.tournaments a {
  text-shadow: 2px 2px 1px rgba(0,0,0,0.7);
  background-color: #454545;
  color: #dbdbdb;
  display: block;
  font-size: 2.5em;
  font-weight: bold;
  padding: 1% 3%;
  text-align: center;
  text-decoration: none;
  width: 40%;
  margin: 0.2em auto;

  &.test {
    background-color: #353535;
    font-size: 1.2em;
    width: 30%;
  }

  &.current {
    background-color: #405060;
    width: 60%;
    font-size: 3.5em;
  }
}
</style>
