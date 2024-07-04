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
          <div class="dashboard-title">Rules</div>
        </div>
        <div class="dashboard-contents">
          <div class="dashboard-inputs">
            <form-value title="Enable Rules" :type="FormTypeToggle" v-model="config.enabled"
              @update:modelValue="onValueUpdate" :validation="v$.enabled">Send rules to users when they join your server.
              This also allows
              users to view the rules by doing <code>/rules</code>.</form-value>
            <form-value title="Enable DMs" :type="FormTypeToggle" v-model="config.dms_enabled"
              @update:modelValue="onValueUpdate" :validation="v$.dms_enabled">When enabled, users will
              also receive the rules in their direct
              messages.</form-value>

            <form-value title="Rules" :type="FormTypeBlank" :hideBorder="true">
              <table class="min-w-full border-spacing-2">
                <tbody class="divide-y divide-gray-200 dark:divide-secondary-light">
                  <tr v-for="(rule, index) in this.rules" :key="index" :class="[
                    this.selectedIndex != null ? 'select-none' : '',
                    this.selectedIndex == index
                      ? 'dark:bg-secondary-dark bg-gray-100'
                      : '',
                  ]" v-on:mousemove="this.mouseMoveHandler(index)">
                    <td :class="[
                      'pr-3 whitespace-nowrap py-4 text-sm dark:text-gray-50 space-x-2 grid grid-cols-2',
                      this.isDraggable ? 'cursor-move' : '',
                    ]" v-on:mousedown="this.mouseDownHandler(index)">
                      <font-awesome-icon v-if="this.isDraggable" icon="grip-vertical" />
                      <a v-if="!this.isDraggable" @click="this.moveRule(index, index - 1)">
                        <font-awesome-icon :class="[index > 0 ? '' : 'opacity-20 touch-none']" icon="chevron-up" />
                      </a>
                      <a v-if="!this.isDraggable" @click="this.moveRule(index, index + 1)">
                        <font-awesome-icon :class="[
                          index < this.rules.length - 1
                            ? ''
                            : 'opacity-20 touch-none',
                        ]" icon="chevron-down" />
                      </a>
                    </td>
                    <td class="pr-3 text-sm dark:text-gray-50 w-auto">
                      <input v-if="rule.selected" type="text"
                        class="bg-white dark:bg-secondary-dark relative w-full pl-3 pr-3 text-left border border-gray-300 dark:border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm"
                        v-model="rule.newValue" :maxlength="this.maxRuleLength"
                        @keypress="this.onEditRuleKeyPress($event, index)" />
                      <div class="break-all" v-else v-html="marked(rule.value, true)" />
                    </td>
                    <td class="whitespace-nowrap py-4 text-sm dark:text-gray-50 space-x-2">
                      <a v-if="rule.selected" @click="this.onSaveRule(index)"
                        class="text-primary hover:text-primary-dark cursor-pointer">Confirm</a>
                      <a v-else @click="this.onSelectRule(index)"
                        class="text-primary hover:text-primary-dark cursor-pointer">Edit</a>
                      <a v-if="rule.selected" @click="this.onCancelRule(index)"
                        class="text-primary hover:text-primary-dark cursor-pointer">Cancel</a>
                      <a v-else @click="this.onDeleteRule(index)"
                        class="text-primary hover:text-primary-dark cursor-pointer">Delete</a>
                    </td>
                  </tr>
                  <tr>
                    <td />
                    <td>
                      <input v-if="this.rules.length < this.maxRuleCount" type="text"
                        class="bg-white dark:bg-secondary-dark relative w-full pl-3 pr-10 mt-2 text-left border border-gray-300 dark:border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm"
                        placeholder="Add rule" :maxlength="this.maxRuleLength" @blur="this.onRuleBlur()"
                        @keypress="this.onRuleKeyPress($event)" v-model="rule" />
                    </td>
                  </tr>
                </tbody>
              </table>
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

import { toHTML } from "@/components/discord-markdown";

import {
  FormTypeBlank,
  FormTypeToggle,
} from "@/components/dashboard/FormValueEnum";

import FormValue from "@/components/dashboard/FormValue.vue";
import UnsavedChanges from "@/components/dashboard/UnsavedChanges.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";

import endpoints from "@/api/endpoints";
import dashboardAPI from "@/api/dashboard";

import {
  getErrorToast,
  getSuccessToast,
  getValidationToast,
  navigateToErrors,
} from "@/utilities";

const maxRuleCount = 25;
const maxRuleLength = 250;

