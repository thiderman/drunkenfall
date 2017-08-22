<template>
  <div>
    <tournament-preview v-if="tournament && !tournament.isStarted"></tournament-preview>
    <tournament-overview v-if="tournament && tournament.isStarted"></tournament-overview>
  </div>
</template>

<script>
import TournamentOverview from '../components/TournamentOverview'
import TournamentPreview from '../components/TournamentPreview'
import _ from 'lodash'

export default {
  name: 'TournamentView',

  components: {
    TournamentOverview,
    TournamentPreview,
  },

  data () {
    return {
      id: null,
    }
  },

  computed: {
    tournament () {
      return this.$store.getters.getTournament(
        this.$route.params.tournament
      )
    },
    user () {
      return this.$store.state.user
    },
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
          this.$route.router.push('/towerfall' + j.redirect)
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
          this.$route.router.push('/towerfall' + j.redirect)
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
      next: { method: "GET", url: "/api/towerfall{/id}/next/" },
      getData: { method: "GET", url: "/api/towerfall/tournament{/id}/" }
    }
    this.api = this.$resource("/api/towerfall", {}, customActions)
  },
}
</script>

<style lang="scss">
</style>
