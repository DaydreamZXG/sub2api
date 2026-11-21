# Sub2API（野菜）交接文档

写于 2026-11-21。给接手的 Devin 会话：先读完这份，再动手。

## 1. 用户要什么

- 目标：普通人一看就懂、界面好看，同时保留网关的高级能力（线路、模型、价格、缓存、日志、计费、利润、代理、分销）。
- 原话：“我想着是咱们应该改成普通人一看就懂就知道怎么用的，当然UI也要好看”；“后端还是别大改，咱们主要还是改UI的”；“小白可能就是自己手动接agent，不会用命令行”。
- 核心流程：注册 → 充值 → 建 API 密钥 → 复制地址和密钥 → 手动填进 Claude Code / Codex / Gemini CLI 等工具。
- 沟通：用中文，简短，说人话，不堆术语。

## 2. 工作规则（必须遵守）

- **每步一个 PR**，从 `main` 新开分支（`devin/$(date +%s)-<slug>`）。
- 做完一步发一次说明和截图（浅色、深色、手机宽度）。
- **都不合并，等用户说“合并”**。合并后发现的问题另开 PR 修，不直接改 `main`。
- 改动前先给方案；大的设计选择先问用户。

## 3. 代码

- 仓库：https://github.com/DaydreamZXG/sub2api ，`main` 最新是 `152b3a40`（PR #39 的合并），**没有未合并的 PR**。
- 后端 Go（`backend/`），前端 Vue 3 + TS（`frontend/`）。
- 已完成并合并（PR #8 起到 #39）：用户页面全部重做（首页、密钥、充值、套餐、邀请、状态、用量、渠道、个人中心、暗色模式）；后台“利润”（成本包、利润报表、只算真充值、按币种显示收款）；代理（按名下用户消费分润、提现审核、后台把已有用户挂到代理名下、分润按月消费分档、代理说明页）；分销（分销商身份、客户密钥加价、按成本从分销商余额扣、客户域名独立门户）；普通密钥查用量不再显示上游成本。
- 分销扣款：分销商余额扣实际成本；客户密钥额度扣“成本 × 加价倍数”（1–100，建好不能改）；客户密钥必须选线路；分销商余额不足时客户密钥停用。
- 客户域名：后台设置项 `reseller_portal`（domain / name / contact / logo），接口 `GET/PUT /api/v1/admin/reseller-portal`。在客户域名上，`/api/*`（除 `/api/v3/` 视频任务接口）、`/setup/*`、`/.well-known/*` 返回 404；`/v1/*`、`/v1beta/*`、客户页面可用。代码：`backend/internal/server/reseller_portal.go`、`backend/internal/service/reseller_service.go`、前端 `frontend/src/portal/`。

## 4. 服务器（线上）

- 服务器地址、SSH 端口和用户不写在这里（仓库是公开的），找用户要。用户是 `ubuntu`（root 不能登录），sudo 免密。
- **新号需要重新添加 SSH 私钥作为 secret**（旧号里叫 `ovh`），密钥不在仓库里。
- 项目目录 `/home/ubuntu/project/sub2api-public`（`compose.yaml`、`.env`、`DEPLOYMENT.txt` 记录了每次改动和撤销办法）。
- 容器：`sub2api-public`（镜像 `sub2api-yecai:152b3a40`，版本 0.2.23，端口 `127.0.0.1:18081`）、`sub2api-public-postgres`、`sub2api-public-redis`。
- 镜像在服务器上从源码构建，源码在 `/home/ubuntu/project/sub2api-yecai-src/152b3a40`。更新流程：备份 → 构建新镜像 → 用备份数据另起临时一套试跑 → 只改 `.env` 里的 `APP_IMAGE` → `docker compose up -d --wait --no-deps sub2api` → 检查。
- 部署前备份：`/opt/backups/sub2api-public-predeploy-20261120T165834Z`（数据库 dump、`.env`、数据目录、旧镜像名）。回滚步骤写在 `DEPLOYMENT.txt` 末尾。
- **禁止**：`docker compose down -v`、`docker prune`、重启 Docker、动旧的 `sub2api-deploy` 或服务器上其他服务。
- 管理员账号和密码在服务器 `.env` 的 `ADMIN_EMAIL` / `ADMIN_PASSWORD`。
- 站点设置里“开放注册”目前是关的。

## 5. 域名

- Cloudflare 里有两个账号的域名，`yeschoy.us` 和 `yeschoy.uk` 在同一个账号下。**新号需要重新连接 Cloudflare**。
- 主站：`https://yeschoy.us`。Cloudflare A 记录指向服务器，开代理。Nginx `/etc/nginx/conf.d/yeschoy-us.conf` → `127.0.0.1:18081`，保留原始 Host。证书 Let's Encrypt，certbot dns-cloudflare 自动续。
- 客户域名：`https://yeschoy.uk`。同样的做法（Nginx `yeschoy-uk.conf`）。该域名原有的收信 MX 和 OpenAI 验证 TXT 记录没动。
- `yeschoy.pro/.app/.com/.vip` 是服务器上其他服务在用的，不要动。Nginx 只 reload，不 restart。

## 6. 还没做完的

1. **用户要做**：登录 `https://yeschoy.us` 后台 → 确认“管理员合规声明”（必须本人确认，不能代点）→ “利润 → 分销商 → 客户域名”填 `yeschoy.uk` 和一个不同于“野菜”的名字，保存。截至交接时还没存（`https://yeschoy.uk/api/v1/settings/public` 仍返回 200，即还在显示主站）。
2. **存好后要检查**：`yeschoy.uk` 上 `/api/v1/settings/public`、`/setup/status`、`/.well-known/x` 都是 404；首页变成客户页面；`/v1/models` 不带密钥返回 401；`/api/v3/contents/generations/tasks` 可达；主站 `yeschoy.us` 不受影响。
3. 上线后用一把小额度客户密钥做一次真实调用，核对分销商扣成本、客户额度按加价倍数扣。站里还没接上游账号。
4. 后台利润页：如果用户靠转账后手动加余额收钱，要打开“后台手动加的余额算真充值”。
5. 可选后续功能（都要新 PR）：客户门户看单次调用记录、看每个模型的客户单价、每个分销商独立域名。
