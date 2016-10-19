<template>
  <div>
    <header>
      <div class="content">
        <div class="title">Drunken TowerFall</div>
      </div>
      <div class="links">
        <a v-link="{name: 'new'}" v-if="user.level(levels.producer)">New Tournament</a>
        <a @click="clear" v-if="user.level(levels.producer)">Clear tests</a>
        <a v-link="{name: 'facebook'}" v-if="!user.authenticated">Facebook</a>
        <a href="/api/facebook/login" v-if="user.level(levels.producer) && user.authenticated">Re-facebook</a>
      </div>
      <div class="clear"></div>
    </header>

    <div class="tournaments" :class="{ loading: !tournament }">
      <div v-for="tournament in tournaments"
        :tournament="tournament.id" track-by="id">

        <a v-link="{ name: 'tournament', params: { tournament: tournament.id }}"
          :class="{ test: !tournament.name.startsWith('DrunkenFall')}">
          {{tournament.name}}
        </a>
      </div>
    </div>

    <h1 v-if="tournaments.length === 0">
      Tournaments will appear here soon! &lt;3
    </h1>
  </div>
</template>

<script>
import _ from "lodash"
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
      this.$on("tournament_list", (tournaments) => {
        this.$set("tournaments", tournaments)
        // enable event to propagate
        return true
      })

      if (!_.isEmpty(this.$root.tournaments)) {
        console.debug("Getting tournaments from root")
        this.$set('tournaments', this.$root.tournaments)
      }
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
  background-color: #405060;
  color: #dbdbdb;
  display: block;
  font-size: 2.5em;
  font-weight: bold;
  padding: 1% 3%;
  text-align: center;
  text-decoration: none;
  width: 60%;
  margin: 0.2em auto;

  &.test {
    background-color: #353535;
    font-size: 1.5em;
    width: 40%;
  }
}
</style>
