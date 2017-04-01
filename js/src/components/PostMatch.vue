<template>
  <div>
    <header>
      <div class="content">
        <div class="title">
          {{tournament.name}} / {{match.kind | capitalize}} {{match.index +1}}
        </div>
      </div>
      <div class="links" v-if="user.level(levels.judge)">
      </div>
      <div class="clear"></div>
    </header>

    <div class="chart"></div>

    <div class="clear"></div>
  </div>
</template>

<script>
import * as d3 from "d3"
import ControlPlayer from './ControlPlayer.vue'
import PreviewPlayer from './PreviewPlayer.vue'
import LivePlayer from './LivePlayer.vue'
import Match from '../models/Match.js'
import Tournament from '../models/Tournament.js'
import * as levels from "../models/Level.js"
import _ from 'lodash'

export default {
  name: 'PostMatch',
  components: {
    ControlPlayer,
    PreviewPlayer,
    LivePlayer,
  },

  data () {
    return {
      match: new Match(),
      tournament: new Tournament(),
      user: this.$root.user,
      levels: levels,
    }
  },

  computed: {
    can_commit: function () {
      return true
    }
  },

  methods: {
    commit: function () {
      // TODO this could potentially be a class
      let payload = {
        'state': _.map(this.$refs.players, (controlPlayer) => {
          return _.pick(controlPlayer, ['ups', 'downs', 'shot', 'reason'])
        })
      }

      console.log(payload)
      this.api.commit({ id: this.tournament.id, kind: this.match.kind, index: this.match.index }, payload).then(function (res) {
        console.log(res)
        this.$set('match', Match.fromObject(res.data.match))

        _.each(this.$refs.players, (controlPlayer) => { controlPlayer.reset() })
      }, function (res) {
        console.log('error when setting score')
        console.log(res)
      })
    },

    refresh: function () {
      // Hax to make vue refresh the entire page.
      // Since nothing on this page is properly bound to components right now
      // the updates won't trigger properly.
      this.$set('updated', Date.now())
    },
    setData: function (tournament) {
      let match
      let index = tournament.current.index
      let kind = tournament.current.kind

      if (kind === 'tryout') {
        kind = 'tryouts'
      } else if (kind === 'semi') {
        kind = 'semis'
      }

      if (kind === 'final') {
        match = Match.fromObject(tournament[kind])
      } else {
        match = Match.fromObject(tournament[kind][index])
      }

      if (!match.isStarted) {
        // If we're on the first match, there is no previous, so bail.
        if (index === 0 && kind === 'tryouts') {
          this.$set('tournament', Tournament.fromObject(tournament))
          return
        }

        index = tournament.previous.index
        kind = tournament.previous.kind

        if (kind === 'tryout') {
          kind = 'tryouts'
        } else if (kind === 'semi') {
          kind = 'semis'
        }

        if (kind === 'final') {
          match = Match.fromObject(tournament[kind])
        } else {
          match = Match.fromObject(tournament[kind][index])
        }
      }

      this.$set('match', match)
      this.$set('tournament', Tournament.fromObject(tournament))
      this.renderChart()
    },
    renderChart: function () {
      // set the dimensions and margins of the graph
      var margin = {top: 70, right: 50, bottom: 50, left: 50}
      var width = window.innerWidth - margin.left - margin.right
      var height = window.innerHeight - (margin.top * 2) - margin.bottom

      // set the ranges
      var x = d3.scaleLinear().range([0, width])
      var y = d3.scaleLinear().range([height, 0])

      // define the line
      var valueline = d3.line()
          .x(function (d, i) {
            return x(i)
          })
          .y(function (d) {
            return y(d)
          })

      // append the svg obgect to the body of the page
      // appends a 'group' element to 'svg'
      // moves the 'group' element to the top left margin
      d3.select("svg").remove()

      var svg = d3.select("body").append("svg")
          .attr("width", width + margin.left + margin.right)
          .attr("height", height + margin.top + margin.bottom)
          .append("g")
          .attr("transform",
                "translate(" + margin.left + "," + margin.top + ")")

      var match = this.match
      var data = match.chartData

      var roundCount = data[0] ? data[0].length - 1 : 0
      var maxScore = match.length

      // Scale the range of the data
      x.domain([0, roundCount])
      y.domain([0, maxScore])

      svg.selectAll(".line")
        .data(data)
        .enter()
        .append("path")
        .attr("d", valueline)
        .attr("class", function (d, i) { return "line " + match.players[i].color })

      // Add the X Axis
      svg.append("g")
        .attr("class", "axis")
        .attr("transform", "translate(0," + height + ")")
        .call(d3.axisBottom(x))

      // Add the Y Axis
      svg.append("g")
        .attr("class", "axis")
        .call(d3.axisRight(y))
        .attr("transform", "translate(" + width + ", 0)")

      svg.append("g")
        .attr("class", "axis")
        .call(d3.axisLeft(y))

      svg.append("g")
        .attr("class", "legend")
        .attr("transform", "translate(10, 0)")

      var player = d3.select(".legend").selectAll(".player")
          .data(match.players)
          .enter()
          .append("g")
          .attr("class", "player")
          .attr("transform", function (d, i) {
            return "translate(0," + (i * 60) + ")"
          })

      player.append("rect")
        .attr("class", function (d) { return d.color })
        .attr("width", 48)
        .attr("height", 48)
        .attr("rx", 16)
        .attr("ry", 16)

      player.append("text")
        .text(function (d) { return d.person.nick })
        .attr("transform", "translate(55, 35)")
    },
  },

  created: function () {
    console.debug("Creating API resource")
    let customActions = {
      commit: { method: "POST", url: "/api/towerfall/tournament{/id}{/kind}{/index}/commit/" },
      start: { method: "GET", url: "/api/towerfall/tournament{/id}{/kind}{/index}/start/" },
      end: { method: "GET", url: "/api/towerfall/tournament{/id}{/kind}{/index}/end/" },
      reset: { method: "GET", url: "/api/towerfall/tournament{/id}{/kind}{/index}/reset/" },
      getTournamentData: { method: "GET", url: "/api/towerfall/tournament{/id}/" }
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
          console.log(res)
          this.setData(res.data.tournament)
        }, function (res) {
          console.log('error when getting tournament')
          console.log(res)
        })
      } else {
        // Something is set - we're clicking on a link and can reuse the
        // already existing data immediately
        this.setData(
          to.router.app.get(to.params.tournament),
          to.params.kind,
          parseInt(to.params.match)
        )
      }
    }
  }
}
</script>

