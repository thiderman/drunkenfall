<template>
  <div>
    <header>
      <div class="content">
        <div class="title">
          {{tournament.name}} / {{match.title}} / Round {{round}}
        </div>
      </div>
      <div class="links" v-if="user.isJudge">
        <a v-if="match.canStart" @click="start">Start match</a>

        <a v-if="match.isRunning" @click="commit"
          v-bind:class="{'disabled': !can_commit}">End round</a>
        <a v-if="match.canEnd" @click="end">End match</a>

        <a v-if="match.isRunning" @click="reset"
          class="danger">Reset match</a>

      </div>
      <div class="clear"></div>
    </header>

    <div class="control" v-if="user.isJudge">
      <template v-for="(player, index) in match.players" ref="players">
        <control-player :index="index"></control-player>
      </template>
    </div>

    <div class="control" v-if="!user.isJudge">
      <template v-if="!match.isStarted" v-for="(player, index) in match.players" ref="players">
        <preview-player :index="index" :player="player" :match="match"></preview-player>
      </template>

      <template v-if="match.isStarted" v-for="(player, index) in match.players" ref="players">
        <live-player :index="index" :player="player" :match="match"></live-player>
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
import _ from 'lodash'

export default {
  name: 'Match',
  components: {
    ControlPlayer,
    PreviewPlayer,
    LivePlayer,
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
    match () {
      let kind = this.$route.params.kind
      let idx = this.$route.params.match

      if (kind === 'final') {
        return this.tournament.final
      }
      kind = kind + 's'
      return this.tournament[kind][idx]
    },
    can_commit: function () {
      return true
    },
    capitalizedKind: function () {
      return _.capitalize(this.match.kind)
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
    }
    this.api = this.$resource("/api/towerfall", {}, customActions)
  },
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
