<template>
  <dl class="bg-white dark:bg-secondary-dark dark:border-secondary-light rounded-lg shadow-sm border border-gray-300 p-5">
    <dt class="flex mb-4">
      <span class="text-xs font-bold uppercase text-gray-500 dark:text-gray-200 flex-1">{{ $props.name }}</span>
      <div class="flex-shrink-0">
        <font-awesome-icon :icon="$props.icon" :class="[$props.iconStyle, 'w-4 h-4 dark:text-gray-50']" />
      </div>
    </dt>
    <dd>
      <div class="text-4xl font-bold text-gray-900 dark:text-gray-50">
        {{ Intl.NumberFormat().format($props.amount) }}
      </div>
      <div class="text-xs mt-2" v-if="$props.previousAmount != null">
        <span v-if="$props.amount > $props.previousAmount" class="text-trendUp font-semibold">
          <font-awesome-icon icon="arrow-trend-up" class="mr-1" />
          +{{ (($props.amount / $props.previousAmount * 100) - 100).toFixed(2) }}%
        </span>
        <span v-else-if="$props.amount < $props.previousAmount" class="text-trendDown font-semibold">
          <font-awesome-icon icon="arrow-trend-down" class="mr-1" />
          -{{ (($props.previousAmount / $props.amount * 100) - 100).toFixed(2) }}%
        </span>
        <span v-else>
          No change
        </span>
        vs last period
      </div>
    </dd>
  </dl>
</template>

<script>
export default {
  props: ["name", "amount", "previousAmount", "icon", "iconStyle"],
};
</script>

<style></style>