<style lang="scss" >

@import "../variables.scss";

.control {
  height: 85vh;
  padding: 0.8%;
}

.player {
  height: 25%;
  display: block;
}

.chart div {
  font: 10px sans-serif;
  background-color: steelblue;
  text-align: right;
  padding: 3px;
  margin: 1px;
  color: white;
}

.axis {
  stroke: #dbdbdb;
  stroke-width: 3px;

  text {
    stroke-width: 1px;
  }

  line, path {
    stroke: #dbdbdb;
    stroke-width: 3px;
  }
}

.line {
  fill: none;
  stroke-width: 3px;

  &.green  { stroke: $green ; }
  &.blue   { stroke: $blue  ; }
  &.pink   { stroke: $pink  ; }
  &.orange { stroke: $orange; }
  &.white  { stroke: $white ; }
  &.yellow { stroke: $yellow; }
  &.cyan   { stroke: $cyan  ; }
  &.purple { stroke: $purple; }
  &.red    { stroke: $red; }
}

.player {
  text {
    fill: #dbdbdb;
    font-size: 32px;
  }

  rect {
    &.green  { fill: $green ; }
    &.blue   { fill: $blue  ; }
    &.pink   { fill: $pink  ; }
    &.orange { fill: $orange; }
    &.white  { fill: $white ; }
    &.yellow { fill: $yellow; }
    &.cyan   { fill: $cyan  ; }
    &.purple { fill: $purple; }
    &.red    { fill: $red; }
  }
}

</style>
