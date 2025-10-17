<template>
  <div class="flex-1 focus:outline-none bg-white dark:bg-secondary">
    <main class="z-0 flex-1">
      <div class="font-medium pb-20 min-h-screen">
        <div class="dashboard-container">
          <div class="dashboard-title-container">
            <div class="dashboard-title">My Guilds</div>
            <button type="button"
              class="-mx-1.5 bg-white text-gray-500 rounded-lg focus:ring-2 focus:ring-gray-500 p-1.5 inline-flex h-8 w-8 hover:bg-gray-100 dark:bg-secondary-dark dark:text-gray-50 dark:hover:bg-secondary-light"
              @click="refreshGuildList()" aria-label="Refresh guild list">
              <span class="sr-only">Refresh guild list</span>
              <font-awesome-icon icon="arrows-rotate" :class="[
                $store.getters.isLoadingGuilds ? 'fa-spin' : '',
                'w-5 h-5',
              ]" />
            </button>
          </div>
          <div class="dashboard-content">
            <!-- <div v-if="$store.getters.isLoadingGuilds && $store.getters.getGuilds.length === 0"
              class="mt-4 p-6 justify-center flex items-center dark:text-gray-50">
              <LoadingIcon class="mr-3" />
              Loading your guilds...
            </div> -->
            <div v-if="!$store.getters.isLoadingGuilds"
              class="mt-4 bg-white dark:bg-secondary-dark shadow-sm rounded-md border-gray-300 dark:border-secondary-light border">
              <ul role="list" class="divide-y divide-gray-200 dark:divide-secondary-light">
                <li v-if="$store.getters.getGuilds.length === 0">
                  <div class="p-4">
                    <p class="font-medium text-center max-w-xl mx-auto">
                      Failed to get a list of your guilds. Please allow Welcomer
                      to view all your guilds or try refreshing.
                    </p>
                  </div>
                </li>
                <li v-else v-for="guild in $store.getters.getGuilds" :key="guild.id" @click="setSelectedGuild(guild)">
                  <button class="block hover:bg-gray-50 dark:hover:bg-secondary w-full rounded-md">
                    <div class="px-4 py-4 flex items-center space-x-5 group">
                      <div class="flex-shrink-0">
                        <div class="flex -space-x-1">
                          <img :alt="`Guild icon for ${guild.name}`" :class="[
                            !guild.has_welcomer | !guild.has_elevation
                              ? 'saturate-0'
                              : '',
                            'w-10 h-10 rounded-lg',
                          ]" v-lazy="{
                            src:
                              guild.icon !== ''
                                ? `https://cdn.discordapp.com/icons/${guild.id}/${guild.icon}.webp?size=128`
                                : '/assets/discordServer.svg',
                            error: '/assets/discordServer.svg',
                          }" />
                        </div>
                      </div>
                      <div class="min-w-0 flex-1 sm:flex sm:items-center sm:justify-between">
                        <div class="truncate">
                          <div class="flex text-sm">
                            <p class="font-bold truncate dark:text-gray-50">
                              <span v-if="guild.has_welcomer_pro"
                                class="mr-2 inline-flex items-center p-2 rounded-md text-xs font-medium bg-primary-light text-white">
                                <font-awesome-icon icon="heart" />
                              </span>
                              <span v-else-if="guild.has_custom_backgrounds"
                                class="mr-2 inline-flex items-center p-2 rounded-md text-xs font-medium bg-gray-500 text-white">
                                <font-awesome-icon icon="heart" />
                              </span>
                              {{ guild.name }}
                            </p>
                          </div>
                        </div>
                      </div>
                      <div class="flex-shrink-0" v-if="guild.has_elevation">
                        <ChevronRightIcon v-if="guild.has_welcomer" class="h-5 w-5 text-gray-400" aria-hidden="true" />
                        <button v-else type="button" class="cta-button bg-primary group-hover:bg-primary-dark">
                          Invite
                        </button>
                      </div>
                    </div>
                  </button>
                </li>
              </ul>
            </div>
          </div>
        </div>
      </div>

      <Footer />
    </main>
  </div>
</template>

<script>
import Footer from "@/components/Footer.vue";
import { ChevronRightIcon, PlusIcon } from "@heroicons/vue/outline";
import FormValue from "@/components/dashboard/FormValue.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";
import store from "@/store/index";

import {
  OpenBotInvite,
  PrimaryBotId
} from "@/constants";

export default {
  components: { FormValue, ChevronRightIcon, LoadingIcon, PlusIcon, Footer },
  setup() {
    store.dispatch("fetchGuilds");
  },
  methods: {
    refreshGuildList() {
      this.$store.dispatch("refreshGuilds");
    },

    setSelectedGuild(guild) {
      if (guild.has_welcomer) {
        this.$store.commit("setSelectedGuild", guild.id);
        this.$router.push({
          name: "dashboard.guild.overview",
          params: {
            guildID: guild.id,
          },
        });
      } else {
        OpenBotInvite(PrimaryBotId, guild.id, this.refreshGuildList);
      }
    },
  },
};
</script>
