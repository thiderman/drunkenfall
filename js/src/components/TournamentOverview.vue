<template>
  <div>
    <header>
      <div class="content">
        <div class="title">{{tournament.name}}</div>
      </div>
      <div class="links">
        <div class="action" @click="start"
          v-if="user.isJudge && tournament.canStart">Start</div>
        <div class="action" @click="next"
          v-if="user.isJudge && tournament.isRunning">Next match</div>
        <div class="action" @click="reshuffle"
          v-if="user.isProducer && tournament.canShuffle">Reshuffle</div>
        <div class="action" @click="log"
          v-if="user.isProducer">Log</div>

        <router-link :to="{ name: 'credits', params: { tournament: tournament.id }}"
          v-if="user.isProducer && tournament.isEnded">
          Roll credits
        </router-link>

        <router-link :to="{ name: 'participants', params: { tournament: tournament.id }}"
          v-if="user.isProducer && canParticipants">
          Participants
        </router-link>

        <router-link :to="{ name: 'runnerups', params: { tournament: tournament.id }}"
          v-if="user.isCommentator && shouldBackfill">
          Backfill semis
        </router-link>
      </div>
      <div class="clear"></div>
    </header>

    <div class="subheader" v-if="user.isCommentator && nextMatch && !tournament.isEnded">
      <div v-if="!nextMatch.isScheduled">
        <p>
          Pause until
          <span>{{tournament.current.kind}}</span>
          <span>{{tournament.current.index+1}}</span>:
        </p>
        <div class="links">
          <a @click="setTime(10)">10 min</a>
          <a @click="setTime(7)">7 min</a>
          <a @click="setTime(5)">5 min</a>
          <a @click="setTime(3)">3 min</a>
        </div>
        <div class="clear"></div>
      </div>
      <div v-if="nextMatch.isScheduled">
        <p class="center">
          <span>{{tournament.current.kind}}</span>
          <span>{{tournament.current.index+1}}</span> scheduled at
          {{nextMatch.scheduled.format("HH:mm")}}
        </p>
        <div class="clear"></div>
      </div>
    </div>

    <div class="category tryouts">
      <h3>Tryouts</h3>
      <div class="matches">
        <template v-for="m in tournament.tryouts">
          <match-overview :match="m" :class="'match ' + m.kind"></match-overview>
        </template>
      </div>
      <div class="clear"></div>
    </div>

    <div class="category semis">
      <h3>Semi-finals</h3>
      <div class="matches">
        <template v-for="m in tournament.semis">
          <match-overview :match="m" :class="'match ' + m.kind"></match-overview>
        </template>
      </div>
      <div class="clear"></div>
    </div>
    <div class="category final">
      <h3>Final</h3>
      <div class="matches">
        <match-overview :match="tournament.final" class="match final"></match-overview>
      </div>
    </div>
  </div>
</template>

<script>
import MatchOverview from './MatchOverview'
import Match from '../models/Match'
import _ from 'lodash'

