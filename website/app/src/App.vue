<template>
  <router-view />
</template>

<script>
import store from "./store/index";

export default {
  setup() {
    store.dispatch("fetchCurrentUser");
  },
};
</script>
