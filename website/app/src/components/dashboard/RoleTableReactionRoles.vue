<template>
  <div>
    <table class="min-w-full border-spacing-2">
      <tbody class="divide-y divide-gray-200 dark:divide-secondary-light">
        <tr v-for="(reaction_role, index) in this.$props.selectedRoles" :key="index">
          <td class="w-8 pr-3 text-center">
            <img :src="getEmojiURL(reaction_role.emoji)" alt="Selected Emoji" class="inline-block mr-2 align-middle w-6 max-h-6" v-if="reaction_role.emoji"/>
          </td>
          <td class="pr-3 text-sm dark:text-gray-50 py-3.5">
            <font-awesome-icon icon="circle" class="text-gray-400 inline w-4 h-4 mr-1 border-primary" :style="$store.getters.getGuildRoleById(reaction_role.role_id)?.color ? {
              color: `${getHexColor(
                $store.getters.getGuildRoleById(reaction_role.role_id)?.color
              )}`,
            } : {}" />
            <div v-if="
              ($store.getters.getGuildRoleById(reaction_role.role_id)?.is_elevated && $store.getters.getGuildRoleById(reaction_role.role_id)?.is_assignable) ||
              !$store.getters.getGuildRoleById(reaction_role.role_id)?.is_assignable" class="inline" title="This role is elevated or not assignable">
              <font-awesome-icon icon="exclamation-triangle" class="text-red-500 w-4 h-4 mr-1" />
            </div>
            {{ $store.getters.getGuildRoleById(reaction_role.role_id)?.name }}
          </td>
          <td class="pr-3 text-sm dark:text-gray-50 py-3.5" v-if="$props.type != 'emoji'">
            <span>{{ reaction_role.name }}</span>
            <p class="text-xs text-gray-500 dark:text-gray-400" v-if="$props.type == 'dropdown' && reaction_role.description">{{ reaction_role.description }}</p>
          </td>
          <td class="whitespace-nowrap py-4 px-2 text-right text-sm dark:text-gray-50 space-x-2">
            <a @click="this.onRemoveRole(reaction_role.role_id)" class="text-primary hover:text-primary-dark cursor-pointer">
              <font-awesome-icon icon="close" />
            </a>
          </td>
        </tr>
      </tbody>
    </table>
    <div class="space-y-2 mb-8">
      <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
        <div>
          <Popover as="div" v-slot="{ open, close }" class="relative">
            <PopoverButton class="relative w-full py-2 pl-3 pr-8 text-left bg-white dark:bg-secondary-dark border border-gray-300 dark:border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm">
              <div class="h-6 flex items-center justify-center w-full" v-if="reactionRole.emoji">
                <img :src="getEmojiURL(reactionRole.emoji)" alt="Selected Emoji" class="inline-block mr-2 align-middle w-6 max-h-6" />
              </div>
              <span v-else class="text-gray-400">Select Emoji</span>
              <span class="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
                <ChevronDownIcon :class="[
                  open ? 'transform rotate-180' : '',
                  'w-5 h-5 text-gray-400 transition-all duration-100',
                ]" aria-hidden="true" />
              </span>
            </PopoverButton>
            <transition :show="open" leave-active-class="transition duration-100 ease-in" leave-from-class="opacity-100" leave-to-class="opacity-0">
              <PopoverPanel class="block w-full overflow-auto text-base bg-white dark:bg-secondary rounded-md shadow-sm sm:text-sm border border-gray-300 dark:border-secondary-light">
                <DiscordEmojiPicker customGroupLabel="Server Emojis" :customEmojis="$store.getters.getCurrentSelectedGuild.emojis" @select="(emoji) => { reactionRole.emoji = reactionRole.emoji == emoji.value ? '' : emoji.value; close(); }" />
              </PopoverPanel>
            </transition>
          </Popover>
          <div v-if="v$.emoji?.$invalid" class="errors">
            <span v-bind:key="index" v-for="(message, index) in v$.emoji.$errors">{{ message.$message }}&nbsp;</span>
          </div>
        </div>
        <form-value :type="FormTypeRoleList" :show-role-popup="true" v-model="reactionRole.role_id" :inline-form-value="true" :hide-border="true" :hide-slot="true" title="Role" :options="selectableRoles" :validation="v$.role_id" />
      </div>
      <div class="grid grid-cols-1 md:grid-cols-2 gap-2">
        <form-value :type="FormTypeText" :inline-form-value="true" :hide-border="true" :hide-slot="true" v-if="$props.type != 'emoji'" v-model="reactionRole.name" placeholder="Name" :validation="v$.name" />
        <form-value :type="FormTypeText" :inline-form-value="true" :hide-border="true" :hide-slot="true" v-if="$props.type == 'dropdown'" v-model="reactionRole.description" placeholder="Description" :validation="v$.description" />
      </div>
      <button :disabled="reactionRole.role_id == undefined" @click="addRole" class="cta-button bg-primary hover:bg-primary-dark w-full md:w-auto disabled:opacity-50 disabled:cursor-not-allowed">
        Add Role
      </button>
    </div>
  </div>
