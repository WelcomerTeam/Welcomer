<template>
  <div :class="[
    'sm:grid sm:grid-cols-3 sm:gap-4 dark:text-gray-50 border-gray-300 dark:border-secondary-light py-4 sm:py-6 items-start',
    $props.hideBorder ? '' : 'border-b',
  ]">
    <label class="block font-medium text-gray-700 dark:text-gray-50" :for="componentId" v-if="!$props.inlineFormValue">
      {{ title }}
      <div v-if="$props.inlineSlot">
        <div class="text-gray-600 dark:text-gray-400 text-sm col-span-3 mt-2 sm:mt-0">
          <slot></slot>
        </div>
      </div>
    </label>
    <div :class="['mt-1 sm:mt-0', $props.inlineFormValue ? 'sm:col-span-3' : 'sm:col-span-2']">
      <div v-if="type == FormTypeToggle">
        <Switch :id="componentId" :modelValue="modelValue" @update:modelValue="updateValue($event)" @blur="blur"
          :disabled="$props.disabled" :class="[
            $props.validation?.$invalid
              ? 'ring-red-500 border-red-500 dark:ring-red-500 dark:border-red-500'
              : '',
            $props.disabled
              ? 'bg-gray-100 dark:bg-secondary-light text-neutral-500'
              : modelValue
                ? 'bg-green-500 focus:ring-green-500'
                : 'bg-gray-400 focus:ring-gray-400',
            'relative inline-flex flex-shrink-0 h-6 w-11 border-2 border-transparent rounded-full cursor-pointer transition-colors ease-in-out duration-200 focus:outline-none focus:ring-2 focus:ring-offset-2',
          ]">
          <span :class="[
            modelValue ? 'translate-x-5' : 'translate-x-0',
            'pointer-events-none relative inline-block h-5 w-5 rounded-full bg-white shadow transform ring-0 transition ease-in-out duration-200',
          ]">
            <span :class="[
              modelValue
                ? 'opacity-0 ease-out duration-100'
                : 'opacity-100 ease-in duration-200',
              'absolute inset-0 h-full w-full flex items-center justify-center transition-opacity',
            ]" aria-hidden="true">
              <svg class="w-3 h-3 text-gray-400" fill="none" viewBox="0 0 12 12">
                <path d="M4 8l2-2m0 0l2-2M6 6L4 4m2 2l2 2" stroke="currentColor" stroke-width="2" stroke-linecap="round"
                  stroke-linejoin="round" />
              </svg>
            </span>
            <span :class="[
              modelValue
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
        <div v-if="$props.validation?.$invalid" class="errors">
          <span v-bind:key="index" v-for="(message, index) in $props.validation?.$errors">{{ message.$message }}</span>
        </div>
      </div>

      <div v-else-if="type == FormTypeChannelList">
        <Listbox :id="componentId" as="div" :modelValue="modelValue" @update:modelValue="updateValue($event)" @blur="blur"
          :disabled="$props.disabled">
          <div class="relative">
            <ListboxButton :disabled="$props.disabled" :class="[
              $props.validation?.$invalid
                ? 'border-red-500 ring-red-500 dark:ring-red-500 dark:border-red-500'
                : '',
              $props.disabled
                ? 'bg-gray-100 dark:bg-secondary-light text-neutral-500'
                : 'bg-white dark:bg-secondary-dark',
              'relative w-full py-2 pl-3 pr-10 text-left border border-gray-300 dark:border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm',
            ]">
              <div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                <font-awesome-icon icon="hashtag" class="w-5 h-5 text-gray-400" aria-hidden="true" />
              </div>
              <div v-if="$store.getters.isLoadingGuild"
                class="block ml-10 h-6 sm:h-5 animate-pulse bg-gray-200 w-48 rounded-md"></div>
              <span v-else class="block pl-10 truncate">{{
                modelValue == null
                ? "No channel selected"
                : $store.getters.getGuildChannelById(modelValue)?.name ||
                `Unknown channel ${modelValue}`
              }}</span>
              <span class="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
                <SelectorIcon class="w-5 h-5 text-gray-400" aria-hidden="true" />
              </span>
            </ListboxButton>

            <transition leave-active-class="transition duration-100 ease-in" leave-from-class="opacity-100"
              leave-to-class="opacity-0">
              <ListboxOptions
                class="absolute z-20 w-full mt-1 overflow-auto text-base bg-white dark:bg-secondary-dark rounded-md shadow-sm max-h-60 ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                <div v-if="$store.getters.isLoadingGuild" class="flex py-5 w-full justify-center">
                  <LoadingIcon />
                </div>
                <div v-else>
                  <ListboxOption as="template" v-slot="{ active, selected }" v-if="nullable" :value="null">
                    <li :class="[
                      active
                        ? 'text-white bg-primary'
                        : 'text-gray-900 dark:text-gray-50',
                      'cursor-default select-none relative py-2 pl-3 pr-9',
                    ]">
                      <span :class="[
                        selected ? 'font-semibold' : 'font-normal',
                        'block truncate',
                      ]">
                        Unselect
                      </span>

                      <span v-if="selected" :class="[
                        active ? 'text-white' : 'text-primary',
                        'absolute inset-y-0 right-0 flex items-center pr-4',
                      ]">
                        <CheckIcon class="w-5 h-5" aria-hidden="true" />
                      </span>
                    </li>
                  </ListboxOption>
                  <ListboxOption as="template" v-for="channel in this.filterChannels(
                    $store.getters.getGuildChannels
                  )" :key="channel.id" :value="channel.id" v-slot="{ active, selected }">
                    <li :class="[
                      active
                        ? 'text-white bg-primary'
                        : 'text-gray-900 dark:text-gray-50',
                      'cursor-default select-none relative py-2 pl-3 pr-9',
                    ]">
                      <span :class="[
                        selected ? 'font-semibold' : 'font-normal',
                        'block truncate',
                      ]">
                        <font-awesome-icon icon="hashtag" :class="[
                          active ? 'text-white' : 'text-gray-400',
                          'inline w-4 h-4 mr-1',
                        ]" aria-hidden="true" />
                        {{ channel.name }}
                      </span>

                      <span v-if="selected" :class="[
                        active ? 'text-white' : 'text-primary',
                        'absolute inset-y-0 right-0 flex items-center pr-4',
                      ]">
                        <CheckIcon class="w-5 h-5" aria-hidden="true" />
                      </span>
                    </li>
                  </ListboxOption>
                </div>
              </ListboxOptions>
            </transition>
          </div>
        </Listbox>
        <div v-if="$props.validation?.$invalid" class="errors">
          <span v-bind:key="index" v-for="(message, index) in $props.validation?.$errors">{{ message.$message }}</span>
        </div>
      </div>

      <div v-else-if="type == FormTypeChannelListCategories">
        <Listbox :id="componentId" as="div" :disabled="$props.disabled" :modelValue="modelValue"
          @update:modelValue="updateValue($event)" @blur="blur">
          <div class="relative">
            <ListboxButton :disabled="$props.disabled" :class="[
              $props.validation?.$invalid
                ? 'ring-red-500 border-red-500 dark:ring-red-500 dark:border-red-500'
                : '',
              $props.disabled
                ? 'bg-gray-100 dark:bg-secondary-light text-neutral-500'
                : 'bg-white dark:bg-secondary-dark',
              'relative w-full py-2 pl-3 pr-10 text-left border border-gray-300 dark:border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm',
            ]">
              <div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                <font-awesome-icon icon="hashtag" class="w-5 h-5 text-gray-400" aria-hidden="true" />
              </div>
              <div v-if="$store.getters.isLoadingGuild"
                class="block ml-10 h-6 sm:h-5 animate-pulse bg-gray-200 w-48 rounded-md"></div>
              <span v-else class="block pl-10 truncate">{{
                modelValue == null
                ? "No channel selected"
                : $store.getters.getGuildChannelById(modelValue)?.name ||
                `Unknown channel ${modelValue}`
              }}</span>
              <span class="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
                <SelectorIcon class="w-5 h-5 text-gray-400" aria-hidden="true" />
              </span>
            </ListboxButton>

            <transition leave-active-class="transition duration-100 ease-in" leave-from-class="opacity-100"
              leave-to-class="opacity-0">
              <ListboxOptions
                class="absolute z-20 w-full mt-1 overflow-auto text-base bg-white dark:bg-secondary-dark rounded-md shadow-sm max-h-60 ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                <div v-if="$store.getters.isLoadingGuild" class="flex py-5 w-full justify-center">
                  <LoadingIcon />
                </div>
                <div v-else>
                  <ListboxOption as="template" v-slot="{ active, selected }" v-if="nullable" :value="null">
                    <li :class="[
                      active
                        ? 'text-white bg-primary'
                        : 'text-gray-900 dark:text-gray-50',
                      'cursor-default select-none relative py-2 pl-3 pr-9',
                    ]">
                      <span :class="[
                        selected ? 'font-semibold' : 'font-normal',
                        'block truncate',
                      ]">
                        Unselect
                      </span>

                      <span v-if="selected" :class="[
                        active ? 'text-white' : 'text-primary',
                        'absolute inset-y-0 right-0 flex items-center pr-4',
                      ]">
                        <CheckIcon class="w-5 h-5" aria-hidden="true" />
                      </span>
                    </li>
                  </ListboxOption>
                  <div v-for="category in $store.getters.getPackedGuildChannels" :key="category">
                    <div class="py-3" v-if="category.name && category.channels.length !== 0">
                      <span class="pl-2 text-xs font-bold uppercase">{{
                        category.name
                      }}</span>
                    </div>
                    <ListboxOption as="template" v-for="channel in category.channels" :key="channel.id"
                      :value="channel.id" v-slot="{ active, selected }">
                      <li :class="[
                        active
                          ? 'text-white bg-primary'
                          : 'text-gray-900 dark:text-gray-50',
                        'cursor-default select-none relative py-2 pl-3 pr-9',
                      ]">
                        <span :class="[
                          selected ? 'font-semibold' : 'font-normal',
                          'block truncate',
                        ]">
                          <font-awesome-icon icon="hashtag" :class="[
                            active ? 'text-white' : 'text-gray-400',
                            'inline w-4 h-4 mr-1',
                          ]" />
                          {{ channel.name }}
                        </span>

                        <span v-if="selected" :class="[
                          active ? 'text-white' : 'text-primary',
                          'absolute inset-y-0 right-0 flex items-center pr-4',
                        ]">
                          <CheckIcon class="w-5 h-5" aria-hidden="true" />
                        </span>
                      </li>
                    </ListboxOption>
                  </div>
                </div>
              </ListboxOptions>
            </transition>
          </div>
        </Listbox>
        <div v-if="$props.validation?.$invalid" class="errors">
          <span v-bind:key="index" v-for="(message, index) in $props.validation?.$errors">{{ message.$message }}</span>
        </div>
      </div>

      <div v-else-if="type == FormTypeRoleList">
        <Listbox :id="componentId" as="div" :modelValue="modelValue" @update:modelValue="updateValue($event)" @blur="blur"
          :disabled="$props.disabled">
          <div class="relative">
            <ListboxButton :disabled="$props.disabled" :class="[
              $props.validation?.$invalid
                ? 'ring-red-500 border-red-500 dark:ring-red-500 dark:border-red-500'
                : '',
              $props.disabled
                ? 'bg-gray-100 dark:bg-secondary-light text-neutral-500'
                : 'bg-white dark:bg-secondary-dark',
              'relative w-full py-2 pl-3 pr-10 text-left border border-gray-300 dark:border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm',
            ]">
              <div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                <font-awesome-icon icon="at" class="w-5 h-5 text-gray-400" aria-hidden="true" />
              </div>
              <div v-if="$store.getters.isLoadingGuild"
                class="block ml-10 h-6 sm:h-5 animate-pulse bg-gray-200 w-48 rounded-md"></div>
              <span v-else class="block pl-10 truncate">{{
                modelValue == null
                ? "No role selected"
                : $store.getters.getGuildRoleById(modelValue)?.name ||
                `Unknown role ${modelValue}`
              }}</span>
              <span class="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
                <SelectorIcon class="w-5 h-5 text-gray-400" aria-hidden="true" />
              </span>
            </ListboxButton>

            <transition leave-active-class="transition duration-100 ease-in" leave-from-class="opacity-100"
              leave-to-class="opacity-0">
              <ListboxOptions
                class="absolute z-20 w-full mt-1 overflow-auto text-base bg-white dark:bg-secondary-dark rounded-md shadow-sm max-h-60 ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                <div v-if="$store.getters.isLoadingGuild" class="flex py-5 w-full justify-center">
                  <LoadingIcon />
                </div>
                <div v-else>
                  <ListboxOption as="template" v-slot="{ active, selected }" v-if="nullable" :value="null">
                    <li :class="[
                      active
                        ? 'text-white bg-primary'
                        : 'text-gray-900 dark:text-gray-50',
                      'cursor-default select-none relative py-2 pl-3 pr-9',
                    ]">
                      <span :class="[
                        selected ? 'font-semibold' : 'font-normal',
                        'block truncate',
                      ]">
                        Unselect
                      </span>

                      <span v-if="selected" :class="[
                        active ? 'text-white' : 'text-primary',
                        'absolute inset-y-0 right-0 flex items-center pr-4',
                      ]">
                        <CheckIcon class="w-5 h-5" aria-hidden="true" />
                      </span>
                    </li>
                  </ListboxOption>
                  <ListboxOption as="template" v-for="role in $store.getters.getGuildRoles" :key="role.id"
                    :value="role.id" v-slot="{ active, selected }" :disabled="!role.is_assignable">
                    <li :class="[
                      role.is_assignable
                        ? ''
                        : 'bg-gray-200 dark:bg-secondary-light',
                      active
                        ? 'text-white bg-primary'
                        : 'text-gray-900 dark:text-gray-50',
                      'cursor-default select-none relative py-2 pl-3 pr-9',
                    ]">
                      <span :class="[
                        selected ? 'font-semibold' : 'font-normal',
                        'block truncate',
                      ]">
                        <font-awesome-icon icon="circle" :class="[
                          active ? 'text-white' : 'text-gray-400',
                          'inline w-4 h-4 mr-1 border-primary',
                        ]" :style="{ color: `${getHexColor(role?.color)}` }" />
                        {{ role.name }}
                      </span>

                      <span v-if="selected" :class="[
                        active ? 'text-white' : 'text-primary',
                        'absolute inset-y-0 right-0 flex items-center pr-4',
                      ]">
                        <CheckIcon class="w-5 h-5" aria-hidden="true" />
                      </span>
                    </li>
                  </ListboxOption>
                </div>
              </ListboxOptions>
            </transition>
          </div>
        </Listbox>
        <div v-if="$props.validation?.$invalid" class="errors">
          <span v-bind:key="index" v-for="(message, index) in $props.validation?.$errors">{{ message.$message }}</span>
        </div>
      </div>

      <div v-else-if="type == FormTypeMemberList">
        <Listbox :id="componentId" as="div" :modelValue="modelValue" @update:modelValue="updateValue($event)" @blur="blur"
          :disabled="$props.disabled">
          <div class="relative">
            <ListboxButton :disabled="$props.disabled" :class="[
              $props.validation?.$invalid
                ? 'ring-red-500 border-red-500 dark:ring-red-500 dark:border-red-500'
                : '',
              $props.disabled
                ? 'bg-gray-100 dark:bg-secondary-light text-neutral-500'
                : 'bg-white dark:bg-secondary-dark',
              'relative w-full py-2 pl-3 pr-10 text-left border border-gray-300 dark:border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm',
            ]">
              <div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                <font-awesome-icon icon="user" class="w-5 h-5 text-gray-400" aria-hidden="true" />
              </div>
              <div v-if="$store.getters.isLoadingGuild"
                class="block ml-10 h-6 sm:h-5 animate-pulse bg-gray-200 w-48 rounded-md"></div>
              <span v-else class="block pl-10 truncate">
                {{
                  modelValue == null
                  ? "No member selected"
                  : $store.getters.getGuildMemberById(modelValue)
                    ?.display_name || `Unknown member ${modelValue}`
                }}
              </span>
              <span class="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
                <SelectorIcon class="w-5 h-5 text-gray-400" aria-hidden="true" />
              </span>
            </ListboxButton>

            <transition leave-active-class="transition duration-100 ease-in" leave-from-class="opacity-100"
              leave-to-class="opacity-0">
              <ListboxOptions
                class="absolute z-20 w-full mt-1 text-base bg-white dark:bg-secondary-dark rounded-md shadow-sm ring-1 ring-primary ring-opacity-5 focus:outline-none sm:text-sm">
                <div v-if="$store.getters.isLoadingGuild" class="flex py-5 w-full justify-center">
                  <LoadingIcon />
                </div>
                <div v-else>
                  <ListboxOption as="template" v-slot="{ active, selected }" v-if="nullable" :value="null">
                    <li :class="[
                      active
                        ? 'text-white bg-primary'
                        : 'text-gray-900 dark:text-gray-50',
                      'cursor-default select-none relative py-2 pl-3 pr-9',
                    ]">
                      <span :class="[
                        selected ? 'font-semibold' : 'font-normal',
                        'block truncate',
                      ]">
                        Unselect
                      </span>

                      <span v-if="selected" :class="[
                        active ? 'text-white' : 'text-primary',
                        'absolute inset-y-0 right-0 flex items-center pr-4',
                      ]">
                        <CheckIcon class="w-5 h-5" aria-hidden="true" />
                      </span>
                    </li>
                  </ListboxOption>
                  <div class="w-full p-2">
                    <AutocompleteInput type="text"
                      class="w-full border-gray-300 dark:border-secondary-light bg-white dark:bg-secondary-dark rounded-md sm:text-sm"
                      placeholder="Start typing a name or user id..." v-model="query" @input="onQueryChange()" />
                  </div>
                  <div class="overflow-auto max-h-60">
                    <ListboxOption as="template" :key="this.query" :value="this.query" v-slot="{ active, selected }"
                      v-if="this.isValidSnowflake">
                      <li :class="[
                        active
                          ? 'text-white bg-primary'
                          : 'text-gray-900 dark:text-gray-50',
                        'cursor-default select-none relative py-2 pl-3 pr-9',
                      ]">
                        <span :class="[
                          selected ? 'font-semibold' : 'font-normal',
                          'block truncate',
                        ]">
                          Use Id "{{ this.query }}"
                        </span>
                        <span v-if="selected" :class="[
                          active ? 'text-white' : 'text-primary',
                          'absolute inset-y-0 right-0 flex items-center pr-4',
                        ]">
                          <CheckIcon class="w-5 h-5" aria-hidden="true" />
                        </span>
                      </li>
                    </ListboxOption>
                    <ListboxOption as="template" v-for="member in $store.getters.getGuildMemberResults" :key="member.id"
                      :value="member.id" v-slot="{ active, selected }">
                      <li :class="[
                        active
                          ? 'text-white bg-primary'
                          : 'text-gray-900 dark:text-gray-50',
                        'cursor-default select-none relative py-2 pl-3 pr-9',
                      ]">
                        <span :class="[
                          selected ? 'font-semibold' : 'font-normal',
                          'block truncate',
                        ]">
                          <img :alt="`Member ${member.display_name}`" v-lazy="`https://cdn.discordapp.com/avatars/${member.id}/${member.avatar}.webp?size=32`
                            " class="flex-shrink-0 inline w-4 h-4 mr-1 rounded-full object-contain" />
                          {{ member.display_name }}
                        </span>

                        <span v-if="selected" :class="[
                          active ? 'text-white' : 'text-primary',
                          'absolute inset-y-0 right-0 flex items-center pr-4',
                        ]">
                          <CheckIcon class="w-5 h-5" aria-hidden="true" />
                        </span>
                      </li>
                    </ListboxOption>
                  </div>
                </div>
              </ListboxOptions>
            </transition>
          </div>
        </Listbox>
        <div v-if="$props.validation?.$invalid" class="errors">
          <span v-bind:key="index" v-for="(message, index) in $props.validation?.$errors">{{ message.$message }}</span>
        </div>
      </div>

      <div v-else-if="type == FormTypeEmojiList">
        <Listbox :id="componentId" as="div" :modelValue="modelValue" @update:modelValue="updateValue($event)" @blur="blur"
          :disabled="$props.disabled">
          <div class="relative">
            <ListboxButton :disabled="$props.disabled" :class="[
              $props.validation?.$invalid
                ? 'ring-red-500 border-red-500 dark:ring-red-500 dark:border-red-500'
                : '',
              $props.disabled
                ? 'bg-gray-100 dark:bg-secondary-light text-neutral-500'
                : 'bg-white dark:bg-secondary-dark',
              'relative w-full py-2 pl-3 pr-10 text-left border border-gray-300 dark:border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm',
            ]">
              <div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                <img v-if="modelValue != undefined" :src="`https://cdn.discordapp.com/emojis/${modelValue}.webp`"
                  class="w-5 h-5 object-contain" />
                <font-awesome-icon v-else icon="face-laugh" class="w-5 h-5 text-gray-400" aria-hidden="true" />
              </div>
              <span class="block pl-10 truncate">{{
                modelValue == null
                ? "No emoji selected"
                : $store.getters.getGuildEmojiById(modelValue)?.name ||
                `Unknown emoji ${modelValue}`
              }}</span>
              <span class="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
                <SelectorIcon class="w-5 h-5 text-gray-400" aria-hidden="true" />
              </span>
            </ListboxButton>

            <transition leave-active-class="transition duration-100 ease-in" leave-from-class="opacity-100"
              leave-to-class="opacity-0">
              <ListboxOptions
                class="absolute z-20 w-full mt-1 text-base bg-white dark:bg-secondary-dark rounded-md shadow-sm ring-1 ring-primary ring-opacity-5 focus:outline-none sm:text-sm">
                <div class="w-full p-2">
                  <AutocompleteInput type="text"
                    class="w-full border-gray-300 dark:border-secondary-light bg-white dark:bg-secondary-dark rounded-md sm:text-sm"
                    placeholder="Start typing a name or emoji id..." />
                </div>
                <div class="overflow-auto max-h-60">
                  <ListboxOption as="template" :value="null" v-slot="{ active, selected }" v-if="nullable">
                    <li :class="[
                      active
                        ? 'text-white bg-primary'
                        : 'text-gray-900 dark:text-gray-50',
                      'cursor-default select-none relative py-2 pl-3 pr-9',
                    ]">
                      <span :class="[
                        selected ? 'font-semibold' : 'font-normal',
                        'block truncate',
                      ]">
                        Unselect
                      </span>

                      <span v-if="selected" :class="[
                        active ? 'text-white' : 'text-primary',
                        'absolute inset-y-0 right-0 flex items-center pr-4',
                      ]">
                        <CheckIcon class="w-5 h-5" aria-hidden="true" />
                      </span>
                    </li>
                  </ListboxOption>
                  <ListboxOption as="template" v-for="emoji in $store.getters.getGuildEmojis" :key="emoji.id"
                    :value="emoji.id" v-slot="{ active, selected }">
                    <li :class="[
                      active
                        ? 'text-white bg-primary'
                        : 'text-gray-900 dark:text-gray-50',
                      'cursor-default select-none relative py-2 pl-3 pr-9',
                    ]">
                      <span :class="[
                        selected ? 'font-semibold' : 'font-normal',
                        'block truncate',
                      ]">
                        <img :alt="`Emoji ${emoji.name}`" v-lazy="`https://cdn.discordapp.com/emojis/${emoji.id}.${emoji.is_animated ? 'gif' : 'png'
                          }`
                          " class="flex-shrink-0 inline w-4 h-4 mr-1 object-contain" />
                        {{ emoji.name }}
                      </span>

                      <span v-if="selected" :class="[
                        active ? 'text-white' : 'text-primary',
                        'absolute inset-y-0 right-0 flex items-center pr-4',
                      ]">
                        <CheckIcon class="w-5 h-5" aria-hidden="true" />
                      </span>
                    </li>
                  </ListboxOption>
                </div>
              </ListboxOptions>
            </transition>
          </div>
        </Listbox>
        <div v-if="$props.validation?.$invalid" class="errors">
          <span v-bind:key="index" v-for="(message, index) in $props.validation?.$errors">{{ message.$message }}</span>
        </div>
      </div>

      <div v-else-if="type == FormTypeGuildList">
        <Listbox :id="componentId" as="div" :modelValue="modelValue" @update:modelValue="updateValue($event)" @blur="blur"
          :disabled="$props.disabled">
          <div class="relative">
            <ListboxButton :disabled="$props.disabled" :class="[
              $props.validation?.$invalid
                ? 'ring-red-500 border-red-500 dark:ring-red-500 dark:border-red-500'
                : '',
              $props.disabled
                ? 'bg-gray-100 dark:bg-secondary-light text-neutral-500'
                : 'bg-white dark:bg-secondary-dark',
              'relative w-full py-2 pl-3 pr-10 text-left border border-gray-300 dark:border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm',
            ]">
              <div v-if="$store.getters.isLoadingGuilds"
                class="block h-6 sm:h-5 animate-pulse bg-gray-200 w-48 rounded-md"></div>
              <span v-else class="block truncate">{{
                modelValue == null
                ? "No guild selected"
                : $store.getters.getGuildById(modelValue)?.name ||
                `Unknown guild ${modelValue}`
              }}</span>
              <span class="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
                <SelectorIcon class="w-5 h-5 text-gray-400" aria-hidden="true" />
              </span>
            </ListboxButton>

            <transition leave-active-class="transition duration-100 ease-in" leave-from-class="opacity-100"
              leave-to-class="opacity-0">
              <ListboxOptions
                class="absolute z-20 w-full mt-1 overflow-auto text-base bg-white dark:bg-secondary-dark rounded-md shadow-sm max-h-60 ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
                <div v-if="$store.getters.isLoadingGuilds" class="flex py-5 w-full justify-center">
                  <LoadingIcon />
                </div>
                <div v-else>
                  <ListboxOption as="template" v-slot="{ active, selected }" v-if="nullable" :value="null">
                    <li :class="[
                      active
                        ? 'text-white bg-primary'
                        : 'text-gray-900 dark:text-gray-50',
                      'cursor-default select-none relative py-2 pl-3 pr-9',
                    ]">
                      <span :class="[
                        selected ? 'font-semibold' : 'font-normal',
                        'block truncate',
                      ]">
                        Unselect
                      </span>

                      <span v-if="selected" :class="[
                        active ? 'text-white' : 'text-primary',
                        'absolute inset-y-0 right-0 flex items-center pr-4',
                      ]">
                        <CheckIcon class="w-5 h-5" aria-hidden="true" />
                      </span>
                    </li>
                  </ListboxOption>
                  <ListboxOption as="template" v-for="guild in $store.getters.getGuilds" :key="guild.id"
                    :value="guild.id" v-slot="{ active, selected }" :disabled="!guild.has_welcomer">
                    <li :class="[
                      guild.has_welcomer
                        ? ''
                        : 'bg-gray-200 dark:bg-secondary-light',
                      active
                        ? 'text-white bg-primary'
                        : 'text-gray-900 dark:text-gray-50',
                      'cursor-default select-none relative py-2 pl-3 pr-9',
                    ]">
                      <span :class="[
                        selected ? 'font-semibold' : 'font-normal',
                        'block truncate',
                      ]">
                        <img :alt="`Guild ${guild.name}`" v-lazy="`https://cdn.discordapp.com/icons/${guild.id}/${guild.icon}.webp?size=32`"
                          class="flex-shrink-0 inline w-4 h-4 mr-1 rounded-full object-contain" v-if="guild.icon != ''" />
                        {{ guild.name }}
                      </span>

                      <span v-if="selected" :class="[
                        active ? 'text-white' : 'text-primary',
                        'absolute inset-y-0 right-0 flex items-center pr-4',
                      ]">
                        <CheckIcon class="w-5 h-5" aria-hidden="true" />
                      </span>
                    </li>
                  </ListboxOption>
                </div>
              </ListboxOptions>
            </transition>
          </div>
        </Listbox>
        <div v-if="$props.validation?.$invalid" class="errors">
          <span v-bind:key="index" v-for="(message, index) in $props.validation?.$errors">{{ message.$message }}</span>
        </div>
      </div>

      <div v-else-if="type == FormTypeColour">
        <Listbox :id="componentId" as="div" :modelValue="modelValue" @update:modelValue="updateValue($event)" @blur="blur"
          :disabled="$props.disabled">
          <div class="relative">
            <ListboxButton :disabled="$props.disabled" :class="[
              $props.validation?.$invalid
                ? 'ring-red-500 border-red-500 dark:ring-red-500 dark:border-red-500'
                : '',
              $props.disabled
                ? 'bg-gray-100 dark:bg-secondary-light text-neutral-500'
                : 'bg-white dark:bg-secondary-dark',
              'relative w-full py-2 pl-3 pr-10 text-left border border-gray-300 dark:border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm',
            ]">
              <div class="absolute inset-y-0 left-0 flex items-center pl-3 pointer-events-none">
                <font-awesome-icon icon="square" class="inline w-4 h-4 mr-1 border-primary"
                  :style="{ color: `${parseCSSValue(modelValue)}` }" />
              </div>
              <span class="block pl-10 truncate">{{
                parseCSSValue(modelValue)
              }}</span>
              <span class="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
                <SelectorIcon class="w-5 h-5 text-gray-400" aria-hidden="true" />
              </span>
            </ListboxButton>

            <transition leave-active-class="transition duration-100 ease-in" leave-from-class="opacity-100"
              leave-to-class="opacity-0">
              <ListboxOptions class="absolute z-10 mt-1">
                <ColorPicker theme="dark" :color="parseCSSValue(modelValue)" @changeColor="onColorChange"
                  :sucker-hide="true" />
              </ListboxOptions>
            </transition>
          </div>
        </Listbox>
      </div>

      <div v-else-if="type == FormTypeText">
        <AutocompleteInput :id="componentId" type="text" :class="[
          $props.validation?.$invalid
            ? 'ring-red-500 border-red-500 dark:ring-red-500 dark:border-red-500'
            : '',
          $props.disabled
            ? 'bg-gray-100 dark:bg-secondary-light text-neutral-500'
            : 'bg-white dark:bg-secondary-dark',
          'flex-1 shadow-sm block w-full min-w-0 border-gray-300 dark:border-secondary-light rounded-md focus:ring-primary focus:border-primary sm:text-sm',
        ]" :disabled="$props.disabled" placeholder="Enter text here..." :value="modelValue"
          :maxlength="$props.maxLength" @input="updateValue($event)" @blur="blur" />
        <div v-if="$props.validation?.$invalid" class="errors">
          <span v-bind:key="index" v-for="(message, index) in $props.validation?.$errors">{{ message.$message }}</span>
        </div>
      </div>

      <div v-else-if="type == FormTypeNumber">
        <input :id="componentId" type="number" :class="[
          $props.validation?.$invalid
            ? 'ring-red-500 border-red-500 dark:ring-red-500 dark:border-red-500'
            : '',
          $props.disabled
            ? 'bg-gray-100 dark:bg-secondary-light text-neutral-500'
            : 'bg-white dark:bg-secondary-dark',
          'flex-1 shadow-sm block w-full min-w-0 border-gray-300 dark:border-secondary-light rounded-md focus:ring-primary focus:border-primary sm:text-sm',
        ]" :disabled="$props.disabled" :value="modelValue" @input="updateValue($event.target.value)" @blur="blur" />
        <div v-if="$props.validation?.$invalid" class="errors">
          <span v-bind:key="index" v-for="(message, index) in $props.validation?.$errors">{{ message.$message }}</span>
        </div>
      </div>

      <div v-else-if="type == FormTypeTextArea">
        <AutocompleteInput type="text" :isTextarea="true" :id="componentId" :class="[
          $props.validation?.$invalid
            ? 'ring-red-500 border-red-500 dark:ring-red-500 dark:border-red-500'
            : '',
          $props.disabled
            ? 'bg-gray-100 dark:bg-secondary-light text-neutral-500'
            : 'bg-white dark:bg-secondary-dark',
          'flex-1 shadow-sm block w-full min-w-0 border-gray-300 dark:border-secondary-light rounded-md focus:ring-primary focus:border-primary sm:text-sm',
        ]" rows="4" :disabled="$props.disabled" placeholder="Enter text here..." :value="modelValue"
          @input="updateValue($event)" @blur="blur" />
        <div v-if="$props.validation?.$invalid" class="errors">
          <span v-bind:key="index" v-for="(message, index) in $props.validation?.$errors">{{ message.$message }}</span>
        </div>
      </div>

      <div v-else-if="type == FormTypeDropdown">
        <Listbox :id="componentId" as="div" :disabled="$props.disabled" :modelValue="modelValue"
          @update:modelValue="updateValue($event)" @blur="blur">
          <div class="relative">
            <ListboxButton :disabled="$props.disabled" :class="[
              $props.validation?.$invalid
                ? 'ring-red-500 border-red-500 dark:ring-red-500 dark:border-red-500'
                : '',
              $props.disabled
                ? 'bg-gray-100 dark:bg-secondary-light text-neutral-500'
                : 'bg-white dark:bg-secondary-dark',
              'relative w-full py-2 pl-3 pr-10 text-left border border-gray-300 dark:border-secondary-light rounded-md shadow-sm cursor-default focus:outline-none focus:ring-1 focus:ring-primary focus:border-primary sm:text-sm',
            ]">
              <div v-if="$props.isLoading" class="block h-6 sm:h-5 animate-pulse bg-gray-200 w-48 rounded-md"></div>
              <span v-else class="block truncate">{{
                modelValue == null ? "No value selected" : getKey(modelValue)
              }}</span>
              <span class="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
                <SelectorIcon class="w-5 h-5 text-gray-400" aria-hidden="true" />
              </span>
            </ListboxButton>

            <ListboxOptions
              class="absolute z-20 w-full mt-1 overflow-auto text-base bg-white dark:bg-secondary-dark rounded-md shadow-sm max-h-60 ring-1 ring-black ring-opacity-5 focus:outline-none sm:text-sm">
              <div v-if="$props.isLoading" class="flex py-5 w-full justify-center">
                <LoadingIcon />
              </div>
              <div v-else>
                <ListboxOption as="template" v-slot="{ active, selected }" v-if="nullable" :value="null">
                  <li :class="[
                    active
                      ? 'text-white bg-primary'
                      : 'text-gray-900 dark:text-gray-50',
                    'cursor-default select-none relative py-2 pl-3 pr-9',
                  ]">
                    <span :class="[
                      selected ? 'font-semibold' : 'font-normal',
                      'block truncate',
                    ]">
                      Unselect
                    </span>

                    <span v-if="selected" :class="[
                      active ? 'text-white' : 'text-primary',
                      'absolute inset-y-0 right-0 flex items-center pr-4',
                    ]">
                      <CheckIcon class="w-5 h-5" aria-hidden="true" />
                    </span>
                  </li>
                </ListboxOption>
                <ListboxOption as="template" v-for="value in $props.values" :key="value" :value="value.value"
                  v-slot="{ active, selected }">
                  <li :class="[
                    active
                      ? 'text-white bg-primary'
                      : 'text-gray-900 dark:text-gray-50',
                    'cursor-default select-none relative py-2 pl-3 pr-9',
                  ]">
                    <span :class="[
                      selected ? 'font-semibold' : 'font-normal',
                      'block truncate',
                    ]">
                      {{ value.key }}
                    </span>

                    <span v-if="selected" :class="[
                      active ? 'text-white' : 'text-primary',
                      'absolute inset-y-0 right-0 flex items-center pr-4',
                    ]">
                      <CheckIcon class="w-5 h-5" aria-hidden="true" />
                    </span>
                  </li>
                </ListboxOption>
              </div>
            </ListboxOptions>
          </div>
        </Listbox>
        <div v-if="$props.validation?.$invalid" class="errors">
          <span v-bind:key="index" v-for="(message, index) in $props.validation?.$errors">{{ message.$message }}</span>
        </div>
      </div>

      <div v-else-if="type == FormTypeEmbed">
        <embed-builder :id="componentId" :modelValue="modelValue" @update:modelValue="updateValue($event)" @blur="blur"
          :disabled="$props.disabled" :invalid="$props.validation?.$invalid" />
        <div v-if="$props.validation?.$invalid" class="errors">
          <span v-bind:key="index" v-for="(message, index) in $props.validation?.$errors">{{ message.$message }}</span>
        </div>
      </div>

      <div v-else-if="type == FormTypeBackground">
        <background-selector :id="componentId" :modelValue="modelValue" @update:modelValue="updateValue($event)"
          @update:files="updateFiles($event)" @blur="blur" :files="$props.files" :disabled="$props.disabled"
          :invalid="$props.validation?.$invalid" :customImages="$props.customImages" />
        <div v-if="$props.validation?.$invalid" class="errors">
          <span v-bind:key="index" v-for="(message, index) in $props.validation?.$errors">{{ message.$message }}</span>
        </div>
      </div>

      <div v-else-if="type == FormTypeCustom">
        <slot name="custom"></slot>
      </div>

      <div v-else-if="type == FormTypeNumberWithConfirm">
        <div class="flex flex-row space-x-2">
          <input :id="componentId" type="number" :class="[
            $props.validation?.$invalid
              ? 'ring-red-500 border-red-500 dark:ring-red-500 dark:border-red-500'
              : '',
            $props.disabled
              ? 'bg-gray-100 dark:bg-secondary-light text-neutral-500'
              : 'bg-white dark:bg-secondary-dark',
            'flex-1 shadow-sm block w-full min-w-0 border-gray-300 dark:border-secondary-light rounded-md focus:ring-primary focus:border-primary sm:text-sm',
          ]" :disabled="$props.disabled" :value="modelValue" @input="updateValue($event.target.value)" @blur="blur" />
          <button type="button" :class="[$props.disabled ? 'bg-gray-100 dark:bg-secondary-light text-neutral-500' : 'bg-primary hover:bg-primary-dark', 'cta-button']" @click="save(modelValue)" :disabled="$props.disabled">
            Save
          </button>
        </div>

        <div v-if="$props.validation?.$invalid" class="errors">
          <span v-bind:key="index" v-for="(message, index) in $props.validation?.$errors">{{ message.$message }}</span>
        </div>
      </div>
    </div>

    <div class="text-gray-600 dark:text-gray-400 text-sm col-span-3 mt-2 sm:mt-0" v-if="!$props.inlineSlot">
      <slot></slot>
    </div>

    <div v-if="type == FormTypeBlank">
      <div v-if="$props.validation?.$invalid" class="errors">
        <span v-bind:key="index" v-for="(message, index) in $props.validation?.$errors">{{ message.$message }}</span>
      </div>
    </div>

  </div>
