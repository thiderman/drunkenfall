<template>
  <div>
    <header>
      <div class="content">
        <div class="title">Drunken TowerFall / {{tournament.name}} / Edit</div>
      </div>
      <div class="clear"></div>
    </header>

    <form v-on:submit="edit">
      <textarea v-model="data"
        autocomplete="off" autocorrect="off" autocapitalize="off" spellcheck="false"
        cols="30" id="" name="" rows="40"></textarea>

      <input id="doit" type="submit" value="Edit tournament"/>
    </form>
  </div>
</template>

<script>
import Tournament from '../models/Tournament'

export default {
  name: 'Edit',

  data () {
    return {
      tournament: new Tournament(),
      data: "",
    }
  },

  methods: {
    edit (e) {
      e.preventDefault()
      this.api.edit({ id: this.tournament.id }, this.data).then((res) => {
        console.log("edit response:", res)
        var j = res.json()
        this.$route.router.push('/towerfall' + j.redirect)
      }, (err) => {
        console.error(`editing tournament ${this.tournament} failed`, err)
      })
    },
    setData: function (tournament) {
      console.log("setData tournament", tournament)
      this.$set('tournament', Tournament.fromObject(tournament))
      this.$set('data', JSON.stringify(tournament, null, 2))
    }
  },

  created: function () {
    let customActions = {
      getTournamentData: { method: "GET", url: "/api/towerfall/tournament{/id}/" },
      edit: { method: "POST", url: "/api/towerfall/{id}/edit/" },
    }
    this.api = this.$resource("/api/towerfall", {}, customActions)
  },

  route: {
    data ({ to }) {
      // listen for tournaments from App
      this.$on(`tournament${to.params.tournament}`, (tournament) => {
        console.debug("New tournament from App:", tournament)
        this.setData(tournament)
      })

      if (to.router.app.tournaments.length === 0) {
        // Nothing is set - we're reloading the page and we need to get the
        // data manually
        this.api.getTournamentData({ id: to.params.tournament }).then(function (res) {
          this.setData(
            res.data.tournament,
          )
        }, function (res) {
          console.log('error when getting tournament')
          console.log(res)
        })
      } else {
        // Something is set - we're clicking on a link and can reuse the
        // already existing data immediately
        this.setData(
          to.router.app.get(to.params.tournament),
        )
      }
    }
  }
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
  background-color: #405060;
  padding: 0.4em 0.8em;
}

textarea {
  text-align: left;
  color: #dbdbdb;
  background-color: #333339;
  width: 90%;
  height: 60%;
  box-shadow: inset 2px 2px 2px 0px rgba(0,0,0,0.5);
  outline: none;
  border: none;
  padding: 1em;
}

</style>
