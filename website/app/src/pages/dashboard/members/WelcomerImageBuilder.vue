<template>
  <div class="builder-container">
    <div v-if="isDataError" class="w-full h-screen flex items-center justify-center">
      <p>Failed to load builder</p>
      <button @click="fetchConfig" class="ml-2 px-3 py-1 bg-primary text-white rounded">
        Retry
      </button>
      <button @click="goBack" class="ml-2 px-3 py-1 bg-gray-600 text-white rounded">
        Go Back
      </button>
    </div>
    <div v-else-if="!isDataFetched" class="w-full h-screen flex items-center justify-center">
      <LoadingIcon :isLight="true" />
    </div>
    <div v-else class="builder-portal">
      <div class="builder-canvas">
        <div class="m-4 flex gap-2 z-50 absolute">
          <button class="border border-secondary-light shadow-md bg-secondary px-4 py-1 rounded-full w-fit cursor-pointer" @click="x = defaultX; y = defaultY">
            {{ defaultX - x }}, {{ defaultY - y }}
            {{ selectedObject > -1 ? 'object selected: ' + selectedObject : '' }}
          </button>
          <button class="border border-secondary-light shadow-md bg-secondary px-4 py-1 rounded-full w-fit cursor-pointer" v-if="zoom != defaultZoom" @click="zoom = defaultZoom">
            {{ Math.round((zoom/defaultZoom) * 100) }}%
            <span class="sr-only">Reset zoom</span>
          </button>
        </div>
        <div class="canvas" :style="getCanvasStyle(x, y)">
          <div v-for="(obj, index) in image_config.layers" :key="index">
            <div @mouseover="onLayerMouseOver(index)" @mouseleave="onLayerMouseLeave()" :style="getObjectStyleBase(obj, index)">
              <div v-if="obj.type == CustomWelcomerImageLayerTypeText" :style="getObjectStyle(obj, index)" class="pointer-events-none"><span v-html="marked(obj.value, true)"></span></div>
              <img v-else-if="obj.type == CustomWelcomerImageLayerTypeImage" :style="getObjectStyle(obj, index)" class="pointer-events-none" :src="obj.value" />
              <div v-else-if="obj.type == CustomWelcomerImageLayerTypeShapeRectangle || obj.type == CustomWelcomerImageLayerTypeShapeCircle" :style="getObjectStyle(obj, index)" class="pointer-events-none"></div>

              <div v-if="hoveredObject == index || selectedObject == index" class="outline outline-[#0078D7] outline-2 w-full h-full" @mousedown="onGrabStart(0)">
                <div v-if="selectedObject == index" class="absolute -left-[11px] -top-[11px] w-2 h-2 m-2 bg-white border-2 border-[#0078D7] cursor-grab" @mousedown="onGrabStart(1)"></div>
                <div v-if="selectedObject == index" class="absolute -right-[11px] -top-[11px] w-2 h-2 m-2 bg-white border-2 border-[#0078D7] cursor-grab" @mousedown="onGrabStart(2)"></div>
                <div v-if="selectedObject == index" class="absolute -left-[11px] -bottom-[11px] w-2 h-2 m-2 bg-white border-2 border-[#0078D7] cursor-grab" @mousedown="onGrabStart(3)"></div>
                <div v-if="selectedObject == index" class="absolute -right-[11px] -bottom-[11px] w-2 h-2 m-2 bg-white border-2 border-[#0078D7] cursor-grab" @mousedown="onGrabStart(4)"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="builder-sidebar">
        <div class="p-4 border-b border-secondary-light">
          <span class="font-semibold text-sm">Layers</span>
          <ul>
            <li v-for="(obj, index) in image_config.layers" :key="index" @click="onLayerClick(index)" @mouseover="onLayerMouseOver(index)" @mouseleave="onLayerMouseLeave()"
              :class="[selectedObject == index ? 'bg-primary text-white' : hoveredObject == index ? 'bg-secondary-light text-white' : 'hover:bg-secondary-light cursor-pointer', ' whitespace-nowrap overflow-ellipsis overflow-hidden px-2 py-1 rounded-md hover:cursor-pointer']">
              <span v-if="obj.type == CustomWelcomerImageLayerTypeText">
                <font-awesome-icon icon="text" class="mr-2" />
                {{ obj.value || 'Text ' + (index+1) }}
              </span>
              <span v-else-if="obj.type == CustomWelcomerImageLayerTypeImage">
                <font-awesome-icon icon="image" class="mr-2" />
                {{ obj.value || 'Image ' + (index+1) }}
              </span>
              <span v-else-if="obj.type == CustomWelcomerImageLayerTypeShapeRectangle">
                <font-awesome-icon icon="square" class="mr-2" />
                Rectangle {{ index+1 }}
              </span>
              <span v-else-if="obj.type == CustomWelcomerImageLayerTypeShapeCircle">
                <font-awesome-icon icon="circle" class="mr-2" />
                Circle {{ index+1 }}
              </span>
            </li>
          </ul>
        </div>
        <div class="overflow-y-auto flex-1 divide-secondary-light divide-y">
          <div class="p-4" v-if="selectedObject > -1">
            <span class="font-semibold text-sm mb-2 block">Position</span>
            <div class="grid grid-cols-2 gap-2">
              <InputCalculator type="number" v-model="image_config.layers[selectedObject].position[0]">
                X
              </InputCalculator> 
              <InputCalculator type="number" v-model="image_config.layers[selectedObject].position[1]">
                Y
              </InputCalculator>
            </div>
          </div>
          <div class="p-4" v-if="selectedObject > -1">
            <span class="font-semibold text-sm mb-2 block">Layout</span>
            <div class="grid grid-cols-2 gap-2">
              <InputCalculator min="16" type="number" v-model="image_config.layers[selectedObject].dimensions[0]">
                Width
              </InputCalculator> 
              <InputCalculator min="16" type="number" v-model="image_config.layers[selectedObject].dimensions[1]">
                Height
              </InputCalculator>
            </div>
          </div>
          <div class="p-4" v-if="selectedObject > -1">
            <span class="font-semibold text-sm mb-2 block">Appearance</span>
            <span class="block text-neutral-500 text-xs font-medium">Border Radius</span>
            <div class="grid grid-cols-2 gap-2 mt-1">
              <InputCalculator type="number" min="0" minPercentage="0" maxPercentage="100" v-model="image_config.layers[selectedObject].border_radius[0]">
                Top Left
              </InputCalculator>
              <InputCalculator type="number" min="0" minPercentage="0" maxPercentage="100" v-model="image_config.layers[selectedObject].border_radius[1]">
                Top Right
              </InputCalculator>
              <InputCalculator type="number" min="0" minPercentage="0" maxPercentage="100" v-model="image_config.layers[selectedObject].border_radius[3]">
                Bottom Left
              </InputCalculator>
              <InputCalculator type="number" min="0" minPercentage="0" maxPercentage="100" v-model="image_config.layers[selectedObject].border_radius[2]">
                Bottom Right
              </InputCalculator>
            </div>
          </div>
          <div class="p-4" v-if="selectedObject > -1 && image_config.layers[selectedObject].type == CustomWelcomerImageLayerTypeText">
            <span class="font-semibold text-sm mb-2 block">Typography</span>
            <Listbox as="div" :model="image_config.layers[selectedObject].typography.font_family" @update:modelValue="updateTypographyFontFamily($event)">
              <div class="relative">
                <ListboxButton class="bg-secondary-dark relative w-full py-2 pl-3 pr-10 text-left border border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm">
                  {{ fonts[image_config.layers[selectedObject].typography.font_family]?.name || 'Select Font' }}
                </ListboxButton>
                <ListboxOptions class="absolute z-20 w-full mt-1 overflow-auto text-base bg-white dark:bg-secondary-dark rounded-md shadow-sm max-h-60 ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                  <ListboxOption v-for="(font, fontKey) in fonts" :key="fontKey" :value="fontKey" v-slot="{ active, selected }">
                    <li :class="[
                      active ? 'text-white bg-primary' : 'text-gray-900 dark:text-gray-50',
                      'cursor-default select-none relative py-2 pl-3 pr-9']">                    
                      {{ font.name }}
                    </li>
                  </ListboxOption>
                </ListboxOptions>
              </div>
            </Listbox>
            <div class="grid grid-cols-2 gap-2 mt-2">
              <Listbox as="div" v-model="image_config.layers[selectedObject].typography.font_weight">
                <div class="relative">
                  <ListboxButton class="bg-secondary-dark relative w-full py-2 pl-3 pr-10 text-left border border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm">
                    {{ fonts[image_config.layers[selectedObject].typography.font_family]?.weights[image_config.layers[selectedObject].typography.font_weight] ? image_config.layers[selectedObject].typography.font_weight : (fonts[image_config.layers[selectedObject].typography.font_family] ? fonts[image_config.layers[selectedObject].typography.font_family].default : 'normal') }}
                  </ListboxButton>
                  <ListboxOptions class="absolute z-20 w-full mt-1 overflow-auto text-base bg-white dark:bg-secondary-dark rounded-md shadow-sm max-h-60 ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                    <ListboxOption v-for="(weight, weightKey) in fonts[image_config.layers[selectedObject].typography.font_family]?.weights" :key="weightKey" :value="weightKey" v-slot="{ active, selected }">
                      <li :class="[
                        active ? 'text-white bg-primary' : 'text-gray-900 dark:text-gray-50',
                        'cursor-default select-none relative py-2 pl-3 pr-9']">                    
                        {{ weightKey }}
                      </li>
                    </ListboxOption>
                  </ListboxOptions>
                </div>
              </Listbox>
              <InputCalculator type="number" min="8" v-model="image_config.layers[selectedObject].typography.font_size"></InputCalculator>
              <InputCalculator type="number" min="0.1" step="0.1" v-model="image_config.layers[selectedObject].typography.line_height">Line Height</InputCalculator>
              <InputCalculator type="number" step="0.1" v-model="image_config.layers[selectedObject].typography.letter_spacing">Letter Spacing</InputCalculator>
            </div>
            <span class="block text-neutral-500 text-xs font-medium mt-2">Alignment</span>
            <div class="grid grid-cols-2 gap-2 mt-1">
              <Listbox as="div" v-model="image_config.layers[selectedObject].typography.horizontal_alignment">
                <div class="relative">
                  <ListboxButton class="bg-secondary-dark relative w-full py-2 pl-3 pr-10 text-left border border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm">
                    {{ image_config.layers[selectedObject].typography.horizontal_alignment || 'Left' }}
                  </ListboxButton>
                  <ListboxOptions class="absolute z-20 w-full mt-1 overflow-auto text-base bg-white dark:bg-secondary-dark rounded-md shadow-sm max-h-60 ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                    <ListboxOption value="left" v-slot="{ active }"><li :class="[active ? 'text-white bg-primary' : 'text-gray-900 dark:text-gray-50','cursor-default select-none relative py-2 pl-3 pr-9']">Left</li></ListboxOption>
                    <ListboxOption value="center" v-slot="{ active }"><li :class="[active ? 'text-white bg-primary' : 'text-gray-900 dark:text-gray-50','cursor-default select-none relative py-2 pl-3 pr-9']">Center</li></ListboxOption>
                    <ListboxOption value="right" v-slot="{ active }"><li :class="[active ? 'text-white bg-primary' : 'text-gray-900 dark:text-gray-50','cursor-default select-none relative py-2 pl-3 pr-9']">Right</li></ListboxOption>
                  </ListboxOptions>
                </div>
              </Listbox>
              <Listbox as="div" v-model="image_config.layers[selectedObject].typography.vertical_alignment">
                <div class="relative">
                  <ListboxButton class="bg-secondary-dark relative w-full py-2 pl-3 pr-10 text-left border border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm">
                    {{ image_config.layers[selectedObject].typography.vertical_alignment || 'Center' }}
                  </ListboxButton>
                  <ListboxOptions class="absolute z-20 w-full mt-1 overflow-auto text-base bg-white dark:bg-secondary-dark rounded-md shadow-sm max-h-60 ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                    <ListboxOption value="start" v-slot="{ active }"><li :class="[active ? 'text-white bg-primary' : 'text-gray-900 dark:text-gray-50','cursor-default select-none relative py-2 pl-3 pr-9']">Top</li></ListboxOption>
                    <ListboxOption value="center" v-slot="{ active }"><li :class="[active ? 'text-white bg-primary' : 'text-gray-900 dark:text-gray-50','cursor-default select-none relative py-2 pl-3 pr-9']">Center</li></ListboxOption>
                    <ListboxOption value="end" v-slot="{ active }"><li :class="[active ? 'text-white bg-primary' : 'text-gray-900 dark:text-gray-50','cursor-default select-none relative py-2 pl-3 pr-9']">Bottom</li></ListboxOption>
                  </ListboxOptions>
                </div>
              </Listbox>
            </div>
          </div>
          <!-- color fill -->
          <div class="p-4" v-if="selectedObject > -1 && (
            image_config.layers[selectedObject].type == CustomWelcomerImageLayerTypeText ||
            image_config.layers[selectedObject].type == CustomWelcomerImageLayerTypeShapeRectangle ||
            image_config.layers[selectedObject].type == CustomWelcomerImageLayerTypeShapeCircle)">
            <span class="font-semibold text-sm mb-2 block">Fill</span>
            <Listbox as="div" class="flex-1">
            <div class="relative">
              <ListboxButton
                class="relative w-full py-2 pl-3 pr-10 text-left bg-white border border-gray-300 dark:bg-secondary dark:border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm">
                <div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                  <font-awesome-icon icon="square" class="inline w-4 h-4 mr-1 border-primary" :style="{
                    color: `${image_config.layers[selectedObject].fill}`,
                  }" />
                </div>
                <span class="block pl-10 truncate">{{
                  image_config.layers[selectedObject].fill.toUpperCase()
                }}</span>
                <span class="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
                  <SelectorIcon class="w-5 h-5 text-gray-400" aria-hidden="true" />
                </span>
              </ListboxButton>

              <transition leave-active-class="transition duration-100 ease-in" leave-from-class="opacity-100"
                leave-to-class="opacity-0">
                <ListboxOptions class="absolute z-10 mt-1">
                  <ColorPicker theme="dark" :color="image_config.layers[selectedObject].fill || '#000000'"
                    @changeColor="image_config.layers[selectedObject].fill = rgbaToHex($event)" :sucker-hide="true" />
                </ListboxOptions>
              </transition>
            </div>
          </Listbox>
          </div>
          <!-- stroke -->
          <div class="p-4" v-if="selectedObject > -1 && image_config.layers[selectedObject].type == CustomWelcomerImageLayerTypeText">
            <span class="font-semibold text-sm mb-2 block">Text</span>
            <AutocompleteInput type="text" :isTextarea="true"
              class="border rounded p-2 bg-transparent w-full"
              placeholder="Message Content" rows="4" :value="image_config.layers[selectedObject].value" @update:modelValue="image_config.layers[selectedObject].value = $event"/>
          </div>
          <div class="p-4" v-if="selectedObject > -1 && image_config.layers[selectedObject].type == CustomWelcomerImageLayerTypeImage">
            <span class="font-semibold text-sm mb-2 block">Image URL</span>
            <AutocompleteInput type="text"
              class="border rounded p-2 bg-transparent w-full"
              placeholder="Image URL" :value="image_config.layers[selectedObject].value" @update:modelValue="image_config.layers[selectedObject].value = $event"/>
          </div>
        </div>
      </div>
    </div>
    
    <unsaved-changes :unsavedChanges="unsavedChanges" :isChangeInProgress="isChangeInProgress"
    v-on:save="saveConfig"></unsaved-changes>
    <Toast />
  </div>
