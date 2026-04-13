# macOS 透明代理实现原理 — mihomo 源码分析

## 1. TUN 设备创建

macOS 内核提供 utun (User Tunnel) 驱动。mihomo 通过以下系统调用序列创建 utun 设备。

### 1.1 创建控制 socket

```c
int tunFd = socket(AF_SYSTEM, SOCK_DGRAM, SYSPROTO_CONTROL/*2*/);
```

`AF_SYSTEM` 是 macOS 专有的地址族，用于和内核扩展/控制接口通信。`SYSPROTO_CONTROL` 表示 Kernel Control Socket 协议。

### 1.2 获取 utun 控制 ID

```c
struct ctl_info ctlInfo;
strcpy(ctlInfo.ctl_name, "com.apple.net.utun_control");
ioctl(tunFd, CTLIOCGINFO, &ctlInfo);  // 获取 ctlInfo.ctl_id
```

### 1.3 connect 创建 utun 设备

```c
struct sockaddr_ctl sc = {
    .sc_id   = ctlInfo.ctl_id,
    .sc_unit = ifIndex + 1,    // utun0 → unit=1, utun1 → unit=2
};
connect(tunFd, (struct sockaddr*)&sc, sizeof(sc));
```

`connect()` 不返回新 fd，而是将 `tunFd` 绑定到内核 utun 控制单元，内核侧随之创建 `utunN` 接口。之后对 `tunFd` 的 `read()/write()` 直接收发 IP 包。**后文所有 `tunFd` 均指此 fd。**

### 1.4 设置 MTU

utun 设备创建后默认 MTU 通常为 1500。mihomo 将其设为 9000，因为 utun 是纯虚拟设备，不受物理链路限制，更大的 MTU 意味着每个 IP 包能承载更多数据，减少用户态/内核态切换次数和包头开销，提升吞吐量。

`SIOCSIFMTU` 等网络管理 ioctl 要求 fd 类型为 `AF_INET`，`tunFd`（`AF_SYSTEM`）不满足，所以临时创建一个 `AF_INET` socket 做管理句柄（下面 1.5、1.6 同理）。`mgmtFd` 不收发数据，用完即 `close()`。

```c
int mgmtFd = socket(AF_INET, SOCK_DGRAM, 0);  // 临时管理句柄，非数据通道
struct ifreq_mtu ifr;
strcpy(ifr.ifr_name, "utun0");
ifr.ifr_mtu = 9000;
ioctl(mgmtFd, SIOCSIFMTU, &ifr);
close(mgmtFd);
```

### 1.5 配置 IPv4 地址（SIOCAIFADDR）

新创建的 utun 接口没有 IP 地址，必须分配一个才能：
1. **作为路由网关** — 2.1 步添加路由时，gateway 指向此地址（如 198.18.0.1），内核才知道把匹配的流量送进 utun
2. **System 栈需要监听** — 4.1 步在此地址上 `listen()`，接收 NAT 改写后的 TCP 连接
3. **DNS 劫持** — utun 地址的 next IP（198.18.0.2:53）被自动加入 DNS 劫持列表

地址选用 `198.18.0.0/15`（RFC 2544 基准测试保留段），不会与真实网络冲突。

```c
struct ifaliasreq {
    char          ifra_name[IFNAMSIZ];
    struct sockaddr_in ifra_addr;      // 接口地址 (如 198.18.0.1)
    struct sockaddr_in ifra_dstaddr;   // 点对点目标地址
    struct sockaddr_in ifra_mask;      // 子网掩码
};
ioctl(mgmtFd, SIOCAIFADDR, &ifReq);  // mgmtFd = socket(AF_INET, SOCK_DGRAM, 0)
```

### 1.6 配置 IPv6 地址（SIOCAIFADDR_IN6）

同理，为 utun 分配 IPv6 地址以支持 IPv6 流量的路由捕获和处理。

```c
struct in6_aliasreq {
    char              ifra_name[16];
    struct sockaddr_in6 ifra_addr;
    struct sockaddr_in6 ifra_dstaddr;
    struct sockaddr_in6 ifra_mask;
    uint32_t            ifra_flags;    // IN6_IFF_NODAD | IN6_IFF_SECURED
    struct addr_lifetime ifra_lifetime; // vltime = pltime = 0xFFFFFFFF (infinite)
};
ioctl(mgmtFd6, SIOCAIFADDR_IN6/*0x8080266A*/, &ifReq6);  // mgmtFd6 = socket(AF_INET6, SOCK_DGRAM, 0)
```

