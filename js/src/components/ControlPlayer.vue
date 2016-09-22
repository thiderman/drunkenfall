<template>
  <div class="player {{player.preferred_color}}">
    <div class="button" @click="score(index, -1)">
      <div><p>-</p></div>
    </div>

    <div class="shots">
       <div @click="manual_shot" v-bind:class="{'give': shot}">
         <div v-if="shot" class="mark">{{player.shots + 1}}</div>
         <div v-if="!shot" class="mark">{{player.shots}}</div>
         <div v-if="reason !== ''" class="reason">{{reason}}</div>
       </div>
     </div>

     <div class="slider {{player.color}}" @click="reset">
       <div><p>{{player.displayName}}</p></div>
     </div>

     <div v-if="match.length == 10" class="scores">
       <div v-for="n in 10"
         class="{{bullet_class(player, index, n+1)}}">
         <p>{{n+1}}</p>
       </div>
     </div>

     <div v-if="match.length == 20" class="scores final">
       <div v-for="n in 20"
         class="{{bullet_class(player, index, n+1)}}">
         <p>{{n+1}}</p>
       </div>
     </div>

     <div class="button" @click="score(index, 1)">
       <div><p>+</p></div>
     </div>
   </div>
</template>

<script>
import Match from "../models/Match.js"

export default {
  name: 'ControlPlayer',

  props: {
    player: {},
    match: {
      coerce: (val) => { return Match.fromObject(val) }
    },
    index: 0,
    shot: false,
    ups: 0,
    downs: 0,
    reason: ''
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

  methods: {
    bullet_class: function (player, playerIndex, n) {
      if (n > player.kills && n <= (player.kills + this.ups)) {
        return 'up'
      } else if (n === player.kills && this.downs === -1) {
        return 'down'
      } else if (n <= player.kills) {
        return 'kill'
      }
      return ''
    },

    score: function (playerIndex, score) {
      if (!this.match.isStarted) {
        return
      }

      console.log('setting score ' + playerIndex + ' - ' + score)
      if (score === 1 && this.ups < 3) {
        // We can only allow up to three kills per round...
        this.ups += 1
        if (this.ups === 3) {
          this.shot = true
          this.reason = 'sweep'
        }
      } else if (score === -1 && this.downs === 0) {
        // ...and only one suicide.
        this.downs -= 1
        this.shot = true
        this.reason = 'suicide'
      }
    },

    reset: function () {
      this.ups = 0
      this.downs = 0
      this.shot = false
      this.reason = ''
    },

    manual_shot: function () {
      if (!this.shot) {
        this.shot = true
        this.reason = 'manual'
      } else {
        this.shot = false
        this.reason = ''
      }
    }
  }
}
</script>

<style lang="scss" scoped>
@import "../variables.scss";

.player {
  width: 100%;
  position: relative;

  >div {
    float: left;
    height: 100%;
  }

  .button, .shots {
    width: 10%;
    position: relative;
    text-shadow: 2px 2px 2px rgba(0,0,0,0.8);

    >div {
      box-shadow: 2px 2px 3px rgba(0,0,0,0.3);
      width: 80%;
      height: 50%;
      background-color: #333339;
      cursor: pointer;
      text-align: center;
      vertical-align: middle;
      position: absolute;
      top: 50%;
      left: 50%;
      transform: translateX(-50%) translateY(-50%);
      display: flex;
      align-items: center;
      justify-content: center;
    }
  }
  .button {
    div {
      font-size: 10vh;
      user-select: none;
      -ms-user-select: none;
      -moz-user-select: none;
      -webkit-user-select: none;
      p {
        margin-top: -13%;
      }
    }
   }

  .shots {

    div {

      width: 80%;
      display: block;
      margin: 0 auto 1%;
      user-select: none;
      -ms-user-select: none;
      -moz-user-select: none;
      -webkit-user-select: none;

      &.give {
        background-color: #508850;
        p {color: #fff;}
      }

      &.mark {
        padding-top: 2%;
        font-size: 6vh;
      }
      &.reason {
        margin-top: -4%;
        font-size: 2vh;
      }
    }
  }

  .slider {
    width: 15%;
    text-shadow: 2px 2px 2px rgba(0,0,0,0.8);
    font-weight: bold;
    font-size: 1.6vw;

    &.green  { color: $green ; }
    &.blue   { color: $blue  ; }
    &.pink   { color: $pink  ; }
    &.orange { color: $orange; }
    &.white  { color: $white ; }
    &.yellow { color: $yellow; }
    &.cyan   { color: $cyan  ; }
    &.purple { color: $purple; }
    &.red    { color: $red; }

    div {
      width: 80%;
      height: 50%;
      font-size: 1.6em;
      cursor: pointer;

      text-align: center;
      vertical-align: middle;
      position: relative;
      top: 50%;
      left: 50%;
      transform: translateX(-50%) translateY(-50%);
      display: flex;
      align-items: center;
      justify-content: center;
      overflow: hidden;
    }
  }
  .scores {
    // http://stackoverflow.com/questions/6865194/fluid-width-with-equally-spaced-divs
    width: 55%;
    text-align: justify;
    text-justify: distribute;
    // background-color: #333339;

    &.final {
      div {
        width: 4%;
        margin: 0 0.5%;
      }
    }

    div {
      // position: relative;
      // top: 25%;
      width: 8%;
      margin: 0 1%;
      height: 8vh;
      vertical-align: top;
      display: inline-block;
      zoom: 1;

      background-color: #333339;
      border-radius: 10000px;
      box-shadow: 2px 2px 3px rgba(0,0,0,0.3);

      text-align: center;
      vertical-align: middle;
      position: relative;
      top: 50%;
      transform: translateX(0) translateY(-50%);

      p {
        position: relative;
        top: 50%;
        transform: translateX(0) translateY(-50%);
        font-size: 1.5em;
        color: #666672;
      }

      transition: 0.2s;

      &.kill {
        background-color: #dbdbdb;
        p {color: #151515;}
      }
      &.up {
        background-color: #508850;
        p {color: #fff;}
      }
      &.down {
        background-color: #885050;
        p {color: #fff;}
      }
    }
  }
}
</style>
