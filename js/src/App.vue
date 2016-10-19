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
      reconnections: 0,
      user: new User(),
    }
  },
  methods: {
    reconnect: function () {
      this.$delete("ws")

      if (this.reconnections < 30) {
        console.warn(`closed uncleanly, reconnecting (try number ${this.reconnections})`)
        setTimeout(() => {
          this.reconnections = this.reconnections + 1
          this.connect()
        }, 500)
      } else {
        console.warn("Tried too many times, stopping.")
        this.reconnections = 0
      }
    },

    // (Re-)Connect the websocket.
    // Is safe to run when the connection is already up - then it will be a noop.
    connect: function () {
      if (!this.ws) {
        console.log('Setting up new websocket')
        this.$set('ws', new WebSocket('ws://' + window.location.host + '/api/towerfall/auto-updater'))

        let timeoutId

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
            if (timeoutId) {
              clearTimeout(timeoutId)
            }

            timeoutId = setTimeout(() => {
              if (this.ws && this.ws.readyState === 0) {
                // we're already trying to reconnect, don't try more
              } else {
                // the server probably died, try to connect again
                console.warn("The server probably died, try to reconnect a few times")
                this.reconnect()
              }
            }, 10000)
            console.debug("Ping")
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

        this.ws.onopen = () => {
          this.reconnections = 0
          console.debug("websocket connected:", this.ws)
        }
        this.ws.onerror = (errorEvent) => { console.error("websocket error:", errorEvent) }
        this.ws.onclose = (closeEvent) => {
          console.debug("websocket closed", closeEvent)
          if (!closeEvent.wasClean) {
            this.reconnect()
          }
        }
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