- `IN6_IFF_NODAD` — 跳过 Duplicate Address Detection，避免启动延迟（虚拟设备不存在地址冲突）
- `IN6_IFF_SECURED` — 标记为安全地址，防止被其他接口的隐私扩展覆盖

### 1.7 设置非阻塞 + 批量收包参数

```c
fcntl(tunFd, F_SETFL, O_NONBLOCK);
int batchSize = 57; // (512*1024 / MTU) + 1
setsockopt(tunFd, SYSPROTO_CONTROL/*2*/, UTUN_OPT_MAX_PENDING_PACKETS/*16*/, &batchSize, sizeof(int));
```

`UTUN_OPT_MAX_PENDING_PACKETS` 控制内核为 utun 缓存的最大待读包数，配合 `recvmsg_x` 批量收包使用。

---

## 2. 路由表操作

### 2.1 添加路由（AF_ROUTE socket）

mihomo 通过 `auto-route` 将流量导向 utun，实现方式是操作系统路由表：

```c
int routeFd = socket(AF_ROUTE, SOCK_RAW, 0);  // 路由操作句柄

struct rt_msghdr rtm = {
    .rtm_type    = RTM_ADD,
    .rtm_flags   = RTF_UP | RTF_STATIC | RTF_GATEWAY,
    .rtm_version = RTM_VERSION,
    .rtm_seq     = 1,
    .rtm_addrs   = RTA_DST | RTA_NETMASK | RTA_GATEWAY,
};
// Addrs[RTAX_DST]     = 目标网段 (如 0.0.0.0/1, 128.0.0.0/1)
// Addrs[RTAX_NETMASK] = 子网掩码
// Addrs[RTAX_GATEWAY] = utun 的 gateway 地址 (如 198.18.0.1)

write(routeFd, &routeMessage, len);
close(routeFd);
```

典型做法是添加 `0.0.0.0/1` + `128.0.0.0/1` 两条路由覆盖全部 IPv4 流量（避免覆盖默认路由 `0.0.0.0/0`）。

### 2.2 刷新 DNS 缓存

路由变更后执行：
```c
execl("/usr/bin/dscacheutil", "dscacheutil", "-flushcache", NULL);
```

---

## 3. 收发包的系统调用

utun 设备的数据格式：**4 字节 AF 头 + IP 包**。前 4 字节标识协议族（`AF_INET=2` 或 `AF_INET6=30`）。

### 3.1 批量收包 — recvmsg_x（Darwin 私有系统调用）

```c
struct msghdr_x {
    struct msghdr msg;
    uint32_t      datalen;
};
int n = syscall(SYS_RECVMSG_X, tunFd, msgHdrs, count, MSG_DONTWAIT, 0, 0);
```

`recvmsg_x` 是 macOS XNU 内核的私有批量收包接口，一次系统调用可读取多个消息。每个消息通过 `iovec` 散列读取：`iovec[0]` = 4 字节 AF 头，`iovec[1]` = IP 包内容。

### 3.2 批量发包 — sendmsg_x（可选）

```c
int n = syscall(SYS_SENDMSG_X, tunFd, msgHdrs, count, MSG_DONTWAIT, 0, 0);
```

默认关闭（`SendMsgX=false`），因为多线程下载时可能触发内核冻结。

### 3.3 单包写入 — writev

默认写入路径：
```c
struct iovec iov[2] = {
    { .iov_base = "\x00\x00\x00\x02", .iov_len = 4 },  // AF_INET header
    { .iov_base = ip_packet,           .iov_len = len },
};
writev(tunFd, iov, 2);
```

### 3.4 单包读取 — readv（非批量模式）

```c
struct iovec iov[2] = {
    { .iov_base = header_buf, .iov_len = 4 },
    { .iov_base = packet_buf, .iov_len = mtu },
};
readv(tunFd, iov, 2);
```

### 3.5 I/O 多路复用 — kqueue

等待 utun fd 可读时使用 `kqueue`（而非 `poll`/`epoll`）：

```c
int kq = kqueue();
struct kevent kevents[2] = {
    { .ident = stopFd, .filter = EVFILT_READ, .flags = EV_ADD|EV_ENABLE },
    { .ident = tunFd,  .filter = EVFILT_READ, .flags = EV_ADD|EV_ENABLE },
};
kevent(kq, kevents, 2, NULL, 0, NULL);  // 注册
kevent(kq, NULL, 0, revents, 2, NULL);  // 阻塞等待
```

