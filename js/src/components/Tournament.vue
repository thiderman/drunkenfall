<template>
  <div>
    <tournament-preview v-if="!tournament.isStarted" :tournament="tournament" :user="user" :levels="levels"></tournament-preview>
    <tournament-overview v-if="tournament.isStarted" :tournament="tournament" :user="user" :levels="levels"></tournament-overview>
  </div>
</template>

<script>
import Tournament from '../models/Tournament'
import TournamentOverview from '../components/TournamentOverview'
import TournamentPreview from '../components/TournamentPreview'
import * as levels from "../models/Level"
import _ from 'lodash'

export default {
  name: 'Tournament',

  components: {
    TournamentOverview,
    TournamentPreview,
  },

  data () {
    return {
      tournament: new Tournament(),
      user: this.$root.user,
      levels: levels,
    }
  },

  computed: {
    runnerups: function () {
      let t = this.tournament

      if (!t.runnerups) {
        return []
      }

      return _.map(t.runnerups, (runnerupName) => {
        return _.find(t.players, { name: runnerupName })
      })
    }
  },

  methods: {
    start: function () {
      if (this.tournament) {
        this.api.start({ id: this.tournament.id }).then((res) => {
          console.log("start response:", res)
          let j = res.json()
          this.$route.router.go('/towerfall' + j.redirect)
        }, (err) => {
          console.error(`start for ${this.tournament} failed`, err)
        })
      } else {
        console.error("start called with no tournament")
      }
    },
    next: function () {
      if (this.tournament) {
        this.api.next({ id: this.tournament.id }).then((res) => {
          console.debug("next response:", res)
          let j = res.json()
          this.$route.router.go('/towerfall' + j.redirect)
        }, (err) => {
          console.error(`next for ${this.tournament} failed`, err)
        })
      } else {
        console.error("next called with no tournament")
      }
    }
  },

  created: function () {
    console.debug("Creating API resource")
    let customActions = {
      start: { method: "GET", url: "/api/towerfall{/id}/start/" },
      next: { method: "GET", url: "/api/towerfall{/id}/next/" }
    }
    this.api = this.$resource("/api/towerfall", {}, customActions)
  },

  route: {
    data ({ to }) {
      // listen for tournaments from App
      this.$on(`tournament${to.params.tournament}`, (tournament) => {
        console.debug("Received new tournament from App:", tournament)
        this.$set('tournament', tournament)
        // enable event to propagate
        return true
      })

      if (!_.isEmpty(this.$root.tournaments)) {
        let tournament = _.find(this.$root.tournaments, { id: to.params.tournament })
        if (tournament) {
          console.debug("Getting tournament from root:", tournament)
          this.$set('tournament', tournament)
        }
      }
    }
  }
}
</script>

<style lang="scss">
</style>
