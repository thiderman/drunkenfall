<template>
  <div>
    <template v-for="(player, index) in match.players" ref="players">
      <live-player :index="index + 1" :player="player" :match="match"></live-player>
    </template>
    <div class="clear"></div>
  </div>
</template>

<script>
import LivePlayer from './LivePlayer.vue'
import Match from '../models/Match.js'
import Tournament from '../models/Tournament.js'

export default {
  name: 'ScoreScreen',
  components: {
    LivePlayer,
  },

  data () {
    return {
      match: new Match(),
      tournament: new Tournament(),
      user: this.$root.user,
    }
  },

  methods: {
    setData: function (tournament) {
      console.log("setData tournament", tournament)
      let kind = tournament.current.kind
      let index = tournament.current.index
      let match
      if (kind === 'tryout') {
        kind = 'tryouts'
      } else if (kind === 'semi') {
        kind = 'semis'
      }

      if (kind === 'final') {
        match = Match.fromObject(tournament[kind])
      } else {
        match = Match.fromObject(tournament[kind][index])
      }

      // HACK(thiderman): So, this is pretty nasty, but it works. We don't
      // want this screen to update to the new match until it is started, but
      // we also at the same time want to show the /next/ screen with the next
      // match data. To avoid this, we just simply don't set the data on this
      // until the match has started.
      // This will break if we have to reload the page since there will be no
      // previous state.
      if (match.isStarted) {
        this.$set('match', match)
        this.$set('tournament', Tournament.fromObject(tournament))
      }
    }
  },

  created: function () {
    console.debug("Creating API resource")
    let customActions = {
      getTournamentData: { method: "GET", url: "/api/towerfall/tournament{/id}/" }
    }
    this.api = this.$resource("/api/towerfall", {}, customActions)
  },

  route: {
    data ({ to }) {
      // listen for tournaments from App
      this.$on(`tournament${to.params.tournament}`, (tournament) => {
        console.debug("New tournament from App:", tournament)
        this.setData(tournament)
      })

      if (to.router.app.tournaments.length === 0) {
        // Nothing is set - we're reloading the page and we need to get the
        // data manually
        this.api.getTournamentData({ id: to.params.tournament }).then(function (res) {
          this.setData(
            res.data.tournament,
          )
        }, function (res) {
          console.log('error when getting tournament')
          console.log(res)
        })
      } else {
        // Something is set - we're clicking on a link and can reuse the
        // already existing data immediately
        this.setData(
          to.router.app.get(to.params.tournament),
        )
      }
    }
  }
}
</script>

<style lang="scss" >

control {
  height: 85vh;
  padding: 0.8%;
}

.player {
  height: 25%;
  display: block;
}

</style>
