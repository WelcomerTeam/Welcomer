<template>
  <div class="dashboard-container">
    <div v-if="this.isDataError">
      <div class="mb-4">Data Error</div>
      <button @click="this.fetchConfig">Retry</button>
    </div>
    <div v-else>
      <div v-if="!this.isDataFetched" class="flex py-5 w-full justify-center">
        <LoadingIcon />
      </div>
      <div v-else>
        <div class="dashboard-title-container">
          <div class="dashboard-title">Accounts</div>
        </div>
        <div class="dashboard-contents">
          <div class="dashboard-inputs">

            <div v-if="isDataFetched" class="mt-4 bg-white dark:bg-secondary-dark shadow-sm rounded-md border-gray-300 dark:border-secondary-light border">
              <ul role="list" class="divide-y divide-gray-200 dark:divide-secondary-light">
                <li>
                  <div class="block hover:bg-gray-50 dark:hover:bg-secondary w-full">
                    <div class="px-4 py-4 flex items-center space-x-5 group">
                      <div class="flex-shrink-0">
                        <div class="flex overflow-hidden -space-x-1">
                          <img alt="" class="w-10 h-10 rounded-lg" v-lazy="{
                              src: getAccountByPlatform('patreon')?.thumb_url || '/assets/patreonIcon.svg',
                              error: '/assets/patreonIcon.svg',
                            }" />
                        </div>
                      </div>
                      <div class="min-w-0 flex-1 sm:flex sm:items-center sm:justify-between">
                        <div class="truncate">
                          <div class="flex text-sm">
                            <p class="truncate dark:text-gray-50 text-left">
                              <span class="text-sm font-semibold text-primary">Patreon</span>
                              <br />
                              <span>{{ getAccountByPlatform("patreon")?.name || 'Not yet linked' }}</span>
                              <br />
                              <span class="text-xs text-gray-600 dark:text-gray-400">{{ getAccountByPlatform('patreon')?.tier_id ? 'Current Tier: ' + getPatreonTierName(getAccountByPlatform('patreon')?.tier_id) : 'Not currently pledging' }}</span>
                            </p>
                          </div>
                        </div>
                      </div>
                      <div class="flex-shrink-0">
                        <button v-if="!getAccountByPlatform('patreon')" @click="gotoPatreonLink" type="button" class="cta-button bg-primary group-hover:bg-primary-dark">
                          Connect
                        </button>
                      </div>
                    </div>
                  </div>
                </li> 
              </ul>
            </div>
          </div>
        </div>
        <div class="dashboard-title-container">
          <div class="dashboard-title">Memberships</div>
        </div>
        <div class="dashboard-contents">
          <div class="dashboard-inputs">

            <div v-if="isDataFetched"
              class="mt-4 bg-white dark:bg-secondary-dark shadow-sm rounded-md border-gray-300 dark:border-secondary-light border">
              <ul role="list" class="divide-y divide-gray-200 dark:divide-secondary-light">
                <li v-if="memberships.length === 0">
                  <div class="p-4">
                    <p class="font-medium text-center max-w-xl mx-auto">
                      You do not have any memberships!
                    </p>
                  </div>
                </li>
                <li v-else v-for="membership in memberships" :key="membership.membership_uuid">
                  <div class="block hover:bg-gray-50 dark:hover:bg-secondary w-full">
                    <div class="px-4 py-4 flex items-center space-x-5 group">
                      <div class="flex-shrink-0">
                        <div class="flex overflow-hidden -space-x-1">
                          <div v-if="membership.guild_id == 0" class="w-10 h-10 rounded-lg dark:bg-white bg-black opacity-20"></div>
                          <img v-else alt=""
                            :class="[
                              membership.guild_id > 0 || isMembershipActive(membership) ? '' : 'saturate-0', 'w-10 h-10 rounded-lg']"
                            v-lazy="{
                              src: membership.guild_icon !== ''
                                ? `https://cdn.discordapp.com/icons/${membership.guild_id}/${membership.guild_icon}.webp?size=128`
                                : '/assets/discordServer.svg',
                              error: '/assets/discordServer.svg',
                            }" />
                        </div>
                      </div>
                      <div class="min-w-0 flex-1 sm:flex sm:items-center sm:justify-between">
                        <div class="truncate">
                          <div class="flex text-sm">
                            <p class="truncate dark:text-gray-50 text-left">
                              <span class="text-sm font-semibold text-primary">
                                <font-awesome-icon title="Discord subscription" v-if="membership.platform_type == PlatformTypeDiscord" :icon="['fab','discord']" />
                                <font-awesome-icon title="Patreon subscription" v-if="membership.platform_type == PlatformTypePatreon" :icon="['fab','patreon']" />
                                <font-awesome-icon title="Paypal purchase" v-if="membership.platform_type == PlatformTypePaypal || membership.platform_type == PlatformTypePaypalSubscription" :icon="['fab','paypal']" />
                                <font-awesome-icon title="Paypal subscription" v-if="membership.platform_type == PlatformTypePaypalSubscription" :icon="['fas','rotate-right']" />
                                {{ getMembershipTypeLabel(membership) }}
                              </span>
                              <br />
                              {{ membership.guild_id > 0 ? membership.guild_name : 'Unassigned' }}
                              <br />
                              <span class="text-xs text-gray-600 dark:text-gray-400">
                                {{ getMembershipStatusLabel(membership) }}
                                <span v-if="
                                  membershipExpiresInFuture(membership) &&
                                  !isCustomBackgroundsMembership(membership) &&
                                  (
                                    (membership.platform_type !== PlatformTypeDiscord && membership.platform_type !== PlatformTypePatreon)
                                    || getDaysLeftOfMembership(membership) <= 30
                                  )
                                ">• {{ getMembershipDurationLeft(membership) }}</span>
                              </span>
                            </p>
                          </div>
                        </div>
                      </div>
                      <div class="flex-shrink-0">
                        <button class="items-center rounded-full text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2" @click="showDiscordPopup = true" v-if="membership.platform_type == PlatformTypeDiscord">?</button>
                        <Menu as="div" class="relative inline-block text-left" v-else-if="isMembershipAssignable(membership)">
                          <div>
                            <MenuButton class="flex items-center rounded-full text-gray-400 hover:text-gray-600 focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2">
                              <span class="sr-only">Open options</span>
                              <DotsVerticalIcon class="h-5 w-5" aria-hidden="true" />
                            </MenuButton>
                          </div>

                          <transition enter-active-class="transition ease-out duration-100" enter-from-class="transform opacity-0 scale-95" enter-to-class="transform opacity-100 scale-100" leave-active-class="transition ease-in duration-75" leave-from-class="transform opacity-100 scale-100" leave-to-class="transform opacity-0 scale-95">
                            <MenuItems class="absolute right-0 z-10 mt-2 w-56 origin-top-right rounded-md bg-white dark:bg-secondary shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none">
                              <div class="py-1">
                                <MenuItem v-slot="{ active }">
                                  <button v-if="membership.guild_id == this.$store.getters.getSelectedGuildID" @click="removeMembership(membership)" type="button" :class="[active ? 'hover:bg-gray-50 dark:hover:bg-secondary-light' : '', 'block px-4 py-2 text-sm w-full']">Remove membership</button>
                                  <button v-else-if="membership.guild_id > 0" @click="addMembership(membership)" type="button" :class="[active ? 'hover:bg-gray-50 dark:hover:bg-secondary-light' : '', 'block px-4 py-2 text-sm w-full']">Transfer membership</button>
                                  <button v-else @click="addMembership(membership)" type="button" :class="[active ? 'hover:bg-gray-50 dark:hover:bg-secondary-light' : '', 'block px-4 py-2 text-sm w-full']">Add membership</button>
                                </MenuItem>
                              </div>
                            </MenuItems>
                          </transition>
                        </Menu>
                      </div>
                    </div>
                  </div>
                </li>
              </ul>
            </div>

          </div>
        </div>

        <Popup :open="showDiscordPopup" @close="showDiscordPopup = false">
          <template v-slot:title>
            Managing your Discord subscription
          </template>

          <p>
            To manage your Discord subscription, go to <b>User Settings</b> → <b>Subscriptions</b> → <b>App Subscriptions</b> on your discord client.
            This will let you see all the current and past subscriptions you have with Welcomer. You can cancel or change your payment method from there.
          </p>
          <p class=" text-sm opacity-75">
            Discord subscriptions are not fully supported on mobile devices. Please use the desktop app or use your web browser to manage your subscriptions.
          </p>
          <p class="mt-8">
            <a href="/support" target="_blank" class="mt-4 text-primary">Need help?</a>
          </p>

          <img src="/assets/discord_subscription.png" alt="Screenshot of an active discord subscription on the discord web client" class="mt-4" />
        </Popup>

        <Popup :open="showPaypalPopup" @close="showPaypalPopup = false">
          <template v-slot:title>
            Managing your Paypal subscription
          </template>
        </Popup>

        <Popup :open="showPatreonPopup" @close="showPatreonPopup = false">
          <template v-slot:title>
            Managing your Patreon subscription
          </template>
        </Popup>

        <div class="border-primary bg-primary text-white border p-6 lg:p-12 rounded-lg shadow-sm h-fit mt-16">
          <h3 class="text-2xl font-bold sm:text-3xl">
            Like what you see?
          </h3>
          <p class="mt-4 text-sm leading-6">Unlock more Welcomer features or custom Welcomer backgrounds on any server you choose. Select from monthly, biannual or yearly plans to suit your needs.</p>

          <a href="/premium" target="_blank" type="button" class="bg-white hover:bg-gray-200 flex items-center justify-center px-5 py-3 mt-8 text-base font-medium text-primary border border-transparent rounded-md cursor-pointer w-full">
            Learn More
          </a>
        </div>

        <unsaved-changes :unsavedChanges="unsavedChanges" :isChangeInProgress="isChangeInProgress"
          @save="saveConfig"></unsaved-changes>
      </div>
    </div>
  </div>