</template>

<style>
.code_editor>.code_area>pre {
  display: flex;
}
</style>

<script>
import LoadingIcon from "@/components/LoadingIcon.vue";
import AutocompleteInput from "@/components/AutocompleteInput.vue";

import {
  Listbox,
  ListboxButton,
  ListboxLabel,
  ListboxOption,
  ListboxOptions,
  Switch,
} from "@headlessui/vue";
import { CheckIcon, SelectorIcon } from "@heroicons/vue/solid";
import { XIcon } from "@heroicons/vue/outline";
import parse from "parse-css-color";

import { ColorPicker } from "vue-color-kit";
import "vue-color-kit/dist/vue-color-kit.css";

import {
  FormTypeBlank,
  FormTypeToggle,
  FormTypeChannelList,
  FormTypeChannelListCategories,
  FormTypeRoleList,
  FormTypeMemberList,
  FormTypeEmojiList,
  FormTypeColour,
  FormTypeText,
  FormTypeTextArea,
  FormTypeNumber,
  FormTypeDropdown,
  FormTypeEmbed,
  FormTypeBackground,
  FormTypeGuildList,
  FormTypeCustom,
  FormTypeNumberWithConfirm,
} from "./FormValueEnum";
import EmbedBuilder from "./EmbedBuilder.vue";
import BackgroundSelector from "./BackgroundSelector.vue";

