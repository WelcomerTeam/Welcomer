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
          <div class="dashboard-title">Bot Customisation</div>
        </div>
        <div class="dashboard-contents">
          <div class="dashboard-inputs">
              <!-- Premium Lock Message -->
              <div v-if="!$store.getters.guildHasWelcomerPro" class="mb-4">
                <div class="border-primary bg-primary text-white border p-6 lg:p-12 rounded-lg shadow-sm h-fit">
                  <h3 class="text-2xl font-bold sm:text-3xl">
                    You've found a premium feature!
                  </h3>
                  <p class="mt-4 text-sm leading-6">Upgrade to Welcomer Pro to customize your bot's appearance on this server.</p>

                  <a href="/premium" target="_blank" type="button" class="bg-white hover:bg-gray-200 flex items-center justify-center px-5 py-3 mt-8 text-base font-medium text-primary border border-transparent rounded-md cursor-pointer w-full">
                    Learn More
                  </a>
                </div>
              </div>

              <!-- Avatar Section -->
              <div class="mb-8">
                <div class="dashboard-heading">Avatar</div>
                <div class="flex items-center space-x-4 mb-4">
                  <div class="w-24 h-24 rounded-lg overflow-hidden bg-gray-200 dark:bg-secondary-light flex items-center justify-center">
                    <img v-if="config.avatarPreview" :src="config.avatarPreview" alt="Avatar preview" class="w-full h-full object-cover" />
                    <span v-else class="text-gray-500">No avatar</span>
                  </div>
                  <div class="flex-1">
                    <p class="text-sm text-gray-600 dark:text-gray-400 mb-2">The recommended size is 512x512 pixels, max 5MB</p>
                    <input
                      ref="avatarInput"
                      type="file"
                      accept="image/png,image/jpeg,image/webp"
                      @change="onAvatarSelect"
                      class="hidden" />
                    <button
                      @click="$refs.avatarInput.click()"
                      :disabled="!$store.getters.guildHasWelcomerPro"
                      class="px-4 py-2 bg-primary text-white rounded-md hover:bg-primary-dark transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                      Upload Avatar
                    </button>
                    <button
                      v-if="config.avatar || config.avatarPreview"
                      @click="clearAvatar"
                      :disabled="!$store.getters.guildHasWelcomerPro"
                      class="ml-2 px-4 py-2 bg-red-500 text-white rounded-md hover:bg-red-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                      Remove
                    </button>
                  </div>
                </div>
              </div>

              <!-- Banner Section -->
              <div class="mb-8">
                <div class="dashboard-heading">Banner</div>
                <div class="flex flex-col space-y-4 mb-4">
                  <div class="w-full max-w-7xl rounded-lg overflow-hidden bg-gray-200 dark:bg-secondary-light flex items-center justify-center">
                    <div v-if="config.bannerPreview" class="w-full aspect-[17/6]">
                      <img :src="config.bannerPreview" alt="Banner preview" class="w-full h-full object-cover" />
                    </div>
                    <span v-else class="text-gray-500">No banner</span>
                  </div>
                  <div>
                    <p class="text-sm text-gray-600 dark:text-gray-400 mb-2">The recommended size is 1024x256 pixels, max 10MB</p>
                    <input
                      ref="bannerInput"
                      type="file"
                      accept="image/png,image/jpeg,image/webp"
                      @change="onBannerSelect"
                      class="hidden" />
                    <button
                      @click="$refs.bannerInput.click()"
                      :disabled="!$store.getters.guildHasWelcomerPro"
                      class="px-4 py-2 bg-primary text-white rounded-md hover:bg-primary-dark transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                      Upload Banner
                    </button>
                    <button
                      v-if="config.banner || config.bannerPreview"
                      @click="clearBanner"
                      :disabled="!$store.getters.guildHasWelcomerPro"
                      class="ml-2 px-4 py-2 bg-red-500 text-white rounded-md hover:bg-red-600 transition-colors disabled:opacity-50 disabled:cursor-not-allowed">
                      Remove
                    </button>
                  </div>
                </div>
              </div>

              <!-- Nickname Section -->
              <form-value title="Bot Nickname" :type="FormTypeText" v-model="config.nickname"
                @update:modelValue="onValueUpdate" :validation="v$.nickname" :disabled="!$store.getters.guildHasWelcomerPro">Set a custom nickname for your bot.</form-value>

              <!-- Bio Section -->
              <form-value title="Bot Bio" :type="FormTypeTextArea" v-model="config.bio"
                @update:modelValue="onValueUpdate" :validation="v$.bio" :disabled="!$store.getters.guildHasWelcomerPro">Set a custom bio for your bot.</form-value>

              <unsaved-changes :unsavedChanges="unsavedChanges" :isChangeInProgress="isChangeInProgress"
                v-on:save="saveConfig"></unsaved-changes>
            </div>
          </div>
        </div>
      </div>
  </div>