</template>

<style lang="scss">
body {
  @apply overflow-hidden;
}
.builder-container {
  @apply h-screen;
}
.builder-portal {
  @apply h-screen bg-secondary-light flex flex-row;
}
.builder-canvas {
  @apply flex-1 min-w-[50%] overflow-hidden relative;
}
.builder-sidebar {
  @apply bg-secondary border-secondary-light w-64 border-l shadow-md flex flex-col;
}
</style>

<script>
import { ref } from "vue";
import store from "@/store/index";
import { useRoute } from "vue-router";

import AutocompleteInput from "@/components/AutocompleteInput.vue";
import InputCalculator from "@/components/dashboard/InputCalculator.vue";
import UnsavedChanges from "@/components/dashboard/UnsavedChanges.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";

import dashboardAPI from "@/api/dashboard";
import endpoints from "@/api/endpoints";

import Toast from "@/components/dashboard/Toast.vue";

import { ColorPicker } from "vue-color-kit";
import "vue-color-kit/dist/vue-color-kit.css";

import {
  Listbox,
  ListboxButton,
  ListboxOption,
  ListboxOptions,
} from "@headlessui/vue";

import {
  getErrorToast,
  getSuccessToast,
  getValidationToast,
  navigateToErrors,
  marked,
} from "@/utilities";

