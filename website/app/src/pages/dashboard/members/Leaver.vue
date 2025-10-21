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
          <div class="dashboard-title">Leaver</div>
        </div>
        <div class="dashboard-contents">
          <div class="dashboard-inputs">
            <div class="dashboard-heading">Auto Deletion</div>
            <form-value title="Auto Delete Leaver Messages" :type="FormTypeToggle"
              v-model="config.auto_delete_leaver_messages" @update:modelValue="onValueUpdate"
              :validation="v$.auto_delete_leaver_messages" :hideBorder="true" :disabled="!$store.getters.guildHasWelcomerPro"></form-value>
            
            <div class="pl-8 border-b border-gray-300 dark:border-secondary-light">
              <form-value title="Message Lifetime" :type="FormTypeDuration" v-model="config.leaver_message_lifetime"
                @update:modelValue="onValueUpdate" :validation="v$.leaver_message_lifetime"
                :hideBorder="true" :disabled="!config.auto_delete_leaver_messages || !$store.getters.guildHasWelcomerPro">This is the duration before a leaver message
                is automatically deleted.</form-value>
              <div v-if="!$store.getters.guildHasWelcomerPro" class="border-primary text- border p-4 rounded-lg shadow-sm h-fit mt-4 text-secondary dark:text-gray-50 mb-4">
                Auto deletion of leaver messages requires a Welcomer Pro subscription.
                <a href="/premium" class="underline">Learn more</a>
              </div>
            </div>
          </div>
          <div class="dashboard-inputs">
            <form-value title="Enable Leaver" :type="FormTypeToggle" v-model="config.enabled"
              @update:modelValue="onValueUpdate" :validation="v$.enabled">Send messages in a channel when users leave your
              server.</form-value>

            <form-value title="Leaver Channel" :type="FormTypeChannelListCategories" v-model="config.channel"
              @update:modelValue="onValueUpdate" :validation="v$.channel" :inlineSlot="true" :nullable="true">This is the
              channel we will send leaver messages to.</form-value>

            <form-value title="Leaver Message" :type="FormTypeEmbed" v-model="config.message_json"
              @update:modelValue="onValueUpdate" :validation="v$.message_json" :inlineSlot="true">This is the message that
              will be sent when users leave.
              <a target="_blank" href="/formatting" class="text-primary hover:text-primary-dark">Click here</a>
              to view all the formatting tags you can use for custom text.
            </form-value>
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
import { helpers, requiredIf } from "@vuelidate/validators";

import {
  FormTypeBlank,
  FormTypeToggle,
  FormTypeChannelListCategories,
  FormTypeEmbed,
  FormTypeDuration,
} from "@/components/dashboard/FormValueEnum";

import EmbedBuilder from "@/components/dashboard/EmbedBuilder.vue";
import FormValue from "@/components/dashboard/FormValue.vue";
import UnsavedChanges from "@/components/dashboard/UnsavedChanges.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";

import dashboardAPI from "@/api/dashboard";
import endpoints from "@/api/endpoints";

import {
  getErrorToast,
  getSuccessToast,
  getValidationToast,
  navigateToErrors,
  isValidJson,
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

    let config = ref({});

    const validation_rules = computed(() => {
      const validation_rules = {
        enabled: {},
        auto_delete_leaver_messages: {},
        leaver_message_lifetime: {
          minValue: helpers.withMessage("The lifetime is not valid", (value) => {
            return value === undefined || value === null || value >= 0;
          }),
          maxValue: helpers.withMessage("The lifetime cannot be longer than one day", (value) => {
            return value === undefined || value === null || value <= 86400;
          }),
        },
        channel: {
          required: helpers.withMessage("The channel must be selected if you do not send a DM", requiredIf(
            config.value.enabled
          ))
        },
        message_json: {
          required: helpers.withMessage("The message is required", requiredIf(
            config.value.enabled
          )),
          isValidJson: helpers.withMessage("The message is not valid JSON", (value) => {
            return !value || isValidJson(value);
          }),
        },
      };

      return validation_rules;
    });

    const v$ = useVuelidate(validation_rules, config, { $rewardEarly: true });

    return {
      FormTypeBlank,
      FormTypeToggle,
      FormTypeChannelListCategories,
      FormTypeEmbed,
      FormTypeDuration,

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
        endpoints.EndpointGuildLeaver(this.$store.getters.getSelectedGuildID),
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
        endpoints.EndpointGuildLeaver(this.$store.getters.getSelectedGuildID),
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
