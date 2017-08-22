<template>
  <div>
    <header>
      <div class="content">
        <div class="title">
          {{tournament.name}} / Events
        </div>
      </div>
      <div class="links" v-if="user.level(levels.judge)">
      </div>
      <div class="clear"></div>
    </header>

    <basic-event v-for="event in events" :event="event"></basic-event>

    <div class="clear"></div>
  </div>
</template>

<script>
import BasicEvent from '../components/events/BasicEvent.vue'
import Event from '../models/Event.js'
import _ from 'lodash'

export default {
  name: 'Log',

  components: {
    BasicEvent,
  },

  computed: {
    user () {
      return this.$store.state.user
    },
    tournament () {
      return this.$store.getters.getTournament(
        this.$route.params.tournament
      )
    },
    events () {
      let events
      events = this.tournament.events
      _.each(this.tournament.tryouts, (tryout) => {
        events = _.concat(events, tryout.events)
      })
      _.each(this.tournament.semis, (semi) => {
        events = _.concat(events, semi.events)
      })
      events = _.concat(events, this.tournament.final.events)

      events = _.omitBy(events, _.isNil) // TODO(thiderman): Why are there nil items? hwat
      events = _.sortBy(events, [(o) => { return o.time }])
      events = _.reverse(events)
      events = _.map(events, Event.fromObject)

      this.$set('events', events)
    },
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
          console.log(res)
          this.setData(res.data.tournament)
        }, function (res) {
          console.log('error when getting tournament')
          console.log(res)
        })
      } else {
        // Something is set - we're clicking on a link and can reuse the
        // already existing data immediately
        this.setData(
          to.router.app.get(to.params.tournament),
          to.params.kind,
          parseInt(to.params.match)
        )
      }
    }
  }
}
</script>

<style lang="scss">
@import "../variables.scss";

</style>
