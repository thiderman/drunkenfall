<template>
  <div v-if="tournament">
    <header>
      <div class="content">
        <div class="title">
          {{tournament.name}} / {{match.title}}
        </div>
      </div>
      <div class="links" v-if="user.isJudge">
      </div>
      <div class="clear"></div>
    </header>

    <div class="chart"></div>

    <div class="clear"></div>
  </div>
</template>

<script>
import * as d3 from "d3"

export default {
  name: 'PostMatch',

  computed: {
    user () {
      return this.$store.state.user
    },
    tournament () {
      return this.$store.getters.getTournament(
        this.$route.params.tournament
      )
    },
    match () {
      let match
      let index = this.tournament.current.index
      let kind = this.tournament.current.kind

      if (kind === 'final') {
        return this.tournament.final
      }

      kind = kind + 's'

      match = this.tournament[kind][index]

      // We don't want to update until the next match has been
      // started. If we do, the graphs are removed as soon as the
      // judges end the previous match.
      // Also, if we're on the first match there is no previous one,
      // so don't try to grab the previous one in that case.
      if (!match.isStarted || (kind === 'tryouts' && index === 0)) {
        index = this.tournament.previous.index
        kind = this.tournament.previous.kind + 's'
        console.log([index, kind])
        return this.tournament[kind][index]
      }

      return match
    },
  },

  watch: {
    tournament (nt, ot) {
      console.log(nt)
      console.log(ot)
      if (nt) {
        this.renderChart()
      }
    }
  },

  methods: {
    renderChart () {
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
      var maxScore = match.end

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