const fonts = {
  "Balsamiq Sans": {
    name: " Balsamiq Sans",
    default: "regular",
    weights: {
      "regular": "400",
      "bold": "700",
    },
  },

  "Fredoka": {
    name: " Fredoka",
    default: "regular",
    weights: {
      "300": "300",
      "regular": "400",
      "500": "500",
      "600": "600",
      "bold": "700",
    },
  },

  "Inter": {
    name: " Inter",
    default: "regular",
    weights: {
      "100": "100",
      "200": "200",
      "300": "300",
      "regular": "400",
      "500": "500",
      "600": "600",
      "bold": "700",
      "800": "800",
      "900": "900",
    },
  },

  "Luckiest Guy": {
    name: "Luckiest Guy",
    default: "regular",
    weights: {
      "regular": "400",
    },
  },

  "Mada": {
    name: "Mada",
    default: "regular",
    weights: {
      "200": "200",
      "300": "300",
      "regular": "400",
      "500": "500",
      "600": "600",
      "bold": "700",
      "800": "800",
      "900": "900",
    },
  },

  "Nunito": {
    name: "Nunito",
    default: "regular",
    weights: {
      "200": "200",
      "300": "300",
      "regular": "400",
      "600": "600",
      "bold": "700",
      "800": "800",
      "900": "900",
      "1000": "1000",
    },
  },

  "Poppins": {
    name: "Poppins",
    default: "regular",
    weights: {
      "100": "100",
      "200": "200",
      "300": "300",
      "regular": "400",
      "500": "500",
      "600": "600",
      "bold": "700",
      "800": "800",
      "900": "900",
    },
  },

  "Raleway": {
    name: "Raleway",
    default: "regular",
    weights: {
      "100": "100",
      "200": "200",
      "300": "300",
      "regular": "400",
      "500": "500",
      "600": "600",
      "bold": "700",
      "800": "800",
      "900": "900",
    },
  },

  // web safe fonts
  "Arial": {
    name: "Arial",
    websafe: true,
    default: "regular",
    weights: {
      "regular": "400",
      "bold": "700",
    },
  },

  "Verdana": {
    name: "Verdana",
    websafe: true,
    default: "regular",
    weights: {
      "regular": "400",
      "bold": "700",
    },
  },

  "Tahoma": {
    name: "Tahoma",
    websafe: true,
    default: "regular",
    weights: {
      "regular": "400",
      "bold": "700",
    },
  },

  "Trebuchet MS": {
    name: "Trebuchet MS",
    websafe: true,
    default: "regular",
    weights: {
      "regular": "400",
      "bold": "700",
    },
  },

  "Times New Roman": {
    name: "Times New Roman",
    websafe: true,
    default: "regular",
    weights: {
      "regular": "400",
      "bold": "700",
    },
  },

  "Georgia": {
    name: "Georgia",
    websafe: true,
    default: "regular",
    weights: {
      "regular": "400",
      "bold": "700",
    },
  },

  "Garamond": {
    name: "Garamond",
    websafe: true,
    default: "regular",
    weights: {
      "regular": "400",
      "bold": "700",
    },
  },

  "Courier New": {
    name: "Courier New",
    websafe: true,
    default: "regular",
    weights: {
      "regular": "400",
      "bold": "700",
    },
  },
}