</template>

<script>
import useVuelidate from "@vuelidate/core";
import { helpers } from "@vuelidate/validators";
import { computed, ref } from "vue";
import LoadingIcon from "@/components/LoadingIcon.vue";
import FormValue from "@/components/dashboard/FormValue.vue";
import UnsavedChanges from "@/components/dashboard/UnsavedChanges.vue";
import {
  FormTypeText,
  FormTypeTextArea,
} from "@/components/dashboard/FormValueEnum";
import dashboardAPI from "@/api/dashboard";
import endpoints from "@/api/endpoints";
import { getErrorToast, getSuccessToast, getValidationToast, navigateToErrors } from "@/utilities";

export default {
  components: {
    LoadingIcon,
    FormValue,
    UnsavedChanges,
  },
  setup() {
    let isDataFetched = ref(false);
    let isDataError = ref(false);
    let unsavedChanges = ref(false);
    let isChangeInProgress = ref(false);

    let config = ref({
      nickname: "",
      bio: "",
      avatar: "",
      banner: "",
      avatarPreview: "",
      bannerPreview: "",
    });

    const validation_rules = computed(() => {
      const validation_rules = {
        nickname: {
          maxLength: helpers.withMessage("Nickname cannot exceed 32 characters", (value) => {
              return value.length <= 32
            }),
        },
        bio: {
          maxLength: helpers.withMessage("Bio cannot exceed 190 characters", (value) => {
              return value.length <= 190
            }),
        },
      };

      return validation_rules;
    });

    const v$ = useVuelidate(validation_rules, config, { $rewardEarly: true });

    return {
      FormTypeText,
      FormTypeTextArea,

      isDataFetched,
      isDataError,
      unsavedChanges,
      isChangeInProgress,

      config,
      v$,
    };
  },
  mounted() {
    this.fetchConfig();
  },
  methods: {
    fetchConfig() {
      this.isDataFetched = false;
      this.isDataError = false;

      dashboardAPI.getConfig(
        endpoints.EndpointGuildCustomisation(this.$store.getters.getSelectedGuildID),
        ({ config }) => {
          this.config = {
            nickname: config.nickname,
            bio: config.bio,
            avatar: null,
            banner: null,
            avatarPreview: config.avatar ? `https://cdn.discordapp.com/guilds/${this.$store.getters.getSelectedGuildID}/users/${config.user_id}/avatars/${config.avatar}.png` : "",
            bannerPreview: config.banner ? `https://cdn.discordapp.com/guilds/${this.$store.getters.getSelectedGuildID}/users/${config.user_id}/banners/${config.banner}.png` : "",
          };
          this.isDataFetched = true;
          this.isDataError = false;
        },
        (error) => {
          this.$store.dispatch("createToast", getErrorToast(error));

          this.isDataFetched = false;
          this.isDataError = true;
        }
      );
    },

    onValueUpdate() {
      this.unsavedChanges = true;
    },

    onAvatarSelect(event) {
      const file = event.target.files?.[0];
      if (!file) return;

      this.readImageFile(file, (base64, error) => {
        if (error) {
          this.$store.dispatch("createToast", getErrorToast(error));
          return;
        }

        this.config.avatar = base64;
        this.config.avatarPreview = base64;

        this.onValueUpdate();
      }, 5 * 1024 * 1024, 1024, 1024);
    },

    onBannerSelect(event) {
      const file = event.target.files?.[0];
      if (!file) return;

      this.readImageFile(file, (base64, error) => {
        if (error) {
          this.$store.dispatch("createToast", getErrorToast(error));
          return;
        }

        this.config.banner = base64;
        this.config.bannerPreview = base64;

        this.onValueUpdate();
      }, 10 * 1024 * 1024, 1024, 256);
    },

    readImageFile(file, callback, maxSize, maxWidth, maxHeight) {
      // Validate file type
      const validTypes = ["image/png", "image/jpeg", "image/webp"];
      if (!validTypes.includes(file.type)) {
        callback(null, "Only PNG, JPEG, and WebP images are supported");
        return;
      }

      const reader = new FileReader();

      reader.onload = () => {
        const base64 = reader.result;

        // Validate dimensions
        const img = new Image();
        img.onload = () => {
          if (maxSize && file.size > maxSize) {
            callback(null, `Image size cannot exceed ${maxSize / (1024 * 1024)}MB`);
            return;
          }

          if (maxWidth && img.width > maxWidth) {
            callback(null, `Image width cannot exceed ${maxWidth}px`);
            return;
          }

          if (maxHeight && img.height > maxHeight) {
            callback(null, `Image height cannot exceed ${maxHeight}px`);
            return;
          }

          callback(base64, null);
        };
        img.onerror = () => {
          callback(null, "Failed to validate image dimensions");
        };
        img.src = base64;
      };

      reader.onerror = () => {
        callback(null, "Failed to read file");
      };

      reader.readAsDataURL(file);
    },

    clearAvatar() {
      this.config.avatar = "";
      this.config.avatarPreview = "";

      this.onValueUpdate();

      if (this.$refs.avatarInput) {
        this.$refs.avatarInput.value = "";
      }
    },

    clearBanner() {
      this.config.banner = "";
      this.config.bannerPreview = "";
      this.onValueUpdate();

      if (this.$refs.bannerInput) {
        this.$refs.bannerInput.value = "";
      }
    },

    async saveConfig() {
      const validForm = await this.v$.$validate();

      if (!validForm) {
        this.$store.dispatch("createToast", getValidationToast());
        navigateToErrors();
        return;
      }

      this.isChangeInProgress = true;

      dashboardAPI.doPost(
        endpoints.EndpointGuildCustomisation(this.$store.getters.getSelectedGuildID),
        {
          nickname: this.config.nickname,
          bio: this.config.bio,
          avatar: this.config.avatar,
          banner: this.config.banner,
        },
        null,
        ({ config }) => {
          this.$store.dispatch("createToast", getSuccessToast());

          this.config = {
            nickname: config.nickname,
            bio: config.bio,
            avatar: null,
            banner: null,
            avatarPreview: config.avatar ? `https://cdn.discordapp.com/guilds/${this.$store.getters.getSelectedGuildID}/users/${config.user_id}/avatars/${config.avatar}.png` : "",
            bannerPreview: config.banner ? `https://cdn.discordapp.com/guilds/${this.$store.getters.getSelectedGuildID}/users/${config.user_id}/banners/${config.banner}.png` : "",
          };

          this.unsavedChanges = false;
          this.isChangeInProgress = false;

          // Clear the file inputs
          if (this.$refs.avatarInput) this.$refs.avatarInput.value = "";
          if (this.$refs.bannerInput) this.$refs.bannerInput.value = "";
        },
        (error) => {
          this.$store.dispatch("createToast", getErrorToast(error));

          this.isChangeInProgress = false;
        }
      );
    },
  },
};
</script>
