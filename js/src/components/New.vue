<template>
  <div>
    <header>
      <div class="content">
        <div class="title">Drunken TowerFall / New tournament</div>
      </div>
      <div class="clear"></div>
    </header>

    <div class="clear"></div>

    <form v-on:submit="create">
      <label for="name">Name</label>
      <input class="text name" v-model="name" name="name" type="text" value="" placeholder="Name"/>
      <label for="id">Id</label>
      <input class="text id" v-model="id" name="id" type="text" value="" placeholder="ID"/>
      <label for="scheduled">Scheduled to start at</label>
      <input class="text" v-date="scheduled"/>
      <div class="fake-checkbox">
        <label for="checkbox">Fake</label>
        <input type="checkbox" id="checkbox" v-model="fake">
      </div>

      <input class="submit" type="submit"/>
    </form>

    <div class="button">
    </div>
  </div>
</template>

<script>
import moment from "moment"

export default {
  name: 'New',

  data () {
    return {
      name: '',
      id: '',
      scheduled: null,
      fake: false,
    }
  },

  created () {
    this.scheduled = moment().hour(20).minute(0).second(0).format()
  },

  methods: {
    create (event) {
      event.preventDefault()

      var payload = {
        name: this.name,
        id: this.id,
        scheduled: moment(this.scheduled),
        fake: this.fake,
      }

      this.$http.post('/api/towerfall/new/', payload).then((res) => {
        var j = res.json()
        this.$router.push('/towerfall' + j.redirect)
      }, (res) => {
        console.warn("creating tournament failed", res)
      })
    },
  }
}
</script>

<style lang="scss">
@import "../pikaday.scss";

form {
  display: flex;
  flex-direction: column;
  align-items: center;

  label {
    font-size: 1.5em;
    margin: 1.5em 0 0.5em 0;
  }

  div.fake-checkbox {
    margin: 1.5em 0 0.5em 0;
  }

  input.text {
    width: 11em;
  }

  input.submit {
    border: none;
    margin: 1.5em;
    background-color: #333333;
    font-size: 2em;
    width: auto;
  }
}
</style>
