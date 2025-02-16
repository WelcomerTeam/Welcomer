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
          <div class="dashboard-title">TimeRoles</div>
        </div>
        <div class="dashboard-contents">
          <div class="dashboard-inputs">
            <form-value title="Enable TimeRoles" :type="FormTypeToggle" v-model="config.enabled"
              @update:modelValue="onValueUpdate" :validation="v$.enabled">TimeRoles allow you to reward users for being in your server for a period of time, by giving them roles.</form-value>

            <form-value title="Roles" :type="FormTypeBlank" :hideBorder="true" :validation="v$.roles">
              <role-table-time-roles :roles="$store.getters.getAssignableGuildRoles" :selectedRoles="config.roles"
                @removeRole="onRemoveRole" @selectRole="onSelectRole" @updateRole="onUpdateRole"></role-table-time-roles>
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
} from "@/components/dashboard/FormValueEnum";

import UnsavedChanges from "@/components/dashboard/UnsavedChanges.vue";
import EmbedBuilder from "@/components/dashboard/EmbedBuilder.vue";
import FormValue from "@/components/dashboard/FormValue.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";
import RoleTableTimeRoles from "@/components/dashboard/RoleTableTimeRoles.vue";

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
    RoleTableTimeRoles,
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
        enabled: {},
        roles: {
          required: helpers.withMessage("No roles have been selected", requiredIf(config.value.enabled))
        },
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
        endpoints.EndpointGuildTimeroles(this.$store.getters.getSelectedGuildID),
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

      dashboardAPI.setConfig(
        endpoints.EndpointGuildTimeroles(this.$store.getters.getSelectedGuildID),
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

    onSelectRole(roleID) {
      let role = this.$store.getters.getGuildRoleById(roleID);
      if (role !== undefined) {
        this.config.roles.push({
          role_id: role.id,
          seconds: 0
        });
        this.config.roles.sort(
          (a, b) =>
            this.$store.getters.getGuildRoleById(a)?.position -
            this.$store.getters.getGuildRoleById(b)?.position
        );
        this.onValueUpdate();
      }
    },

    onRemoveRole(roleID) {
      this.config.roles = this.config.roles.filter((role) => role.role_id !== roleID);
      this.onValueUpdate();
    },

    onUpdateRole({ roleID, seconds }) {
      this.config.roles.forEach((role) => {
        if (role.role_id == roleID) {
          role.seconds = seconds;
        }

        if (role.seconds == "") {
          role.seconds = 0;
        }
      })
      this.onValueUpdate();
    }
  },
};
</script>
