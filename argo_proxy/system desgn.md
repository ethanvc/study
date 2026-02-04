# Argo Proxy 系统设计

基于 [prd.md](./prd.md) 的架构与模块设计。

---

## 1. 概述

Chrome 扩展，按域名/路径将请求路由到不同代理服务器，供测试/开发做环境切换。核心能力：拦截请求并根据规则设置代理、规则配置（域名或路径前缀 → 代理）、启用/禁用总开关、Popup 与 Settings 界面、配置持久化。

---

## 2. 整体架构

```
┌─────────────────────────────────────────────────────────────┐
│                    Chrome Extension (MV3)                    │
├─────────────────────────────────────────────────────────────┤
│  Popup (popup.html)     │  主菜单：启用/禁用、Host 列表、设置  │
│  └─ 子菜单：Host → 代理列表                                   │
├─────────────────────────────────────────────────────────────┤
│  Settings (settings.html) │  代理列表 CRUD、规则 CRUD（匹配条件  │
│                          │  域名/路径 → 选择代理）              │
├─────────────────────────────────────────────────────────────┤
│  Service Worker          │  读取配置；生成/更新 DNR 规则；      │
│  (background)            │  响应 Popup/Settings 的读写请求     │
├─────────────────────────────────────────────────────────────┤
│  chrome.storage.local    │  持久化：enabled、proxies、rules   │
└─────────────────────────────────────────────────────────────┘
         │
         ▼
  declarativeNetRequest    按规则对请求做 redirect 至对应代理
  (或 webRequest + 代理网关，见 5)
```

- **Popup**：仅做展示与总开关、进入 Host 子菜单、打开 Settings，不直接写存储，通过 message 与 Service Worker 通信。
- **Settings**：代理与规则的增删改，写 `chrome.storage.local`（可由 SW 封装或直接写）。
- **Service Worker**：总开关与规则变更时，同步到 declarativeNetRequest 或代理配置，使「按域名/路径走不同代理」生效。

---

## 3. 数据模型

**代理 (Proxy)**

- `id`：唯一标识（如 UUID）
- `name`：展示用名称（如「团队 A 代理」）
- `type`：`http` | `https` | `socks`
- `host`：代理主机
- `port`：端口
- `username`、`password`：可选，认证用；存储时需考虑仅本地、不上传（见非功能）

**规则 (Rule)**

- `id`：唯一标识
- `matchType`：`domain` | `pathPrefix`
- `value`：域名（如 `api.example.com`）或路径前缀（如 `/api/order`）
- `proxyId`：关联的代理 id
- `order`：排序，数值越小越优先匹配

**全局状态**

- `enabled`：boolean，总开关；为 false 时所有请求直连，不应用任何规则。

---

## 4. 请求匹配与代理生效

- 请求发出时，由 Service Worker 侧根据 `enabled` 与当前 `rules` 决定是否、如何代理。
- **匹配逻辑**：按 `order` 升序对规则排序，对请求 URL 的 host 与 path 依次匹配；命中即采用该规则的 `proxyId`，不再继续。
- **生效方式**（二选一或组合，视 Chrome 能力与实现选型）：
  - **declarativeNetRequest**：根据规则动态生成 redirect 规则，将匹配到的请求 redirect 到「代理网关」URL（网关再根据原 URL 与配置转发到对应上游代理）。规则变更时调用 `chrome.declarativeNetRequest.updateDynamicRules`。
  - **代理网关 + 全局代理**：扩展只控制「当前使用的代理」或「PAC」；若不做 PAC（PRD 边界），则需一个本地或远端代理网关，扩展将请求 redirect 到该网关，网关按 host/path 查表转发到不同上游代理。此时扩展侧只维护规则与网关 URL 的映射即可。

首版建议：优先用 **declarativeNetRequest + 单一代理网关 URL**（或每个代理对应一个网关 URL），规则里存「匹配条件 → proxyId」，SW 里把 proxyId 映射为网关 redirect URL，便于实现与后续替换为本地网关。

---

## 5. 模块划分

| 模块               | 职责                                                                                                          |
| ------------------ | ------------------------------------------------------------------------------------------------------------- |
| **Popup**          | 总开关 UI、Host 列表入口与子菜单（只读展示 host → 代理）、打开 Settings；与 SW 通信读 enabled/rules/proxies。 |
| **Settings**       | 代理列表与规则的表单增删改；读写 storage；可调用 SW 通知「配置已更新」以刷新 DNR。                            |
| **Service Worker** | 读取/监听 storage；根据 enabled + rules 生成 DNR 动态规则或更新代理配置；响应 popup/settings 的 messaging。   |
| **Storage**        | 使用 `chrome.storage.local` 存 `enabled`、`proxies`、`rules`；敏感字段（如 password）考虑仅内存或加密后存。   |

Popup 与 Settings 不共享 DOM，仅通过 storage 与 messaging 与 SW 同步状态。

---

## 6. 关键流程

**总开关切换（Popup）**

1. 用户切换 Switch → Popup 发 message 给 SW（或直接写 storage）。
2. SW 监听到 `enabled` 变更 → 若为 false 则清除/禁用 DNR 规则（或恢复直连）；若为 true 则根据当前 rules 重新生成 DNR 规则。

**规则变更（Settings）**

1. 用户在 Settings 增删改规则/代理 → 写 `chrome.storage.local`。
2. SW 监听 `chrome.storage.onChanged` → 重新生成并提交 DNR 动态规则（仅当 `enabled === true`）。

**Host 列表展示（Popup 子菜单）**

