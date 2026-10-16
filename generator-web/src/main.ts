import JSZip from "jszip";
import { TEMPLATE_FILE_PATHS } from "./templateManifest";

type DatabaseConfig = {
  type?: string;
  host: string;
  port: string;
  database: string;
  username: string;
  password: string;
};

type ServiceInput = {
  name: string;
  httpPort: string;
  grpcPort: string;
  db: DatabaseConfig;
};

type ServiceSpec = {
  name: string;
  httpPort: number;
  grpcPort: number;
  db: DatabaseConfig;
};

type GenerationRequest = {
  projectName: string;
  includeTraefik: boolean;
  traefikPort: number;
  includeAuth: boolean;
  authHttpPort: number;
  authGrpcPort: number;
  authDb: DatabaseConfig;
  services: ServiceSpec[];
};

const RAW_BASE =
  "https://raw.githubusercontent.com/jiwomdf/go-fiber-microservice-generator/main/";
const EXCLUDED_TOP_LEVEL = new Set([".git", ".DS_Store", "generator-web"]);
const servicesEl = document.getElementById("services") as HTMLDivElement;
const countEl = document.getElementById("service-count") as HTMLInputElement;
const formEl = document.getElementById("generator-form") as HTMLFormElement;
const statusEl = document.getElementById("status") as HTMLDivElement;
const buttonEl = document.getElementById(
  "generate-button",
) as HTMLButtonElement;
const applyServicesButtonEl = document.getElementById(
  "apply-services-button",
) as HTMLButtonElement;

function slugify(value: string): string {
  return value
    .trim()
    .toLowerCase()
    .replace(/[^a-z0-9]+/g, "-")
    .replace(/^-+|-+$/g, "");
}

function pascalCase(value: string): string {
  return value
    .split(/[^A-Za-z0-9]+/)
    .filter(Boolean)
    .map((part) => part.charAt(0).toUpperCase() + part.slice(1).toLowerCase())
    .join("");
}

function replaceAllLiteral(value: string, from: string, to: string): string {
  return value.split(from).join(to);
}

function setStatus(message: string, isError = false): void {
  statusEl.textContent = message;
  statusEl.className = isError ? "status error" : "status";
}

function renderServiceInputs(): void {
  const count = Math.max(0, Math.min(20, Number(countEl.value || 0)));
  servicesEl.innerHTML = "";

  for (let i = 0; i < count; i += 1) {
    const card = document.createElement("div");
    card.className = "service-card extra-service-card";
    card.innerHTML = `
      <div class="service-card-header">
        <div class="service-card-title">
          <span class="service-pill">Microservice</span>
          <strong>Service ${i + 1}</strong>
        </div>
      </div>

      <div class="service-subsection non-db-section">
        <h4 class="subsection-title">Service Config</h4>
        <div class="grid">
          <label>
            Service Name
            <span class="label-caption">Unique entity name (e.g. inventory)</span>
            <input name="service-name" placeholder="inventory" required>
          </label>
          <label>
            HTTP Port
            <span class="label-caption">REST/HTTP application port</span>
            <input name="http-port" placeholder="auto">
          </label>
          <label>
            gRPC Port
            <span class="label-caption">Internal gRPC port</span>
            <input name="grpc-port" placeholder="auto">
          </label>
        </div>
      </div>

      <div class="service-subsection db-section">
        <h4 class="subsection-title">Database Config</h4>
        <div class="grid">
          <label>
            Database
            <span class="label-caption">Select database engine</span>
            <select name="db-type">
              <option value="postgres" selected>PostgreSQL</option>
              <option value="mysql" disabled>MySQL (Coming Soon)</option>
              <option value="mongodb" disabled>MongoDB (Coming Soon)</option>
              <option value="sqlite" disabled>SQLite (Coming Soon)</option>
              <option value="mariadb" disabled>MariaDB (Coming Soon)</option>
            </select>
          </label>
          <label>
            DB Host
            <span class="label-caption">Database host</span>
            <input name="db-host" placeholder="localhost">
          </label>
          <label>
            DB Port
            <span class="label-caption">PostgreSQL port</span>
            <input name="db-port" placeholder="5432">
          </label>
          <label>
            DB Name
            <span class="label-caption">Target database name</span>
            <input name="db-name" placeholder="somedatabase">
          </label>
          <label>
            DB Username
            <span class="label-caption">Database username</span>
            <input name="db-user" placeholder="postgres">
          </label>
          <label>
            DB Password
            <span class="label-caption">Database password</span>
            <input name="db-password" placeholder="your-password">
          </label>
        </div>
      </div>
    `;
    card.querySelectorAll("input, select").forEach((inp) => {
      inp.addEventListener("input", renderProjectTree);
      inp.addEventListener("change", renderProjectTree);
    });
    servicesEl.appendChild(card);
  }
}

function applyServiceInputs(): void {
  const count = Number(countEl.value || 0);
  if (!Number.isInteger(count) || count < 0 || count > 20) {
    servicesEl.innerHTML = "";
    setStatus("Number Of New Services must be between 0 and 20.", true);
    return;
  }

  renderServiceInputs();
  const authCheckbox = document.getElementById(
    "include-auth-service",
  ) as HTMLInputElement | null;
  const isAuthEnabled = authCheckbox?.checked ?? false;

  if (count === 0) {
    if (isAuthEnabled) {
      setStatus("Auth-only project is ready to generate.");
    } else {
      setStatus("Add at least one extra service or include Auth Service.");
    }
    return;
  }

  setStatus(
    `${count} service input${count > 1 ? "s are" : " is"} ready to fill.`,
  );
}

function readValue(id: string): string {
  const element = document.getElementById(id) as HTMLInputElement | null;
  return element?.value.trim() ?? "";
}

function readDatabase(prefix: string): DatabaseConfig {
  const typeEl = document.getElementById(
    `${prefix}-db-type`,
  ) as HTMLSelectElement | null;
  return {
    type: typeEl?.value || "postgres",
    host: readValue(`${prefix}-db-host`),
    port: readValue(`${prefix}-db-port`),
    database: readValue(`${prefix}-db-name`),
    username: readValue(`${prefix}-db-user`),
    password: (
      document.getElementById(`${prefix}-db-password`) as HTMLInputElement
    ).value,
  };
}

function readServiceInputs(): ServiceInput[] {
  return Array.from(document.querySelectorAll(".service-card"))
    .filter((card) => card.querySelector('input[name="service-name"]'))
    .map((card) => ({
      name: (
        card.querySelector('input[name="service-name"]') as HTMLInputElement
      ).value.trim(),
      httpPort: (
        card.querySelector('input[name="http-port"]') as HTMLInputElement
      ).value.trim(),
      grpcPort: (
        card.querySelector('input[name="grpc-port"]') as HTMLInputElement
      ).value.trim(),
      db: {
        type:
          (
            card.querySelector('select[name="db-type"]') as HTMLSelectElement | null
          )?.value || "postgres",
        host: (
          card.querySelector('input[name="db-host"]') as HTMLInputElement
        ).value.trim(),
        port: (
          card.querySelector('input[name="db-port"]') as HTMLInputElement
        ).value.trim(),
        database: (
          card.querySelector('input[name="db-name"]') as HTMLInputElement
        ).value.trim(),
        username: (
          card.querySelector('input[name="db-user"]') as HTMLInputElement
        ).value.trim(),
        password: (
          card.querySelector('input[name="db-password"]') as HTMLInputElement
        ).value,
      },
    }))
    .filter((service) => service.name);
}

