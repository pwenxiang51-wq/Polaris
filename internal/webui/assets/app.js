(() => {
  "use strict";

  const translations = {
    en: {
      privateConsole: "PRIVATE INFRASTRUCTURE CONSOLE",
      loginTitle: "Your servers.<br>One clear view.",
      loginSubtitle: "A lightweight monitoring station, built for the small machines that keep your world running.",
      reportEvery: "REPORT INTERVAL", smallFootprint: "SMALL-HOST FRIENDLY", secureByDefault: "SECURE BY DEFAULT",
      welcomeBack: "Welcome back", loginCardSubtitle: "Sign in to your private console", username: "Username", password: "Password",
      showPassword: "Show", hidePassword: "Hide", humanVerification: "Complete the human verification to sign in.", signIn: "Sign in",
      sessionNote: "Protected session · HttpOnly cookie · Turnstile verification", footerPrivate: "Private by design. Yours to operate.",
      liveMonitor: "LIVE MONITOR", privateInfrastructure: "Private infrastructure", signOut: "Sign out", overviewEyebrow: "INFRASTRUCTURE / OVERVIEW",
      dashboardTitle: "Everything, in orbit.", dashboardSubtitle: "A live view of your nodes, resource pressure and network health.", addProbe: "Add probe",
      totalNodes: "TOTAL NODES", registeredProbes: "Registered probes", onlineNodes: "ONLINE NODES", reportingNow: "Reporting right now",
      averageCpu: "AVERAGE CPU", acrossOnline: "Across online nodes", networkThroughput: "NETWORK THROUGHPUT", combinedTraffic: "Combined in + out",
      yourNodes: "Your nodes", nodeSubtitle: "Host-level metrics from your installed probes.", refreshing: "Refreshes every 10s", noNodesTitle: "Your orbit is quiet.",
      noNodesText: "Create a probe token, install the lightweight agent on a Linux host, and your first node will appear here.", addFirstProbe: "Add your first probe",
      footerMonitor: "SERVER MONITOR", footerLocal: "Metrics collected by your own probes", newProbeKicker: "NEW MONITORING AGENT", createProbeTitle: "Bring a node into orbit.",
      createProbeText: "Create a one-time token for a Linux probe. The token is shown only once.", nodeDisplayName: "Display name", nodeIdentifier: "Node identifier",
      nodeIdHint: "Use letters, numbers, dots, underscores or hyphens.", generateToken: "Generate one-time token", tokenCreated: "TOKEN CREATED", saveTokenNow: "Save this token now.",
      tokenWarning: "Polaris stores only a token hash. This raw token will not be shown again.", agentToken: "Agent token", copy: "Copy", agentConfig: "Agent environment",
      copyConfig: "Copy agent configuration", agentNextStep: "Install the polaris-agent binary and place these values in a root-readable environment file.", done: "Done",
      online: "ONLINE", offline: "OFFLINE", cpu: "CPU", memory: "MEMORY", disk: "DISK", load: "LOAD", ingress: "IN", egress: "OUT",
      received: "Received", sent: "Sent", publicIPv4: "PUBLIC IPV4", publicIPv6: "PUBLIC IPV6", operatingSystem: "OPERATING SYSTEM", kernel: "KERNEL",
      uptime: "UPTIME", processor: "PROCESSOR", location: "LOCATION", network: "NETWORK / ASN", lastSeen: "LAST REPORT", latency: "NETWORK LATENCY",
      noPublicIP: "Not detected", unknown: "Unknown", firstReportWaiting: "Waiting for the first report", deleteNode: "Revoke token and remove node",
      deleteConfirm: "Remove this node and revoke its probe token? This cannot be undone.", tokenCreatedToast: "Probe token created. Copy it before closing this window.",
      copySuccess: "Copied to clipboard.", copyFailed: "Clipboard access failed. Select and copy the text manually.", nodeRemoved: "Node and probe token removed.",
      loginFailed: "Sign-in failed. Check your credentials and verification.", verificationNotConfigured: "Turnstile is not configured. Ask the administrator to finish setup.",
      loading: "Loading…", requestFailed: "The request failed. Try again.", authExpired: "Your session expired. Sign in again.", lastUpdated: "Updated", noLatency: "No probe data",
      country: "COUNTRY", hostName: "HOSTNAME", cores: "cores", noMetrics: "Waiting for metrics", passwordRequired: "Please enter your username and password.",
      nodeNameRequired: "Enter a display name and a valid node identifier.", createTokenFailed: "Could not create the probe token.", refreshFailed: "Could not refresh node data.", themeLight: "Light mode", themeDark: "Dark mode",
      languageChinese: "中文", languageEnglish: "English", reportHistory: "LAST 24 HOURS", cpuTrend: "CPU TREND", memoryTrend: "MEMORY TREND", timeout: "Timeout",
      copyToken: "Copy token", probeTarget: "TARGET", latencyMs: "LATENCY", packetLoss: "LOSS", trafficTotal: "TOTAL TRAFFIC", swap: "SWAP",
      timezoneNote: "Host uptime", failedNodeLoad: "Node data is temporarily unavailable.", tokenPrompt: "The token is only displayed once. Rotate it if you lose it.",
      history24h: "LAST 24 HOURS", trafficInTrend: "INBOUND RATE · 24H", trafficOutTrend: "OUTBOUND RATE · 24H", latencyTrend: "TCP CONNECT · 24H",
      demoData: "DEMO DATA", telecom: "China Telecom", unicom: "China Unicom", mobile: "China Mobile", manageNode: "Node details", nodeMetadataTitle: "Node details and allowances", nodeMetadataText: "Optional fields are hidden until set. Identity overrides are used when automatic detection is unavailable.",
      expiryDate: "Expiry date", priceUSD: "Price (USD)", monthlyQuota: "Monthly traffic quota (GiB)", totalQuota: "Total traffic quota (GiB)",
      identityOverrides: "Identity overrides", countryName: "Country / region", countryCode: "Country code", regionName: "Region / state", cityName: "City", ispName: "ISP", asnName: "ASN",
      saveMetadata: "Save details", metadataSaved: "Node details saved.", expiryIn: "Expires in", expired: "Expired", days: "days", monthlyUsage: "Monthly traffic", totalUsage: "Tracked traffic", quotaNear: "Quota nearly reached", noQuota: "No quota configured",
    },
    zh: {
      privateConsole: "私有基础设施控制台", loginTitle: "你的服务器。<br>一屏尽在掌握。", loginSubtitle: "为轻量 VPS 打造的监控面板，让每台小机器的运行状态一目了然。",
      reportEvery: "上报间隔", smallFootprint: "适合小内存主机", secureByDefault: "默认安全设计", welcomeBack: "欢迎回来", loginCardSubtitle: "登录你的私有控制台",
      username: "账户名", password: "密码", showPassword: "显示", hidePassword: "隐藏", humanVerification: "请完成人机验证后登录。", signIn: "登录控制台",
      sessionNote: "安全会话 · HttpOnly Cookie · Cloudflare 人机验证", footerPrivate: "私有部署，完全由你掌控。", liveMonitor: "实时监控", privateInfrastructure: "私有基础设施",
      signOut: "退出登录", overviewEyebrow: "基础设施 / 总览", dashboardTitle: "让一切尽在掌握。", dashboardSubtitle: "实时查看探针、资源占用和网络健康状态。", addProbe: "添加探针",
      totalNodes: "探针总数", registeredProbes: "已注册探针", onlineNodes: "在线节点", reportingNow: "正在实时上报", averageCpu: "平均 CPU", acrossOnline: "仅统计在线节点",
      networkThroughput: "网络吞吐", combinedTraffic: "上行与下行合计", yourNodes: "你的节点", nodeSubtitle: "由节点探针采集的主机级运行数据。", refreshing: "每 10 秒刷新",
      noNodesTitle: "当前轨道还没有节点。", noNodesText: "创建探针令牌，在 Linux 主机上安装轻量 Agent，节点数据就会显示在这里。", addFirstProbe: "添加第一个探针",
      footerMonitor: "服务器监控", footerLocal: "指标由你自己的探针采集", newProbeKicker: "新建监控探针", createProbeTitle: "让一台主机加入轨道。", createProbeText: "为 Linux 探针生成一次性令牌。原始令牌只显示一次。",
      nodeDisplayName: "显示名称", nodeIdentifier: "节点标识", nodeIdHint: "仅使用英文字母、数字、点、下划线或连字符。", generateToken: "生成一次性令牌", tokenCreated: "令牌已创建",
      saveTokenNow: "请立即保存令牌。", tokenWarning: "Polaris 只保存令牌哈希，不会再次显示这段原始令牌。", agentToken: "探针令牌", copy: "复制", agentConfig: "探针环境变量",
      copyConfig: "复制探针配置", agentNextStep: "安装 polaris-agent，并将这些配置保存到仅管理员可读的环境文件中。", done: "完成", online: "在线", offline: "离线", cpu: "CPU", memory: "内存", disk: "磁盘", load: "负载",
      ingress: "下行", egress: "上行", received: "接收流量", sent: "发送流量", publicIPv4: "公网 IPv4", publicIPv6: "公网 IPv6", operatingSystem: "操作系统", kernel: "内核版本", uptime: "运行时间",
      processor: "处理器", location: "地理位置", network: "网络 / ASN", lastSeen: "最近上报", latency: "网络延迟", noPublicIP: "未检测到", unknown: "未知", firstReportWaiting: "等待首次上报",
      deleteNode: "撤销令牌并删除节点", deleteConfirm: "确定删除该节点并撤销其探针令牌？此操作无法撤销。", tokenCreatedToast: "探针令牌已创建，请在关闭窗口前复制保存。", copySuccess: "已复制到剪贴板。",
      copyFailed: "无法访问剪贴板，请手动选择并复制文本。", nodeRemoved: "节点和探针令牌已删除。", loginFailed: "登录失败，请检查账户、密码和人机验证。", verificationNotConfigured: "尚未配置 Turnstile，请联系管理员完成配置。",
      loading: "加载中…", requestFailed: "请求失败，请重试。", authExpired: "登录状态已过期，请重新登录。", lastUpdated: "更新于", noLatency: "暂无探测数据", country: "国家或地区", hostName: "主机名", cores: "核", noMetrics: "等待监控数据",
      passwordRequired: "请输入账户名和密码。", nodeNameRequired: "请输入显示名称和有效的节点标识。", createTokenFailed: "无法创建探针令牌。", refreshFailed: "无法刷新节点数据。", themeLight: "浅色主题", themeDark: "暗夜主题",
      languageChinese: "English", languageEnglish: "中文", reportHistory: "最近 24 小时", cpuTrend: "CPU 趋势", memoryTrend: "内存趋势", timeout: "超时", copyToken: "复制令牌", probeTarget: "探测目标", latencyMs: "延迟", packetLoss: "丢包率", trafficTotal: "累计流量", swap: "交换分区",
      timezoneNote: "主机运行时间", failedNodeLoad: "暂时无法读取节点数据。", tokenPrompt: "令牌只展示一次，遗失后请重新生成以轮换令牌。",
      history24h: "最近 24 小时", trafficInTrend: "下行速率 · 24 小时", trafficOutTrend: "上行速率 · 24 小时", latencyTrend: "TCP 建连 · 24 小时",
      demoData: "演示数据", telecom: "中国电信", unicom: "中国联通", mobile: "中国移动", manageNode: "节点详情", nodeMetadataTitle: "节点详情与套餐信息", nodeMetadataText: "可选字段留空时不显示。自动识别不可用时，可以填写身份覆盖信息。",
      expiryDate: "到期日期", priceUSD: "价格（美元）", monthlyQuota: "月流量额度（GiB）", totalQuota: "总流量额度（GiB）",
      identityOverrides: "手动身份覆盖", countryName: "国家或地区", countryCode: "国家代码", regionName: "州 / 省 / 地区", cityName: "城市", ispName: "运营商 ISP", asnName: "ASN",
      saveMetadata: "保存详情", metadataSaved: "节点详情已保存。", expiryIn: "剩余", expired: "已过期", days: "天", monthlyUsage: "本月流量", totalUsage: "累计统计流量", quotaNear: "流量额度即将用尽", noQuota: "未设置流量额度",
    },
  };

  const state = {
    language: safeStorageGet("polaris-language") || "en",
    theme: safeStorageGet("polaris-theme") || "dark",
    config: null,
    nodes: [],
    historyByNode: {},
    historyFetchedAt: {},
    historyLoading: {},
    editingNodeID: "",
    turnstileToken: "",
    turnstileWidgetId: null,
    nodesTimer: null,
    loadingNodes: false,
    createdToken: "",
    toastTimer: null,
  };

  const $ = (selector) => document.querySelector(selector);
  const $$ = (selector) => Array.from(document.querySelectorAll(selector));
  const sampleAcross = window.PolarisChartUtils.sampleAcross;
  const translate = (key) => (translations[state.language] && translations[state.language][key]) || translations.en[key] || key;

  function safeStorageGet(key) {
    try { return window.localStorage.getItem(key); } catch (_) { return null; }
  }

  function safeStorageSet(key, value) {
    try { window.localStorage.setItem(key, value); } catch (_) { /* Storage may be disabled by the browser. */ }
  }

  function applyLanguage() {
    document.documentElement.lang = state.language === "zh" ? "zh-CN" : "en";
    $$('[data-i18n]').forEach((element) => {
      const key = element.getAttribute("data-i18n");
      if (key in translations[state.language]) element.innerHTML = translate(key);
    });
    const languageLabel = state.language === "zh" ? translate("languageChinese") : translate("languageEnglish");
    $("#languageLabel").textContent = languageLabel;
    $("#dashLanguageLabel").textContent = languageLabel;
    updateThemeLabels();
    renderNodes();
  }

  function applyTheme() {
    document.documentElement.dataset.theme = state.theme;
    safeStorageSet("polaris-theme", state.theme);
    updateThemeLabels();
  }

  function updateThemeLabels() {
    const nextTheme = state.theme === "dark" ? "themeLight" : "themeDark";
    const glyph = state.theme === "dark" ? "☼" : "☾";
    ["themeLabel", "dashThemeLabel"].forEach((id) => { const el = document.getElementById(id); if (el) el.textContent = translate(nextTheme); });
    ["themeGlyph", "dashThemeGlyph"].forEach((id) => { const el = document.getElementById(id); if (el) el.textContent = glyph; });
  }

  function toggleLanguage() {
    state.language = state.language === "en" ? "zh" : "en";
    safeStorageSet("polaris-language", state.language);
    applyLanguage();
  }

  function toggleTheme() {
    state.theme = state.theme === "dark" ? "light" : "dark";
    applyTheme();
  }

  async function request(path, options = {}) {
    const headers = new Headers(options.headers || {});
    if (options.body !== undefined && !headers.has("Content-Type")) headers.set("Content-Type", "application/json");
    if (options.mutation === true) headers.set("X-Polaris-Request", "1");
    const response = await fetch(path, {
      method: options.method || "GET",
      headers,
      body: options.body,
      credentials: "same-origin",
      cache: "no-store",
      redirect: "error",
    });
    if (response.status === 204) return null;
    const payload = await response.json().catch(() => ({}));
    if (!response.ok) {
      const error = new Error(payload.error || translate("requestFailed"));
      error.status = response.status;
      throw error;
    }
    return payload;
  }

  function showView(view) {
    $("#authView").hidden = view !== "auth";
    $("#dashboardView").hidden = view !== "dashboard";
    if (view === "dashboard") startRefresh(); else stopRefresh();
  }

  function showLoginError(message) {
    const el = $("#loginError");
    el.textContent = message;
    el.hidden = false;
  }

  function clearLoginError() {
    $("#loginError").hidden = true;
    $("#loginError").textContent = "";
  }

  function setButtonBusy(button, busy, labelKey) {
    button.disabled = busy;
    if (busy) {
      button.dataset.previousText = button.innerHTML;
      button.textContent = translate(labelKey || "loading");
    } else if (button.dataset.previousText) {
      button.innerHTML = button.dataset.previousText;
      delete button.dataset.previousText;
    }
  }

  function loadTurnstile(siteKey) {
    if (!siteKey) {
      $("#turnstileHint").textContent = translate("verificationNotConfigured");
      $("#loginSubmit").disabled = true;
      return;
    }
    const script = document.createElement("script");
    script.src = "https://challenges.cloudflare.com/turnstile/v0/api.js?render=explicit";
    script.async = true;
    script.defer = true;
    script.onload = () => {
      if (!window.turnstile) return;
      state.turnstileWidgetId = window.turnstile.render("#turnstileWidget", {
        sitekey: siteKey,
        action: "login",
        theme: state.theme,
        callback: (token) => { state.turnstileToken = token; clearLoginError(); },
        "expired-callback": () => { state.turnstileToken = ""; },
        "error-callback": () => { state.turnstileToken = ""; },
      });
    };
    script.onerror = () => { $("#turnstileHint").textContent = translate("humanVerification"); };
    document.head.appendChild(script);
  }

  async function initialize() {
    applyTheme();
    applyLanguage();
    try {
      state.config = await request("/api/public-config");
      loadTurnstile(state.config.turnstile_sitekey);
      try {
        await request("/api/me");
        showView("dashboard");
        await loadNodes();
      } catch (error) {
        showView("auth");
      }
    } catch (_) {
      showView("auth");
      showLoginError(translate("requestFailed"));
    }
  }

  async function handleLogin(event) {
    event.preventDefault();
    clearLoginError();
    const username = $("#username").value.trim();
    const password = $("#password").value;
    if (!username || !password) {
      showLoginError(translate("passwordRequired"));
      return;
    }
    if (!state.turnstileToken) {
      showLoginError(translate("humanVerification"));
      return;
    }
    const button = $("#loginSubmit");
    setButtonBusy(button, true);
    try {
      await request("/api/login", {
        method: "POST", mutation: true,
        body: JSON.stringify({ username, password, turnstile_token: state.turnstileToken }),
      });
      state.turnstileToken = "";
      $("#password").value = "";
      showView("dashboard");
      await loadNodes();
    } catch (error) {
      showLoginError(error.message || translate("loginFailed"));
      if (window.turnstile && state.turnstileWidgetId !== null) window.turnstile.reset(state.turnstileWidgetId);
      state.turnstileToken = "";
    } finally {
      setButtonBusy(button, false, "signIn");
    }
  }

  async function logout() {
    try {
      await request("/api/logout", { method: "POST", mutation: true, body: "{}" });
    } catch (_) { /* Expired sessions are cleared by the next login. */ }
    state.nodes = [];
    state.historyByNode = {};
    state.historyFetchedAt = {};
    stopRefresh();
    showView("auth");
    clearLoginError();
  }

  async function loadNodes() {
    if (state.loadingNodes) return;
    state.loadingNodes = true;
    try {
      const payload = await request("/api/nodes");
      state.nodes = Array.isArray(payload.nodes) ? payload.nodes : [];
      const validIDs = new Set(state.nodes.map((node) => node.node_id));
      Object.keys(state.historyByNode).forEach((id) => { if (!validIDs.has(id)) { delete state.historyByNode[id]; delete state.historyFetchedAt[id]; } });
      renderNodes();
      await loadHistoriesIfStale();
      $("#lastRefreshLabel").textContent = `${translate("lastUpdated")} ${new Date().toLocaleTimeString()}`;
      $("#globalMessage").hidden = true;
    } catch (error) {
      if (error.status === 401) {
        showView("auth");
        showLoginError(translate("authExpired"));
      } else {
        const message = $("#globalMessage");
        message.textContent = translate("refreshFailed");
        message.hidden = false;
      }
    } finally {
      state.loadingNodes = false;
    }
  }

  async function loadHistoriesIfStale(force = false) {
    const due = state.nodes.filter((node) => {
      const id = node.node_id;
      return id && !state.historyLoading[id] && (force || !state.historyFetchedAt[id] || Date.now() - state.historyFetchedAt[id] > 60000);
    });
    await Promise.all(due.map(async (node) => {
      const id = node.node_id;
      state.historyLoading[id] = true;
      try {
        const payload = await request(`/api/nodes/${encodeURIComponent(id)}/history`);
        state.historyByNode[id] = Array.isArray(payload.history) ? payload.history : [];
        state.historyFetchedAt[id] = Date.now();
      } catch (error) {
        if (error.status === 401) throw error;
      } finally {
        delete state.historyLoading[id];
      }
    }));
    renderNodes();
  }

  function startRefresh() {
    stopRefresh();
    state.nodesTimer = window.setInterval(loadNodes, 10000);
  }

  function stopRefresh() {
    if (state.nodesTimer) window.clearInterval(state.nodesTimer);
    state.nodesTimer = null;
  }

  function renderNodes() {
    const nodesGrid = $("#nodesGrid");
    if (!nodesGrid) return;
    const now = Date.now();
    const online = state.nodes.filter((node) => isOnline(node, now));
    $("#summaryTotal").textContent = String(state.nodes.length);
    $("#summaryOnline").textContent = String(online.length);
    const averageCPU = online.length ? online.reduce((sum, node) => sum + (Number(node.cpu_percent) || 0), 0) / online.length : 0;
    $("#summaryCPU").innerHTML = `${averageCPU.toFixed(1)}<small>%</small>`;
    const traffic = online.reduce((sum, node) => sum + (Number(node.rx_rate) || 0) + (Number(node.tx_rate) || 0), 0);
    $("#summaryTraffic").textContent = fmtRate(traffic);
    $("#emptyState").hidden = state.nodes.length > 0;
    nodesGrid.hidden = state.nodes.length === 0;
    nodesGrid.innerHTML = state.nodes.map((node) => renderNode({ ...node, history: state.historyByNode[node.node_id] || [] }, now)).join("");
    nodesGrid.querySelectorAll("[data-remove-node]").forEach((button) => {
      button.addEventListener("click", () => removeNode(button.dataset.removeNode));
    });
    nodesGrid.querySelectorAll("[data-edit-node]").forEach((button) => {
      button.addEventListener("click", () => openMetadataModal(button.dataset.editNode));
    });
  }

  function isOnline(node, now = Date.now()) {
    const lastSeen = Date.parse(node.last_seen || "");
    return Number.isFinite(lastSeen) && now - lastSeen < 50000;
  }

  function osLogo(osID) {
    const id = String(osID || "").toLowerCase();
    if (id === "ubuntu") {
      return `<svg class="os-logo" viewBox="0 0 24 24" aria-hidden="true"><circle cx="12" cy="12" r="8.1" fill="none" stroke="#E95420" stroke-width="2.8"/><circle cx="12" cy="3.9" r="2.1" fill="#E95420"/><circle cx="19.1" cy="16.1" r="2.1" fill="#E95420"/><circle cx="4.9" cy="16.1" r="2.1" fill="#E95420"/><path d="M12 6.8v2.1m4.2 6.1-1.8-1.1m-8.4 0-1.8 1.1" stroke="#E95420" stroke-width="1.8" stroke-linecap="round"/></svg>`;
    }
    if (id === "debian") {
      return `<svg class="os-logo" viewBox="0 0 24 24" aria-hidden="true"><path d="M13.6 3.4c-3.9-.8-7.7 2.1-7.7 6.1 0 3.2 2.4 5.7 5.4 5.7 2.5 0 4.4-1.7 4.4-4.1 0-1.8-1.2-3.1-2.9-3.1-1.4 0-2.4.9-2.4 2.2 0 .8.5 1.4 1.3 1.4.6 0 1-.4 1-1" fill="none" stroke="#D70A53" stroke-width="2.1" stroke-linecap="round"/><path d="M7.2 15.2c1.6 2.1 4.2 3.4 7 3.1 2.7-.3 4.7-2.2 5-4.6" fill="none" stroke="#D70A53" stroke-width="1.8" stroke-linecap="round"/></svg>`;
    }
    return `<svg class="os-logo" viewBox="0 0 24 24" aria-hidden="true"><path d="M12 2.5 20.5 7v10L12 21.5 3.5 17V7L12 2.5Z" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linejoin="round"/><path d="M3.8 7.2 12 12l8.2-4.8M12 12v9" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round"/></svg>`;
  }

  function renderNode(node, now) {
    const online = isOnline(node, now);
    const countryCode = String(node.country_code || "").toUpperCase();
    const flag = countryFlag(countryCode);
    const locationParts = [node.city, node.region, node.country].filter(Boolean);
    const location = locationParts.length ? locationParts.join(", ") : translate("unknown");
    const cpu = safePercent(node.cpu_percent);
    const mem = safePercent(node.mem_percent);
    const disk = safePercent(node.disk_percent);
    const latencyRows = Array.isArray(node.latencies) && node.latencies.length ? node.latencies.map((probe) => {
      const loss = safePercent(probe.loss_percent);
      const broken = loss >= 100;
      const latency = broken ? translate("timeout") : `${(Number(probe.latency_ms) || 0).toFixed(1)} ms`;
      return `<div class="latency-row"><span class="latency-name"><span class="latency-dot" style="background:${broken ? "var(--red)" : loss > 0 ? "var(--orange)" : "var(--green-bright)"}"></span>${esc(displayProbeName(probe.name || probe.target || "Probe"))}</span><span class="latency-ms ${broken ? "bad" : ""}">${latency}</span><span class="latency-loss ${loss > 0 ? "bad" : ""}">${loss.toFixed(0)}% ${esc(translate("packetLoss"))}</span></div>`;
    }).join("") : `<div class="field-hint">${esc(translate("noLatency"))}</div>`;
    const publicIPv4 = node.public_ipv4 || "";
    const publicIPv6 = node.public_ipv6 || "";
    const networkParts = [node.isp, node.asn ? `AS${node.asn.replace(/^AS/i, "")}` : ""].filter(Boolean);
    const network = networkParts.length ? networkParts.join(" · ") : translate("unknown");
    const history = Array.isArray(node.history) ? node.history : [];
    const statusText = online ? translate("online") : translate("offline");
    const lastSeenText = node.last_seen ? new Date(node.last_seen).toLocaleString() : translate("firstReportWaiting");
    const cpuCores = Number(node.cpu_count) > 0 ? `${Number(node.cpu_count)} ${translate("cores")}` : "—";
    const operatingSystem = [node.os, node.arch].filter(Boolean).join(" · ") || translate("unknown");
    const kernel = node.kernel || "—";
    const load = [node.load1, node.load5, node.load15].map((item) => Number(item || 0).toFixed(2)).join(" / ");
    const uptime = fmtDuration(node.uptime_seconds);
    const rxRate = Number(node.rx_rate) || 0;
    const txRate = Number(node.tx_rate) || 0;
    const swapUsed = Number(node.swap_used_bytes) || 0;
    const swapTotal = Number(node.swap_total_bytes) || 0;
    const rxBytes = Number(node.rx_bytes) || 0;
    const txBytes = Number(node.tx_bytes) || 0;

    return `<article class="node-card ${online ? "is-online" : "is-offline"}">
      <div class="node-card-head">
        <div class="node-title-row">
          <div class="node-title-left"><div class="country-flag" aria-label="${esc(countryCode || "Unknown country")}">${flag}</div><div class="node-name-wrap"><h3 class="node-name">${esc(node.name || node.node_id || "Unnamed node")}</h3><div class="node-subname"><span>${esc(node.hostname || node.node_id || "—")}</span><span>·</span><span>${esc(countryCode || "--")}</span></div></div></div>
          <span class="node-status-pill ${online ? "" : "offline"}">${esc(statusText)}</span>${node.demo_data ? `<span class="node-badge demo-badge">${esc(translate("demoData"))}</span>` : ""}
        </div>
        <div class="node-actions">
          <span class="node-badge green-badge">${esc(countryCode || "GLOBAL")}</span>
          <span class="node-badge">${esc(String(node.arch || "Linux"))}</span>
          <span class="node-badge blue-badge">${osLogo(node.os_id)}${esc(String(node.os || "Linux"))}</span>
          ${node.city ? `<span class="node-badge">${esc(node.city)}</span>` : ""}
          <span class="node-quick-actions"><button class="mini-action metadata-action" type="button" title="${esc(translate("manageNode"))}" aria-label="${esc(translate("manageNode"))}" data-edit-node="${esc(node.node_id)}">✎</button><button class="mini-action" type="button" title="${esc(translate("deleteNode"))}" aria-label="${esc(translate("deleteNode"))}" data-remove-node="${esc(node.node_id)}">×</button></span>
        </div>
      </div>
      <div class="node-body">
        <div class="resource-grid">
          ${resourceItem("◉", "cpu", translate("cpu"), `${cpu.toFixed(1)}%`, cpu)}
          ${resourceItem("▥", "memory", translate("memory"), `${mem.toFixed(1)}%`, mem, `${fmtBytes(node.mem_used_bytes)} / ${fmtBytes(node.mem_total_bytes)}`)}
          ${resourceItem("▰", "disk", translate("disk"), `${disk.toFixed(1)}%`, disk, `${fmtBytes(node.disk_used_bytes)} / ${fmtBytes(node.disk_total_bytes)}`)}
          ${resourceItem("⌁", "load", translate("load"), load, Math.min(100, (Number(node.load1) || 0) / Math.max(1, Number(node.cpu_count) || 1) * 100), `${cpuCores}`)}
        </div>
        <div class="node-stat-strip">
          <div class="node-stat"><div class="node-stat-label"><span class="resource-icon">↓</span>${esc(translate("received"))}</div><div class="node-stat-value down">${fmtRate(rxRate)}</div><div class="field-hint">${esc(fmtBytes(rxBytes))}</div></div>
          <div class="node-stat"><div class="node-stat-label"><span class="resource-icon">↑</span>${esc(translate("sent"))}</div><div class="node-stat-value up">${fmtRate(txRate)}</div><div class="field-hint">${esc(fmtBytes(txBytes))}</div></div>
          <div class="node-stat"><div class="node-stat-label"><span class="resource-icon">◷</span>${esc(translate("uptime"))}</div><div class="node-stat-value">${esc(uptime)}</div><div class="field-hint">${esc(translate("trafficTotal"))}: ${fmtBytes(rxBytes + txBytes)}</div></div>
        </div>
        <div class="node-detail-grid">
          ${detailItem(translate("publicIPv4"), publicIPv4 || translate("noPublicIP"), true)}
          ${detailItem(translate("publicIPv6"), publicIPv6 || translate("noPublicIP"), true)}
          ${detailItem(translate("operatingSystem"), operatingSystem)}
          ${detailItem(translate("kernel"), kernel)}
          ${detailItem(translate("location"), location)}
          ${detailItem(translate("network"), network)}
          ${detailItem(translate("processor"), `${cpuCores} · ${esc(String(node.arch || "—"))}`)}
          ${detailItem(translate("lastSeen"), lastSeenText)}
          ${detailItem(translate("swap"), `${fmtBytes(swapUsed)} / ${fmtBytes(swapTotal)}`)}
          ${detailItem(translate("hostName"), node.hostname || "—")}
        </div>
        <div class="latency-section"><div class="latency-heading"><span>${esc(translate("latency"))}</span><span>${esc(translate("history24h"))}</span></div><div class="latency-list">${latencyRows}</div></div>
        ${renderManagedMetadata(node)}
        <div class="sparkline-wrap">
          <div class="sparkline"><div class="sparkline-head"><span>${esc(translate("cpuTrend"))}</span><strong>${cpu.toFixed(1)}%</strong></div>${sparkline(history.map((point) => Number(point.cpu_percent) || 0), "var(--blue)")}</div>
          <div class="sparkline"><div class="sparkline-head"><span>${esc(translate("memoryTrend"))}</span><strong>${mem.toFixed(1)}%</strong></div>${sparkline(history.map((point) => Number(point.mem_percent) || 0), "var(--purple)")}</div>
          <div class="sparkline"><div class="sparkline-head"><span>${esc(translate("trafficInTrend"))}</span><strong>${fmtRate(rxRate)}</strong></div>${sparklineScaled(history.map((point) => Number(point.rx_rate) || 0), "var(--green-bright)")}</div>
          <div class="sparkline"><div class="sparkline-head"><span>${esc(translate("trafficOutTrend"))}</span><strong>${fmtRate(txRate)}</strong></div>${sparklineScaled(history.map((point) => Number(point.tx_rate) || 0), "var(--orange)")}</div>
          ${(Array.isArray(node.latencies) ? node.latencies : []).map((probe) => {
            const values = history.map((point) => {
              const sample = Array.isArray(point.latencies) ? point.latencies.find((item) => item.name === probe.name && item.target === probe.target) : null;
              return sample && Number(sample.loss_percent) < 100 ? Number(sample.latency_ms) || 0 : null;
            }).filter((value) => value !== null);
            const latest = Number(probe.latency_ms) || 0;
            const label = `${translate("latencyTrend")}: ${displayProbeName(probe.name || probe.target)}`;
            return `<div class="sparkline"><div class="sparkline-head"><span>${esc(label)}</span><strong>${Number(probe.loss_percent) >= 100 ? esc(translate("timeout")) : `${latest.toFixed(1)} ms`}</strong></div>${sparklineScaled(values, "var(--cyan, var(--blue))")}</div>`;
          }).join("")}
        </div>
      </div>
    </article>`;
  }

  function renderManagedMetadata(node) {
    const items = [];
    if (node.expiry_date) {
      const expiry = new Date(`${node.expiry_date}T23:59:59Z`);
      const daysLeft = Math.ceil((expiry.getTime() - Date.now()) / 86400000);
      const label = daysLeft < 0 ? translate("expired") : `${translate("expiryIn")} ${daysLeft} ${translate("days")}`;
      items.push(`<div class="managed-chip ${daysLeft <= 7 ? "is-warning" : ""}"><span>${esc(translate("expiryDate"))}</span><strong>${esc(node.expiry_date)} · ${esc(label)}</strong></div>`);
    }
    if (node.price_usd !== undefined && node.price_usd !== null) items.push(`<div class="managed-chip"><span>${esc(translate("priceUSD"))}</span><strong>$${Number(node.price_usd).toFixed(2)}</strong></div>`);
    const monthlyQuota = Number(node.monthly_quota_bytes) || 0;
    if (monthlyQuota > 0) items.push(quotaChip(translate("monthlyUsage"), Number(node.monthly_usage_bytes) || 0, monthlyQuota));
    const totalQuota = Number(node.total_quota_bytes) || 0;
    if (totalQuota > 0) items.push(quotaChip(translate("totalUsage"), Number(node.total_usage_bytes) || 0, totalQuota));
    if (!items.length) return "";
    return `<div class="managed-section">${items.join("")}</div>`;
  }

  function quotaChip(label, used, quota) {
    const ratio = quota > 0 ? Math.min(100, used / quota * 100) : 0;
    return `<div class="managed-chip ${ratio >= 90 ? "is-warning" : ""}"><span>${esc(label)}</span><strong>${fmtBytes(used)} / ${fmtBytes(quota)}</strong><div class="quota-track"><span style="width:${ratio.toFixed(2)}%"></span></div>${ratio >= 90 ? `<small>${esc(translate("quotaNear"))}</small>` : ""}</div>`;
  }

  function sparklineScaled(values, stroke) {
    const filtered = values.filter((value) => Number.isFinite(Number(value))).map(Number);
    const width = 280;
    const height = 34;
    if (!filtered.length) return `<svg viewBox="0 0 ${width} ${height}" aria-hidden="true"><line x1="0" y1="${height - 2}" x2="${width}" y2="${height - 2}" stroke="var(--line-strong)" stroke-width="1"/></svg>`;
    const maxValue = Math.max(...filtered);
    const minValue = Math.min(...filtered);
    const span = maxValue - minValue;
    const points = sampleAcross(filtered, 240).map((value, index, series) => {
      const x = series.length === 1 ? width : index / (series.length - 1) * width;
      const y = span === 0 ? height / 2 : height - 3 - ((value - minValue) / span) * (height - 7);
      return `${x.toFixed(1)},${y.toFixed(1)}`;
    }).join(" ");
    return `<svg viewBox="0 0 ${width} ${height}" preserveAspectRatio="none" aria-hidden="true"><line x1="0" y1="${height - 2}" x2="${width}" y2="${height - 2}" stroke="var(--line-strong)" stroke-width="1"/><polyline points="${points}" fill="none" stroke="${stroke}" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" vector-effect="non-scaling-stroke"/></svg>`;
  }

  function resourceItem(icon, type, label, value, percent, extra = "") {
    const valueHtml = esc(String(value));
    return `<div class="resource-item"><div class="resource-top"><span class="resource-label"><span class="resource-icon">${icon}</span>${esc(label)}</span><span class="resource-value">${valueHtml}</span></div><div class="resource-track"><span class="resource-fill ${type}" style="width:${safePercent(percent)}%"></span></div>${extra ? `<div class="field-hint">${esc(extra)}</div>` : ""}</div>`;
  }

  function detailItem(label, value, strong = false) {
    return `<div class="detail-line"><div class="detail-label">${esc(label)}</div><div class="detail-value ${strong ? "strong" : ""}" title="${esc(String(value))}">${esc(String(value))}</div></div>`;
  }

  function sparkline(values, stroke) {
    const width = 280;
    const height = 34;
    if (!values.length) return `<svg viewBox="0 0 ${width} ${height}" aria-hidden="true"><line x1="0" y1="${height - 2}" x2="${width}" y2="${height - 2}" stroke="var(--line-strong)" stroke-width="1"/></svg>`;
    const trimmed = sampleAcross(values, 240);
    const points = trimmed.map((value, index) => {
      const x = trimmed.length === 1 ? width : (index / (trimmed.length - 1)) * width;
      const y = height - 3 - (safePercent(value) / 100) * (height - 7);
      return `${x.toFixed(1)},${y.toFixed(1)}`;
    }).join(" ");
    return `<svg viewBox="0 0 ${width} ${height}" preserveAspectRatio="none" aria-hidden="true"><line x1="0" y1="${height - 2}" x2="${width}" y2="${height - 2}" stroke="var(--line-strong)" stroke-width="1"/><polyline points="${points}" fill="none" stroke="${stroke}" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" vector-effect="non-scaling-stroke"/></svg>`;
  }

  function displayProbeName(name) {
    const key = ({ CT: "telecom", CU: "unicom", CM: "mobile" })[String(name || "").toUpperCase()];
    return key ? translate(key) : String(name || "Probe");
  }

  function countryFlag(code) {
    if (!/^[A-Z]{2}$/.test(code)) return "🌐";
    return String.fromCodePoint(...Array.from(code).map((letter) => 127397 + letter.charCodeAt(0)));
  }

  function safePercent(value) {
    const number = Number(value);
    if (!Number.isFinite(number)) return 0;
    return Math.max(0, Math.min(100, number));
  }

  function fmtBytes(input) {
    let value = Number(input) || 0;
    if (value < 0) value = 0;
    if (value < 1024) return `${value.toFixed(0)} B`;
    const units = ["KB", "MB", "GB", "TB", "PB"];
    let index = -1;
    do { value /= 1024; index++; } while (value >= 1024 && index < units.length - 1);
    return `${value >= 100 ? value.toFixed(0) : value >= 10 ? value.toFixed(1) : value.toFixed(2)} ${units[index]}`;
  }

  function fmtRate(value) {
    return `${fmtBytes(value)}/s`;
  }

  function fmtDuration(input) {
    let seconds = Math.max(0, Math.floor(Number(input) || 0));
    const days = Math.floor(seconds / 86400);
    seconds %= 86400;
    const hours = Math.floor(seconds / 3600);
    seconds %= 3600;
    const minutes = Math.floor(seconds / 60);
    if (days > 0) return `${days}d ${hours}h`;
    if (hours > 0) return `${hours}h ${minutes}m`;
    if (minutes > 0) return `${minutes}m`;
    return `${seconds}s`;
  }

  function esc(value) {
    return String(value == null ? "" : value).replace(/[&<>"']/g, (character) => ({
      "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;",
    })[character]);
  }

  function openMetadataModal(nodeID) {
    const node = state.nodes.find((item) => item.node_id === nodeID);
    if (!node) return;
    state.editingNodeID = nodeID;
    $("#metadataNodeTitle").textContent = `${translate("nodeMetadataTitle")} · ${node.name || nodeID}`;
    $("#metaExpiryDate").value = node.expiry_date || "";
    $("#metaPriceUSD").value = node.price_usd == null ? "" : String(node.price_usd);
    $("#metaMonthlyQuota").value = node.monthly_quota_bytes ? (Number(node.monthly_quota_bytes) / (1024 ** 3)).toFixed(3).replace(/0+$/, "").replace(/\.$/, "") : "";
    $("#metaTotalQuota").value = node.total_quota_bytes ? (Number(node.total_quota_bytes) / (1024 ** 3)).toFixed(3).replace(/0+$/, "").replace(/\.$/, "") : "";
    $("#metaIPv4").value = node.public_ipv4_override || "";
    $("#metaIPv6").value = node.public_ipv6_override || "";
    $("#metaCountry").value = node.country_override || "";
    $("#metaCountryCode").value = node.country_code_override || "";
    $("#metaRegion").value = node.region_override || "";
    $("#metaCity").value = node.city_override || "";
    $("#metaISP").value = node.isp_override || "";
    $("#metaASN").value = node.asn_override || "";
    $("#metadataFormError").hidden = true;
    $("#metadataModal").hidden = false;
    window.setTimeout(() => $("#metaExpiryDate").focus(), 0);
  }

  function closeMetadataModal() {
    $("#metadataModal").hidden = true;
    state.editingNodeID = "";
  }

  async function updateNodeMetadata(event) {
    event.preventDefault();
    const errorBox = $("#metadataFormError");
    errorBox.hidden = true;
    const toQuotaBytes = (value) => value === "" ? 0 : Math.round(Number(value) * (1024 ** 3));
    const priceText = $("#metaPriceUSD").value.trim();
    const payload = {
      expiry_date: $("#metaExpiryDate").value,
      price_usd: priceText === "" ? null : Number(priceText),
      monthly_quota_bytes: toQuotaBytes($("#metaMonthlyQuota").value.trim()),
      total_quota_bytes: toQuotaBytes($("#metaTotalQuota").value.trim()),
      public_ipv4_override: $("#metaIPv4").value.trim(),
      public_ipv6_override: $("#metaIPv6").value.trim(),
      country_override: $("#metaCountry").value.trim(),
      country_code_override: $("#metaCountryCode").value.trim().toUpperCase(),
      region_override: $("#metaRegion").value.trim(),
      city_override: $("#metaCity").value.trim(),
      isp_override: $("#metaISP").value.trim(),
      asn_override: $("#metaASN").value.trim(),
    };
    if ((priceText !== "" && (!Number.isFinite(payload.price_usd) || payload.price_usd < 0)) ||
        [payload.monthly_quota_bytes, payload.total_quota_bytes].some((value) => !Number.isSafeInteger(value) || value < 0)) {
      errorBox.textContent = translate("requestFailed");
      errorBox.hidden = false;
      return;
    }
    const button = $("#saveMetadataButton");
    setButtonBusy(button, true, "loading");
    try {
      await request(`/api/nodes/${encodeURIComponent(state.editingNodeID)}/metadata`, { method: "PATCH", mutation: true, body: JSON.stringify(payload) });
      closeMetadataModal();
      await loadNodes();
      showToast(translate("metadataSaved"));
    } catch (error) {
      errorBox.textContent = error.message || translate("requestFailed");
      errorBox.hidden = false;
    } finally {
      setButtonBusy(button, false, "saveMetadata");
    }
  }

  function openNodeModal() {
    state.createdToken = "";
    $("#nodeFormPane").hidden = false;
    $("#tokenPane").hidden = true;
    $("#nodeFormError").hidden = true;
    $("#nodeForm").reset();
    $("#nodeModal").hidden = false;
    window.setTimeout(() => $("#nodeName").focus(), 0);
  }

  function closeNodeModal() {
    $("#nodeModal").hidden = true;
    $("#createdToken").value = "";
    $("#agentConfigSnippet").value = "";
    state.createdToken = "";
  }

  async function createNodeToken(event) {
    event.preventDefault();
    const nodeName = $("#nodeName").value.trim();
    const nodeId = $("#nodeId").value.trim();
    const errorBox = $("#nodeFormError");
    errorBox.hidden = true;
    if (!nodeName || !/^[A-Za-z0-9][A-Za-z0-9._-]{0,47}$/.test(nodeId)) {
      errorBox.textContent = translate("nodeNameRequired");
      errorBox.hidden = false;
      return;
    }
    const button = $("#createTokenButton");
    setButtonBusy(button, true);
    try {
      const payload = await request("/api/nodes/token", {
        method: "POST", mutation: true,
        body: JSON.stringify({ node_id: nodeId, node_name: nodeName }),
      });
      state.createdToken = payload.token;
      $("#createdToken").value = payload.token;
      $("#createdToken").type = "password";
      $("#revealTokenButton").textContent = translate("showPassword");
      const configLines = [
        `POLARIS_SERVER_URL="${envEscape(payload.server_url)}"`,
        `POLARIS_NODE_ID="${envEscape(payload.node_id)}"`,
        `POLARIS_NODE_NAME="${envEscape(payload.node_name)}"`,
        `POLARIS_AGENT_TOKEN="${envEscape(payload.token)}"`,
        `POLARIS_PROBE_TARGETS="CT=114.114.114.114:53;CU=123.125.81.6:53;CM=www.10086.cn:443"`,
      ];
      $("#agentConfigSnippet").value = configLines.join("\n");
      $("#nodeFormPane").hidden = true;
      $("#tokenPane").hidden = false;
      await loadNodes();
      showToast(translate("tokenCreatedToast"));
    } catch (error) {
      errorBox.textContent = error.message || translate("createTokenFailed");
      errorBox.hidden = false;
    } finally {
      setButtonBusy(button, false, "generateToken");
    }
  }

  function envEscape(value) {
    return String(value).replace(/\\/g, "\\\\").replace(/"/g, "\\\"").replace(/\n/g, "");
  }

  async function removeNode(nodeID) {
    if (!window.confirm(translate("deleteConfirm"))) return;
    try {
      await request(`/api/nodes/${encodeURIComponent(nodeID)}`, { method: "DELETE", mutation: true, body: "{}" });
      await loadNodes();
      showToast(translate("nodeRemoved"));
    } catch (error) {
      if (error.status === 401) {
        showView("auth");
        showLoginError(translate("authExpired"));
      } else showToast(error.message || translate("requestFailed"));
    }
  }

  async function copyText(value) {
    try {
      await navigator.clipboard.writeText(value);
      showToast(translate("copySuccess"));
    } catch (_) {
      showToast(translate("copyFailed"));
    }
  }

  function showToast(message) {
    const toast = $("#toast");
    toast.textContent = message;
    toast.hidden = false;
    if (state.toastTimer) window.clearTimeout(state.toastTimer);
    state.toastTimer = window.setTimeout(() => { toast.hidden = true; }, 3000);
  }

  function bindEvents() {
    $("#languageButton").addEventListener("click", toggleLanguage);
    $("#dashLanguageButton").addEventListener("click", toggleLanguage);
    $("#themeButton").addEventListener("click", toggleTheme);
    $("#dashThemeButton").addEventListener("click", toggleTheme);
    $("#loginForm").addEventListener("submit", handleLogin);
    $("#passwordToggle").addEventListener("click", () => {
      const input = $("#password");
      input.type = input.type === "password" ? "text" : "password";
      $("#passwordToggle").textContent = translate(input.type === "password" ? "showPassword" : "hidePassword");
    });
    $("#logoutButton").addEventListener("click", logout);
    $("#addNodeButton").addEventListener("click", openNodeModal);
    $("#emptyAddNodeButton").addEventListener("click", openNodeModal);
    $("#closeNodeModal").addEventListener("click", closeNodeModal);
    $("#closeMetadataModal").addEventListener("click", closeMetadataModal);
    $("#metadataForm").addEventListener("submit", updateNodeMetadata);
    $("#metadataModal").addEventListener("click", (event) => { if (event.target === $("#metadataModal")) closeMetadataModal(); });
    $("#doneTokenButton").addEventListener("click", closeNodeModal);
    $("#nodeModal").addEventListener("click", (event) => { if (event.target === $("#nodeModal")) closeNodeModal(); });
    $("#nodeForm").addEventListener("submit", createNodeToken);
    $("#refreshButton").addEventListener("click", loadNodes);
    $("#copyTokenButton").addEventListener("click", () => copyText(state.createdToken));
    $("#copyConfigButton").addEventListener("click", () => copyText($("#agentConfigSnippet").value));
    $("#revealTokenButton").addEventListener("click", () => {
      const input = $("#createdToken");
      input.type = input.type === "password" ? "text" : "password";
      $("#revealTokenButton").textContent = translate(input.type === "password" ? "showPassword" : "hidePassword");
    });
    document.addEventListener("keydown", (event) => { if (event.key === "Escape" && !$("#nodeModal").hidden) closeNodeModal(); if (event.key === "Escape" && !$("#metadataModal").hidden) closeMetadataModal(); });
  }

  bindEvents();
  initialize();
})();
