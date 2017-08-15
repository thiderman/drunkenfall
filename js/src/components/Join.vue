<template>
  <div>
    <header>
      <div class="content">
        <div class="title">Drunken TowerFall / {{tournament.name}} / Join</div>
      </div>
      <div class="clear"></div>
    </header>

    <div class="clear"></div>

    <h1>Joining {{tournament.name}}</h1>

    <p>
      Drunken TowerFall has <i>drunken</i> in its name for a reason. A key
      part of the gameplay is getting shots for various reasons dictated by
      the judges.
    </p>

    <p>
      Do you agree that during gameplay you will receive hard liquor shots?
    </p>

    <form v-on:submit="join">
      <input type="checkbox" id="checkbox" v-model="approve">
      <label for="checkbox">Yes, I agree</label>

      <div class="clear"></div>

      <input id="doit" :class="{ready: approve}" type="submit" value="Join tournament"/>
    </form>

    <div class="button">
    </div>
  </div>
</template>

<script>
export default {
  name: 'Join',

  data () {
    return {
      approve: false,
    }
  },

  computed: {
    tournament () {
      return this.$store.getters.getTournament(
        this.$route.params.tournament
      )
    },
  },

  methods: {
    join (e) {
      e.preventDefault()
      if (!this.approve) {
        console.log("Not joining - not approved")
        return
      }

      this.api.join({ id: this.tournament.id }).then((res) => {
        console.log("join response:", res)
        var j = res.json()
        this.$route.router.push('/towerfall' + j.redirect)
      }, (err) => {
        console.error(`joining tournament ${this.tournament} failed`, err)
      })
    },
  },

  created: function () {
    console.debug("Creating API resource")
    let customActions = {
      join: { method: "GET", url: "/api/towerfall/{id}/join/" },
    }
    this.api = this.$resource("/api/towerfall", {}, customActions)
  },
}
</script>

<style lang="scss" scoped>

* {
  text-align: center;
}

h1 {
  font-size: 4em;
}

p {
  font-size: 1.5em;
  width: 25em;
  margin: 1em auto;
}

label {
  font-size: 2.5em;
}

#doit {
  margin: 1em auto;
  width: 350px;
  border: none;
  font-size: 2em;
  transition: 1.0s;
  color: #dbdbdb;
  background-color: #333333;
  padding: 0.4em 0.8em;

  &.ready {
    background-color: #405060;
  }
}

</style>
