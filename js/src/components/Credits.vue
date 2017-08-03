<template>
  <div>
    <div id="executive">
      <img :alt="executive.nick" :src="executive.avatar"/>
      <h2>Executive Producer</h2>
      <h1>{{executive.name}}</h1>
      <h3 :class="executive.color">{{executive.nick}}</h3>
    </div>

    <div id="producers">
      <h1>Producers</h1>

      <div class="first">
        <div class="producer" v-for="p in first_producers">
          <img :alt="p.nick" :src="p.avatar"/>
          <div>
            <h1>{{p.name}}</h1>
            <h3 :class="p.color">{{p.nick}}</h3>
          </div>
        </div>
      </div>

      <div class="second">
        <div class="producer" v-for="p in second_producers">
          <img :alt="p.nick" :src="p.avatar"/>
          <div>
            <h1>{{p.name}}</h1>
            <h3 :class="p.color">{{p.nick}}</h3>
          </div>
        </div>
      </div>
    </div>

    <div id="rolling">
      <h1>DrunkenFall IV</h1>
      <h2>A Song of Dig and Jalo</h2>

      <div id="players">
        <h1>Combatants</h1>
        <div class="player" v-for="p in players">
          <img :alt="p.nick" :src="p.avatar"/>
          <div>
            <h1 :class="p.color">{{p.nick}}</h1>
          </div>
        </div>
      </div>
      <div class="clear"></div>
      <div id="thanks">
        <h1>Special thanks</h1>
        <div>
          <img alt="Matt Makes Games" src="/static/img/mmg.png"/>
          <p>
            For the many hundreds of hours of arrows, fun, anger, screaming,
            joy, pain, brambles, regret, lasers, bombs, and love.<br/>
            We couldn't have done this without you. Thank you. &lt;3
          </p>
        </div>
      </div>

      <div class="clear"></div>

      <div id="harmed">
        <span>{{archersHarmed}}</span> archers were harmed in the making of this broadcast
      </div>
    </div>


    <div id="return">
      <h4>DrunkenFall will return in</h4>
      <h1>DrunkenFall V</h1>
      <h2 class="orange">Dusk on East Mountain</h2>
      <h3>Sep 30th</h3>
    </div>

  </div>
</template>

<script>
import Person from '../models/Person.js'
import _ from "lodash"

let pause = 2000
let duration = 6000
let roll = 8 * 1000

export default {
  name: 'Credits',

  data () {
    return {
      executive: Person,
      first_producers: [],
      second_producers: [],
      players: [],
      archersHarmed: 0,
    }
  },

  created: function () {
    console.debug("Creating API resource")
    let customActions = {
      getCredits: { method: "GET", url: "/api/towerfall{/id}/credits/" }
    }
    this.api = this.$resource("/api/towerfall", {}, customActions)

    document.getElementsByTagName("body")[0].className = "scroll-less"
    this.producerCards()
  },
  methods: {
    producerCards: function () {
      let $vue = this
      setTimeout(function () {
        let exec = document.getElementById("executive")
        exec.className = "active"

        setTimeout(function () {
          exec.className = ""

          setTimeout(function () {
            let prod = document.getElementById("producers")
            prod.className = "active"
            let first = prod.getElementsByClassName("first")[0]
            first.className = "first active"

            setTimeout(function () {
              first.className = "first"
              setTimeout(function () {
                let second = prod.getElementsByClassName("second")[0]
                second.className = "second active"

                setTimeout(function () {
                  prod.className = ""
                  $vue.rolling()
                }, duration)
              }, pause)
            }, duration)
          }, pause)
        }, duration)
      }, pause)
    },
    rolling: function () {
      setTimeout(function () {
        let rolling = document.getElementById("rolling")
        rolling.className = "active"

        setTimeout(function () {
          rolling.className = ""

          let returns = document.getElementById("return")
          returns.className = "active"
        }, roll)
      }, pause)
    }
  },
  route: {
    data ({ to }) {
      this.api.getCredits({ id: to.params.tournament }).then(function (res) {
        console.log(res)
        let data = res.json()

        this.$set("executive", Person.fromObject(data.executive))
        this.$set("first_producers", _.map(data.producers.slice(0, 4), Person.fromObject))
        this.$set("second_producers", _.map(data.producers.slice(4, 7), Person.fromObject))
        this.$set("players", _.map(data.players, Person.fromObject))
        this.$set("archersHarmed", data.archers_harmed)
      }, function (res) {
        console.log('error when getting credits data')
        console.log(res)
      })
    }
  }
}
</script>

<style lang="scss" scoped>

@import "../variables.scss";

.green  {color: $green;}
.blue   {color: $blue;}
.pink   {color: $pink;}
.orange {color: $orange;}
.white  {color: $white;}
.yellow {color: $yellow;}
.cyan   {color: $cyan;}
.purple {color: $purple;}
.red    {color: $red;}


h1 {
  font-size: 4em;
}


#executive, #producers, #producer > div, #producer > h1 {
  transition: opacity 2s ease-in-out;
  opacity: 0;
}

