<template>
  <div class="relative min-h-screen">
    <Header />

    <main>
      <div id="overview" class="relative bg-secondary">
        <div class="px-6 py-12 bg-secondary w-full max-w-7xl mx-auto">
          <p class="text-3xl font-bold text-left text-white flex tracking-tight">
            Welcome Image Backgrounds
          </p>
        </div>
        <div>
          <BackgroundCarousel />
        </div>
      </div>

      <div id="backgrounds">
        <div class="bg-white">
          <div class="hero-preview">
            <div class="px-4 pt-8 mx-auto max-w-7xl sm:px-6">
              <div class="sm:flex sm:flex-col sm:align-center">
                <!-- <div class="mb-4 grid grid-cols-4 gap-4">
                  <input type="text"
                    class="col-span-4 sm:col-span-3 border-gray-300 dark:border-secondary-light bg-white dark:bg-secondary-dark rounded-md sm:text-sm"
                    placeholder="" v-model="query" @input="onQueryChange()" />

                  <Listbox as="div" class="col-span-4 sm:col-span-1">
                    <div class="relative">
                      <ListboxButton
                        class="bg-white dark:bg-secondary-dark relative w-full py-2 pl-3 pr-10 text-left border border-gray-300 dark:border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm">
                        <span class="block truncate">Groups</span>
                        <span class="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
                          <SelectorIcon class="w-5 h-5 text-gray-400" aria-hidden="true" />
                        </span>
                      </ListboxButton>

                      <transition leave-active-class="transition duration-100 ease-in" leave-from-class="opacity-100"
                        leave-to-class="opacity-0">
                        <ListboxOptions
                          class="absolute z-20 w-full mt-1 overflow-auto text-base bg-white dark:bg-secondary-dark rounded-md shadow-sm max-h-60 ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                          <ListboxOption as="template" v-for="group in groups" :key="group.id" :value="group.id">
                            <li
                              class="text-gray-900 dark:text-gray-50 cursor-default select-none relative py-2 pl-3 pr-9 hover:font-semibold font-normal block truncate hover:bg-primary"
                              @click="scrollTo(group.id)">
                              {{ group.name }}
                            </li>
                          </ListboxOption>
                        </ListboxOptions>
                      </transition>
                    </div>
                  </Listbox>
                </div> -->

                <div class="space-y-12">
                  <div v-for="category in backgrounds" :key="category" :id="category.id">
                    <div class="text-xs font-bold uppercase my-4 text-gray-900">
                      {{ category.name }}
                    </div>
                    <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-2">
                      <button as="template" v-for="image in category.images" :key="image">
                        <img :title="image.name" v-lazy="{
                          src: `/assets/backgrounds/${image.name}.webp`,
                        }" :class="[
                          $props.modelValue == image.name
                            ? 'border-primary ring-primary ring-4'
                            : '',
                          'hover:brightness-75 rounded-md focus:outline-none focus:ring-4 focus:ring-primary focus:border-primary aspect-[10/3] w-full',
                        ]" />
                      </button>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <BackgroundPreview />
    </main>
    <Footer />
  </div>
</template>

<script>
import Header from "@/components/Header.vue";
import Footer from "@/components/Footer.vue";

import backgrounds from "@/backgrounds.json";

import {
  Listbox,
  ListboxButton,
  ListboxLabel,
  ListboxOption,
  ListboxOptions,
} from "@headlessui/vue";

import { SelectorIcon } from "@heroicons/vue/solid";

import BackgroundCarousel from "@/components/BackgroundCarousel.vue";
import BackgroundPreview from "@/components/BackgroundPreview.vue";

export default {
  components: {
    Listbox,
    ListboxButton,
    ListboxLabel,
    ListboxOption,
    ListboxOptions,
    SelectorIcon,

    BackgroundCarousel,
    BackgroundPreview,

    Header,
    Footer,
  },
  setup() {
    return {
      backgrounds,
    };
  },
  methods: {
    scrollTo(id) {
      let elem = document.getElementById(id);
      if (elem) {
        elem.scrollIntoView({
          behavior: "smooth",
          block: "start",
          inline: "start",
        });
      }
    },
  },
};
</script>
