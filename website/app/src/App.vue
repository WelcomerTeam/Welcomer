<template>
  <router-view />
</template>

<script>
import { useStore } from "vuex";

export default {
  setup() {
    const store = useStore();

    if (import.meta.env.SSR) {
      // Match the client's initial "loading" state so hydration lines up.
      store.commit("loadingUser");
    } else {
      store.dispatch("fetchCurrentUser");
    }
  },
};
</script>
