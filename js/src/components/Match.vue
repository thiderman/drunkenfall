<template>
  <div>
    <header>
      <div class="content">
        <div class="title">
          {{tournament.name}} / {{match.kind | capitalize}} {{match.index +1}} / Round {{match.commits.length + 1}}
        </div>
      </div>
      <div class="links" v-if="user.level(levels.judge)">
        <a v-if="match.canStart" @click="start">Start match</a>

        <a v-if="match.isRunning" @click="commit"
          v-bind:class="{'disabled': !can_commit}">End round</a>
        <a v-if="match.canEnd" @click="end">End match</a>

        <a v-if="match.isRunning" @click="reset"
          class="danger">Reset match</a>

      </div>
      <div class="clear"></div>
    </header>

    <div class="control" v-if="user.level(levels.judge)">
      <template v-for="player in match.players" v-ref:players>
        <control-player :index="$index" :player="player" :match="match"
                        :downs="0" :ups="0">
      </template>
    </div>

    <div class="control" v-if="!user.level(levels.judge)">
      <template v-if="!match.isStarted" v-for="player in match.players" v-ref:players>
        <preview-player :index="$index" :player="player" :match="match">
      </template>

      <template v-if="match.isStarted" v-for="player in match.players" v-ref:players>
        <live-player :index="$index + 1" :player="player" :match="match">
      </template>
    </div>

    <div class="clear"></div>
  </div>
</template>

<script>
import ControlPlayer from './ControlPlayer.vue'
import PreviewPlayer from './PreviewPlayer.vue'
import LivePlayer from './LivePlayer.vue'
import Match from '../models/Match.js'
import Tournament from '../models/Tournament.js'
import * as levels from "../models/Level.js"
import _ from 'lodash'

export default {
  name: 'Match',
  components: {
    ControlPlayer,
    PreviewPlayer,
    LivePlayer,
  },

  data () {
    return {
      match: new Match(),
      tournament: new Tournament(),
      user: this.$root.user,
      levels: levels,
    }
  },

  computed: {
    can_commit: function () {
      return true
    }
  },

  methods: {
    commit: function () {
      // TODO this could potentially be a class
      let payload = {
        'state': _.map(this.$refs.players, (controlPlayer) => {
          return _.pick(controlPlayer, ['ups', 'downs', 'shot', 'reason'])
        })
      }

      console.log(payload)
      this.api.commit({ id: this.tournament.id, kind: this.match.kind, index: this.match.index }, payload).then(function (res) {
        console.log(res)
        let data = res.json()
        this.$set('match', Match.fromObject(data.match))

        _.each(this.$refs.players, (controlPlayer) => { controlPlayer.reset() })
      }, function (res) {
        console.log('error when setting score')
        console.log(res)
      })
    },

    refresh: function () {
      // Hax to make vue refresh the entire page.
      // Since nothing on this page is properly bound to components right now
      // the updates won't trigger properly.
      this.$set('updated', Date.now())
    },
    end: function () {
      this.api.end({ id: this.tournament.id, kind: this.match.kind, index: this.match.index }).then(function (res) {
        console.log(res)
        this.$route.router.push('/towerfall/' + this.tournament.id + '/')
      }, function (res) {
        console.log('error when getting tournament')
        console.log(res)
      })
    },
    start: function () {
      this.api.start({ id: this.tournament.id, kind: this.match.kind, index: this.match.index }).then(function (res) {
        console.log(res)
        this.setData(
          res.data.tournament,
          this.match.kind,
          this.match.index
        )
      }, function (res) {
        console.log('error when getting tournament')
        console.log(res)
      })
    },
    reset: function () {
      this.api.reset({ id: this.tournament.id, kind: this.match.kind, index: this.match.index }).then(function (res) {
        console.log(res)
        this.setData(
          res.data.tournament,
          this.match.kind,
          this.match.index
        )
      }, function (res) {
        console.log('error when getting tournament')
        console.log(res)
      })
    },
    setData: function (tournament, kind, match) {
      if (kind === 'tryout') {
        kind = 'tryouts'
      } else if (kind === 'semi') {
        kind = 'semis'
      }

      if (kind === 'final') {
        this.$set('match', Match.fromObject(tournament[kind]))
      } else {
        this.$set('match', Match.fromObject(tournament[kind][match]))
      }

      this.$set('tournament', Tournament.fromObject(tournament))
    }
  },

  created: function () {
    console.debug("Creating API resource")
    let customActions = {
      commit: { method: "POST", url: "/api/towerfall/tournament{/id}{/kind}{/index}/commit/" },
      start: { method: "GET", url: "/api/towerfall/tournament{/id}{/kind}{/index}/start/" },
      end: { method: "GET", url: "/api/towerfall/tournament{/id}{/kind}{/index}/end/" },
      reset: { method: "GET", url: "/api/towerfall/tournament{/id}{/kind}{/index}/reset/" },
      getTournamentData: { method: "GET", url: "/api/towerfall/tournament{/id}/" }
    }
    this.api = this.$resource("/api/towerfall", {}, customActions)
  },

  route: {
    data ({ to }) {
      // listen for tournaments from App
      this.$on(`tournament${to.params.tournament}`, (tournament) => {
        console.debug("New tournament from App:", tournament)
        this.setData(tournament, to.params.kind, parseInt(to.params.match))
      })

      if (to.router.app.tournaments.length === 0) {
        // Nothing is set - we're reloading the page and we need to get the
        // data manually
        this.api.getTournamentData({ id: to.params.tournament }).then(function (res) {
          console.log(res)
          let data = res.json()
          this.setData(
            data.tournament,
            to.params.kind,
            parseInt(to.params.match)
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
          to.params.kind,
          parseInt(to.params.match)
        )
      }
    }
  }
}
</script>

<style lang="scss" >

.control {
  height: 85vh;
  padding: 0.8%;
}

.player {
  height: 25%;
  display: block;
}

</style>
