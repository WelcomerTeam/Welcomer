import { createWebHistory, createRouter } from "vue-router";

import Home from "@/pages/Home.vue";
import Dashboard from "@/pages/Dashboard.vue";

const routes = [
  {
    path: "/",
    name: "home",
    component: Home,
  },
  {
    path: "/premium",
    name: "premium",
    component: () => import("@/pages/Premium.vue"),
  },
  {
    path: "/invite",
    name: "invite",
    component: () => import("@/pages/Invite.vue"),
  },
  {
    path: "/support",
    name: "support",
    component: () => import("@/pages/Support.vue"),
  },
  {
    path: "/phishing",
    name: "phishing",
    component: () => import("@/pages/Phishing.vue"),
  },
  {
    path: "/backgrounds",
    name: "backgrounds",
    component: () => import("@/pages/Backgrounds.vue"),
  },
  {
    path: "/status",
    name: "status",
    component: () => import("@/pages/Status.vue"),
  },
  {
    path: "/faq",
    name: "faq",
    component: () => import("@/pages/FAQ.vue"),
  },
  {
    path: "/formatting",
    name: "formatting",
    component: () => import("@/pages/Formatting.vue"),
  },
  {
    path: "/borderwall/:key?",
    name: "borderwall",
    component: () => import("@/pages/Borderwall.vue"),
  },
  {
    path: "/terms",
    name: "terms",
    component: () => import("@/pages/Terms.vue"),
  },
  {
    path: "/privacy",
    name: "privacy",
    component: () => import("@/pages/Privacy.vue"),
  },
  {
    name: "dashboard.guild.welcomer_builder",
    path: "/dashboard/:guildID/builder",
    component: () => import("@/pages/dashboard/members/WelcomerImageBuilder.vue"),
  },
  {
    path: "/dashboard",
    name: "dashboard",
    component: Dashboard,
    children: [
      {
        path: "",
        name: "dashboard.guilds",
        component: () => import("@/pages/GuildSelector.vue"),
      },
      {
        path: ":guildID",
        component: () => import("@/pages/DashboardGuild.vue"),
        children: [
          {
            name: "dashboard.guild.overview",
            path: "",
            component: () => import("@/pages/dashboard/Home.vue"),
          },
          {
            name: "dashboard.guild.memberships",
            path: "memberships",
            component: () => import("@/pages/dashboard/Memberships.vue"),
          },
          {
            name: "dashboard.guild.custombots",
            path: "custombots",
            component: () => import("@/pages/dashboard/CustomBots.vue"),
          },
          {
            name: "dashboard.guild.settings",
            path: "settings",
            component: () => import("@/pages/dashboard/Settings.vue"),
          },

          {
            name: "dashboard.guild.welcomer",
            path: "welcomer",
            component: () => import("@/pages/dashboard/members/Welcomer.vue"),
          },
          {
            name: "dashboard.guild.rules",
            path: "rules",
            component: () => import("@/pages/dashboard/members/Rules.vue"),
          },
          {
            name: "dashboard.guild.borderwall",
            path: "borderwall",
            component: () => import("@/pages/dashboard/members/Borderwall.vue"),
          },
          {
            name: "dashboard.guild.leaver",
            path: "leaver",
            component: () => import("@/pages/dashboard/members/Leaver.vue"),
          },
          {
            name: "dashboard.guild.tempchannels",
            path: "tempchannels",
            component: () =>
              import("@/pages/dashboard/members/Tempchannels.vue"),
          },

          {
            name: "dashboard.guild.autoroles",
            path: "autorole",
            component: () => import("@/pages/dashboard/roles/Autorole.vue"),
          },
          {
            name: "dashboard.guild.freeroles",
            path: "freeroles",
            component: () => import("@/pages/dashboard/roles/Freeroles.vue"),
          },
          {
            name: "dashboard.guild.timeroles",
            path: "timeroles",
            component: () => import("@/pages/dashboard/roles/Timeroles.vue"),
          },

          {
            path: "example",
            name: "dashboard.guild.example",
            component: () => import("@/pages/dashboard/debug/Example.vue"),
          },
        ],
      },
    ],
  },
  {
    path: "/:catchAll(.*)",
    component: () => import("@/pages/NotFound.vue"),
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
});

export default router;
