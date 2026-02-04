# 原型说明

静态 HTML，用于体验 Popup 与设置页的交互和布局。

**体验方式**

1. **Popup**：用浏览器直接打开 [popup.html](popup.html)（建议窗口收窄到约 320px 宽以模拟扩展弹窗）。
2. **设置**：打开 [options.html](options.html)；或在 popup 中点击「设置」会在新标签页打开（需允许弹窗）。

**交互**

- **popup.html**：切换「启用代理路由」、点击「Host 列表」进入子菜单查看 host → 代理列表、子菜单顶部 ‹ 返回、点击「设置」打开 options。
- **options.html**：查看代理列表与路由规则的结构；添加/编辑/删除按钮为占位，无持久化。
