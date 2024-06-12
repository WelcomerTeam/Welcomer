<template>
  <div>
    <button class="hidden dark:block" @click="setDarkTheme(false)">
      <div
        class="flex-shrink-0 h-8 inline-flex items-center justify-center rounded-md text-blue-50 w-8 bg-secondary-light">
        <font-awesome-icon icon="moon" class="w-4 h-4" aria-hidden="true" />
      </div>
    </button>
    <button class="dark:hidden block" @click="setDarkTheme(true)">
      <div class="flex-shrink-0 h-8 inline-flex items-center justify-center rounded-md text-white w-8 bg-primary">
        <font-awesome-icon icon="sun" class="w-4 h-4" aria-hidden="true" />
      </div>
    </button>
  </div>
</template>

<script>
export default {
  methods: {
    setDarkTheme(isEnabled) {
      if (isEnabled) {
        localStorage.theme = "dark";
        document.documentElement.classList.add("dark");
      } else {
        localStorage.theme = "light";
        document.documentElement.classList.remove("dark");
      }
    },
  },
};
</script>
