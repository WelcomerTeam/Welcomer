<template>
  <div>
    <component :is="$props.isTextarea ? 'textarea' : 'input'" ref="textarea" :value="value" @input="onInput"
      @keydown.down="highlightNext" @keydown.up="highlightPrev" @keydown.enter="selectSuggestion"
      @keydown.tab="selectSuggestion" @selectionchange="selectionChange" :class="$props.class" :placeholder="$props.placeholder"
      :disabled="$props.disabled" :rows="$props.isTextarea ? 4 : undefined" :id="$props.id" :type="$props.type"></component>
    <div v-if="showSuggestions"
      class="absolute rounded-lg shadow-lg mt-1 z-50 border dark:border-secondary-light dark:bg-secondary bg-gray-50 border-gray-300 overflow-hidden"
      :style="{ top: `${position.top}px`, left: `${position.left}px`, width: `${position.width}px` }">
      <p class="p-2 text-xs font-bold group uppercase text-secondary-light dark:text-gray-400">
        {{ suggestionTitle }}
      </p>
      <ul class="px-2 pb-2 rounded-lg">
        <li ref="suggestionsList" v-for="(suggestion, index) in suggestions" :key="index" :class="{
          'dark:bg-secondary-light bg-gray-200': index === highlightedIndex,
          'relative px-2 py-1 cursor-pointer rounded-md': true
        }" @click="selectSuggestionAsIndex(index)" @mouseover="highlightedIndex = index">
          <div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none"
            v-if="suggestion.iconType === iconTypeChannel">
            <font-awesome-icon icon="hashtag" class="w-5 h-5 text-gray-400" aria-hidden="true" />
          </div>
          <div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none"
            v-else-if="suggestion.iconType === iconTypeRole">
            <font-awesome-icon icon="user-tag" class="w-5 h-5 text-gray-400" aria-hidden="true" />
          </div>
          <div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none"
            v-else-if="suggestion.iconType === iconTypeIcon">
            <img :src="suggestion.icon" class="w-full max-w-5 max-h-5" />
          </div>
          <span :class="{
            'block truncate': true,
            'pl-10': !(suggestion.iconType == undefined),
          }">{{ suggestion.name }}</span>
        </li>
      </ul>
    </div>
  </div>
</template>

<script>
import store from "@/store/index";

import { ref } from 'vue';

var iconTypeChannel = 1;
var iconTypeRole = 1;
var iconTypeIcon = 2;

const formattingTags = [
  "{{User.ID}}",
  "{{User.Name}}",
  "{{User.Username}}",
  "{{User.Discriminator}}",
  "{{User.GlobalName}}",
  "{{User.Mention}}",
  "{{User.CreatedAt}}",
  "{{User.JoinedAt}}",
  "{{User.Avatar}}",
  "{{User.Bot}}",
  "{{User.Pending}}",
  "{{Guild.ID}}",
  "{{Guild.Name}}",
  "{{Guild.Icon}}",
  "{{Guild.Splash}}",
  "{{Guild.Members}}",
  "{{Ordinal(Guild.Members)}}",
  "{{Guild.MembersJoined}}",
  "{{Ordinal(Guild.MembersJoined)}}",
  "{{Guild.Banner}}",
  "{{Invite.Code}}",
  "{{Invite.Uses}}",
  "{{Invite.Inviter}}",
  "{{Invite.ChannelID}}",
  "{{Invite.CreatedAt}}",
  "{{Invite.ExpiresAt}}",
  "{{Invite.MaxAge}}",
  "{{Invite.MaxUses}}",
  "{{Invite.Temporary}}",
];

