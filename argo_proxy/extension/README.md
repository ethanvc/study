# Argo Proxy 扩展

Chrome 扩展代码与配置均在此目录。

**加载方式**

1. Chrome 打开 `chrome://extensions`，开启「开发者模式」。
2. 点击「加载已解压的扩展程序」。
3. **必须选择本目录**（即包含 `manifest.json` 的 `extension` 文件夹），不要选上级 `argo_proxy`。

**若出现 ERR_FILE_NOT_FOUND / “Your file couldn't be accessed”**

- 说明当前加载的根目录不对（例如之前选的是 `argo_proxy`）。
- 在 `chrome://extensions` 中**移除**该扩展，再重新「加载已解压的扩展程序」，这次**只选 `extension` 这一层目录**。
