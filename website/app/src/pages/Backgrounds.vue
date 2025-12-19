<template>
  <div class="relative min-h-screen">
    <Header />

    <main>
      <div id="overview" class="relative bg-secondary">
        <div class="px-6 py-12 bg-secondary w-full max-w-7xl mx-auto">
          <h1 class="text-3xl font-bold text-left text-white tracking-tight">
            Welcome Image Backgrounds
          </h1>
        </div>
        <div>
          <BackgroundCarousel />
        </div>
      </div>

      <div id="backgrounds">
        <div class="bg-white text-neutral-900">
          <div class="hero-preview px-4 mx-auto max-w-7xl sm:px-6 sm:flex sm:flex-col sm:align-center">
            <div class="space-y-12">
              <div v-for="category in backgrounds" :key="category" :id="category.id">
                <div class="text-xs font-bold uppercase my-4 text-gray-900">
                  {{ category.name }}
                </div>
                <div class="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-4 gap-2">
                  <button as="template" v-for="image in category.images" :key="image">
                    <img :title="image.name" :alt="'Background image ' + image.name" v-lazy="{
                      src: `/assets/backgrounds/${image.name}.webp`,
                    }" :class="[
                        $props.modelValue == image.name
                          ? 'border-primary ring-primary ring-4'
                          : '',
                        'hover:brightness-75 rounded-md focus:outline-none focus:ring-4 focus:ring-primary focus:border-primary aspect-[10/3] w-full',
                      ]" />
                      <span class="sr-only">Select background {{ image.name }}</span>
                  </button>
                </div>
              </div>
            </div>

            <div class="border-primary bg-primary text-white border p-6 lg:p-12 rounded-lg shadow-sm h-fit mt-16">
              <h3 class="text-2xl font-bold sm:text-3xl">
                Looking for more?
              </h3>
              <p class="mt-4 text-sm leading-6">If you do not like these images, you can always use your own! You can unlock custom backgrounds forever for your server with a one-time purchase. Unlock animated backgrounds and more Welcomer features on any server you choose. Select from monthly, biannual or yearly plans to suit your needs.</p>

              <a href="/premium" target="_blank" type="button" class="bg-white hover:bg-gray-200 flex items-center justify-center px-5 py-3 mt-8 text-base font-medium text-primary border border-transparent rounded-md cursor-pointer w-full">
                Learn More
              </a>
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
