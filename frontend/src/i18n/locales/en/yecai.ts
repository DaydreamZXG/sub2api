export default {
  nav: { aria: 'Page navigation', pains: 'What we fix', prices: 'Prices', faq: 'FAQ', docs: 'Docs', client: 'Client' },
  cta: { start: 'Get started', login: 'Sign in', console: 'Open console' },
  flow: { aria: 'Your tools reach every model through one key', key: 'Your key' },
  hero: {
    eyebrow: 'AI model access',
    titleA: 'Connect to AI',
    titleB: 'without the setup',
    lead: 'Wire AI models into Claude Code or Codex. Install the YesChoy client and pick a classmate for one click, or paste in the endpoint yourself. Prices are in yuan, every request is on record, and unused balance is refundable.'
  },
  pain: {
    before: 'Before',
    after: 'Now',
    setup: {
      tag: 'Setup',
      titleA: 'Before, you edited config files.',
      titleB: 'Now, pick an AI classmate.',
      b1: 'Search the docs for the endpoint',
      b2: 'Open a config file and fill it in line by line',
      b3: 'Repeat for every tool',
      a1: 'Download YesChoy and sign in',
      a2: 'Pick an AI classmate: Juzi, Aqing, Xiaojing…',
      a3: 'She connects Claude Code and Codex for you'
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
  live: { label: 'Tokens processed across the site, live', today: '{tokens} today · {requests} requests in total' },
  client: {
    download: 'Download YesChoy',
    windows: 'Download for Windows',
    macos: 'Download for macOS',
    manual: 'For Windows and macOS. Prefer not to install it? Sign in and copy the settings from the web page.',
    window: 'YesChoy · nine AI classmates',
    shotAlt: 'AI classmates in YesChoy: Juzi, Aqing, Xiaojing, Yueyue and Doubao'
  },
  price: {
    eyebrow: 'Prices',
    title: 'And now, the price.',
    official: 'Official',
    perMillionIn: 'per million input tokens',
    off: '{pct}% below official',
    note: 'See the full list for other models. Official prices are in USD, for comparison only; discounts use about ¥{rate} to $1.',
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
  shell: {
    garden: 'Garden', start: 'Home', connect: 'Connect a tool', prices: 'Prices', keys: 'Manage keys', bills: 'Bills', topUp: 'Top up',
    help: 'Help', more: 'More', plans: 'Monthly plans', invite: 'Invite & earn', reseller: 'Reselling', admin: 'Admin', account: 'My account'
  },
  connect: {
    title: 'Connect a tool',
    lead: 'Two ways to do it. Pick either one.',
    best: 'Easiest',
    clientTitle: 'Install YesChoy, pick a classmate',
    clientText: 'YesChoy has nine AI classmates. Sign in, pick one, and she connects Claude Code and Codex on your computer, no config files.',
    appConnect: 'Connect to the app',
    appConnectHint: 'Opening YesChoy. If nothing happens, install it first.',
    appConnectNeedsHttps: 'This needs the site to be opened over https.',
    appCheaper: 'Some models cost less when you connect with the app.',
    appCheaperLink: 'See prices',
    appLinesOnly: 'Your lines only work through the app. Connect with YesChoy above.',
    appKeyNote: 'This key is on an app-only line: it gets the app price through YesChoy. Pasted into another tool it pays the normal price, or does not work.',
    appConnectFailed: 'Could not connect. Please try again later.',
    fillTitle: 'Or fill it into your tool yourself',
    fillLead: 'Most chat apps and coding plugins have an “API address” and a “key” field in their settings. Copy the values below into them; no terminal needed.',
    fillStep: 'Copy these into the tool’s settings',
    addrV1: 'If the tool wants the address to end in /v1',
    modelId: 'Model ID',
    modelIdHint: 'Copy the full model name from the price list',
    toPrices: 'Price list',
    toolGuide: 'Step-by-step guide for your tool',
    toolGuideHint: 'Not sure where these go? Pick your tool and follow the steps on screen.',
    pickKey: 'Pick a key',
    loading: 'Loading your keys…',
    newKey: 'Create a key',
    newKeyFor: 'New key: {line}',
    keyName: '{line} key',
    noLine: 'Your account has no line available yet. Please contact support.',
    created: 'Key created',
    createFailed: 'Could not create the key. Please try again later.',
    pickShell: 'Where do you run {tool}?',
    shell: { unix: 'macOS / Linux terminal', powershell: 'Windows PowerShell', cmd: 'Windows Command Prompt' },
    paste: 'Copy this, paste it into the terminal and press Enter',
    files: 'The tool for this key needs config files. Follow the steps once.',
    copy: 'Copy command',
    copyShort: 'Copy',
    copied: 'Copied',
    after: 'Then start {tool} in the same terminal window.',
    fullGuide: 'Want it saved in a config file, or use another tool? See the full setup',
    openGuide: 'Open the setup steps',
    cliTitle: 'I use the command line',
    baseUrl: 'Base URL',
    keyOf: 'Key ({name})',
    manageKeys: 'Manage all keys: limits, expiry, IP rules'
  },
  bills: {
    title: 'Bills',
    lead: 'Every request, the model it used and what it cost, is listed here.',
    range: 'Time range', today: 'Today', week: 'Last 7 days', month: 'Last 30 days',
    spent: 'Spent in this period', requests: 'Requests', tokens: 'Total tokens',
    empty: 'No usage in this period.',
    more: 'Show {n} more',
    advanced: 'Need filters, charts and export? Open the advanced view'
  },
  topUp: {
    title: 'Top up',
    lead: 'You get exactly what you pay: ¥1 is ¥1. You are charged only for what you use.',
    pay: 'Top up now', buyPlan: 'See monthly plans', redeem: 'I have a code', orders: 'See top-up history',
    minTitle: 'Start small', minAny: 'You choose the amount. Start small and see if it suits you.',
    bonusTitle: '{pct}% extra on larger top-ups', bonusText: 'Top up {min} or more in one payment and get {pct}% extra.',
    refundTitle: 'Unused balance is refundable', refundText: 'You can ask for a refund of balance you have not used. Contact support to arrange it.',
    clearTitle: 'Every charge is on record', clearText: 'What each request cost is listed in your bills, any time.'
  },
  help: {
    title: 'Help',
    lead: 'Stuck? Do not struggle alone, reach out to us.',
    contactTitle: 'Get help from a person', noContact: 'The site owner has not added contact details yet.', docs: 'Read the docs',
    newsTitle: 'Announcements'
  },
  guide: {
    title: 'Setup and error guide',
    lead: 'Pick your tool and see only its setup steps and common errors.',
    zhOnly: 'This guide is only available in Chinese for now.',
    loadFailed: 'The guide did not load. Refresh the page and try again.'
  },
  agentGuide: {
    title: 'Becoming an agent',
    lead: 'Bring customers in, and a part of what they spend is yours for as long as they stay.',
    card: 'Bring customers in and earn a share of what they spend. Rules and an example.',
    termsTitle: 'What you earn',
    terms: {
      rate: { name: 'Share', text: 'Agreed with each agent; what counts is the share set on your account. It can rise in tiers: the more your customers spend in a month, the higher the share on what comes after.' },
      basis: { name: 'Counted on', text: 'Money your customers topped up themselves and then spent. Spending gifted balance, or using a monthly plan, does not count.' },
      start: { name: 'From when', text: 'What a customer spends after they are under you. Nothing is added for earlier spending.' },
      duration: { name: 'For how long', text: 'For as long as the customer keeps using the service and you are an agent.' }
    },
    example: 'For example: at a share of 5%, a customer you brought in spends ¥10,000 of their own top-ups this month, and you earn ¥500. If they spend ¥10,000 again next month, you earn another ¥500. If "8% from ¥5,000 in the month" was agreed, the first ¥5,000 earns 5% and the next ¥5,000 earns 8%: ¥650 in all.',
    join: {
      title: 'How a customer comes under you',
      i1: 'Send the customer your invite link. If they sign up through it, they are under you from the start.',
      i2: 'If the customer already has an account, or did not use your link: tell us which account it is, and once confirmed we put them under you.'
    },
    rules: {
      title: 'Rules',
      i1: 'You cannot count yourself, or another account of yours, as a customer.',
      i2: 'A customer is under one agent at a time. If several people say they brought the customer in, the one we confirmed first counts.',
      i3: "As an agent you no longer get the top-up rebate for your customers; you get the share of their spending instead, not both.",
      i4: 'If a customer is refunded later, what was already credited to you stays.',
      i5: 'The share is credited about every 10 minutes, into the quota on the invite page.',
      i6: 'On the first of each month the tiers start again from the lowest one.'
    },
    payout: {
      title: 'How you are paid',
      i1: 'On the invite page, choose "Request a payout" and fill in the amount and how to pay you.',
      i2: 'We pay you outside the site and mark the request paid; the amount is then taken from your quota. One request can be open at a time.',
      i3: 'You can also move the quota to your own balance instead.'
    },
    apply: {
      title: 'Want to be an agent',
      text: 'Get in touch and tell us what kind of customers you can bring. Once confirmed, we set you up and agree on the share.',
      toHelp: 'Contact us on the Help page'
    }
  },
  tickets: {
    title: 'Tickets',
    lead: 'Tell us what went wrong and we will reply here.',
    new: 'New ticket',
    subject: 'The problem in one line',
    subjectHint: 'e.g. Claude Code says 401',
    category: 'Topic',
    cat: { usage: 'Setup and usage', billing: 'Top-ups and charges', account: 'Account', other: 'Other' },
    content: 'Details',
    contentHint: 'Which tool, what you did, and the error you saw. Paste the error text if you can.',
    submit: 'Submit',
    empty: 'No tickets yet.',
    loadFailed: 'Tickets cannot be loaded right now. Refresh in a moment.',
    failed: 'That did not go through. Please try again later.',
    status: { open: 'Waiting for reply', answered: 'Answered', closed: 'Closed' },
    author: { user: 'Me', staff: 'Support', ai: 'AI assistant' },
    replyHint: 'Add details or ask a follow-up',
    reopenHint: 'This ticket is closed. Sending a message reopens it.',
    send: 'Send',
    close: 'Solved, close the ticket',
    more: 'Show more'
  },
  growth: {
    title: 'My plant',
    off: '{percent}% off the app price',
    next: 'Top up {amount} more to reach "{name}": {percent}% off the app price',
    top: 'You are at the top level',
    note: 'Levels follow your total top-ups; refunded amounts do not count. The discount applies when you connect through the app.',
    manual: 'Your app price was set individually, so the level discount is not added on top.'
  },
  theme: { label: 'Appearance', light: 'Light', dark: 'Dark', system: 'Follow system' },
  garden: {
    title: 'Garden',
    lead: 'The bonus on a top-up is planted as a crop first. It ripens in {hours} hours; come back and harvest it into your balance. What you paid for arrives at once, as usual. Leave a ripe crop too long and neighbours take some.',
    failed: 'The garden cannot be opened right now. Try again later.',
    ripe: 'Ripe, ready to harvest',
    growing: 'Growing',
    ripeIn: 'Ripe in {time}',
    hoursMinutes: '{hours} h {minutes} min',
    minutes: '{minutes} min',
    harvest: 'Harvest',
    harvestDone: '{amount} harvested into your balance',
    harvestFailed: 'Could not harvest. Refresh and try again.',
    emptyTitle: 'Nothing planted yet',
    empty: 'When a top-up comes with a bonus, the bonus is planted here.',
    off: 'The garden is closed for now.',
    harvested: 'Harvested',
    stolen: '{amount} taken by neighbours',
    emptyPlot: 'Empty bed',
    source: { invite: 'Invite reward', gift: 'A gift from us' },
    water: 'Water',
    waterDone: 'Watered. This crop ripens {minutes} minutes sooner.',
    steal: 'Take some',
    stealDone: 'You took {amount}; it is in your balance',
    stealCaught: 'Wangcai caught you. You leave with nothing.',
    size: { small: 'Small crop', medium: 'Medium crop', large: 'Large crop' },
    state: {
      taken: 'You took from this one',
      caught: 'You were caught here',
      protected: 'Just ripe: the owner goes first',
      open: 'Not harvested yet',
      pickedClean: 'Nothing left to take'
    },
    dog: {
      name: 'Wangcai',
      fed: 'Fed today and on guard. It catches {percent}% of the people who come to take your crops; a higher level makes it fiercer.',
      hungry: 'Wangcai has not eaten today and is not guarding. Feed it so it catches people who come to take your crops.',
      feed: 'Feed Wangcai',
      feedDone: 'Wangcai is fed and on guard',
      onGuard: 'On guard'
    },
    neighbours: {
      title: "Today's neighbours",
      triesLeft: '{left} tries left today (out of {limit})',
      rules: 'Neighbours change every day; people you invited and the person who invited you always stay. Water a growing crop to ripen it {minutes} minutes sooner. A crop its owner has not harvested {protect} hours after ripening is open: each try takes {percent}%. You need a crop of your own to take from others.',
      none: 'No neighbour has a crop growing today.',
      pinned: 'Always here',
      raided: 'Took from you',
      crops: '{count} crops',
      loading: 'Looking over the fence…',
      noCrops: 'Their crops are all harvested.'
    },
    ranks: {
      title: 'This week',
      thieves: 'Took the most',
      dogs: 'Dogs that caught the most',
      none: 'Nobody yet this week',
      me: 'Me',
      caught: 'Caught {count}'
    },
    invite: {
      title: 'Invite a friend to the garden',
      lead: 'When a friend signs up with your link and pays for the first time, you get a {amount} crop. They also stay your neighbour.',
      copy: 'Copy invite link',
      copied: 'Invite link copied'
    },
    feed: {
      title: 'Activity',
      visit: 'Visit them',
      mine: {
        water: "You watered {name}'s crop",
        steal: "You took {amount} from {name}'s garden",
        caught: "Wangcai caught you in {name}'s garden"
      },
      theirs: {
        water: '{name} watered your crop',
        steal: '{name} took {amount} from you',
        caught: 'Wangcai caught {name} in your garden'
      }
    },
    errors: {
      GARDEN_CLOSED: 'The garden is closed for now.',
      GARDEN_NOT_NEIGHBOUR: "They are not one of today's neighbours.",
      GARDEN_OWN_CROP: 'This is your own crop.',
      GARDEN_CROP_PROTECTED: 'Only the owner can harvest this crop for now.',
      GARDEN_NEED_CROP: 'You need a crop of your own before taking from others.',
      GARDEN_STEAL_LIMIT: 'No tries left today. Come back tomorrow.',
      GARDEN_ALREADY_TRIED: 'You already tried this crop.',
      GARDEN_PICKED_CLEAN: 'There is nothing left to take from this crop.',
      GARDEN_CROP_RIPE: 'This crop is already ripe.',
      GARDEN_WATER_FULL: 'This crop has been watered enough.',
      GARDEN_ALREADY_WATERED: 'You already watered this crop today.',
      GARDEN_CROP_NOT_FOUND: 'This crop is gone.',
      other: 'That did not work. Refresh and try again.'
    }
  },
  priceWords: {
    cheaper: '{pct}% off',
    dearer: '{pct}% more',
    free: 'Free'
  },
  prices: {
    title: 'Prices',
    lead: 'One price per model, in yuan. Prices are per million tokens; you pay for what you use.',
    search: 'Search models, e.g. claude',
    model: 'Model', input: 'Input', output: 'Output',
    loading: 'Loading prices…',
    none: 'No prices to show yet.',
    noMatch: 'No model matches. Try another keyword.',
    from: '+', varies: 'This is the standard price. Requests with very long context, at certain hours or with higher reasoning effort are charged at a higher tier; your bill shows each one.',
    line: 'Line', cacheRead: 'Cache read', multiplier: 'Discount',
    official: 'Official price: ${input} input, ${output} output (per million tokens). Struck-out yuan prices use 1 USD ≈ {rate} CNY, for comparison only.',
    appPrice: '¥{price} with the app',
    appHeadline: 'App price',
    appOnly: 'App only',
    appNote: 'The "with the app" price applies when you connect through YesChoy. A key pasted into another tool pays the normal price.',
    foot: 'Click any row to see each line\'s price, cache price and discount.'
  },
  dash: {
    night: 'Up late', morning: 'Good morning', afternoon: 'Good afternoon', evening: 'Good evening', comma: ', ',
    nextTag: 'Next', nextTopUp: 'Add a little balance. You pay for what you use, and the rest is refundable.',
    nextKey: 'Create your first key. Your tool uses it to spend your balance.', nextKeyGo: 'Create',
    nextConnect: 'You have a key. Connect it to your tool to start.', nextConnectGo: 'Connect',
    balance: 'Balance', balanceHint: 'Pay for what you use; the rest is refundable', today: 'Spent today', total: 'Spent in total',
    mates: 'Your AI classmates', matesGo: 'Pick one in the app', mateJuzi: 'Juzi', mateAqing: 'Aqing', mateXiaojing: 'Xiaojing', mateYueyue: 'Yueyue', mateDoubao: 'Doubao',
    requests: '{n} requests', recent: 'Recent usage', allBills: 'All bills',
    empty: 'No usage yet. After you connect a tool and send a request, each one is listed here.',
    expandTip: 'Select a row to see every detail of that request.',
    dModel: 'Model', dInput: 'Input', dOutput: 'Output', dCacheWrite: 'Cache write', dCacheRead: 'Cache read',
    dTiming: 'First token / total', dListCost: 'List price', dMultiplier: 'Discount'
  },
  final: { title: 'Connect it now' }
}
