<template>
  <div>
     <div class="event">
       <img alt="" v-bind:src="event.player.avatar" v-if="event.hasPlayer"/>
       <div class="content" v-html="event.print">
       </div>
       <p class="time">
         {{event.time.fromNow()}}
       </p>
     </div>
  </div>
</template>


<script>
import Event from '../../models/Event.js'

export default {
  name: 'BasicEvent',

  props: {
    event: new Event(),
  }
}
</script>

<style lang="scss">
// This cannot be scoped because the span elements that colorize the
// templates are not created by vue and therefore are not given the scope.

@import "../../variables.scss";

.event {
  font-size: 1.4em;
  float: left;
  margin: 1% 0.7%;
  width: 30%;
  background-color: #333339;
  box-shadow: 2px 2px 3px rgba(0,0,0,0.3);
  position: relative;

  .content {
    padding: 1em 1.3em;

    span {
      text-shadow: 2px 2px 3px rgba(0,0,0,0.5);
      &.green  { color: $green  ; }
      &.blue   { color: $blue   ; }
      &.pink   { color: $pink   ; }
      &.orange { color: $orange ; }
      &.white  { color: $white  ; }
      &.yellow { color: $yellow ; }
      &.cyan   { color: $cyan   ; }
      &.purple { color: $purple ; }
      &.red    { color: $red    ; }
    }
  }

  .time {
    position: absolute;
    bottom: 0.4em;
    right:  0.4em;
    text-align: right;
    font-size: 0.7em;
    color: #555;
    padding: 0.5em;
  }
}

img {
  display: block;
  object-fit: cover;
  border-radius: 100%;
  width:  3em;
  height: 3em;
  margin: 2.5% 4.5% 2.5% 3.5%;
  box-shadow: -1px -1px 6px rgba(0,0,0,0.5);
  background-color: rgba(10,12,14,0.3);
  float: left;
}

</style>
