<template>
  <div class="dashboard-container">
    <div v-if="this.isDataError">
      <div class="mb-4">Data Error</div>
      <button @click="this.fetchConfig">Retry</button>
    </div>
    <div v-else>
      <div class="my-16 text-center">
        <h1 class="dashboard-title">Custom Bots</h1>
        <p class="mt-8 text-gray-600 dark:text-gray-400 text-sm">
          Got a Discord token? Turn it into your very own Welcomer bot that greets, manages, and shows off your server's
          style.
        </p>
      </div>

      <div class="dashboard-components">
        <div class="dashboard-inputs">
          <div v-if="!isDataFetched && customBots.length === 0" class="flex py-5 w-full justify-center">
            <LoadingIcon />
          </div>
          <div v-else-if="isDataError" class="mb-4">
            <div class="text-red-500">Data Error</div>
            <button @click="fetchConfig" class="cta-button bg-primary hover:bg-primary-dark">
              Retry
            </button>
          </div>
          <div v-else>
            <div v-if="limit === 0" class="text-gray-500 mb-4">

              <div class="border-primary bg-primary text-white border p-6 lg:p-12 rounded-lg shadow-sm h-fit mt-16">
                <h3 class="text-2xl font-bold sm:text-3xl">
                  You've found a premium feature!
                </h3>
                <p class="mt-4 text-sm leading-6">Unlock custom bots on your server with a Welcomer Pro plan. Select from monthly, biannual or yearly plans to suit your needs.</p>

                <a href="/premium" target="_blank" type="button" class="bg-white hover:bg-gray-200 flex items-center justify-center px-5 py-3 mt-8 text-base font-medium text-primary border border-transparent rounded-md cursor-pointer w-full">
                  Learn More
                </a>
              </div>

            </div>
            <ul role="list" class="space-y-4" v-else>
              <li v-for="customBot in customBots" :key="customBot.id"
                class="bg-white dark:bg-secondary-dark shadow-sm rounded-md border-gray-300 dark:border-secondary-light border">
                <button
                  :class="['block hover:bg-gray-50 dark:hover:bg-secondary w-full rounded-md', customBot.open ? 'border-b border-gray-300 dark:border-secondary-light rounded-b-none' : '']"
                  @click="customBot.open = !customBot.open; customBot.showTokenInput = false">
                  <div class="px-4 py-4 flex items-center space-x-5 group">
                    <div class="flex-shrink-0">
                      <div class="flex -space-x-1">
                        <img :alt="`Bot icon for ${customBot.application_name}`" class="w-10 h-10 rounded-lg" v-lazy="{
                          src:
                            customBot.application_avatar !== ''
                              ? `https://cdn.discordapp.com/avatars/${customBot.application_id}/${customBot.application_avatar}.webp?size=128`
                              : '/assets/discordServer.svg',
                          error: '/assets/discordServer.svg',
                        }" />
                      </div>
                    </div>
                    <div class="min-w-0 flex-1 sm:flex sm:items-center sm:justify-between">
                      <div class="truncate">
                        <div class="flex text-sm">
                          <p class="font-bold truncate dark:text-gray-50">
                            {{ customBot.application_name }}
                            <span v-if="customBot.environment !== ''" class="font-bold text-sm px-2 py-1 rounded-md ml-2 bg-fuchsia-100 text-fuchsia-800 ring-fuchsia-800">
                              {{ customBot.environment }}
                            </span>
                          </p>
                        </div>
                      </div>
                    </div>
                    <div class="flex-shrink-0" v-if="customBot.shards.length > 0">
                      {{ countGuilds(customBot) }} {{ countGuilds(customBot) === 1 ? 'guild' : 'guilds' }}
                    </div>
                    <div class="flex-shrink-0">
                      <div v-if="!customBot.is_active">Inactive</div>
                      <div v-else-if="customBot.shards.length === 0">Idle</div>
                      <div v-else-if="customBot.shards.length > 0"
                        :class="['font-bold text-sm px-2 py-1 rounded-md', getStyleForShard(customBot.shards[0])]"> {{
                          getLabelForShard(customBot.shards[0]) }} </div>
                    </div>
                    <div class="flex-shrink-0">
                      <ChevronRightIcon :class="['h-5 w-5 text-gray-400 transition-all duration-100', customBot.open ? 'rotate-90' : '']"
                        aria-hidden="true" />
                    </div>
                  </div>
                  <div v-if="customBot.shards[0]?.status == 1" class="bg-red-100 text-red-800 m-2 p-2 rounded-md text-sm">
                    Failed to connect to discord. Please check the token, public key, and ensure the server members intent is enabled
                    in the bot tab on the Discord Developer Portal for your application.
                  </div>
                </button>
                <div v-if="customBot.open" class="p-4 flex flex-col gap-4">
                  <div class="flex gap-2 w-full">
                    <input v-model="customBot.public_key" type="text" autocomplete="off"
                      placeholder="Enter custom bot public key" class="flex-1 shadow-sm block w-full min-w-0 border-gray-300 dark:border-secondary-light rounded-md focus:ring-primary focus:border-primary sm:text-sm text-black">
                    <button type="button" @click="showPublicKeyPopup = true"
                      class="cta-button-dark dark:text-gray-50 bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600">
                      ?
                    </button>
                    <button type="button" @click="updateCustomBot(customBot.id, '', customBot.public_key)"
                      class="cta-button bg-primary hover:bg-primary-dark">
                      Update Public Key
                    </button>
                  </div>
                  <a href="#" @click="copyToClipboard(getInteractionsEndpointURL(customBot))"
                    class="text-gray-600 dark:text-gray-400 text-xs">
                    Your interactions endpoint URL is:
                    <span class="whitespace-nowrap underline">
                      {{ getInteractionsEndpointURL(customBot) }}
                      <font-awesome-icon icon="fa-regular fa-copy"
                        class="w-4 h-4 top-1 text-gray-400 absolute -left-6 hover:visible invisible"
                        aria-hidden="true" />
                    </span>
                  </a>

                  <button v-if="!customBot.showTokenInput" type="button" @click="customBot.showTokenInput = true"
                    class="cta-button-dark dark:text-gray-50 bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600">
                    Update Token
                  </button>
                  <div v-else class="flex items-center space-x-2 mb-4 w-full">
                    <input v-model="customBot.token" type="password" autocomplete="off"
                      placeholder="Enter custom bot token" class="flex-1 shadow-sm block w-full min-w-0 border-gray-300 dark:border-secondary-light rounded-md focus:ring-primary focus:border-primary sm:text-sm text-black">
                    <button type="button" @click="showTokenPopup = true"
                      class="cta-button-dark dark:text-gray-50 bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600">
                      ?
                    </button>
                    <button type="button" @click="updateCustomBot(customBot.id, customBot.token, customBot.public_key)"
                      class="cta-button bg-primary hover:bg-primary-dark">
                      Update Token
                    </button>
                  </div>

                  <div class="flex justify-between gap-2">
                    <button v-if="customBot.shards.length === 0 || (customBot.shards[0]?.status == 1) || !customBot.is_active" type="button"
                      @click="startCustomBot(customBot.id)" class="cta-button bg-green-500 hover:bg-green-600">
                      <LoadingIcon class="mr-3" v-if="isChangeInProgress" />
                      Start Bot
                    </button>
                    <button v-else type="button" @click="stopCustomBot(customBot.id)"
                      class="cta-button bg-gray-500 hover:bg-gray-600">
                      <LoadingIcon class="mr-3" v-if="isChangeInProgress" />
                      Stop Bot
                    </button>
                    <button type="button" @click="deleteCustomBot(customBot.id)"
                      class="cta-button bg-red-500 hover:bg-red-600">
                      Delete Bot
                    </button>
                  </div>
                </div>
              </li>
              <li v-for="n in limit - customBots.length" :key="n"
                class="rounded-md border-gray-300 dark:border-secondary-light border border-dashed">
                <div class="px-4 py-4 flex items-center space-x-5 group">
                  <div class="flex-shrink-0">
                    <div class="flex -space-x-1">
                      <img src="/assets/discordServer.svg" alt="" class="w-10 h-10 rounded-lg saturate-0  " />
                    </div>
                  </div>
                  <div class="min-w-0 flex-1 sm:flex sm:items-center sm:justify-between">
                    <div class="truncate">
                      <div class="flex text-sm">
                        <p class="font-bold truncate dark:text-gray-50">
                          Custom Bot Slot
                        </p>
                      </div>
                    </div>
                  </div>
                </div>
              </li>
              <li v-if="customBots.length < limit"
                class="bg-white dark:bg-secondary-dark shadow-sm rounded-md border-gray-300 dark:border-secondary-light border">
                <button class="block hover:bg-gray-50 dark:hover:bg-secondary w-full rounded-md"
                  @click="showCreateBotPopup = true">
                  <div class="flex p-4">
                    <div class="flex justify-start flex-grow">
                      Create Custom Bot
                    </div>
                    <div class="flex-shrink-0">
                      <PlusIcon class="h-5 w-5 text-gray-400" aria-hidden="true" />
                    </div>
                  </div>
                </button>
              </li>
            </ul>
          </div>
        </div>
      </div>

      <Popup :open="showPublicKeyPopup" @close="showPublicKeyPopup = false">
        <p>You can get your application public key from the General Information tab in your application on the Discord
          Developer Portal. The public key cannot be changed on Discord and is 64 characters long.</p>
        <img src="/assets/custom_bot_public_key.png" alt="" class="w-full max-w-md mt-2 rounded-md">
      </Popup>

      <Popup :open="showTokenPopup" @close="showTokenPopup = false">
        <p>You can get your bot's token from the Bot tab in your application on the Discord Developer Portal. If you
          forgot your token, you will need to reset it. Tokens usually start with M and are between 60 and 75 characters
          long.</p>
        <img src="/assets/custom_bot_token.png" alt="" class="w-full max-w-md mt-2 rounded-md">
      </Popup>

      <Popup :open="showCreateBotPopup" @close="showCreateBotPopup = false" :hideContinueButton="false"
        continueLabel="Create">
        <template v-slot:title>
          Create Custom Bot
        </template>

        <div class="space-y-4">
          <p>
            To create a custom bot, you will need to provide a valid Discord bot token and public key for interactions.
            Head to <a href="https://discord.com/developers/applications" target="_blank"
              class="text-blue-500 underline">Discord Developer Portal</a> to create an application, or use an existing
            one.
          </p>

          <p class="p-2 bg-yellow-100 text-yellow-800 rounded-md text-sm">
            Note: Welcomer will require the server members intent to run properly. Please enable it in the bot tab on
            the
            Discord Developer Portal for your application.
            <img src="/assets/custom_bot_intents.png" alt="" class="w-full max-w-md mt-2 rounded-md">
          </p>

          <div>
            <input v-model="newPublicKey" type="text" autocomplete="off" placeholder="Enter custom bot public key"
              class="flex-1 shadow-sm block w-full min-w-0 border-gray-300 dark:border-secondary-light rounded-md focus:ring-primary focus:border-primary sm:text-sm text-black">
            <span class="text-gray-600 dark:text-gray-400 text-sm">You can get your application public key from the
              General
              Information tab in your application on the Discord Developer Portal. The public key cannot be changed, is
              64
              characters long.</span>
          </div>

          <div>
            <input v-model="newBotToken" type="password" autocomplete="off" placeholder="Enter new custom bot token"
              class="flex-1 shadow-sm block w-full min-w-0 border-gray-300 dark:border-secondary-light rounded-md focus:ring-primary focus:border-primary sm:text-sm text-black">
            <span class="text-gray-600 dark:text-gray-400 text-sm">You can get your bot's token from the Bot tab in your
              application on the Discord Developer Portal. If you forgot your token, you will need to reset it. Tokens
              usually
              start with M and are between 60 and 75 characters long.</span>
          </div>

          <div class="flex justify-end">
            <button type="button"
              class="cta-button bg-green-500 hover:bg-green-600 disabled:bg-gray-100 disabled:dark:bg-secondary-light disabled:text-neutral-500"
              @click="createCustomBot(newBotToken, newPublicKey)"
              :disabled="isChangeInProgress || !newPublicKey || !newBotToken">
              Create
            </button>
          </div>
        </div>
      </Popup>
    </div>
  </div>