function normalizeDbConfig(
  label: string,
  db: DatabaseConfig,
  defaultDatabaseName = "somedatabase",
): DatabaseConfig {
  const normalized: DatabaseConfig = {
    type: db.type || "postgres",
    host: db.host.trim() || "localhost",
    port: db.port.trim() || "5432",
    database: db.database.trim() || defaultDatabaseName,
    username: db.username.trim() || "postgres",
    password: db.password,
  };

  if (!/^\d+$/.test(normalized.port)) {
    throw new Error(`${label} database port must be numeric`);
  }

  return normalized;
}

function parsePort(raw: string, label: string): number {
  if (!raw.trim()) {
    return 0;
  }

  if (!/^\d+$/.test(raw.trim())) {
    throw new Error(`${label} must be a number`);
  }

  return Number(raw.trim());
}

function nextUnusedPort(start: number, used: Set<number>): number {
  let port = start;
  while (used.has(port)) {
    port += 1;
  }
  return port;
}

function normalizeRequest(
  projectNameRaw: string,
  includeTraefik: boolean,
  traefikPortRaw: string,
  includeAuth: boolean,
  authHttpPortRaw: string,
  authGrpcPortRaw: string,
  authDbRaw: DatabaseConfig,
  servicesRaw: ServiceInput[],
): GenerationRequest {
  const projectName = projectNameRaw.trim()
    ? slugify(projectNameRaw)
    : "some-platform";
  if (!projectName) {
    throw new Error("Project name must contain letters or numbers.");
  }
  if (servicesRaw.length > 20) {
    throw new Error("Too many services requested.");
  }
  if (!includeAuth && servicesRaw.length === 0) {
    throw new Error(
      "Please add at least one extra service or include Auth Service.",
    );
  }

  const authDb = includeAuth
    ? normalizeDbConfig("auth-service", authDbRaw, "somedatabase")
    : {
        host: "localhost",
        port: "5432",
        database: "somedatabase",
        username: "postgres",
        password: "",
      };
  const traefikPort = includeTraefik
    ? parsePort(traefikPortRaw, "Traefik port") || 8000
    : 0;
  const authHttpPort =
    parsePort(authHttpPortRaw, "auth-service http port") || 7704;
  const authGrpcPort =
    parsePort(authGrpcPortRaw, "auth-service grpc port") || 57704;

  if (includeTraefik && includeAuth && traefikPort === authHttpPort) {
    throw new Error(`Traefik port already in use: ${traefikPort}`);
  }

  const usedHttp = new Set<number>([7705]);
  if (includeTraefik) {
    usedHttp.add(traefikPort);
  }
  const usedGrpc = new Set<number>([57705]);
  if (includeAuth) {
    usedHttp.add(7704);
    usedHttp.add(authHttpPort);
    usedGrpc.add(57704);
    usedGrpc.add(authGrpcPort);
  }
  let nextHttp = 7704;
  let nextGrpc = 57704;

  const services = servicesRaw.map((raw) => {
    const serviceName = slugify(raw.name);
    if (!serviceName) {
      throw new Error("Service names must contain letters or numbers.");
    }
    if (includeAuth && serviceName === "auth") {
      throw new Error(
        '"auth" is reserved because the project contains auth-service.',
      );
    }

    const db = normalizeDbConfig(
      `${serviceName}-service`,
      raw.db,
      `${serviceName}db`,
    );
    let httpPort = parsePort(raw.httpPort, `${serviceName} http port`);
    if (httpPort === 0) {
      nextHttp = nextUnusedPort(nextHttp + 1, usedHttp);
      httpPort = nextHttp;
    }
    if (usedHttp.has(httpPort)) {
      throw new Error(`HTTP port already in use: ${httpPort}`);
    }

    let grpcPort = parsePort(raw.grpcPort, `${serviceName} grpc port`);
    if (grpcPort === 0) {
      nextGrpc = nextUnusedPort(nextGrpc + 1, usedGrpc);
      grpcPort = nextGrpc;
    }
    if (usedGrpc.has(grpcPort)) {
      throw new Error(`gRPC port already in use: ${grpcPort}`);
    }

    usedHttp.add(httpPort);
    usedGrpc.add(grpcPort);
    return { name: serviceName, httpPort, grpcPort, db };
  });

  const uniqueNames = new Set(services.map((service) => service.name));
  if (uniqueNames.size !== services.length) {
    throw new Error("Duplicate service name in request.");
  }

  return {
    projectName,
    includeTraefik,
    traefikPort,
    includeAuth,
    authHttpPort,
    authGrpcPort,
    authDb,
    services,
  };
}

async function runWithConcurrency<T>(
  items: readonly T[],
  limit: number,
  worker: (item: T, index: number) => Promise<void>,
): Promise<void> {
  let cursor = 0;
  const runners = Array.from(
    { length: Math.min(limit, items.length) },
    async () => {
      while (cursor < items.length) {
        const current = cursor;
        cursor += 1;
        await worker(items[current], current);
      }
    },
  );
  await Promise.all(runners);
}

function rawFileUrl(path: string): string {
  const encoded = path.split("/").map(encodeURIComponent).join("/");
  return `${RAW_BASE}${encoded}`;
}

async function fetchTemplateFiles(
  updateStatus: (message: string) => void,
): Promise<Map<string, string>> {
  const files = new Map<string, string>();
  let completed = 0;

  await runWithConcurrency(TEMPLATE_FILE_PATHS, 8, async (path) => {
    const response = await fetch(rawFileUrl(path), { cache: "no-store" });
    if (!response.ok) {
      throw new Error(`Failed to fetch template file: ${path}`);
    }
    files.set(path, await response.text());
    completed += 1;
    updateStatus(
      `Fetching template files from GitHub (${completed}/${TEMPLATE_FILE_PATHS.length})...`,
    );
  });

  return files;
}

function getFile(files: Map<string, string>, path: string): string {
  const content = files.get(path);
  if (content === undefined) {
    throw new Error(`Missing template file: ${path}`);
  }
  return content;
}

function setFile(
  files: Map<string, string>,
  path: string,
  content: string,
): void {
  files.set(path, content);
}

function removeServiceBlock(text: string, serviceName: string): string {
  const lines = text.split("\n");
  const output: string[] = [];
  let skipping = false;

  for (const line of lines) {
    if (new RegExp(`^  ${serviceName}:\\s*$`).test(line)) {
      skipping = true;
      continue;
    }
    if (skipping && /^  [a-z0-9-]+:\s*$/.test(line)) {
      skipping = false;
    }
    if (!skipping) {
      output.push(line);
    }
  }

  return output.join("\n").replace(/\n{3,}/g, "\n\n");
}

