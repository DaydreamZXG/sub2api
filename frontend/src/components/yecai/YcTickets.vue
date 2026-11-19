<template>
  <section class="yc-surface yc-tk">
    <div class="yc-tk-head">
      <div>
        <h2 class="yc-sec-h2">{{ t('yc.tickets.title') }}</h2>
        <p class="yc-note">{{ t('yc.tickets.lead') }}</p>
      </div>
      <button v-if="!composing" type="button" class="yc-btn yc-btn-lime yc-btn-sm" @click="composing = true">{{ t('yc.tickets.new') }}</button>
    </div>

    <form v-if="composing" class="yc-tk-form" @submit.prevent="submit">
      <label>
        <span>{{ t('yc.tickets.subject') }}</span>
        <input v-model="draft.subject" class="yc-input" :maxlength="TICKET_SUBJECT_MAX" :placeholder="t('yc.tickets.subjectHint')" required />
      </label>
      <div class="yc-tk-cats" role="radiogroup" :aria-label="t('yc.tickets.category')">
        <button
          v-for="c in TICKET_CATEGORIES"
          :key="c"
          type="button"
          role="radio"
          :aria-checked="draft.category === c"
          :class="{ 'is-on': draft.category === c }"
          @click="draft.category = c"
        >{{ t(`yc.tickets.cat.${c}`) }}</button>
      </div>
      <label>
        <span>{{ t('yc.tickets.content') }}</span>
        <textarea v-model="draft.content" class="yc-input yc-tk-area" rows="5" :maxlength="TICKET_CONTENT_MAX" :placeholder="t('yc.tickets.contentHint')" required></textarea>
      </label>
      <div class="yc-tk-actions">
        <button type="submit" class="yc-btn yc-btn-lime yc-btn-sm" :disabled="busy || !canSubmit">{{ t('yc.tickets.submit') }}</button>
        <button type="button" class="yc-btn yc-btn-ghost yc-btn-sm" @click="composing = false">{{ t('common.cancel') }}</button>
      </div>
    </form>

    <p v-if="failed" class="yc-note yc-tk-empty">{{ t('yc.tickets.loadFailed') }}</p>
    <p v-else-if="loaded && !tickets.length && !composing" class="yc-note yc-tk-empty">{{ t('yc.tickets.empty') }}</p>

    <article v-for="ticket in tickets" :key="ticket.id" class="yc-tk-item" :class="{ 'is-open': openId === ticket.id }">
      <button type="button" class="yc-tk-row" :aria-expanded="openId === ticket.id" @click="toggle(ticket)">
        <i :class="`is-${ticket.status}`">{{ t(`yc.tickets.status.${ticket.status}`) }}</i>
        <b>{{ ticket.subject }}</b>
        <time>{{ formatRelativeTime(ticket.last_message_at) }}</time>
      </button>
      <div v-if="openId === ticket.id" class="yc-tk-thread">
        <p v-for="m in messages" :key="m.id" class="yc-tk-msg" :class="`from-${m.author_type}`">
          <small>{{ t(`yc.tickets.author.${m.author_type}`) }} · {{ formatRelativeTime(m.created_at) }}</small>
          {{ m.content }}
        </p>
        <form class="yc-tk-reply" @submit.prevent="send(ticket)">
          <textarea v-model="replyText" class="yc-input yc-tk-area" rows="3" :maxlength="TICKET_CONTENT_MAX" :placeholder="t(ticket.status === 'closed' ? 'yc.tickets.reopenHint' : 'yc.tickets.replyHint')"></textarea>
          <div class="yc-tk-actions">
            <button type="submit" class="yc-btn yc-btn-lime yc-btn-sm" :disabled="busy || !replyText.trim()">{{ t('yc.tickets.send') }}</button>
            <button v-if="ticket.status !== 'closed'" type="button" class="yc-btn yc-btn-ghost yc-btn-sm" :disabled="busy" @click="closeTicket(ticket)">{{ t('yc.tickets.close') }}</button>
          </div>
        </form>
      </div>
    </article>

    <button v-if="tickets.length < total" type="button" class="yc-btn yc-btn-ghost yc-btn-sm yc-tk-more" :disabled="busy" @click="load(page + 1)">{{ t('yc.tickets.more') }}</button>
  </section>
</template>

<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores'
import ticketsAPI, {
  TICKET_CATEGORIES,
  TICKET_CONTENT_MAX,
  TICKET_SUBJECT_MAX,
  type Ticket,
  type TicketCategory,
  type TicketMessage
} from '@/api/tickets'
import { extractApiErrorMessage } from '@/utils/apiError'
import { formatRelativeTime } from '@/utils/format'

const { t } = useI18n()
const appStore = useAppStore()

const tickets = ref<Ticket[]>([])
const total = ref(0)
const page = ref(1)
const loaded = ref(false)
const failed = ref(false)
const busy = ref(false)

const composing = ref(false)
const draft = reactive<{ subject: string; category: TicketCategory; content: string }>({ subject: '', category: 'usage', content: '' })
const canSubmit = computed(() => draft.subject.trim() !== '' && draft.content.trim() !== '')

