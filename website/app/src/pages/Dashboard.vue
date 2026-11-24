<template>
  <div>
    <Header>
      <div class="lg:hidden">
        <button v-if="$route.name != 'dashboard.guilds'"
                class="text-gray-400 focus:outline-none focus:ring-2 focus:ring-inset" @click="sidebarOpen = true">
          <span class="sr-only">Open sidebar</span>
          <MenuAlt1Icon class="w-6 h-6" aria-hidden="true" />
        </button>
      </div>
    </Header>
    <router-view :sidebarOpen="sidebarOpen" v-on:closeSidebar="closeSidebar" />
    
    <Toast />
    <Popups />
  </div>
</template>

<script>
import { ref } from "vue";

import { MenuAlt1Icon } from "@heroicons/vue/outline";

import Header from "@/components/dashboard/Header.vue";
import Toast from "@/components/dashboard/Toast.vue";
import HoistHeading from "@/components/hoist/HoistHeading.vue";
import Popups from "@/components/Popups.vue";
import useGuildRoute from "@/composables/useGuildRoute";


export default {
  components: {
    Header,
    MenuAlt1Icon,
    HoistHeading,
    Popups,
    Toast,
  },
  methods: {
    closeSidebar() {
      this.sidebarOpen = false;
    },
  },
  setup() {
    useGuildRoute();

    return {
      sidebarOpen: ref(false),
    };
  },
};
</script>
