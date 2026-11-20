<template>
  <section class="yc-surface yc-dog" :class="{ hungry: !garden.dog_fed_today }">
    <svg class="yc-dog-face" viewBox="0 0 64 64" role="img" :aria-label="t('yc.garden.dog.name')">
      <path class="ear" d="M14 12 8 34l12-6zM50 12l6 22-12-6z" />
      <rect class="head" x="12" y="14" width="40" height="38" rx="16" />
      <template v-if="garden.dog_fed_today">
        <circle class="eye" cx="25" cy="30" r="3" /><circle class="eye" cx="39" cy="30" r="3" />
      </template>
      <path v-else class="line" d="M21 30h8M35 30h8" />
      <ellipse class="eye" cx="32" cy="39" rx="4" ry="3" />
      <path class="line" :d="garden.dog_fed_today ? 'M25 44q7 6 14 0' : 'M26 47q6-4 12 0'" />
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
        <div v-if="open === n.id" class="yc-nb-crops">
          <p v-if="!visiting" class="yc-nb-none">{{ t('yc.garden.neighbours.loading') }}</p>
          <p v-else-if="!visiting.crops.length" class="yc-nb-none">{{ t('yc.garden.neighbours.noCrops') }}</p>
          <article v-for="crop in visiting?.crops ?? []" :key="crop.id" class="yc-nb-crop">
            <YcPlant class="yc-nb-plant" :stage="stageOf(crop)" :label="t(`yc.garden.size.${crop.size}`)" />
            <p class="yc-nb-size">{{ t(`yc.garden.size.${crop.size}`) }}</p>
            <p class="yc-nb-state">{{ stateOf(crop) }}</p>
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
  appStore.showSuccess(t('yc.garden.waterDone', { minutes: props.garden.water_minutes }))
})
const steal = (crop: GardenNeighbourCrop) => act(crop.id, async () => {
  const result = await stealCrop(crop.id)
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
.yc-dog { display: flex; align-items: center; gap: 18px; margin-top: 16px; padding: 20px 26px; }
.yc-dog-face { flex: none; width: 60px; height: 60px; }
.yc-dog-face .head { fill: var(--yc-lime); stroke: var(--yc-ink); stroke-width: 2; }
.yc-dog-face .ear, .yc-dog-face .eye { fill: var(--yc-ink); }
.yc-dog-face .line { fill: none; stroke: var(--yc-ink); stroke-width: 2.4; stroke-linecap: round; }
.yc-dog.hungry .yc-dog-face .head { fill: var(--yc-fill); }
.dark .yc .yc-dog-face .head { stroke: none; }
.dark .yc .yc-dog-face .eye { fill: var(--yc-on-lime); }
.dark .yc .yc-dog-face .ear { fill: var(--yc-lime); }
.dark .yc .yc-dog-face .line { stroke: var(--yc-on-lime); }
.dark .yc .yc-dog.hungry .yc-dog-face .eye { fill: var(--yc-mut); }
.dark .yc .yc-dog.hungry .yc-dog-face .ear { fill: var(--yc-fill); }
.dark .yc .yc-dog.hungry .yc-dog-face .line { stroke: var(--yc-mut); }
.yc-dog-text { flex: 1; min-width: 0; }
.yc-dog-text h2, .yc-nb-head h2, .yc-feed h2, .yc-ranks h2, .yc-invite h2 { font-size: 17px; font-weight: 800; }
.yc-dog-text p { margin-top: 4px; font-size: 14.5px; line-height: 1.55; color: var(--yc-mut); }
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
.yc-nb-crops { display: grid; grid-template-columns: repeat(auto-fill, minmax(150px, 1fr)); gap: 12px; padding: 4px 0 18px; }
.yc-nb-crops > .yc-nb-none { grid-column: 1 / -1; padding: 0; }
.yc-nb-crop { display: flex; flex-direction: column; align-items: center; gap: 4px; padding: 16px 12px; border-radius: 16px; background: var(--yc-fill); text-align: center; }
.yc-nb-plant { width: 72px; height: 72px; }
.yc-nb-size { font-size: 15px; font-weight: 800; }
.yc-nb-state { min-height: 20px; font-size: 13.5px; color: var(--yc-mut); }
.yc-nb-crop .yc-btn { margin-top: 6px; }
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
  .yc-dog { flex-wrap: wrap; padding: 18px 20px; }
  .yc-nb, .yc-feed, .yc-ranks, .yc-invite { padding: 18px 20px; }
  .yc-ranks-cols { grid-template-columns: 1fr; gap: 18px; }
  .yc-invite { flex-direction: column; align-items: stretch; }
  .yc-nb-head { flex-direction: column; gap: 2px; }
}
</style>