`stopFd` 是一对 `pipe()` fd，用于优雅停止：向 write 端写 1 字节，kqueue 立即返回。

---

## 4. IP 栈处理

从 utun 读出的是原始 IP 包，需要解析协议后分发给代理隧道。

### 4.1 System 栈（默认）— 用户态 NAT 转发到内核 TCP

核心思路：**在用户态解析 IP 包头，做 NAT 改写，然后回写 utun 让内核 TCP/IP 栈处理**。

```
应用 → utun（IP 包）→ 用户态解析 → NAT 改写 IP/TCP 头 → 回写 utun → 内核 TCP → accept()
```

具体流程（TCP）：

1. **启动时监听 TCP**：

```c
// 在 utun 地址上监听一个随机端口
int listenFd = socket(AF_INET, SOCK_STREAM, 0);  // System 栈的 TCP 监听
struct sockaddr_in addr = { .sin_addr.s_addr = inet_addr("198.18.0.1"), .sin_port = 0 };
bind(listenFd, (struct sockaddr*)&addr, sizeof(addr));
listen(listenFd, SOMAXCONN);
// getsockname 获取内核分配的 tcpPort
```

2. **读 IP 包**：从 utun 读出 `[4B AF头][IP头][TCP头][payload]`

3. **NAT 改写**（直接修改 IP 包内存）：

```c
// 原始: src=10.0.0.5:12345 → dst=93.184.216.34:443
// 改写为:
struct iphdr  *ip  = packet;
struct tcphdr *tcp = packet + ip->ihl * 4;

// 记录 NAT 映射: natPort → {原始src, 原始dst}
uint16_t natPort = alloc_nat_port();
nat_table[natPort] = (Session){ .src = ip->saddr, .sport = tcp->source,
                                .dst = ip->daddr, .dport = tcp->dest };

ip->saddr    = inet_addr("198.18.0.2");   // utun 的 next 地址
tcp->source  = htons(natPort);
ip->daddr    = inet_addr("198.18.0.1");   // utun 自身地址
tcp->dest    = htons(tcpPort);             // 上面 listen 的端口

// 重算 checksum
tcp->check = 0;
tcp->check = tcp_checksum(ip, tcp);
ip->check  = 0;
ip->check  = ip_checksum(ip);
```

4. **回写 utun**：内核看到目标是本机 198.18.0.1:tcpPort，走正常 TCP 栈

```c
struct iovec iov[2] = {
    { "\x00\x00\x00\x02", 4 },  // AF_INET
    { packet, packet_len },
};
writev(tunFd, iov, 2);
```

5. **accept 拿到连接**：通过 `getpeername` 获取 natPort，查 NAT 表恢复原始目标

```c
int connFd = accept(listenFd, (struct sockaddr*)&peer, &len);
uint16_t natPort = ntohs(peer.sin_port);
Session *sess = &nat_table[natPort];
// sess->dst + sess->dport = 原始目标 93.184.216.34:443
// 交给代理隧道处理
```

UDP 直接在用户态解析，不经过内核 TCP 栈。

### 4.2 gVisor 栈 — 完全用户态 TCP/IP

使用 Google gVisor 的用户态 TCP/IP 栈。通过 `fdbased` endpoint 直接从 `tunFd` 用 `recvmsg_x` / `writev` 收发原始 IP 帧，在用户空间完成 TCP 三次握手、拥塞控制等全部逻辑，不依赖内核 TCP 栈。

数据路径对比 System 栈少一次 utun 往返：

```
System栈: 应用 → utun → 用户态NAT改写 → 回写utun → 内核TCP → accept() → 代理
gVisor栈: 应用 → utun → gVisor用户态TCP  → 直接得到连接 → 代理
```

### 4.3 Mixed 栈

TCP 走 System，UDP 走 gVisor。兼顾 TCP 性能和 UDP 处理简洁性。

### 4.4 三种栈对比

