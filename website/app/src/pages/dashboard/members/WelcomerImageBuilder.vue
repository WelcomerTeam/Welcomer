<!-- TODO:

- canvas controls
  - image upload
- enhancement: snap for vertical or horizontal alignments
- enhancement: 1:1 aspect ratio when drawing shapes
-->
<template>
  <div class="builder-container">
    <div v-if="isDataError" class="w-full h-dvh flex items-center justify-center">
      <p>Failed to load builder</p>
      <button @click="fetchConfig" class="ml-2 px-3 py-1 bg-primary text-white rounded">
        Retry
      </button>
      <button @click="goBack" class="ml-2 px-3 py-1 bg-gray-600 text-white rounded">
        Go Back
      </button>
    </div>
    <div v-else-if="!isDataFetched" class="w-full h-dvh flex items-center justify-center">
      <LoadingIcon :isLight="true" />
    </div>
    <div v-else class="builder-portal">
      <div class="builder-canvas">
        <div class="bg-secondary-dark border border-secondary-light absolute bottom-12 left-1/2 -translate-x-1/2 rounded-lg shadow-md z-50 flex flex-row divide-x divide-secondary-light">
          <div class="p-2 gap-2 flex flex-row h-full">
            <button @click="selectedAction = 0" :class="[selectedAction == 0 ? 'bg-primary' : 'hover:bg-secondary-light', 'h-12 w-12 rounded-md shadow-md flex justify-center items-center']">
              <font-awesome-icon icon="mouse-pointer" class="w-6 h-6 text-white" aria-hidden="true" />
              <span class="sr-only">Select</span>
            </button>
          </div>
          <div class="p-2 gap-2 flex flex-row h-full">
            <button @click="selectedAction = 1" :class="[selectedAction == 1 ? 'bg-primary' : 'hover:bg-secondary-light', 'h-12 w-12 rounded-md shadow-md flex justify-center items-center']">
              <font-awesome-icon icon="text" class="w-6 h-6 text-white" aria-hidden="true" />
              <span class="sr-only">Text</span>
            </button>
            <button @click="selectedAction = 2" :class="[selectedAction == 2 ? 'bg-primary' : 'hover:bg-secondary-light', 'h-12 w-12 rounded-md shadow-md flex justify-center items-center']">
              <font-awesome-icon icon="image" class="w-6 h-6 text-white" aria-hidden="true" />
              <span class="sr-only">Image</span>
            </button>
            <button @click="selectedAction = 3" :class="[selectedAction == 3 ? 'bg-primary' : 'hover:bg-secondary-light', 'h-12 w-12 rounded-md shadow-md flex justify-center items-center']">
              <font-awesome-icon icon="square" class="w-6 h-6 text-white" aria-hidden="true" />
              <span class="sr-only">Square</span>
            </button>
            <button @click="selectedAction = 4" :class="[selectedAction == 4 ? 'bg-primary' : 'hover:bg-secondary-light', 'h-12 w-12 rounded-md shadow-md flex justify-center items-center']">
              <font-awesome-icon icon="circle" class="w-6 h-6 text-white" aria-hidden="true" />
              <span class="sr-only">Circle</span>
            </button>
          </div>
          <!-- <div class="p-2 gap-2 flex flex-row h-full">
            <button @click="preview = !preview" :class="[preview ? 'bg-primary' : 'bg-secondary hover:bg-primary', 'h-12 w-12 rounded-md shadow-md border border-secondary-light']"></button>
          </div> -->
        </div>
        <div class="m-4 flex gap-2 z-50 absolute">
          <button
            class="border border-secondary-light shadow-md bg-secondary px-4 py-1 rounded-full w-fit cursor-pointer"
            @click="fitCanvas">
            {{ defaultX - x }}, {{ defaultY - y }}
            {{ selectedObject > -1 ? 'object selected: ' + selectedObject : '' }}
          </button>
          <button
            class="border border-secondary-light shadow-md bg-secondary px-4 py-1 rounded-full w-fit cursor-pointer"
            v-if="zoom != defaultZoom" @click="zoom = defaultZoom">
            {{ Math.round((zoom / defaultZoom) * 100) }}%
            <span class="sr-only">Reset zoom</span>
          </button>
        </div>
        <div class="absolute inset-0 pointer-events-none" :style="{
          backgroundImage: 'radial-gradient(circle, rgba(255,255,255,0.05) 1px, transparent 1px)',
          backgroundSize: `${20 * zoom}px ${20 * zoom}px`,
          backgroundPosition: `${x}px ${y}px`,
        }"></div>

        <!-- canvas -->

        <div :class="[selectedAction == 0 ? '' : 'cursor-crosshair', 'canvas']" @mousedown="onCanvasMouseDown" :style="getCanvasStyle(x, y)">
          <div v-for="(obj, index) in image_config.layers" :key="index">
            <div @mouseover="onLayerMouseOver(index)" @mouseleave="onLayerMouseLeave()"
              :style="getObjectStyleBase(obj, index)">
              <div v-if="obj.type == CustomWelcomerImageLayerTypeText" :style="getObjectStyle(obj, index)"
                class="pointer-events-none"><span v-html="marked(obj.value, true)"></span></div>
              <img v-else-if="obj.type == CustomWelcomerImageLayerTypeImage" :style="getObjectStyle(obj, index)"
                class="pointer-events-none" :src="formatText(obj.value)" />
              <div
                v-else-if="obj.type == CustomWelcomerImageLayerTypeShapeRectangle || obj.type == CustomWelcomerImageLayerTypeShapeCircle"
                :style="getObjectStyle(obj, index)" class="pointer-events-none"></div>

              <div v-if="hoveredObject == index || selectedObject == index"
                class="outline outline-[#0078D7] outline-2 w-full h-full" @mousedown="onGrabStart(0)">
                <div v-if="selectedObject == index"
                  class="absolute -left-[11px] -top-[11px] w-2 h-2 m-2 bg-white border-2 border-[#0078D7] cursor-grab"
                  @mousedown="onGrabStart(1)"></div>
                <div v-if="selectedObject == index"
                  class="absolute -right-[11px] -top-[11px] w-2 h-2 m-2 bg-white border-2 border-[#0078D7] cursor-grab"
                  @mousedown="onGrabStart(2)"></div>
                <div v-if="selectedObject == index"
                  class="absolute -left-[11px] -bottom-[11px] w-2 h-2 m-2 bg-white border-2 border-[#0078D7] cursor-grab"
                  @mousedown="onGrabStart(3)"></div>
                <div v-if="selectedObject == index"
                  class="absolute -right-[11px] -bottom-[11px] w-2 h-2 m-2 bg-white border-2 border-[#0078D7] cursor-grab"
                  @mousedown="onGrabStart(4)"></div>
              </div>
            </div>
          </div>
        </div>
      </div>
      <div class="builder-sidebar">
        <div class="p-4 border-b border-secondary-light">
          <span class="font-semibold text-sm">Layers</span>
          <ul>
            <li v-for="(obj, index) in image_config.layers" :key="index" @click="onLayerClick(index)"
              @mouseover="onLayerMouseOver(index)" @mouseleave="onLayerMouseLeave()"
              :class="[selectedObject == index ? 'bg-primary text-white' : hoveredObject == index ? 'bg-secondary-light text-white' : 'hover:bg-secondary-light cursor-pointer', ' whitespace-nowrap overflow-ellipsis overflow-hidden px-2 py-1 rounded-md hover:cursor-pointer']">
              <span v-if="obj.type == CustomWelcomerImageLayerTypeText">
                <font-awesome-icon icon="text" class="mr-2" />
                {{ obj.value || 'Text ' + (index + 1) }}
              </span>
              <span v-else-if="obj.type == CustomWelcomerImageLayerTypeImage">
                <font-awesome-icon icon="image" class="mr-2" />
                {{ obj.value || 'Image ' + (index + 1) }}
              </span>
              <span v-else-if="obj.type == CustomWelcomerImageLayerTypeShapeRectangle">
                <font-awesome-icon icon="square" class="mr-2" />
                Rectangle {{ index + 1 }}
              </span>
              <span v-else-if="obj.type == CustomWelcomerImageLayerTypeShapeCircle">
                <font-awesome-icon icon="circle" class="mr-2" />
                Circle {{ index + 1 }}
              </span>
            </li>
          </ul>
        </div>
        <div class="overflow-y-auto flex-1 divide-secondary-light divide-y" v-if="selectedObject === -1">
          <!-- canvas controls -->
          <div class="p-4">
            <span class="font-semibold text-sm mb-2 block">Layout</span>
            <div class="grid grid-cols-2 gap-2">
              <InputCalculator :min="getMinimumWidth()" max="2000" type="number" v-model="image_config.dimensions[0]">
                Width
              </InputCalculator>
              <InputCalculator :min="getMinimumHeight()" max="2000" type="number" v-model="image_config.dimensions[1]">
                Height
              </InputCalculator>
            </div>
          </div>

          <div class="p-4">
            <span class="font-semibold text-sm mb-2 block">Fill</span>
            <ImageBuilderColourSelector v-model="image_config.fill" />
          </div>
          <div class="p-4">
            <span class="font-semibold text-sm mb-2 block">Stroke</span>
            <Listbox as="div" class="flex-1">
              <div class="relative">
                <ListboxButton
                  class="relative w-full py-2 pl-3 pr-10 text-left bg-white border border-gray-300 dark:bg-secondary dark:border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm">
                  <div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                    <font-awesome-icon icon="square" class="inline w-4 h-4 mr-1 border-primary" :style="{
                      color: `${image_config.stroke.color}`,
                    }" />
                  </div>
                  <span class="block pl-10 truncate">{{
                    image_config.stroke.color.toUpperCase()
                    }}</span>
                  <span class="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
                    <SelectorIcon class="w-5 h-5 text-gray-400" aria-hidden="true" />
                  </span>
                </ListboxButton>

                <transition leave-active-class="transition duration-100 ease-in" leave-from-class="opacity-100"
                  leave-to-class="opacity-0">
                  <ListboxOptions class="absolute z-10 mt-1">
                    <ColorPicker theme="dark" :color="image_config.stroke.color || '#000000'"
                      @changeColor="image_config.stroke.color = rgbaToHex($event)"
                      :sucker-hide="true" />
                  </ListboxOptions>
                </transition>
              </div>
            </Listbox>
            <div class="mt-2">
              <InputCalculator type="number" min="0" :max="Math.min(32, Math.min(image_config.dimensions[0]/2, image_config.dimensions[1]/2))" v-model="image_config.stroke.width">
                Width
              </InputCalculator>
            </div>
          </div>

          <div class="p-4">
            <span class="font-semibold text-sm mb-2 block">Debug</span>
            <code>
              {{ image_config }}
            </code>
          </div>
        </div>


        <div class="overflow-y-auto flex-1 divide-secondary-light divide-y" v-else>
          <!-- object controls -->
          <div class="p-4" v-if="image_config.layers[selectedObject].type == CustomWelcomerImageLayerTypeText">
            <span class="font-semibold text-sm mb-2 block">Text</span>
            <AutocompleteInput type="text" :isTextarea="true" class="border rounded p-2 bg-transparent w-full"
              placeholder="Message Content" rows="4" :value="image_config.layers[selectedObject].value"
              @update:modelValue="image_config.layers[selectedObject].value = $event" />
          </div>
          <div class="p-4" v-if="image_config.layers[selectedObject].type == CustomWelcomerImageLayerTypeImage">
            <span class="font-semibold text-sm mb-2 block">Image URL</span>
            <AutocompleteInput type="text" class="border rounded p-2 bg-transparent w-full" placeholder="Image URL"
              :value="image_config.layers[selectedObject].value"
              @update:modelValue="image_config.layers[selectedObject].value = $event" />
          </div>
          <div class="p-4">
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
          <div class="p-4">
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
          <div class="p-4">
            <span class="font-semibold text-sm mb-2 block">Appearance</span>
            <span class="block text-neutral-500 text-xs font-medium">Border Radius</span>
            <div class="grid grid-cols-2 gap-2 mt-1">
              <InputCalculator type="number" min="0" minPercentage="0" maxPercentage="100" forceStringOutput="true"
                v-model="image_config.layers[selectedObject].border_radius[0]">
                Top Left
              </InputCalculator>
              <InputCalculator type="number" min="0" minPercentage="0" maxPercentage="100" forceStringOutput="true"
                v-model="image_config.layers[selectedObject].border_radius[1]">
                Top Right
              </InputCalculator>
              <InputCalculator type="number" min="0" minPercentage="0" maxPercentage="100" forceStringOutput="true"
                v-model="image_config.layers[selectedObject].border_radius[3]">
                Bottom Left
              </InputCalculator>
              <InputCalculator type="number" min="0" minPercentage="0" maxPercentage="100" forceStringOutput="true"
                v-model="image_config.layers[selectedObject].border_radius[2]">
                Bottom Right
              </InputCalculator>
            </div>
          </div>
          <div class="p-4" v-if="image_config.layers[selectedObject].type == CustomWelcomerImageLayerTypeText">
            <span class="font-semibold text-sm mb-2 block">Typography</span>
            <Listbox as="div" :model="image_config.layers[selectedObject].typography.font_family"
              @update:modelValue="updateTypographyFontFamily($event)">
              <div class="relative">
                <ListboxButton
                  class="bg-secondary-dark relative w-full py-2 pl-3 pr-10 text-left border border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm">
                  {{ fonts[image_config.layers[selectedObject].typography.font_family]?.name || 'Select Font' }}
                </ListboxButton>
                <ListboxOptions
                  class="absolute z-20 w-full mt-1 overflow-auto text-base bg-white dark:bg-secondary-dark rounded-md shadow-sm max-h-60 ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                  <ListboxOption v-for="(font, fontKey) in fonts" :key="fontKey" :value="fontKey"
                    v-slot="{ active, selected }">
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
                  <ListboxButton
                    class="bg-secondary-dark relative w-full py-2 pl-3 pr-10 text-left border border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm">
                    {{
                      fonts[image_config.layers[selectedObject].typography.font_family]?.weights[image_config.layers[selectedObject].typography.font_weight]
                        ? image_config.layers[selectedObject].typography.font_weight :
                        (fonts[image_config.layers[selectedObject].typography.font_family] ?
                          fonts[image_config.layers[selectedObject].typography.font_family].default : 'normal') }}
                  </ListboxButton>
                  <ListboxOptions
                    class="absolute z-20 w-full mt-1 overflow-auto text-base bg-white dark:bg-secondary-dark rounded-md shadow-sm max-h-60 ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                    <ListboxOption
                      v-for="(weight, weightKey) in fonts[image_config.layers[selectedObject].typography.font_family]?.weights"
                      :key="weightKey" :value="weightKey" v-slot="{ active, selected }">
                      <li :class="[
                        active ? 'text-white bg-primary' : 'text-gray-900 dark:text-gray-50',
                        'cursor-default select-none relative py-2 pl-3 pr-9']">
                        {{ weightKey }}
                      </li>
                    </ListboxOption>
                  </ListboxOptions>
                </div>
              </Listbox>
              <InputCalculator type="number" min="8" v-model="image_config.layers[selectedObject].typography.font_size">
              </InputCalculator>
              <InputCalculator type="number" min="0.1" step="0.1"
                v-model="image_config.layers[selectedObject].typography.line_height">Line Height</InputCalculator>
              <InputCalculator type="number" step="0.1"
                v-model="image_config.layers[selectedObject].typography.letter_spacing">Letter Spacing</InputCalculator>
            </div>
            <span class="block text-neutral-500 text-xs font-medium mt-2">Alignment</span>
            <div class="grid grid-cols-2 gap-2 mt-1">
              <Listbox as="div" v-model="image_config.layers[selectedObject].typography.horizontal_alignment">
                <div class="relative">
                  <ListboxButton
                    class="bg-secondary-dark relative w-full py-2 pl-3 pr-10 text-left border border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm">
                    {{ image_config.layers[selectedObject].typography.horizontal_alignment || 'left' }}
                  </ListboxButton>
                  <ListboxOptions
                    class="absolute z-20 w-full mt-1 overflow-auto text-base bg-white dark:bg-secondary-dark rounded-md shadow-sm max-h-60 ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                    <ListboxOption value="left" v-slot="{ active }">
                      <li
                        :class="[active ? 'text-white bg-primary' : 'text-gray-900 dark:text-gray-50', 'cursor-default select-none relative py-2 pl-3 pr-9']">
                        left</li>
                    </ListboxOption>
                    <ListboxOption value="center" v-slot="{ active }">
                      <li
                        :class="[active ? 'text-white bg-primary' : 'text-gray-900 dark:text-gray-50', 'cursor-default select-none relative py-2 pl-3 pr-9']">
                        center</li>
                    </ListboxOption>
                    <ListboxOption value="right" v-slot="{ active }">
                      <li
                        :class="[active ? 'text-white bg-primary' : 'text-gray-900 dark:text-gray-50', 'cursor-default select-none relative py-2 pl-3 pr-9']">
                        right</li>
                    </ListboxOption>
                  </ListboxOptions>
                </div>
              </Listbox>
              <Listbox as="div" v-model="image_config.layers[selectedObject].typography.vertical_alignment">
                <div class="relative">
                  <ListboxButton
                    class="bg-secondary-dark relative w-full py-2 pl-3 pr-10 text-left border border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm">
                    {{ image_config.layers[selectedObject].typography.vertical_alignment || 'center' }}
                  </ListboxButton>
                  <ListboxOptions
                    class="absolute z-20 w-full mt-1 overflow-auto text-base bg-white dark:bg-secondary-dark rounded-md shadow-sm max-h-60 ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                    <ListboxOption value="start" v-slot="{ active }">
                      <li
                        :class="[active ? 'text-white bg-primary' : 'text-gray-900 dark:text-gray-50', 'cursor-default select-none relative py-2 pl-3 pr-9']">
                        start</li>
                    </ListboxOption>
                    <ListboxOption value="center" v-slot="{ active }">
                      <li
                        :class="[active ? 'text-white bg-primary' : 'text-gray-900 dark:text-gray-50', 'cursor-default select-none relative py-2 pl-3 pr-9']">
                        center</li>
                    </ListboxOption>
                    <ListboxOption value="end" v-slot="{ active }">
                      <li
                        :class="[active ? 'text-white bg-primary' : 'text-gray-900 dark:text-gray-50', 'cursor-default select-none relative py-2 pl-3 pr-9']">
                        end</li>
                    </ListboxOption>
                  </ListboxOptions>
                </div>
              </Listbox>
            </div>
          </div>
          <div class="p-4">
            <span class="font-semibold text-sm mb-2 block">Fill</span>
            <ImageBuilderColourSelector v-model="image_config.layers[selectedObject].fill" />
          </div>
          <div class="p-4">
            <span class="font-semibold text-sm mb-2 block">Stroke</span>
            <Listbox as="div" class="flex-1">
              <div class="relative">
                <ListboxButton
                  class="relative w-full py-2 pl-3 pr-10 text-left bg-white border border-gray-300 dark:bg-secondary dark:border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm">
                  <div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                    <font-awesome-icon icon="square" class="inline w-4 h-4 mr-1 border-primary" :style="{
                      color: `${image_config.layers[selectedObject].stroke.color}`,
                    }" />
                  </div>
                  <span class="block pl-10 truncate">{{
                    image_config.layers[selectedObject].stroke.color.toUpperCase()
                    }}</span>
                  <span class="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
                    <SelectorIcon class="w-5 h-5 text-gray-400" aria-hidden="true" />
                  </span>
                </ListboxButton>

                <transition leave-active-class="transition duration-100 ease-in" leave-from-class="opacity-100"
                  leave-to-class="opacity-0">
                  <ListboxOptions class="absolute z-10 mt-1">
                    <ColorPicker theme="dark" :color="image_config.layers[selectedObject].stroke.color || '#000000'"
                      @changeColor="image_config.layers[selectedObject].stroke.color = rgbaToHex($event)"
                      :sucker-hide="true" />
                  </ListboxOptions>
                </transition>
              </div>
            </Listbox>
            <div class="mt-2">
              <InputCalculator type="number" min="0" max="32" v-model="image_config.layers[selectedObject].stroke.width">
              </InputCalculator>
            </div>
          </div>
          <div class="p-4">
            <span class="font-semibold text-sm mb-2 block">Debug</span>
            <code>
              {{ image_config.layers[selectedObject] }}
            </code>
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
  @apply h-dvh;
}

