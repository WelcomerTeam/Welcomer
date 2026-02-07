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
          <div class="dashboard-inputs">
            <form-value title="Enable Reaction Roles" :type="FormTypeToggle" v-model="config.enabled"
                        @update:modelValue="onValueUpdate" :validation="v$.enabled">Reaction Roles allow users to assign themselves roles by reacting to messages.</form-value>
            
            
            <form-value title="Reaction Roles" type="FormTypeBlank" :validation="v$.roles">
              <reaction-roles-table :modelValue="config"
                                    @update:modelValue="onValueUpdate"></reaction-roles-table>
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

import dashboardAPI from "@/api/dashboard";
import endpoints from "@/api/endpoints";
import EmbedBuilder from "@/components/dashboard/EmbedBuilder.vue";
import FormValue from "@/components/dashboard/FormValue.vue";
import {
  FormTypeBlank,
  FormTypeToggle,
} from "@/components/dashboard/FormValueEnum";
import ReactionRolesTable from "@/components/dashboard/ReactionRolesTable.vue";
import RoleTable from "@/components/dashboard/RoleTable.vue";
import UnsavedChanges from "@/components/dashboard/UnsavedChanges.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";
import {
  getErrorToast,
  getSuccessToast,
} from "@/utilities";

export default {
  components: {
    EmbedBuilder,
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

    saveConfig() {
      this.isChangeInProgress = true;

      dashboardAPI.doPost(
        endpoints.EndpointGuildReactionRoles(this.$store.getters.getSelectedGuildID),
        this.config,
        null,
        ({ config }) => {
          this.$store.dispatch("createToast", getSuccessToast())

          this.config = config;
          this.isChangeInProgress = false;
          this.unsavedChanges = false;
        },
        (error) => {
          this.$store.dispatch("createToast", getErrorToast(error));

          this.isChangeInProgress = false;
        }
      );
    },

    onValueUpdate() {
      this.unsavedChanges = true;
    }
  },
};
</script>
