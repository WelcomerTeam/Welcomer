<template>
  <div>
    <transition enter-active-class="transition duration-200 ease-out" enter-from-class="translate-y-1 opacity-0"
      enter-to-class="translate-y-0 opacity-100" leave-active-class="transition duration-150 ease-in"
      leave-from-class="translate-y-0 opacity-100" leave-to-class="translate-y-1 opacity-0">
      <div v-if="unsavedChanges" class="fixed inset-x-0 bottom-0 z-10">
        <div class="bg-primary">
          <div class="max-w-xl mx-auto py-3 px-3 sm:px-6">
            <div class="flex items-center justify-between space-x-4 flex-nowrap">
              <p class="ml-3 font-medium text-white">
                <span class="inline"> You have unsaved changes! </span>
              </p>
              <a href="#"
                class="flex items-center justify-center px-4 py-2 whitespace-nowrap border border-transparent rounded-md shadow-sm text-sm font-medium text-primary bg-white hover:bg-gray-200"
                @click="this.$emit('save')">
                <loading-icon class="mr-3" v-if="$props.isChangeInProgress" />
                Save Changes
              </a>
            </div>
          </div>
        </div>
      </div>
    </transition>
    <confirm-leave :unsavedChanges="unsavedChanges"></confirm-leave>
  </div>
</template>

<script>
import ConfirmLeave from "@/components/ConfirmLeave.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";

export default {
  components: {
    ConfirmLeave,
    LoadingIcon,
  },
  props: {
    unsavedChanges: {
      type: Boolean,
    },
    isChangeInProgress: {
      type: Boolean,
    },
  },
};
</script>
