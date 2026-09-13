<template>
  <dl class="bg-white dark:bg-secondary-dark dark:border-secondary-light rounded-lg shadow-sm border border-gray-300 p-5 flex flex-col">
    <dt class="flex mb-2">
      <span class="text-xs font-bold uppercase text-gray-500 dark:text-gray-200 flex-1">{{ $props.name }}</span>
      <div class="flex-shrink-0">
        <font-awesome-icon :icon="$props.icon" :class="[$props.iconStyle, 'w-4 h-4 dark:text-gray-50']" />
      </div>
    </dt>
    <dd class="flex flex-col flex-1">
      <div class="text-4xl font-bold text-gray-900 dark:text-gray-50 flex-1">
        {{ typeof $props.amount === 'number' ? Intl.NumberFormat().format($props.amount) : $props.amount }}{{ $props.amountSuffix }}
      </div>
      <div class="text-xs mt-2" v-if="$props.previousAmount != null">
        <analytics-card-change :amount="$props.amount" :previousAmount="$props.previousAmount" :isDecreaseGood="$props.isDecreaseGood" />
      </div>
    </dd>
  </dl>
</template>

<script>
import AnalyticsCardChange from './AnalyticsCardChange.vue';

export default {
  props: {
    name: {
      type: String,
      required: true
    },
    amount: {
      type: Number,
      required: true
    },
    previousAmount: {
      type: Number,
      default: null
    },
    amountSuffix: {
      type: String,
      default: ''
    },
    icon: {
      type: String,
      default: ''
    },
    iconStyle: {
      type: String,
      default: ''
    },
    isDecreaseGood: {
      type: Boolean,
      default: false
    }
  },
  components: {
    AnalyticsCardChange,
  },
};
</script>

<style></style>
