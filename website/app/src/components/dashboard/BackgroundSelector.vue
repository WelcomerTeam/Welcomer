<template>
  <!-- <Popover
    as="div"
    v-slot="{ open }"
    class="relative"
    :disabled="$props.disabled"
  > -->
  <Popover as="div" class="relative" :disabled="$props.disabled">
    <!-- <div
      :class="[
        $props.invalid ? 'ring-red-500 border-red-500' : '',
        open ? 'rounded-b-none' : '',
        'border border-gray-300 dark:border-secondary-light p-4 rounded-md flex shadow-sm',
      ]"
    >
      <discord-embed
        class="flex-1"
        :embeds="displayEmbed.embeds"
        :content="displayEmbed.content"
        :isLight="true"
        :isBot="true"
      />

      <div class="flex items-end">
        <div class="relative">
          <PopoverButton
            :class="[
              $props.disabled
                ? 'bg-gray-100 dark:bg-secondary-light text-neutral-500'
                : 'bg-white dark:bg-secondary',
              'relative py-2 pl-3 pr-10 text-left border border-gray-300 dark:border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm',
            ]"
            :disabled="$props.disabled"
          >
            <div class="">
              <font-awesome-icon
                icon="pen-to-square"
                class="w-5 h-5 text-gray-400"
                aria-hidden="true"
              />
            </div>
            <span
              class="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none"
            >
              <ChevronDownIcon
                :class="[
                  open ? 'transform rotate-180' : '',
                  'w-5 h-5 text-gray-400',
                ]"
                aria-hidden="true"
              />
            </span>
          </PopoverButton>
        </div>
      </div>
    </div> -->
    <!-- <transition
      leave-active-class="transition duration-100 ease-in"
      leave-from-class="opacity-100"
      leave-to-class="opacity-0"
    >
      <PopoverPanel
        class="block w-full overflow-auto text-base bg-white dark:bg-secondary rounded-md shadow-sm sm:text-sm rounded-t-none border-t-0"
      > -->
    <div v-if="$props.isLoading" class="flex py-5 w-full justify-center">
      <LoadingIcon />
    </div>
    <div v-else>
      <div :class="[
        'block w-full overflow-auto text-base rounded-md sm:text-sm bg-white border-gray-300 dark:border-secondary-light dark:bg-secondary border',
        true ? '' : 'rounded-t-none border-t-0',
      ]">
        <div class="border-b border-gray-300 dark:border-secondary-light">
          <nav class="flex display-flex justify-evenly" aria-label="Tabs">
            <a v-for="tab in tabs" :key="tab.name" @click="this.page = tab.enabled ? tab.value : this.page" :class="[
              tab.enabled ? '' : ' bg-gray-100',
              tab.value == this.page
                ? 'border-primary text-primary'
                : 'border-transparent text-gray-500 dark:text-gray-50 hover:text-gray-700 dark:hover:text-primary-light',
              'whitespace-nowrap flex py-4 px-1 border-b-2 font-medium text-sm cursor-pointer w-full justify-center',
            ]" :aria-current="tab.value == this.page ? 'page' : undefined">
              <div v-if="tab.icon" class="mr-2">
                <font-awesome-icon :icon="tab.icon" />
              </div>
              {{ tab.name }}
            </a>
          </nav>
        </div>

        <div class="overflow-auto p-4">
          <div class="space-y-12 max-h-72" v-if="this.page == 1">
            <div v-for="category in backgrounds" :key="category" :id="category.id">
              <div class="text-xs font-bold uppercase my-4 text-gray-500 dark:text-gray-100">
                {{ category.name }}
              </div>
              <div class="grid grid-cols-2 gap-2">
                <button as="template" v-for="image in category.images" :key="image" @click="updateValue(image.name)">
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
          <div v-if="this.page == 2" class="space-y-4">
            <div
              class="lg:max-w-lg flex justify-center px-6 pt-5 pb-6 border-2 border-gray-300 dark:border-secondary-light border-dashed rounded-md relative mx-auto mb-4"
              v-if="$store.getters.guildHasWelcomerPro ||
                $store.getters.guildHasCustomBackgrounds
              ">
              <input id="file-upload" name="file-upload" type="file" accept="image/*"
                class="absolute top-0 left-0 w-full h-full opacity-0" @change="onFileUpdate" />
              <div class="space-y-1 text-center" v-if="$props.files.length == 0">
                <div class="flex text-sm text-gray-600 dark:text-gray-200">
                  <label for="file-upload"
                    class="relative cursor-pointer rounded-md font-medium text-primary hover:text-primary focus-within:outline-none focus-within:ring-2 focus-within:ring-offset-2 focus-within:ring-primary">
                    <span>Upload a file</span>
                  </label>
                  <p class="pl-1">or drag and drop</p>
                </div>
                <p class="text-xs text-gray-500 dark:text-gray-100">
                  a
                  <span v-if="$store.getters.guildHasWelcomerPro"><span class="text-primary">GIF</span>, PNG or
                    JPG</span>
                  <span v-else>PNG or JPG</span>
                  up to 20MB
                </p>
              </div>
              <div class="space-y-1 text-center" v-else>
                <div class="absolute top-2 right-2">
                  <button @click="removeFiles">
                    <font-awesome-icon icon="xmark" />
                  </button>
                </div>
                <div class="flex text-sm text-gray-600 dark:text-gray-200">
                  <p>{{ $props.files[0].name }}</p>
                </div>
                <p :class="[
                  $props.files[0].size > 20000000 ? 'text-red-500' : '',
                ]">
                  {{ formatBytes($props.files[0].size) }}MB
                </p>
              </div>
            </div>
            <div v-else class="border-primary border-2 p-4 grid grid-cols-6 gap-4">
              <div class="col-span-6 items-center grid">
                <span class="font-bold leading-6">Looking for more?</span>
                With Welcomer Pro, you can unlock custom backgrounds on your server. You can upload PNG, JPG and even
                animated GIFs!
              </div>
              <a href="/premium" target="_blank" class="col-span-6 items-center grid">
                <button type="button" class="cta-button bg-primary hover:bg-primary-dark w-full">
                  Get Welcomer Pro now
                </button>
              </a>
            </div>
            <div>
              <button as="template" v-for="image in $props.customImages" :key="image"
                @click="updateValue(customPrefix + image)">
                <img v-lazy="customRoot(image)" :class="[
                  $props.modelValue == customPrefix + image
                    ? 'border-primary ring-primary ring-4'
                    : '',
                  'hover:brightness-75 rounded-md focus:outline-none focus:ring-4 focus:ring-primary focus:border-primary',
                ]" />
              </button>
            </div>
          </div>
          <!-- <div v-if="this.page == 3">Unsplash</div> -->
          <div v-if="this.page == 4">
            <div class="sm:flex sm:gap-4 sm:border-gray-300 mb-6 sm:mb-4 align-middle">
              <label class="block font-medium text-gray-700 dark:text-gray-50">
                Use profile colour for backgrounds
              </label>
              <div class="mt-1 sm:mt-0 sm:col-span-2 text-left sm:text-right">
                <Switch :modelValue="$props.modelValue ==
                  solidColourPrefix + solidColourProfileBased
                  " @update:modelValue="
                    updateValue(
                      $event
                        ? solidColourPrefix + solidColourProfileBased
                        : 'default'
                    )
                    " :class="[
                      $props.modelValue ==
                        solidColourPrefix + solidColourProfileBased
                        ? 'bg-green-500 focus:ring-green-500'
                        : 'bg-gray-400 focus:ring-gray-400',
                      'relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2',
                    ]">
                  <span :class="[
                    $props.modelValue ==
                      solidColourPrefix + solidColourProfileBased
                      ? 'translate-x-5'
                      : 'translate-x-0',
                    'pointer-events-none relative inline-block h-5 w-5 rounded-full bg-white shadow transform ring-0 transition ease-in-out duration-200',
                  ]">
                    <span :class="[
                      $props.modelValue ==
                        solidColourPrefix + solidColourProfileBased
                        ? 'opacity-0 ease-out duration-100'
                        : 'opacity-100 ease-in duration-200',
                      'absolute inset-0 h-full w-full flex items-center justify-center transition-opacity',
                    ]" aria-hidden="true">
                      <svg class="w-3 h-3 text-gray-400" fill="none" viewBox="0 0 12 12">
                        <path d="M4 8l2-2m0 0l2-2M6 6L4 4m2 2l2 2" stroke="currentColor" stroke-width="2"
                          stroke-linecap="round" stroke-linejoin="round" />
                      </svg>
                    </span>
                    <span :class="[
                      $props.modelValue ==
                        solidColourPrefix + solidColourProfileBased
                        ? 'opacity-100 ease-in duration-200'
                        : 'opacity-0 ease-out duration-100',
                      'absolute inset-0 h-full w-full flex items-center justify-center transition-opacity',
                    ]" aria-hidden="true">
                      <svg class="w-3 h-3 text-green-500" fill="currentColor" viewBox="0 0 12 12">
                        <path
                          d="M3.707 5.293a1 1 0 00-1.414 1.414l1.414-1.414zM5 8l-.707.707a1 1 0 001.414 0L5 8zm4.707-3.293a1 1 0 00-1.414-1.414l1.414 1.414zm-7.414 2l2 2 1.414-1.414-2-2-1.414 1.414zm3.414 2l4-4-1.414-1.414-4 4 1.414 1.414z" />
                      </svg>
                    </span>
                  </span>
                </Switch>
              </div>
            </div>
            <Listbox as="div" :modelValue="modelValue" @update:modelValue="updateValue($event)" :disabled="$props.modelValue == solidColourPrefix + solidColourProfileBased
              ">
              <div class="mt-1">
                <ListboxButton :class="[
                  $props.modelValue ==
                    solidColourPrefix + solidColourProfileBased
                    ? 'bg-gray-100 dark:bg-secondary-light text-neutral-500'
                    : 'bg-white dark:bg-secondary-dark',
                  'relative w-full py-2 pl-3 pr-10 text-left border border-gray-300 dark:border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm',
                ]">
                  <div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                    <font-awesome-icon icon="square" class="inline w-4 h-4 mr-1 border-primary" :style="{
                      color: `${parseCSSValue(trimPrefix(modelValue))}`,
                    }" />
                  </div>
                  <span class="block pl-10 truncate">{{ parseCSSValue(trimPrefix(modelValue)) }}
                  </span>
                  <span class="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
                    <SelectorIcon class="w-5 h-5 text-gray-400" aria-hidden="true" />
                  </span>
                </ListboxButton>

                <transition leave-active-class="transition duration-100 ease-in" leave-from-class="opacity-100"
                  leave-to-class="opacity-0">
                  <ListboxOptions class="absolute z-10 mt-1">
                    <ColorPicker theme="dark" :color="parseCSSValue(trimPrefix(modelValue))"
                      @changeColor="SetRGBIntToRGB" :sucker-hide="true" />
                  </ListboxOptions>
                </transition>
              </div>
            </Listbox>
          </div>
        </div>
      </div>
    </div>
    <!-- </PopoverPanel>
    </transition> -->
  </Popover>
