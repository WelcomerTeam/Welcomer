<template>
  <dl class="bg-white dark:bg-secondary-dark dark:border-secondary-light rounded-lg shadow-sm border border-gray-300 p-5">
    <dt class="flex mb-4">
      <span class="text-xs font-bold uppercase text-gray-500 dark:text-gray-200 flex-1">{{ $props.name }}</span>
      <div class="flex-shrink-0">
        <font-awesome-icon :icon="$props.icon" :class="[$props.iconStyle, 'w-4 h-4 dark:text-gray-50']" />
      </div>
    </dt>
    <dd>
      <slot />
    </dd>
  </dl>
</template>

<script>
export default {
  props: ["name", "icon", "iconStyle"],
};
</script>

<style></style>
