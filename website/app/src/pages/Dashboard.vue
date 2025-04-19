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
import Header from "@/components/dashboard/Header.vue";
import { MenuAlt1Icon } from "@heroicons/vue/outline";
import { useRoute } from "vue-router";
import store from "@/store/index";
import { ref } from "vue";
import HoistHeading from "@/components/hoist/HoistHeading.vue";

import Popups from "@/components/Popups.vue";
import Toast from "@/components/dashboard/Toast.vue";

export default {
  components: {
    Header,
    MenuAlt1Icon,
    HoistHeading,
    Popups,
    Toast,
  },
  watch: {
    "$route.params.guildID"(to) {
      store.commit("setSelectedGuild", to);
    },
  },
  methods: {
    closeSidebar() {
      this.sidebarOpen = false;
    },
  },
  setup() {
    store.watch(
      () => store.getters.getSelectedGuildID,
      () => {
        if (store.getters.getSelectedGuildID !== undefined) {
          store.dispatch("fillGuild");
        }
      }
    );

    const route = useRoute();

    let guildID = route.params.guildID;
    if (guildID !== undefined) {
      store.commit("setSelectedGuild", guildID);
    }

    let sidebarOpen = ref(false);

    return {
      sidebarOpen,
    };
  },
};
</script>
