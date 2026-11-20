<!-- Sections added by Devin, not in the user's original document. Each is placed in the chapter named by `chapter:` and shown with a 新增 badge. Everything here follows the official page linked under 官方参考; none of it was installed and tested. -->

<!-- chapter: 三 -->
## Factory Droid CLI ✅ 直接支持

Droid 是 Factory 出的终端编程 Agent，官方支持自带密钥（BYOK）：自定义模型写在本机的 `~/.factory/settings.json`，密钥不会上传到 Factory。

**配置步骤（界面方式）**

1. 在 Droid 里输入 `/settings`，打开 Custom models；或输入 `/model`，选 Manage custom models。

2. 选 Add model，选一个 provider，填地址、密钥和模型 ID，然后 Save changes。

3. 保存后的模型出现在 `/model` 的 Custom Models 下，按 `displayName` 显示。

**配置步骤（配置文件方式）**

编辑 `~/.factory/settings.json`，在 `customModels` 里加一项：

```JSON
{
  "customModels": [
    {
      "model": "请替换为模型ID",
      "displayName": "野菜 AI",
      "baseUrl": "https://yeschoy.com/v1",
      "apiKey": "请替换为你的密钥",
      "provider": "generic-chat-completion-api",
      "maxOutputTokens": 16384
    }
  ]
}
```

**注意事项**

- `provider` 决定走哪种协议：`generic-chat-completion-api` 是 OpenAI Chat Completions 兼容格式；`anthropic` 是 Anthropic Messages 格式。

- 用 `anthropic` 时密钥默认放在 `x-api-key` 头里；网关要求 `Authorization: Bearer` 的话，加一行 `"authMode": "bearer"`。

- `apiKey` 可以写成 `${变量名}`，从环境变量读取，不必把密钥明文写进文件。

- `generic-chat-completion-api` 的模型默认不发图片（`noImageSupport` 默认为 `true`），要用识图模型需把它设为 `false`。

- `maxContextLimit` 填模型真实的上下文窗口，Droid 按它压缩对话。

- 旧版的 `~/.factory/config.json` 只在启动时被复制一次，之后请改 `settings.json`。

