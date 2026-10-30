<template>
  <router-view />
</template>

<script>
import store from "./store/index";

export default {
  setup() {
    if (import.meta.env.SSR) {
      // Match the client's initial "loading" state so hydration lines up.
      store.commit("loadingUser");
    } else {
      store.dispatch("fetchCurrentUser");
    }
  },
};
</script>
