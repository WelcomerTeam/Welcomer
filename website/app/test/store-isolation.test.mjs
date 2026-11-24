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
let getCanvasSnapTargets;
let getSnappedPosition;

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
      hmr: false,
    },
    appType: "custom",
  });

  ({ createAppStore } = await viteServer.ssrLoadModule("/src/store/index.js"));
  ({ formatText } = await viteServer.ssrLoadModule("/src/utilities.js"));
  ({ getCanvasSnapTargets, getSnappedPosition } = await viteServer.ssrLoadModule("/src/utils/canvasSnapping.js"));
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

test("snaps the closest object anchor to a target within tolerance", () => {
  assert.deepEqual(getSnappedPosition(44, 20, { edges: [], centers: [50] }, 6), {
    position: 40,
    guide: 50,
  });
});

test("prioritizes the right edge when dragging to the right", () => {
  assert.deepEqual(
    getSnappedPosition(44, 20, { edges: [64], centers: [52] }, 8, "end"),
    {
    position: 44,
    guide: 64,
    }
  );
});

test("canvas snap targets account for the existing canvas border inset", () => {
  assert.deepEqual(getCanvasSnapTargets(1000, 16, [], -1, 0), {
    edges: [-16, 984],
    centers: [484],
    gaps: [],
  });
});

test("snaps a layer to the center of a gap when it fits", () => {
  const targets = {
    edges: [],
    centers: [],
    gaps: [{ start: 200, end: 400 }],
  };

  assert.deepEqual(getSnappedPosition(254, 100, targets, 8), {
    position: 250,
    guide: 300,
  });
});

test("does not snap a layer into a gap that is too narrow", () => {
  const targets = {
    edges: [],
    centers: [],
    gaps: [{ start: 200, end: 400 }],
  };

  assert.deepEqual(getSnappedPosition(254, 201, targets, 8), {
    position: 254,
    guide: null,
  });
});

test("releases a snap when the proposed drag moves beyond the tolerance", () => {
  assert.deepEqual(getSnappedPosition(40, 20, { edges: [], centers: [50] }, 8), {
    position: 40,
    guide: 50,
  });

  assert.deepEqual(getSnappedPosition(60, 20, { edges: [], centers: [50] }, 8), {
    position: 60,
    guide: null,
  });
});
