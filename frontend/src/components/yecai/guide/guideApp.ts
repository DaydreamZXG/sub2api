// The connect-and-errors guide shown under Help. Content comes from src/data/yc-guide.json, which
// tools/yc-guide/build.py generates from the Markdown source, so the HTML it carries is build-time trusted.
// Everything a visitor types goes through esc() or textContent.

export interface GuideError { s: string; c: string; f: string; t: string }
export interface GuideTab { kind: string; label: string; html: string }
export interface GuideSection { id: string; title: string; html: string }
export interface GuideTool {
  num: number; name: string; note: string; status: string; cat: string; new: boolean; updated: boolean
  lead: string; leadHtml: string; tabs: GuideTab[]; errors: GuideError[]; refs: string
}
export interface GuideData {
  title: string; meta: string[]; intro: string; start: GuideSection[]; cats: GuideSection[]; tools: GuideTool[]
  general: { title: string; rest: string; errors: GuideError[] }; pages: GuideSection[]
}

const ESCAPES: Record<string, string> = { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;' }
const esc = (s: unknown) => String(s).replace(/[&<>"]/g, (c) => ESCAPES[c])
const STATUS: Record<string, string> = { direct: '直接支持', convert: '协议转换', limited: '有限支持', no: '不建议' }
const SEARCH_ICON = '<svg viewBox="0 0 24 24"><circle cx="11" cy="11" r="7"/><path d="m20 20-3.6-3.6" stroke-linecap="round"/></svg>'
const svg = (d: string) => `<svg viewBox="0 0 24 24" width="22" height="22" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">${d}</svg>`
const CAT_ICON = [
  svg('<path d="M4 5h16v11H9l-5 4z"/>'),
  svg('<path d="M9 8l-4 4 4 4M15 8l4 4-4 4"/>'),
  svg('<rect x="3" y="4" width="18" height="16" rx="3"/><path d="M7 10l3 2-3 2M13 15h4"/>'),
  svg('<path d="M4 8h13l-3-3M20 16H7l3 3"/>'),
  svg('<circle cx="6" cy="6" r="2.5"/><circle cx="18" cy="12" r="2.5"/><circle cx="6" cy="18" r="2.5"/><path d="M8.5 6.8l7 4M8.5 17.2l7-4"/>')
]
const HERO_ART = `<svg class="hero-art" viewBox="0 0 240 240" aria-hidden="true"><rect x="20" y="20" width="200" height="200" rx="48" fill="#0b0f0c"/><circle cx="120" cy="120" r="74" fill="#c8ff00" opacity=".14"/>
  <path d="M120 186V104" stroke="#c8ff00" stroke-width="9" stroke-linecap="round"/><path d="M120 128c-34 2-52-18-54-46 30-2 52 14 54 46z" fill="#c8ff00"/><path d="M120 110c2-32 22-50 52-50 0 30-18 50-52 50z" fill="#c8ff00"/>
  <path d="M84 190h72" stroke="#f4f6ef" stroke-width="9" stroke-linecap="round" opacity=".9"/><circle cx="182" cy="170" r="5" fill="#c8ff00"/><circle cx="58" cy="150" r="4" fill="#f4f6ef" opacity=".6"/></svg>`

// Codex spells its placeholder out: 请替换为支持Responses的模型ID.
const MODEL_PLACEHOLDER = /请替换为(?:支持Responses的)?模型ID/g

type FilterKey = 'cat' | 'status'

/** Renders the guide into `root` and follows the URL hash (#/tools, #/tool/5/errors, #/errors, #/p/id). Returns the cleanup. */
export function mountGuide(root: HTMLElement, D: GuideData): () => void {
  root.innerHTML = '<nav class="gnav" data-g="nav"></nav><div data-g="main"></div><div class="gtoast" data-g="toast"></div>'
  const part = <T extends HTMLElement = HTMLElement>(name: string) => root.querySelector<T>(`[data-g="${name}"]`)
  const state = { cat: '', status: '', line: 'cn', key: '', model: '', tab: {} as Record<number, string> }
  const toolByNum = new Map(D.tools.map((t) => [t.num, t]))
  const catTitle: Record<string, string> = Object.fromEntries(D.cats.map((c) => [c.id, c.title]))
  const errorTotal = D.general.errors.length + D.tools.reduce((n, t) => n + t.errors.length, 0)

  const badge = (t: GuideTool) => `${t.new ? '<span class="gbadge new">新增</span>' : ''}${t.updated ? '<span class="gbadge new">已更新</span>' : ''}${t.status ? `<span class="gbadge ${t.status}">${STATUS[t.status] ?? ''}</span>` : ''}`
  const hue = (s: string) => [...s].reduce((h, c) => (h * 31 + (c.codePointAt(0) ?? 0)) % 360, 7)
  const avatar = (t: GuideTool) => `<i class="ava p${hue(t.name) % 5}">${esc((t.name.match(/[A-Za-z0-9\u4e00-\u9fff]/) || ['·'])[0].toUpperCase())}</i>`
  const card = (t: GuideTool) => `<a class="gcard" href="#/tool/${t.num}">
    <div class="card-top">${avatar(t)}<b>${esc(t.name)}</b></div><div class="card-badges">${badge(t)}</div>
    ${t.note ? `<em>${esc(t.note)}</em>` : ''}<p>${esc(t.lead)}</p>
    <footer>${t.errors.length ? `${t.errors.length} 条常见报错` : esc(catTitle[t.cat] ?? '')}</footer></a>`
  const errCard = (e: GuideError, tool: GuideTool | null) => `<div class="err">${tool ? `<a class="err-from" href="#/tool/${tool.num}/errors">${esc(tool.name)}</a>` : '<span class="err-from">通用</span>'}
    <div class="err-s">${e.s}</div><dl><dt>原因</dt><dd>${e.c}</dd><dt>怎么办</dt><dd>${e.f}</dd></dl></div>`
  const searchBox = (placeholder: string) => `<div class="search">${SEARCH_ICON}<input data-g="q" type="search" autocomplete="off" placeholder="${placeholder}"></div>`
  const lineSwitch = () => `<span class="seg"><button type="button" data-line="cn" class="${state.line === 'cn' ? 'on' : ''}">大陆</button><button type="button" data-line="global" class="${state.line === 'global' ? 'on' : ''}">全球</button></span>`

  // A query that only names a tool lists that tool's errors on the error page; on the home page its card is enough.
  function find(q: string, onlyErrors: boolean) {
    const words = q.toLowerCase().split(/\s+/).filter(Boolean)
    const errors: Array<[GuideError, GuideTool | null]> = []
    if (!words.length) return { tools: [] as GuideTool[], errors }
    const hit = (text: string) => words.every((w) => text.includes(w))
    const tools = D.tools.filter((t) => hit(`${t.name} ${t.note} ${t.lead} ${catTitle[t.cat] ?? ''}`.toLowerCase()))
    for (const e of D.general.errors) if (hit(e.t.toLowerCase())) errors.push([e, null])
    for (const t of D.tools) {
      for (const e of t.errors) {
        const text = e.t.toLowerCase()
        if (hit(text) || (hit(`${t.name.toLowerCase()} ${text}`) && (onlyErrors || words.some((w) => text.includes(w))))) errors.push([e, t])
      }
    }
    return { tools, errors }
  }
  function results(q: string, onlyErrors: boolean) {
    const { tools, errors } = find(q, onlyErrors)
    let html = ''
    if (!onlyErrors && tools.length) html += `<h2 class="sec">工具<small>${tools.length} 个</small></h2><div class="grid">${tools.map(card).join('')}</div>`
    if (errors.length) html += `<h2 class="sec">报错<small>${errors.length} 条</small></h2><div class="errs prose wide">${errors.slice(0, 60).map(([e, t]) => errCard(e, t)).join('')}</div>${errors.length > 60 ? '<p class="empty">结果太多，只显示前 60 条，再加一个词试试。</p>' : ''}`
    return html || `<p class="empty">没找到“${esc(q)}”。换个说法试试，比如只输入状态码或工具名。</p>`
  }

  const views = {
    home(): string {
      const cats = D.cats.map((c, i) => {
        const mine = D.tools.filter((t) => t.cat === c.id)
        return `<a class="cat" href="#/tools" data-cat="${c.id}"><i class="cat-ic">${CAT_ICON[i % CAT_ICON.length]}</i><b>${esc(c.title)}</b><span>${mine.length} 个工具</span><em>${esc(mine.slice(0, 4).map((t) => t.name.split(/[、（(/]/)[0].trim()).join(' · '))}${mine.length > 4 ? ' …' : ''}</em></a>`
      }).join('')
      const fold = (title: string, html: string) => `<details class="fold"><summary>${esc(title)}</summary><div class="prose">${html}</div></details>`
      const addr = [['OpenAI 格式（大多数软件）', 'https://yeschoy.com/v1'], ['Anthropic / Gemini 原生格式', 'https://yeschoy.com'], ['软件要“完整接口地址”时', 'https://yeschoy.com/v1/chat/completions']]
      const steps = [['拿密钥', '在YesChoy 后台创建一枚 API 密钥'], ['填地址', '把下面的地址填进你的工具'], ['选模型', '从模型定价页复制完整模型 ID']]
      return `<section class="hero"><div class="hero-text"><span class="hero-tag">${D.tools.length} 个工具 · ${errorTotal} 条报错解法</span>
        <h1>${esc(D.title)}</h1><div class="lead"><p>选你的工具，只看它的配置和报错。</p></div>
        ${searchBox('搜工具名或报错，比如 Cursor、401、模型列表为空')}</div>${HERO_ART}</section>
        <ol class="steps3">${steps.map(([a, b], i) => `<li><i>${i + 1}</i><b>${a}</b><span>${b}</span></li>`).join('')}</ol>
        <div data-g="out"><section class="quick" data-sub>
          <header><b>地址速查</b>${lineSwitch()}</header>
          ${addr.map(([label, url]) => `<div class="quick-row"><span>${label}</span><code>${url}</code></div>`).join('')}
          <p>点地址就能复制。密钥在YesChoy 后台创建，模型 ID 从“模型定价”页完整复制。</p></section>
        <h2 class="sec">按类别找工具</h2><div class="cats">${cats}</div>
        <h2 class="sec">新手先看</h2>${D.start.map((s) => fold(s.title, s.html)).join('')}
        ${fold('关于这份指南', `${D.intro}<ul>${D.meta.map((m) => `<li>${esc(m)}</li>`).join('')}</ul>`)}</div>`
    },
    tools(): string {
      const matches = (t: GuideTool) => !state.status || (state.status === 'new' ? t.new : state.status === 'updated' ? t.updated : t.status === state.status)
      const list = D.tools.filter((t) => (!state.cat || t.cat === state.cat) && matches(t))
      const chip = (key: FilterKey, val: string, label: string, n?: number) => `<button type="button" class="chip${state[key] === val ? ' on' : ''}" data-set="${key}" data-val="${val}">${esc(label)}${n === undefined ? '' : `<small>${n}</small>`}</button>`
      const count = (test: (t: GuideTool) => boolean) => D.tools.filter(test).length
      const groups = D.cats.filter((c) => (!state.cat || c.id === state.cat) && list.some((t) => t.cat === c.id))
        .map((c) => `<h2 class="sec">${esc(c.title)}</h2>${c.html ? `<div class="prose">${c.html}</div>` : ''}<div class="grid mt">${list.filter((t) => t.cat === c.id).map(card).join('')}</div>`).join('')
      return `<h1>找到你的工具</h1><p class="lead">点开一个工具，只看它自己的配置步骤、注意事项和常见报错。</p>
        <div class="chips">${chip('cat', '', '全部', D.tools.length)}${D.cats.map((c) => chip('cat', c.id, c.title, count((t) => t.cat === c.id))).join('')}</div>
        <div class="chips">${chip('status', '', '不限支持程度')}${Object.entries(STATUS).map(([k, v]) => chip('status', k, v, count((t) => t.status === k))).join('')}${D.tools.some((t) => t.new) ? chip('status', 'new', '新增', count((t) => t.new)) : ''}${D.tools.some((t) => t.updated) ? chip('status', 'updated', '已更新', count((t) => t.updated)) : ''}</div>
        ${groups || '<p class="empty">这个组合下没有工具。</p>'}`
    },
    tool(num: number, tab?: string): string {
      const t = toolByNum.get(num)
      if (!t) return views.tools()
      const active = t.tabs.find((x) => x.kind === (tab || state.tab[num])) || t.tabs[0]
      state.tab[num] = active.kind
      const i = D.tools.indexOf(t)
      const prev = D.tools[i - 1]
      const next = D.tools[i + 1]
      return `<a class="back" href="#/tools">← 全部工具</a>
        <div class="title-row">${avatar(t)}<h1>${esc(t.name)}</h1>${badge(t)}</div>
        <div class="lead">${t.note ? `<p>${esc(t.note)} · ${esc(catTitle[t.cat] ?? '')}</p>` : ''}${t.leadHtml}</div>
        <div class="fill">
          <label>线路${lineSwitch()}</label>
          <label>我的密钥<input data-g="key" placeholder="sk-...（可不填）" value="${esc(state.key)}" autocomplete="off" spellcheck="false"></label>
          <label>模型 ID<input data-g="model" placeholder="从模型定价页复制（可不填）" value="${esc(state.model)}" autocomplete="off" spellcheck="false"></label>
          <p>填了之后，下面的地址和命令会自动换成你的，复制就能用。只留在这个页面里，刷新就清空。</p>
        </div>
        ${t.tabs.length > 1 ? `<div class="gtabs">${t.tabs.map((x) => `<button type="button" data-tab="${x.kind}" class="${x === active ? 'on' : ''}">${x.label}${x.kind === 'errors' ? `<small>${t.errors.length}</small>` : ''}</button>`).join('')}</div>` : ''}
        <div class="prose" data-sub>${active.html}</div>
        ${t.refs ? `<div class="refs prose">${t.refs}</div>` : ''}
        <div class="pager">${prev ? `<a href="#/tool/${prev.num}">← ${esc(prev.name)}</a>` : '<span></span>'}${next ? `<a href="#/tool/${next.num}">${esc(next.name)} →</a>` : ''}</div>`
    },
    errors(): string {
      const quick = ['401', '404', '429', '403', '超时', '模型列表', '余额', 'v1/v1', '工具调用']
      return `<h1>${esc(D.general.title)}</h1><p class="lead">输入状态码或报错里的几个字，所有工具的报错表一起搜。</p>
        ${searchBox('比如 401、404、context_length_exceeded、一直转圈')}
        <div class="chips">${quick.map((w) => `<button type="button" class="chip" data-q="${w}">${w}</button>`).join('')}</div>
        <div data-g="out"><h2 class="sec">最常见的<small>${D.general.errors.length} 条，适用于所有工具</small></h2>
        <div class="errs">${D.general.errors.map((e) => errCard(e, null)).join('')}</div>
        <div class="prose" data-sub>${D.general.rest}</div></div>`
    },
    p(id: string): string {
      const page = D.pages.find((x) => x.id === id)
      return page ? `<h1>${esc(page.title)}</h1><div class="prose" data-sub>${page.html}</div>` : views.home()
    }
  }

  // Swaps the mainland address for the global one and fills in the visitor's key and model ID, in place.
  function substitute() {
    root.querySelectorAll<HTMLElement>('[data-sub] code').forEach((el) => {
      if (el.dataset.o === undefined) el.dataset.o = el.textContent ?? ''
      let text = el.dataset.o
      if (state.line === 'global' && !text.includes('api.yeschoy.com')) text = text.replace(/(^|[^\w.-])yeschoy\.com/g, '$1api.yeschoy.com')
      if (state.key) text = text.split('请替换为你的密钥').join(state.key)
      if (state.model) text = text.replace(MODEL_PLACEHOLDER, () => state.model)
      if (el.textContent !== text) el.textContent = text
    })
  }
  function enhance(scope: HTMLElement) {
    scope.querySelectorAll<HTMLElement>('pre').forEach((pre) => {
      const lang = pre.querySelector('code')?.className.match(/language-(\w+)/)?.[1]
      if (lang && lang !== 'Plaintext') pre.dataset.lang = lang
      if (pre.querySelector('.copy')) return
      const button = document.createElement('button')
      button.type = 'button'
      button.className = 'copy'
      button.textContent = '复制'
      pre.appendChild(button)
    })
    substitute()
  }

  let toastTimer: ReturnType<typeof setTimeout> | undefined
  async function copy(text: string) {
    const toast = part('toast')
    try {
      await navigator.clipboard.writeText(text)
      if (toast) toast.textContent = '已复制'
    } catch {
      if (toast) toast.textContent = '复制失败，请手动选中复制'
    }
    toast?.classList.add('on')
    clearTimeout(toastTimer)
    toastTimer = setTimeout(() => toast?.classList.remove('on'), 1400)
  }

  function route() {
    const [, name = '', a = '', b] = location.hash.split('/')
    const main = part('main')
    const nav = part('nav')
    if (!main || !nav) return
    const view = name === 'tools' || name === 'tool' || name === 'errors' || name === 'p' ? name : 'home'
    main.innerHTML = view === 'tool' ? views.tool(Number(a), b) : view === 'p' ? views.p(a) : views[view]()
    enhance(main)
    const links: Array<[string, string, number]> = [['#/', '开始', 0], ['#/tools', '找工具', D.tools.length], ['#/errors', '查报错', errorTotal], ...D.pages.map((p): [string, string, number] => [`#/p/${p.id}`, p.title, 0])]
    const current = view === 'tool' ? '#/tools' : view === 'p' ? `#/p/${a}` : view === 'home' ? '#/' : `#/${view}`
    nav.innerHTML = links.map(([href, label, n]) => `<a href="${href}" class="${href === current ? 'on' : ''}">${esc(label)}${n ? `<small>${n}</small>` : ''}</a>`).join('')
  }
  function go(hash: string) {
    if (location.hash === hash) route()
    else location.hash = hash
  }

  function onClick(ev: MouseEvent) {
    const target = ev.target as Element
    const el = target.closest<HTMLElement>('[data-set],[data-tab],[data-line],[data-q],[data-cat],.copy,.prose code,.quick-row code')
    if (el?.dataset.set) {
      state[el.dataset.set as FilterKey] = el.dataset.val ?? ''
      route()
    } else if (el?.dataset.tab) {
      const num = Number(location.hash.split('/')[2])
      state.tab[num] = el.dataset.tab
      history.replaceState(history.state, '', `#/tool/${num}/${el.dataset.tab}`)
      route()
    } else if (el?.dataset.line) {
      state.line = el.dataset.line
      root.querySelectorAll('[data-line]').forEach((b) => b.classList.toggle('on', b === el))
      substitute()
    } else if (el?.dataset.q) {
      const q = part<HTMLInputElement>('q')
      if (q) {
        q.value = el.dataset.q
        q.dispatchEvent(new Event('input', { bubbles: true }))
      }
    } else if (el?.dataset.cat) {
      state.cat = el.dataset.cat
      state.status = ''
    } else if (el?.classList.contains('copy')) {
      void copy(el.parentElement?.querySelector('code')?.textContent ?? '')
    } else if (el && !el.closest('pre')) {
      void copy(el.textContent ?? '')
    }
    const link = target.closest<HTMLAnchorElement>('a[href^="#/"]')
    if (link) {
      ev.preventDefault()
      go(link.getAttribute('href') ?? '#/')
    }
  }
  function onInput(ev: Event) {
    const el = ev.target as HTMLInputElement
    const which = el.dataset.g
    if (which === 'key' || which === 'model') {
      state[which] = el.value.trim()
      substitute()
      return
    }
    const out = part('out')
    if (which !== 'q' || !out) return
    if (out.dataset.o === undefined) out.dataset.o = out.innerHTML
    const q = el.value.trim()
    out.innerHTML = q ? results(q, location.hash.startsWith('#/errors')) : out.dataset.o
    enhance(out)
  }
  function onHashChange() {
    route()
    window.scrollTo(0, 0)
  }

  root.addEventListener('click', onClick)
  root.addEventListener('input', onInput)
  window.addEventListener('hashchange', onHashChange)
  route()
  return () => {
    clearTimeout(toastTimer)
    root.removeEventListener('click', onClick)
    root.removeEventListener('input', onInput)
    window.removeEventListener('hashchange', onHashChange)
  }
}
