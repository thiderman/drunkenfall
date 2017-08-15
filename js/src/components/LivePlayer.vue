<template>
  <div :class="'player ' + player.color">
    <div class="section">
      <h1>P{{index}}</h1>
    </div>

    <div class="section">
      <h2>{{player.kills}} <span>/ {{match.length}}</span></h2>
      <h3>kills</h3>
    </div>

    <div class="section">
      <h2>{{player.shots}}</h2>
      <h3>shots</h3>
    </div>

    <div class="section">
      <img :alt="player.displayName" :src="player.avatar"/>
    </div>
  </div>
</template>

<script>
export default {
  name: 'ControlPlayer',

  props: {
    player: {},
    match: {},
    index: 0,
  },

  computed: {
    classes: function () {
      if (!this.match.isEnded) {
        if (this.index === 0) {
          return 'gold'
        } else if (this.index === 1) {
          return 'silver'
        } else if (this.index === 2 && this.match.kind === 'final') {
          return 'bronze'
        }

        return 'out'
      }

      return this.player.preferred_color
    }
  },

  created: function () {
    document.getElementsByTagName("body")[0].className = "scroll-less"
  }
}
</script>

<style lang="scss" scoped>
@import "../variables.scss";

.player {
  float: left;
  height: 100%;
  width: 25%;
  text-align: center;
  color: white;

  .section {
    height: 25%;
  }

  &.green  {background-color: $green-bg;}
  &.blue   {background-color: $blue-bg;}
  &.pink   {background-color: $pink-bg;}
  &.orange {background-color: $orange-bg;}
  &.white  {background-color: $white-bg;}
  &.yellow {background-color: $yellow-bg;}
  &.cyan   {background-color: $cyan-bg;}
  &.purple {background-color: $purple-bg;}
  &.red    {background-color: $red-bg;}

  h1 {
    font-size: 17vh;
    margin: 0.1em 0 0 0;
    text-shadow: 1vh 1vh 1vh rgba(0,0,0,0.7);
  }

  h2 {
    font-size: 13vh;
    margin: 0 0 0 0;
    text-shadow: 5px 5px 5px rgba(0,0,0,0.7);

    span {
      font-size: 0.4em;
    }
  }

  h3 {
    font-size: 6vh;
    margin: 0em 0em 0em 0em;
    text-shadow: 5px 5px 5px rgba(0,0,0,0.7);
  }

  img {
    display: block;
    object-fit: cover;
    width:  20vw;
    height: 20vw;
    margin: 0px auto;
    box-shadow: -1px -1px 6px rgba(0,0,0,0.5);
    background-color: rgba(10,12,14,0.3);
  }
}

</style>