export default {
  props: {
    value: {
      required: true,
    },
    isTextarea: {
      type: Boolean,
      default: false,
    },
    class: {
      type: String,
    },
    disabled: {
      type: Boolean,
    },
    placeholder: {
      type: String,
    },
    id: {
      type: String,
    },
    type: {
      type: String,
      default: "text",
    },
  },

  emits: ["update:modelValue", "input"],

  setup() {
    var highlightedIndex = ref(-1);
    var showSuggestions = ref(false);
    var suggestions = ref([]);
    var suggestionTitle = ref("");
    var position = ref({ top: 0, left: 0, width: 200 });

    var resizeObserver = null;

    return {
      store,

      highlightedIndex,
      showSuggestions,
      suggestions,
      suggestionTitle,
      position,
      resizeObserver,

      iconTypeChannel,
      iconTypeRole,
      iconTypeIcon,
    };
  },

  mounted() {
    this.resizeObserver = new ResizeObserver(() => {
      if (this.$refs.textarea) {
        this.updatePosition(this.$refs.textarea);
      }
    });

    if (this.$refs.textarea) {
      this.resizeObserver.observe(this.$refs.textarea);
    }
  },

  beforeUnmount() {
    if (this.resizeObserver) {
      this.resizeObserver.disconnect();
    }
  },

  methods: {
    getEligibleText() {
      var content = this.$refs.textarea.value;

      const caretPos = this.$refs.textarea.selectionStart;
      const slice = content.slice(0, caretPos);
      const match = slice.match(/(#|:|\{|\{\{)([\w\.\(\)]*)$/);

      var start;
      var end;

      if (match) {
        start = caretPos - match[0].length;
        end = caretPos;
      }

      return { match, start, end }
    },

    selectionChange(event) {
      var { match } = this.getEligibleText(false);

      if (match) {
        var { suggestions, suggestionTitle } = this.getSuggestions(match[0]);
        this.suggestions = suggestions;
        this.suggestionTitle = suggestionTitle
        this.showSuggestions = this.suggestions.length > 0;
      } else {
        this.suggestions = [];
        this.showSuggestions = false;
      }

      this.updatePosition(event.target);
      this.highlightedIndex = 0;
    },

    onInput(event) {
      this.selectionChange(event);

      this.$emit("update:modelValue", event.target.value);
      this.$emit("input", event.target.value);
    },

    highlightNext(e) {
      if (this.suggestions.length == 0) {
        return;
      }

      if (this.highlightedIndex == this.suggestions.length - 1) {
        this.highlightedIndex = 0;
      } else {
        this.highlightedIndex++;
      }

      this.$refs.suggestionsList[this.highlightedIndex]?.scrollIntoView({
        behavior: "smooth",
        block: "nearest",
      });

      e.preventDefault();
    },

    highlightPrev(e) {
      if (this.suggestions.length == 0) {
        return;
      }

      if (this.highlightedIndex == 0) {
        this.highlightedIndex = this.suggestions.length - 1;
      } else {
        this.highlightedIndex--;
      }

      this.$refs.suggestionsList[this.highlightedIndex]?.scrollIntoView({
        behavior: "smooth",
        block: "nearest",
      });

      e.preventDefault();
    },

    selectSuggestion(e) {
      if (this.suggestions.length > 0) {
        this.selectSuggestionAsIndex(this.highlightedIndex);
        e.preventDefault();
      }
    },

    selectSuggestionAsIndex(index) {
      if (index >= 0 && index < this.suggestions.length) {
        const { start, end } = this.getEligibleText();
        var content = this.$props.value.slice(0, start) + this.suggestions[index].value + " " + this.$props.value.slice(end);

        this.$emit("update:modelValue", content);
        this.$emit("input", content);

        this.highlightedIndex = -1;
        this.suggestions = [];
        this.showSuggestions = false;
      }
    },

    getSuggestions(input) {
      var options = function () {
        if (input.startsWith("#")) {
          const query = input.slice(1);

          return {
            suggestionTitle: query ? `Channels matching #${query}` : `Channels`,
            suggestions: store.getters.getGuildChannels.filter(channel => channel.name.toLowerCase().includes(query.toLowerCase())).map(channel => {
              return {
                name: `${channel.name}`,
                value: `<#${channel.id}>`,
                iconType: iconTypeChannel,
              };
            }),
          };
        }

        if (input.startsWith(":")) {
          const query = input.slice(1);

          // Do not show suggestions if the length is less than 2
          if (query.length <= 1) {
            return
          }

          return {
            suggestionTitle: query ? `Emojis matching :${query}` : `Emojis`,
            suggestions: store.getters.getGuildEmojis.filter(emoji => emoji.name.toLowerCase().includes(query.toLowerCase())).map(emoji => {
              return {
                name: `:${emoji.name}:`,
                value: `<${emoji.animated ? "a" : ""}:${emoji.name}:${emoji.id}>`,
                iconType: iconTypeIcon,
                icon: `https://cdn.discordapp.com/emojis/${emoji.id}.${emoji.animated ? "gif" : "webp"}?size=128`,
              };
            }),
          };
        }

        if (input.startsWith("{")) {
          var query = input.slice(1);
          if (query.startsWith("{")) {
            query = query.slice(1);
          }

          return {
            suggestionTitle: query ? `Tags matching {{${query}}}` : `Tags`,
            suggestions: formattingTags.filter(tag => tag.toLowerCase().includes(query.toLowerCase())).map(tag => {
              return {
                name: tag,
                value: tag,
              };
            }),
          };
        }

        return
      }();

      if (options?.suggestions) {
        options.suggestions = options.suggestions.slice(0, Math.floor(window.innerHeight * 0.4 / 32));
      }

      return options || {
        suggestionTitle: "",
        suggestions: [],
      };
    },

    updatePosition(textarea) {
      this.position = {
        top: textarea.offsetTop + textarea.offsetHeight,
        left: textarea.offsetLeft,
        width: textarea.clientWidth,
      };
    },
  }
};
</script>