</template>

<script>
import { ref } from "vue";

// import {
//   FormTypeBlank,
//   FormTypeToggle,
// } from "@/components/dashboard/FormValueEnum";

import { ChevronRightIcon, PlusIcon } from "@heroicons/vue/outline";

import FormValue from "@/components/dashboard/FormValue.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";
import Popup from "@/components/Popup.vue";

import { Menu, MenuButton, MenuItem, MenuItems } from '@headlessui/vue';

import dashboardAPI from "@/api/dashboard";
import endpoints from "@/api/endpoints";

import {
  getErrorToast
} from "@/utilities";


export default {
  components: {
    FormValue,
    LoadingIcon,
    Menu,
    MenuButton,
    MenuItem,
    MenuItems,
    Popup,
    ChevronRightIcon,
    PlusIcon,
  },
  setup() {
    let isDataFetched = ref(false);
    let isDataError = ref(false);
    let unsavedChanges = ref(false);
    let isChangeInProgress = ref(false);

    let limit = ref(0);
    let customBots = ref([]);

    let newPublicKey = ref("");
    let newBotToken = ref("");

    let showPublicKeyPopup = ref(false);
    let showTokenPopup = ref(false);
    let showCreateBotPopup = ref(false);

    let interval = undefined;

    return {
      isDataFetched,
      isDataError,
      unsavedChanges,
      isChangeInProgress,

      limit,
      customBots,

      newPublicKey,
      newBotToken,

      showPublicKeyPopup,
      showTokenPopup,
      showCreateBotPopup,

      interval,
    }
  },

  mounted() {
    this.fetchConfig();

    this.interval = setInterval(() => {
      this.fetchConfig();
    }, 30000);
  },
  unmounted() {
    clearInterval(this.interval);
  },

  methods: {

    fetchConfig() {
      this.isDataFetched = false;
      this.isDataError = false;

      dashboardAPI.getConfig(
        endpoints.EndpointGuildCustomBots(this.$store.getters.getSelectedGuildID),
        ({ config }) => {
          this.limit = config.limit;
          this.customBots = config.bots;
          this.isDataFetched = true;
          this.isDataError = false;
        },
        (error) => {
          this.$store.dispatch("createToast", getErrorToast(error));

          this.isDataFetched = false;
          this.isDataError = true;
        }
      )
    },

    async createCustomBot(token, public_key) {
      if (!this.validatePublicKey(public_key)) { return; }

      this.validateToken(token, () => {
        this.isChangeInProgress = true;

        dashboardAPI.doPost(
          endpoints.EndpointGuildCustomBot(this.$store.getters.getSelectedGuildID, ""),
          { token, public_key: public_key },
          null,
          () => {
            this.unsavedChanges = false;
            this.isChangeInProgress = false;
            this.showCreateBotPopup = false;
            this.fetchConfig();

            setTimeout(() => {
              this.fetchConfig();
            }, 5000);

            this.$store.dispatch("createPopup", {
              id: "custom-bot-created",
              title: "Custom Bot Created",
              description: "Your custom bot has been created successfully! If it has not started automatically, you can click the 'Start Bot' button to start it.\n\nLastly, you need to setup the Interaction Endpoint URL else interactions will not work. You can find the URL under the custom bot details and is configured in the General Information tab on the Discord Developer Portal for your application.",
              closeFunction: null,
              continueFunction: null,
              showCloseButton: true,
              hideContinueButton: true,
            });
          },
          (error) => {
            this.$store.dispatch("createToast", getErrorToast(error));
            this.isChangeInProgress = false;
          }
        );
      }, (e) => {
        this.$store.dispatch("createToast", {
          title: `Failed to create custom bot: ${e}. Please check the token and public key.`,
          icon: "xmark",
          class: "text-red-500 bg-red-100",
        });
      });
    },

    async updateCustomBot(customBotUUID, token, public_key) {
      if (!this.validatePublicKey(public_key)) { return; }

      var continueFunc = () => {
        this.isChangeInProgress = true;

        dashboardAPI.doPost(
          endpoints.EndpointGuildCustomBot(this.$store.getters.getSelectedGuildID, customBotUUID),
          { token, public_key },
          null,
          () => {
            this.unsavedChanges = false;
            this.isChangeInProgress = false;
            this.fetchConfig();
          },
          (error) => {
            this.$store.dispatch("createToast", getErrorToast(error));
            this.isChangeInProgress = false;
          }
        );
      }

      if (token === "") {
        continueFunc();
      } else {
        this.validateToken(token, continueFunc, () => { });
      }
    },

    startCustomBot(customBotUUID) {
      this.isChangeInProgress = true;

      dashboardAPI.doPost(
        endpoints.EndpointStartGuildCustomBot(this.$store.getters.getSelectedGuildID, customBotUUID),
        {},
        null,
        () => {
          this.unsavedChanges = false;
          this.isChangeInProgress = false;
          this.fetchConfig();

          setTimeout(() => {
            this.fetchConfig();
          }, 5000);
        },
        (error) => {
          this.$store.dispatch("createToast", getErrorToast(error));
          this.isChangeInProgress = false;
        }
      );
    },

    stopCustomBot(customBotUUID) {
      this.isChangeInProgress = true;

      dashboardAPI.doPost(
        endpoints.EndpointStopGuildCustomBot(this.$store.getters.getSelectedGuildID, customBotUUID),
        {},
        null,
        () => {
          this.unsavedChanges = false;
          this.isChangeInProgress = false;
          this.fetchConfig();
        },
        (error) => {
          this.$store.dispatch("createToast", getErrorToast(error));
          this.isChangeInProgress = false;
        }
      );
    },

    deleteCustomBot(customBotUUID) {
      if (!confirm("Are you sure you want to delete this custom bot? This action cannot be undone.")) {
        return;
      }

      this.isChangeInProgress = true;

      dashboardAPI.doAPICall("DELETE",
        endpoints.EndpointGuildCustomBot(this.$store.getters.getSelectedGuildID, customBotUUID),
        {},
        null,
        () => {
          this.unsavedChanges = false;
          this.isChangeInProgress = false;
          this.fetchConfig();
        },
        (error) => {
          this.$store.dispatch("createToast", getErrorToast(error));
          this.isChangeInProgress = false;
        }
      );
    },

    getStyleForShard(shard) {
      if (shard.latency < 0) {
        shard.status = 4;
      }

      switch (shard.status) {
        case 1: // failed
          return "bg-red-200 text-red-800 ring-red-800";
        case 2: // connecting
          return "bg-fuchsia-100 text-fuchsia-700 ring-fuchsia-700";
        case 3: // connected
          return "bg-fuchsia-200 text-fuchsia-800 ring-fuchsia-800";
        case 4: // ready
          return "bg-emerald-200 text-emerald-800 ring-emerald-800";
        case 5: /// stopping
          return "bg-emerald-100 text-emerald-700 ring-emerald-700";
        case 6: // stopped
          return "bg-amber-200 text-amber-800 ring-amber-800";
        default: // idle
          return "bg-gray-200 text-gray-800 ring-gray-800";
      }
    },

    getLabelForShard(shard) {
      if (shard.latency < 0) {
        shard.status = 4;
      }

      switch (shard.status) {
        case 1: // failed
          return "Failed to connect";
        case 2: // connecting
          return "Connecting";
        case 3: // connected
          return "Connected";
        case 4: // ready
          return "Ready";
        case 5: // stopping
          return "Stopping";
        case 6: // stopped
          return "Stopped";
        default: // idle
          return "Idle";
      }
    },

    copyToClipboard(text) {
      navigator.clipboard.writeText(text);

      this.$store.dispatch("createToast", {
        title: "Copied to clipboard",
        icon: "info",
        class: "text-blue-500 bg-blue-100",
      });
    },

    getInteractionsEndpointURL(customBot) {
      return `https://${document.location.host}/interactions?manager=custom_bot_${customBot.id}`;
    },

    countGuilds(customBot) {
      return customBot.shards.reduce((acc, shard) => acc + shard.guilds, 0);
    },

    validatePublicKey(publicKey) {
      if (!publicKey || publicKey.trim() === "") {
        this.$store.dispatch("createToast", {
          title: "Public Key is required",
          icon: "xmark",
          class: "text-red-500 bg-red-100",
        });
        return false;
      }

      const publicKeyRegex = /^[0-9a-fA-F]{64}$/;
      if (!publicKeyRegex.test(publicKey)) {
        this.$store.dispatch("createToast", {
          title: "Invalid Public Key format",
          icon: "xmark",
          class: "text-red-500 bg-red-100",
        });
        return false;
      }

      return true;
    },

    validateToken(token, successCallback, errorCallback) {
      if (!token || token.trim() === "") {
        return errorCallback("token is required");
      }

      const tokenRegex = /^[A-Za-z0-9_\-]{24,28}\.[A-Za-z0-9_\-]{6}\.[A-Za-z0-9_\-]{27,38}$/;
      if (!tokenRegex.test(token)) {
        return errorCallback("invalid token format");
      }

      fetch(`https://discord.com/api/v10/users/@me`, {
        headers: {
          "Authorization": `Bot ${token}`
        }
      })
      .then((response) => {
        if (response.status !== 200) {
          return errorCallback("invalid token");
        }
        return successCallback();
      })
      .catch((error) => {
        console.error("Error validating token:", error);
        return errorCallback(error);
      });
    },
  }
}
</script>