/** The maker a model comes from, read from its name; it is what users pick a model by. */
export interface ModelFamily {
  key: string
  label: string
  rank: number
}

const FAMILIES: [key: string, label: string, match: RegExp][] = [
  ['claude', 'Claude', /claude/],
  ['gpt', 'GPT', /^(gpt|o\d|chatgpt|codex)|dall-e/],
  ['gemini', 'Gemini', /gemini|gemma|imagen|veo-/],
  ['grok', 'Grok', /grok/],
  ['deepseek', 'DeepSeek', /deepseek/],
  ['glm', 'GLM', /glm|cogview|cogvideo/],
  ['qwen', 'Qwen', /qwen|qwq/],
  ['kimi', 'Kimi', /kimi|moonshot/],
  ['doubao', 'Doubao', /doubao/],
  ['minimax', 'MiniMax', /minimax/],
  ['mistral', 'Mistral', /mistral|mixtral|codestral|devstral|pixtral|magistral|voxtral/],
  ['llama', 'Llama', /llama/]
]

/** Null for models whose maker is not recognised. */
export function modelFamily(name: string): ModelFamily | null {
  const lower = name.toLowerCase()
  const rank = FAMILIES.findIndex(([, , match]) => match.test(lower))
  if (rank === -1) return null
  const [key, label] = FAMILIES[rank]
  return { key, label, rank }
}

/** Families in the given model names, each with one model to draw its icon from, in display order. */
export function familiesOf(names: string[]): (ModelFamily & { sample: string })[] {
  const found = new Map<string, ModelFamily & { sample: string }>()
  for (const name of names) {
    const f = modelFamily(name)
    if (f && !found.has(f.key)) found.set(f.key, { ...f, sample: name })
  }
  return [...found.values()].sort((a, b) => a.rank - b.rank)
}
