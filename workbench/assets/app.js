(function () {
  const STORAGE_KEY = "codearch.workbench.profiles.v1";
  const PRESET_PROFILES = [
    { id: "deterministic", label: "Deterministic only", provider: null, locked: true },
    { id: "openai", label: "OpenAI", provider: { name: "openai", model: "gpt-4.1-mini", apiKey: "", baseUrl: "" } },
    { id: "openrouter", label: "OpenRouter", provider: { name: "openai-compatible", model: "openai/gpt-4.1-mini", apiKey: "", baseUrl: "https://openrouter.ai/api/v1" } },
    { id: "mistral", label: "Mistral", provider: { name: "openai-compatible", model: "mistral-small-latest", apiKey: "", baseUrl: "https://api.mistral.ai/v1" } },
  ];
  const loadedProfiles = loadProfiles();

  const state = {
    activeTab: "summary",
    status: null,
    bundles: [],
    bundleCache: {},
    selectedBundle: "",
    profiles: loadedProfiles,
    selectedProfile: loadedProfiles[0] ? loadedProfiles[0].id : "deterministic",
    busy: false,
  };

  const refs = {
    workspaceTitle: document.getElementById("workspace-title"),
    workspaceSubtitle: document.getElementById("workspace-subtitle"),
    bundleReadmeLink: document.getElementById("bundle-readme-link"),
    analyzeForm: document.getElementById("analyze-form"),
    analyzeButton: document.getElementById("analyze-button"),
    analyzeStateBadge: document.getElementById("analyze-state-badge"),
    repoPathInput: document.getElementById("repo-path-input"),
    supportFilesInput: document.getElementById("support-files-input"),
    ignorePatternsInput: document.getElementById("ignore-patterns-input"),
    profileList: document.getElementById("profile-list"),
    bundleList: document.getElementById("bundle-list"),
    systemFacts: document.getElementById("system-facts"),
    statsRow: document.getElementById("stats-row"),
    profileLabelInput: document.getElementById("profile-label-input"),
    profileProviderInput: document.getElementById("profile-provider-input"),
    profileModelInput: document.getElementById("profile-model-input"),
    profileBaseURLInput: document.getElementById("profile-base-url-input"),
    profileAPIKeyInput: document.getElementById("profile-api-key-input"),
    connectionTitle: document.getElementById("connection-title"),
    connectionModePill: document.getElementById("connection-mode-pill"),
    profileHint: document.getElementById("profile-hint"),
    saveProfileButton: document.getElementById("save-profile-button"),
    duplicateProfileButton: document.getElementById("duplicate-profile-button"),
    deleteProfileButton: document.getElementById("delete-profile-button"),
    newProfileButton: document.getElementById("new-profile-button"),
    refreshButton: document.getElementById("refresh-button"),
    reloadDashboardButton: document.getElementById("reload-dashboard-button"),
    tabButtons: Array.from(document.querySelectorAll(".tab-button")),
    panels: {
      summary: document.getElementById("panel-summary"),
      architecture: document.getElementById("panel-architecture"),
      flowchart: document.getElementById("panel-flowchart"),
      issues: document.getElementById("panel-issues"),
      recommendations: document.getElementById("panel-recommendations"),
    },
    toast: document.getElementById("toast"),
  };

  bindEvents();
  initialize();

  async function initialize() {
    await refreshStatus();
    render();
  }

  function bindEvents() {
    refs.analyzeForm.addEventListener("submit", onAnalyzeSubmit);
    refs.saveProfileButton.addEventListener("click", onSaveProfile);
    refs.duplicateProfileButton.addEventListener("click", onDuplicateProfile);
    refs.deleteProfileButton.addEventListener("click", onDeleteProfile);
    refs.newProfileButton.addEventListener("click", onNewProfile);
    refs.refreshButton.addEventListener("click", refreshStatus);
    refs.reloadDashboardButton.addEventListener("click", refreshStatus);
    refs.profileList.addEventListener("click", onProfileListClick);
    refs.bundleList.addEventListener("click", onBundleListClick);
    refs.tabButtons.forEach((button) => button.addEventListener("click", onTabClick));
  }

  async function refreshStatus() {
    try {
      const status = await requestJSON("/api/status");
      state.status = status;
      state.bundles = status.recent_bundles || [];
      if (state.selectedBundle && !state.bundles.some((item) => item.name === state.selectedBundle)) state.selectedBundle = "";
      if (!state.selectedBundle && state.bundles[0]) state.selectedBundle = state.bundles[0].name;
      if (state.selectedBundle) await loadBundle(state.selectedBundle);
      render();
    } catch (error) {
      toast(error.message || "Failed to load workbench status.", true);
    }
  }

  async function loadBundle(bundleName) {
    if (!bundleName) return;
    if (!state.bundleCache[bundleName]) {
      state.bundleCache[bundleName] = await requestJSON(`/api/bundles/${encodeURIComponent(bundleName)}`);
    }
    state.selectedBundle = bundleName;
    render();
  }

  async function onAnalyzeSubmit(event) {
    event.preventDefault();
    const repoPath = refs.repoPathInput.value.trim();
    if (!repoPath) return toast("Repository path is required.", true);

    const profile = currentProfile();
    state.busy = true;
    render();

    const payload = {
      repo_path: repoPath,
      deterministic_only: !profile.provider,
      support_files: splitLines(refs.supportFilesInput.value),
      extra_ignore_patterns: splitLines(refs.ignorePatternsInput.value),
      provider: buildProviderPayload(profile),
    };

    try {
      const response = await requestJSON("/api/analyze", { method: "POST", body: JSON.stringify(payload) });
      state.bundleCache[response.bundle.name] = { summary: response.bundle, data: response.data };
      state.selectedBundle = response.bundle.name;
      state.bundles = [response.bundle].concat(state.bundles.filter((item) => item.name !== response.bundle.name));
      await refreshStatus();
      toast(`Analysis ready for ${response.bundle.project_name || response.result.project_name || "repository"}.`);
    } catch (error) {
      toast(error.message || "Analyze request failed.", true);
    } finally {
      state.busy = false;
      render();
    }
  }

  function onProfileListClick(event) {
    const button = event.target.closest("[data-profile-id]");
    if (!button) return;
    state.selectedProfile = button.dataset.profileId;
    render();
  }

  async function onBundleListClick(event) {
    const button = event.target.closest("[data-bundle-name]");
    if (!button) return;
    try {
      await loadBundle(button.dataset.bundleName);
    } catch (error) {
      toast(error.message || "Failed to load bundle.", true);
    }
  }

  function onTabClick(event) {
    state.activeTab = event.currentTarget.dataset.tab;
    render();
  }

  function onNewProfile() {
    const id = `profile-${Date.now()}`;
    state.profiles = state.profiles.concat({
      id,
      label: "Custom connection",
      provider: { name: "openai-compatible", model: "", apiKey: "", baseUrl: "" },
    });
    state.selectedProfile = id;
    persistProfiles();
    render();
  }

  function onDuplicateProfile() {
    const selected = currentProfile();
    const clone = JSON.parse(JSON.stringify(selected));
    clone.id = `profile-${Date.now()}`;
    clone.label = `${selected.label} Copy`;
    clone.locked = false;
    state.profiles = state.profiles.concat(clone);
    state.selectedProfile = clone.id;
    persistProfiles();
    render();
  }

  function onDeleteProfile() {
    const selected = currentProfile();
    if (selected.locked) return toast("Preset profiles cannot be deleted.", true);
    state.profiles = state.profiles.filter((profile) => profile.id !== selected.id);
    state.selectedProfile = state.profiles[0] ? state.profiles[0].id : "deterministic";
    persistProfiles();
    render();
  }

  function onSaveProfile() {
    const selected = currentProfile();
    selected.label = refs.profileLabelInput.value.trim() || "Connection";
    const providerName = refs.profileProviderInput.value.trim();
    selected.provider = providerName
      ? { name: providerName, model: refs.profileModelInput.value.trim(), apiKey: refs.profileAPIKeyInput.value.trim(), baseUrl: refs.profileBaseURLInput.value.trim() }
      : null;
    persistProfiles();
    render();
    toast(`Saved connection "${selected.label}".`);
  }

  function render() {
    const profile = currentProfile();
    const bundle = currentBundle();
    refs.workspaceTitle.textContent = bundle ? bundle.summary.project_name || "Workbench" : "Codebase Explorer Workbench";
    refs.workspaceSubtitle.textContent = bundle ? `${bundle.summary.project_type || "Repository"} · ${bundle.summary.total_files} files · ${bundle.summary.total_lines} lines` : "Recent bundles and local analysis are available here.";
    refs.bundleReadmeLink.href = bundle ? bundleLink(bundle.summary.name, "README.md") : "#";
    refs.analyzeStateBadge.textContent = state.busy ? "Running" : "Idle";
    refs.analyzeStateBadge.className = `status-pill ${state.busy ? "active" : "neutral"}`;
    refs.analyzeButton.disabled = state.busy;

    refs.connectionTitle.textContent = profile.label;
    refs.connectionModePill.textContent = profile.provider ? profile.provider.name : "Deterministic";
    refs.connectionModePill.className = `status-pill ${profile.provider ? "active" : "neutral"}`;
    refs.profileLabelInput.value = profile.label || "";
    refs.profileProviderInput.value = profile.provider ? profile.provider.name || "" : "";
    refs.profileModelInput.value = profile.provider ? profile.provider.model || "" : "";
    refs.profileBaseURLInput.value = profile.provider ? profile.provider.baseUrl || "" : "";
    refs.profileAPIKeyInput.value = profile.provider ? profile.provider.apiKey || "" : "";
    refs.deleteProfileButton.disabled = !!profile.locked;
    refs.profileHint.textContent = profile.provider ? "API keys stay local to this browser and are never written into the analysis bundle." : "Deterministic mode keeps analysis fully local without calling any provider.";

    refs.profileList.innerHTML = state.profiles.map(renderProfileButton).join("");
    refs.bundleList.innerHTML = state.bundles.length ? state.bundles.map(renderBundleButton).join("") : `<p class="empty">No bundle yet. Run your first analysis.</p>`;
    refs.systemFacts.innerHTML = renderSystemFacts();
    refs.statsRow.innerHTML = renderStats(bundle);
    refs.panels.summary.innerHTML = renderSummaryPanel(bundle);
    refs.panels.architecture.innerHTML = renderArchitecturePanel(bundle);
    refs.panels.flowchart.innerHTML = renderFlowchartPanel(bundle);
    refs.panels.issues.innerHTML = renderIssuesPanel(bundle);
    refs.panels.recommendations.innerHTML = renderRecommendationsPanel(bundle);
    refs.tabButtons.forEach((button) => button.classList.toggle("active", button.dataset.tab === state.activeTab));
    Object.entries(refs.panels).forEach(([name, panel]) => panel.classList.toggle("active", name === state.activeTab));
  }

  function renderProfileButton(profile) {
    const meta = profile.provider ? `${profile.provider.name}${profile.provider.model ? ` · ${profile.provider.model}` : ""}` : "Deterministic only";
    return `<button class="sidebar-card ${profile.id === state.selectedProfile ? "selected" : ""}" type="button" data-profile-id="${escapeHTML(profile.id)}"><strong>${escapeHTML(profile.label)}</strong><span>${escapeHTML(meta)}</span></button>`;
  }

  function renderBundleButton(bundle) {
    return `<button class="sidebar-card ${bundle.name === state.selectedBundle ? "selected" : ""}" type="button" data-bundle-name="${escapeHTML(bundle.name)}"><strong>${escapeHTML(bundle.project_name || bundle.name)}</strong><span>${escapeHTML(bundle.ai_status || "deterministic")} · ${bundle.total_files} files</span></button>`;
  }

  function renderSystemFacts() {
    if (!state.status) return `<p class="empty">Loading system details...</p>`;
    const facts = [
      ["Version", state.status.app_version || "dev"],
      ["Output root", state.status.output_root || "out"],
      ["Cache", state.status.cache_enabled ? state.status.cache_root : "disabled"],
      ["Providers", (state.status.supported_providers || []).map((item) => item.name).join(", ") || "none"],
    ];
    return facts.map(([label, value]) => `<div class="fact-row"><span>${escapeHTML(label)}</span><strong>${escapeHTML(value)}</strong></div>`).join("");
  }

  function renderStats(bundle) {
    const values = bundle
      ? [
          ["Project", bundle.summary.project_name || "Repository", bundle.summary.project_type || "Unknown shape"],
          ["Files", bundle.summary.total_files, bundle.summary.analyzed_path || "Local checkout"],
          ["AI", bundle.summary.ai_status || "disabled", bundle.data.ai.provider || "No provider"],
          ["Changes", bundle.data.changes.frequently_mentioned_areas.length, `${bundle.summary.support_file_count} support file(s)`],
        ]
      : [
          ["Projects", state.bundles.length, "Recent bundles available"],
          ["Connections", state.profiles.length, "Saved locally in this browser"],
          ["Output root", state.status ? state.status.output_root : "out", "Local-first storage"],
          ["Mode", "Deterministic-first", "AI stays optional"],
        ];
    return values.map(([label, value, note]) => `<article class="stat-card"><span>${escapeHTML(label)}</span><strong>${escapeHTML(value)}</strong><p>${escapeHTML(note)}</p></article>`).join("");
  }

  function renderSummaryPanel(bundle) {
    if (!bundle) return emptyPanel("Run an analysis or pick a bundle from the sidebar.");
    return panelTemplate("Project snapshot", [
      contentCard("Summary", `<p>${escapeHTML(bundle.data.project.summary || bundle.data.ai.project_summary || "No summary available.")}</p>`),
      contentCard("Warnings", renderSimpleList(bundle.data.warnings, "No runtime warnings recorded.")),
      contentCard("Languages", renderMetricList(bundle.data.languages.map((item) => `${item.name} · ${item.file_count} files · ${item.line_count} lines`), "No language data.")),
      contentCard("Entry points", renderMetricList(bundle.data.entry_points, "No entry points recorded.")),
      contentCard("Important directories", renderMetricList(bundle.data.important_directories, "No important directories recorded.")),
    ]);
  }

  function renderArchitecturePanel(bundle) {
    if (!bundle) return emptyPanel("Architecture details will appear after a bundle is selected.");
    return panelTemplate("Architecture", [
      contentCard("Narrative", `<p>${escapeHTML(bundle.data.ai.architecture_narrative || "AI architecture narrative is unavailable for this bundle.")}</p>`),
      contentCard("Core modules", renderMetricList(bundle.data.core_modules, "No core module list recorded.")),
      contentCard("Module inventory", renderMetricList((bundle.data.modules || []).slice(0, 12).map((item) => `${item.path} · ${item.file_count} files · ${item.total_lines} lines`), "No module inventory recorded.")),
    ]);
  }

  function renderFlowchartPanel(bundle) {
    if (!bundle) return emptyPanel("Flowchart views depend on a selected bundle.");
    return panelTemplate("Codebase flowchart", [
      contentCard("Module graph", `<p><a class="inline-link" href="${bundleLink(bundle.summary.name, bundle.data.links.architecture_diagram)}" target="_blank" rel="noreferrer">Open raw Mermaid file</a></p><pre class="code-block">${escapeHTML(bundle.data.mermaid.architecture || "No architecture diagram source.")}</pre>`),
      contentCard("Dependency graph", `<p><a class="inline-link" href="${bundleLink(bundle.summary.name, bundle.data.links.dependency_diagram)}" target="_blank" rel="noreferrer">Open raw Mermaid file</a></p><pre class="code-block">${escapeHTML(bundle.data.mermaid.dependencies || "No dependency diagram source.")}</pre>`),
      contentCard("Top modules", renderMetricList((bundle.data.modules || []).slice(0, 8).map((item) => `${item.path} · ${item.entry_point_count} entry point(s)`), "No module topology available.")),
    ]);
  }

  function renderIssuesPanel(bundle) {
    if (!bundle) return emptyPanel("Issue tracking becomes available when a bundle is selected.");
    return panelTemplate("Issue tracking", [
      contentCard("Change note", `<p>${escapeHTML(bundle.data.changes.note || "No support files were linked to this bundle.")}</p>`),
      contentCard("Support files", renderMetricList((bundle.data.changes.sources || []).map((item) => `${item.path} · ${item.kind} · ${item.status}`), "No support files were recorded.")),
      contentCard("Frequently mentioned areas", renderMetricList((bundle.data.changes.frequently_mentioned_areas || []).map((item) => `${item.path} · ${item.mention_count} mention(s) · ${item.confidence}`), "No repository areas were matched strongly enough.")),
      contentCard("Repeated themes", renderMetricList((bundle.data.changes.repeated_themes || []).map((item) => `${item.name} · ${item.mention_count} mention(s)`), "No repeated themes crossed the reporting threshold.")),
    ]);
  }

  function renderRecommendationsPanel(bundle) {
    if (!bundle) return emptyPanel("Recommendations will appear after selecting a bundle.");
    const hotspotLines = (bundle.data.ai.hotspot_explanations || []).map((item) => `${item.path} · ${item.explanation}`);
    const pathLines = (bundle.data.ai.reading_path_explanations || []).map((item) => `${item.path} · ${item.rationale}`);
    const deterministicPath = (bundle.data.reading_path || []).map((item) => `${item.path} · ${item.reason}`);
    return panelTemplate("Recommendations", [
      contentCard("Reading path", renderMetricList(pathLines.length ? pathLines : deterministicPath, "No reading path recommendation available.")),
      contentCard("Hotspot guidance", renderMetricList(hotspotLines, "AI hotspot notes are unavailable for this bundle.")),
      contentCard("Next moves", renderMetricList([
        `Open raw README · ${bundleLink(bundle.summary.name, "README.md")}`,
        `Open viewer snapshot · ${bundleLink(bundle.summary.name, "ui/index.html")}`,
        `Review bundle warnings · ${bundle.data.warnings.length} warning(s)`,
      ], "No follow-up recommendations available.")),
    ]);
  }

  function panelTemplate(title, blocks) {
    return `<div class="panel-header"><div><p class="eyebrow">${escapeHTML(title)}</p><h3>${escapeHTML(title)}</h3></div></div><div class="panel-stack">${blocks.join("")}</div>`;
  }
  function contentCard(title, body) { return `<section class="content-card"><h4>${escapeHTML(title)}</h4>${body}</section>`; }
  function emptyPanel(message) { return `<div class="empty-state"><p>${escapeHTML(message)}</p></div>`; }
  function renderSimpleList(items, emptyText) { return !items || !items.length ? `<p class="empty">${escapeHTML(emptyText)}</p>` : `<ul class="detail-list">${items.map((item) => `<li>${escapeHTML(item)}</li>`).join("")}</ul>`; }
  function renderMetricList(items, emptyText) { return !items || !items.length ? `<p class="empty">${escapeHTML(emptyText)}</p>` : `<div class="metric-list">${items.map((item) => `<article class="metric-row"><span>${escapeHTML(item)}</span></article>`).join("")}</div>`; }

  function buildProviderPayload(profile) {
    return profile.provider ? { name: profile.provider.name || "", model: profile.provider.model || "", api_key: profile.provider.apiKey || "", base_url: profile.provider.baseUrl || "" } : null;
  }
  function currentProfile() { return state.profiles.find((profile) => profile.id === state.selectedProfile) || state.profiles[0]; }
  function currentBundle() { return state.bundleCache[state.selectedBundle] || null; }
  function persistProfiles() { window.localStorage.setItem(STORAGE_KEY, JSON.stringify(state.profiles)); }
  function splitLines(value) { return value.split(/\r?\n/).map((item) => item.trim()).filter(Boolean); }
  function bundleLink(bundleName, relativePath) { return `/bundles/${encodeURIComponent(bundleName)}/${relativePath}`; }
  function cloneProfile(profile) { return JSON.parse(JSON.stringify(profile)); }
  function escapeHTML(value) { return String(value || "").replaceAll("&", "&amp;").replaceAll("<", "&lt;").replaceAll(">", "&gt;").replaceAll('"', "&quot;"); }

  function loadProfiles() {
    try {
      const raw = window.localStorage.getItem(STORAGE_KEY);
      if (!raw) return PRESET_PROFILES.map(cloneProfile);
      const parsed = JSON.parse(raw);
      if (!Array.isArray(parsed) || !parsed.length) return PRESET_PROFILES.map(cloneProfile);
      const ids = new Set(parsed.map((item) => item.id));
      const merged = parsed.map(cloneProfile);
      PRESET_PROFILES.forEach((profile) => { if (!ids.has(profile.id)) merged.unshift(cloneProfile(profile)); });
      return merged;
    } catch (_) {
      return PRESET_PROFILES.map(cloneProfile);
    }
  }

  function requestJSON(url, options) {
    return fetch(url, Object.assign({ headers: { "Content-Type": "application/json" } }, options || {})).then(async (response) => {
      const data = await response.json().catch(() => ({}));
      if (!response.ok) throw new Error(data.error || `Request failed with ${response.status}`);
      return data;
    });
  }

  function toast(message, isError) {
    refs.toast.hidden = false;
    refs.toast.textContent = message;
    refs.toast.className = `toast ${isError ? "error" : "success"}`;
    window.clearTimeout(toast.timer);
    toast.timer = window.setTimeout(() => { refs.toast.hidden = true; }, 3000);
  }
})();
