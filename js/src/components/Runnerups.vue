<template>
  <div v-if="tournament">
    <header>
      <div class="content">
        <div class="title">Drunken TowerFall / {{tournament.name}} / Runnerup backfill</div>
      </div>
      <div class="links">
        <a @click="commit" v-if="canCommit">Commit</a>
      </div>
      <div class="clear"></div>
    </header>

    <div class="clear"></div>

    <h2>Already the semi</h2>
    <div class="players joined">
      <div v-for="player in inSemi" class="player">
        <img :alt="player.person.nick" :src="player.avatar"/>
        <p>{{player.person.name.split(" ")[0]}}</p>
      </div>
    </div>

    <h2>To be added to the semi ({{toSemi.length}} / {{needed}})</h2>
    <div class="players joined">
      <div v-for="person in toSemi" class="player">
        <img @click="remove" :id="person.id" :alt="person.nick" :src="person.avatar"/>
        <p>{{person.name.split(" ")[0]}}</p>
      </div>
    </div>

    <h2>Runnerups</h2>
    <div class="players not-joined">
      <div v-for="person in runnerups" class="player">
        <img @click="add" :id="person.id" :alt="person.nick" :src="person.avatar"/>
        <p>{{person.name.split(" ")[0]}}</p>
      </div>
    </div>
  </div>
</template>

<script>
import _ from 'lodash'

export default {
  name: 'Runnerups',

  data () {
    return {
      selected: [],
    }
  },

  methods: {
    add (e) {
      if (this.selected.length < this.needed) {
        this.selected.push(e.srcElement.id)
      }
    },
    remove (e) {
      this.selected.pop(e.srcElement.id)
    },
    commit () {
      let b = this.selected.join(',')
      console.log(b)

      this.api.backfill({id: this.tournament.id}, b).then((res) => {
        console.debug("backfill response:", res)
        let j = res.json()
        console.log('Redirect to /towerfall' + j.redirect)
        this.$router.push('/towerfall' + j.redirect)
      }, (err) => {
        console.error(`backfill for ${this.tournament} failed`, err)
      })
    }
  },

  computed: {
    tournament () {
      return this.$store.getters.getTournament(
        this.$route.params.tournament
      )
    },
    user () {
      return this.$store.state.user
    },
    runnerups () {
      let $vue = this
      let r = _.filter(this.tournament.runnerups, function (p) {
        let pl = _.find($vue.selected, function (o) {
          return p.id === o
        })
        return pl === undefined
      })
      return _.sortBy(r, ['name'])
    },
    toSemi () {
      let $vue = this
      return _.filter($vue.tournament.runnerups, function (p) {
        let pl = _.find($vue.selected, function (o) {
          return p.id === o
        })
        return pl !== undefined
      })
    },
    inSemi () {
      return _.concat(
        this.tournament.semis[0].players,
        this.tournament.semis[1].players,
      )
    },
    needed () {
      return 8 - _.sumBy(this.tournament.semis, function (m) {
        return m.players.length
      })
    },
    canCommit () {
      return this.selected.length === this.needed
    },
  },

  created: function () {
    this.api = this.$resource("/api/towerfall", {}, {
      toggle: { method: "GET", url: "/api/towerfall/{id}/toggle/{person}" },
      people: { method: "GET", url: "/api/towerfall/people/" },
      backfill: { method: "POST", url: "/api/towerfall{/id}/backfill/" },
    })
  }
}
</script>

<style lang="scss" scoped>

* {
  text-align: center;
}

.players {
  text-align: center;
  width: 80%;
  margin: 10px auto;
  min-height: 120px;

  .player {
    display: inline-block;
    width: 100px;
    margin-top: 0px;
    cursor: pointer;

    img {
      object-fit: cover;
      border-radius: 100%;
      width:  100px;
      height: 100px;
      box-shadow: -1px -1px 6px rgba(0,0,0,0.5);
      background-color: rgba(10,12,14,0.3);
      margin-bottom: -30px;
    }
    p {
      width: 80%;
      text-align: center;
      padding: 0.2em 0.3em;
      margin: 0.5em auto;
      display: inline-block;
      box-shadow: 1px 1px 6px rgba(0,0,0,0.5);
      font-weight: bold;
    }
  }
}
.joined p {
  background-color: #406040;
}
.not-joined p {
  background-color: #604040;
}

</style>
