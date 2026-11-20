<template>
  <span class="yc-odo" role="img" :aria-label="text">
    <template v-for="(ch, i) in text" :key="text.length - i">
      <span v-if="ch < '0' || ch > '9'" class="yc-odo-sep" aria-hidden="true">{{ ch }}</span>
      <span v-else class="yc-odo-col" aria-hidden="true">
        <span
          class="yc-odo-strip"
          :style="{ transform: `translateY(-${(rolled ? Number(ch) : 0) * 10}%)`, transitionDelay: `${(text.length - i) * 35}ms` }"
        >
          <span v-for="n in 10" :key="n">{{ n - 1 }}</span>
        </span>
      </span>
    </template>
  </span>
</template>

<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'

// A number whose digits roll into place like a mileage counter.
const props = defineProps<{ value: number }>()

const text = computed(() => Math.round(props.value).toLocaleString('en-US'))
const rolled = ref(false)

onMounted(() => {
  requestAnimationFrame(() => {
    rolled.value = true
  })
})
</script>

<style scoped>
.yc-odo { display: inline-flex; flex-wrap: wrap; justify-content: center; line-height: 1.08; }
.yc-odo-col { display: inline-block; height: 1.08em; overflow: hidden; }
.yc-odo-strip { display: flex; flex-direction: column; transition: transform 1.4s cubic-bezier(0.2, 0.75, 0.2, 1); }
.yc-odo-strip > span { display: block; height: 1.08em; }
@media (prefers-reduced-motion: reduce) {
  .yc-odo-strip { transition: none; }
}
</style>
