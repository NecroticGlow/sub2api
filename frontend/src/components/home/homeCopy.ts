const zh = {
  skip: '跳转到主要内容',
  scenesNav: '万物可能',
  modelsNav: '模型广场',
  connectNav: '快速接入',
  login: '登录',
  dashboard: '控制台',
  start: '开始创造',
  explore: '探索模型',
  pause: '暂停动效',
  resume: '开启动效',
  reduced: '已减少动态效果',
  menu: '打开导航',
  closeMenu: '关闭导航',
  kicker: '连接智能，让灵感生长',
  title: '万物一体，',
  titleAccent: '创造无界。',
  intro: '让不同的智能，在同一处相遇。从灵感、代码到应用，连接你的每一种可能。',
  heroAlt: '山川、植物、流体与星光汇成一颗悬浮的万物星球',
  heroCaption: '山川有形，想象无界。',
  scroll: '向下，探索万物',
  ecosystem: '一个入口，多种智能',
  ecosystemNote: '可用模型与分组以模型广场为准',
  sceneTitle: '一念生发，',
  sceneAccent: '万物皆有可能。',
  sceneIntro: '把复杂留在背后，让每一次与 AI 的相遇，都更接近你想创造的世界。',
  scenes: [
    {
      image: 'life',
      tag: '01 / INSPIRE',
      title: '让灵感生长',
      text: '写作、构思、知识探索。从一个问题出发，让零散的想法，长成完整的表达。',
      alt: '通透的翡翠玻璃种子中生长着蕨叶与银色根系'
    },
    {
      image: 'flow',
      tag: '02 / THINK',
      title: '让思考流动',
      text: '分析、推理、代码协作。在不同模型之间找到适合的视角，让思路不断向前。',
      alt: '银色飘带与水流融为一体，悬浮在深绿色水面上'
    },
    {
      image: 'cosmos',
      tag: '03 / BUILD',
      title: '让应用成形',
      text: '用熟悉的 API 接入智能。从第一个原型，到你亲手构建的产品与工作流。',
      alt: '翡翠晶体中展开一片由山川、银河与轨道组成的微型宇宙'
    }
  ],
  unionTitle: '把选择留给创造。',
  unionAccent: '把连接交给万物。',
  unionIntro:
    '模型、密钥、分组与用量，在一处有序连接。保留选择的自由，让接入更从容。',
  features: [
    {
      title: '多模型，一处接入',
      text: '按任务选择模型，使用对应接口接入已有应用与开发工具。'
    },
    {
      title: '每一项价格，都清楚',
      text: '在广场比较人民币实付价，查看输入、输出、缓存和分组倍率。'
    },
    {
      title: '从调用，到用量',
      text: '在控制台管理 API 密钥，查看调用记录和消耗，让使用有据可查。'
    }
  ],
  modelsTitle: '当下，值得探索。',
  modelsIntro: '探索当前公开目录中的可用模型。找到适合你的下一种智能。',
  allModels: '查看完整模型广场',
  groups: '个接入分组',
  loading: '正在连接模型目录…',
  modelsEmpty: '前往模型广场，查看当前可用模型与分组。',
  modelsFailed: '模型目录暂时未能加载，可在广场重试。',
  modelNote: '价格、缓存与接入详情，均可在模型广场查看。',
  connectTitle: '从这里，',
  connectAccent: '开始第一次连接。',
  connectIntro: '保留熟悉的开发方式，将万物接入你的应用。',
  steps: [
    {
      title: '选择模型与分组',
      text: '在模型广场比较用途与价格，选择合适的接入分组。'
    },
    {
      title: '创建 API 密钥',
      text: '登录控制台创建密钥，配置到应用的环境变量中。'
    },
    {
      title: '发起第一次调用',
      text: '设置接口地址和模型 ID，用熟悉的工具开始构建。'
    }
  ],
  getKey: '获取 API 密钥',
  docs: '查看文档',
  copy: '复制代码',
  copied: '已复制',
  copyFailed: '复制失败，请选择代码手动复制。',
  codeHint: 'OpenAI 兼容接口示例 · 请使用所选分组支持的模型 ID',
  prompt: '你好，万物。让我们开始创造。',
  faqTitle: '你可能还想知道',
  faqs: [
    {
      question: '如何选择适合自己的模型？',
      answer:
        '先在模型广场查看当前公开模型、分组与定价，再根据写作、推理或开发等具体任务进行选择。同一模型在不同分组的价格与接入方式可能不同。'
    },
    {
      question: '首页和广场的价格如何理解？',
      answer:
        '模型广场以人民币展示中转实付价，并计入生效倍率。输入、输出与缓存价格分别列出；阶梯、时段与分组差异请展开价格明细查看。'
    },
    {
      question: '已经在使用 SDK 或 AI 开发工具，如何接入？',
      answer:
        '在工具中填写对应的接口地址、API 密钥与模型 ID。OpenAI 兼容接口可参考上方示例；其他接口协议请以所选模型和分组的说明为准。'
    }
  ],
  closingTitle: '万物已连接。',
  closingAccent: '下一步，由你创造。',
  closingText: '每一个想法，都值得一次开始。',
  footerLine: '万物一体 · 智能无界',
  artwork: '首页视觉由 AI 生成',
  backTop: '回到顶部'
}