.builder-portal {
  @apply h-dvh bg-secondary-light flex flex-row;
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
import ImageBuilderColourSelector from "@/components/dashboard/ImageBuilderColourSelector.vue";
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
  marked,
  formatText,
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
    ImageBuilderColourSelector,
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
    "image_config.stroke.width"(newValue, oldValue) {
      if (newValue !== oldValue) {
        this.fitCanvas();
      }
    },
    "image_config": {
      handler(newValue, oldValue) {
        // excludes initial assignment
        if (Object.keys(oldValue).length !== 0) {
          this.onValueUpdate();
        }
      },
      deep: true
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

    // 0: Select
    // 1: Text
    // 2: Image
    // 3: Square
    // 4: Circle
    let selectedAction = ref(0);
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

      selectedAction, selectedObject, hoveredObject, selectedGrab,

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
      if (document.activeElement && (document.activeElement.tagName == "INPUT" || document.activeElement.tagName == "TEXTAREA")) return;

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
          case "Delete":
          case "Backspace": {
            this.image_config.layers.splice(this.selectedObject, 1);
            this.selectedObject = -1;
            break;
          }
          case "PageUp": {
            // Move layer up
            if (this.selectedObject > 0) {
              const temp = this.image_config.layers[this.selectedObject];
              this.image_config.layers[this.selectedObject] = this.image_config.layers[this.selectedObject - 1];
              this.image_config.layers[this.selectedObject - 1] = temp;
              this.selectedObject--;
            }
            break;
          }
          case "PageDown": {
            // Move layer down
            if (this.selectedObject < this.image_config.layers.length - 1) {
              const temp = this.image_config.layers[this.selectedObject];
              this.image_config.layers[this.selectedObject] = this.image_config.layers[this.selectedObject + 1];
              this.image_config.layers[this.selectedObject + 1] = temp;
              this.selectedObject++;
            }
            break;
          }
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
      if (e.button == 0 && this.selectedAction == 0) {
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
    onCanvasMouseDown(event) {
      // handles drawing new layers. users can click which will create a 100x100 layer,
      // or click and drag to create a custom sized layer.

      if (this.selectedAction === 0) return; // only handle in draw modes

      const canvasRect = this.$el.querySelector('.builder-canvas').getBoundingClientRect();

      if (event.clientX < canvasRect.left || event.clientX > canvasRect.right ||
      event.clientY < canvasRect.top || event.clientY > canvasRect.bottom) return;

      const canvas = this.$el.querySelector('.canvas');
      const canvasRect2 = canvas.getBoundingClientRect();

      let startX = ((event.clientX - canvasRect2.left) / this.zoom) - 16;
      let startY = ((event.clientY - canvasRect2.top) / this.zoom) - 16;

      let isDrawing = false;
      let endX = startX;
      let endY = startY;
      let tempLayerIndex = -1;

      const onMouseMove = (moveEvent) => {
        isDrawing = true;
        endX = (moveEvent.clientX - canvasRect2.left) / this.zoom;
        endY = (moveEvent.clientY - canvasRect2.top) / this.zoom;

        const x = Math.round(Math.min(startX, endX));
        const y = Math.round(Math.min(startY, endY));
        const width = Math.round(Math.abs(endX - startX));
        const height = Math.round(Math.abs(endY - startY));

        // Create temp layer on first move
        if (tempLayerIndex === -1) {
          tempLayerIndex = this.createLayer(x, y, width, height);
        } else {
          // Update temp layer in real time
          const tempLayer = this.image_config.layers[tempLayerIndex];
          if (tempLayer) {
          tempLayer.position[0] = x;
          tempLayer.position[1] = y;
          tempLayer.dimensions[0] = Math.max(10, width);
          tempLayer.dimensions[1] = Math.max(10, height);
          }
        }
      };

      const onMouseUp = () => {
        window.removeEventListener('mousemove', onMouseMove);
        window.removeEventListener('mouseup', onMouseUp);

        if (!isDrawing && tempLayerIndex === -1) {
          // Click: create 100x100 at clicked position
          let x = Math.round(startX);
          let y = Math.round(startY);
          let width = this.selectedAction == 1 ? 200 : 100;
          let height = this.selectedAction == 1 ? 32 : 100;

          this.createLayer(x, y, width, height);
        } else if (tempLayerIndex !== -1) {
          // Select the newly created layer
          this.selectedObject = tempLayerIndex;
        }
      };

      window.addEventListener('mousemove', onMouseMove);
      window.addEventListener('mouseup', onMouseUp);
    },

    createLayer(x, y, width, height) {
      let newLayer = {
        type: this.selectedAction - 1,
        position: [x, y],
        dimensions: [Math.max(10, width), Math.max(10, height)],
        rotation: 0,
        inverted_x: false,
        inverted_y: false,
        border_radius: ["0", "0", "0", "0"],
        fill: this.selectedAction != 2 ? "#ffffff" : "#ffffff00", // transparent for images
        stroke: {
          color: "#ffffff00",
          width: 0,
        },
        value: "",
      };

      if (newLayer.type == CustomWelcomerImageLayerTypeText) {
        newLayer.typography = {
          font_family: "Inter",
          font_weight: "regular",
          font_size: 24,
          line_height: 1.2,
          letter_spacing: 0,
          horizontal_alignment: "left",
          vertical_alignment: "center",
        };
        newLayer.value = "New Text";
      }

      this.image_config.layers.unshift(newLayer); // add layer to top layer
      this.selectedAction = 0; // reset to select mode
      this.selectedObject = 0; // select new layer

      return 0;
    },

    getMinimumWidth() {
      let minWidth = 100;
      for (let layer of this.image_config.layers) {
        minWidth = Math.max(minWidth, layer.position[0] + layer.dimensions[0]);
      }
      return minWidth;
    },

    getMinimumHeight() {
      let minHeight = 100;
      for (let layer of this.image_config.layers) {
        minHeight = Math.max(minHeight, layer.position[1] + layer.dimensions[1]);
      }
      return minHeight;
    },

    marked(text, embed) {
      return marked(text, embed);
    },

    formatText(text) {
      return formatText(text);
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
        endpoints.EndpointGuildWelcomerBuilder(this.$store.getters.getSelectedGuildID),
        ({ config }) => {
          this.config = config;
          this.isDataFetched = true;
          this.isDataError = false;

          this.image_config = this.parseDict(config.custom_builder_data);

          this.$nextTick(() => {
            this.preemptivelyLoadFonts();
            this.fitCanvas();
          });
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
      this.isChangeInProgress = true;

      this.config.custom_builder_data = JSON.stringify(this.image_config);

      dashboardAPI.doPost(
        endpoints.EndpointGuildWelcomerBuilder(this.$store.getters.getSelectedGuildID),
        this.config,
        this.files,
        ({ config }) => {
          this.$store.dispatch("createToast", getSuccessToast());

          this.image_config = this.parseDict(config.custom_builder_data);

          this.$nextTick(() => {
            this.preemptivelyLoadFonts();
            this.fitCanvas();
          });

          this.files = [];
          
          this.$nextTick(() => {
            this.unsavedChanges = false;
            this.isChangeInProgress = false;
          });
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
        background: this.getFillAsCSS(this.image_config.fill || '#ffffff'),
        overflow: this.preview ? "hidden" : "visible",
        border: (this.image_config.stroke?.width > 0 ? this.image_config.stroke.width + "px solid " + this.getFillAsCSS(this.image_config.stroke.color) : "none")
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

      styles.background = (obj.type != CustomWelcomerImageLayerTypeText ? this.getFillAsCSS(obj.fill) : "transparent")
      styles.color = (obj.type == CustomWelcomerImageLayerTypeText ? this.getFillAsCSS(obj.fill) : "inherit")


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

        styles.fontSize = (obj.typography?.font_size && obj.typography.font_size != 0 ? obj.typography.font_size + "px" : "auto")
        styles.lineHeight = (obj.typography?.line_height && obj.typography.line_height != 0 ? obj.typography.line_height + "em" : "normal")
        styles.letterSpacing = (obj.typography?.letter_spacing && obj.typography.letter_spacing != 0 ? obj.typography.letter_spacing + "px" : "normal")
        // styles["-webkit-text-stroke"] = (obj.stroke?.width > 0 ? obj.stroke.width + "px " + this.getFillAsCSS(obj.stroke.color) : "none")
        styles.textShadow = (obj.stroke?.width > 0 ? this.generateTextShadow(obj.stroke.width, this.getFillAsCSS(obj.stroke.color)) : "none")

        styles.display = "flex"
        styles.justifyContent = this.normalizeHorizontalAlignment(obj.typography?.horizontal_alignment)
        styles.alignItems = this.normalizeVerticalAlignment(obj.typography?.vertical_alignment)
        styles.whiteSpace = "pre-wrap"
      } else {
        styles.border = (obj.stroke?.width > 0 ? obj.stroke.width + "px solid " + this.getFillAsCSS(obj.stroke.color) : "none")
      }

      return styles;
    },

    generateTextShadow(strokeWidth, strokeColor) {
      // approximate text stroke using multiple text shadows
      let shadows = [];
      let p = strokeWidth * strokeWidth;
      for (let dx = -strokeWidth; dx <= strokeWidth; dx++) {
        for (let dy = -strokeWidth; dy <= strokeWidth; dy++) {
          // round out stroke
          if (dx*dx+dy*dy <= p) {
            shadows.push(`${dx}px ${dy}px 0 ${strokeColor}`);
          }
        }
      }
      return shadows.join(", ");
    },

    onLayerClick(index) {
      if (this.selectedObject != index) this.selectedObject = index;
    },

    onLayerMouseOver(index) {
      if (this.selectedAction != 0) return; // only hover in select mode
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
          obj.position[0] += x * (1 / this.zoom);
          obj.position[1] += y * (1 / this.zoom);
          break;
        case 1: // top-left
          obj.position[0] += x * (1 / this.zoom);
          obj.dimensions[0] -= x * (1 / this.zoom);
          obj.position[1] += y * (1 / this.zoom);
          obj.dimensions[1] -= y * (1 / this.zoom);
          break;
        case 2: // top-right
          obj.dimensions[0] += x * (1 / this.zoom);
          obj.position[1] += y * (1 / this.zoom);
          obj.dimensions[1] -= y * (1 / this.zoom);
          break;
        case 3: // bottom-left
          obj.position[0] += x * (1 / this.zoom);
          obj.dimensions[0] -= x * (1 / this.zoom);
          obj.dimensions[1] += y * (1 / this.zoom);
          break;
        case 4: // bottom-right
          obj.dimensions[0] += x * (1 / this.zoom);
          obj.dimensions[1] += y * (1 / this.zoom);
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

        this.x = (container.offsetWidth - canvas.offsetWidth) / 2;
        this.y = 0;
        this.zoom = Math.min(1, (container.offsetWidth - (padding * 2)) / canvas.offsetWidth);

        this.defaultZoom = this.zoom;
        this.defaultX = this.x;
        this.defaultY = this.y;
      });
    },

    getFillAsCSS(value) {
      if (value.startsWith("#")) {
        return value;
      }

      if (value == "solid:profile") {
        return "linear-gradient(to right, #ffffff, #000000)";
      }

      return "url('https://beta.welcomer.gg/assets/backgrounds/" + value + ".webp') center / cover no-repeat";
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