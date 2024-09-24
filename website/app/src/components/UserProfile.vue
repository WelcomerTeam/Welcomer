<template>
  <div class="space-x-4 hidden items-center justify-end lg:w-0 md:flex md:flex-1">
    <div v-if="$store.getters.isLoadingUser" class="px-10">
      <LoadingIcon class="text-white" />
    </div>
    <div v-else-if="!$store.getters.isLoggedIn" class="space-x-4 flex">
      <router-link :to="{ name: 'invite' }">
        <button type="button" class="cta-button bg-primary hover:bg-primary-dark">
          Invite Welcomer
        </button>
      </router-link>
      <a href="/login" class="hover:text-gray-300 text-base text-white whitespace-nowrap my-auto">Log in</a>
    </div>
    <div v-else class="space-x-4 flex">
      <router-link :to="{ name: 'dashboard.guilds' }">
        <button type="button" class="cta-button bg-primary hover:bg-primary-dark">
          Dashboard
        </button>
      </router-link>
      <PopoverGroup as="nav" class="hidden space-x-6 md:flex">
        <div class="inline-flex my-auto space-x-4">
          <Popover class="relative z-10" v-slot="{ open }">
            <PopoverButton :class="[
              open ? 'text-gray-300' : 'text-white',
              'group focus:outline-none hover:text-gray-300 inline-flex items-center rounded-md text-base',
            ]">
              <span v-if="$store.getters.getCurrentUser.discriminator == '0'">{{ $store.getters.getCurrentUser.global_name
              }}</span>
              <span v-else>{{ $store.getters.getCurrentUser.username }}#{{
                $store.getters.getCurrentUser.discriminator }}</span>
              <ChevronDownIcon :class="[
                open ? 'text-gray-300' : 'text-white',
                'group-hover:text-gray-300 h-5 ml-1 w-5',
              ]" aria-hidden="true" />
            </PopoverButton>

            <transition enter-active-class="transition duration-200 ease-out" enter-from-class="translate-y-1 opacity-0"
              enter-to-class="translate-y-0 opacity-100" leave-active-class="transition duration-150 ease-in"
              leave-from-class="translate-y-0 opacity-100" leave-to-class="translate-y-1 opacity-0">
              <PopoverPanel
                class="absolute bg-secondary-dark max-w-md mt-3 px-2 rounded-md sm:px-0 transform w-screen z-10 left-full -translate-x-full">
                <div class="popover-container">
                  <!-- <router-link to="/profile" -->
                  <div
                    class="gap-6 px-5 py-6 relative rounded-lg sm:gap-8 sm:p-6 group bg-primary text-white rounded-b-none grid grid-cols-4">
                    <!-- class="gap-6 px-5 py-6 relative rounded-lg sm:gap-8 sm:p-6 group bg-primary hover:bg-primary-dark text-white rounded-b-none grid grid-cols-4"> -->
                    <img class="object-cover col-span-1 aspect-square w-16 h-16" :src="`https://cdn.discordapp.com/avatars/${$store.getters.getCurrentUser.id
                      }/${$store.getters.getCurrentUser.avatar}.${$store.getters.getCurrentUser.avatar.startsWith('a_')
                        ? 'gif'
                        : 'webp'
                      }?size=128`" />
                    <div class="col-span-3 flex items-center">
                      <div>
                        <h2 class="font-bold text-xl">
                          <span v-if="$store.getters.getCurrentUser.discriminator == '0'">{{
                            $store.getters.getCurrentUser.global_name }}</span>
                          <span v-else>{{ $store.getters.getCurrentUser.username }}#{{
                            $store.getters.getCurrentUser.discriminator }}</span>
                        </h2>
                        <div class="space-x-2 space-y-2">
                          <font-awesome-icon :title="badge.name" v-for="badge in $store.getters.getCurrentUser
                            .badges" v-bind:key="badge.name" :icon="badge.icon"
                            :class="['p-2 bg-white rounded-md', badge.colour]" />
                        </div>
                      </div>
                    </div>
                  </div>
                  <div class="gap-6 grid px-5 py-6 relative rounded-lg sm:gap-6 sm:p-8 bg-secondary-dark">
                    <router-link to="/premium" class="group -m-3 flex hover:bg-secondary items-start p-2 rounded-lg">
                      <div class="flex-shrink-0">
                        <div class="popover-panel-icon">
                          <font-awesome-icon icon="heart" class="w-6 h-6" aria-hidden="true" />
                        </div>
                      </div>
                      <div class="my-auto ml-4 leading-none">
                        <p class="text-lg font-medium text-white">
                          Get Welcomer Pro
                        </p>
                        <p class="text-sm text-gray-300">
                          Get new features or just help contribute to the bot
                        </p>
                      </div>
                    </router-link>
                    <div class="flex justify-between items-center">
                      <a href="/logout" class="text-white hover:text-gray-300 underline">Logout</a>
                      <ThemeToggle />
                    </div>
                  </div>
                </div>
              </PopoverPanel>
            </transition>
          </Popover>
        </div>
      </PopoverGroup>
    </div>
  </div>
</template>

<script>

import {
  Popover,
  PopoverButton,
  PopoverGroup,
  PopoverPanel,
} from "@headlessui/vue";
import { ChevronDownIcon } from "@heroicons/vue/solid";
import LoadingIcon from "./LoadingIcon.vue";
import ThemeToggle from "./ThemeToggle.vue";

export default {
  components: {
    Popover,
    PopoverButton,
    PopoverGroup,
    PopoverPanel,
    ChevronDownIcon,
    LoadingIcon,
    ThemeToggle,
  }
};
</script>
