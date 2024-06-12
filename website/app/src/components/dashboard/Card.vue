<template>
  <div
    class="overflow-hidden bg-white dark:bg-secondary-dark dark:border-secondary-light rounded-lg shadow-sm border border-gray-300">
    <div class="p-5">
      <div class="flex items-center">
        <div class="flex-shrink-0">
          <font-awesome-icon :icon="$props.icon" class="w-6 h-6 dark:text-gray-50" />
        </div>
        <div class="flex-1 w-0 ml-5">
          <dl>
            <dt class="text-sm font-medium text-gray-500 dark:text-gray-200 truncate">
              {{ $props.name }}
            </dt>
            <dd>
              <div class="text-lg font-medium text-gray-900 dark:text-gray-50">
                {{ Intl.NumberFormat().format($props.amount) }}
              </div>
            </dd>
          </dl>
        </div>
      </div>
    </div>
    <div class="px-5 py-3 bg-gray-50" v-if="$props.showViewAll">
      <div class="text-sm">
        <a :href="$props.href" class="font-medium text-secondary hover:text-secondary-light">
          View all
        </a>
      </div>
    </div>
  </div>
</template>

<script>
export default {
  props: ["name", "amount", "showViewAll", "href", "icon"],
};
</script>

<style></style>
