<script setup lang="ts">
import GameTable from './components/GameTable.vue'
</script>

<template>
  <div class="app">
    <header>
      <h1>扑克游戏</h1>
    </header>
    <main>
      <GameTable />
    </main>
  </div>
</template>

<style>
.app {
  min-height: 100vh;
  background-color: #1a1a1a;
  color: white;
  padding: 20px;
}

header {
  text-align: center;
  margin-bottom: 30px;
}

h1 {
  color: #42b983;
  font-size: 2.5em;
  margin: 0;
}

main {
  max-width: 1200px;
  margin: 0 auto;
}

* {
  margin: 0;
  padding: 0;
  box-sizing: border-box;
}

body {
  font-family: Arial, sans-serif;
  line-height: 1.6;
}
</style>