const en: typeof zh = {
  skip: 'Skip to main content',
  scenesNav: 'Possibilities',
  modelsNav: 'Model plaza',
  connectNav: 'Get connected',
  login: 'Sign in',
  dashboard: 'Dashboard',
  start: 'Start creating',
  explore: 'Explore models',
  pause: 'Pause motion',
  resume: 'Enable motion',
  reduced: 'Reduced motion',
  menu: 'Open navigation',
  closeMenu: 'Close navigation',
  kicker: 'CONNECTED INTELLIGENCE. GROWING IDEAS.',
  title: 'Everything connects.',
  titleAccent: 'Possibility unfolds.',
  intro:
    'A place where different kinds of intelligence meet. Connect your ideas, your code, and everything you want to build.',
  heroAlt:
    'A floating world formed from mountains, plants, flowing silver and starlight',
  heroCaption: 'Rooted in nature. Open to possibility.',
  scroll: 'Scroll to explore',
  ecosystem: 'One entry. Many intelligences.',
  ecosystemNote: 'See the model plaza for available models and groups',
  sceneTitle: 'One thought.',
  sceneAccent: 'A world of possibilities.',
  sceneIntro:
    'Make room for ideas. Bring every encounter with AI closer to the world you want to create.',
  scenes: [
    {
      image: 'life',
      tag: '01 / INSPIRE',
      title: 'Let ideas grow',
      text: 'Writing, ideation and discovery. Start with a question and turn scattered thoughts into a complete expression.',
      alt: 'A fern and silver roots growing inside a translucent jade glass seed'
    },
    {
      image: 'flow',
      tag: '02 / THINK',
      title: 'Let thinking flow',
      text: 'Analysis, reasoning and coding. Find a useful perspective across models and keep your thinking moving.',
      alt: 'A silver ribbon merging into flowing water above a dark green pool'
    },
    {
      image: 'cosmos',
      tag: '03 / BUILD',
      title: 'Bring ideas to life',
      text: 'Connect intelligence through familiar APIs. Build your first prototype, your product or your own workflow.',
      alt: 'An emerald crystal opening into a miniature universe of mountains and stars'
    }
  ],
  unionTitle: 'Your freedom to create.',
  unionAccent: 'Our place to connect.',
  unionIntro:
    'Models, keys, groups and usage, brought together. Keep your choices open and your connections simple.',
  features: [
    {
      title: 'Many models. One place.',
      text: 'Choose a model for your task and connect to your existing applications through its supported API.'
    },
    {
      title: 'See what you pay',
      text: 'Compare relay prices in CNY, including input, output, caching and group multipliers.'
    },
    {
      title: 'From calls to usage',
      text: 'Manage API keys, review request records and track consumption in your dashboard.'
    }
  ],
  modelsTitle: 'Worth exploring, right now.',
  modelsIntro:
    'Explore available models in the current public catalog. Find the intelligence for your next idea.',
  allModels: 'Explore the full model plaza',
  groups: 'access groups',
  loading: 'Connecting to the model catalog…',
  modelsEmpty:
    'Visit the model plaza to see currently available models and groups.',
  modelsFailed:
    'The catalog could not be loaded. Please try again in the model plaza.',
  modelNote: 'Find pricing, caching and connection details in the model plaza.',
  connectTitle: 'Your next idea.',
  connectAccent: 'Your first connection.',
  connectIntro: 'Keep the tools you know. Bring Wanwu into what you build.',
  steps: [
    {
      title: 'Choose a model and group',
      text: 'Compare models and prices in the plaza, then choose an access group.'
    },
    {
      title: 'Create an API key',
      text: 'Sign in to create a key and add it to your application environment.'
    },
    {
      title: 'Make your first call',
      text: 'Set the API address and model ID, and start building with familiar tools.'
    }
  ],
  getKey: 'Get an API key',
  docs: 'Read the docs',
  copy: 'Copy code',
  copied: 'Copied',
  copyFailed: 'Copy failed. Please select and copy the code manually.',
  codeHint:
    'OpenAI-compatible example · Use a model supported by your chosen group',
  prompt: 'Hello, Wanwu. Let us create something.',
  faqTitle: 'A few things to know',
  faqs: [
    {
      question: 'How do I choose the right model?',
      answer:
        'Explore the currently available models, groups and prices in the model plaza, then choose for your writing, reasoning or development task. Pricing and access can vary across groups for the same model.'
    },
    {
      question: 'How should I read the prices?',
      answer:
        'The model plaza shows relay prices in CNY with the effective multiplier applied. Input, output and caching are listed separately. Open the pricing details for context tiers, time periods and group differences.'
    },
    {
      question: 'Can I use my existing SDK or AI tools?',
      answer:
        'Configure your tool with the appropriate API address, API key and model ID. The example above is for OpenAI-compatible APIs. For other protocols, follow the guidance for your selected model and group.'
    }
  ],
  closingTitle: 'Everything is connected.',
  closingAccent: 'What happens next is yours.',
  closingText: 'Every idea deserves a beginning.',
  footerLine: 'Connected intelligence. Endless possibility.',
  artwork: 'Homepage artwork generated with AI',
  backTop: 'Back to top'
}

export const homeCopy = { zh, en }
