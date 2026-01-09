<template>
  <div>
    <div v-if="$store.getters.isLoadingGuild" class="justify-center flex h-10 items-center">
      <LoadingIcon />
    </div>
    <div v-else>
      <div class="flex items-center flex-shrink-0 px-4">
        <img v-if="$store.getters.getCurrentSelectedGuild" class="w-10 h-10 rounded-lg" :src="$store.getters.getCurrentSelectedGuild?.icon !== ''
            ? `https://cdn.discordapp.com/icons/${$store.getters.getCurrentSelectedGuild?.id}/${$store.getters.getCurrentSelectedGuild?.icon}.webp?size=128`
            : '/assets/discordServer.svg'
          " alt="Server icon" />
        <div class="pl-2 overflow-hidden dark:text-gray-50">
          <router-link @click="$emit('onTabClick')" :to="{ name: 'dashboard.guild.overview', params: $route.params }"
            v-if="$store.getters.getCurrentSelectedGuild">
            <h3 class="truncate font-bold leading-none hover:underline">
              {{ $store.getters.getCurrentSelectedGuild?.name }}
            </h3>
          </router-link>
          <h3 v-else class="truncate font-bold leading-none">
            No Server Selected
          </h3>
          <router-link @click="$emit('onTabClick')"
            class="text-xs leading-none font-semibold text-gray-600 dark:text-gray-300 hover:underline"
            :to="{ name: 'dashboard.guilds', params: $route.params }">Change Server</router-link>
        </div>
      </div>

      <!-- Sidebar -->
      <nav
        class="flex flex-col flex-1 px-3 mt-5 overflow-y-auto divide-y divide-gray-300 dark:divide-secondary custom-scroll"
        aria-label="Sidebar" v-if="$store.getters.getCurrentSelectedGuild">
        <div v-for="(nav, index) in navigation" v-bind:key="index" :class="[index === 0 ? '' : 'pt-3 mt-3']">
          <div>
            <span class="text-xs font-bold group uppercase text-secondary-light dark:text-gray-50" v-if="nav.title">{{
              nav.title }}</span>
            <router-link @click="$emit('onTabClick')" v-for="item in nav.items" :key="item.name"
              :to="{ name: item.linkname, params: $route.params }" :class="[
                $route.name === item.linkname
                  ? 'text-secondary dark:text-gray-50 bg-gray-200 dark:bg-secondary'
                  : 'text-gray-600 dark:text-gray-400',
                'hover:text-secondary dark:hover:text-gray-300 hover:bg-gray-200 dark:hover:bg-secondary group flex items-center px-2 py-2 text-sm leading-6 font-semibold rounded-md',
                item.class,
              ]">
              <font-awesome-icon :icon="($route.name === item.linkname ? 'fa-solid' : 'fa-regular') + ' ' + item.icon"
                :class="[
                  'flex group-hover:hidden flex-shrink-0 w-6 h-6 mr-4',
                  item.class,
                ]" aria-hidden="true" />
              <font-awesome-icon :icon="'fa-solid ' + item.icon" :class="[
                'hidden group-hover:flex flex-shrink-0 w-6 h-6 mr-4',
                item.class,
              ]" aria-hidden="true" />
              {{ item.name }}
            </router-link>
          </div>
        </div>
      </nav>
    </div>
  </div>
</template>

<script>
import {
  Dialog,
  Menu,
  MenuItem,
} from "@headlessui/vue";

import LoadingIcon from "@/components/LoadingIcon.vue";

const navigation = [
  {
    items: [
      {
        name: "Memberships",
        linkname: "dashboard.guild.memberships",
        icon: "fa-heart",
        class: "text-primary",
      },
      {
        name: "Custom Bots",
        linkname: "dashboard.guild.custombots",
        icon: "fa-robot",
        class: "text-primary",
      },
      {
        name: "Bot Customisation",
        linkname: "dashboard.guild.customisation",
        icon: "fa-tag",
        class: "text-primary",
      }
    ],
  },
  {
    items: [
      {
        name: "Server Overview",
        linkname: "dashboard.guild.overview",
        icon: "fa-chart-line",
      },
      {
        name: "Bot Settings",
        linkname: "dashboard.guild.settings",
        icon: "fa-wrench",
      },
    ],
  },
  {
    items: [
      {
        name: "Borderwall",
        linkname: "dashboard.guild.borderwall",
        icon: "fa-door-closed",
      },
      {
        name: "Leaver",
        linkname: "dashboard.guild.leaver",
        icon: "fa-user-minus",
      },
      {
        name: "Rules",
        linkname: "dashboard.guild.rules",
        icon: "fa-list-ol",
      },
      {
        name: "Welcomer",
        linkname: "dashboard.guild.welcomer",
        icon: "fa-user-plus",
      },
    ],
  },
  {
    items: [
      {
        name: "AutoRoles",
        linkname: "dashboard.guild.autoroles",
        icon: "fa-user-check",
      },
      {
        name: "FreeRoles",
        linkname: "dashboard.guild.freeroles",
        icon: "fa-list-check",
      },
      {
        name: "TimeRoles",
        linkname: "dashboard.guild.timeroles",
        icon: "fa-user-clock",
      },
      {
        name: "TempChannels",
        linkname: "dashboard.guild.tempchannels",
        icon: "fa-microphone-lines",
      },
    ],
  },
  // {
  //   items: [
  //     {
  //       name: "Examples",
  //       linkname: "dashboard.guild.example",
  //       icon: "fa-list-check",
  //     },
  //   ],
  // },
];

export default {
  components: {
    Dialog,
    LoadingIcon,
    Menu,
    MenuItem,
  },
  setup() {
    return {
      navigation,
    };
  },
};
</script>
