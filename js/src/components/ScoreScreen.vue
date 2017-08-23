<template>
  <div v-if="tournament">
    <template v-for="(player, index) in match.players" ref="players">
      <live-player :index="index" :player="player" :match="match"></live-player>
    </template>
    <div class="clear"></div>
  </div>
</template>

<script>
import LivePlayer from './LivePlayer.vue'

export default {
  name: 'ScoreScreen',
  components: {
    LivePlayer,
  },

  computed: {
    tournament () {
      return this.$store.getters.getTournament(
        this.$route.params.tournament
      )
    },
    // TODO: We're missing the old hack that only set the match once
    // it was started. Right now, this will move as soon as the match
    // is ended. Presumably this will be fixed once the tournament
    // match structure is flattened.
    match () {
      let kind = this.tournament.current.kind
      let idx = this.tournament.current.index

      if (kind === 'final') {
        return this.tournament.final
      }
      kind = kind + 's'
      return this.tournament[kind][idx]
    },
  },
}
</script>

<style lang="scss" >

control {
  height: 85vh;
  padding: 0.8%;
}

.player {
  height: 25%;
  display: block;
}

</style>
