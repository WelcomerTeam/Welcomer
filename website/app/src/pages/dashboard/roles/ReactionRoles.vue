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
          <div class="dashboard-title">Reaction Roles</div>
        </div>
        <div class="dashboard-contents">
          <!-- <discord-emoji-picker
            customGroupLabel="Server Emojis"
            :customEmojis="$store.getters.getCurrentSelectedGuild.emojis" /> -->
          <div class="dashboard-inputs">
            <form-value title="Enable Reaction Roles" :type="FormTypeToggle" v-model="config.enabled"
            @update:modelValue="onValueUpdate" :validation="v$.enabled">Reaction Roles allow users to assign themselves roles by reacting to messages.</form-value>
            
            
            <form-value title="Configurations" type="FormTypeBlank" :hideBorder="true" :validation="v$.roles">
              <reaction-roles-table :modelValue="config"
              @update:modelValue="onConfigUpdate"></reaction-roles-table>
            </form-value>
            {{  config  }}
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

import UnsavedChanges from "@/components/dashboard/UnsavedChanges.vue";
import EmbedBuilder from "@/components/dashboard/EmbedBuilder.vue";
import FormValue from "@/components/dashboard/FormValue.vue";
import RoleTable from "@/components/dashboard/RoleTable.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";
import DiscordEmojiPicker from "@/components/DiscordEmojiPicker.vue";

import dashboardAPI from "@/api/dashboard";
import endpoints from "@/api/endpoints";

import {
  getErrorToast,
  getSuccessToast,
  getValidationToast,
  navigateToErrors,
} from "@/utilities";

import {
  FormTypeBlank,
  FormTypeToggle,
} from "@/components/dashboard/FormValueEnum";

import ReactionRolesTable from "@/components/dashboard/ReactionRolesTable.vue";

export default {
  components: {
    EmbedBuilder,
    DiscordEmojiPicker,
    FormValue,
    LoadingIcon,
    ReactionRolesTable,
    RoleTable,
    UnsavedChanges,
  },
  setup() {
    let isDataFetched = ref(false);
    let isDataError = ref(false);
    let unsavedChanges = ref(false);
    let isChangeInProgress = ref(false);
    
    let config = ref({});
    
    const validation_rules = computed(() => {
      const validation_rules = {
      };
      
      return validation_rules;
    });
    
    const v$ = useVuelidate(validation_rules, config, { $rewardEarly: true });
    
    return {
      FormTypeBlank,
      FormTypeToggle,
      
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
      endpoints.EndpointGuildReactionRoles(this.$store.getters.getSelectedGuildID),
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
  },
};
</script>
