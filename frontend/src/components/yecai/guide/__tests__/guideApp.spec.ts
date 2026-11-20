import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { mountGuide, type GuideData } from '../guideApp'

const data: GuideData = {
  title: '接入指南',
  meta: ['更新于今天'],
  intro: '<p>简介</p>',
  start: [{ id: 's1', title: '一分钟版本', html: '<p>三步</p>' }],
  cats: [{ id: 'cli', title: '终端 Agent', html: '' }],
  tools: [
    {
      num: 1, name: 'Claude Code', note: '', status: 'convert', cat: 'cli', new: false, updated: true,
      lead: '用 Anthropic 协议', leadHtml: '<p>用 Anthropic 协议</p>',
      tabs: [
        { kind: 'steps', label: '怎么配置', html: '<pre><code class="language-Bash">export ANTHROPIC_BASE_URL="https://yeschoy.com"\nexport ANTHROPIC_AUTH_TOKEN="请替换为你的密钥"</code></pre>' },
        { kind: 'errors', label: '常见报错', html: '<p>报错表</p>' }
      ],
      errors: [{ s: '401 Unauthorized', c: '密钥不对', f: '重新复制密钥', t: '401 unauthorized 密钥不对 重新复制密钥' }],
      refs: ''
    },
    {
      num: 2, name: 'Amp', note: '', status: 'direct', cat: 'cli', new: true, updated: false,
      lead: '编程 Agent', leadHtml: '<p>编程 Agent</p>', errors: [], refs: '',
      tabs: [{ kind: 'steps', label: '怎么配置', html: '<pre><code>model = "请替换为支持Responses的模型ID"\nalias = "请替换为模型ID"</code></pre>' }]
    }
  ],
  general: { title: '通用报错', rest: '', errors: [{ s: '404 Not Found', c: '地址错了', f: '检查 /v1', t: '404 not found 地址错了 检查 /v1' }] },
  pages: [{ id: 'safe', title: '安全提醒', html: '<p>别泄露密钥</p>' }]
}

describe('mountGuide', () => {
  let root: HTMLElement
  let unmount: () => void

  beforeEach(() => {
    window.scrollTo = vi.fn()
    location.hash = ''
    root = document.createElement('div')
    document.body.appendChild(root)
    unmount = mountGuide(root, data)
  })
  afterEach(() => {
    unmount()
    root.remove()
    location.hash = ''
  })

  const open = (hash: string) => {
    location.hash = hash
    window.dispatchEvent(new HashChangeEvent('hashchange'))
  }
  const type = (name: string, value: string) => {
    const input = root.querySelector<HTMLInputElement>(`[data-g="${name}"]`)!
    input.value = value
    input.dispatchEvent(new Event('input', { bubbles: true }))
  }

  it('starts on the home page with the tool and error counts', () => {
    expect(root.querySelector('.hero-tag')?.textContent).toBe('2 个工具 · 2 条报错解法')
    expect(root.querySelector('[data-g="nav"] a.on')?.textContent).toBe('开始')
  })

  it('filters the tool list by the "updated" chip', () => {
    open('#/tools')
    expect(root.querySelectorAll('.gcard')).toHaveLength(2)
    root.querySelector<HTMLElement>('[data-set="status"][data-val="updated"]')!.click()
    expect([...root.querySelectorAll('.gcard b')].map((b) => b.textContent)).toEqual(['Claude Code'])
  })

  it('puts the global address and the visitor key into the commands of a tool page', () => {
    open('#/tool/1')
    const code = () => root.querySelector('pre code')!.textContent
    expect(code()).toContain('https://yeschoy.com"')
    root.querySelector<HTMLElement>('[data-line="global"]')!.click()
    type('key', 'sk-demo')
    expect(code()).toContain('https://api.yeschoy.com"')
    expect(code()).toContain('ANTHROPIC_AUTH_TOKEN="sk-demo"')
    expect(root.querySelector('pre')?.dataset.lang).toBe('Bash')
  })

  it('fills the model ID into both spellings of the placeholder', () => {
    open('#/tool/2')
    type('model', 'gpt-x')
    expect(root.querySelector('pre code')!.textContent).toBe('model = "gpt-x"\nalias = "gpt-x"')
  })

  it('switches tabs on a tool page and keeps the tab in the URL', () => {
    open('#/tool/1')
    root.querySelector<HTMLElement>('[data-tab="errors"]')!.click()
    expect(location.hash).toBe('#/tool/1/errors')
    expect(root.querySelector('.prose[data-sub]')?.textContent).toBe('报错表')
  })

  it('searches the errors of every tool and names the tool an error belongs to', () => {
    open('#/errors')
    type('q', '401')
    expect([...root.querySelectorAll('.err-s')].map((e) => e.textContent)).toEqual(['401 Unauthorized'])
    expect(root.querySelector('a.err-from')?.getAttribute('href')).toBe('#/tool/1/errors')
    type('q', '<b>nothing</b>')
    expect(root.querySelector('.empty')?.textContent).toContain('<b>nothing</b>')
    expect(root.querySelector('.empty b')).toBeNull()
  })

  it('lists a tool\'s errors when the error search only names the tool', () => {
    open('#/errors')
    type('q', 'claude')
    expect([...root.querySelectorAll('.err-s')].map((e) => e.textContent)).toEqual(['401 Unauthorized'])
  })

  it('follows in-page links without leaving the route', () => {
    root.querySelector<HTMLAnchorElement>('a.cat')!.click()
    expect(location.hash).toBe('#/tools')
  })

  it('stops reacting to the URL after it is unmounted', () => {
    unmount()
    open('#/tools')
    expect(root.querySelector('.hero-tag')).not.toBeNull()
  })
})
