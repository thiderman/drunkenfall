<template>
  <router-view
    keep-alive
    transition>
  </router-view>
</template>

<script>
/* eslint-env browser */
import Tournament from "./models/Tournament.js"
import _ from 'lodash'
import User from './models/User.js'

export default {
  data () {
    return {
      // All the tournaments
      // TODO(thiderman): It's practically pointless to always update all of them
      // but this was the quickest way. In the future, we can probably just update
      // either the match, or just the running tournament.
      tournaments: [],
      // The main websocket object
      ws: null,
      user: new User(),
    }
  },
  methods: {
    // (Re-)Connect the websocket.
    // Is safe to run when the connection is already up - then it will be a noop.
    connect: function () {
      if (!this.ws) {
        console.log('Setting up new websocket')
        this.$set('ws', new WebSocket('ws://' + window.location.host + '/api/towerfall/auto-updater'))

        this.ws.onmessage = (event) => {
          let res
          try {
            res = JSON.parse(event.data)
          } catch (e) {
            console.error("Failed to parse message data:", e)
            return
          }

          // If p is set, this is a ping message that only serves to keep the connection open.
          // Break immediately.
          if (res.p !== undefined) {
            return
          }

          console.log(event)
          if (res.data) {
            if (res.data.tournaments) {
              // The main bulk update. This contains the latest state.
              console.log('Updating tournaments')
              let tournaments = _.map(res.data.tournaments, Tournament.fromObject)
              this.$set('tournaments', tournaments)

              // first broadcast the entire list, in case the page needs it (mainly TournamentList)
              this.$broadcast("tournament_list", tournaments)
              // then, broadcast each tournament separately, to simplify the logic of most pages
              _.each(tournaments, (tournament) => { this.$broadcast(`tournament${tournament.id}`, tournament) })
              return
            }

            console.log('no tournaments received, did not update anything')
            return
          }

          console.log('Unknown websocket update:', res)
        }

        this.ws.onopen = () => { console.debug("websocket connected:", this.ws) }
        this.ws.onerror = (errorEvent) => { console.error("websocket error:", errorEvent) }
        this.ws.onclose = (closeEvent) => { console.warn("websocket closed", closeEvent) }
      }
    },

    get: function (tid) {
      return _.find(this.tournaments, { id: tid })
    }
  }
}
</script>

<style lang="scss">
@import "./style.scss";

@font-face { font-family: Archer; src: url('/static/Archer.ttf'); }

</style>