**官方参考**：[Factory：Custom Models (BYOK)](https://docs.factory.com/model-independence/byok)

---

<!-- chapter: 三 -->
## GitHub Copilot CLI ✅ 直接支持

GitHub Copilot 的命令行版本官方支持添加自己的模型供应商（BYOK），OpenAI 兼容接口和 Anthropic 接口都可以。这里说的是 Copilot CLI，不是编辑器里的 Copilot 插件。

**配置步骤**

打开或新建 `~/.copilot/providers.json`，写入：

```JSON
{
  "providers": [
    {
      "name": "yeschoy",
      "type": "openai",
      "baseUrl": "https://yeschoy.com/v1",
      "apiKey": "请替换为你的密钥"
    }
  ],
  "models": [
    {
      "id": "请替换为模型ID",
      "provider": "yeschoy"
    }
  ]
}
```

1. 文件里已经有内容时，把这两项加进原有的 `providers` 和 `models`，不要整个覆盖。

2. 保存后重启 Copilot CLI。

3. 启动时指定模型：`copilot --model yeschoy/请替换为模型ID`；会话中用 `/model` 切换。

**注意事项**

- 这个文件里的密钥是明文保存的，不要提交到仓库或发给别人。写成 `$变量名` 不会去读环境变量，会被当成普通文字。

- `type` 可选 `openai`（默认，OpenAI 兼容）、`azure`、`anthropic`。

- 模型只支持 Responses 接口时，在 provider 里加 `"wireApi": "responses"`，默认是 `completions`。

- 模型和接口需要支持工具调用和流式输出。

- 不认识的模型默认按 128,000 输入 token 处理，可在模型上用 `maxPromptTokens`、`maxOutputTokens` 填真实上限。

**官方参考**：[GitHub Docs：Adding LLM models to GitHub Copilot CLI](https://docs.github.com/en/copilot/how-tos/copilot-cli/customize-copilot/use-byok-models)

---

<!-- chapter: 三 -->
## Kimi Code CLI ✅ 直接支持

Kimi Code CLI 是月之暗面的终端编程 Agent，除了自家服务，也能在配置文件里接入 OpenAI 兼容或 Anthropic 协议的第三方服务。

**配置步骤（配置文件方式）**

编辑 `~/.kimi-code/config.toml`，加一个供应商和一个模型：

```TOML
[providers.yeschoy]
type = "openai"
base_url = "https://yeschoy.com/v1"
api_key = "请替换为你的密钥"

[models."yeschoy-model"]
provider = "yeschoy"
model = "请替换为模型ID"
max_context_size = 200000
```

**配置步骤（界面方式）**

1. 在 Kimi Code CLI 里输入 `/provider` 打开供应商管理。

2. 在 `[ Add New Platform ]` 一行按回车，按提示添加。

**注意事项**

- `type` 决定协议：`openai` 是 Chat Completions 及兼容服务，`anthropic` 是 Anthropic Messages，`openai_responses` 是 Responses 接口。

- 密钥只从配置文件读：要么写 `api_key`，要么用 `api_key_env = "变量名"` 指定环境变量，两者只能选一个。只在终端里 `export` 密钥不会生效。

- `max_context_size` 请按模型定价页填真实的上下文长度，上面的 200000 只是示例。

- 思考、识图、工具调用等能力按模型名前缀自动匹配；没被识别的模型需要在模型条目上自己声明 `capabilities`，写法见官方文档。

- 网关把思考内容放在非标准字段时，在模型条目上设置 `reasoning_key`。

**官方参考**：[Kimi Code：Providers and models](https://www.kimi.com/code/docs/en/kimi-code-cli/configuration/providers)、[Kimi Code：Configuration files](https://www.kimi.com/code/docs/en/kimi-code-cli/configuration/config-files)

---

<!-- chapter: 二 -->
## Zed ✅ 直接支持

Zed 是一款带 AI Agent 的代码编辑器，官方支持添加 OpenAI 兼容的自定义供应商。

**配置步骤（图形界面）**

1. 打开命令面板，执行 `agent: open settings`。

2. 在 LLM Providers 里点 Add Provider，填供应商名称、API 地址 `https://yeschoy.com/v1` 和模型 ID。

3. 在这个供应商的设置里填入 API 密钥。

**配置步骤（配置文件方式）**

在 Zed 的 `settings.json` 里加入：

```JSON
{
  "language_models": {
    "openai_compatible": {
      "yeschoy": {
        "api_url": "https://yeschoy.com/v1",
        "available_models": [
          {
            "name": "请替换为模型ID",
            "display_name": "野菜 AI",
            "max_tokens": 128000
          }
        ]
      }
    }
  }
}
```

**注意事项**

- 密钥不要写进 `settings.json`：在界面里填（保存在系统钥匙串），或设置环境变量。变量名是供应商名转成大写下划线再加 `_API_KEY`，上面的例子是 `YESCHOY_API_KEY`。

- `max_tokens` 是模型的上下文长度，请按模型定价页填真实值，128000 只是示例。

- 自定义模型默认开启工具调用、关闭图片输入。

- 模型只支持 Responses 接口时，在模型上加 `"capabilities": { "chat_completions": false }`。

**官方参考**：[Zed：Use API Access](https://zed.dev/docs/ai/use-api-access)

---

<!-- chapter: 二 -->
## GitHub Copilot（VS Code）⚠️ 有限支持

VS Code 里的 Copilot 聊天官方支持自带密钥，并且可以用 Custom Endpoint 供应商接入任意兼容接口（Chat Completions、Responses、Anthropic Messages 三种格式）。它只管聊天，代码补全不走你的地址。

**配置步骤**

1. 在聊天输入框的模型选择器里点 Manage Language Models（齿轮图标），或在命令面板执行 Chat: Manage Language Models。

2. 点 Add Models，选 Custom Endpoint。

3. 填分组名、显示名称和 API 密钥。

4. API 类型选 Chat Completions。

5. VS Code 会打开 `chatLanguageModels.json`，按下面的样子填好模型后保存。

6. 回到聊天，在模型选择器里选这个模型；没出现就重启 VS Code。

```JSON
[
  {
    "name": "野菜 AI",
    "vendor": "customendpoint",
    "apiKey": "${input:yeschoyApiKey}",
    "apiType": "chat-completions",
    "models": [
      {
        "id": "请替换为模型ID",
        "name": "野菜 AI",
        "url": "https://yeschoy.com/v1/chat/completions",
        "toolCalling": true,
        "vision": true,
        "maxInputTokens": 200000,
        "maxOutputTokens": 64000
      }
    ]
  }
]
```

**注意事项**

- 自带密钥只用于聊天和标题生成这类辅助任务。代码补全、语义搜索仍然需要 GitHub 账号和 Copilot 套餐，不会走你的地址。

- `url` 建议填完整接口地址。只填到域名或 `/v1` 时，VS Code 会按 API 类型自己补路径。

- `apiType` 可选 `chat-completions`（默认）、`responses`、`messages`。用 `messages` 时密钥放在 `x-api-key` 头里，其余放在 `Authorization: Bearer`。

- 用 Responses 接口而服务端不保存历史时，在模型上加 `"statefulResponses": false`。

- `maxInputTokens`、`maxOutputTokens`、`vision` 请按模型定价页的真实情况填，上面只是示例。

- Copilot Business / Enterprise 账号需要管理员在 GitHub 的 Copilot 策略里打开“Bring Your Own Language Model Key in VS Code”。

**官方参考**：[VS Code：AI language models](https://code.visualstudio.com/docs/agent-customization/language-models)

---

<!-- chapter: 三 -->
## Antigravity CLI 🔄 协议转换

Antigravity 的命令行版本（`agy`）官方支持用自己的 Gemini API 密钥，并可以指向自定义的 Gemini 兼容地址，接法和 Gemini CLI 相同，Base URL 不带 `/v1`。Antigravity 的 IDE 不能这样接，见第 20 节。

**配置步骤**

1. 编辑 `~/.gemini/antigravity-cli/settings.json`，写入：

```JSON
{
  "modelProvider": "gemini"
}
```

然后设置环境变量并启动：

```Bash
export GEMINI_API_KEY="请替换为你的密钥"
export GOOGLE_GEMINI_BASE_URL="https://yeschoy.com"
agy
```

**注意事项**

- 只设置 `GEMINI_API_KEY` 不会生效，必须先把 `modelProvider` 设为 `gemini`。

- 设了 `modelProvider` 却没有 `GEMINI_API_KEY` 时，CLI 无法启动。

- 启动后会跳过登录，顶部显示 Gemini API key；这种方式下 `/logout` 不起作用。

- `export` 只对当前终端有效，要长期使用请写进 `~/.zshrc` 或 `~/.bashrc`。

- 想换回账号登录：删掉 `modelProvider` 后重启 CLI。

**官方参考**：[Antigravity：Authenticate](https://antigravity.google/docs/install/)

---

<!-- chapter: 一 -->
## Jan ✅ 直接支持

Jan 是一款开源的桌面聊天软件，官方支持添加自定义供应商，OpenAI 兼容和 Anthropic 兼容两种格式都可以。

**配置步骤**

1. 在 Jan 里打开添加自定义供应商（Add Custom Provider）的对话框。

2. API 格式选 OpenAI-compatible（默认）。

3. Provider name 填"野菜 AI"，Base URL 填 `https://yeschoy.com/v1`，API key 填 `请替换为你的密钥`。

4. 点 Create。

5. Jan 会自动从接口拉取模型列表；拉不到时，在这个供应商的 Models 区域点 + 手动添加模型 ID。

**注意事项**

- OpenAI 兼容格式的 Base URL 要带 `/v1`，漏掉会 404。末尾的斜杠会被自动去掉。

- API 格式选错时，即使地址和密钥都对，请求也会失败。

- API key 是必填项。

**官方参考**：[Jan：Custom Endpoints](https://docs.jan.ai/desktop/remote-models/custom-endpoint)

---

<!-- chapter: 一 -->
## LibreChat ✅ 直接支持

LibreChat 是自己部署的网页聊天平台，由管理员在 `librechat.yaml` 里添加自定义端点（custom endpoint）。

**配置步骤**

1. 在 `librechat.yaml` 里新增一个自定义端点，完整结构照官方参考里的示例写。

2. `baseURL` 填 `https://yeschoy.com/v1`。

3. `apiKey` 建议写成环境变量引用，例如 `${YESCHOY_API_KEY}`，再在部署环境里设置这个变量。

4. 模型列表里填完整的模型 ID，保存后重启 LibreChat。

**注意事项**

- LibreChat 默认会在 `baseURL` 后面自动加 `/chat/completions`，所以只填到 `/v1`。如果你填的是完整接口地址，需要把 `directEndpoint` 设为 `true`。

- `apiKey` 和 `baseURL` 都可以填 `user_provided`，让每个用户在界面里自己填。

- 用 `provider: anthropic` 时，`baseURL` 填接口根地址（不带 `/v1`），LibreChat 会走 Anthropic 原生的 `/v1/messages`。

**官方参考**：[LibreChat：Custom Endpoint Object Structure](https://www.librechat.ai/docs/configuration/librechat_yaml/object_structure/custom_endpoint)

---

<!-- chapter: 五 -->
## AnythingLLM ✅ 直接支持

AnythingLLM 是知识库问答平台，用官方的 Generic OpenAI 供应商可以接入任意 OpenAI 兼容接口。官方提醒这是面向开发者的选项，每一项都要填对。

**配置步骤**

1. 在设置的 LLM 供应商里选 Generic OpenAI。

2. Base URL 填 `https://yeschoy.com/v1`，API 密钥填 `请替换为你的密钥`，模型填完整的模型 ID，其余项（如 Max Tokens）按模型实际情况填。

3. 保存后，先在一个空的工作区里发一句短消息测试，再试文档问答和 Agent。

**注意事项**

- Base URL 填到 `/v1` 为止，不要填完整的 `/chat/completions` 地址。

- 输出上限默认用 `max_tokens` 字段发送。接口要求 `max_completion_tokens` 时，在服务端设置环境变量 `GENERIC_OPEN_AI_MODEL_MAX_TOKEN_KEY="max_completion_tokens"` 并重启。

- 能聊天但流式输出不正常时，按官方文档关闭 Generic OpenAI 的流式输出。

**官方参考**：[AnythingLLM：OpenAI (Generic) LLM](https://docs.anythingllm.com/setup/llm-configuration/cloud/openai-generic)

---

<!-- chapter: 三 -->
## Amp ✅ 直接支持

Amp 是一款编程 Agent，官方的 Model Routing 支持添加 Custom URL 连接，可以接任意 OpenAI 兼容或 Anthropic 兼容的地址。

**配置步骤**

1. 按官方 Model Routing 文档，新增一个 Custom URL 连接。

2. API 格式选 `chat-completions`（默认），Base URL 填 `https://yeschoy.com/v1`，API key 填 `请替换为你的密钥`。

3. 表单会实时显示最终请求地址，确认它是 `https://yeschoy.com/v1/chat/completions` 再保存。

**注意事项**

- 只填 Base URL，路径由 Amp 按格式自动补：`chat-completions` 补 `/chat/completions`，`responses` 补 `/responses`，`anthropic-messages` 补 `/v1/messages`。

- 用 `anthropic-messages` 时 Base URL 填 `https://yeschoy.com`；已经带 `/v1` 也可以，不会重复。

- Base URL 里已经包含接口路径时 Amp 会给出警告。

- 密钥以 `Authorization: Bearer` 发送；需要额外请求头时用 Headers 设置或命令行的 `--headers`。

**官方参考**：[Amp：Model Routing](https://ampcode.com/docs/customize/model-routing)
