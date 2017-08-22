<template>
  <div v-if="tournament">
    <header>
      <div class="content">
        <div class="title">
          {{tournament.name}} / Events
        </div>
      </div>
      <div class="links" v-if="user.isJudge">
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

      return events
    }
  },

  created: function () {
    console.debug("Creating API resource")
    let customActions = {
      getTournamentData: { method: "GET", url: "/api/towerfall/tournament{/id}/" }
    }
    this.api = this.$resource("/api/towerfall", {}, customActions)
  },
}
</script>

<style lang="scss">
@import "../variables.scss";

</style>
