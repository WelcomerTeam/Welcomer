<template>
  <div>
    <div v-if="onboardingStep >= 0">
      <h2 class="text-2xl font-bold mb-8 text-center">Select Message Type</h2>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-2 mb-4">
        <button :disabled="true || !$props.isSetup" @click="$props.reactionRole.is_system_message = false" :class="[$props.reactionRole.is_system_message === false ? ' border-primary' : 'border-gray-300 dark:border-secondary-light', 'p-8 border rounded-lg shadow-sm h-fit mb-4 disabled:bg-secondary-light']">
          <h2 class="font-bold text-lg">Use an existing message</h2>
          <span>Use an existing message sent by any user, you cannot configure this message on the dashboard.</span>
        </button>
        <button :disabled="!$props.isSetup" @click="$props.reactionRole.is_system_message = true" :class="[$props.reactionRole.is_system_message === true ? ' border-primary' : 'border-gray-300 dark:border-secondary-light', 'p-8 border rounded-lg shadow-sm h-fit mb-4 disabled:bg-secondary-light']">
          <h2 class="font-bold text-lg">Create New Message</h2>
          <span>Create a message sent by Welcomer, you can then configure this message on the dashboard.</span>
        </button>
      </div>

      <div v-if="$props.reactionRole.is_system_message === false" class="mb-4">
        <form-value :type="FormTypeText" title="Message Link" v-model="$props.reactionRole.message_link" placeholder="https://discord.com/channels/123456789012345678/123456789012345678/123456789012345678" :validation="v$.message_link">
          Enter the link to the message you want to use for reaction roles. You can get this by selecting a message and clicking <b>Copy Message Link</b> on Discord.
        </form-value>
      </div>
      <div v-if="$props.reactionRole.is_system_message === true" class="mb-4">
        <form-value :type="FormTypeChannelListCategories" title="Channel" v-model="$props.reactionRole.channel_id" :validation="v$.channel_id"/>
        <form-value :type="FormTypeEmbed" title="Message" v-model="$props.reactionRole.message" class="mt-4" :validation="v$.message"/>
      </div>
    </div>
    <div v-if="onboardingStep >= 1">
      <h2 class="text-2xl font-bold mb-8 mt-8 text-center">Select Reaction Role Type</h2>
      <div class="grid  grid-cols-1 md:grid-cols-3 gap-2 mb-4">
        <button @click="$props.reactionRole.type = 'emoji'" :class="[$props.reactionRole.type === 'emoji' ? ' border-primary' : 'border-gray-300 dark:border-secondary-light', 'p-8 border rounded-lg shadow-sm h-fit disabled:bg-secondary-light']">
          <h2 class="font-semibold text-center mb-2">Emojis</h2>
          <img src="/assets/reaction_roles_emoji.png" alt="Emoji Reaction Roles" class="mt-4 mx-auto" />
        </button>
        <button @click="$props.reactionRole.type = 'buttons'" :disabled="$props.reactionRole.is_system_message == false" :class="[$props.reactionRole.type === 'buttons' ? ' border-primary' : 'border-gray-300 dark:border-secondary-light', 'p-8 border rounded-lg shadow-sm h-fit disabled:bg-secondary-light']">
          <h2 class="font-semibold text-center mb-2">Buttons</h2>
          <img src="/assets/reaction_roles_buttons.png" alt="Button Reaction Roles" class="mt-4 mx-auto" />
        </button>
        <button @click="$props.reactionRole.type = 'dropdown'" :disabled="$props.reactionRole.is_system_message == false" :class="[$props.reactionRole.type === 'dropdown' ? ' border-primary' : 'border-gray-300 dark:border-secondary-light', 'p-8 border rounded-lg shadow-sm h-fit disabled:bg-secondary-light']">
          <h2 class="font-semibold text-center mb-2">Dropdown</h2>
          <img src="/assets/reaction_roles_dropdown.png" alt="Dropdown Reaction Roles" class="mt-4 mx-auto" />
        </button>
      </div>
    </div>
    <div v-if="onboardingStep >= 2">
      <h2 class="text-2xl font-bold mb-8 mt-8 text-center">Select Roles</h2>

      <role-table-reaction-roles
        :roles="$store.getters.getAssignableGuildRoles"
        :selectedRoles="$props.reactionRole.roles"
        :type="$props.reactionRole.type"
        @removeRole="onRemoveReactionRole"
        @selectRole="onSelectReactionRole"
      ></role-table-reaction-roles>
    </div>

    <div v-if="onboardingStep == 0" class="flex justify-end">
      <button @click="validateOnboardingStepZero" class="px-4 py-2 bg-blue-600 text-white rounded w-full md:w-auto">
        <loading-icon class="inline-block mr-2" :isLight="true" v-if="showOnboardingLoading" />
        Next
      </button>
    </div>
    <div v-else-if="onboardingStep == 1" class="flex justify-end">
      <button :disabled="!($props.reactionRole.type == 'emoji' || $props.reactionRole.type == 'buttons' || $props.reactionRole.type == 'dropdown')" @click="validateOnboardingStepOne" class="px-4 py-2 bg-blue-600 text-white rounded w-full md:w-auto disabled:opacity-50 disabled:cursor-not-allowed">
        Next
      </button>
    </div>
    <div v-else class="flex justify-end gap-2">
      <button v-if="!$props.isSetup" :disabled="!$props.reactionRole.roles.length" @click="$emit('delete')" class="px-4 py-2 cta-button bg-red-500 hover:bg-red-600 text-white rounded w-full md:w-auto disabled:opacity-50 disabled:cursor-not-allowed">
        Delete
      </button>
      <button :disabled="!$props.reactionRole.roles.length" @click="$emit('save')" class="px-4 py-2 cta-button bg-green-500 hover:bg-green-600 text-white rounded w-full md:w-auto disabled:opacity-50 disabled:cursor-not-allowed">
        {{ $props.isSetup ? 'Create' : 'Save' }}
      </button>
    </div>
  </div>
