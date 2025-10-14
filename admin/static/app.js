async function loadConfig() {
  try {
    const res = await fetch("/api/config");
    const cfg = await res.json();
    renderConfig(cfg);
  } catch (err) {
    document.body.innerHTML = `<div class="p-10 text-red-500">Failed to load configuration: ${err}</div>`;
  }
}

function renderConfig(cfg) {
  // Header info
  document.getElementById("gateway-info").textContent =
    `${cfg.name} v${cfg.version} • Schema: ${cfg.schema} • Port: ${cfg.server.port}`;

  // Routes
  const routesEl = document.getElementById("routes");
  routesEl.innerHTML = "";

  cfg.routes.forEach(route => {
    const backends = route.backends.map(b => `
      <li class="ml-2">→ ${b.method || "GET"} <span class="text-blue-400">${b.url}</span></li>
    `).join("");

    const plugins = route.plugins.map(p => `<span class="bg-blue-800 text-blue-200 px-2 py-0.5 rounded text-xs mr-1">${p.name}</span>`).join("");

    routesEl.innerHTML += `
      <div class="card">
        <div class="flex justify-between items-center mb-2">
          <span class="font-semibold">${route.method}</span>
          <span class="text-blue-400">${route.path}</span>
        </div>
        <div class="mb-2">
          <span class="text-sm text-gray-400">Aggregate:</span> ${route.aggregate || "-"}<br>
          <span class="text-sm text-gray-400">Transform:</span> ${route.transform || "-"}
        </div>
        <div class="mb-2">${plugins}</div>
        <ul class="text-sm text-gray-300">${backends}</ul>
      </div>
    `;
  });

  // Plugins
  const pluginsEl = document.getElementById("plugins");
  pluginsEl.innerHTML = cfg.plugins.map(p => `
    <div class="card text-center">
      <h3 class="font-semibold mb-1">${p.name}</h3>
      <pre>${JSON.stringify(p.config, null, 2)}</pre>
    </div>
  `).join("");
}

loadConfig();
