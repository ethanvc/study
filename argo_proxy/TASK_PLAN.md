# Argo Proxy Chrome 扩展开发任务计划

## 目标与依据

- **目标**：开发 MV3 Chrome 扩展，实现按域名/路径将请求路由到不同代理，供测试/开发环境切换。
- **依据**：[prd.md](prd.md)、[system desgn.md](system%20desgn.md)、[prototype/popup.html](prototype/popup.html)、[prototype/settings.html](prototype/settings.html)。

---

## 阶段一：项目骨架与配置

- **1.1** 在 `argo_proxy/` 下按系统设计建立目录：`popup/`、`settings/`、`background/`、`shared/`、`icons/`。
- **1.2** 编写 `manifest.json`（Manifest V3）：`action` 指向 popup；`background.service_worker`；`options_ui` 指向 settings 页；权限至少包含 `storage`、`declarativeNetRequest`（及所需 `declarativeNetRequestWithHostAccess` 或 host 权限）；提供 16/48 图标占位。
- **1.3** 在 `shared/storage-schema.js`（或等价模块）中定义数据结构和 Key：`enabled`（boolean）、`proxies`（数组，元素含 id/name/type/host/port/username/password）、`rules`（数组，元素含 id/matchType/value/proxyId/order）；约定 `proxyId === 'direct'` 表示直连。

---

## 阶段二：Storage 与 Background

- **2.1** 实现 storage 读写封装（可在 background 或 shared）：读取/写入 `enabled`、`proxies`、`rules`；监听 `chrome.storage.onChanged`，在 background 内处理。
- **2.2** 实现 Service Worker（`background/service-worker.js`）：启动时从 storage 读取配置；`onChanged` 时若 `enabled` 或 `rules`/`proxies` 变化，更新 DNR 动态规则（见 2.3）；提供 messaging 接口：获取当前配置（getConfig）、设置 enabled（setEnabled），供 Popup 调用。
- **2.3** 实现 DNR 规则逻辑：根据 `enabled` 与 `rules`（按 order 排序）生成 redirect 规则；首版采用「redirect 到代理网关」方案：每条规则映射为一条 DNR 动态规则，将匹配的 URL redirect 到网关 URL（网关参数或 path 携带原 URL 及 proxyId）；若 `enabled === false` 则清除全部动态规则。规则数量上限 50（与系统设计一致）。

---

## 阶段三：Popup

- **3.1** 将 [prototype/popup.html](prototype/popup.html) 的布局与样式迁移到 `popup/popup.html` + `popup/popup.css`，保持主菜单（启用/禁用、Host 列表、设置）与子菜单（Host 列表 + 代理下拉）结构；有子菜单的项保留箭头，设置项无箭头。
- **3.2** 实现 `popup/popup.js`：从 background 拉取 `enabled`、`proxies`、`rules`；总开关与 `enabled` 双向同步（点击写 storage 或发 message）；Host 列表子菜单展示「匹配条件 → 代理」列表，代理为下拉（含「直连」+ 各 proxy 名称），默认展示当前生效代理，选择变更时写回 storage（或通过 message 通知 SW）；无规则时显示「暂无规则」并引导打开 settings；点击「设置」调用 `chrome.runtime.openOptionsPage()`。
- **3.3** 总开关关闭时，Host 列表项显示「已暂停」等状态，子菜单仍可打开查看/编辑；子菜单支持返回主菜单。

---

## 阶段四：Settings 页

- **4.1** 将 [prototype/settings.html](prototype/settings.html) 的布局与样式迁移到 `settings/settings.html` + `settings/settings.css`，保留「代理列表」与「路由规则」两大块及添加/编辑/删除按钮区域。
- **4.2** 实现 `settings/settings.js`：从 storage 读取并渲染代理列表与规则列表；代理 CRUD：表单（name、type、host、port、用户名/密码可选）增删改，写 `chrome.storage.local`；规则 CRUD：匹配类型（域名/路径前缀）、匹配值、选择代理（含直连），排序 order，写 storage；规则数量与代理数量限制与系统设计一致（如规则上限 50）；保存后无需额外调用，依赖 storage.onChanged 触发 SW 更新 DNR。

---

## 阶段五：联调与验收

- **5.1** 端到端验证：配置一条「域名 → 代理 A」规则，开启总开关，访问该域名，确认请求被 redirect 到预期网关（或代理）；关闭总开关，确认直连；重启浏览器后配置仍在。
- **5.2** 按 [prd.md](prd.md) 验收标准检查：规则与代理持久化；未匹配规则直连；Popup/Settings 与系统设计一致。
- **5.3** 安全与边界：确认代理凭证仅存本地；不申请超出需求的权限；首版不做 PAC、不做系统级代理、不做导入导出与多配置集（留 V2）。

---

## 交付物

- 在 **argo_proxy 根目录** 新增任务计划文档（建议文件名：`TASK_PLAN.md` 或 `任务计划.md`），内容为本计划正文（含上述阶段与子项），便于团队按任务勾选与排期。

---

## 可选与后续（V2）

- 代理认证（用户名/密码）在请求中自动带出；敏感字段加密或仅内存。
- 规则导入导出、多配置集切换。