1. Popup 打开时（或进入子菜单时）向 SW 发 message 请求当前 `rules` + `proxies`。
2. SW 从 storage 读取并返回「匹配条件 → 代理名」列表；Popup 按 order 渲染，无规则时显示「暂无规则」并引导去 Settings。

**打开设置**

- Popup 中点击「设置」→ `chrome.runtime.openOptionsPage()`。

---

### 6.5 监听并记录 tab 发出的请求列表（可选）

**现状**：当前实现**未**监听单个请求，仅通过 declarativeNetRequest 声明 redirect 规则，由浏览器内核匹配并重定向，扩展侧无“每个请求”的回调，也无法拿到“某 tab 的请求列表”。

**若需“监听并记录某 tab 发出的所有请求”**，可按以下方式补充实现。

**API 选型**

- **chrome.webRequest**：在 MV3 中仍可做**只读**监听（不阻塞、不修改请求）。在 background 中注册 `onBeforeRequest` 或 `onCompleted`，即可在每次请求时收到回调（含 `tabId`、`url`、`method`、`type` 等），用于汇总成“请求列表”。
- **chrome.declarativeNetRequest.getMatchedRules**：可查询一段时间内**匹配到的规则**及对应请求信息，适合“哪些请求被重定向了”的统计，而非“某 tab 的全部请求”的实时列表；且需开启 `declarativeNetRequestFeedback` 权限。

**推荐**：用 **webRequest.onBeforeRequest**（或 onCompleted）在 Service Worker 中监听，按 `tabId` 聚合为请求列表。

**实现要点**

1. **权限**：manifest 中已有 `host_permissions: ["<all_urls>"]` 时，可直接使用 `chrome.webRequest.onBeforeRequest`；无需额外声明 `webRequest` 权限（MV3 下 host 权限即够）。
2. **监听注册**：在 `extension/background/service-worker.js` 中：
   - `chrome.webRequest.onBeforeRequest.addListener(callback, { urls: ['<all_urls>'] }, ['requestBody'])`（无需 requestBody 可省略第三参数）。
   - callback 参数包含 `requestId`、`url`、`method`、`type`、`tabId`、`frameId` 等。
3. **数据结构**：在内存中维护「按 tab 聚合的请求列表」，例如：
   - `requestLogByTab: Map<tabId, Array<{ url, method, type, timestamp, requestId? }>>`；
   - 单 tab 列表长度设上限（如 200 条），超出则 FIFO 丢弃；或只保留最近 N 秒。
4. **生命周期**：在 `chrome.tabs.onRemoved` 中根据 `tabId` 删除对应列表，避免内存泄漏。
5. **与 Popup 的联动**：若 Popup 需展示“当前 tab 的请求列表”，可通过 messaging 向 SW 请求当前 `tabId` 的列表（SW 用 `chrome.tabs.query({ active: true, currentWindow: true })` 取 tabId，或由 Popup 在发 message 时带上 tabId）；SW 返回该 tab 的请求数组。

**与 DNR 的关系**

- 使用 webRequest 做**只读**监听不会与 declarativeNetRequest 冲突：DNR 负责 redirect，webRequest 只观察，不修改请求。
- 若需在“请求列表”中标注“该请求被哪条规则重定向”，可在 callback 中根据 `url` 用当前 `rules` 做一次同步匹配计算，或后续再查 `getMatchedRules`（需 `declarativeNetRequestFeedback`）。

**可选权限**

- 若使用 `getMatchedRules` 做匹配反馈，需在 manifest 的 `permissions` 中增加 `declarativeNetRequestFeedback`。

---

## 7. Chrome API 与权限

- **storage**：`chrome.storage.local` 持久化配置。
- **declarativeNetRequest**：动态规则做请求 redirect（需 host 权限或 declarativeNetRequest 权限及匹配的 host）。
- **scripting / tabs**：若需在 Popup 中获取当前 tab 的 host 做展示，可用 `chrome.tabs.query` 等（可选）。
- **options_ui**：在 manifest 中声明 `options_ui` 指向 settings 页。

敏感权限（如 `<all_urls>`）按最小必要原则申请，并在隐私说明中注明仅用于代理路由、数据不上传。

---

## 8. 非功能与安全

- **安全**：代理凭证仅存本地（storage.local），不随请求发往除代理服务器外的第三方；Settings 与 Popup 仅在扩展上下文中运行。
- **兼容**：Manifest V3；Service Worker 无持久化进程，逻辑保持无状态、以 storage 为事实来源。
- **性能**：规则数量设上限（如 50），避免 DNR 规则过多；Storage 变更后批量更新 DNR，避免频繁调用 API。

---

## 9. 文件与目录建议

Chrome 扩展代码与配置均位于 **extension/** 目录下，加载已解压的扩展程序时选择该目录。

```
argo_proxy/
├── extension/              # Chrome 扩展（加载时选此目录）
│   ├── manifest.json       # MV3, options_ui, background
│   ├── popup/
│   │   ├── popup.html
│   │   ├── popup.js
│   │   └── popup.css
│   ├── settings/
│   │   ├── settings.html
│   │   ├── settings.js
│   │   └── settings.css
│   ├── background/
│   │   └── service-worker.js
│   ├── shared/
│   │   └── storage-schema.js
│   └── icons/
├── prototype/              # 静态原型
├── prd.md
├── system desgn.md
└── TASK_PLAN.md
```

以上设计可直接作为实现与代码结构依据，并与 [prd.md](./prd.md) 的验收标准对齐。