| | System | gVisor | Mixed |
|---|--------|--------|-------|
| **TCP 实现** | 内核 TCP 栈 | 用户态 gVisor TCP | 内核 TCP 栈 |
| **UDP 实现** | 用户态直接解析 | 用户态 gVisor UDP | 用户态 gVisor UDP |
| **TCP 数据路径** | utun → 用户态 → utun → 内核 → accept（**两次穿越 utun**） | utun → gVisor（**一次穿越**） | 同 System |
| **是否需要 NAT** | TCP 需要（改写 IP/TCP 头 + 维护映射表 + 重算 checksum） | 不需要 | TCP 需要 |
| **TCP 性能** | 好。内核 TCP 经过多年优化，拥塞控制、快速重传等成熟 | 略差。用户态实现开销更大，拥塞算法不如内核完善 | 同 System |
| **UDP 性能** | 好。无 NAT，直接解析 | 好 | 同 gVisor |
| **兼容性** | 高。不需要额外 build tag | 需要 `-tags with_gvisor` 编译，二进制体积增大约 10MB | 同 gVisor |
| **内存占用** | 低 | 较高。gVisor 栈需要维护自己的连接状态、收发缓冲区 | 中等 |
| **复杂度** | NAT 映射表管理复杂，需处理端口耗尽、超时回收 | 架构简洁，utun 读出即交给 gVisor 处理 | 两套逻辑并存 |
| **适用场景** | 追求 TCP 吞吐量（大文件下载、视频流） | 追求低延迟、大量短连接 | 默认推荐，综合最优 |

---

## 5. 出站接口绑定 — 防止回环

代理出站流量必须走物理网卡（如 en0），不能再进 utun。macOS 使用 `IP_BOUND_IF`：

```c
int ifIndex = if_nametoindex("en0");
setsockopt(outFd, IPPROTO_IP, IP_BOUND_IF, &ifIndex, sizeof(ifIndex));
// IPv6:
setsockopt(outFd, IPPROTO_IPV6, IPV6_BOUND_IF, &ifIndex, sizeof(ifIndex));
```

`outFd` 是代理出站连接的 socket（连接远程代理服务器的那个）。

这是 macOS 特有的 socket 选项（Linux 用 `SO_BINDTODEVICE`），将 socket 绑定到指定接口索引。

---

## 6. 网络变化监控 — AF_ROUTE socket

```c
int monitorFd = socket(AF_ROUTE, SOCK_RAW, 0);  // 路由监控句柄（常驻）
fcntl(monitorFd, F_SETFL, O_NONBLOCK);
```

持续 `read(monitorFd)` 监听路由表变化。收到 `RTM_*` 消息时解析，若是路由变更事件则触发回调（刷新接口缓存、重置 DNS 连接）。

### 6.1 检测默认网关接口

```c
int mib[] = { CTL_NET, PF_ROUTE, 0, AF_UNSPEC, NET_RT_DUMP, 0 };
size_t len;
sysctl(mib, 6, NULL, &len, NULL, 0);
char *buf = malloc(len);
sysctl(mib, 6, buf, &len, NULL, 0);
// 解析 route messages，找 dst=0.0.0.0 mask=0.0.0.0 且 RTF_UP|RTF_GATEWAY 的条目
// 该条目的 rtm_index 即为默认出口接口
```

### 6.2 Network Extension 模式下的接口检测

作为 VPN 扩展运行时无法直接读路由表，改用 `connect()` 探测：

```c
int probeFd = socket(AF_INET, SOCK_STREAM, 0);  // 临时探测句柄
struct sockaddr_in target = {
    .sin_family = AF_INET,
    .sin_addr.s_addr = inet_addr("10.255.255.255"),
    .sin_port = htons(80),
};
connect(probeFd, (struct sockaddr*)&target, sizeof(target)); // 异步，不会真正建连

struct sockaddr_in local;
socklen_t len = sizeof(local);
getsockname(probeFd, (struct sockaddr*)&local, &len);  // 拿到内核选择的出口 IP
// 通过出口 IP 反查接口
```

---

## 7. Redir 模式 — PF NAT 查表

用 PF 防火墙做 TCP 重定向时，mihomo 通过 `DIOCNATLOOK` ioctl 恢复原始目标地址。

### 7.1 PF rdr 规则（用户手动配置）

```
# /etc/pf.conf
rdr pass on lo0 proto tcp from any to any -> 127.0.0.1 port 7892
```

### 7.2 DIOCNATLOOK 查询