</template>

<style lang="css">
.hu-color-picker {
  min-width: 218px;
}
</style>

<script>
import LoadingIcon from "@/components/LoadingIcon.vue";

import {
  Listbox,
  ListboxButton,
  ListboxLabel,
  ListboxOption,
  ListboxOptions,
  Switch,
  SwitchGroup,
  SwitchLabel,
  Popover,
  PopoverButton,
  PopoverPanel,
} from "@headlessui/vue";

import { CheckIcon, SelectorIcon, ChevronDownIcon } from "@heroicons/vue/solid";
import { XIcon } from "@heroicons/vue/outline";

import { ref } from "vue";
import DiscordEmbed from "@/components/DiscordEmbed.vue";

import { ColorPicker } from "vue-color-kit";
import "vue-color-kit/dist/vue-color-kit.css";
import parse from "parse-css-color";

import backgrounds from "@/backgrounds.json";

const tabs = [
  { name: "Welcomer", value: 1, enabled: true },
  { name: "Solid Colour", value: 4, enabled: true },
  { name: "Custom", value: 2, enabled: true },
  // { name: "Unsplash", icon: ["fab", "unsplash"], value: 3, enabled: true },
];

const backgroundRoot = (id) => `/assets/backgrounds/${id}.webp`;
const customRoot = (id) => `/api/welcomer/preview/${id}`;

