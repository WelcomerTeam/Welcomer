<template>
  <div class="flex min-h-screen overflow-hidden bg-gray-100">
    <TransitionRoot as="template" :show="$props.sidebarOpen">
      <Dialog as="div" static class="fixed inset-0 z-40 flex lg:hidden" @close="this.$emit('closeSidebar')"
        :open="$props.sidebarOpen">
        <TransitionChild as="template" enter="transition-opacity ease-linear duration-300" enter-from="opacity-0"
          enter-to="opacity-100" leave="transition-opacity ease-linear duration-300" leave-from="opacity-100"
          leave-to="opacity-0">
          <DialogOverlay class="fixed inset-0 bg-gray-600 bg-opacity-25" />
        </TransitionChild>
        <TransitionChild as="template" enter="transition ease-in-out duration-300 transform"
          enter-from="-translate-x-full" enter-to="translate-x-0" leave="transition ease-in-out duration-300 transform"
          leave-from="translate-x-0" leave-to="-translate-x-full">
          <div
            class="relative flex flex-col flex-1 w-full max-w-xs bg-gray-100 border-r dark:bg-secondary-dark dark:border-secondary-light shadow-inner">
            <TransitionChild as="template" enter="ease-in-out duration-300" enter-from="opacity-0" enter-to="opacity-100"
              leave="ease-in-out duration-300" leave-from="opacity-100" leave-to="opacity-0">
              <div class="absolute top-0 right-0 pt-2 -mr-12">
                <button
                  class="flex items-center justify-center w-10 h-10 ml-1 rounded-full focus:outline-none focus:ring-2 focus:ring-inset focus:ring-white bg-secondary-dark"
                  @click="this.$emit('closeSidebar')">
                  <span class="sr-only">Close sidebar</span>
                  <XIcon class="w-6 h-6 text-white" aria-hidden="true" />
                </button>
              </div>
            </TransitionChild>
            <div class="pt-5 pb-4 overflow-hidden overflow-y-auto custom-scroll">
              <DashboardSidebar @onTabClick="this.$emit('closeSidebar')" />
            </div>
          </div>
        </TransitionChild>
        <div class="flex-shrink-0 w-14" aria-hidden="true">
          <!-- Dummy element to force sidebar to shrink to fit close icon -->
        </div>
      </Dialog>
    </TransitionRoot>

    <div class="hidden lg:flex lg:flex-shrink-0">
      <div class="flex flex-col w-64">
        <div
          class="flex flex-col flex-grow pt-5 pb-4 overflow-y-auto custom-scroll bg-gray-100 border-r dark:border-secondary-light dark:bg-secondary-dark shadow-inner">
          <DashboardSidebar />
        </div>
      </div>
    </div>

    <div class="flex-1 focus:outline-none bg-white dark:bg-secondary">
      <HoistHeading />

      <main class="relative z-0 flex-1 min-h-full pb-9">
        <div class="font-medium pb-40">
          <div v-if="$store.getters.isLoadingGuild">
            <div class="dashboard-container flex justify-center">
              <LoadingIcon />
            </div>
          </div>
          <div v-else-if="!$store.getters.guildHasWelcomer">
            <div class="dashboard-container">Welcomer isn't here!</div>
          </div>
          <router-view v-else />
        </div>
      </main>
    </div>

    <Toast />
  </div>
</template>

<script>
import Header from "@/components/dashboard/Header.vue";
import DashboardSidebar from "@/components/dashboard/Sidebar.vue";
import Toast from "@/components/dashboard/Toast.vue";
import HoistHeading from "@/components/hoist/HoistHeading.vue";

import {
  Dialog,
  DialogOverlay,
  Menu,
  MenuItem,
  TransitionChild,
  TransitionRoot,
} from "@headlessui/vue";

import { XIcon } from "@heroicons/vue/outline";

import LoadingIcon from "@/components/LoadingIcon.vue";

export default {
  props: {
    sidebarOpen: {
      type: Boolean,
    },
  },
  components: {
    Dialog,
    DialogOverlay,
    Menu,
    MenuItem,
    TransitionChild,
    TransitionRoot,
    Toast,
    XIcon,

    Header,
    HoistHeading,
    DashboardSidebar,
    LoadingIcon,
  },
};
</script>
