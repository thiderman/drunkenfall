<template>
  <div>
    <header>
      <div class="content">
        <div class="title">Drunken TowerFall</div>
      </div>
      <div class="links">
      </div>
      <div class="clear"></div>
    </header>

    <form @submit="submit">
      <input name="id" v-model="id" type="hidden" value=""/>

      <label for="nick">Display name</label>
      <input class="text" name="nick" v-model="nick" type="text" value=""/>

      <label for="name">Full name</label>
      <input class="text smaller" name="name" v-model="name" type="text" value=""/>

      <div id="join-images" class="images">
        <label for="images">Preferred archer</label>

        <input type="image" @click="character" id="green" src="/static/img/green-selected.png">
        <input type="image" @click="character" id="blue" src="/static/img/blue-selected.png">
        <input type="image" @click="character" id="pink" src="/static/img/pink-selected.png">
        <input type="image" @click="character" id="orange" src="/static/img/orange-selected.png">
        <input type="image" @click="character" id="white" src="/static/img/white-selected.png">
        <input type="image" @click="character" id="yellow" src="/static/img/yellow-selected.png">
        <input type="image" @click="character" id="cyan" src="/static/img/cyan-selected.png">
        <input type="image" @click="character" id="purple" src="/static/img/purple-selected.png">
        <input type="image" @click="character" id="red" src="/static/img/red-selected.png">
      </div>

      <input type="submit" class="submit" value="Go go go!"
        v-bind:class="{'isReady': isReady}"/>
    </form>

  </div>
</template>

<script>
export default {
  name: 'FacebookFinalize',

  data () {
    return {
      color: "",
    }
  },

  methods: {
    character (event) {
      event.preventDefault()
      this.clear()
      var img = event.srcElement
      this.$data.color = img.id
      img.className = 'selected'
    },
    clear () {
      var elem = document.getElementById('join-images').getElementsByTagName('input')
      for (var i = 0; i < elem.length; i++) {
        var item = elem[i]
        item.className = ''
      }
    },
    submit (event) {
      let $vue = this
      event.preventDefault()
      if (this.isReady === false) {
        console.log('Did not fill in all details')
        return
      }

      var payload = {
        id: this.id,
        name: this.name,
        nick: this.nick,
        color: this.color
      }

      this.$http.post('/api/facebook/register', payload).then((res) => {
        var j = res.json()
        $vue.$router.push('/towerfall' + j.redirect)
      }, (res) => {
        console.log('fail')
        console.log(res)
      })
    }
  },

  computed: {
    isReady: function () {
      return this.$data.color !== '' && this.$data.name !== ''
    },
    id () {
      return this.$route.query.id
    },
    name () {
      return this.$route.query.name
    },
    nick () {
      return this.$route.query.nick
    },
  },
}
</script>

<style lang="scss" scoped>
form {
  div {
    width: 80%;
    margin: 10px auto;
  }

  label {
    display: block;
    text-align: center;
    font-size: 1.5em;
    margin: 1.5em 0 0.5em 0;
  }

  input.text {
    width: 11em;
    margin: 0 auto;
    display: inherit;
    text-align: center;
    box-shadow: inset 5px 5px 4px rgba(0,0,0,0.1);
    border-radius: 0.3em;

    &:focus {
      outline: 0;
    }

    &.smaller {
      font-size: 1em;
      width: 16em;
    }
  }

  input.submit {
    margin: 20px auto;
    width: 250px;
    border: none;
    font-size: 2em;
    transition: 1.0s;
    background-color: #333333;

    &.isReady {
      background-color: #405060;
    }
  }
  h2 {
    text-align: center;
  }

  .images {
    margin-top: 3em;

    #green {background-image: url(/static/img/green-selected.png);}
    #blue {background-image: url(/static/img/blue-selected.png);}
    #pink {background-image: url(/static/img/pink-selected.png);}
    #orange {background-image: url(/static/img/orange-selected.png);}
    #white {background-image: url(/static/img/white-selected.png);}
    #yellow {background-image: url(/static/img/yellow-selected.png);}
    #cyan {background-image: url(/static/img/cyan-selected.png);}
    #purple {background-image: url(/static/img/purple-selected.png);}

    .selected {
      opacity: 1;
    }

    input {
      image-rendering: pixelated;
      width: 9.5%;
      margin: 0.3%;
      opacity: 0.3;
      transition: 0.3s;
      cursor: pointer;
    }
  }
}
</style>
