<template>
  <div>
    <header>
      <div class="content">
        <div class="title">{{tournament.name}}</div>
      </div>
      <div class="links">
        <a v-if="tournament.canStart"
          v-link="{ name: 'join', params: { tournament: tournament.id }}">Join</a>
        <div class="action" @click="start"
          v-if="user.level(levels.judge) && tournament.canStart">Start</div>
        <div class="action" @click="next"
          v-if="user.level(levels.judge) && tournament.isRunning">Next match</div>
      </div>
      <div class="clear"></div>
    </header>

    <div class="category tryouts">
      <h3>Tryouts</h3>
      <div class="matches">
        <template v-for="m in tournament.tryouts">
          <match-overview :match="m" class="match {{m.kind}}">
        </template>
      </div>
      <div class="clear"></div>
    </div>

    <div class="category semis">
      <h3>Semi-finals</h3>
      <div class="matches">
        <template v-for="m in tournament.semis">
          <match-overview :match="m" class="match {{m.kind}}">
        </template>
      </div>
      <div class="clear"></div>

      <div v-if="runnerups.length > 0">
        <h3>Runnerups</h3>
        <div class="runnerups">
          <template v-for="player in runnerups">
            <div class="runnerup">
              <p class="name">{{player.displayName}}</p>
              <p class="score">
                <b>{{player.score}}</b> points
                /
                <b>{{player.matches}}</b> matches
              </p>
              <div class="clear"></div>
            </div>
          </template>
        </div>
      </div>

    </div>
    <div class="category final">
      <h3>Final</h3>
      <div class="matches">
        <match-overview :match="tournament.final" class="match final">
      </div>
    </div>
  </div>
</template>

<script>
import MatchOverview from './MatchOverview'
import Tournament from '../models/Tournament'
import * as levels from "../models/Level"
import User from '../models/User'
import _ from 'lodash'

export default {
  name: 'TournamentOverview',

  components: {
    MatchOverview
  },

  props: {
    tournament: new Tournament(),
    user: new User(),
    levels: levels,
  },

  computed: {
    runnerups: function () {
      let t = this.tournament

      if (!t.runnerups) {
        return []
      }

      return _.map(t.runnerups, (runnerupName) => {
        console.log("runnerup map", runnerupName)
        return _.find(t.players, { displayName: runnerupName })
      })
    }
  },

  methods: {
    start: function () {
      if (this.tournament) {
        this.api.start({ id: this.tournament.id }).then((res) => {
          console.log("start response:", res)
          let j = res.json()
          this.$route.router.go('/towerfall' + j.redirect)
        }, (err) => {
          console.error(`start for ${this.tournament} failed`, err)
        })
      } else {
        console.error("start called with no tournament")
      }
    },
    next: function () {
      if (this.tournament) {
        this.api.next({ id: this.tournament.id }).then((res) => {
          console.debug("next response:", res)
          let j = res.json()
          this.$route.router.go('/towerfall' + j.redirect)
        }, (err) => {
          console.error(`next for ${this.tournament} failed`, err)
        })
      } else {
        console.error("next called with no tournament")
      }
    }
  },

  created: function () {
    console.debug("Creating API resource")
    let customActions = {
      start: { method: "GET", url: "/api/towerfall{/id}/start/" },
      next: { method: "GET", url: "/api/towerfall{/id}/next/" },
      getData: { method: "GET", url: "/api/towerfall/tournament{/id}/" }
    }
    this.api = this.$resource("/api/towerfall", {}, customActions)
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



  &.gold, &.silver, &.bronze {
    color: #fff;
    text-shadow: 1px 1px 5px rgba(0,0,0,0.3);
  }

}

.runnerups {
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

</style>
