# Argo Proxy 系统设计

基于 [prd.md](./prd.md) 的架构与模块设计。

---

## 1. 概述

Chrome 扩展，按域名/路径将请求路由到不同代理服务器，供测试/开发做环境切换。核心能力：规则配置（域名或路径前缀 → 代理）、启用/禁用总开关、Popup 与 Settings 界面、配置持久化。

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

```
argo_proxy/
├── manifest.json           # MV3, permissions, options_ui 指向 settings 页, background
├── popup/
│   ├── popup.html          # 主菜单：启用/禁用、Host 列表、设置
│   ├── popup.js            # 逻辑与子菜单切换、与 SW 通信
│   └── popup.css
├── settings/
│   ├── settings.html       # 设置页
│   ├── settings.js         # 代理与规则表单、storage 读写
│   └── settings.css
├── background/
│   └── service-worker.js   # storage 监听、DNR 更新、messaging
├── shared/
│   └── storage-schema.js   # 数据结构常量、默认值（可选）
└── icons/                  # 扩展图标
```

以上设计可直接作为实现与代码结构依据，并与 [prd.md](./prd.md) 的验收标准对齐。
