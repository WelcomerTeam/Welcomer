<template>
  <div class="fixed top-6 right-6 space-y-3 z-50">
    <transition-group enter-active-class="transition duration-200 ease-out" enter-from-class="translate-x-1 opacity-0"
      enter-to-class="translate-x-0 opacity-100" leave-active-class="transition duration-150 ease-in"
      leave-from-class="translate-x-0 opacity-100" leave-to-class="translate-x-1 opacity-0">
      <div v-for="toast in $store.getters.getToasts" v-bind:key="toast.id">
        <div id="toast-default"
          class="flex items-center w-full max-w-xs p-4 text-gray-500 bg-white dark:bg-secondary-dark dark:text-gray-50 rounded-lg shadow-sm"
          role="alert">
          <font-awesome-icon :icon="toast.icon || 'info'" :class="[
            toast.class || 'bg-blue-100 text-blue-500',
            'p-2 rounded-lg w-4 h-4',
          ]" />
          <div class="mx-5 text-sm font-normal flex-1 capitalize">
            {{ toast.title }}
          </div>
          <button type="button"
            class="-mx-1.5 bg-white dark:bg-secondary-light text-gray-500 dark:text-gray-50 dark:hover:text-gray-300 hover:text-gray-900 rounded-lg focus:ring-2 focus:ring-gray-300 dark:focus:ring-secondary-light p-1.5 hover:bg-gray-100 dark:hover:bg-primary inline-flex h-8 w-8"
            data-dismiss-target="#toast-default" @click="hideToast(toast.id)" aria-label="Close">
            <XIcon />
            <span class="sr-only">Close</span>
          </button>
        </div>
      </div>
    </transition-group>
  </div>
</template>

<script>
import { XIcon } from "@heroicons/vue/outline";
export default {
  components: {
    XIcon,
  },

  methods: {
    hideToast(toastID) {
      this.$store.dispatch("removeToast", toastID);
    },
  },
};
</script>
