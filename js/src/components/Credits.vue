<template>
  <div>
    <div id="executive">
      <img alt="thiderman" src="https://scontent-arn2-1.xx.fbcdn.net/v/t1.0-9/18920658_1655829387790533_6528694072985925355_n.jpg?oh=8ba01db0f250fe4c1bbe90de09931d60&oe=59EBF2DB"/>
      <h2>Executive Producer</h2>
      <h1>Lowe Thiderman</h1>
      <h3 class="purple">thiderman</h3>
    </div>

    <div id="producers">
      <h1>Producers</h1>

      <div class="first">
        <div class="producer">
          <img alt="thiderman" src="https://scontent-arn2-1.xx.fbcdn.net/v/t1.0-9/18920658_1655829387790533_6528694072985925355_n.jpg?oh=8ba01db0f250fe4c1bbe90de09931d60&oe=59EBF2DB"/>
          <div>
            <h1>Magnus Ulenius</h1>
            <h3 class="red">GoosE</h3>
          </div>
        </div>
        <div class="producer">
          <img alt="thiderman" src="https://scontent-arn2-1.xx.fbcdn.net/v/t1.0-9/18920658_1655829387790533_6528694072985925355_n.jpg?oh=8ba01db0f250fe4c1bbe90de09931d60&oe=59EBF2DB"/>
          <div>
            <h1>Ida Andreasson</h1>
            <h3 class="blue">Queen Obscene</h3>
          </div>
        </div>
        <div class="producer">
          <img alt="thiderman" src="https://scontent-arn2-1.xx.fbcdn.net/v/t1.0-9/18920658_1655829387790533_6528694072985925355_n.jpg?oh=8ba01db0f250fe4c1bbe90de09931d60&oe=59EBF2DB"/>
          <div>
            <h1>Karl-Astrid Krantz</h1>
            <h3 class="pink">QRL-Astrid</h3>
          </div>
        </div>
        <div class="producer">
          <img alt="thiderman" src="https://scontent-arn2-1.xx.fbcdn.net/v/t1.0-9/18920658_1655829387790533_6528694072985925355_n.jpg?oh=8ba01db0f250fe4c1bbe90de09931d60&oe=59EBF2DB"/>
          <div>
            <h1>Jonathan Gustafsson</h1>
            <h3 class="orange">Gishten</h3>
          </div>
        </div>
      </div>

      <div class="second">
        <div class="producer">
          <img alt="thiderman" src="https://scontent-arn2-1.xx.fbcdn.net/v/t1.0-9/18920658_1655829387790533_6528694072985925355_n.jpg?oh=8ba01db0f250fe4c1bbe90de09931d60&oe=59EBF2DB"/>
          <div>
            <h1>Mike Goeppner</h1>
            <h3 class="white">Skolpadda</h3>
          </div>
        </div>
        <div class="producer">
          <img alt="thiderman" src="https://scontent-arn2-1.xx.fbcdn.net/v/t1.0-9/18920658_1655829387790533_6528694072985925355_n.jpg?oh=8ba01db0f250fe4c1bbe90de09931d60&oe=59EBF2DB"/>
          <div>
            <h1>Li Larsson</h1>
            <h3 class="red">Moijra</h3>
          </div>
        </div>
        <div class="producer">
          <img alt="thiderman" src="https://scontent-arn2-1.xx.fbcdn.net/v/t1.0-9/18920658_1655829387790533_6528694072985925355_n.jpg?oh=8ba01db0f250fe4c1bbe90de09931d60&oe=59EBF2DB"/>
          <div>
            <h1>Daniel Dala Thiderman</h1>
            <h3 class="blue">Dalan</h3>
          </div>
        </div>
      </div>

    </div>

  </div>
</template>

<script>
let pause = 2000
let duration = 6000

export default {
  created: function () {
    setTimeout(function () {
      let exec = document.getElementById("executive")
      exec.className = "active"

      setTimeout(function () {
        exec.className = ""

        setTimeout(function () {
          let prod = document.getElementById("producers")
          prod.className = "active"
        }, pause)
      }, duration)
    }, pause)
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

#executive, #producer > div, #producer > h1 {
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
  .first, .second {
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

  .second {
    display: none;
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
</style>
