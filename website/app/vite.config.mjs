import { defineConfig } from "vite";
import vue from "@vitejs/plugin-vue";
import analyze from "rollup-plugin-analyzer";

// https://vitejs.dev/config/
export default defineConfig({
  resolve: {
    alias: [{ find: '@', replacement: '/src' }],
  },
  plugins: [vue()],
  server: {
    proxy: {
      "/api": {
        target: "http://localhost:1000/api",
        rewrite: (path) => path.replace(/^\/api/, ""),
        changeOrigin: true,
      },
      "/(login|logout|callback)": {
        target: "http://localhost:1000",
        changeOrigin: true,
      },
    },
  },
  build: {
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