const openId = ref<number | null>(null)
const messages = ref<TicketMessage[]>([])
const replyText = ref('')

async function run(action: () => Promise<void>) {
  if (busy.value) return
  busy.value = true
  try {
    await action()
  } catch (error) {
    appStore.showError(extractApiErrorMessage(error, t('yc.tickets.failed')))
  } finally {
    busy.value = false
  }
}

async function load(target: number) {
  await run(async () => {
    const result = await ticketsAPI.list(target)
    tickets.value = target === 1 ? result.items : [...tickets.value, ...result.items]
    total.value = result.total
    page.value = target
  })
}

async function openThread(ticket: Ticket) {
  const detail = await ticketsAPI.getById(ticket.id)
  ticket.status = detail.status
  messages.value = detail.messages
  replyText.value = ''
  openId.value = ticket.id
}

function toggle(ticket: Ticket) {
  if (openId.value === ticket.id) {
    openId.value = null
    return
  }
  void run(() => openThread(ticket))
}

function submit() {
  void run(async () => {
    const created = await ticketsAPI.create({ subject: draft.subject, category: draft.category, content: draft.content })
    tickets.value.unshift(created)
    total.value += 1
    Object.assign(draft, { subject: '', content: '' })
    composing.value = false
    await openThread(tickets.value[0])
  })
}

function send(ticket: Ticket) {
  void run(async () => {
    messages.value.push(await ticketsAPI.reply(ticket.id, replyText.value))
    replyText.value = ''
    ticket.status = 'open'
  })
}

function closeTicket(ticket: Ticket) {
  void run(async () => {
    await ticketsAPI.close(ticket.id)
    ticket.status = 'closed'
  })
}

onMounted(async () => {
  try {
    const result = await ticketsAPI.list(1)
    tickets.value = result.items
    total.value = result.total
  } catch (error) {
    console.error('Failed to load tickets:', error)
    failed.value = true
  } finally {
    loaded.value = true
  }
})
</script>

<style scoped>
.yc-tk { margin-top: 20px; padding: 26px 30px; }
.yc-tk-head { display: flex; align-items: flex-start; justify-content: space-between; gap: 16px; margin-bottom: 6px; }
.yc-tk-head .yc-note { margin-top: 6px; }
.yc-tk-head .yc-btn { flex: none; }
.yc-tk-form, .yc-tk-reply { display: grid; gap: 14px; }
.yc-tk-form { margin: 16px 0 6px; padding: 20px; border-radius: 18px; background: var(--yc-fill); }
.yc-tk-form label { display: grid; gap: 7px; font-size: 13.5px; font-weight: 600; color: var(--yc-mut); }
.yc-tk-area { height: auto; padding: 12px 16px; line-height: 1.6; resize: vertical; }
.yc-tk-cats { display: flex; flex-wrap: wrap; gap: 8px; }
.yc-tk-cats button { padding: 7px 14px; border-radius: 999px; background: var(--yc-paper); box-shadow: inset 0 0 0 1px var(--yc-line); font-size: 13.5px; color: var(--yc-mut); }
.yc-tk-cats button.is-on { background: var(--yc-ink); color: var(--yc-paper); box-shadow: none; }
.yc-tk-actions { display: flex; flex-wrap: wrap; gap: 10px; }
.yc-tk-empty { margin-top: 14px; }
.yc-tk-item { border-top: 1px solid var(--yc-line); }
.yc-tk-head + .yc-tk-item, .yc-tk-form + .yc-tk-item { margin-top: 14px; }
.yc-tk-row { display: flex; align-items: center; gap: 12px; width: 100%; padding: 15px 0; text-align: left; }
.yc-tk-row b { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; font-size: 15.5px; font-weight: 600; }
.yc-tk-row time { flex: none; font-size: 12.5px; color: var(--yc-dim); }
.yc-tk-row i { flex: none; padding: 3px 10px; border-radius: 999px; font-style: normal; font-size: 12px; font-weight: 600; background: var(--yc-fill); color: var(--yc-mut); }
.yc-tk-row i.is-answered { background: var(--yc-lime); color: #0b0f0c; }
.yc-tk-row i.is-open { box-shadow: inset 0 0 0 1px var(--yc-ink); color: var(--yc-ink); background: none; }
.yc-tk-thread { display: grid; gap: 10px; padding: 2px 0 20px; }
.yc-tk-msg { justify-self: start; max-width: 86%; padding: 12px 16px; border-radius: 16px; background: var(--yc-fill); font-size: 14.5px; line-height: 1.7; white-space: pre-wrap; overflow-wrap: anywhere; }
.yc-tk-msg small { display: block; margin-bottom: 2px; font-size: 12px; color: var(--yc-dim); }
.yc-tk-msg.from-user { justify-self: end; background: var(--yc-ink); color: var(--yc-paper); }
.yc-tk-msg.from-user small { color: inherit; opacity: 0.6; }
.yc-tk-reply { margin-top: 6px; }
.yc-tk-more { margin-top: 14px; }
@media (max-width: 760px) {
  .yc-tk { padding: 22px 20px; }
  .yc-tk-head { flex-direction: column; }
  .yc-tk-msg { max-width: 100%; }
}
</style>