</template>

<script>
import { computed, ref } from "vue";

import { Popover, PopoverButton, PopoverPanel } from "@headlessui/vue";
import { ChevronDownIcon } from "@heroicons/vue/solid";
import useVuelidate from "@vuelidate/core";
import { helpers, requiredIf } from "@vuelidate/validators";

import dashboardAPI from "@/api/dashboard";
import endpoints from "@/api/endpoints";
import EmbedBuilder from "@/components/dashboard/EmbedBuilder.vue";
import FormValue from "@/components/dashboard/FormValue.vue";
import {
  FormTypeChannelListCategories,
  FormTypeEmbed,
  FormTypeText
} from "@/components/dashboard/FormValueEnum";
import RoleTableReactionRoles from "@/components/dashboard/RoleTableReactionRoles.vue";
import DiscordEmojiPicker from "@/components/DiscordEmojiPicker.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";
import store from "@/store";
import {
  getErrorToast,
  getValidationToast,
  navigateToErrors,
  isValidJson
} from "@/utilities";

export default {
  components: {
    ChevronDownIcon,
    DiscordEmojiPicker,
    EmbedBuilder,
    FormValue,
    LoadingIcon,
    Popover,
    PopoverButton,
    PopoverPanel,
    RoleTableReactionRoles,
  },
  props: {
    isSetup: {
      type: Boolean,
      required: true,
    },
    reactionRole: {
      type: Object,
      required: true,
    },
  },
  emits: ["save"],

  setup(props) {
    let onboardingStep = ref(props.reactionRole.is_system_message == undefined ? 0 : 2);
    let showOnboardingLoading = ref(false);
    // Onboarding steps:
    // 0 - Select message
    // 1 - Select type and roles

    const validation_rules = computed(() => {
      const validation_rules = {
        message_link: {
          required: helpers.withMessage("Message Link is required.", requiredIf(() => props.reactionRole.is_system_message === false)),
          discordMessageLink: helpers.withMessage("Please enter a valid Discord message link.", (value) => !value || /^https:\/\/discord\.com\/channels\/\d+\/\d+\/\d+$/.test(value)),
          currentServer: helpers.withMessage("The message link must be for the current server.", (value) => {
            if (!value) return true;
            
            const messageLinkParts = value.split("/");
            const guildID = messageLinkParts[messageLinkParts.length - 3];
            return guildID == store.getters.getSelectedGuildID;
          }),
        },
        channel_id: {
          required: helpers.withMessage("Channel is required.", requiredIf(
            () => props.reactionRole.is_system_message === true
          )),
        },
        message: {
          required: helpers.withMessage("The message is required", requiredIf(
            () => props.reactionRole.is_system_message === true
          )),
          isValidJson: helpers.withMessage("The message is not valid JSON", (value) => {
            return !value || isValidJson(value);
          }),
        },
      };

      return validation_rules;
    })

    const v$ = useVuelidate(validation_rules, props.reactionRole, { $rewardEarly: true });

    return {
      FormTypeChannelListCategories,
      FormTypeEmbed,
      FormTypeText,

      onboardingStep,
      showOnboardingLoading,

      v$,
    };
  },

  methods: {
    async validateOnboardingStepZero() {
      const validForm = await this.v$.$validate();

      if (!validForm) {
        this.$store.dispatch("createToast", getValidationToast());
        navigateToErrors();

        return;
      }

      if (this.$props.reactionRole.message_link) {
        this.showOnboardingLoading = true;
        const messageLinkParts = this.$props.reactionRole.message_link.split("/");

        this.$props.reactionRole.channel_id = messageLinkParts[messageLinkParts.length - 2];
        this.$props.reactionRole.message_id = messageLinkParts[messageLinkParts.length - 1];

        dashboardAPI.doPost(
          endpoints.EndpointCheckMessage(this.$store.getters.getSelectedGuildID, this.$props.reactionRole.channel_id, this.$props.reactionRole.message_id),
          {},
          null,
          () => {
            this.showOnboardingLoading = false;
            this.onboardingStep = 1;
          },
          (error) => {
            this.$store.dispatch("createToast", getErrorToast(error));
          }
        )
      } else {
        this.onboardingStep = 1;
      }
    },

    validateOnboardingStepOne() {
      this.onboardingStep = 2;
    },

    onSelectReactionRole(reaction_role) {
      this.$props.reactionRole.roles.push(reaction_role);
    },

    onRemoveReactionRole(roleID) {
      this.$props.reactionRole.roles = this.$props.reactionRole.roles.filter((role) => role.role_id != roleID);
    },
  }
};
</script>