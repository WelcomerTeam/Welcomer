<template>
  <AnalyticsHeader :tab="tab" @update:tab="tab = $event" />
  <AnalyticsView_Overview v-if="tab === AnalyticsTab_Overview" />
</template>

<script>
import { ref } from "vue";

import AnalyticsHeader from "@/components/dashboard/analytics/AnalyticsHeader.vue";
import AnalyticsView_Overview from "@/components/dashboard/analytics/AnalyticsView_Overview.vue";

import {
  AnalyticsTab_Overview,
  DefaultAnalyticsTab,
} from "@/constants";

export default {
  components: {
    AnalyticsHeader,
    AnalyticsView_Overview,
  },
  setup() {
    var tab = ref(DefaultAnalyticsTab);

    return {
      tab,

      AnalyticsTab_Overview,
    };
  },
};
</script>
