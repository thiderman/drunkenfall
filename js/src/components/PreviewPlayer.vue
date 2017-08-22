<template>
  <div :class="'player ' + player.color">
    <img :alt="player.displayName" :src="player.avatar"/>

    <div class="protector">
      <div v-if="showName" class="super-ribbon">
      {{player.firstName}}
      </div>

      <div :class="'ribbon ' + player.color">
        <strong class="ribbon-content">
          {{player.displayName}}
        </strong>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  name: 'PreviewPlayer',

  props: {
    player: {},
    match: {},
    index: 0,
  },

  computed: {
    showName: function () {
      // Don't show the first name tag super ribbon if the player name is the
      // same as that.
      if (this.player.firstName.toLowerCase() === this.player.displayName.toLowerCase()) {
        return false
      }
      return true
    }
  },
}
</script>

<style lang="scss" scoped>
@import "../variables.scss";
@import "../ribbon.scss";

.player {
  float: left;
  height: 50%;
  width: 50%;
  margin-top: 4em;

  img {
    display: block;
    object-fit: cover;
    border-radius: 100%;
    width:  300px;
    height: 300px;
    margin: 0px auto;
    box-shadow: -1px -1px 6px rgba(0,0,0,0.5);
    background-color: rgba(10,12,14,0.3);
  }

  .super-ribbon {
    margin: -3.5em auto 2.2em;
    width: 20%;
  }
  .ribbon {
    font-size: 2em;
    width: 60%;
  }
}

</style>
