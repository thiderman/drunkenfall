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
          Local time: 00:03 CET
        </p>
      </div>
      <div class="clear"></div>
    </header>

    <h1>Starting in</h1>
    <div class="timer">
      {{timer}}
    </div>

    <div class="players">
      <template v-for="player in match.players" v-ref:players>
        <preview-player :index="$index + 1" :player="player" :match="match">
      </template>
    </div>
  </div>
</template>

<script>
import PreviewPlayer from './PreviewPlayer.vue'
import Match from '../models/Match.js'
import Tournament from '../models/Tournament.js'

export default {
  name: 'NextScreen',
  components: {
    PreviewPlayer,
  },

  data () {
    return {
      match: new Match(),
      tournament: new Tournament(),
      timer: "04:36",
    }
  },

  methods: {
    setData: function (tournament) {
      let kind = tournament.current.kind
      let index = tournament.current.index
      if (kind === 'tryout') {
        kind = 'tryouts'
      } else if (kind === 'semi') {
        kind = 'semis'
      }

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
