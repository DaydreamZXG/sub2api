export default {
  nav: { aria: 'Page navigation', docs: 'Docs' },
  cta: { start: 'Get started', login: 'Sign in', console: 'Open console' },
  flow: { aria: 'Your tools reach every model through one key', key: 'Your key' },
  hero: {
    eyebrow: 'AI model access',
    titleA: 'Connect to AI',
    titleB: 'without the setup',
    lead: 'Install the desktop app and connect Claude or Codex in one click. Prices are in yuan, every request is on record, and unused balance is refundable.'
  },
  pain: {
    before: 'Before',
    after: 'Now',
    setup: {
      tag: 'Setup',
      titleA: 'Before, you edited config files.',
      titleB: 'Now, one click in the app.',
      b1: 'Search the docs for the endpoint',
      b2: 'Open a config file and fill it in line by line',
      b3: 'Repeat for every tool',
      a1: 'Download the app and sign in',
      a2: 'It finds Claude and Codex on your computer',
      a3: 'Select Connect, and you are done'
    },
    price: {
      tag: 'Prices',
      titleA: 'Before, you worked out multipliers.',
      titleB: 'Now, prices are in yuan.',
      formula: 'Top-up rate × group multiplier × model multiplier × official price = ?',
      b1: 'The same multiplier can mean very different prices',
      b2: 'Pick a group first to learn the price',
      b3: 'Charges are in credits, not money',
      a1: 'One yuan price per model',
      a2: 'What you pay is what you get: ¥1 is ¥1',
      a3: 'Multipliers and raw prices are one click away'
    },
    bill: {
      tag: 'Bills',
      titleA: 'Before, you saw one total.',
      titleB: 'Now, you see every request.',
      b1: 'Balance drops and you cannot tell why',
      b2: 'No way to check which model was used',
      b3: 'Money paid in never comes back',
      a1: 'Each request shows its model and charge',
      a2: 'Input, output and cache listed separately',
      a3: 'Unused balance is refundable'
    },
    help: {
      tag: 'Support',
      titleA: 'Before, you debugged errors alone.',
      titleB: 'Now, someone answers.',
      b1: 'An error message you cannot read',
      b2: 'No idea if the fault is yours or the service',
      b3: 'Nobody to ask',
      a1: 'Ask in the QQ group',
      a2: 'Open a ticket and get a follow-up',
      a3: 'Connection, cache and network issues are all covered'
    }
  },
  live: { label: 'Tokens processed across the site · live', today: 'Today', requests: 'Requests in total' },
  client: {
    download: 'Download the app',
    windows: 'Download for Windows',
    macos: 'Download for macOS',
    manual: 'For Windows and macOS. Prefer not to install it? Sign in and copy the settings from the web page.',
    window: 'Desktop app · My apps',
    shotAlt: 'My apps in the desktop app: Claude and Codex found on this computer, ready to connect in one click'
  },
  price: {
    eyebrow: 'Price list',
    title: 'And now, the price.',
    official: 'Official',
    perMillionIn: 'per million input tokens',
    off: '{pct}% below official',
    note: 'See the full list for other models. Official prices are converted at about ¥{rate} to $1, for comparison only.',
    output: 'Output: ¥{price} per million tokens',
    pick: 'Select a model to see its price',
    all: 'All models and prices'
  },
  faq: {
    eyebrow: 'FAQ',
    title: 'Before you start',
    q1: 'Do I need to know how to code?',
    a1: 'No. Install the app and connect in one click, or copy and paste once as the page shows.',
    q2: 'What is a key?',
    a2: 'A password that starts with sk-. Your tool uses it to spend your balance. It is yours alone, so do not share it.',
    q3: 'What is a token?',
    a3: 'The unit models use to measure usage, roughly word count: an English word is about 1 to 2 tokens. Prices are quoted per million tokens.',
    q4: 'How am I charged?',
    a4: 'Each request deducts its actual usage from your balance. Prices differ by model, and the bill shows what you spent.',
    q5: 'Which tools are supported?',
    a5: 'Claude Code, Codex, Gemini CLI, OpenCode and other common tools.'
  },
  shell: { start: 'Home', keys: 'Keys', bills: 'Bills', topUp: 'Top up', admin: 'Admin', account: 'My account' },
  dash: {
    night: 'Up late', morning: 'Good morning', afternoon: 'Good afternoon', evening: 'Good evening', comma: ', ',
    nextTag: 'Next', nextTopUp: 'Add a little balance. You pay for what you use, and the rest is refundable.',
    nextKey: 'Create your first key. Your tool uses it to spend your balance.', nextKeyGo: 'Create',
    nextConnect: 'You have a key. Connect it to your tool to start.', nextConnectGo: 'Connect',
    balance: 'Balance', balanceHint: 'Pay for what you use; the rest is refundable', today: 'Spent today', total: 'Spent in total',
    requests: '{n} requests', recent: 'Recent usage', allBills: 'All bills',
    empty: 'No usage yet. After you connect a tool and send a request, each one is listed here.',
    expandTip: 'Select a row to see every detail of that request.',
    dModel: 'Model', dInput: 'Input', dOutput: 'Output', dCacheWrite: 'Cache write', dCacheRead: 'Cache read',
    dTiming: 'First token / total', dListCost: 'List price', dMultiplier: 'Multiplier'
  },
  final: { title: 'Connect it now' }
}