import { getHexColor } from "@/utilities";

export default {
  components: {
    Listbox,
    ListboxButton,
    ListboxLabel,
    ListboxOption,
    ListboxOptions,
    Switch,
    CheckIcon,
    SelectorIcon,
    XIcon,
    ColorPicker,
    LoadingIcon,
    EmbedBuilder,
    BackgroundSelector,
    AutocompleteInput,
  },

  props: {
    title: {
      type: String,
      required: true,
    },
    description: {
      type: String,
      required: false,
    },
    type: {
      type: Number,
      required: true,
      validator(value) {
        return [
          FormTypeBlank,
          FormTypeToggle,
          FormTypeChannelList,
          FormTypeChannelListCategories,
          FormTypeRoleList,
          FormTypeMemberList,
          FormTypeEmojiList,
          FormTypeColour,
          FormTypeText,
          FormTypeTextArea,
          FormTypeNumber,
          FormTypeDropdown,
          FormTypeEmbed,
          FormTypeBackground,
          FormTypeGuildList,
          FormTypeCustom,
          FormTypeNumberWithConfirm,
        ].includes(value);
      },
    },
    disabled: {
      type: Boolean,
      required: false,
    },
    modelValue: {
      type: null,
      required: false,
    },
    nullable: {
      type: Boolean,
      required: false,
    },
    invalid: {
      type: Boolean,
    },
    maxLength: {
      type: Number,
      required: false,
      default: 524288
    },
    isLoading: {
      type: Boolean,
      required: false,
    },
    values: {
      required: false,
    },
    errors: {
      type: Array,
    },
    validation: {
      type: Object,
    },
    files: {
      type: Object,
      required: false,
    },
    inlineFormValue: {
      type: Boolean,
      required: false,
    },
    inlineSlot: {
      type: Boolean,
      required: false,
    },
    hideBorder: {
      type: Boolean,
      required: false,
    },
    customImages: {
      type: Array,
      required: false,
    },
    channelFilter: {
      type: Number,
      required: false,
    },
    allowAlpha: {
      type: Boolean,
      required: false,
    }
  },

  emits: ["update:modelValue", "update:files", "blur", "save"],

  setup() {
    let componentId = "formvalue_" + Math.random().toString(16).slice(2);

    let query = "";
    let isValidSnowflake = false;

    const idRegex = new RegExp("([0-9]{15,20})");

    return {
      componentId,

      FormTypeBlank,
      FormTypeToggle,
      FormTypeChannelList,
      FormTypeChannelListCategories,
      FormTypeRoleList,
      FormTypeMemberList,
      FormTypeEmojiList,
      FormTypeColour,
      FormTypeText,
      FormTypeTextArea,
      FormTypeNumber,
      FormTypeDropdown,
      FormTypeEmbed,
      FormTypeBackground,
      FormTypeGuildList,
      FormTypeCustom,
      FormTypeNumberWithConfirm,

      idRegex,

      query,
      isValidSnowflake,
    };
  },

  methods: {
    getHexColor,

    refreshStore() {
      switch (props.type) {
        case FormTypeChannelList:
          break;
        case FormTypeChannelListCategories:
          break;
        case FormTypeRoleList:
          break;
        case FormTypeMemberList:
          break;
        case FormTypeEmojiList:
          break;
        case FormTypeGuildList:
          break;
      }
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

      if (a == 1 || !this.$props.allowAlpha) {
        return `#${r.toString(16).toUpperCase().padStart(2, "0")}${g
          .toString(16)
          .toUpperCase()
          .padStart(2, "0")}${b.toString(16).toUpperCase().padStart(2, "0")}`;
      } else {
        a = Math.round(a * 100) / 100;

        return `rgba(${r}, ${g}, ${b}, ${a})`;
      }
    },

    onColorChange(color) {
      var { r, g, b, a } = color.rgba;

      if (a == 1) {
        this.updateValue(color.hex);
      } else {
        a = Math.round(a * 100) / 100;

        this.updateValue(`rgba(${r}, ${g}, ${b}, ${a})`);
      }
    },

    updateValue(value) {
      if (this.$props.disabled) {
        return;
      }
      this.$emit("update:modelValue", value);
    },

    updateFiles(value) {
      if (this.$props.disabled) {
        return;
      }
      this.$emit("update:files", value);
    },

    blur() {
      this.$props.validation?.$touch();
    },

    save(value) {
      this.$emit("save", value);
    },

    filterChannels(channels) {
      if (this.$props.channelFilter == undefined) {
        return channels
      }

      return channels.filter((channel) => channel.type === this.$props.channelFilter);
    },

    getKey(value) {
      let matches = this.$props.values.filter((v) => {
        return v.value == value;
      });

      if (matches.length === 0) {
        return value;
      } else {
        return matches[0].key;
      }
    },

    onQueryChange() {
      this.isValidSnowflake = this.query.match(this.idRegex) != undefined;
      this.fetchGuildMemberByQuery(this);
    },

    fetchGuildMemberByQuery: (self) => {
      // TODO: Move into component.
      self.$store.dispatch("fetchGuildMembersByQuery", self.query);
    },
  },
};
</script>