const CustomWelcomerImageLayerTypeText = 0;
const CustomWelcomerImageLayerTypeImage = 1;
const CustomWelcomerImageLayerTypeShapeRectangle = 2;
const CustomWelcomerImageLayerTypeShapeCircle = 3;

export default {
  components: {
    ColorPicker,
    AutocompleteInput,
    InputCalculator,
    UnsavedChanges,
    LoadingIcon,
    Toast,
    Listbox,
    ListboxButton,
    ListboxOption,
    ListboxOptions,
  },
  watch: {
    "$route.params.guildID"(to) {
      store.commit("setSelectedGuild", to);
    },
  },
  setup() {
    store.watch(
    () => store.getters.getSelectedGuildID,
    () => {
      if (store.getters.getSelectedGuildID !== undefined) {
        store.dispatch("fillGuild");
      }
    }
    );
    
    const route = useRoute();
    
    let guildID = route.params.guildID;
    if (guildID !== undefined) {
      store.commit("setSelectedGuild", guildID);
    }
    
    let isDataFetched = ref(false);
    let isDataError = ref(false);
    let unsavedChanges = ref(false);
    let isChangeInProgress = ref(false);
    
    let config = ref({});
    let files = ref([]);
    
    let image_config = ref({});

    let x = ref(0);
    let y = ref(0);
    let zoom = ref(1.0);

    // Used to make zoom and coordinates relative to default fit
    let defaultX = ref(0);
    let defaultY = ref(0);
    let defaultZoom = ref(1.0);

    let preview = ref(false);

    let selectedObject = ref(-1);
    let hoveredObject = ref(-1);
    let selectedGrab = ref(-1);

    return {
      config,
      files,
      image_config,
      
      isDataFetched,
      isDataError,
      unsavedChanges,
      isChangeInProgress,
      
      getErrorToast,
      x, y, zoom,
      defaultX, defaultY, defaultZoom,

      preview,
      fonts,

      selectedObject, hoveredObject, selectedGrab,

      CustomWelcomerImageLayerTypeText, CustomWelcomerImageLayerTypeImage, CustomWelcomerImageLayerTypeShapeRectangle, CustomWelcomerImageLayerTypeShapeCircle,
    }
  },
  
  mounted() {
    this.fetchConfig();

    window.addEventListener('resize', this.fitCanvas);
    window.addEventListener('orientationchange', this.fitCanvas);
    window.addEventListener('load', this.fitCanvas);

    window.addEventListener('keydown', (e) => {
      // do nothing if anything is focused
      if (document.activeElement && document.activeElement !== document.body) return;

      if (this.selectedObject == -1) {
        let incr = e.shiftKey ? 16 : 4;
        switch (e.key) {
          case "+":
          case "=": this.zoom = Math.min(2.0, this.zoom + 0.1); break;
          case "-":
          case "_": this.zoom = Math.max(0.1, this.zoom - 0.1); break;
          case "ArrowLeft": this.x += incr; break;
          case "ArrowRight": this.x -= incr; break;
          case "ArrowUp": this.y += incr; break;
          case "ArrowDown": this.y -= incr; break;
        }
      } else {
        let obj = this.image_config.layers[this.selectedObject];
        if (!obj) return;

        let incr = e.shiftKey ? 8 : 1;

        switch (e.key) {
          case "ArrowLeft": obj.position[0] -= incr; break;
          case "ArrowRight": obj.position[0] += incr; break;
          case "ArrowUp": obj.position[1] -= incr; break;
          case "ArrowDown": obj.position[1] += incr; break;
        } 
      }
    });

    window.addEventListener('wheel', (e) => {
      // check cursor is over the builder-canvas
      const canvasRect = this.$el.querySelector('.builder-canvas').getBoundingClientRect();
      
      // check cursor is inside canvas area
      if (e.clientX < canvasRect.left || e.clientX > canvasRect.right || e.clientY < canvasRect.top || e.clientY > canvasRect.bottom) return;

      if (e.ctrlKey || e.metaKey) {
        // apply fixed zoom behaviour
        e.preventDefault();

        if (e.deltaY < 0) {
          this.zoom = Math.min(2.0, this.zoom + 0.1);
        } else {
          this.zoom = Math.max(0.1, this.zoom - 0.1);
        }
      } else {
        // apply inverse of wheel delta
        this.x -= e.deltaX;
        this.y -= e.deltaY;
      }
    });

    // Middle‑mouse drag to pan
    let isMiddleDown = false;
    let lastMouse = { x: 0, y: 0 };

    const onMouseDown = (e) => {
      if (e.button == 0) {
        // check mouse is inside canvas area
        const canvasRect = this.$el.querySelector('.builder-canvas').getBoundingClientRect();
        if (e.clientX < canvasRect.left || e.clientX > canvasRect.right ||
            e.clientY < canvasRect.top || e.clientY > canvasRect.bottom) return;

        if (e.clientX == 0 && e.clientY == 0) return; // ignore invalid coordinates

        // check if clicking on any layers using builder canvas children.
        for (let i = 0; i < this.$el.querySelector('.canvas').children.length; i++) {
          const child = this.$el.querySelector('.canvas').children[i];
          if (child.contains(e.target)) {
            this.onLayerClick(i);
            return;
          }
        }

        this.onLayerClick(-1);

      } else if (e.button == 1) {
        // middle button, start panning
        isMiddleDown = true;
        lastMouse.x = e.clientX;
        lastMouse.y = e.clientY;
        // prevent native autoscroll / middle-click default
        e.preventDefault();
        document.body.style.cursor = 'grabbing';
      }
    };

    const onMouseUp = (e) => {
      if (e.button == 0) {
        // left button
        this.onGrabEnd();
      } else {
        isMiddleDown = false;
        document.body.style.cursor = '';
      }
    };

    const onMouseMove = (e) => {
      const dx = e.clientX - lastMouse.x;
      const dy = e.clientY - lastMouse.y;

      lastMouse.x = e.clientX;
      lastMouse.y = e.clientY;

      if (isMiddleDown) {
        // Match wheel behavior (scroll/pan): apply inverse of pointer delta
        this.x += dx;
        this.y += dy;
      } else {
        this.onGrabMove(dx, dy)
      }
    };

    // prevent auxclick default on middle button as well
    const onAuxClick = (e) => {
      if (e.button === 1) e.preventDefault();
    };

    this.$el.addEventListener('mousedown', onMouseDown);
    this.$el.addEventListener('auxclick', onAuxClick);
    window.addEventListener('mousemove', onMouseMove);
    window.addEventListener('mouseup', onMouseUp);

  },
  
  methods: {
    marked(text, embed) {
        return marked(text, embed);
    },

    rgbaToHex(color) {
      var { r, g, b, a } = color.rgba;
      return (
        "#" +
        ((1 << 24) + (r << 16) + (g << 8) + b)
          .toString(16)
          .slice(1)
          .toUpperCase() +
        (a !== undefined && a < 1 ? Math.round(a * 255).toString(16).padStart(2, "0").toUpperCase() : "")
      );
    },

    fetchConfig() {
      this.isDataFetched = false;
      this.isDataError = false;
      
      dashboardAPI.getConfig(
      endpoints.EndpointGuildWelcomer(this.$store.getters.getSelectedGuildID),
      ({ config }) => {
        this.config = config;
        this.isDataFetched = true;
        this.isDataError = false;
        
        this.image_config = this.parseDict(config.images.custom_builder_data);
        this.preemptivelyLoadFonts();
        this.fitCanvas();
      },
      (error) => {
        this.$store.dispatch("createToast", getErrorToast(error));
        
        this.isDataFetched = true;
        this.isDataError = true;
      }
      );
    },
    
    parseDict(data) {
      try {
        return JSON.parse(data);
      } catch {
        return {};
      }
    },
    
    async saveConfig() {
      const validForm = await this.v$.$validate();
      
      if (!validForm) {
        this.$store.dispatch("createToast", getValidationToast());
        navigateToErrors();
        
        return;
      }
      
      this.isChangeInProgress = true;
      
      dashboardAPI.doPost(
      endpoints.EndpointGuildWelcomer(this.$store.getters.getSelectedGuildID),
      this.config,
      this.files,
      ({ config }) => {
        this.$store.dispatch("createToast", getSuccessToast());
        
        this.config = config;
        this.files = [];
        this.unsavedChanges = false;
        this.isChangeInProgress = false;
      },
      (error) => {
        this.$store.dispatch("createToast", getErrorToast(error));
        
        this.isChangeInProgress = false;
      }
      );
    },
    
    onValueUpdate() {
      this.unsavedChanges = true;
    },
    
    onFilesUpdate(event) {
      this.files = event;
      this.onValueUpdate();
    },
    
    goBack() {
      this.$router.push({
        name: "dashboard.guild.welcomer",
        params: { guildID: this.$store.getters.getSelectedGuildID },
      });
    },
    
    getCanvasStyle(x, y) {
      // this outputs as a style=""
      
      let dimensions = this.image_config?.dimensions || [1000, 300];

      return {
        left: "calc(" + x + "px)",
        top: "calc(50% + " + y + "px)",
        transform: "translateY(-50%) scale(" + this.zoom + ")",
        position: "absolute",
        width: (dimensions[0] || 1000) + "px",
        height: (dimensions[1] || 300) + "px",
        backgroundColor: this.getFillAsCSS(this.image_config.fill || '#ffffff'),
        overflow : this.preview ? "hidden" : "visible",
      };
    },

    getObjectStyleBase(obj, index) {
      return {
        position: "absolute",
        userSelect: "none",

        zIndex: this.image_config.layers.length - index,

        width: (obj.dimensions && obj.dimensions[0] ? obj.dimensions[0] + "px" : "auto"),
        height: (obj.dimensions && obj.dimensions[1] ? obj.dimensions[1] + "px" : "auto"),

        left: (obj.position && obj.position[0] ? obj.position[0] + "px" : "0px"),
        top: (obj.position && obj.position[1] ? obj.position[1] + "px" : "0px"),

        // transform:
        //   (obj.rotation ? "rotate(" + obj.rotation + "deg) " : "") +
        //   "scale(" + (obj.inverted_x ? "-1" : "1") + ", " + (obj.inverted_y ? "-1" : "1") + ")",
      }
    },

    getObjectStyle(obj, index) {
      let styles = this.getObjectStyleBase(obj, index);

      styles.zIndex = 0;
      styles.transform = "";
      styles.left = "0px";
      styles.top = "0px";

      styles.borderRadius = obj.type == CustomWelcomerImageLayerTypeShapeCircle ? "100%" : 
          this.normalizeBorderRadius(obj.border_radius[0]) + " " +
          this.normalizeBorderRadius(obj.border_radius[1]) + " " +
          this.normalizeBorderRadius(obj.border_radius[2]) + " " +
          this.normalizeBorderRadius(obj.border_radius[3])

      styles.backgroundColor = (obj.type != CustomWelcomerImageLayerTypeText ? this.getFillAsCSS(obj.fill) : "transparent")
      styles.color = (obj.type == CustomWelcomerImageLayerTypeText ? this.getFillAsCSS(obj.fill) : "inherit")
      styles.border = (obj.stroke?.width > 0 ? obj.stroke.width + "px solid " + this.getFillAsCSS(obj.stroke.color) : "none")

      if (obj.type == CustomWelcomerImageLayerTypeText) {
        let font = this.fonts[obj.typography?.font_family];
        if (font) {
          if (font.websafe) {
            styles.fontFamily = obj.typography.font_family;
            styles.fontWeight = (obj.typography?.font_weight && obj.typography.font_weight != "" ? font.weights[obj.typography.font_weight] : "normal")
          } else {
            styles.fontFamily = "'" + obj.typography.font_family + "', sans-serif";
            styles.fontWeight = (obj.typography?.font_weight && obj.typography.font_weight != "" ? font.weights[obj.typography.font_weight] : "normal")
          }
        } else {
          console.warn("font not found:", obj.typography?.font_family);
          styles.fontFamily = "inherit";
          styles.fontWeight = "normal";
        }

        styles.fontSize      = (obj.typography?.font_size && obj.typography.font_size != 0 ? obj.typography.font_size + "px" : "auto")
        styles.lineHeight    = (obj.typography?.line_height && obj.typography.line_height != 0 ? obj.typography.line_height + "em" : "normal")
        styles.letterSpacing = (obj.typography?.letter_spacing && obj.typography.letter_spacing != 0 ? obj.typography.letter_spacing + "px" : "normal")

        styles.display        = "flex"
        styles.justifyContent = this.normalizeHorizontalAlignment(obj.typography?.horizontal_alignment)
        styles.alignItems     = this.normalizeVerticalAlignment(obj.typography?.vertical_alignment)
        styles.whiteSpace     = "pre-wrap"
      }

      return styles;
    },

    onLayerClick(index) {
      if (this.selectedObject != index) this.selectedObject = index;
    },

    onLayerMouseOver(index) {
      this.hoveredObject = index;
    },

    onLayerMouseLeave() {
      this.hoveredObject = -1;
    },

    onGrabStart(index) {
      if (this.selectedGrab > -1) return; // already grabbing

      document.body.style.cursor = 'grabbing';
      this.selectedGrab = index;
    },

    onGrabEnd() {
      document.body.style.cursor = 'default';
      this.selectedGrab = -1;
    },

    onGrabMove(x, y) {
      if (this.selectedGrab == -1) return;

      let obj = this.image_config.layers[this.selectedObject];
      if (!obj) return;

      let minWidth = Math.max(10, obj.stroke.width * 2);
      let minHeight = Math.max(10, obj.stroke.width * 2);

      if (obj.dimensions[0] < minWidth && obj.dimensions[1] < minHeight) return;

      switch (this.selectedGrab) {
        case 0: // nothing, move entire object
          obj.position[0] += x * (1/this.zoom);
          obj.position[1] += y * (1/this.zoom);
          break;
        case 1: // top-left
          obj.position[0] += x * (1/this.zoom);
          obj.dimensions[0] -= x * (1/this.zoom);
          obj.position[1] += y * (1/this.zoom);
          obj.dimensions[1] -= y * (1/this.zoom);
          break;
        case 2: // top-right
          obj.dimensions[0] += x * (1/this.zoom);
          obj.position[1] += y * (1/this.zoom);
          obj.dimensions[1] -= y * (1/this.zoom);
          break;
        case 3: // bottom-left
          obj.position[0] += x * (1/this.zoom);
          obj.dimensions[0] -= x * (1/this.zoom);
          obj.dimensions[1] += y * (1/this.zoom);
          break;
        case 4: // bottom-right
          obj.dimensions[0] += x * (1/this.zoom);
          obj.dimensions[1] += y * (1/this.zoom);
          break;
      }

      obj.dimensions[0] = Math.max(minWidth, obj.dimensions[0]);
      obj.dimensions[1] = Math.max(minHeight, obj.dimensions[1]);

      // round position and dimensions to integers
      obj.position[0] = Math.round(obj.position[0]);
      obj.position[1] = Math.round(obj.position[1]);
      obj.dimensions[0] = Math.round(obj.dimensions[0]);
      obj.dimensions[1] = Math.round(obj.dimensions[1]);
    },

    normalizeBorderRadius(value) {
      value = String(value);
      if (value.endsWith("%")) return value;
      if (value !== "") return value + "px";
      return "0px";
    },

    normalizeHorizontalAlignment(value) {
      if (value === "center" || value === "right") return value;
      return "left";
    },

    normalizeVerticalAlignment(value) {
      if (value === "start" || value === "end") return value;
      return "center";
    },

    updateTypographyFontFamily(value) {
      let obj = this.image_config.layers[this.selectedObject];
      if (!obj) return;

      let font = this.fonts[obj.typography.font_family];
      if (!font) return;

      this.image_config.layers[this.selectedObject].typography.font_family = value;
      this.image_config.layers[this.selectedObject].typography.font_weight = font.default;

      this.loadFont(value, font.default);
    },
    
    preemptivelyLoadFonts() {
      if (!this.image_config.layers) return;

      for (let layer of this.image_config.layers) {
        if (layer.type != CustomWelcomerImageLayerTypeText) continue;
        if (!layer.typography || !layer.typography.font_family) continue;

        this.loadFont(layer.typography.font_family, layer.typography.font_weight);
      }
    },

    loadFont(fontFamily, fontWeight) {
      let font = this.fonts[fontFamily];
      if (!font) return;

      if (font.websafe) return; // no need to load web safe fonts

      let weight = font.weights[fontWeight] || font.weights[font.default] || "400";

      const linkId = `font-${fontFamily.replace(/\s+/g, '-')}-${weight}`;
      if (document.getElementById(linkId)) return; // already loaded

      const link = document.createElement('link');
      link.id = linkId;
      link.rel = 'stylesheet';
      link.href = `https://fonts.googleapis.com/css2?family=${fontFamily.replace(/\s+/g, '+')}:wght@${weight}&display=block`;
      document.head.appendChild(link);
    },

    fitCanvas() {
      // center and fit the canvas inside the container with padding

      this.$nextTick(() => {
        const padding = 32;
        const container = this.$el.querySelector('.builder-canvas');
        if (!container) return;

        const canvas = this.$el.querySelector('.canvas');
        if (!canvas) return;

        this.x = (container.clientWidth - canvas.clientWidth) / 2;
        this.y = 0;
        this.zoom = Math.min(1, (container.clientWidth - (padding * 2)) / canvas.clientWidth);

        this.defaultZoom = this.zoom;
        this.defaultX = this.x;
        this.defaultY = this.y;
      });
    },
    
    getFillAsCSS(value) {
      if (value.startsWith("#")) {
        return value;
      }
      
      if (value == "profile") {
        return this.generateLightHex();
      }
    },
    
    simpleSeededRandom(seedText) {
      // Turn seedText into a 32-bit integer hash
      let h = 0;
      for (let i = 0; i < seedText.length; i++) {
        h = (h * 31 + seedText.charCodeAt(i)) >>> 0;
      }

      // Simple LCG (Linear Congruential Generator)
      function next() {
        h = (h * 1664525 + 1013904223) >>> 0;
        return h & 0xff; // return 8-bit number
      }

      return [next(), next(), next()];
    },

    generateLightHex() {
      let r, g, b, luminance;
      
      let i = 0;

      do {
        i++;

        // Generate random RGB values (biased towards lighter colors)
        [r, g, b] = this.simpleSeededRandom(this.config + i);

        // Calculate luminance using the same formula as the Go code
        luminance = (0.299 * r + 0.587 * g + 0.114 * b) / 255;
        
      } while (luminance <= 0.7);
      
      // Convert to hex
      const toHex = (n) => n.toString(16).padStart(2, '0');
      return `#${toHex(r)}${toHex(g)}${toHex(b)}`;
    },
  }
}
</script>