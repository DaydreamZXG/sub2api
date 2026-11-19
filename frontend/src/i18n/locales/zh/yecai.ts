export default {
  nav: { aria: '页面导航', pains: '解决什么', prices: '价格', faq: '常见问题', docs: '文档' },
  cta: { start: '开始使用', login: '登录', console: '进入控制台', how: '看看解决了什么' },
  flow: { aria: '你的工具通过一把密钥连到各家模型', key: '你的密钥' },
  hero: {
    eyebrow: 'AI 模型接入服务',
    titleA: '接入 AI',
    titleB: '不用再折腾配置',
    lead: 'Claude Code、Codex 这些工具，选好就能接上。价格直接写人民币，每一笔都查得到，没用完的可以退。'
  },
  pain: {
    before: '以前',
    after: '现在',
    setup: {
      tag: '接入',
      titleA: '以前，要自己改配置文件。',
      titleB: '现在，复制一次就接上。',
      b1: '翻文档找接口地址',
      b2: '打开配置文件，一项一项填',
      b3: '换一个工具，再来一遍',
      a1: '选你在用的工具',
      a2: '复制页面给出的内容',
      a3: '粘贴进去，就能用'
    },
    price: {
      tag: '价格',
      titleA: '以前，要自己算倍率。',
      titleB: '现在，直接写人民币。',
      formula: '充值汇率 × 分组倍率 × 模型倍率 × 官方价 = ？',
      b1: '同样的倍率，实际价格差好几倍',
      b2: '先选分组，才知道多少钱',
      b3: '扣的是额度，不是钱',
      a1: '每个模型一个人民币价格',
      a2: '充多少就是多少，¥1 就是 ¥1',
      a3: '想看倍率和原始价，点开就有'
    },
    bill: {
      tag: '账单',
      titleA: '以前，只看得到一个总数。',
      titleB: '现在，一笔一笔看得见。',
      b1: '余额少了，不知道花在哪',
      b2: '用了哪个模型，对不上',
      b3: '充进去的钱拿不回来',
      a1: '每次请求用了哪个模型、扣了多少',
      a2: '输入、输出、缓存分开列',
      a3: '没用完的余额可以退'
    },
    help: {
      tag: '服务',
      titleA: '以前，报错只能自己查。',
      titleB: '现在，卡住了有人回。',
      b1: '一串英文报错，看不懂',
      b2: '不知道是自己的问题还是站的问题',
      b3: '找不到人问',
      a1: 'QQ 群里直接问',
      a2: '提工单，有人跟进',
      a3: '接不上、缓存、网络问题都可以找我们'
    }
  },
  price: {
    eyebrow: '价格',
    title: '最后，说价格。',
    official: '官方价',
    perMillionIn: '每百万 Token 输入',
    off: '约 {zhe} 折',
    note: '以上为每百万 Token 的输入价，输出价和其他模型见完整价目表。官方价按 1 美元 ≈ {rate} 元折算，仅供对比。',
    all: '全部模型和价格'
  },
  faq: {
    eyebrow: '常见问题',
    title: '开始之前',
    q1: '我需要会写代码吗？',
    a1: '不需要。你只要会用 Claude Code 或 Codex 这类工具，按页面提示复制粘贴一次就行。',
    q2: '密钥是什么？',
    a2: '一串以 sk- 开头的密码，工具靠它来使用你的余额。它只属于你，不要发给别人。',
    q3: 'Token 是什么？',
    a3: '模型计算用量的单位，大致相当于字数：一个汉字约 1 到 2 个 Token。价格按每百万 Token 标注。',
    q4: '怎么收费？',
    a4: '按每次请求的实际用量扣余额，不同模型价格不同。花了多少随时能在账单里看到。',
    q5: '支持哪些工具？',
    a5: 'Claude Code、Codex、Gemini CLI、OpenCode 等常见工具都可以接。'
  },
  shell: { start: '开始', keys: '密钥', bills: '账单', topUp: '充值', admin: '管理后台', account: '我的账号' },
  dash: {
    night: '夜深了', morning: '早上好', afternoon: '下午好', evening: '晚上好', comma: '，',
    nextTag: '下一步', nextTopUp: '先充一点余额，用多少扣多少，没用完可以退。',
    nextKey: '创建你的第一把密钥，工具靠它来使用余额。', nextKeyGo: '去创建',
    nextConnect: '密钥有了，把它接到你的工具里就能用。', nextConnectGo: '去接入',
    balance: '余额', balanceHint: '用多少扣多少，没用完可以退', today: '今天花了', total: '累计花了',
    requests: '{n} 次请求', recent: '最近使用', allBills: '全部账单',
    empty: '还没有使用记录。接入工具并发出第一次请求后，这里会显示每一笔的明细。',
    expandTip: '点任意一行可以展开，看这次请求的全部细节。',
    dModel: '模型', dInput: '输入', dOutput: '输出', dCacheWrite: '缓存写入', dCacheRead: '缓存读取',
    dTiming: '首字 / 总用时', dListCost: '标准价', dMultiplier: '倍率'
  },
  final: { title: '现在就把它接上' }
}