#executive {
  position: absolute;
  top: 25%;
  right: 10%;
  left:  10%;
  bottom: 20%;
  text-align: center;

  &.active {
    opacity: 1;
    display: block;
  }

  img {
    float: right;
    height: 400px;
    width:  400px;
    object-fit: cover;
    border-radius: 100%;
    box-shadow: 2px 3px 6px rgba(0,0,0,0.5);
    background-color: rgba(10,12,14,0.3);
  }

  h1 {
    font-size: 10em;
    margin: 0.3em;
    text-shadow: 3px 3px 3px rgba(0,0,0,0.7);
  }

  h2 {
    font-size: 3em;
    margin-bottom: -1em;
    text-shadow: 1px 1px 1px rgba(0,0,0,0.7);
    color: #ccc;
  }

  h3 {
    font-size: 3em;
    margin-top: -1em;
  }
}

#producers {
  position: relative;

  &.active {
    opacity: 1;
  }
  .first, .second {
    position: absolute;
    top: 100%;
    right: 10%;
    left:  10%;
    bottom: 20%;
    transition: opacity 2s ease-in-out;
    opacity: 0;
    &.active {
      opacity: 1;
    }
  }

  > h1 {
    color: #aaa;
    font-size: 8em;
    margin: 0.5em;
  }

  /* First element is floated left */
  .first .producer {
    &:nth-child(odd) {
      h1, h3 {text-align: left;}
      img {float: left;}
    }
    &:nth-child(even) {
      h1, h3 {text-align: right;}
      img {float: right;}
    }
  }

  /* First element is floated right */
  .second .producer {
    &:nth-child(even) {
      h1, h3 {text-align: left;}
      img {float: left;}
    }
    &:nth-child(odd) {
      h1, h3 {text-align: right;}
      img {float: right;}
    }
  }

  .producer {
    width: 80vw;
    margin: 0px auto;
    height: 200px;
    margin-bottom: -100px;


    img {
      height: 200px;
      width:  200px;
      object-fit: cover;
      border-radius: 100%;
      box-shadow: 2px 2px 3px rgba(0,0,0,0.5);
      background-color: rgba(10,12,14,0.3);
    }

    div {
      margin: 0px 225px;
    }

    h1 {
      font-size: 5.5em;
      text-shadow: 3px 3px 3px rgba(0,0,0,0.7);
    }

    h3 {
      font-size: 3em;
      margin-top: -1.3em;
    }
  }
}

@keyframes credits {
  0% { top:  100%; }
  100% {
    top: -500%;
    display: none;
  }
}

#rolling {
  display: none;
  position: absolute;

  &.active {
    animation: 120s credits linear;
    display: block;
  }

  > h1 {
    margin-top: 30px;
    font-size: 15em;
    text-shadow: 3px 3px 3px rgba(0,0,0,0.7);
  }
  > h2 {
    margin-top: -2.2em;
    margin-bottom: 5em;
    font-size: 5em;
    text-align: center;
    text-shadow: 3px 3px 3px rgba(0,0,0,0.7);
    color: #aaa;
  }

  #players {
    width: 80%;
    margin: 0px auto;

    .player {
      width: 30%;
      margin: 1%;
      height: 150px;
      margin-bottom: 50px;
      float: left;

      h1, h3 {
        text-align: left;
      }
      img {
        float: left;
      }

      img {
        height: 150px;
        width:  150px;
        object-fit: cover;
        border-radius: 100%;
        box-shadow: 2px 2px 3px rgba(0,0,0,0.5);
        background-color: rgba(10,12,14,0.3);
      }

      div {
        height: 100%;
        margin: 0px 0px 0px 170px;
        display: flex;
        align-items: center;
        /* justify-content: center; */

        h1 {
          display: block;
          font-size: 3em;
          text-shadow: 3px 3px 3px rgba(0,0,0,0.7);
        }
      }
    }
  }
}

#thanks {
  margin: 6em auto;
  width: 40%;

  div {
    height: 220px;
    display: flex;
    align-items: center;

    img {
      float: left;
      margin-right: 60px;
    }
    p {
      width: 47%;
      font-size: 1.75em;
      text-align: center;
    }
  }
}

#harmed {
  margin-top: 6em;
  clear: both;
  text-align: center;
  font-size: 3em;

  span {
    text-shadow: 2px 2px 2px rgba(0,0,0,0.7);
  }
}

#return {
  transition: opacity 2s ease-in-out;
  opacity: 0;
  position: absolute;
  top: 0%;
  right: 10%;
  left:  10%;
  bottom: 20%;
  margin-top: 10em;
  text-align: center;

  &.active {
    opacity: 1;
  }

  > * {
    text-shadow: 2px 2px 4px rgba(0,0,0,0.7);
  }

  h1 {
    font-size: 14em;
    margin: 10px 0 -100px;
  }
  h2 {
    font-size: 6em;
  }
  h3 {
    font-size: 3em;
  }
  h4 {
    color: #777;
    font-size: 2em;
  }

}
</style>