</template>

<script>
import { ref } from "vue";

import {
  FormTypeBlank,
  FormTypeToggle,
} from "@/components/dashboard/FormValueEnum";

import UnsavedChanges from "@/components/dashboard/UnsavedChanges.vue";
import EmbedBuilder from "@/components/dashboard/EmbedBuilder.vue";
import FormValue from "@/components/dashboard/FormValue.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";
import Popup from "@/components/Popup.vue";

import userAPI from "@/api/user";
import endpoints from "@/api/endpoints";

import {
  getErrorToast,
  getSuccessToast,
  getValidationToast,
  navigateToErrors,
} from "@/utilities";

import { Menu, MenuButton, MenuItem, MenuItems } from '@headlessui/vue'
import { DotsVerticalIcon } from '@heroicons/vue/solid'
import {
  OpenPatreonLink,
  PlatformTypePatreon,
  PlatformTypePaypal,
  PlatformTypePaypalSubscription,
  PlatformTypeDiscord,
} from "../../constants";

export default {
  components: {
    FormValue,
    EmbedBuilder,
    UnsavedChanges,
    LoadingIcon,
    Menu,
    MenuButton,
    MenuItem,
    MenuItems,
    DotsVerticalIcon,
    Popup,
  },
  setup() {
    let isDataFetched = ref(false);
    let isDataError = ref(false);
    let unsavedChanges = ref(false);
    let isChangeInProgress = ref(false);

    let memberships = ref([]);
    let accounts = ref([]);

    let showDiscordPopup = ref(false);
    let showPaypalPopup = ref(false);
    let showPatreonPopup = ref(false);

    return {
      FormTypeBlank,
      FormTypeToggle,

      isDataFetched,
      isDataError,
      unsavedChanges,
      isChangeInProgress,

      accounts,
      memberships,

      PlatformTypeDiscord,
      PlatformTypePatreon,
      PlatformTypePaypal,
      PlatformTypePaypalSubscription,

      showDiscordPopup,
      showPaypalPopup,
      showPatreonPopup,
    };
  },

  mounted() {
    this.fetchConfig();
  },

  methods: {
    isMembershipAssignable(membership) {
      return (this.isMembershipIdle(membership) || this.isMembershipActive(membership)) && membership.platform_type !== PlatformTypeDiscord;
    },

    isMembershipActive(membership) {
      return membership.membership_status === "active";
    },

    isMembershipIdle(membership) {
      return membership.membership_status === "idle" || membership.guild_id === 0;
    },

    isCustomBackgroundsMembership(membership) {
      return membership.membership_type === "customBackgrounds" || membership.membership_type === "legacyCustomBackgrounds";
    },

    getAccountByPlatform(platform) {
      return this.accounts.find((account) => account.platform === platform);
    },

    fetchConfig() {
      this.isDataFetched = false;
      this.isDataError = false;

      function cmp(a, b) {
        if (a > b) return +1;
        if (a < b) return -1;
        return 0;
      }

      userAPI.getMemberships(
        ({ memberships, accounts }) => {
          this.memberships = memberships
            .sort((a, b) => {
              return cmp(b.guild_id == this.$store.getters.getSelectedGuildID, a.guild_id == this.$store.getters.getSelectedGuildID) || cmp(this.isMembershipAssignable(b), this.isMembershipAssignable(a)) || cmp(b.membership_type, a.membership_type) || cmp(a.expires_at, b.expires_at);
            });
          this.accounts = accounts;
          this.isDataFetched = true;
          this.isDataError = false;
        },
        (error) => {
          this.$store.dispatch("createToast", getErrorToast(error));

          this.isDataFetched = true;
          this.isDataError = true;
        }
      );
    },

    removeMembership(membership) {
      if (confirm("Are you sure you want to remove this membership?")) {
        userAPI.assignMembership(
          membership.membership_uuid,
          null,
          () => {
            this.$store.dispatch("createToast", {
              title: "Membership removed from your server.",
              icon: "check",
              class: "text-green-500 bg-green-100",
            });

            this.fetchConfig();
          },
          (error) => {
            this.$store.dispatch("createToast", getErrorToast(error));
          }
        );
      }
    },

    addMembership(membership) {
      userAPI.assignMembership(
        membership.membership_uuid,
        this.$store.getters.getSelectedGuildID,
        () => {
          this.$store.dispatch("createToast", {
            title: "🎉 Membership assigned from your server.",
            icon: "check",
            class: "text-green-500 bg-green-100",
          });

          this.fetchConfig();
        },
        (error) => {
          this.$store.dispatch("createToast", getErrorToast(error));
        }
      );
    },

    membershipExpiresInFuture(membership) {
      const expiresAt = new Date(membership.expires_at);
      const now = new Date();

      return expiresAt > now;
    },

    getDaysLeftOfMembership(membership) {
      const expiresAt = new Date(membership.expires_at);
      const now = new Date();

      const diff = expiresAt - now;

      const days = Math.floor(diff / (1000 * 60 * 60 * 24));

      return days;
    },

    getMembershipDurationLeft(membership) {
      const days = this.getDaysLeftOfMembership(membership);

      if (days > 0) {
        return `Expires in ${days} days`;
      }

      const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));

      return `Expires in ${hours} hours`;
    },

    getPatreonTierName(tier_id) {
      switch (tier_id) {
        case 23606682:
          return "Welcomer Pro";
        default:
          return "Unknown Tier";
      }
    },

    getMembershipTypeLabel(membership) {
      switch (membership.membership_type) {
        case "unknown":
          return "Unknown";
        case "legacyCustomBackgrounds":
          return "Custom Backgrounds";
        case "legacyWelcomerPro":
          return "Welcomer Pro";
        case "welcomerPro":
          return "Welcomer Pro";
        case "customBackgrounds":
          return "Custom Backgrounds";
        default:
          return "Unknown";
      }
    },

    getMembershipStatusLabel(membership) {
      switch (membership.membership_status) {
        case "active":
          return "Active";
        case "expired":
          return "Expired";
        case "refunded":
          return "Refunded";
        case "removed":
          return "Removed";
        default:
          return "Idle";
      }
    },

    gotoPatreonLink() {
      OpenPatreonLink(this.fetchConfig)
    },

    async saveConfig() {
      const validForm = await this.v$.$validate();

      if (!validForm) {
        this.$store.dispatch("createToast", getValidationToast());
        navigateToErrors();

        return;
      }

      this.isChangeInProgress = true;

      dashboardAPI.setConfig(
        endpoints.EndpointGuild(this.$store.getters.getSelectedGuildID),
        this.config,
        null,
        ({ config }) => {
          this.$store.dispatch("createToast", getSuccessToast());

          this.config = config;
          this.unsavedChanges = false;
          this.isChangeInProgress = false;
        },
        (error) => {
          this.$store.dispatch("createToast", getErrorToast(error));

          this.isChangeInProgress = false;
        }
      );
    },

    onValueUpdate() {
      this.unsavedChanges = true;
    },
  },
};
</script>
