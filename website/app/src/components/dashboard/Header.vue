<template>
  <header>
    <Popover class="relative w-full shadow bg-secondary-dark">
      <div class="min-h-full px-6 mx-auto sm:px-6">
        <div class="flex items-center justify-between md:justify-start md:space-x-10 py-6">
          <slot />
          <router-link to="/">
            <div class="flex justify-start">
              <img class="w-auto h-8" src="/assets/logo.svg" alt="Welcomer Logo" />
              <span class="my-auto ml-2 text-xl font-bold text-white">Welcomer</span>
            </div>
          </router-link>
          <div class="-my-2 -mr-2 md:hidden">
            <PopoverButton
              class="focus:outline-none hover:text-gray-400 inline-flex items-center justify-center p-2 rounded-md text-gray-300">
              <span class="sr-only">Open menu</span>
              <MenuIcon class="w-6 h-6" aria-hidden="true" />
            </PopoverButton>
          </div>
          <PopoverGroup as="nav" class="hidden space-x-6 md:flex">
            <div class="inline-flex my-auto space-x-4">
              <Popover class="relative" v-slot="{ open }" v-if="Toggle_ShowFeaturesOnDashboard">
                <PopoverButton :class="[
                  open ? 'text-gray-300' : 'text-white',
                  'group focus:outline-none hover:text-gray-300 inline-flex items-center rounded-md text-base',
                ]">
                  <span>Features</span>
                  <ChevronDownIcon :class="[
                    open ? 'text-gray-300' : 'text-white',
                    'group-hover:text-gray-300 h-5 ml-1 w-5',
                  ]" aria-hidden="true" />
                </PopoverButton>

                <transition enter-active-class="transition duration-200 ease-out"
                  enter-from-class="translate-y-1 opacity-0" enter-to-class="translate-y-0 opacity-100"
                  leave-active-class="transition duration-150 ease-in" leave-from-class="translate-y-0 opacity-100"
                  leave-to-class="translate-y-1 opacity-0">
                  <PopoverPanel
                    class="absolute bg-secondary-dark lg:max-w-lg max-w-md mt-3 px-2 rounded-md sm:px-0 transform w-screen z-10 left-1/2 -translate-x-1/2">
                    <div class="popover-container">
                      <div class="gap-6 grid px-5 py-6 relative rounded-lg sm:gap-8 sm:p-8 bg-secondary-dark">
                        <router-link v-for="item in NavigationFeatures" :key="item.name" :to="item.href"
                          class="group -m-3 flex hover:bg-secondary items-start p-2 rounded-lg">
                          <div class="flex-shrink-0">
                            <div class="popover-panel-icon">
                              <font-awesome-icon :icon="item.icon" class="w-6 h-6" aria-hidden="true" />
                            </div>
                          </div>
                          <div class="my-auto ml-4 leading-none">
                            <p class="text-lg font-medium text-white">
                              {{ item.name }}
                            </p>
                            <p class="text-sm text-gray-300">
                              {{ item.description }}
                            </p>
                          </div>
                        </router-link>
                        <router-link class="text-white underline hover:text-gray-300" to="/features">View all
                          features</router-link>
                      </div>
                    </div>
                  </PopoverPanel>
                </transition>
              </Popover>

              <Popover class="relative z-40" v-slot="{ open }">
                <PopoverButton :class="[
                  open ? 'text-gray-300' : 'text-white',
                  'group focus:outline-none hover:text-gray-300 inline-flex items-center rounded-md text-base',
                ]">
                  <span>Help</span>
                  <ChevronDownIcon :class="[
                    open ? 'text-gray-300' : 'text-white',
                    'group-hover:text-gray-300 h-5 ml-1 w-5',
                  ]" aria-hidden="true" />
                </PopoverButton>

                <transition enter-active-class="transition duration-200 ease-out"
                  enter-from-class="translate-y-1 opacity-0" enter-to-class="translate-y-0 opacity-100"
                  leave-active-class="transition duration-150 ease-in" leave-from-class="translate-y-0 opacity-100"
                  leave-to-class="translate-y-1 opacity-0">
                  <PopoverPanel
                    class="absolute bg-secondary-dark lg:max-w-lg max-w-md mt-3 px-2 rounded-md sm:px-0 transform w-screen z-10 left-1/2 -translate-x-1/2">
                    <div class="popover-container">
                      <div class="gap-6 grid px-5 py-6 relative rounded-lg sm:gap-8 sm:p-8 bg-secondary-dark">
                        <router-link v-for="item in NavigationResources" :key="item.name" :to="item.href"
                          class="group -m-3 flex hover:bg-secondary items-start p-2 rounded-lg">
                          <div class="flex-shrink-0">
                            <div class="popover-panel-icon">
                              <font-awesome-icon :icon="item.icon" :path="item.icon" class="w-6 h-6" aria-hidden="true" />
                            </div>
                          </div>
                          <div class="my-auto ml-4 leading-none">
                            <p class="text-lg font-medium text-white">
                              {{ item.name }}
                            </p>
                            <p class="text-sm text-gray-300">
                              {{ item.description }}
                            </p>
                          </div>
                        </router-link>
                      </div>
                    </div>
                  </PopoverPanel>
                </transition>
              </Popover>
            </div>
          </PopoverGroup>
          <UserProfile />
        </div>
      </div>

      <transition enter-active-class="duration-200 ease-out" enter-from-class="scale-95 opacity-0"
        enter-to-class="scale-100 opacity-100" leave-active-class="duration-100 ease-in"
        leave-from-class="scale-100 opacity-100" leave-to-class="scale-95 opacity-0">
        <PopoverPanel focus class="navbar-mobile-panel">
          <div class="navbar-mobile-menu">
            <div class="px-5 pt-5 pb-6">
              <div class="flex items-center justify-between">
                <div class="flex justify-start">
                  <img class="w-auto h-8" src="/assets/logo.svg" alt="Workflow" />
                  <span class="my-auto ml-2 text-xl font-bold text-white">Welcomer</span>
                </div>
                <div class="-mr-2">
                  <PopoverButton class="navbar-mobile-close">
                    <span class="sr-only">Close menu</span>
                    <XIcon class="w-6 h-6" aria-hidden="true" />
                  </PopoverButton>
                </div>
              </div>
              <div class="mt-6">
                <UserProfileCompact />
              </div>
            </div>

            <div class="px-4 py-4">
              <div class="grid grid-cols-2">
                <router-link :to="{ name: 'invite' }" v-if="!$store.getters.isLoggedIn" class="navbar-mobile-menu-item">
                  <div class="popover-panel-icon bg-primary">
                    <font-awesome-icon icon="plus" class="navbar-mobile-menu-item-icon" aria-hidden="true" />
                  </div>
                  <span class="navbar-mobile-menu-item-text">
                    Invite Welcomer
                  </span>
                </router-link>
                <router-link :to="{ name: 'dashboard.guilds' }" v-else class="navbar-mobile-menu-item">
                  <div class="popover-panel-icon bg-primary">
                    <font-awesome-icon icon="toolbox" class="navbar-mobile-menu-item-icon" aria-hidden="true" />
                  </div>
                  <span class="navbar-mobile-menu-item-text"> Dashboard </span>
                </router-link>

                <router-link to="/premium" class="navbar-mobile-menu-item">
                  <div class="popover-panel-icon">
                    <font-awesome-icon icon="heart" class="navbar-mobile-menu-item-icon" aria-hidden="true" />
                  </div>
                  <span class="navbar-mobile-menu-item-text">
                    Get Welcomer Pro
                  </span>
                </router-link>
              </div>
            </div>

            <div class="px-4 py-4" v-if="Toggle_ShowFeaturesOnDashboard">
              <span class="pl-3 font-bold uppercase text-gray-200">Features</span>
              <nav class="grid grid-cols-2">
                <router-link v-for="item in NavigationFeatures" :key="item.name" :to="item.href" class="navbar-mobile-menu-item">
                  <div class="popover-panel-icon">
                    <font-awesome-icon :icon="item.icon" class="navbar-mobile-menu-item-icon" aria-hidden="true" />
                  </div>
                  <span class="navbar-mobile-menu-item-text">
                    {{ item.name }}
                  </span>
                </router-link>
              </nav>
            </div>

            <div class="px-4 py-4">
              <span class="pl-3 font-bold uppercase text-gray-200">Help</span>
              <nav class="grid grid-cols-2">
                <router-link v-for="item in NavigationResources" :key="item.name" :to="item.href"
                  class="navbar-mobile-menu-item">
                  <div class="popover-panel-icon">
                    <font-awesome-icon :icon="item.icon" class="navbar-mobile-menu-item-icon" aria-hidden="true" />
                  </div>
                  <span class="navbar-mobile-menu-item-text">
                    {{ item.name }}
                  </span>
                </router-link>
              </nav>
            </div>
          </div>
        </PopoverPanel>
      </transition>
    </Popover>
  </header>
</template>

<script>
import {
  Popover,
  PopoverButton,
  PopoverGroup,
  PopoverPanel,
} from "@headlessui/vue";
import { MenuIcon, XIcon } from "@heroicons/vue/outline";
import { ChevronDownIcon } from "@heroicons/vue/solid";

import UserProfile from "@/components/UserProfile.vue";
import UserProfileCompact from "@/components/UserProfileCompact.vue";

import {
  NavigationFeatures,
  NavigationResources,
  Toggle_ShowFeaturesOnDashboard
} from "@/constants";

export default {
  components: {
    Popover,
    PopoverButton,
    PopoverGroup,
    PopoverPanel,
    UserProfile,
    UserProfileCompact,
    ChevronDownIcon,
    MenuIcon,
    XIcon,
  },
  setup() {
    return {
      NavigationFeatures,
      NavigationResources,
      Toggle_ShowFeaturesOnDashboard,
    };
  }
}
</script>