```c
int pfFd = open("/dev/pf", O_RDONLY);  // PF 防火墙句柄

struct pfioc_natlook nl = {
    .af        = AF_INET,
    .proto     = IPPROTO_TCP,
    .direction = PF_OUT,
    .saddr     = client_ip,       // 客户端 IP
    .sport     = client_port,     // 客户端端口
    .daddr     = local_ip,        // 本地监听 IP（被 rdr 改写后的目标）
    .dport     = local_port,      // redir-port
};
ioctl(pfFd, DIOCNATLOOK, &nl);
close(pfFd);
// nl.rdaddr = 原始目标 IP
// nl.rdport = 原始目标端口
```

`DIOCNATLOOK` 的值计算：`IOC_INOUT | ((sizeof(pfioc_natlook) & 0x1FFF) << 16) | ('D' << 8) | 23`。

---

## 8. 进程识别 — sysctl pcblist

```c
size_t len;
sysctlbyname("net.inet.tcp.pcblist_n", NULL, &len, NULL, 0);
char *buf = malloc(len);
sysctlbyname("net.inet.tcp.pcblist_n", buf, &len, NULL, 0);
// UDP: "net.inet.udp.pcblist_n"
```

返回内核 PCB (Protocol Control Block) 列表的二进制 dump。每个条目包含 `xinpcb_n` + `xsocket_n` 结构体：

```
偏移 inp+18..20: 源端口 (big-endian)
偏移 inp+44:     inp_vflag (0x1=IPv4, 0x2=IPv6)
偏移 inp+64..80: IPv6 源地址 / 偏移 inp+76..80: IPv4 源地址
偏移 so+68..72:  so_last_pid (最近使用该 socket 的进程 PID)
```

匹配到 PID 后，通过 `proc_info` 系统调用获取进程路径：
```c
char path[1024];
syscall(SYS_PROC_INFO,
    PROC_CALLNUM_PIDINFO/*2*/,
    pid,
    PROC_PIDPATHINFO/*11*/,
    0,
    path,
    sizeof(path)/*1024*/);
```

注意：`xinpcb_n` 结构体大小依赖 macOS 版本。`kern.osrelease` 主版本号 >= 22（macOS 13 Ventura）时为 408 字节，更早版本为 384 字节。

---

## 9. 完整调用链总结

```
┌──── 应用发起连接 (如 curl https://example.com) ────┐
│                                                     │
│  内核路由表:                                        │
│    0.0.0.0/1   → utun gateway                       │
│    128.0.0.0/1 → utun gateway                       │
│  两条 /1 覆盖全部 IPv4 地址空间，                     │
│  但不替换原有 default route (0.0.0.0/0)，             │
│  因最长前缀匹配优先命中 /1；                          │
│  TUN 关闭时删除这两条即可恢复原路由。                  │
│  (RTM_ADD via AF_ROUTE socket)                      │
│                                                     │
▼                                                     │
utun 设备 (由 AF_SYSTEM + connect 创建)               │
│                                                     │
│  读取: recvmsg_x(tunFd, msgHdrs[], count)           │
│  数据: [4B AF头][IP头][TCP/UDP头][payload]            │
│                                                     │
▼                                                     │
IP 栈处理                                              │
├─ System栈: 解析IP/TCP头 → NAT改写 → writev回utun     │
│  → 内核TCP → accept() → 恢复原始目标                  │
├─ gVisor栈: fdbased endpoint → 用户态TCP/IP           │
│  → 直接得到 TCP conn / UDP packet                    │
└─ Mixed栈: TCP走System, UDP走gVisor                   │
                                                      │
▼                                                      │
代理隧道 (HandleTCPConn / HandleUDPPacket)              │
├─ 进程识别: sysctl(pcblist_n) → SYS_PROC_INFO(PID)    │
├─ DNS劫持: 匹配目标端口53 → 转发给内置DNS引擎           │
└─ 规则匹配 → 选择出站代理                               │
                                                      │
▼                                                      │
outFd (出站 socket)                                     │
│  setsockopt(outFd, IPPROTO_IP, IP_BOUND_IF, en0_idx)│
│  确保流量走物理网卡，不回流 utun                        │
│                                                     │
▼                                                     │
物理网卡 (en0) → 远程代理服务器 ───────────────────────┘
```

## 参考资料
- [mihomo Meta 分支](https://github.com/MetaCubeX/mihomo/tree/Meta)
- [sing-tun](https://github.com/metacubex/sing-tun) — 实际 TUN 设备和路由管理库
- [XNU utun 实现](https://github.com/apple-oss-distributions/xnu/blob/main/bsd/net/if_utun.c)
