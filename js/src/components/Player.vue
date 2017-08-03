<template>
  <div :class="'player ' + classes">
    <p>{{display_name}}</p>
  </div>
</template>

<script>
export default {
  name: 'Player',

  props: {
    player: Object,
    match: {},
    index: 0
  },

  computed: {
    avatar: function () {
      return this.player.avatar
    },
    display_name: function () {
      return this.player.person.nick
    },
    classes: function () {
      if (this.match.isEnded) {
        if (this.index === this.match.kill_order[0]) {
          return 'gold'
        } else if (this.index === this.match.kill_order[1]) {
          return 'silver'
        } else if (this.index === this.match.kill_order[2] && this.match.kind === 'final') {
          return 'bronze'
        }

        return 'out'
      }

      return this.player.color
    }
  }
}
</script>

<style lang="scss" scoped>
</style>
