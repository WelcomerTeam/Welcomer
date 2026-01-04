<template>
  <Listbox as="div" :model-value="modelValue" @update:modelValue="$emit('update:modelValue', $event)" v-slot="{ open }">
    <div class="relative">
      <ListboxButton
        class="bg-secondary-dark relative w-full py-2 pl-3 pr-10 text-left border border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm">
        {{ displayText }}
      </ListboxButton>
      <ListboxOptions v-if="open" class="absolute z-20 w-full mt-1 overflow-hidden text-base bg-secondary-dark rounded-md shadow-sm max-h-60 ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm" @keydown.escape="onClose">
        <input
          ref="searchInput"
          v-model="searchQuery"
          type="text"
          placeholder="Search fonts..."
          class="w-full px-3 py-2 bg-secondary border-b border-secondary-light text-white focus:outline-none"
          @click.stop />
        <div 
          class="overflow-y-auto max-h-48"
          @scroll="onScroll"
          ref="scrollContainer">
          <div :style="{ height: topPlaceholderHeight + 'px' }"></div>
          <ListboxOption
            v-for="item in visibleOptions"
            :key="item.key"
            :value="item.key"
            v-slot="{ active }">
            <li
              :class="[
                active ? 'text-white bg-primary' : 'text-gray-50',
                'cursor-default select-none relative py-2 pl-3 pr-9 h-9']">
              {{ item.label }}
            </li>
          </ListboxOption>
          <div :style="{ height: bottomPlaceholderHeight + 'px' }"></div>
          <li v-if="filteredOptions.length === 0" class="px-3 py-2 text-gray-400 text-center">
            No options found
          </li>
        </div>
      </ListboxOptions>
      <template v-if="trackOpenState(open)"></template>
    </div>
  </Listbox>
</template>

<script>
import { Listbox, ListboxButton, ListboxOption, ListboxOptions } from '@headlessui/vue';

const ITEM_HEIGHT = 36;
const VISIBLE_ITEMS = 6;

export default {
  components: {
    Listbox,
    ListboxButton,
    ListboxOption,
    ListboxOptions,
  },
  props: {
    modelValue: {
      type: String,
      required: true,
    },
    options: {
      type: Object,
      required: true,
    },
    displayFormatter: {
      type: Function,
      default: (key, option) => option.name || option,
    },
  },
  emits: ['update:modelValue'],
  data() {
    return {
      searchQuery: '',
      scrollTop: 0,
      previousOpen: false,
    };
  },
  computed: {
    optionsArray() {
      return Object.entries(this.options).map(([key, value]) => ({
        key,
        label: this.displayFormatter(key, value),
      }));
    },
    filteredOptions() {
      if (!this.searchQuery.trim()) return this.optionsArray;

      const query = this.searchQuery.toLowerCase();
      return this.optionsArray.filter(item =>
        item.label.toLowerCase().includes(query)
      );
    },
    startIndex() {
      return Math.max(0, Math.floor(this.scrollTop / ITEM_HEIGHT));
    },
    endIndex() {
      return Math.min(
        this.filteredOptions.length,
        this.startIndex + VISIBLE_ITEMS + 1
      );
    },
    visibleOptions() {
      return this.filteredOptions.slice(this.startIndex, this.endIndex);
    },
    topPlaceholderHeight() {
      return this.startIndex * ITEM_HEIGHT;
    },
    bottomPlaceholderHeight() {
      return Math.max(0, (this.filteredOptions.length - this.endIndex) * ITEM_HEIGHT);
    },
    displayText() {
      return this.displayFormatter(this.modelValue, this.options[this.modelValue]) || 'Select option';
    },
    selectedIndex() {
      return this.filteredOptions.findIndex(item => item.key === this.modelValue);
    },
  },
  watch: {
    searchQuery() {
      this.scrollTop = 0;
      this.$refs.scrollContainer.scrollTop = 0;
      this.$forceUpdate();
    },
  },
  methods: {
    onScroll(event) {
      this.scrollTop = event.target.scrollTop;
    },
    scrollToSelected() {
      this.$nextTick(() => {
        const scrollContainer = this.$refs.scrollContainer;
        if (scrollContainer && this.selectedIndex >= 0) {
          const targetScroll = Math.max(0, this.selectedIndex * ITEM_HEIGHT - (VISIBLE_ITEMS / 2) * ITEM_HEIGHT);
          scrollContainer.scrollTop = targetScroll;
          this.scrollTop = targetScroll;
        } else if (scrollContainer) {
          scrollContainer.scrollTop = 0;
          this.scrollTop = 0;
        }
      });
    },
    onOpen() {
      this.searchQuery = '';
      this.scrollTop = 0;
      this.$nextTick(() => {
        this.$refs.searchInput?.focus();
        this.scrollToSelected();
      });
    },
    onClose() {
      this.searchQuery = '';
      this.scrollTop = 0;
    },
    trackOpenState(open) {
      if (open !== this.previousOpen) {
        this.previousOpen = open;
        if (open) {
          this.onOpen();
        } else {
          this.onClose();
        }
      }
      return false;
    },
  },
};
</script>
