import assert from "node:assert/strict";
import { after, before, test } from "node:test";
import path from "node:path";
import { fileURLToPath } from "node:url";

import { createServer } from "vite";

const appRoot = path.resolve(path.dirname(fileURLToPath(import.meta.url)), "..");
const originalNodeEnv = process.env.NODE_ENV;
let viteServer;
let createAppStore;
let formatText;

before(async () => {
  process.env.NODE_ENV = "production";
  viteServer = await createServer({
    configFile: false,
    root: appRoot,
    resolve: {
      alias: {
        "@": path.join(appRoot, "src"),
      },
    },
    optimizeDeps: {
      noDiscovery: true,
    },
    server: {
      middlewareMode: true,
    },
    appType: "custom",
  });

  ({ createAppStore } = await viteServer.ssrLoadModule("/src/store/index.js"));
  ({ formatText } = await viteServer.ssrLoadModule("/src/utilities.js"));
});

after(async () => {
  await viteServer?.close();

  if (originalNodeEnv === undefined) {
    delete process.env.NODE_ENV;
  } else {
    process.env.NODE_ENV = originalNodeEnv;
  }
});

test("app stores have isolated module state", () => {
  const firstStore = createAppStore();
  const secondStore = createAppStore();

  firstStore.commit("setCurrentUser", { id: "first" });
  firstStore.commit("addToast", { id: 1 });
  secondStore.commit("setCurrentUser", { id: "second" });

  assert.equal(firstStore.getters.getCurrentUser.id, "first");
  assert.equal(secondStore.getters.getCurrentUser.id, "second");
  assert.deepEqual(secondStore.getters.getToasts, []);
  assert.equal(formatText("{{User.ID}}", firstStore), "first");
  assert.equal(formatText("{{User.ID}}", secondStore), "second");
});

test("guild selection mutation is pure and action clears selected guild state", async () => {
  const store = createAppStore();

  store.commit("setSelectedGuild", "guild-1");

  assert.equal(store.getters.getSelectedGuildID, "guild-1");
  assert.equal(store.getters.isLoadingGuild, false);

  await store.dispatch("selectGuild", null);

  assert.equal(store.getters.getSelectedGuildID, null);
  assert.equal(store.getters.getCurrentSelectedGuild, null);
});
