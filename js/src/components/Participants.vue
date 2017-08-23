<template>
  <div v-if="tournament">
    <header>
      <div class="content">
        <div class="title">Drunken TowerFall / {{tournament.name}} / Participants</div>
      </div>
      <div class="clear"></div>
    </header>

    <div class="clear"></div>

    <h2>Not joined ({{notJoined.length}})</h2>
    <div class="players not-joined">
      <div v-for="person in notJoined" class="player">
        <img @click="toggle" :id="person.id" :alt="person.nick" :src="'https://graph.facebook.com/'+person.facebook_id+'/picture?width=9999'"/>
        <p>{{person.name.split(" ")[0]}}</p>
      </div>
    </div>

    <h2>Joined ({{tournament.players.length}}/32)</h2>
    <div class="players joined">
      <div v-for="player in joined" class="player">
        <img @click="toggle" :id="player.person.id" :alt="player.person.nick" :src="player.avatar"/>
        <p>{{player.person.name.split(" ")[0]}}</p>
      </div>
    </div>
  </div>
</template>

<script>
import Tournament from '../models/Tournament'
import _ from 'lodash'

export default {
  name: 'Join',

  data () {
    return {
      people: [],
    }
  },

  methods: {
    toggle (e) {
      let person = e.srcElement
      this.api.toggle({ id: this.tournament.id, person: person.id }).then((res) => {
        console.log("join response:", res)
        var j = res.json()
        this.$route.router.push('/towerfall' + j.redirect)
      }, (err) => {
        console.error(`joining tournament ${this.tournament} failed`, err)
      })
    },
    setData: function (tournament) {
      console.log("setData tournament", tournament)
      this.$set('tournament', Tournament.fromObject(tournament))
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
    joined () {
      return _.sortBy(this.tournament.players, ['person.name'])
    },
    notJoined () {
      let $vue = this
      return _.filter(this.people, function (o) {
        let p = _.find($vue.tournament.players, function (p) {
          return p.person.id === o.id
        })
        return p === undefined
      })
    },
  },

  created: function () {
    console.debug("Creating API resource")
    let customActions = {
      getTournamentData: { method: "GET", url: "/api/towerfall/tournament{/id}/" },
      toggle: { method: "GET", url: "/api/towerfall/{id}/toggle/{person}" },
      people: { method: "GET", url: "/api/towerfall/people/" },
    }
    this.api = this.$resource("/api/towerfall", {}, customActions)

    // TODO: It could be that this makes more sense to have as
    // something in the Vuex storage.
    this.api.people().then(function (res) {
      console.log(res)
      this.$set(this.$data, 'people', _.sortBy(JSON.parse(res.data).people, ['name']))
    }, function (res) {
      console.log('error when getting people')
      console.log(res)
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
