<template>
  <div :class="'player ' + player.color">
    <div class="section">
      <h1>P{{index+1}}</h1>
    </div>

    <div class="section">
      <h2>{{player.kills}} <span>/ {{match.end}}</span></h2>
    </div>

    <div class="section">
      <h2>{{player.shots}} <span>{{player.shots | pluralize('shot')}}</span></h2>
    </div>

    <img :alt="player.displayName" :src="player.avatar"/>
  </div>
</template>

<script>
export default {
  name: 'LivePlayer',

  props: {
    player: {},
    match: {},
    index: 0,
  },

  created () {
    document.getElementsByTagName("body")[0].className = "scroll-less"
  }
}
</script>

<style lang="scss" scoped>
@import "../variables.scss";

.player {
  position: absolute;
  top: 0;
  bottom: 0;
  float: left;
  height: 100%;
  width: 25%;
  text-align: center;
  color: white;

  &:nth-child(1) {left: 0}
  &:nth-child(2) {left: 25%}
  &:nth-child(3) {left: 50%}
  &:nth-child(4) {left: 75%}

  .section {
    height: 21vh;
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
    position: absolute;
    bottom: 1.2%;
    right: 4%;
    left:  4%;
    display: block;
    object-fit: cover;
    width:  92%;
    max-height: 33vh;
    margin: 0px auto;
    box-shadow: -1px -1px 6px rgba(0,0,0,0.5);
    background-color: rgba(10,12,14,0.3);
  }
}

</style>
