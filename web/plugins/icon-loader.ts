import { readFileSync, existsSync, readdirSync } from "node:fs";
import { join } from "node:path";
import type { Plugin } from "vite";

const iconMap: Record<string, Set<string>> = {};
const jsonCache: Record<string, any> = {};
const collectionExistsCache: Record<string, boolean> = {};
let installedCollectionsCache: Set<string> | null = null;

const sourceFileRe = /\.(vue|[cm]?[jt]sx?)$/;
const iconColonRe = /\bi-([a-z0-9-]+):([a-z0-9-]+)\b/g;
const iconTokenRe = /\bi-([a-z0-9-]+)\b/g;

type IconWhitelist = Record<string, string[]>;

type ProductionOptions = {
  pruneIcons?: boolean;
  keepAliases?: boolean;
  dropMeta?: boolean;
};

type IconifyIconsOptions = {
  whitelist?: IconWhitelist;
  production?: ProductionOptions;
  whitelistCheck?: {
    enabled?: boolean;
    throwOnMissing?: boolean;
  };
};

const defaultProductionOptions: Required<ProductionOptions> = {
  pruneIcons: true,
  keepAliases: true,
  dropMeta: true,
};

const defaultWhitelistCheckOptions = {
  enabled: true,
  throwOnMissing: false,
};

