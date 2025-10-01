<template>
  <Popover as="div" v-slot="{ open }" class="relative">
    <div :class="[
      $props.invalid
        ? 'ring-red-500 border-red-500'
        : 'border-gray-300 dark:border-secondary-light',
      '',
    ]">
      <PopoverButton :class="[
        $props.disabled
          ? 'bg-gray-100 dark:bg-secondary-light text-neutral-500'
          : 'bg-white dark:bg-secondary',
        'relative py-2 pl-3 pr-10 text-left border border-gray-300 dark:border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm w-full',
      ]" :disabled="$props.disabled">
        <div>
          <span>{{ friendlyString }}</span>
        </div>
        <span class="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
          <ChevronDownIcon :class="[
            open ? 'transform rotate-180' : '',
            'w-5 h-5 text-gray-400 transition-all duration-100',
          ]" aria-hidden="true" />
        </span>
      </PopoverButton>
    </div>
    <transition leave-active-class="transition duration-100 ease-in" leave-from-class="opacity-100"
      leave-to-class="opacity-0">
      <PopoverPanel
        class="block w-full overflow-auto text-base bg-white dark:bg-secondary rounded-md shadow-sm sm:text-sm rounded-t-none border-t-0">
        <div class="border-gray-300 dark:border-secondary-light rounded-md border shadow-sm p-4 space-y-1">
          <div v-if="showYears" class="flex items-center gap-2"><input v-model="years" type="number" min="0"
              class="flex-1 shadow-sm block w-4 max-w-32 border-gray-300 dark:border-secondary-light dark:bg-secondary-dark rounded-md focus:ring-primary focus:border-primary sm:text-sm"
              @input="onUpdate" />
            years</div>
          <div v-if="showDays" class="flex items-center gap-2"><input v-model="days" type="number" min="0"
              class="flex-1 shadow-sm block w-4 max-w-32 border-gray-300 dark:border-secondary-light dark:bg-secondary-dark rounded-md focus:ring-primary focus:border-primary sm:text-sm"
              @input="onUpdate" />
            days</div>
          <div v-if="showHours" class="flex items-center gap-2"><input v-model="hours" type="number" min="0"
              class="flex-1 shadow-sm block w-4 max-w-32 border-gray-300 dark:border-secondary-light dark:bg-secondary-dark rounded-md focus:ring-primary focus:border-primary sm:text-sm"
              @input="onUpdate" />
            hours</div>
          <div v-if="showMinutes" class="flex items-center gap-2"><input v-model="minutes" type="number" min="0"
              class="flex-1 shadow-sm block w-4 max-w-32 border-gray-300 dark:border-secondary-light dark:bg-secondary-dark rounded-md focus:ring-primary focus:border-primary sm:text-sm"
              @input="onUpdate" />
            minutes</div>
          <div v-if="showSeconds" class="flex items-center gap-2"><input v-model="seconds" type="number" min="0"
              class="flex-1 shadow-sm block w-4 max-w-32 border-gray-300 dark:border-secondary-light dark:bg-secondary-dark rounded-md focus:ring-primary focus:border-primary sm:text-sm"
              @input="onUpdate" />
            seconds</div>
        </div>
      </PopoverPanel>
    </transition>
  </Popover>
</template>

<script>
import {
  Popover,
  PopoverButton,
  PopoverPanel,
} from '@headlessui/vue';
import { ChevronDownIcon } from '@heroicons/vue/solid';


export default {
  components: {
    Popover,
    PopoverButton,
    PopoverPanel,
    ChevronDownIcon,
  },

  props: {
    modelValue: {
      type: Number,
      required: true
    },
    disabled: {
      type: Boolean,
    },
    invalid: {
      type: Boolean,
    },
    blankDisplay: {
      type: String,
      default: 'Immediately',
    },

    showYears: {
      type: Boolean,
      default: true,
    },
    showDays: {
      type: Boolean,
      default: true,
    },
    showDays: {
      type: Boolean,
      default: true,
    },
    showHours: {
      type: Boolean,
      default: true,
    },
    showMinutes: {
      type: Boolean,
      default: true,
    },
    showSeconds: {
      type: Boolean,
      default: true,
    },
  },

  emits: ['update:modelValue'],

  data() {
    return {
      years: 0,
      days: 0,
      hours: 0,
      minutes: 0,
      seconds: 0,

      friendlyString: ''
    };
  },

  watch: {
    modelValue: {
      immediate: true,
      handler(newValue) {
        let remaining = newValue;

        if (this.showYears) {
          this.years = Math.floor(remaining / (365 * 24 * 60 * 60));
          remaining = remaining % (365 * 24 * 60 * 60);
        }

        if (this.showDays) {
          this.days = Math.floor(remaining / (24 * 60 * 60));
          remaining %= (24 * 60 * 60);
        }

        if (this.showHours) {
          this.hours = Math.floor(remaining / (60 * 60));
          remaining %= (60 * 60);
        }

        if (this.showMinutes) {
          this.minutes = Math.floor(remaining / 60);
          remaining %= 60;
        }

        this.seconds = remaining % 60;

        const parts = [];
        if (this.years) parts.push(`${this.years} year${this.years > 1 ? 's' : ''}`);
        if (this.days) parts.push(`${this.days} day${this.days > 1 ? 's' : ''}`);
        if (this.hours) parts.push(`${this.hours} hour${this.hours > 1 ? 's' : ''}`);
        if (this.minutes) parts.push(`${this.minutes} minute${this.minutes > 1 ? 's' : ''}`);
        if (this.seconds) parts.push(`${this.seconds} second${this.seconds > 1 ? 's' : ''}`);
        if (parts.length === 0) {
          this.friendlyString = this.blankDisplay;
          return;
        }

        if (parts.length > 1) {
          this.friendlyString = parts.slice(0, -1).join(', ') + ' and ' + parts.slice(-1);
        } else {
          this.friendlyString = parts.join('');
        }
      }
    }
  },

  methods: {
    onUpdate() {
      this.updateValue(
        (this.showYears ? this.years : 0) * 365 * 24 * 60 * 60 +
        (this.showDays ? this.days : 0) * 24 * 60 * 60 +
        (this.showHours ? this.hours : 0) * 60 * 60 +
        (this.showMinutes ? this.minutes : 0) * 60 +
        (this.showSeconds ? this.seconds : 0)
      );
    },
    updateValue(value) {
      this.$emit("update:modelValue", value);
    },
  }
}
</script>