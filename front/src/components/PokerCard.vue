<template>
  <div class="poker-card">
    <div class="card-inner">
      <div class="card-front">
        <div class="card-value">{{ card.value }}</div>
        <div class="card-suit">{{ card.suit }}</div>
        <div class="card-rank">{{ card.rank }}</div>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
const props = defineProps<{
  card:object;
}>();
// console.log("value:",props.card);


</script>

<style scoped>
.poker-card {
  width: 120px;
  height: 180px;
  perspective: 1000px;
  cursor: pointer;
}

.card-inner {
  position: relative;
  width: 100%;
  height: 100%;
  text-align: center;
  transition: transform 0.6s;
  transform-style: preserve-3d;
}

.is-flipped .card-inner {
  transform: rotateY(180deg);
}

.card-front,
.card-back {
  position: absolute;
  width: 100%;
  height: 100%;
  backface-visibility: hidden;
  border-radius: 10px;
  box-shadow: 0 4px 8px rgba(138, 55, 55, 0.1);
}

.card-front {
  background-color: white;
  border: 2px solid #ddd;
  display: flex;
  flex-direction: column;
  justify-content: center;
  align-items: center;
}

.card-back {
  background-color: #2c3e50;
  transform: rotateY(180deg);
}

.card-value {
  color:black;
  font-size: 24px;
  font-weight: bold;
  margin-bottom: 5px;
}

.card-suit {
  color:red;
  font-size: 32px;
}
.card-rank {
  color:orange;
  font-size: 32px;
}
</style> 