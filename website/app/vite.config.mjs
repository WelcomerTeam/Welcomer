import vue from "@vitejs/plugin-vue";
import analyze from "rollup-plugin-analyzer";
import { defineConfig } from "vite";

import { prerenderedRoutes } from "./src/router/prerendered.js";

// https://vitejs.dev/config/
export default defineConfig({
  resolve: {
    alias: [{ find: '@', replacement: '/src' }],
  },
  plugins: [vue()],
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:15011/api",
        rewrite: (path) => path.replace(/^\/api/, ""),
        changeOrigin: true,
      },
      "^/(login|logout|callback|patreon_link|patreon_callback)": {
        target: "http://localhost:15011",
        changeOrigin: true,
      },
    },
  },
  ssr: {
    // Bundle Font Awesome so the icon library populated in main.js is the same
    // instance vue-fontawesome reads from during pre-rendering.
    noExternal: [/^@fortawesome\//],
  },
  ssgOptions: {
    dirStyle: "flat",
    includedRoutes: () => prerenderedRoutes,
    // Tag the root container with the route it was rendered for, so the
    // client can skip hydration when this HTML is served as an SPA fallback.
    onPageRendered: (route, html) =>
      html.replace('<div id="app"', `<div id="app" data-ssg-path="${route}"`),
  },
  build: {
    outDir: process.env.WEBSITE_OUT_DIR || "dist",
    rollupOptions: {
      plugins: [analyze({ limit: 10 })],
      output: {
        // manualChunks: {
        //   "dashboard-scaffolding": [
        //     "./src/pages/DashboardGuild.vue",
        //     "./src/pages/GuildSelector.vue",
        //   ],
        //   "dashboard-pages": [
        //     "./src/pages/dashboard/Memberships.vue",
        //     "./src/pages/dashboard/Home.vue",
        //     "./src/pages/dashboard/Settings.vue",
        //     "./src/pages/dashboard/members/Welcomer.vue",
        //     "./src/pages/dashboard/members/Rules.vue",
        //     "./src/pages/dashboard/members/Borderwall.vue",
        //     "./src/pages/dashboard/members/Autorole.vue",
        //     "./src/pages/dashboard/members/Leaver.vue",
        //     "./src/pages/dashboard/roles/Freeroles.vue",
        //     "./src/pages/dashboard/roles/Timeroles.vue",
        //     "./src/pages/dashboard/roles/Tempchannels.vue",
        //     "./src/pages/dashboard/debug/Example.vue",
        //     "./src/pages/NotFound.vue",
        //   ],
        // },
      },
    },
  },
});
