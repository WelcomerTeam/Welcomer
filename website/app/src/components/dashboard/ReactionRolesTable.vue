<template>
  <div>
    <table class="min-w-full border-spacing-2">
      <thead>
        <tr>
          <th></th>
          <th></th>
          <th class="text-right">
            <button type="button" class="cta-button bg-primary hover:bg-primary-dark" @click="openCreatePopup()">
              Create Reaction Role
            </button>
          </th>
        </tr>
      </thead>
      <tbody class="divide-y divide-gray-200 dark:divide-secondary-light">
        <tr v-for="(reactionRole, index) in this.$props.modelValue.reaction_roles" :key="index">
          <td class="py-3">
            <Switch v-model="reactionRole.enabled" @click="onValueUpdate" :class="[
              reactionRole.enabled ? 'bg-green-500 focus:ring-green-500' : 'bg-gray-400 focus:ring-gray-400',
              'relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2',
            ]">
              <span :class="[
                reactionRole.enabled ? 'translate-x-5' : 'translate-x-0',
                'pointer-events-none relative inline-block h-5 w-5 rounded-full bg-white shadow transform ring-0 transition ease-in-out duration-200',
              ]">
                <span :class="[
                  reactionRole.enabled
                    ? 'opacity-0 ease-out duration-100'
                    : 'opacity-100 ease-in duration-200',
                  'absolute inset-0 h-full w-full flex items-center justify-center transition-opacity',
                ]" aria-hidden="true">
                  <svg class="w-3 h-3 text-gray-400" fill="none" viewBox="0 0 12 12">
                    <path d="M4 8l2-2m0 0l2-2M6 6L4 4m2 2l2 2" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                          stroke-linejoin="round" />
                  </svg>
                </span>
                <span :class="[
                  reactionRole.enabled
                    ? 'opacity-100 ease-in duration-200'
                    : 'opacity-0 ease-out duration-100',
                  'absolute inset-0 h-full w-full flex items-center justify-center transition-opacity',
                ]" aria-hidden="true">
                  <svg class="w-3 h-3 text-green-500" fill="currentColor" viewBox="0 0 12 12">
                    <path
                      d="M3.707 5.293a1 1 0 00-1.414 1.414l1.414-1.414zM5 8l-.707.707a1 1 0 001.414 0L5 8zm4.707-3.293a1 1 0 00-1.414-1.414l1.414 1.414zm-7.414 2l2 2 1.414-1.414-2-2-1.414 1.414zm3.414 2l4-4-1.414-1.414-4 4 1.414 1.414z" />
                  </svg>
                </span>
              </span>
            </Switch>
          </td>
          <td class="py-3">
            <discord-embed class="flex-1" :embeds="reactionRole.is_system_message ? parseDict(reactionRole.embed)?.embeds : []" :content="reactionRole.is_system_message ? parseDict(reactionRole.embed)?.content : ''" :buttons="getReactionRoleButtons(reactionRole.type, reactionRole.roles)" :isLight="true" :showAuthor="false" />
          </td>
          <td class="py-3 text-right">
            <button class="relative py-2 px-2 border border-gray-300 dark:border-secondary-light rounded-md shadow-sm focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm" @click="reactionRole.showPopup = true">
              <font-awesome-icon icon="pen-to-square" class="w-5 h-5 text-gray-400" aria-hidden="true" />
            </button>
          </td>
          <popup :open="reactionRole.showPopup" @close="reactionRole.showPopup = false; onValueUpdate();">
            <reaction-roles-table-item :isSetup="false" :reactionRole="reactionRole" @save="saveReactionRole(reactionRole, false)" @delete="removeReactionRole(index)" />
          </popup>
        </tr>
      </tbody>
    </table>
    <popup :open="showCreatePopup" @close="showCreatePopup = false;">
      <reaction-roles-table-item :isSetup="true" :reactionRole="createPopupData" @save="saveReactionRole(createPopupData, true)" />
    </popup>
  </div>
</template>

<script>
import { ref } from "vue";

import { Switch } from "@headlessui/vue";

import DiscordEmbed from "@/components/DiscordEmbed.vue";

import Popup from "../Popup.vue";


import ReactionRolesTableItem from "./ReactionRolesTableItem.vue";

export default {
  components: {
    DiscordEmbed,
    Popup,
    Switch,
    ReactionRolesTableItem,
  },
  props: {
    modelValue: {
      type: Object,
      required: true,
    },
  },

  setup() {
    let showCreatePopup = ref(false);
    let createPopupData = ref({});

    return {
      showCreatePopup,
      createPopupData,
    };
  },

  emits: ["update:modelValue"],

  methods: {
    parseDict(data) {
      try {
        return JSON.parse(data);
      } catch {
        return {};
      }
    },

    openCreatePopup() {
      this.resetCreatePopup();
      this.showCreatePopup = true;
    },

    resetCreatePopup() {
      this.createPopupData = {
        enabled: true,
        is_system_message: undefined,
        embed: "{\"embeds\":[{\"description\":\"React below to get roles!\"}]}",
        roles: [],
      }
    },

    onValueUpdate() {
      this.$emit("update:modelValue", this.modelValue);
    },

    saveReactionRole(reactionRole, isSetup) {
      if (isSetup) {
        this.modelValue.reaction_roles.push(reactionRole);
        this.showCreatePopup = false;
        this.resetCreatePopup();
      } else {
        reactionRole.showPopup = false;
      }
      
      this.onValueUpdate();
    },
    
    removeReactionRole(index) {
      this.modelValue.reaction_roles.splice(index, 1);
      this.onValueUpdate();
    },

    getReactionRoleButtons(reactionRoleType, roles) {
      if (reactionRoleType == "dropdown") {
        return [{
          style: 6,
          options: roles.map((roleOption) => ({
            label: roleOption.name,
            description: roleOption.description,
            emoji: roleOption.emoji,
          })),
        }];
      }

      let buttons = [];

      roles.forEach((roleOption) => {
        buttons.push({
          type: roleOption.style,
          label: roleOption.name,
          emoji: roleOption.emoji,
          style: reactionRoleType == "emoji" ? 5 : roleOption.style,
        });
      });

      return buttons;
    }
  }
}
</script>