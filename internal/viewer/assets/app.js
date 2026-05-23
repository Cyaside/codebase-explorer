(function () {
  const data = window.CODEARCH_VIEWER_DATA;
  if (!data) {
    document.body.innerHTML = "<main class='error-state'><h1>Viewer data missing</h1><p>The bundle does not contain viewer-data.js.</p></main>";
    return;
  }

  const text = (value) => (value === null || value === undefined || value === "" ? "—" : String(value));
  const html = (value) =>
    String(value)
      .replaceAll("&", "&amp;")
      .replaceAll("<", "&lt;")
      .replaceAll(">", "&gt;")
      .replaceAll('"', "&quot;");

  const sectionList = (items, renderer, emptyText) => {
    if (!items || items.length === 0) {
      return `<p class="empty">${html(emptyText)}</p>`;
    }
    return `<div class="stack">${items.map(renderer).join("")}</div>`;
  };

  const bundleTitle = document.getElementById("bundle-title");
  const bundleSubtitle = document.getElementById("bundle-subtitle");
  const projectName = document.getElementById("project-name");
  const projectSummary = document.getElementById("project-summary");
  const summaryCards = document.getElementById("summary-cards");

  bundleTitle.textContent = data.bundle_name || "Codebase Explorer";
  bundleSubtitle.textContent = data.project.analyzed_path || "Local bundle";
  projectName.textContent = data.project.name || data.bundle_name || "Codebase Explorer";
  projectSummary.textContent = data.project.summary || "No project summary was generated.";

  summaryCards.innerHTML = [
    metricCard("Type", text(data.project.type)),
    metricCard("Files", text(data.metrics.total_files)),
    metricCard("Lines", text(data.metrics.total_lines)),
    metricCard("AI", text(data.ai.status)),
  ].join("");

  document.getElementById("overview-body").innerHTML = [
    infoBlock("Warnings", data.warnings && data.warnings.length ? `<div class="stack">${data.warnings.map((warning) => `<article class="item"><p>${html(warning)}</p></article>`).join("")}</div>` : `<p class="empty">No runtime warnings for this bundle.</p>`),
    infoBlock("Primary languages", sectionList(data.languages, (language) => `<article class="item"><strong>${html(language.name)}</strong><span>${language.file_count} files · ${language.line_count} lines</span></article>`, "No language summary available.")),
    infoBlock("Important directories", pillList(data.important_directories)),
    infoBlock("Entry points", pillList(data.entry_points)),
    infoBlock("AI project summary", data.ai.project_summary ? `<p>${html(data.ai.project_summary)}</p>` : `<p class="empty">AI project summary is unavailable for this run.</p>`),
  ].join("");

  document.getElementById("architecture-body").innerHTML = [
    infoBlock("Core modules", sectionList(data.modules, (module) => `<article class="item"><strong>${html(module.path)}</strong><span>${module.file_count} files · ${module.total_lines} lines · ${module.entry_point_count} entry point(s)</span></article>`, "No module summary available.")),
    infoBlock("AI architecture narrative", data.ai.architecture_narrative ? `<p>${html(data.ai.architecture_narrative)}</p>` : `<p class="empty">AI architecture narrative is unavailable for this run.</p>`),
  ].join("");

  document.getElementById("hotspots-body").innerHTML = sectionList(
    data.hotspots,
    (hotspot) => {
      const explanation = (data.ai.hotspot_explanations || []).find((item) => item.path === hotspot.path);
      return `
        <article class="item">
          <strong>${html(hotspot.path)}</strong>
          <span>score ${hotspot.score.toFixed(2)} · ${hotspot.line_count} lines · ${hotspot.import_count} imports</span>
          <p>${html((hotspot.reasons || []).join("; ") || "No detailed heuristic reason recorded.")}</p>
          ${explanation ? `<p class="ai-note">AI: ${html(explanation.explanation)}</p>` : ""}
        </article>
      `;
    },
    "No hotspot candidates were recorded."
  );

  document.getElementById("dependencies-body").innerHTML = sectionList(
    data.dependencies,
    (risk) => `
      <article class="item">
        <strong>${html(risk.path)}</strong>
        <span>${risk.import_count} import-like references</span>
        <p>${html(risk.reason)}</p>
      </article>
    `,
    "No concentrated dependency risks were recorded."
  );

  document.getElementById("reading-path-body").innerHTML = sectionList(
    data.reading_path,
    (item) => {
      const rationale = (data.ai.reading_path_explanations || []).find((entry) => entry.path === item.path);
      return `
        <article class="item">
          <strong>${html(item.path)}</strong>
          <p>${html(item.reason)}</p>
          ${rationale ? `<p class="ai-note">AI: ${html(rationale.rationale)}</p>` : ""}
        </article>
      `;
    },
    "No reading path guidance was recorded."
  );

  document.getElementById("changes-body").innerHTML = [
    infoBlock("Status", `<p>${html(data.changes.note || (data.changes.available ? "Support files were parsed for this bundle." : "No supporting files were available for this bundle."))}</p>`),
    infoBlock("Supporting files", sectionList(data.changes.sources, (source) => `
      <article class="item">
        <strong>${html(source.path)}</strong>
        <span>${html(source.kind)} &middot; ${html(source.format)} &middot; ${html(source.status)}</span>
        ${source.message ? `<p>${html(source.message)}</p>` : ""}
      </article>
    `, "No supporting files were recorded.")),
    infoBlock("Frequently mentioned areas", sectionList(data.changes.frequently_mentioned_areas, (area) => `
      <article class="item">
        <strong>${html(area.path)}</strong>
        <span>${area.mention_count} mention(s) &middot; ${html(area.confidence)} confidence</span>
        ${area.reasons && area.reasons.length ? `<p>${html(area.reasons.join("; "))}</p>` : ""}
      </article>
    `, "No repository areas were matched strongly enough.")),
    infoBlock("Repeated themes", sectionList(data.changes.repeated_themes, (theme) => `
      <article class="item">
        <strong>${html(theme.name)}</strong>
        <span>${theme.mention_count} mention(s) &middot; ${theme.source_count} source(s)</span>
        ${theme.related_areas && theme.related_areas.length ? `<p>Related areas: ${html(theme.related_areas.join(", "))}</p>` : ""}
      </article>
    `, "No repeated themes met the reporting threshold."))
  ].join("");

  const graphViews = data.graphs?.views || [];
  document.getElementById("diagrams-body").innerHTML = graphViews.length
    ? graphViews.map((view) => {
        const nodes = view.nodes || [];
        const labels = new Map(nodes.map((node) => [node.id, node.label]));
        const edges = view.edges || [];
        const rows = edges.slice(0, 12).map((edge) => `<article class="item"><strong>${html(labels.get(edge.source) || edge.source)} → ${html(labels.get(edge.target) || edge.target)}</strong><span>${html(edge.relation)}</span></article>`);
        return infoBlock(`${view.id} · ${nodes.length} nodes · ${edges.length} relationships`, rows.length
          ? `<div class="stack">${rows.join("")}</div>${edges.length > rows.length ? `<p class="empty">Showing ${rows.length} relationships. Open graphs.json for the full set.</p>` : ""}`
          : `<p class="empty">No relationships in this view.</p>`);
      }).join("") + (data.links?.architecture_diagram ? `<p><a class="viewer-link" href="../${html(data.links.architecture_diagram)}">Open structured graph data</a></p>` : "")
    : `<p class="empty">This older bundle has no structured graph data.</p>`;

  document.getElementById("bundle-links").innerHTML = [
    linkItem("Root README", data.links.root),
    linkItem("Overview", data.links.overview),
    linkItem("Architecture", data.links.architecture),
    linkItem("Dependencies", data.links.dependencies),
    linkItem("Hotspots", data.links.hotspots),
    linkItem("Reading Path", data.links.reading_path),
    linkItem("Changes", data.links.changes),
  ].join("");

  function metricCard(label, value) {
    return `<article class="metric-card"><span>${html(label)}</span><strong>${html(value)}</strong></article>`;
  }

  function infoBlock(title, body) {
    return `<section class="block"><h4>${html(title)}</h4>${body}</section>`;
  }

  function pillList(items) {
    if (!items || items.length === 0) {
      return `<p class="empty">No entries recorded.</p>`;
    }
    return `<div class="pill-row">${items.map((item) => `<span class="pill">${html(item)}</span>`).join("")}</div>`;
  }

  function linkItem(label, href) {
    return `<li><a class="viewer-link" href="../${html(href)}">${html(label)}</a></li>`;
  }
})();
