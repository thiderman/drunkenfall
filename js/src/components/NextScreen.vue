<template>
  <div v-if="tournament">
    <header>
      <div class="content">
        <div class="title">
          {{tournament.name}} / {{match.title}}
        </div>
      </div>
      <div class="links">
        <p class="time">
          Local time: {{clock.time}}
        </p>
      </div>
      <div class="clear"></div>
    </header>

    <h1>Starting in</h1>
    <div class="timer">
      {{countdown.time}}
    </div>

    <div class="players">
      <template v-for="(player, index) in playersReversed" ref="players">
        <preview-player :index="index + 1" :player="player" :match="match"></preview-player>
      </template>
    </div>
  </div>
</template>

<script>
import PreviewPlayer from './PreviewPlayer.vue'
import {Countdown, Clock} from '../models/Timer'
import _ from 'lodash'

export default {
  name: 'NextScreen',
  components: {
    PreviewPlayer,
  },

  data () {
    return {
      countdown: new Countdown(),
      clock: new Clock(),
    }
  },

  computed: {
    tournament () {
      return this.$store.getters.getTournament(
        this.$route.params.tournament
      )
    },
    match () {
      let kind = this.tournament.current.kind
      let idx = this.tournament.current.index
      let match

      if (kind === 'final') {
        match = this.tournament.final
      } else {
        kind = kind + 's'
        match = this.tournament[kind][idx]
      }

      this.countdown.start(match.scheduled)
      this.clock.start()

      return match
    },
    playersReversed () {
      return _.reverse(this.match.players)
    },
  },
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