</template>

<script>
import {
  Listbox,
  ListboxButton,
  ListboxOption,
  ListboxOptions,
  Popover,
  PopoverButton,
  PopoverPanel,
} from "@headlessui/vue";

import DiscordEmojiPicker from "@/components/DiscordEmojiPicker.vue";

import { ref, computed } from "vue";

import useVuelidate from "@vuelidate/core";
import { helpers, required, requiredIf } from "@vuelidate/validators";

import { CheckIcon, ChevronDownIcon, SelectorIcon } from "@heroicons/vue/solid";
import { XIcon } from "@heroicons/vue/outline";
import EmbedBuilder from "./EmbedBuilder.vue";
import FormValue from "./FormValue.vue";
import { FormTypeBlank, FormTypeRoleList, FormTypeText } from "./FormValueEnum";
import UnsavedChanges from "./UnsavedChanges.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";

import { getHexColor, getRolePermissionListAsString, getValidationToast, navigateToErrors } from "@/utilities";

export default {
  props: {
    roles: {
      type: Object,
      required: true,
    },
    selectedRoles: {
      type: Object,
      required: true,
    },
    type: {
      type: String,
      required: true,
    }
  },

  emits: ["removeRole", "selectRole"],

  components: {
    Listbox,
    ListboxButton,
    ListboxOption,
    ListboxOptions,
    Popover,
    PopoverButton,
    PopoverPanel,

    DiscordEmojiPicker,
    FormValue,
    EmbedBuilder,
    UnsavedChanges,
    XIcon,
    LoadingIcon,
    CheckIcon,
    ChevronDownIcon,
    SelectorIcon,
  },

  mounted() {
    this.parseSelectedRoles();
  },

  setup(props) {
    let selectableRoles = ref([]);

    let reactionRole = ref({
      emoji: "",
      name: "",
      description: "",
      role_id: undefined,
    }); 

    const validation_rules = computed(() => {
      const validation_rules = {
        emoji: {
          required: helpers.withMessage("No emoji has been selected", requiredIf(props.type == 'emoji')),
          not_used: helpers.withMessage("This emoji has already been used", (value) => {
            return props.type != 'emoji' || !props.selectedRoles.some((role) => role.emoji === value);
          }),
        },
        role_id: {
          required: helpers.withMessage("No role has been selected", required),
        },
        name: {
          required: helpers.withMessage("No name has been provided", requiredIf(props.type != 'emoji')),
          length: helpers.withMessage("The name cannot be longer than 50 characters", (value) => {
            return value.length <= 50;
          }),
        },
        description: {
          length: helpers.withMessage("The description cannot be longer than 100 characters", (value) => {
            return props.type != 'dropdown' || (value.length <= 100);
          }),
        },
      };

      return validation_rules;
    }); 

    const v$ = useVuelidate(validation_rules, reactionRole, { $rewardEarly: true });

    return {
      FormTypeBlank,
      FormTypeRoleList,
      FormTypeText,

      selectableRoles,
      reactionRole,
      v$,
    };
  },

  watch: {
    $props: {
      handler: function () {
        this.parseSelectedRoles();
      },
      deep: true,
      immediate: true,
    }
  },

  methods: {
    getHexColor,

    async addRole() {
      const validForm = await this.v$.$validate();

      if (!validForm) {
        this.$store.dispatch("createToast", getValidationToast());
        navigateToErrors();

        return;
      }

      this.$emit("selectRole", {
        emoji: this.reactionRole.emoji,
        role_id: this.reactionRole.role_id,
        name: this.reactionRole.name,
        description: this.reactionRole.description,
      });

      this.reactionRole.emoji = "";
      this.reactionRole.name = "";
      this.reactionRole.description = "";
      this.reactionRole.role_id = undefined;
      this.v$.$reset();
    },

    getEmojiURL(emoji) {
      let isNumbers = /^[0-9]+$/.test(emoji);
      if (isNumbers) {
        return `https://cdn.discordapp.com/emojis/${emoji}.png?size=64`;
      } else {
        let twemojiCode = Array.from(emoji)
          .map((part) => part.codePointAt(0))
          .map((code) => code.toString(16))
          .join("-");
        return `https://twemoji.maxcdn.com/v/latest/72x72/${twemojiCode}.png`;
      }
    },

    onRemoveRole(roleID) {
      this.$emit("removeRole", roleID);
    },

    parseSelectedRoles() {
      let selectableRoles = [];

      this.$props.roles.forEach((reaction_role) => {
        var assigned_roles = this.$props.selectedRoles.find((element) => {
          return element.role_id == reaction_role.id;
        });

        if (assigned_roles === undefined) {
          selectableRoles.push(reaction_role.role_id);
        };
      });

      this.selectableRoles = selectableRoles;
    },
  },
};
</script>
