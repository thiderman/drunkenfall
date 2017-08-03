<template>
  <div>
    <header v-if="user.authenticated">
      <div class="content">
        <div class="title">{{tournament.name}}</div>
      </div>
      <div class="links">
        <a v-if="tournament.canStart && user.level(30)" @click="start">Start</a>
        <a v-if="tournament.canStart && user.level(100)" @click="usurp">Usurp</a>
        <a v-if="user.level(10)" v-link="{ name: 'join', params: { tournament: tournament.id }}">Join</a>
        <a v-if="user.level(50)" v-link="{ name: 'participants', params: { tournament: tournament.id }}">Participants</a>
      </div>
      <div class="clear"></div>
    </header>

    <h1>
      Starting soon
    </h1>

    <div class="players">
      <div v-for="player in tournament.players" class="player">
        <img alt="{{player.person.nick}}" :src="player.avatar"/>
      </div>
      <div class="clear"></div>
    </div>

    <div class="protector">
      <div class="super-ribbon">
        drunkenfall.com
      </div>

      <div class="ribbon">
        <strong class="ribbon-content">
          {{ countdown }}
        </strong>
      </div>
    </div>
  </div>
</template>

<script>
import Tournament from '../models/Tournament'
import User from '../models/User'
import * as levels from "../models/Level"
import moment from 'moment'
import _ from 'lodash'

export default {
  name: 'TournamentPreview',

  props: {
    tournament: new Tournament(),
    user: new User(),
    levels: levels,
    countdown: "00:00:00",
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
    usurp: function () {
      if (this.tournament) {
        this.api.usurp({ id: this.tournament.id }).then((res) => {
          console.log("usurp response:", res)
          let j = res.json()
          this.$route.router.push('/towerfall' + j.redirect)
        }, (err) => {
          console.error(`usurp for ${this.tournament} failed`, err)
        })
      } else {
        console.error("usurp called with no tournament")
      }
    },
    join () {
      this.api.join({ id: this.tournament.id }).then((res) => {
        console.log("join response:", res)
        var j = res.json()
        this.$route.router.push('/towerfall' + j.redirect)
      }, (err) => {
        console.error(`joining tournament ${this.tournament} failed`, err)
      })
    }
  },

  created: function () {
    // API definitions
    console.debug("Creating API resource")
    let customActions = {
      start: { method: "GET", url: "/api/towerfall{/id}/start/" },
      usurp: { method: "GET", url: "/api/towerfall{/id}/usurp/" },
      getData: { method: "GET", url: "/api/towerfall/tournament/{id}/" }
    }
    this.api = this.$resource("/api/towerfall", {}, customActions)

    // Also create the clock countdown
    this.$set('countdown', '00:00:00')
    this.$watch('tournament', (newVal) => {
      var eventTime = newVal.scheduled.unix()
      var currentTime = moment().unix()
      var diffTime = eventTime - currentTime
      var d = moment.duration(diffTime, 'seconds') // duration
      var interval = 1000
      var intervalId = 0

      function pad (n, width) {
        n = n + ''
        return n.length >= width ? n : new Array(width - n.length + 1).join("0") + n
      }

      intervalId = setInterval(() => {
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
          clearInterval(intervalId)
          return
        }

        this.$set(
          'countdown',
          pad(d.hours(), 2) + ":" +
          pad(d.minutes(), 2) + ":" +
          pad(d.seconds(), 2)
        )
      }, interval)
    })
  }
}
</script>

<style lang="scss" scoped>
@import "../variables.scss";
@import "../ribbon.scss";

h1 {
  font-size: 6em;
  margin-top: 0;
  margin-bottom: 0.4em;
  padding-top: 0.2em;
  text-shadow: 5px 5px 10px rgba(0,0,0,0.7);
}

h2 {
  margin: -1.5em 0 1em;
  font-size: 2.5em;
  text-align: center;
  strong {
    text-shadow: 2px 2px 4px rgba(0,0,0,0.7);
    color: #a090a0;
  }
}

.ribbon {
  font-size: 40px;
}
.super-ribbon {
  margin: -3em auto 2.5em;
}

.players {
  text-align: center;
  width: 80%;
  margin: 100px auto;

  .player {
    display: inline-block;
    width: 130px;
    margin-top: -30px;

    img {
      object-fit: cover;
      border-radius: 100%;
      width:  150px;
      height: 150px;
      box-shadow: -1px -1px 6px rgba(0,0,0,0.5);
      background-color: rgba(10,12,14,0.3);
    }
    .ribbon {
      width: 88%;
    }
  }
}
</style>
