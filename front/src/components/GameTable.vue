<script setup lang="ts">
import { onMounted } from 'vue';
import { useGameStore } from '../stores/game';
import PokerCard from './PokerCard.vue';

const gameStore = useGameStore();

onMounted(() => {
  gameStore.initializeWebSocket();
});
</script>
<template>
  <div class="game-table">
    <div class="player-hand">
      <PokerCard
        v-for="card in gameStore.playerHand"
        :key="card.value"
        :card="card"
        class="poker-card"
      />
    </div>
  </div>
</template>

<style scoped>
.player-hand {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(100px, 1fr)); /* 设置每张卡片的最小宽度 */
  gap: 10px; /* 卡片之间的间距 */
  padding: 10px;

}

.poker-card {
  display: inline-block;
  /* 如果需要，可以为每张卡片添加额外样式 */
}
</style>