function patchScaffoldScript(files: Map<string, string>): void {
  const path = "scripts/scaffold-service.sh";
  let content = getFile(files, path);
  if (!content.includes('source_dir="$repo_root/templates/service-template"')) {
    content = replaceAllLiteral(
      content,
      'source_dir="$repo_root/user-service"',
      'source_dir="$repo_root/templates/service-template"\nif [[ ! -d "$source_dir" ]]; then\n  source_dir="$repo_root/user-service"\nfi',
    );
  }
  if (content.includes('marker = "\\n  traefik:\\n"')) {
    content = replaceAllLiteral(
      content,
      'marker = "\\n  traefik:\\n"\nif marker not in text:\n    raise SystemExit("could not find traefik service block in docker-compose.yml")\ntext = text.replace(marker, block + marker, 1)',
      'marker = "\\n  traefik:\\n"\nif marker in text:\n    text = text.replace(marker, block + marker, 1)\nelse:\n    text = text.rstrip() + "\\n" + block + "\\n"',
    );
  }
  setFile(files, path, content);
}

function prepareServiceTemplate(files: Map<string, string>): void {
  const templatePrefix = "templates/service-template/";
  const hasTemplate = Array.from(files.keys()).some((path) =>
    path.startsWith(templatePrefix),
  );
  if (hasTemplate) {
    return;
  }

  for (const [path, content] of Array.from(files.entries())) {
    if (!path.startsWith("user-service/")) {
      continue;
    }
    files.set(path.replace(/^user-service\//, templatePrefix), content);
  }
}

function removeGeneratedUserService(files: Map<string, string>): void {
  for (const path of Array.from(files.keys())) {
    if (path.startsWith("user-service/")) {
      files.delete(path);
    }
  }

  let compose = getFile(files, "docker-compose.yml");
  compose = removeServiceBlock(compose, "user-service");
  compose = compose.replace(/\n\s+- user-service/g, "");
  setFile(files, "docker-compose.yml", compose);

  let routes = getFile(files, "traefik/dynamic/routes.yml");
  routes = routes.replace(
    "\n    user-api:\n      entryPoints:\n        - web\n      rule: PathPrefix(`/api/v1/user`)\n      middlewares:\n        - user-strip-v1\n        - user-addprefix\n        - protected-common\n      service: user-service\n",
    "",
  );
  routes = routes.replace(
    "\n    user-service:\n      loadBalancer:\n        servers:\n          - url: http://user-service:7705\n",
    "",
  );
  routes = routes.replace(
    "\n    user-strip-v1:\n      stripPrefix:\n        prefixes:\n          - /api/v1\n\n    user-addprefix:\n      addPrefix:\n        prefix: /api/user-service/v1\n",
    "",
  );
  setFile(files, "traefik/dynamic/routes.yml", routes);
}

function removeGeneratorWeb(files: Map<string, string>): void {
  let compose = getFile(files, "docker-compose.yml");
  compose = removeServiceBlock(compose, "generator-web");
  compose = compose.replace(/\n\s+- generator-web/g, "");
  setFile(files, "docker-compose.yml", compose);
}

function updateConfigSettings(
  files: Map<string, string>,
  path: string,
  httpPort: number,
  grpcPort: number,
  db: DatabaseConfig,
): void {
  const lines = getFile(files, path).split("\n");
  let inPostgres = false;

  for (let index = 0; index < lines.length; index += 1) {
    const line = lines[index];
    const trimmed = line.trim();

    if (trimmed.startsWith("port:") && line.startsWith("    ")) {
      lines[index] = `    port: "${httpPort}"`;
    }
    if (trimmed.startsWith("grpc_port:")) {
      lines[index] = `  grpc_port: "${grpcPort}"`;
    }
    if (trimmed === "postgres:") {
      inPostgres = true;
      continue;
    }
    if (inPostgres && !line.startsWith("  ") && trimmed !== "") {
      inPostgres = false;
    }
    if (!inPostgres) {
      continue;
    }

    if (trimmed.startsWith("host:")) {
      lines[index] = `  host: "${db.host}"`;
    } else if (trimmed.startsWith("port:")) {
      lines[index] = `  port: "${db.port}"`;
    } else if (trimmed.startsWith("database:")) {
      lines[index] = `  database: "${db.database}"`;
    } else if (trimmed.startsWith("user:")) {
      lines[index] = `  user: "${db.username}"`;
    } else if (trimmed.startsWith("password:")) {
      lines[index] = `  password: "${db.password}"`;
    }
  }

  setFile(files, path, lines.join("\n"));
}

function updateComposeServiceEnv(
  files: Map<string, string>,
  serviceName: string,
  httpPort: number,
  db: DatabaseConfig,
): void {
  const path = "docker-compose.yml";
  const lines = getFile(files, path).split("\n");
  let currentService = "";
  let inEnvironment = false;
  let inPorts = false;

  for (let index = 0; index < lines.length; index += 1) {
    const line = lines[index];
    const match = line.match(/^  ([a-z0-9-]+):\s*$/);
    if (match) {
      currentService = match[1];
      inEnvironment = false;
      inPorts = false;
      continue;
    }

    if (currentService !== serviceName) {
      continue;
    }

    const trimmed = line.trim();
    if (line.startsWith("    environment:")) {
      inEnvironment = true;
      inPorts = false;
      continue;
    }
    if (line.startsWith("    ports:")) {
      inPorts = true;
      inEnvironment = false;
      continue;
    }
    if (
      inEnvironment &&
      line.startsWith("    ") &&
      !line.startsWith("      ")
    ) {
      inEnvironment = false;
    }
    if (inPorts && line.startsWith("    ") && !line.startsWith("      ")) {
      inPorts = false;
    }
    if (!inEnvironment) {
      if (inPorts && line.startsWith('      - "')) {
        lines[index] = `      - "${httpPort}:${httpPort}"`;
        break;
      }
      if (line.startsWith('      - "')) {
        continue;
      }
      continue;
    }

    if (trimmed.startsWith("POSTGRES_HOST:")) {
      lines[index] = `      POSTGRES_HOST: "host.docker.internal"`;
    } else if (trimmed.startsWith("POSTGRES_PORT:")) {
      lines[index] = `      POSTGRES_PORT: "${db.port}"`;
    } else if (trimmed.startsWith("POSTGRES_DATABASE:")) {
      lines[index] = `      POSTGRES_DATABASE: "${db.database}"`;
    } else if (trimmed.startsWith("POSTGRES_USER:")) {
      lines[index] = `      POSTGRES_USER: "${db.username}"`;
    } else if (trimmed.startsWith("POSTGRES_PASSWORD:")) {
      lines[index] = `      POSTGRES_PASSWORD: "${db.password}"`;
    } else if (trimmed.startsWith("SERVER_HTTP_PORT:")) {
      lines[index] = `      SERVER_HTTP_PORT: "${httpPort}"`;
    }
  }

  setFile(
    files,
    path,
    ensureComposeServiceExtraHosts(lines.join("\n"), serviceName),
  );
}

function ensureComposeServiceExtraHosts(
  text: string,
  serviceName: string,
): string {
  const lines = text.split("\n");
  const output: string[] = [];
  let currentService = "";
  let inserted = false;
  let hasExtraHosts = false;

  for (let index = 0; index < lines.length; index += 1) {
    const line = lines[index];
    const match = line.match(/^  ([a-z0-9-]+):\s*$/);
    if (match) {
      if (currentService === serviceName && !inserted && !hasExtraHosts) {
        output.push("    extra_hosts:");
        output.push('      - "host.docker.internal:host-gateway"');
      }
      currentService = match[1];
      inserted = false;
      hasExtraHosts = false;
      output.push(line);
      continue;
    }

    if (currentService === serviceName) {
      if (line.trim() === "extra_hosts:") {
        hasExtraHosts = true;
      }
      if (!inserted && !hasExtraHosts && line.trim() === "ports:") {
        output.push("    extra_hosts:");
        output.push('      - "host.docker.internal:host-gateway"');
        inserted = true;
      }
    }

    output.push(line);
  }

  if (currentService === serviceName && !inserted && !hasExtraHosts) {
    output.push("    extra_hosts:");
    output.push('      - "host.docker.internal:host-gateway"');
  }

  return output.join("\n");
}

function updateServiceConfig(
  files: Map<string, string>,
  serviceName: string,
  httpPort: number,
  grpcPort: number,
  db: DatabaseConfig,
): void {
  updateConfigSettings(
    files,
    `${serviceName}/config.dev.yaml`,
    httpPort,
    grpcPort,
    db,
  );
  updateConfigSettings(
    files,
    `${serviceName}/config.local.yaml`,
    httpPort,
    grpcPort,
    db,
  );
  updateComposeServiceEnv(files, serviceName, httpPort, db);
}

function updateTraefikBasePorts(
  files: Map<string, string>,
  authHttpPort: number,
): void {
  let routes = getFile(files, "traefik/dynamic/routes.yml");
  routes = replaceAllLiteral(
    routes,
    "http://auth-service:7704",
    `http://auth-service:${authHttpPort}`,
  );
  setFile(files, "traefik/dynamic/routes.yml", routes);

  let middlewares = getFile(files, "traefik/dynamic/middlewares.yml");
  middlewares = replaceAllLiteral(
    middlewares,
    "http://auth-service:7704",
    `http://auth-service:${authHttpPort}`,
  );
  setFile(files, "traefik/dynamic/middlewares.yml", middlewares);
}

function updateTraefikPort(
  files: Map<string, string>,
  traefikPort: number,
): void {
  let compose = getFile(files, "docker-compose.yml");
  compose = replaceAllLiteral(
    compose,
    '- "8000:8000"',
    `- "${traefikPort}:${traefikPort}"`,
  );
  setFile(files, "docker-compose.yml", compose);

  let traefikConfig = getFile(files, "traefik/traefik.yml");
  traefikConfig = replaceAllLiteral(
    traefikConfig,
    'address: ":8000"',
    `address: ":${traefikPort}"`,
  );
  setFile(files, "traefik/traefik.yml", traefikConfig);
}

function appendComposeService(
  files: Map<string, string>,
  serviceName: string,
  httpPort: number,
  db: DatabaseConfig,
): void {
  const block = `
  ${serviceName}:
    build:
      context: ./${serviceName}
    container_name: ${serviceName}
    environment:
      POSTGRES_HOST: "host.docker.internal"
      POSTGRES_PORT: "${db.port}"
      POSTGRES_DATABASE: "${db.database}"
      POSTGRES_USER: "${db.username}"
      POSTGRES_PASSWORD: "${db.password}"
      POSTGRES_SCHEMA: public
      SERVER_HTTP_PORT: "${httpPort}"
    extra_hosts:
      - "host.docker.internal:host-gateway"
    ports:
      - "${httpPort}:${httpPort}"
`;
  let compose = getFile(files, "docker-compose.yml");
  if (compose.includes("\n  traefik:\n")) {
    compose = compose.replace("\n  traefik:\n", `${block}\n  traefik:\n`);
  } else {
    compose = `${compose.trimEnd()}\n${block}\n`;
  }
  setFile(files, "docker-compose.yml", compose);
}

function appendTraefikRoute(
  files: Map<string, string>,
  entity: string,
  serviceName: string,
  httpPort: number,
  includeAuth = true,
): void {
  let routes = getFile(files, "traefik/dynamic/routes.yml");
  const commonMiddleware = includeAuth ? "protected-common" : "public-common";
  const routerBlock = `
    ${entity}-api:
      entryPoints:
        - web
      rule: PathPrefix(\`/api/v1/${entity}\`)
      middlewares:
        - ${entity}-strip-v1
        - ${entity}-addprefix
        - ${commonMiddleware}
      service: ${serviceName}
`;
  const serviceBlock = `
    ${serviceName}:
      loadBalancer:
        servers:
          - url: http://${serviceName}:${httpPort}
`;
  const middlewareBlock = `
    ${entity}-strip-v1:
      stripPrefix:
        prefixes:
          - /api/v1

    ${entity}-addprefix:
      addPrefix:
        prefix: /api/${serviceName}/v1
`;
  routes = routes.replace("\n  services:\n", `${routerBlock}\n  services:\n`);
  routes = routes.replace(
    "\n  middlewares:\n",
    `${serviceBlock}\n  middlewares:\n`,
  );
  routes = `${routes.trimEnd()}\n${middlewareBlock}`;
  setFile(files, "traefik/dynamic/routes.yml", routes);
}

function scaffoldService(
  files: Map<string, string>,
  spec: ServiceSpec,
  includeAuth = true,
  includeTraefik = true,
): void {
  const entity = spec.name;
  const serviceName = `${entity}-service`;
  const entityPascal = pascalCase(entity);
  const entityPlural = `${entity}s`;
  const entityPluralPascal = `${entityPascal}s`;
  const entityUpper = entity.toUpperCase();
  const sourcePrefix = "templates/service-template/";
  const targetPrefix = `${serviceName}/`;

  const pathRenames: Record<string, string> = {
    "domain/user.go": `domain/${entity}.go`,
    "data/usecase/user.go": `data/usecase/${entity}.go`,
    "data/usecase/test/user_test.go": `data/usecase/test/${entity}_test.go`,
    "data/repository/postgres/user.go": `data/repository/postgres/${entity}.go`,
    "data/repository/postgres/test/user_test.go": `data/repository/postgres/test/${entity}_test.go`,
    "data/delivery/http/handler/user.go": `data/delivery/http/handler/${entity}.go`,
    "data/delivery/http/handler/test/user_test.go": `data/delivery/http/handler/test/${entity}_test.go`,
    "data/delivery/grpc/proto/user.proto": `data/delivery/grpc/proto/${entity}.proto`,
  };

  const copiedPaths: string[] = [];
  for (const [path, content] of Array.from(files.entries())) {
    if (!path.startsWith(sourcePrefix)) {
      continue;
    }
    if (path.includes("/.vscode/")) {
      continue;
    }

    const relativePath = path.slice(sourcePrefix.length);
    const renamedRelativePath = pathRenames[relativePath] ?? relativePath;
    const targetPath = `${targetPrefix}${renamedRelativePath}`;
    files.set(targetPath, content);
    copiedPaths.push(targetPath);
  }

  const specificReplacements: Array<[string, string]> = [
    ["user-service", serviceName],
    ["TestUserRepository", `Test${entityPascal}Repository`],
    ["TestUserUsecase", `Test${entityPascal}Usecase`],
    ["TestUserHandler", `Test${entityPascal}Handler`],
    ["TestUser", `Test${entityPascal}`],
    ["UserRepoTestCase", `${entityPascal}RepoTestCase`],
    ["CreateUserRequest", `Create${entityPascal}Request`],
    ["UpdateUserRequest", `Update${entityPascal}Request`],
    ["GetUserByIdRequest", `Get${entityPascal}ByIdRequest`],
    ["DeleteUserRequest", `Delete${entityPascal}Request`],
    ["DeleteUserResponse", `Delete${entityPascal}Response`],
    ["UserListResponse", `${entityPascal}ListResponse`],
    ["UserResponse", `${entityPascal}Response`],
    ["UserHandler", `${entityPascal}Handler`],
    ["UserUsecase", `${entityPascal}Usecase`],
    ["UserRepository", `${entityPascal}Repository`],
    ["UserRepo", `${entityPascal}Repo`],
    ["UserService", `${entityPascal}Service`],
    ["NewUserHandler", `New${entityPascal}Handler`],
    ["NewUserUsecase", `New${entityPascal}Usecase`],
    ["NewUserRepo", `New${entityPascal}Repo`],
    ["GetAllUsers", `GetAll${entityPluralPascal}`],
    ["GetUserById", `Get${entityPascal}ById`],
    ["CreateUser", `Create${entityPascal}`],
    ["UpdateUser", `Update${entityPascal}`],
    ["DeleteUser", `Delete${entityPascal}`],
    ["/user", `/${entity}`],
    ['"users"', `"${entityPlural}"`],
    [" public.users ", ` public.${entityPlural} `],
    ["user endpoints", `${entity} endpoints`],
    ["Create user", `Create ${entity}`],
    ["Get all users", `Get all ${entityPlural}`],
    ["Get user by ID", `Get ${entity} by ID`],
    ["Update user by ID", `Update ${entity} by ID`],
    ["Delete user by ID", `Delete ${entity} by ID`],
    ["User Routes", `${entityPascal} Routes`],
  ];

  const entityFiles = new Set([
    `${targetPrefix}domain/errors.go`,
    `${targetPrefix}domain/${entity}.go`,
    `${targetPrefix}data/usecase/${entity}.go`,
    `${targetPrefix}data/usecase/test/${entity}_test.go`,
    `${targetPrefix}data/repository/postgres/${entity}.go`,
    `${targetPrefix}data/repository/postgres/test/${entity}_test.go`,
    `${targetPrefix}data/delivery/http/handler/${entity}.go`,
    `${targetPrefix}data/delivery/http/handler/test/${entity}_test.go`,
    `${targetPrefix}data/delivery/grpc/proto/${entity}.proto`,
    `${targetPrefix}openapi.yaml`,
  ]);

  for (const path of copiedPaths) {
    let content = getFile(files, path);
    for (const [from, to] of specificReplacements) {
      content = replaceAllLiteral(content, from, to);
    }
    if (entityFiles.has(path)) {
      content = replaceAllLiteral(content, "Users", entityPluralPascal);
      content = replaceAllLiteral(content, "User", entityPascal);
    }
    if (path === `${targetPrefix}config/struct.go`) {
      content = replaceAllLiteral(
        content,
        'ClientID:          "user",',
        `ClientID:          "${entity}",`,
      );
    }
    if (path === `${targetPrefix}Dockerfile`) {
      content = replaceAllLiteral(
        content,
        "EXPOSE 7705",
        `EXPOSE ${spec.httpPort}`,
      );
    }
    if (path === `${targetPrefix}openapi.yaml`) {
      content = replaceAllLiteral(
        content,
        "url: http://localhost:7704/api/",
        `url: http://localhost:${spec.httpPort}/api/`,
      );
    }
    if (path === `${targetPrefix}domain/errors.go`) {
      content = replaceAllLiteral(
        content,
        'const StatusCodePrefix = "USER"',
        `const StatusCodePrefix = "${entityUpper}"`,
      );
    }
    setFile(files, path, content);
  }

  appendComposeService(files, serviceName, spec.httpPort, spec.db);
  if (includeTraefik) {
    appendTraefikRoute(files, entity, serviceName, spec.httpPort, includeAuth);
  }
  updateServiceConfig(
    files,
    serviceName,
    spec.httpPort,
    spec.grpcPort,
    spec.db,
  );
}

function removeGeneratedAuthService(files: Map<string, string>): void {
  for (const path of Array.from(files.keys())) {
    if (path.startsWith("auth-service/")) {
      files.delete(path);
    }
  }

  let compose = getFile(files, "docker-compose.yml");
  compose = removeServiceBlock(compose, "auth-service");
  compose = compose.replace(/\n\s+- auth-service/g, "");
  compose = compose.replace(/\n\s+depends_on:\s*\n(?=\s+[a-z])/g, "\n");
  setFile(files, "docker-compose.yml", compose);

  let routes = getFile(files, "traefik/dynamic/routes.yml");
  routes = routes.replace(
    "\n    auth-login:\n      entryPoints:\n        - web\n      rule: Path(`/api/v1/login`)\n      middlewares:\n        - auth-strip-v1\n        - auth-addprefix\n        - public-common\n      service: auth-service\n",
    "",
  );
  routes = routes.replace(
    "\n    auth-api:\n      entryPoints:\n        - web\n      rule: PathPrefix(`/api/v1/auth`)\n      middlewares:\n        - auth-strip-v1\n        - auth-addprefix\n        - public-common\n      service: auth-service\n",
    "",
  );
  routes = routes.replace(
    "\n    auth-service:\n      loadBalancer:\n        servers:\n          - url: http://auth-service:7704\n",
    "",
  );
  routes = routes.replace(
    "\n    auth-strip-v1:\n      stripPrefix:\n        prefixes:\n          - /api/v1\n\n    auth-addprefix:\n      addPrefix:\n        prefix: /api/auth-service/v1\n",
    "",
  );
  routes = routes.replace("  services:\n  middlewares:", "  services:\n\n  middlewares:\n");
  setFile(files, "traefik/dynamic/routes.yml", routes);
}

function removeTraefik(files: Map<string, string>): void {
  for (const path of Array.from(files.keys())) {
    if (path.startsWith("traefik/")) {
      files.delete(path);
    }
  }

  let compose = getFile(files, "docker-compose.yml");
  compose = removeServiceBlock(compose, "traefik");
  compose = compose.replace(/\n\s+- traefik/g, "");
  setFile(files, "docker-compose.yml", compose);
}

function cleanEmptyDependsOn(text: string): string {
  return text.replace(/\n\s+depends_on:\s*\n(?=\s+[a-z])/g, "\n");
}

async function buildProjectZip(
  request: GenerationRequest,
  updateStatus: (message: string) => void,
): Promise<Blob> {
  const files = await fetchTemplateFiles(updateStatus);

  patchScaffoldScript(files);
  prepareServiceTemplate(files);
  removeGeneratedUserService(files);
  removeGeneratorWeb(files);

  if (request.includeAuth) {
    updateServiceConfig(
      files,
      "auth-service",
      request.authHttpPort,
      request.authGrpcPort,
      request.authDb,
    );
    if (request.includeTraefik) {
      updateTraefikBasePorts(files, request.authHttpPort);
    }
  } else {
    removeGeneratedAuthService(files);
  }

  for (const service of request.services) {
    updateStatus(`Scaffolding ${service.name}-service...`);
    scaffoldService(
      files,
      service,
      request.includeAuth,
      request.includeTraefik,
    );
  }

  if (request.includeTraefik) {
    updateTraefikPort(files, request.traefikPort);
  } else {
    removeTraefik(files);
  }

  let compose = getFile(files, "docker-compose.yml");
  compose = cleanEmptyDependsOn(compose);
  setFile(files, "docker-compose.yml", compose);

  updateStatus("Building zip archive...");
  const zip = new JSZip();
  for (const [path, content] of Array.from(files.entries()).sort(([a], [b]) =>
    a.localeCompare(b),
  )) {
    const topLevel = path.split("/")[0];
    if (EXCLUDED_TOP_LEVEL.has(topLevel)) {
      continue;
    }

    zip.file(`${request.projectName}/${path}`, content, {
      unixPermissions: path.endsWith(".sh") ? 0o755 : 0o644,
    });
  }

  return zip.generateAsync({ type: "blob", compression: "DEFLATE" });
}

function downloadBlob(filename: string, blob: Blob): void {
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = filename;
  document.body.appendChild(link);
  link.click();
  link.remove();
  URL.revokeObjectURL(url);
}

async function generateProject(event: SubmitEvent): Promise<void> {
  event.preventDefault();
  setStatus("");
  buttonEl.disabled = true;

  try {
    const includeAuth =
      (
        document.getElementById(
          "include-auth-service",
        ) as HTMLInputElement | null
      )?.checked ?? false;

    const includeTraefik =
      (
        document.getElementById(
          "include-traefik",
        ) as HTMLInputElement | null
      )?.checked ?? false;

    const request = normalizeRequest(
      readValue("project-name"),
      includeTraefik,
      readValue("traefik-port"),
      includeAuth,
      readValue("auth-http-port"),
      readValue("auth-grpc-port"),
      readDatabase("auth"),
      readServiceInputs(),
    );

    setStatus("Preparing template...");
    const zipBlob = await buildProjectZip(request, setStatus);
    downloadBlob(`${request.projectName}.zip`, zipBlob);
    setStatus("Project generated. Download should start automatically.");
  } catch (error) {
    const message =
      error instanceof Error ? error.message : "Generation failed.";
    setStatus(message, true);
  } finally {
    buttonEl.disabled = false;
  }
}

const baseServiceSection = document.getElementById("base-service-section");
const extraServicesSection = document.getElementById("extra-services-section");
const readyToGenerateCard = document.getElementById("ready-to-generate-card");

const includeAuthCheckbox = document.getElementById(
  "include-auth-service",
) as HTMLInputElement | null;
const baseServiceContainer = document.getElementById(
  "base-service-container",
) as HTMLDivElement | null;

includeAuthCheckbox?.addEventListener("change", () => {
  const isChecked = includeAuthCheckbox.checked;
  if (baseServiceSection) {
    if (isChecked) {
      baseServiceSection.classList.remove("hidden");
    } else {
      baseServiceSection.classList.add("hidden");
    }
  }
  if (baseServiceContainer) {
    baseServiceContainer.style.display = isChecked ? "" : "none";
  }
  const count = Number(countEl?.value || 0);
  if (isChecked || count > 0) {
    readyToGenerateCard?.classList.remove("hidden");
  } else {
    const extraCards = document.querySelectorAll(".extra-service-card");
    if (extraCards.length === 0) {
      readyToGenerateCard?.classList.add("hidden");
    }
  }
  renderProjectTree();
});

const includeTraefikCheckbox = document.getElementById(
  "include-traefik",
) as HTMLInputElement | null;
const traefikPortContainer = document.getElementById(
  "traefik-port-container",
) as HTMLDivElement | null;

includeTraefikCheckbox?.addEventListener("change", () => {
  if (traefikPortContainer) {
    traefikPortContainer.style.display = includeTraefikCheckbox.checked
      ? ""
      : "none";
  }
  renderProjectTree();
});

// Initialize initial display states
if (baseServiceSection && includeAuthCheckbox) {
  if (includeAuthCheckbox.checked) {
    baseServiceSection.classList.remove("hidden");
  } else {
    baseServiceSection.classList.add("hidden");
  }
}
if (baseServiceContainer && includeAuthCheckbox) {
  baseServiceContainer.style.display = includeAuthCheckbox.checked
    ? ""
    : "none";
}
if (traefikPortContainer && includeTraefikCheckbox) {
  traefikPortContainer.style.display = includeTraefikCheckbox.checked
    ? ""
    : "none";
}

function showConfigSections(): void {
  const isAuthEnabled = includeAuthCheckbox?.checked ?? false;
  if (isAuthEnabled) {
    baseServiceSection?.classList.remove("hidden");
  }
  extraServicesSection?.classList.remove("hidden");
  readyToGenerateCard?.classList.remove("hidden");
}

function getTreeData(): {
  projectName: string;
  includeTraefik: boolean;
  traefikPort: string;
  includeAuth: boolean;
  authHttpPort: string;
  services: Array<{ name: string; httpPort: string }>;
} {
  const rawName = readValue("project-name");
  const projectName = rawName ? slugify(rawName) : "some-platform";

  const includeTraefik =
    (document.getElementById("include-traefik") as HTMLInputElement | null)
      ?.checked ?? false;
  const traefikPort = readValue("traefik-port") || "8000";

  const includeAuth =
    (document.getElementById("include-auth-service") as HTMLInputElement | null)
      ?.checked ?? false;
  const authHttpPort = readValue("auth-http-port") || "7704";

  const extraCards = Array.from(
    document.querySelectorAll(".extra-service-card"),
  );
  const services: Array<{ name: string; httpPort: string }> = [];

  if (extraCards.length > 0) {
    extraCards.forEach((card, idx) => {
      const nameInput = card.querySelector(
        'input[name="service-name"]',
      ) as HTMLInputElement | null;
      const portInput = card.querySelector(
        'input[name="http-port"]',
      ) as HTMLInputElement | null;
      const rawServiceName = nameInput?.value.trim() ?? "";
      const name = rawServiceName
        ? slugify(rawServiceName)
        : `service-${idx + 1}`;
      const port = portInput?.value.trim() || "auto";
      services.push({ name, httpPort: port });
    });
  } else {
    const count = Math.max(0, Math.min(20, Number(countEl?.value || 0)));
    for (let i = 0; i < count; i += 1) {
      services.push({ name: `service-${i + 1}`, httpPort: "auto" });
    }
  }

  return {
    projectName,
    includeTraefik,
    traefikPort,
    includeAuth,
    authHttpPort,
    services,
  };
}

function renderProjectTree(): void {
  const container = document.getElementById("project-tree-container");
  if (!container) return;

  const {
    projectName,
    includeTraefik,
    traefikPort,
    includeAuth,
    authHttpPort,
    services,
  } = getTreeData();

  const rootLabel = document.getElementById("tree-root-label");
  if (rootLabel) {
    rootLabel.textContent = `${projectName}/`;
  }

  const totalServices = (includeAuth ? 1 : 0) + services.length;
  const estFiles =
    30 +
    (includeTraefik ? 4 : 0) +
    (includeAuth ? 28 : 0) +
    services.length * 28;
  const statsBadge = document.getElementById("structure-stats-badge");
  if (statsBadge) {
    statsBadge.textContent = `${totalServices} service${totalServices === 1 ? "" : "s"}${includeTraefik ? " + Traefik" : ""} • ~${estFiles} files`;
  }

  const line = (
    guide: string,
    icon: string,
    name: string,
    badgeType = "",
    badgeLabel = "",
    note = "",
    isFolder = false,
    isRoot = false,
  ) => {
    const nameClass = isRoot
      ? "tree-root-folder"
      : isFolder
        ? "tree-folder"
        : "tree-file";
    const badgeHtml = badgeType
      ? `<span class="tree-badge ${badgeType}">${badgeLabel}</span>`
      : "";
    const noteHtml = note ? `<span class="tree-note">${note}</span>` : "";
    return `<div class="tree-line"><span class="tree-guide">${guide}</span><span class="tree-icon">${icon}</span><span class="${nameClass}">${name}</span>${badgeHtml}${noteHtml}</div>`;
  };

  const lines: string[] = [];
  lines.push(line("", "📁", `${projectName}/`, "", "", "", true, true));
  lines.push(line("├── ", "📁", "scripts/", "", "", "", true));
  lines.push(
    line(
      "│   └── ",
      "📄",
      "scaffold-service.sh",
      "",
      "",
      "CLI service generator",
    ),
  );
  lines.push(line("├── ", "📁", "templates/", "", "", "", true));
  lines.push(
    line(
      "│   └── ",
      "📁",
      "service-template/",
      "template",
      "generator source",
      "clean architecture blueprint",
      true,
    ),
  );

  if (includeTraefik) {
    lines.push(
      line(
        "├── ",
        "📁",
        "traefik/",
        "gateway",
        `gateway :${traefikPort}`,
        "",
        true,
      ),
    );
    lines.push(line("│   ├── ", "📁", "dynamic/", "", "", "", true));
    lines.push(line("│   │   ├── ", "📄", "middlewares.yml"));
    lines.push(line("│   │   └── ", "📄", "routes.yml"));
    lines.push(line("│   ├── ", "📄", "traefik.yml"));
    lines.push(line("│   └── ", "📄", "README.md"));
  }

  if (includeAuth) {
    lines.push(
      line(
        "├── ",
        "📁",
        "auth-service/",
        "auth",
        `auth HTTP :${authHttpPort}`,
        "",
        true,
      ),
    );
    lines.push(line("│   ├── ", "📁", "config/", "", "", "", true));
    lines.push(line("│   ├── ", "📁", "data/", "", "", "", true));
    lines.push(
      line("│   │   ├── ", "📁", "delivery/", "", "", "http & grpc", true),
    );
    lines.push(
      line("│   │   │   ├── ", "📁", "http/handler/", "", "", "", true),
    );
    lines.push(
      line("│   │   │   │   ├── ", "📁", "test/", "", "", "handler test", true),
    );
    lines.push(line("│   │   │   │   │   └── ", "📄", "auth_test.go"));
    lines.push(line("│   │   │   │   └── ", "📄", "auth.go"));
    lines.push(
      line("│   │   │   └── ", "📁", "grpc/proto/", "", "", "", true),
    );
    lines.push(line("│   │   ├── ", "📁", "repository/", "", "", "", true));
    lines.push(
      line("│   │   │   └── ", "📁", "postgres/", "", "", "", true),
    );
    lines.push(
      line("│   │   │       ├── ", "📁", "test/", "", "", "repo test", true),
    );
    lines.push(line("│   │   │       │   └── ", "📄", "auth_test.go"));
    lines.push(line("│   │   │       └── ", "📄", "auth.go"));
    lines.push(line("│   │   └── ", "📁", "usecase/", "", "", "", true));
    lines.push(
      line("│   │       ├── ", "📁", "test/", "", "", "usecase test", true),
    );
    lines.push(line("│   │       │   └── ", "📄", "auth_test.go"));
    lines.push(line("│   │       └── ", "📄", "auth.go"));
    lines.push(
      line("│   ├── ", "📁", "domain/", "", "", "entities & errors", true),
    );
    lines.push(line("│   ├── ", "📄", "config.dev.yaml"));
    lines.push(line("│   ├── ", "📄", "Dockerfile"));
    lines.push(line("│   ├── ", "📄", "go.mod"));
    lines.push(line("│   └── ", "📄", "main.go"));
  }

  services.forEach((s) => {
    const sName = `${s.name}-service`;
    lines.push(
      line(
        "├── ",
        "📁",
        `${sName}/`,
        "service",
        `service :${s.httpPort}`,
        "",
        true,
      ),
    );
    lines.push(line("│   ├── ", "📁", "config/", "", "", "", true));
    lines.push(line("│   ├── ", "📁", "data/", "", "", "", true));
    lines.push(
      line("│   │   ├── ", "📁", "delivery/", "", "", "http & grpc", true),
    );
    lines.push(
      line("│   │   │   ├── ", "📁", "http/handler/", "", "", "", true),
    );
    lines.push(
      line("│   │   │   │   ├── ", "📁", "test/", "", "", "handler test", true),
    );
    lines.push(line("│   │   │   │   │   └── ", "📄", `${s.name}_test.go`));
    lines.push(line("│   │   │   │   └── ", "📄", `${s.name}.go`));
    lines.push(
      line("│   │   │   └── ", "📁", "grpc/proto/", "", "", "", true),
    );
    lines.push(line("│   │   ├── ", "📁", "repository/", "", "", "", true));
    lines.push(
      line("│   │   │   └── ", "📁", "postgres/", "", "", "", true),
    );
    lines.push(
      line("│   │   │       ├── ", "📁", "test/", "", "", "repo test", true),
    );
    lines.push(line("│   │   │       │   └── ", "📄", `${s.name}_test.go`));
    lines.push(line("│   │   │       └── ", "📄", `${s.name}.go`));
    lines.push(line("│   │   └── ", "📁", "usecase/", "", "", "", true));
    lines.push(
      line("│   │       ├── ", "📁", "test/", "", "", "usecase test", true),
    );
    lines.push(line("│   │       │   └── ", "📄", `${s.name}_test.go`));
    lines.push(line("│   │       └── ", "📄", `${s.name}.go`));
    lines.push(
      line("│   ├── ", "📁", "domain/", "", "", "entities & errors", true),
    );
    lines.push(line("│   ├── ", "📄", "config.dev.yaml"));
    lines.push(line("│   ├── ", "📄", "Dockerfile"));
    lines.push(line("│   ├── ", "📄", "go.mod"));
    lines.push(line("│   ├── ", "📄", "openapi.yaml"));
    lines.push(line("│   └── ", "📄", "main.go"));
  });

  lines.push(
    line("├── ", "📄", "docker-compose.yml", "", "", "orchestration"),
  );
  lines.push(line("└── ", "📄", "README.md", "", "", "quick start guide"));

  container.innerHTML = lines.join("");
}

function generateAsciiTree(): string {
  const {
    projectName,
    includeTraefik,
    traefikPort,
    includeAuth,
    authHttpPort,
    services,
  } = getTreeData();

  const lines: string[] = [];
  lines.push(`${projectName}/`);
  lines.push("├── scripts/");
  lines.push("│   └── scaffold-service.sh");
  lines.push("├── templates/");
  lines.push("│   └── service-template/");

  if (includeTraefik) {
    lines.push(`├── traefik/                       [gateway :${traefikPort}]`);
    lines.push("│   ├── dynamic/");
    lines.push("│   │   ├── middlewares.yml");
    lines.push("│   │   └── routes.yml");
    lines.push("│   ├── traefik.yml");
    lines.push("│   └── README.md");
  }

  if (includeAuth) {
    lines.push(
      `├── auth-service/                  [core auth :${authHttpPort}]`,
    );
    lines.push("│   ├── config/");
    lines.push("│   ├── data/");
    lines.push("│   │   ├── delivery/");
    lines.push("│   │   │   ├── http/handler/");
    lines.push("│   │   │   │   ├── test/");
    lines.push("│   │   │   │   │   └── auth_test.go");
    lines.push("│   │   │   │   └── auth.go");
    lines.push("│   │   │   └── grpc/proto/");
    lines.push("│   │   ├── repository/");
    lines.push("│   │   │   └── postgres/");
    lines.push("│   │   │       ├── test/");
    lines.push("│   │   │       │   └── auth_test.go");
    lines.push("│   │   │       └── auth.go");
    lines.push("│   │   └── usecase/");
    lines.push("│   │       ├── test/");
    lines.push("│   │       │   └── auth_test.go");
    lines.push("│   │       └── auth.go");
    lines.push("│   ├── domain/");
    lines.push("│   ├── config.dev.yaml");
    lines.push("│   ├── Dockerfile");
    lines.push("│   ├── go.mod");
    lines.push("│   └── main.go");
  }

  services.forEach((s) => {
    lines.push(
      `├── ${s.name}-service/             [service :${s.httpPort}]`,
    );
    lines.push("│   ├── config/");
    lines.push("│   ├── data/");
    lines.push("│   │   ├── delivery/");
    lines.push("│   │   │   ├── http/handler/");
    lines.push("│   │   │   │   ├── test/");
    lines.push(`│   │   │   │   │   └── ${s.name}_test.go`);
    lines.push(`│   │   │   │   └── ${s.name}.go`);
    lines.push("│   │   │   └── grpc/proto/");
    lines.push("│   │   ├── repository/");
    lines.push("│   │   │   └── postgres/");
    lines.push("│   │   │       ├── test/");
    lines.push(`│   │   │       │   └── ${s.name}_test.go`);
    lines.push(`│   │   │       └── ${s.name}.go`);
    lines.push("│   │   └── usecase/");
    lines.push("│   │       ├── test/");
    lines.push(`│   │       │   └── ${s.name}_test.go`);
    lines.push(`│   │       └── ${s.name}.go`);
    lines.push("│   ├── domain/");
    lines.push("│   ├── config.dev.yaml");
    lines.push("│   ├── Dockerfile");
    lines.push("│   ├── go.mod");
    lines.push("│   ├── openapi.yaml");
    lines.push("│   └── main.go");
  });

  lines.push("├── docker-compose.yml");
  lines.push("└── README.md");

  return lines.join("\n");
}

const copyTreeBtn = document.getElementById("copy-tree-button");
copyTreeBtn?.addEventListener("click", () => {
  const ascii = generateAsciiTree();
  void navigator.clipboard.writeText(ascii).then(() => {
    const textEl = document.getElementById("copy-tree-text");
    if (textEl) {
      const orig = textEl.textContent;
      textEl.textContent = "Copied!";
      setTimeout(() => {
        textEl.textContent = orig;
      }, 2000);
    }
  });
});

document
  .getElementById("project-name")
  ?.addEventListener("input", renderProjectTree);
document
  .getElementById("traefik-port")
  ?.addEventListener("input", renderProjectTree);
document
  .getElementById("auth-http-port")
  ?.addEventListener("input", renderProjectTree);
document
  .getElementById("service-count")
  ?.addEventListener("input", renderProjectTree);

const projectDbSelect = document.getElementById(
  "project-database",
) as HTMLSelectElement | null;
projectDbSelect?.addEventListener("change", () => {
  const val = projectDbSelect.value;
  const authDbSelect = document.getElementById(
    "auth-db-type",
  ) as HTMLSelectElement | null;
  if (authDbSelect) {
    authDbSelect.value = val;
  }
  document
    .querySelectorAll<HTMLSelectElement>('select[name="db-type"]')
    .forEach((sel) => {
      sel.value = val;
    });
  renderProjectTree();
});

document
  .getElementById("auth-db-type")
  ?.addEventListener("change", renderProjectTree);

applyServicesButtonEl.addEventListener("click", () => {
  showConfigSections();
  applyServiceInputs();
  renderProjectTree();
});

formEl.addEventListener("submit", (event) => {
  void generateProject(event as SubmitEvent);
});
servicesEl.innerHTML = "";
renderProjectTree();

// Theme toggle
const themeToggleBtn = document.getElementById("theme-toggle-btn");
const themeIconSun = document.getElementById("theme-icon-sun");
const themeIconMoon = document.getElementById("theme-icon-moon");

function updateThemeUI(isDark: boolean): void {
  if (isDark) {
    document.documentElement.classList.add("dark");
    themeIconSun?.classList.remove("hidden");
    themeIconMoon?.classList.add("hidden");
  } else {
    document.documentElement.classList.remove("dark");
    themeIconSun?.classList.add("hidden");
    themeIconMoon?.classList.remove("hidden");
  }
}

const savedTheme = localStorage.getItem("theme");
const prefersDark = window.matchMedia("(prefers-color-scheme: dark)").matches;
const isInitialDark = savedTheme === "dark" || (!savedTheme && prefersDark);
updateThemeUI(isInitialDark);

themeToggleBtn?.addEventListener("click", () => {
  const isDark = document.documentElement.classList.contains("dark");
  const nextTheme = isDark ? "light" : "dark";
  localStorage.setItem("theme", nextTheme);
  updateThemeUI(!isDark);
});
