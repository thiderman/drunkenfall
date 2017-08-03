<template>
  <router-link to="{ name: 'match', params: { kind: match.kind, match: match.index }}">
    <template v-for="x in 4">
      <player v-if="match.players.length > x" :player="match.players[x]" :match="match" :index="x"></player>

      <div v-if="match.players.length <= x" class="player prefill">
        <p>???</p>
      </div>
    </template>
  </router-link>
</template>

<script>
import Player from './Player.vue'

export default {
  name: 'MatchOverview',
  components: {
    Player
  },
  props: {
    match: {}
  }
}
</script>

<style lang="scss" scoped>
@import "../variables.scss";

.match {
  float: left;
  margin: 10px;
  cursor: pointer;
  text-align: center;
  box-shadow: 2px 2px 3px rgba(0,0,0,0.3);

  .player {
    float: left;
    width: 50%;
    height: 40%;
    line-height: 120%;
    text-align: center;
    overflow: hidden;
    white-space: nowrap;
    padding: 0.2em 0;
    text-shadow: 1px 1px 1px rgba(0,0,0,0.7);

    &:nth-child(1), &:nth-child(4) {
      background-color: #333;
      img {
        left: -20px;
      }
    }
    &:nth-child(2), &:nth-child(3) {
      background-color: #383838;
      img {
        right: -20px;
      }
    }

    img {
      object-fit: cover;
      border-radius: 100%;
      width:  60px;
      height: 60px;
      box-shadow: -1px -1px 6px rgba(0,0,0,0.5);
      background-color: rgba(10,12,14,0.3);
      position: absolute;
      top: 0;
    }

    &.prefill {
      color: #555;
    }

    &.green  { color: $green ; }
    &.blue   { color: $blue  ; }
    &.pink   { color: $pink  ; }
    &.orange { color: $orange; }
    &.white  { color: $white ; }
    &.yellow { color: $yellow; }
    &.cyan   { color: $cyan  ; }
    &.purple { color: $purple; }
    &.red    { color: $red; }

    &.gold {
      background-color: #daa520;
    }
    &.silver {
      background-color: #999;
    }
    &.bronze {
      background-color: #8C7853;
    }
    &.out {
      color: #777;
      text-shadow: 1px 1px 5px rgba(0,0,0,0.3);
    }
    &.out:nth-child(1), &.out:nth-child(4) {
      background-color: #433;
    }
    &.out:nth-child(2), &.out:nth-child(3) {
      background-color: #4a3838;
    }

  }

  img {
    width: 100%;
  }
}
</style>
