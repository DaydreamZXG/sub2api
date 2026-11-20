<template>
  <section class="yc-dog" :class="{ hungry: !garden.dog_fed_today }">
    <svg class="yc-dog-yard" viewBox="0 0 132 84" role="img" :aria-label="t('yc.garden.dog.name')">
      <path class="roof" d="M4 40 38 10l34 30z" />
      <path class="wall" d="M12 40h52v40H12z" />
      <path class="door" d="M28 80V60a10 10 0 0 1 20 0v20z" />
      <g :transform="garden.dog_fed_today ? 'translate(78 22)' : 'translate(78 34)'">
        <path class="ear" d="M6 6 2 24l10-5zM42 6l4 18-10-5z" />
        <rect class="head" x="4" y="8" width="40" height="38" rx="16" />
        <template v-if="garden.dog_fed_today">
          <circle class="eye" cx="17" cy="24" r="3" /><circle class="eye" cx="31" cy="24" r="3" />
        </template>
        <path v-else class="line" d="M13 24h8M27 24h8" />
        <ellipse class="eye" cx="24" cy="33" rx="4" ry="3" />
        <path class="line" :d="garden.dog_fed_today ? 'M17 38q7 6 14 0' : 'M18 41q6-4 12 0'" />
      </g>
    </svg>
    <div class="yc-dog-text">
      <h2>{{ t('yc.garden.dog.name') }}</h2>
      <p v-if="garden.dog_fed_today">{{ t('yc.garden.dog.fed', { percent: garden.dog_catch_percent }) }}</p>
      <p v-else>{{ t('yc.garden.dog.hungry') }}</p>
    </div>
    <button v-if="!garden.dog_fed_today" type="button" class="yc-btn yc-btn-lime yc-btn-sm" :disabled="busy === 'dog'" @click="feedTheDog">
      {{ t('yc.garden.dog.feed') }}
    </button>
    <span v-else class="yc-badge yc-badge-lime">{{ t('yc.garden.dog.onGuard') }}</span>
  </section>

  <section ref="neighboursEl" class="yc-surface yc-nb">
    <header class="yc-nb-head">
      <h2>{{ t('yc.garden.neighbours.title') }}</h2>
      <p>{{ t('yc.garden.neighbours.triesLeft', { left: triesLeft, limit: garden.daily_steal_limit }) }}</p>
    </header>
    <p class="yc-nb-rules">
      {{ t('yc.garden.neighbours.rules', { protect: garden.protect_hours, percent: garden.steal_percent, minutes: garden.water_minutes }) }}
    </p>
    <p v-if="neighboursFailed" class="yc-nb-none">{{ t('yc.garden.failed') }}</p>
    <p v-else-if="neighbours && !neighbours.length" class="yc-nb-none">{{ t('yc.garden.neighbours.none') }}</p>
    <ul v-else-if="neighbours" class="yc-nb-list">
      <li v-for="n in neighbours" :key="n.id">
        <button type="button" class="yc-nb-row" :aria-expanded="open === n.id" @click="toggle(n.id)">
          <i>{{ n.name.charAt(0).toUpperCase() }}</i>
          <span class="yc-nb-name">{{ n.name }}</span>
          <span v-if="n.took_from_you" class="yc-badge yc-badge-lime">{{ t('yc.garden.neighbours.raided') }}</span>
          <span v-else-if="n.pinned" class="yc-badge">{{ t('yc.garden.neighbours.pinned') }}</span>
          <span class="yc-nb-count">{{ t('yc.garden.neighbours.crops', { count: n.crops }) }}</span>
        </button>
        <div v-if="open === n.id" class="yc-nb-crops yc-field-inset">
          <p v-if="!visiting" class="yc-nb-none">{{ t('yc.garden.neighbours.loading') }}</p>
          <p v-else-if="!visiting.crops.length" class="yc-nb-none">{{ t('yc.garden.neighbours.noCrops') }}</p>
          <article
            v-for="crop in visiting?.crops ?? []" :key="crop.id" class="yc-nb-crop yc-plot"
            :class="[{ ripe: crop.can_steal }, fx[crop.id] ? `fx-${fx[crop.id]}` : '']"
          >
            <div class="yc-plot-bed">
              <YcPlant class="yc-plot-plant" :stage="stageOf(crop)" :label="t(`yc.garden.size.${crop.size}`)" />
            </div>
            <p class="yc-plot-name">{{ t(`yc.garden.size.${crop.size}`) }}</p>
            <p class="yc-plot-state">{{ stateOf(crop) }}</p>
            <button v-if="crop.can_water" type="button" class="yc-btn yc-btn-ghost yc-btn-sm" :disabled="busy === crop.id" @click="water(crop)">
              {{ t('yc.garden.water') }}
            </button>
            <button v-else-if="crop.can_steal" type="button" class="yc-btn yc-btn-lime yc-btn-sm" :disabled="busy === crop.id || triesLeft <= 0" @click="steal(crop)">
              {{ t('yc.garden.steal') }}
            </button>
          </article>
        </div>
      </li>
    </ul>
  </section>

  <section v-if="feed.length" class="yc-surface yc-feed">
    <h2>{{ t('yc.garden.feed.title') }}</h2>
    <ul>
      <li v-for="(item, i) in feed" :key="i">
        <span>{{ feedText(item) }}</span>
        <button v-if="canGoBack(item)" type="button" class="yc-feed-go" @click="goTo(item.user_id)">{{ t('yc.garden.feed.visit') }}</button>
        <time>{{ when(item.created_at) }}</time>
      </li>
    </ul>
  </section>

  <section v-if="ranks && (ranks.thieves.length || ranks.dogs.length)" class="yc-surface yc-ranks">
    <h2>{{ t('yc.garden.ranks.title') }}</h2>
    <div class="yc-ranks-cols">
      <div v-for="board in boards" :key="board.key">
        <h3>{{ t(`yc.garden.ranks.${board.key}`) }}</h3>
        <p v-if="!board.rows.length" class="yc-nb-none">{{ t('yc.garden.ranks.none') }}</p>
        <ol v-else>
          <li v-for="(row, i) in board.rows" :key="i" :class="{ me: row.me }">
            <b>{{ i + 1 }}</b>
            <span class="yc-nb-name">{{ row.name }}</span>
            <span v-if="row.me" class="yc-badge yc-badge-lime">{{ t('yc.garden.ranks.me') }}</span>
            <em class="yc-num">{{ board.key === 'thieves' ? yuan(row.amount) : t('yc.garden.ranks.caught', { count: row.count }) }}</em>
          </li>
        </ol>
      </div>
    </div>
  </section>

  <section v-if="inviteLink" class="yc-surface yc-invite">
    <div>
      <h2>{{ t('yc.garden.invite.title') }}</h2>
      <p>{{ t('yc.garden.invite.lead', { amount: yuan(garden.invite_crop_amount) }) }}</p>
      <code>{{ inviteLink }}</code>
    </div>
    <button type="button" class="yc-btn yc-btn-lime yc-btn-sm" @click="copyToClipboard(inviteLink, t('yc.garden.invite.copied'))">
      {{ t('yc.garden.invite.copy') }}
    </button>
  </section>
