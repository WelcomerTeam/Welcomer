<template>
  <div class="dashboard-container">
    <div class="dashboard-title-container">
      <div class="dashboard-title">Dashboard example</div>
    </div>
    <div class="dashboard-content">
      <form-value title="Nullable" v-model="nullable" :type="FormTypeToggle" />

      <form-value title="Disabled" v-model="disabled" :type="FormTypeToggle" />

      <form-value title="Invalid" v-model="invalid" :type="FormTypeToggle" />

      <div class="dashboard-heading">
        <a>Form Types</a>
      </div>

      <form-value title="Form Type Toggle" v-model="config.enabled" :type="FormTypeToggle" :nullable="nullable"
        :disabled="disabled" :invalid="invalid" />

      <form-value title="Form Type Channel List" v-model="config.channel" :type="FormTypeChannelList" :nullable="nullable"
        :disabled="disabled" :invalid="invalid" />
      <form-value title="Form Type Channel List Categories" v-model="config.channelTwo"
        :type="FormTypeChannelListCategories" :nullable="nullable" :disabled="disabled" :invalid="invalid" />
      <form-value title="Form Type Role List" v-model="config.role" :type="FormTypeRoleList" :nullable="nullable"
        :disabled="disabled" :invalid="invalid" />
      <form-value title="Form Type Member List" v-model="config.member" :type="FormTypeMemberList" :nullable="nullable"
        :disabled="disabled" :invalid="invalid" />
      <form-value title="Form Type Emoji List" v-model="config.emoji" :type="FormTypeEmojiList" :nullable="nullable"
        :disabled="disabled" :invalid="invalid" />
      <form-value title="Form Type Colour" v-model="config.colour" :type="FormTypeColour" :nullable="nullable"
        :disabled="disabled" :invalid="invalid" />
      <form-value title="Form Type Text" v-model="config.message" :type="FormTypeText" :nullable="nullable"
        :disabled="disabled" :invalid="invalid" />
      <form-value title="Form Type Number" v-model="config.count" :type="FormTypeNumber" :nullable="nullable"
        :disabled="disabled" :invalid="invalid" />
      <form-value title="Form Type TextArea" v-model="config.rules" :type="FormTypeTextArea" :nullable="nullable"
        :disabled="disabled" :invalid="invalid" />

      <form-value title="Form Type Dropdown" v-model="config.select" :type="FormTypeDropdown" :values="['a', 'b', 'c']"
        :nullable="nullable" :disabled="disabled" :invalid="invalid" />

      <form-value title="Form Type Embed" v-model="config.rich_message" :type="FormTypeEmbed" :nullable="nullable"
        :disabled="disabled" :invalid="invalid" />
    </div>

    <pre class="mt-12">{{ config }}</pre>
  </div>
</template>

<script>
import { ref } from "vue";
import FormValue from "@/components/dashboard/FormValue.vue";
import store from "@/store/index";

import {
  FormTypeToggle,
  FormTypeChannelList,
  FormTypeChannelListCategories,
  FormTypeRoleList,
  FormTypeMemberList,
  FormTypeEmojiList,
  FormTypeColour,
  FormTypeText,
  FormTypeNumber,
  FormTypeTextArea,
  FormTypeDropdown,
  FormTypeEmbed,
} from "@/components/dashboard/FormValueEnum";

var config = ref({
  enabled: true,
  channel: "485801351352418304",
  channelTwo: null,
  role: null,
  member: "865142611434143755",
  emoji: null,
  colour: 0,
  message: "",
  count: 0,
  rules: "",
  select: "a",
  rich_message: "{}",
});

export default {
  components: {
    FormValue,
  },
  setup() {
    let nullable = ref(false);
    let disabled = ref(false);
    let invalid = ref(false);

    return {
      nullable,
      disabled,
      invalid,
      config,

      FormTypeToggle,
      FormTypeChannelList,
      FormTypeChannelListCategories,
      FormTypeRoleList,
      FormTypeMemberList,
      FormTypeEmojiList,
      FormTypeColour,
      FormTypeText,
      FormTypeNumber,
      FormTypeTextArea,
      FormTypeDropdown,
      FormTypeEmbed,

      store,
    };
  },
};
</script>
