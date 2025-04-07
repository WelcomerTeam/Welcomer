<template>
  <TransitionRoot as="template" :show="open">
    <Dialog class="relative z-10" :open="true">
      <TransitionChild as="template" enter="ease-out duration-300" enter-from="opacity-0" enter-to="opacity-100" leave="ease-in duration-200" leave-from="opacity-100" leave-to="opacity-0">
        <div class="fixed inset-0 bg-black/50 transition-opacity" />
      </TransitionChild>

      <div class="fixed inset-0 z-10 w-screen overflow-y-auto">
        <div class="flex min-h-full justify-center p-4 text-center items-center">
          <TransitionChild as="template" enter="ease-out duration-300" enter-from="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95" enter-to="opacity-100 translate-y-0 sm:scale-100" leave="ease-in duration-200" leave-from="opacity-100 translate-y-0 sm:scale-100" leave-to="opacity-0 translate-y-4 sm:translate-y-0 sm:scale-95">
            <DialogPanel :class="['bg-white text-secondary dark:bg-secondary dark:text-gray-50 relative transform overflow-y-auto rounded-md text-left shadow-xl transition-all sm:my-8 sm:w-full sm:max-w-[1024px]']">
              <div class="flex gap-4 align-top p-6 dark:bg-secondary-dark border-b dark:border-secondary-light bg-gray-100 border-gray-200 shadow-inner">
                  <div class="flex-1">
                    <slot name="title"></slot>
                  </div>
                  <button v-if="showCloseButton" @click="$emit('close')">
                      <font-awesome-icon icon="times" />
                  </button>
              </div>

              <div class="p-6">
                <slot></slot>
              </div>
            </DialogPanel>
          </TransitionChild>
        </div>
      </div>
    </Dialog>
  </TransitionRoot>
</template>

<script>
import {
  Dialog,
  DialogBackdrop,
  DialogPanel,
  TransitionChild,
  TransitionRoot,
} from '@headlessui/vue'

export default {
  components: {
    Dialog,
    DialogBackdrop,
    DialogPanel,
    TransitionChild,
    TransitionRoot,
  },

  props: {
    open: {
      type: Boolean,
      required: true,
    },
    showCloseButton: {
      type: Boolean,
      default: true,
    },
  },
}

</script>