<template>
  <div v-if="tournament">
    <header v-if="user.authenticated">
      <div class="content">
        <div class="title">{{tournament.name}}</div>
      </div>
      <div class="links">
        <a v-if="tournament.canStart && user.isCommentator" @click="start">Start</a>
        <a v-if="tournament.canStart && user.isProducer" @click="usurp">Usurp</a>

        <router-link v-if="user.isPlayer"
          :to="{ name: 'join', params: { tournament: tournament.id }}">
          Join
        </router-link>
        <router-link v-if="user.isJudge"
          :to="{ name: 'participants', params: { tournament: tournament.id }}">
          Participants
        </router-link>
      </div>
      <div class="clear"></div>
    </header>

    <h1>
      Starting soon
    </h1>

    <div class="players">
      <div v-for="player in tournament.players" class="player">
        <img :alt="player.person.nick" :src="player.avatar"/>
      </div>
      <div class="clear"></div>
    </div>

    <div class="protector">
      <div class="super-ribbon">
        drunkenfall.com
      </div>

      <div class="ribbon">
        <strong class="ribbon-content">
          {{ countdown.time }}
        </strong>
      </div>
    </div>
  </div>
</template>

<script>
import {Countdown} from '../models/Timer.js'

export default {
  name: 'TournamentPreview',

  data () {
    return {
      countdown: new Countdown(),
    }
  },

  computed: {
    tournament () {
      let t = this.$store.getters.getTournament(
        this.$route.params.tournament
      )
      this.countdown.start(t.scheduled)
      return t
    },
    user () {
      return this.$store.state.user
    },
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
    this.api = this.$resource("/api/towerfall", {}, {
      start: { method: "GET", url: "/api/towerfall{/id}/start/" },
      usurp: { method: "GET", url: "/api/towerfall{/id}/usurp/" },
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
