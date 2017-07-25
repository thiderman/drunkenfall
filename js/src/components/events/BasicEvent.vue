<template>
  <div>
     <div class="event">
       <img alt="" src="{{event.player.avatar}}" v-if="event.hasPlayer"/>
       <div>
         {{event.print}}
       </div>
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

<style lang="scss" scoped>

.event {
  font-size: 1.8em;
  float: left;
  margin: 1% 0.7%;
  width: 40%;
  background-color: #333339;
  box-shadow: 2px 2px 3px rgba(0,0,0,0.3);

  div {
      padding: 1em 1.3em;
  }
}

img {
  display: block;
  object-fit: cover;
  border-radius: 100%;
  width:  3em;
  height: 3em;
  margin: 4%;
  box-shadow: -1px -1px 6px rgba(0,0,0,0.5);
  background-color: rgba(10,12,14,0.3);
  float: left;
}

</style>