export default {
  components: {
    FormValue,
    UnsavedChanges,
    LoadingIcon,
  },
  setup() {
    let isDataFetched = ref(false);
    let isDataError = ref(false);
    let unsavedChanges = ref(false);
    let isChangeInProgress = ref(false);

    let config = ref({});
    let rules = ref([]);

    const validation_rules = computed(() => {
      const validation_rules = {
        enabled: {},
        dms_enabled: {},
        rules: {
          required: helpers.withMessage(
            "The rules are required",
            requiredIf(config.value.enabled)
          ),
        },
      };

      return validation_rules;
    });

    const v$ = useVuelidate(validation_rules, config, { $rewardEarly: true });

    let rule = ref("");

    let selectedIndex = ref(null);

    let isDraggable = matchMedia("(pointer:fine)").matches;

    return {
      FormTypeBlank,
      FormTypeToggle,

      isDataFetched,
      isDataError,
      unsavedChanges,
      isChangeInProgress,

      config,
      rules,
      v$,

      rule,

      maxRuleCount,
      maxRuleLength,

      selectedIndex,

      isDraggable,
    };
  },

  beforeDestroy() {
    window.removeEventListener("mouseup", this.mouseUpHandler);
  },

  mounted() {
    window.addEventListener("mouseup", this.mouseUpHandler);

    this.fetchConfig();
  },

  methods: {
    setConfig(config) {
      this.config = config;

      this.rules = [];
      this.config.rules.forEach((rule) => {
        this.rules.push({
          value: rule,
          selected: false,
        });
      });
    },

    fetchConfig() {
      this.isDataFetched = false;
      this.isDataError = false;

      dashboardAPI.getConfig(
        endpoints.EndpointGuildRules(this.$store.getters.getSelectedGuildID),
        ({ config }) => {
          this.setConfig(config);
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

      this.config.rules = [];
      this.rules.forEach((rule) => {
        this.config.rules.push(rule.value);
      });

      dashboardAPI.setConfig(
        endpoints.EndpointGuildRules(this.$store.getters.getSelectedGuildID),
        this.config,
        null,
        ({ config }) => {
          this.$store.dispatch("createToast", getSuccessToast());

          this.setConfig(config);
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

    marked(input, embed) {
      if (input) {
        return toHTML(input, {
          embed: embed,
          discordCallback: {
            user: function (user) {
              return `@${user.id}`;
            },
            channel: function (channel) {
              return `#${channel.id}`;
            },
            role: function (role) {
              return `@${role.id}`;
            },
            everyone: function () {
              return `@everyone`;
            },
            here: function () {
              return `@here`;
            },
          },
          cssModuleNames: {
            "d-emoji": "emoji",
          },
        });
      }
      return "";
    },

    onSelectRule(index) {
      this.rules.forEach((rule) => {
        rule.selected = false;
      });

      this.rules[index].selected = true;
      this.rules[index].newValue = this.rules[index].value;
    },

    onSaveRule(index) {
      if (this.rules[index].newValue.trim() == "") {
        this.onDeleteRule(index);
      } else {
        this.rules[index].selected = false;

        if (this.rules[index].value !== this.rules[index].newValue) {
          this.onValueUpdate();
        }

        this.rules[index].value = this.rules[index].newValue;
      }
    },

    onCancelRule(index) {
      this.rules[index].selected = false;
    },

    onDeleteRule(index) {
      this.rules.splice(index, 1);
      this.onValueUpdate();
    },

    onEditRuleKeyPress(event, index) {
      if (event.key === "Enter") {
        event.preventDefault();
        this.onSaveRule(index);
      }
    },

    onRuleKeyPress(event) {
      if (event.key === "Enter") {
        event.preventDefault();
        this.onRuleBlur();
      }
    },

    onRuleBlur() {
      this.rule = this.rule.trim();

      if (this.rule != "") {
        this.rules.push({
          value: this.rule,
          selected: false,
        });
        this.rule = "";
        this.onValueUpdate();
      }
    },

    mouseDownHandler(index) {
      this.selectedIndex = index;
    },

    mouseUpHandler() {
      this.selectedIndex = null;
    },

    mouseMoveHandler(index) {
      if (this.selectedIndex != null && this.selectedIndex != index) {
        this.moveRule(index, this.selectedIndex);
        this.selectedIndex = index;
      }
    },

    moveRule(index, newIndex) {
      if (index >= 0 && index <= this.rules.length - 1) {
        var temp = this.rules[index];
        this.rules[index] = this.rules[newIndex];
        this.rules[newIndex] = temp;
        this.onValueUpdate();
      }
    },
  },
};
</script>
