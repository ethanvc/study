# 概念
## 代理
支持http、https、socks5代理。
代理都已一个名称，内部有一个数字id，用于引用代理。代理支持改名。

## Profile
包含一个规则列表、默认动作，默认动作可以是发给代理，BLOCK，或者DIRECT。

## 规则（Rule）
一条规则包含如下元素：
- 规则类型：
  - host contains
  - host regex
  - url contains
  - url regex
  - path contains
  - path regex
- 是否启用：用于临时禁用一条规则。
- 规则值：根据规则类型+规则值，判断请求是否命中规则。
- 动作：匹配之后的动作。

# 界面
## 代理列表
- 代理列表
- 新增代理
- 删除代理
- 编辑代理

## Profile界面
- 规则列表
- 新增规则
- 删除规则
- 编辑规则
- 展示默认动作
- 编辑默认动作