export default function IconifyIcons(
  options: IconifyIconsOptions = {},
): Plugin {
  const name = "load-iconify-icons";
  const virtualModuleId = `virtual:${name}`;
  const resolvedVirtualModuleId = `\0${virtualModuleId}`;

  let isBuild = false;
  let productionOptions: Required<ProductionOptions> = defaultProductionOptions;
  let whitelistCheckOptions = defaultWhitelistCheckOptions;
  let orderedCollectionPrefixes: string[] = [];

  function getCollectionsRootPath() {
    return join(process.cwd(), "node_modules/@iconify-json");
  }

  function getInstalledCollections() {
    if (installedCollectionsCache) return installedCollectionsCache;

    const collections = new Set<string>();
    const root = getCollectionsRootPath();
    if (!existsSync(root)) {
      installedCollectionsCache = collections;
      return collections;
    }

    for (const entry of readdirSync(root, { withFileTypes: true })) {
      if (!entry.isDirectory() && !entry.isSymbolicLink()) continue;
      if (entry.name.startsWith(".")) continue;

      const collectionName = entry.name;
      const collectionPath = join(root, collectionName, "icons.json");
      if (existsSync(collectionPath)) {
        collections.add(collectionName);
      }
    }

    installedCollectionsCache = collections;
    return collections;
  }

  function getCollectionPath(prefix: string) {
    return join(
      process.cwd(),
      `node_modules/@iconify-json/${prefix}/icons.json`,
    );
  }

  function hasCollection(prefix: string) {
    if (!getInstalledCollections().has(prefix)) {
      collectionExistsCache[prefix] = false;
      return false;
    }

    if (collectionExistsCache[prefix] !== undefined) {
      return collectionExistsCache[prefix];
    }
    const exists = existsSync(getCollectionPath(prefix));
    collectionExistsCache[prefix] = exists;
    return exists;
  }

  function clearIconMap() {
    for (const prefix of Object.keys(iconMap)) {
      delete iconMap[prefix];
    }
  }

  function addIcon(prefix: string, iconName: string) {
    if (!hasCollection(prefix)) return;
    if (!iconMap[prefix]) iconMap[prefix] = new Set();
    iconMap[prefix].add(iconName);
  }

  function applyWhitelist() {
    const whitelist = options.whitelist;
    if (!whitelist) return;

    for (const [prefix, iconNames] of Object.entries(whitelist)) {
      if (!hasCollection(prefix)) continue;

      for (const iconName of iconNames) {
        addIcon(prefix, iconName);
      }
    }
  }

  function collectWhitelistMissingEntries() {
    const missing: string[] = [];
    const whitelist = options.whitelist;
    if (!whitelist) return missing;

    for (const [prefix, iconNames] of Object.entries(whitelist)) {
      const fullSet = getFullSet(prefix);
      if (!fullSet) {
        missing.push(
          `set '${prefix}' is missing (install @iconify-json/${prefix})`,
        );
        continue;
      }

      for (const iconName of new Set(iconNames)) {
        const existsInIcons = Boolean(fullSet.icons?.[iconName]);
        const existsInAliases = Boolean(fullSet.aliases?.[iconName]);
        if (!existsInIcons && !existsInAliases) {
          missing.push(`${prefix}:${iconName}`);
        }
      }
    }

    return missing;
  }

  function resolveCollectionAndName(tokenBody: string) {
    for (const prefix of orderedCollectionPrefixes) {
      const prefixToken = `${prefix}-`;
      if (!tokenBody.startsWith(prefixToken)) continue;

      const iconName = tokenBody.slice(prefixToken.length);
      if (iconName.length > 0) {
        return { prefix, iconName };
      }
    }

    return null;
  }

  function collectIconsFromCode(code: string) {
    let colonMatch: RegExpExecArray | null;
    while ((colonMatch = iconColonRe.exec(code)) !== null) {
      const [, prefix, iconName] = colonMatch;
      addIcon(prefix, iconName);
    }
    iconColonRe.lastIndex = 0;

    let match: RegExpExecArray | null;
    while ((match = iconTokenRe.exec(code)) !== null) {
      const resolved = resolveCollectionAndName(match[1]);
      if (!resolved) continue;
      addIcon(resolved.prefix, resolved.iconName);
    }
    iconTokenRe.lastIndex = 0;
  }

  function walkSourceFiles(dir: string): string[] {
    const files: string[] = [];
    const stack = [dir];

    while (stack.length > 0) {
      const current = stack.pop();
      if (!current) continue;

      for (const entry of readdirSync(current, { withFileTypes: true })) {
        const absPath = join(current, entry.name);

        if (entry.isDirectory()) {
          if (entry.name === "node_modules" || entry.name.startsWith(".")) {
            continue;
          }
          stack.push(absPath);
          continue;
        }

        if (sourceFileRe.test(absPath)) {
          files.push(absPath);
        }
      }
    }

    return files;
  }

  function scanWorkspaceSource() {
    const srcDir = join(process.cwd(), "src");
    if (!existsSync(srcDir)) return;

    for (const filePath of walkSourceFiles(srcDir)) {
      try {
        const code = readFileSync(filePath, "utf8");
        collectIconsFromCode(code);
      } catch {
        // 忽略偶发读取失败（比如文件正在写入）
      }
    }
  }

  function getFullSet(prefix: string) {
    if (jsonCache[prefix]) return jsonCache[prefix];
    const path = getCollectionPath(prefix);
    if (existsSync(path)) {
      const data = JSON.parse(readFileSync(path, "utf8"));
      jsonCache[prefix] = data;
      return data;
    }
    return null;
  }

  function buildProductionCollection(fullSet: any, usedIcons: Set<string>) {
    const filteredIcons: Record<string, any> = {};
    const filteredAliases: Record<string, any> = {};
    const queued = new Set<string>(usedIcons);
    const queue = [...queued];

    while (queue.length > 0) {
      const iconName = queue.shift();
      if (!iconName) continue;

      if (fullSet.icons?.[iconName]) {
        filteredIcons[iconName] = fullSet.icons[iconName];
        continue;
      }

      const alias = fullSet.aliases?.[iconName];
      if (alias && productionOptions.keepAliases) {
        filteredAliases[iconName] = alias;
        const parent = alias.parent;
        if (parent && !queued.has(parent)) {
          queued.add(parent);
          queue.push(parent);
        }
      }
    }

    const collectionData: Record<string, any> = {
      ...fullSet,
      icons: filteredIcons,
    };

    if (
      productionOptions.keepAliases &&
      Object.keys(filteredAliases).length > 0
    ) {
      collectionData.aliases = filteredAliases;
    } else {
      delete collectionData.aliases;
    }

    if (productionOptions.dropMeta) {
      delete collectionData.chars;
      delete collectionData.categories;
      delete collectionData.themes;
      delete collectionData.info;
    }

    return collectionData;
  }

  return {
    name,
    enforce: "pre",

    configResolved(config) {
      isBuild = config.command === "build";
      orderedCollectionPrefixes = [...getInstalledCollections()].sort(
        (a, b) => b.length - a.length,
      );
      productionOptions = {
        ...defaultProductionOptions,
        ...options.production,
      };
      whitelistCheckOptions = {
        ...defaultWhitelistCheckOptions,
        ...options.whitelistCheck,
      };
    },

    resolveId(id) {
      if (id === virtualModuleId) return resolvedVirtualModuleId;
    },

    async load(id) {
      if (id !== resolvedVirtualModuleId) return;

      clearIconMap();
      scanWorkspaceSource();
      applyWhitelist();

      if (isBuild && whitelistCheckOptions.enabled) {
        const missing = collectWhitelistMissingEntries();
        if (missing.length > 0) {
          const message = `[Iconify] Whitelist missing entries:\n - ${missing.join("\n - ")}`;
          if (whitelistCheckOptions.throwOnMissing) {
            this.error(message);
          } else {
            this.warn(message);
          }
        }
      }

      // Nuxt UI 会在这个虚拟模块里放一批默认图标。
      try {
        const nuxtUiConfig = await this.load({
          id: "virtual:nuxt-ui-app-config",
        });
        if (typeof nuxtUiConfig?.code === "string") {
          collectIconsFromCode(nuxtUiConfig.code);
        }
      } catch {
        // dev/build 某些阶段该模块可能还不存在，忽略即可
      }

      let outputCode = `import { addCollection } from "@iconify/vue";\n\n`;
      outputCode += `export function AllIcons() {\n`;

      for (const prefix in iconMap) {
        const fullSet = getFullSet(prefix);
        if (!fullSet) continue;
        let collectionData: any = fullSet;

        // 生产环境：只保留用到的图标 (Tree-shaking)
        if (isBuild && productionOptions.pruneIcons) {
          const usedIcons = iconMap[prefix];
          collectionData = buildProductionCollection(fullSet, usedIcons);
        } else if (isBuild && productionOptions.dropMeta) {
          collectionData = {
            ...fullSet,
          };
          delete collectionData.chars;
          delete collectionData.categories;
          delete collectionData.themes;
          delete collectionData.info;
        }

        const data = JSON.stringify(collectionData);
        outputCode += `  addCollection(${data});\n`;
      }

      outputCode += `}\n`;

      console.log(
        `📦 [Iconify] Bundled sets: ${Object.keys(iconMap).join(", ") || "none"}`,
      );
      return outputCode;
    },
  };
}