export default {
  name: 'TournamentOverview',

  components: {
    MatchOverview
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
    match () {
      let kind = this.$route.params.kind
      let idx = this.$route.params.match

      if (kind === 'final') {
        return this.tournament.final
      }
      kind = kind + 's'
      return this.tournament[kind][idx]
    },
    shouldBackfill () {
      let c = this.tournament.current
      let ps = _.sumBy(this.tournament.semis, (m) => { return m.players.length })

      if (c.kind === 'semi' && c.index === 0 && ps < 8) {
        return true
      }
      return false
    },
    canParticipants () {
      return this.tournament.current.kind === 'tryout'
    },
    nextMatch () {
      let kind = this.tournament.current.kind
      let idx = this.tournament.current.index

      if (kind === 'tryout') {
        kind = 'tryouts'
      } else if (kind === 'semi') {
        kind = 'semis'
      }

      let m
      if (kind === 'final') {
        m = Match.fromObject(this.tournament[kind])
      } else {
        m = Match.fromObject(this.tournament[kind][idx])
      }
      return m
    },

    runnerups () {
      let t = this.tournament

      if (!t.runnerups) {
        return []
      }

      return _.map(t.runnerups, (runnerup) => {
        return _.find(t.players, function (p) {
          return p.person.id === runnerup.id
        })
      })
    },
  },

  methods: {
    start () {
      if (this.tournament) {
        this.api.start({ id: this.tournament.id }).then((res) => {
          console.log("start response:", res)
          let j = res.json()
          this.$route.router.push('/towerfall' + j.redirect)
        }, (err) => {
          console.error(`start for ${this.tournament} failed`, err)
        })
      } else {
        console.error("start called with no tournament")
      }
    },
    next () {
      if (this.tournament) {
        this.api.next({ id: this.tournament.id }).then((res) => {
          console.debug("next response:", res)
          let j = res.json()
          this.$router.push('/towerfall' + j.redirect)
        }, (err) => {
          console.error(`next for ${this.tournament} failed`, err)
        })
      } else {
        console.error("next called with no tournament")
      }
    },
    reshuffle () {
      if (this.tournament) {
        this.api.reshuffle({ id: this.tournament.id }).then((res) => {
          console.debug("reshuffle response:", res)
          let j = res.json()
          this.$route.router.push('/towerfall' + j.redirect)
        }, (err) => {
          console.error(`reshuffle for ${this.tournament} failed`, err)
        })
      } else {
        console.error("reshuffle called with no tournament")
      }
    },
    log () {
      this.$router.push({
        name: "log",
        params: {
          tournament: this.tournament.id
        }
      })
    },
    setTime (x) {
      this.api.setTime({ id: this.tournament.id, time: x }).then((res) => {
        console.debug("settime response:", res)
        let j = res.json()
        console.log('Redirect to /towerfall' + j.redirect)
        // this.$route.router.push('/towerfall' + j.redirect)
      }, (err) => {
        console.error(`settime for ${this.tournament} failed`, err)
      })
    },

    selectRunnerup (p) {
      if (this.isSelected(p)) {
        // TODO(thiderman): Doesn't work. Fuck this.
        console.log("selected, to remove", this.selected)
        this.selected = _.remove(this.selected, function (o) {
          console.log(o.person.id, p.person.id)
          return o.person.id === p.person.id
        })
        return
      }

      this.selected.push(p)
    },
    isSelected (p) {
      return _.find(this.selected, p) !== undefined
    },
  },

  created () {
    this.api = this.$resource("/api/towerfall", {}, {
      start: { method: "GET", url: "/api/towerfall{/id}/start/" },
      next: { method: "GET", url: "/api/towerfall{/id}/next/" },
      reshuffle: { method: "GET", url: "/api/towerfall{/id}/reshuffle/" },
      setTime: { method: "GET", url: "/api/towerfall{/id}/time{/time}" },
    })
  },
}
</script>

<style lang="scss" scoped>
@import "../variables.scss";

.tournament {
  position: relative;
}

.tryouts, .semis, .final {
  width: 29%;
  float: left;
  margin-left: 3%;
  position: relative;
}

.category h3 {
  text-align: center;
  font-size: 200%;
  margin: 4%;
}

.match {
  width: 100%;
  font-size: 150%;
  display: block;
  position: relative;

}

.runnerups, .selected-runnerups {
  width: 100%;
  margin: 10px;
  box-shadow: 2px 2px 3px rgba(0,0,0,0.3);

  .runnerup {
    padding: 0.1em 0.3em;
    font-size: 24px;
    color: #aaa;

    p {
      margin: 1px;
      &.name {
        float: left;
        // font-weight: bold;
      }
      &.score {
        float: right;
      }
      b {
        text-shadow: 1px 1px 1px rgba(0,0,0,0.4);
      }
    }
  }

  .runnerup:nth-child(odd) {
    background-color: #333;
  }
  .runnerup:nth-child(even) {
    background-color: #272727;
  }
}

.selected-runnerups {
  .runnerup:nth-child(odd) {
    background-color: #405040;
  }
  .runnerup:nth-child(even) {
    background-color: #394939;
  }

  .button {
    width: 50px;
    margin: 10px auto;
    padding: 0.3em 0.5em;
    background-color: #405060;
    cursor: pointer;
    text-shadow: 1px 1px 1px rgba(0,0,0,0.4);
    text-align: center;
  }
}

.subheader {
  width: 80%;
  margin: 30px auto;
  background-color: #333339;
  box-shadow: 2px 2px 3px rgba(0,0,0,0.3);
  text-shadow: 2px 2px 3px rgba(0,0,0,0.5);

  p {
    font-size: 2em;
    padding: 0.3em 0.5em;
    float: left;

    span {
      text-transform: capitalize;
    }
  }

  .links {
    float: right;
    a, .action {
      margin: 15px  !important;
      font-size: 24px;
      float: right;
      background-color: #405060;
      color: #dbdbdb;
      display: block;
      font-weight: bold;
      padding: 7px 30px;
      text-align: center;
      text-decoration: none;
      margin: 10px auto;
      min-width: 60px;
    }
  }
}

</style>
