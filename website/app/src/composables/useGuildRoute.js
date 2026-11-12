import { watch } from "vue";

import { useRoute } from "vue-router";
import { useStore } from "vuex";

export default function useGuildRoute() {
  const route = useRoute();
  const store = useStore();

  watch(
    () => route.params.guildID,
    (guildID) => {
      store.dispatch("selectGuild", guildID);
    },
    { immediate: true }
  );
}