</template>

<script setup lang="ts">
import { computed, nextTick, onMounted, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import {
  feedDog, getGardenFeed, getGardenNeighbours, getGardenRanks, getNeighbourGarden, stealCrop, waterCrop,
  type GardenFeedItem, type GardenNeighbour, type GardenNeighbourCrop, type GardenNeighbourGarden, type GardenRanks,
  type MyGarden
} from '@/api/garden'
import userAPI from '@/api/user'
import { useClipboard } from '@/composables/useClipboard'
import YcPlant from '@/components/yecai/YcPlant.vue'
import { useYcMoney } from '@/composables/useYcMoney'
import { useAppStore } from '@/stores'
import { extractApiErrorCode } from '@/utils/apiError'

// What neighbours do in each other's gardens: the dog, today's neighbours and their crops, and the activity feed.
const props = defineProps<{ garden: MyGarden; now: number }>()
const emit = defineEmits<{ changed: [] }>()

const { t, te, locale } = useI18n()
const { yuan } = useYcMoney()
const appStore = useAppStore()

const neighbours = ref<GardenNeighbour[] | null>(null)
const neighboursFailed = ref(false)
const open = ref<number | null>(null)
const visiting = ref<GardenNeighbourGarden | null>(null)
const feed = ref<GardenFeedItem[]>([])
const busy = ref<number | 'dog' | null>(null)
// What just happened to a neighbour's bed, shown for a moment as a small animation.
const fx = ref<Record<number, 'water' | 'steal' | 'caught'>>({})
function play(cropId: number, kind: 'water' | 'steal' | 'caught') {
  fx.value = { ...fx.value, [cropId]: kind }
  setTimeout(() => {
    const { [cropId]: _done, ...rest } = fx.value
    fx.value = rest
  }, 900)
}
const ranks = ref<GardenRanks | null>(null)
const neighboursEl = ref<HTMLElement | null>(null)
const affCode = ref('')
const { copyToClipboard } = useClipboard()

const boards = computed(() => [
  { key: 'thieves' as const, rows: ranks.value?.thieves ?? [] },
  { key: 'dogs' as const, rows: ranks.value?.dogs ?? [] }
])

// Shown only when inviting earns a crop.
const inviteLink = computed(() =>
  affCode.value ? `${window.location.origin}/register?aff=${encodeURIComponent(affCode.value)}` : ''
)

// Someone who took from the user, or was caught trying, is listed among the neighbours for a day.
const canGoBack = (item: GardenFeedItem) =>
  !item.mine && item.kind !== 'water' && !!neighbours.value?.some((n) => n.id === item.user_id)

async function goTo(id: number) {
  if (open.value !== id) toggle(id)
  await nextTick()
  neighboursEl.value?.scrollIntoView?.({ behavior: 'smooth', block: 'start' })
}

const triesLeft = computed(() => Math.max(0, props.garden.daily_steal_limit - props.garden.steals_today))

const isRipe = (crop: GardenNeighbourCrop) => new Date(crop.ripe_at).getTime() <= props.now
function stageOf(crop: GardenNeighbourCrop): number {
  if (isRipe(crop)) return 4
  const planted = new Date(crop.planted_at).getTime()
  const span = new Date(crop.ripe_at).getTime() - planted
  return span <= 0 ? 3 : Math.min(3, Math.max(0, Math.floor(((props.now - planted) / span) * 4)))
}

function stateOf(crop: GardenNeighbourCrop): string {
  if (crop.tried === 'steal') return t('yc.garden.state.taken')
  if (crop.tried === 'caught') return t('yc.garden.state.caught')
  if (!isRipe(crop)) return t('yc.garden.growing')
  if (new Date(crop.open_at).getTime() > props.now) return t('yc.garden.state.protected')
  return t(crop.can_steal ? 'yc.garden.state.open' : 'yc.garden.state.pickedClean')
}

function feedText(item: GardenFeedItem): string {
  const side = item.mine ? 'mine' : 'theirs'
  return t(`yc.garden.feed.${side}.${item.kind}`, { name: item.name, amount: yuan(item.amount) })
}

function when(value: string): string {
  return new Date(value).toLocaleString(locale.value, { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' })
}

function failedWith(error: unknown) {
  const key = `yc.garden.errors.${extractApiErrorCode(error)}`
  appStore.showError(te(key) ? t(key) : t('yc.garden.errors.other'))
}

async function loadFeed() {
  try {
    feed.value = await getGardenFeed()
  } catch (error) {
    console.error('Failed to load the garden feed:', error)
  }
}

async function loadRanks() {
  try {
    ranks.value = await getGardenRanks()
  } catch (error) {
    console.error('Failed to load the garden boards:', error)
  }
}

async function loadInvite() {
  if (props.garden.invite_crop_amount <= 0 || !appStore.cachedPublicSettings?.affiliate_enabled) return
  try {
    affCode.value = (await userAPI.getAffiliateDetail()).aff_code
  } catch (error) {
    console.error('Failed to load the invite link:', error)
  }
}

async function visit(id: number) {
  try {
    visiting.value = await getNeighbourGarden(id)
  } catch (error) {
    open.value = null
    failedWith(error)
  }
}

function toggle(id: number) {
  visiting.value = null
  open.value = open.value === id ? null : id
  if (open.value) void visit(id)
}

async function act(key: number | 'dog', action: () => Promise<void>) {
  busy.value = key
  try {
    await action()
  } catch (error) {
    failedWith(error)
  } finally {
    busy.value = null
    if (open.value) await visit(open.value)
    await Promise.all([loadFeed(), loadRanks()])
    emit('changed')
  }
}

const feedTheDog = () => act('dog', async () => {
  await feedDog()
  appStore.showSuccess(t('yc.garden.dog.feedDone'))
})
const water = (crop: GardenNeighbourCrop) => act(crop.id, async () => {
  await waterCrop(crop.id)
  play(crop.id, 'water')
  appStore.showSuccess(t('yc.garden.waterDone', { minutes: props.garden.water_minutes }))
})
const steal = (crop: GardenNeighbourCrop) => act(crop.id, async () => {
  const result = await stealCrop(crop.id)
  play(crop.id, result.caught ? 'caught' : 'steal')
  if (result.caught) appStore.showError(t('yc.garden.stealCaught'))
  else appStore.showSuccess(t('yc.garden.stealDone', { amount: yuan(result.amount) }))
})

onMounted(async () => {
  void loadFeed()
  void loadRanks()
  void loadInvite()
  try {
    neighbours.value = await getGardenNeighbours()
  } catch (error) {
    console.error('Failed to load neighbours:', error)
    neighboursFailed.value = true
  }
})
</script>

<style scoped>
/* The yard under the field: the kennel and the dog, on the same dark ground in both themes. */
.yc-dog { display: flex; align-items: center; gap: 18px; padding: 16px 26px 20px; border-top: 1px dashed rgb(244 246 241 / 0.16); border-radius: 0 0 28px 28px; background: #0b0f0c; color: #f4f6f1; }
.dark .yc .yc-dog { box-shadow: 0 0 0 1px var(--yc-line); }
.yc-dog-yard { flex: none; width: 124px; height: 80px; }
.yc-dog-yard .roof { fill: var(--yc-lime); }
.yc-dog-yard .wall { fill: #f4f6f1; }
.yc-dog-yard .door { fill: #0b0f0c; }
.yc-dog-yard .head { fill: var(--yc-lime); }
.yc-dog-yard .ear { fill: #f4f6f1; }
.yc-dog-yard .eye { fill: #0b0f0c; }
.yc-dog-yard .line { fill: none; stroke: #0b0f0c; stroke-width: 2.4; stroke-linecap: round; }
.yc-dog.hungry .yc-dog-yard .head { fill: #5d665f; }
.yc-dog.hungry .yc-dog-yard .ear { fill: #3a423c; }
.yc-dog:not(.hungry) .yc-dog-yard g { animation: yc-dog-bob 2.8s ease-in-out infinite; }
@keyframes yc-dog-bob { 50% { translate: 0 -3px; } }
@media (prefers-reduced-motion: reduce) { .yc-dog .yc-dog-yard g { animation: none; } }
.yc-dog-text { flex: 1; min-width: 0; }
.yc-dog-text h2, .yc-nb-head h2, .yc-feed h2, .yc-ranks h2, .yc-invite h2 { font-size: 17px; font-weight: 800; }
.yc-dog-text p { margin-top: 4px; font-size: 14.5px; line-height: 1.55; color: rgb(244 246 241 / 0.66); }
.yc-nb, .yc-feed, .yc-ranks, .yc-invite { margin-top: 16px; padding: 22px 26px; }
.yc-nb-head { display: flex; align-items: baseline; justify-content: space-between; gap: 12px; }
.yc-nb-head p { font-size: 14px; color: var(--yc-mut); }
.yc-nb-rules { margin-top: 6px; font-size: 14px; line-height: 1.6; color: var(--yc-mut); }
.yc-nb-none { padding: 14px 0 4px; font-size: 14.5px; color: var(--yc-mut); }
.yc-nb-list { margin-top: 12px; }
.yc-nb-list > li { border-top: 1px solid var(--yc-line); }
.yc-nb-row { display: flex; align-items: center; gap: 12px; width: 100%; padding: 12px 0; text-align: left; font-size: 15px; }
.yc-nb-row i { display: grid; place-items: center; flex: none; width: 32px; height: 32px; border-radius: 50%; background: var(--yc-fill); font-style: normal; font-weight: 800; }
.yc-nb-name { min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-weight: 650; }
.yc-nb-count { margin-left: auto; flex: none; font-size: 14px; color: var(--yc-mut); }
.yc-nb-crops { display: grid; grid-template-columns: repeat(auto-fill, minmax(150px, 1fr)); gap: 12px; margin: 2px 0 18px; }
.yc-nb-crops > .yc-nb-none { grid-column: 1 / -1; padding: 6px 4px; color: rgb(244 246 241 / 0.66); }
.yc-nb-crop.yc-plot { min-height: 188px; }
.yc-feed ul { margin-top: 10px; }
.yc-feed li { display: flex; justify-content: space-between; gap: 16px; padding: 10px 0; border-top: 1px solid var(--yc-line); font-size: 14.5px; }
.yc-feed time { flex: none; color: var(--yc-mut); }
.yc-feed li > span { flex: 1; min-width: 0; }
.yc-feed-go { flex: none; font-weight: 700; text-decoration: underline; text-underline-offset: 3px; }
.yc-ranks-cols { display: grid; grid-template-columns: 1fr 1fr; gap: 28px; margin-top: 12px; }
.yc-ranks h3 { font-size: 14px; font-weight: 700; color: var(--yc-mut); }
.yc-ranks li { display: flex; align-items: center; gap: 10px; padding: 9px 0; border-top: 1px solid var(--yc-line); font-size: 14.5px; }
.yc-ranks ol { margin-top: 8px; }
.yc-ranks li b { flex: none; width: 20px; font-weight: 800; color: var(--yc-mut); }
.yc-ranks li.me b { color: inherit; }
.yc-ranks li em { margin-left: auto; flex: none; font-style: normal; font-weight: 700; }
.yc-invite { display: flex; align-items: center; justify-content: space-between; gap: 18px; }
.yc-invite > div { min-width: 0; }
.yc-invite p { margin-top: 4px; font-size: 14.5px; line-height: 1.55; color: var(--yc-mut); }
.yc-invite code { display: block; margin-top: 8px; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 13.5px; }
@media (max-width: 760px) {
  .yc-dog { flex-wrap: wrap; padding: 14px 18px 18px; }
  .yc-dog-yard { width: 96px; height: 62px; }
  .yc-nb-crops { grid-template-columns: repeat(2, minmax(0, 1fr)); gap: 10px; }
  .yc-nb, .yc-feed, .yc-ranks, .yc-invite { padding: 18px 20px; }
  .yc-ranks-cols { grid-template-columns: 1fr; gap: 18px; }
  .yc-invite { flex-direction: column; align-items: stretch; }
  .yc-nb-head { flex-direction: column; gap: 2px; }
}
</style>