const solidColourPrefix = "solid:";
const unsplashPrefix = "unsplash:";
const customPrefix = "custom:";

const solidColourProfileBased = "profile";

export default {
  components: {
    Listbox,
    ListboxButton,
    ListboxLabel,
    ListboxOption,
    ListboxOptions,
    Switch,
    SwitchGroup,
    SwitchLabel,
    CheckIcon,
    SelectorIcon,
    ChevronDownIcon,
    XIcon,
    LoadingIcon,
    Popover,
    PopoverButton,
    PopoverPanel,
    DiscordEmbed,
    ColorPicker,
  },

  props: {
    modelValue: {
      type: null,
      required: false,
    },
    disabled: {
      type: Boolean,
    },
    invalid: {
      type: Boolean,
    },
    files: {
      type: Object,
      required: false,
    },
    customImages: {
      type: Array,
      required: false,
    },
  },

  setup(props) {
    let initialPage = 1;

    if (props.modelValue.startsWith(customPrefix)) {
      initialPage = 2;
    } else if (props.modelValue.startsWith(solidColourPrefix)) {
      initialPage = 4;
    }

    let page = ref(initialPage);

    let displayEmbed = ref({
      embeds: [
        {
          // image: {
          //   url: getBackgroundName(props.modelValue),
          // },
        },
      ],
    });

    let solidColourIsProfileBased = ref(
      props.modelValue == solidColourPrefix + solidColourProfileBased
    );

    return {
      tabs,
      page,
      displayEmbed,

      solidColourIsProfileBased,

      solidColourPrefix,
      solidColourProfileBased,
      unsplashPrefix,
      customPrefix,

      backgroundRoot,
      customRoot,

      backgrounds,
    };
  },

  emits: ["update:modelValue", "update:files"],

  methods: {
    updateValue(value) {
      this.$emit("update:modelValue", value);
    },

    updateFiles(value) {
      this.$emit("update:files", value);
    },

    RGBIntToRGB(rgbInt, defaultValue) {
      if (rgbInt.startsWith(solidColourPrefix)) {
        rgbInt = rgbInt.slice(solidColourPrefix.length);
      } else {
        rgbInt = defaultValue;
      }

      if (parseInt(rgbInt, 10) != rgbInt) {
        rgbInt = defaultValue;
      }

      return (
        "#" +
        parseInt(rgbInt == undefined ? defaultValue : rgbInt, 10)
          .toString(16)
          .slice(-6)
          .padStart(6, "0")
      );
    },

    removeFiles() {
      this.updateValue("default");
      this.updateFiles([]);
    },

    onFileUpdate(event) {
      if (event.target.files.length > 0) {
        if (event.target.files[0].size > 20000000) {
          this.$store.dispatch("createToast", {
            title: "Your file is too large. It must be 20MB or less!",
            icon: "xmark",
            class: "text-red-500 bg-red-100",
          });

          return;
        } else {
          this.$store.dispatch("createToast", {
            title:
              "Your custom background will be uploaded when changes are saved.",
            icon: "info",
            class: "text-blue-500 bg-blue-100",
          });
        }
      }

      this.updateValue("custom:upload");
      this.updateFiles(event.target.files);
    },

    SetRGBIntToRGB(color) {
      var { r, g, b, a } = color.rgba;

      if (a == 1) {
        this.updateValue(solidColourPrefix + color.hex);
      } else {
        a = Math.round(a * 100) / 100;

        this.updateValue(solidColourPrefix + `rgba(${r}, ${g}, ${b}, ${a})`);
      }
    },

    trimPrefix(value) {
      return value.replace(solidColourPrefix, "");
    },

    formatBytes(size) {
      var mb = size / 1024000;
      return mb.toFixed(2);
    },

    parseCSSValue(value, defaultValue) {
      var result;

      result = parse(value);

      if (result == null) {
        result = parse(defaultValue);
      }

      if (result == null) {
        result = parse("#FFFFFF");
      }

      var [r, g, b] = result.values;
      var a = result.alpha;

      if (a == 1) {
        return `#${r.toString(16).toUpperCase().padStart(2, "0")}${g
          .toString(16)
          .toUpperCase()
          .padStart(2, "0")}${b.toString(16).toUpperCase().padStart(2, "0")}`;
      } else {
        a = Math.round(a * 100) / 100;

        return `rgba(${r}, ${g}, ${b}, ${a})`;
      }
    },
  },
};
</script>
