<template>
  <div>
    <header>
      <div class="content">
        <div class="title">
          {{tournament.name}} / {{match.kind | capitalize}} {{match.index +1}}
        </div>
      </div>
      <div class="links">
        <p class="time">
          Local time: {{clock}}
        </p>
      </div>
      <div class="clear"></div>
    </header>

    <h1>Starting in</h1>
    <div class="timer">
      {{timer}}
    </div>

    <div class="players">
      <template v-for="player in playersReversed" v-ref:players>
        <preview-player :index="$index + 1" :player="player" :match="match">
      </template>
    </div>
  </div>
</template>

<script>
import PreviewPlayer from './PreviewPlayer.vue'
import Match from '../models/Match.js'
import Tournament from '../models/Tournament.js'
import moment from 'moment'
import _ from 'lodash'

export default {
  name: 'NextScreen',
  components: {
    PreviewPlayer,
  },

  data () {
    return {
      match: new Match(),
      tournament: new Tournament(),
      timer: "00:00",
      clock: "00:00 (+02:00)",
      intervalID: 0,
    }
  },

  computed: {
    playersReversed: function () {
      return _.reverse(this.match.players)
    }
  },

  methods: {
    setData: function (tournament) {
      let kind = tournament.current.kind
      let index = tournament.current.index

      if (kind === 'final') {
        this.$set('match', Match.fromObject(tournament[kind]))
      } else {
        this.$set('match', Match.fromObject(tournament[kind][index]))
      }

      this.$set('tournament', Tournament.fromObject(tournament))
    }
  },

  created: function () {
    console.debug("Creating API resource")
    let customActions = {
      getTournamentData: { method: "GET", url: "/api/towerfall/tournament{/id}/" }
    }
    this.api = this.$resource("/api/towerfall", {}, customActions)

    // Update the clock
    setInterval(() => {
      this.$set("clock", moment().format("HH:mm (Z)"))
    }, 1000)

    // XXX(thiderman): Super duplicated from TournamentPreview.vue
    this.$watch('tournament', (t) => {
      // If there is already a clock ticking, kill it.
      if (this.intervalID !== 0) {
        clearInterval(this.intervalID)
      }

      var eventTime = this.match.scheduled.unix()
      var currentTime = moment().unix()
      var diffTime = eventTime - currentTime
      var d = moment.duration(diffTime, 'seconds') // duration
      var interval = 1000

      function pad (n, width) {
        n = n + ''
        return n.length >= width ? n : new Array(width - n.length + 1).join("0") + n
      }

      this.intervalID = setInterval(() => {
        d = moment.duration(d - interval, 'milliseconds')

        // NOTE: Due to this messing with the chroma-key on stream, we're removing this! Byeeeeeeeeeeeeeeeee!
        // During the last minute, make sure to add the pulse class.
        // Do so for every second, so that reloads will make sense as well.
        // if (d.hours() === 0 && d.minutes() === 0) {
        //   document.getElementsByTagName("body")[0].className = "red-pulse"
        // }

        // If we're ever at a negative interval, stop immediately.
        // Technically we probably only really need the seconds here, but
        // if we use all of them any future cases will be fixed immediately.
        if (_.some([d.hours(), d.minutes(), d.seconds()], (n) => n < 0)) {
          console.log("Closing interval.")
          document.getElementsByTagName("body")[0].className = ""
          clearInterval(this.intervalID)
          return
        }

        this.$set(
          'timer',
          pad(d.minutes(), 2) + ":" +
          pad(d.seconds(), 2)
        )
      }, interval)
    })
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

<style lang="scss" scoped>
@import "../variables.scss";
@import "../ribbon.scss";

.players {
  width: 100%;

  .player {
    width: 25%;
    display: block;
    float: left;
  }
}

h1 {
  margin-top: 100px;
  margin-bottom: -1em;
}

.timer {
  margin: 0 auto 0.25em;
  width: 3em;
  font-size: 12em;
  text-align: center;
  padding: 0.08em 0.4em;

  text-shadow: 3px 3px 3px rgba(0,0,0,0.7);
}

.time {
  font-size: 1.5em;
  padding: 16px 40px;
}

</style>
