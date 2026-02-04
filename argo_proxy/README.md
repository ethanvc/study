# Argo Proxy

按域名/路径将请求路由到不同代理的 Chrome 扩展（React + Vite + Zustand + Tailwind）。

**技术栈**：React 18、Vite 5、Zustand、Tailwind CSS、TypeScript。

**脚本**

- `npm install`：安装依赖。
- `npm run build`：构建扩展到 **dist/**（会从 static/ 复制 manifest、background、icons）。
- `npm run dev`：监听源码变更并持续构建。

**加载扩展**：Chrome → `chrome://extensions` → 开发者模式 → 加载已解压的扩展程序 → 选择 **dist** 目录。

**目录说明**

- **src/**：React 源码（popup、settings、shared 类型与 store）。
- **dist/**：构建产物，为 Chrome 可加载的扩展目录（运行 `npm run build` 后生成）。
- **static/**：扩展静态资源（manifest、background、icons），构建时复制到 dist。
