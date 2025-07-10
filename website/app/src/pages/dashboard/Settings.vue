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
          <div class="dashboard-title">Bot Settings</div>
        </div>
        <div class="dashboard-contents">
          <div class="dashboard-inputs">
            <form-value title="Embed Colour" :type="FormTypeColour" v-model="config.embed_colour"
            @update:modelValue="onValueUpdate" :validation="v$.embed_colour">This changes the embed colour accent on any commands you run with Welcomer</form-value>
          </div>
          <div class="dashboard-inputs">
            <div class="dashboard-heading">Server Web Page</div>
            <form-value title="Show Server on Website" :type="FormTypeToggle" v-model="config.site_guild_visible"
            @update:modelValue="onValueUpdate" :validation="v$.site_guild_visible">When enabled, users will be able to publicly see your server information on the website.</form-value>

            <form-value title="Show Staff on Website" :type="FormTypeToggle" v-model="config.site_staff_visible"
            @update:modelValue="onValueUpdate" :validation="v$.site_staff_visible">When enabled, your staff will be shown on your server's page on the website.</form-value>

            <form-value title="Allow Users to Join on Website" :type="FormTypeToggle" v-model="config.site_allow_invites"
            @update:modelValue="onValueUpdate" :validation="v$.site_allow_invites">When enabled, users will be able to use Welcomer to get an invite for your server through the website. If you have a vanity invite, this will be used instead.</form-value>
          </div>
          <unsaved-changes :unsavedChanges="unsavedChanges" :isChangeInProgress="isChangeInProgress"
            @save="saveConfig"></unsaved-changes>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { computed, ref } from "vue";

import useVuelidate from "@vuelidate/core";

import {
  FormTypeColour,
  FormTypeToggle,
} from "@/components/dashboard/FormValueEnum";

import UnsavedChanges from "@/components/dashboard/UnsavedChanges.vue";
import EmbedBuilder from "@/components/dashboard/EmbedBuilder.vue";
import FormValue from "@/components/dashboard/FormValue.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";

import dashboardAPI from "@/api/dashboard";
import endpoints from "@/api/endpoints";

import {
  getErrorToast,
  getSuccessToast,
  getValidationToast,
  navigateToErrors,
} from "@/utilities";

export default {
  components: {
    FormValue,
    EmbedBuilder,
    UnsavedChanges,
    LoadingIcon,
  },
  setup() {
    let isDataFetched = ref(false);
    let isDataError = ref(false);
    let unsavedChanges = ref(false);
    let isChangeInProgress = ref(false);

    let assigned_roles = ref([]);
    let roles = ref([]);

    let config = ref({});

    const validation_rules = computed(() => {
      const validation_rules = {
        embed_colour: {},
        site_splash_url: {},
        site_staff_visible: {},
        site_guild_visible: {},
        site_allow_invites: {}
      };

      return validation_rules;
    });

    const v$ = useVuelidate(validation_rules, config, { $rewardEarly: true });

    return {
      FormTypeColour,
      FormTypeToggle,

      isDataFetched,
      isDataError,
      unsavedChanges,
      isChangeInProgress,

      assigned_roles,
      roles,

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
        endpoints.EndpointGuildSettings(this.$store.getters.getSelectedGuildID),
        ({ config }) => {
          this.config = config;
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

    async saveConfig() {
      const validForm = await this.v$.$validate();

      if (!validForm) {
        this.$store.dispatch("createToast", getValidationToast());
        navigateToErrors();

        return;
      }

      this.isChangeInProgress = true;

      dashboardAPI.doPost(
        endpoints.EndpointGuildSettings(this.$store.getters.getSelectedGuildID),
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
