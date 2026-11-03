<template>
  <AnalyticsHeader :tab="tab" @update:tab="tab = $event" />
  <AnalyticsView_Overview v-if="tab === AnalyticsTab_Overview" />
  <AnalyticsView_Retention v-else-if="tab === AnalyticsTab_Retention" />
  <AnalyticsView_UserActivity v-else-if="tab === AnalyticsTab_UserActivity" />
  <AnalyticsView_UserDemographics v-else-if="tab === AnalyticsTab_UserDemographics" />
  <AnalyticsView_MessageActivity v-else-if="tab === AnalyticsTab_MessageActivity" />
  <AnalyticsView_VoiceActivity v-else-if="tab === AnalyticsTab_VoiceActivity" />
  <AnalyticsView_RoleActivity v-else-if="tab === AnalyticsTab_RoleActivity" />
  <AnalyticsView_Invites v-else-if="tab === AnalyticsTab_Invites" />
  <AnalyticsView_Borderwall v-else-if="tab === AnalyticsTab_Borderwall" />
</template>

<script>
import { provide, reactive, ref } from "vue";

import AnalyticsHeader from "@/components/dashboard/analytics/AnalyticsHeader.vue";
import AnalyticsView_Overview from "@/components/dashboard/analytics/AnalyticsView_Overview.vue";
import AnalyticsView_Retention from "@/components/dashboard/analytics/AnalyticsView_Retention.vue";
import AnalyticsView_UserActivity from "@/components/dashboard/analytics/AnalyticsView_UserActivity.vue";
import AnalyticsView_UserDemographics from "@/components/dashboard/analytics/AnalyticsView_UserDemographics.vue";
import AnalyticsView_MessageActivity from "@/components/dashboard/analytics/AnalyticsView_MessageActivity.vue";
import AnalyticsView_VoiceActivity from "@/components/dashboard/analytics/AnalyticsView_VoiceActivity.vue";
import AnalyticsView_RoleActivity from "@/components/dashboard/analytics/AnalyticsView_RoleActivity.vue";
import AnalyticsView_Invites from "@/components/dashboard/analytics/AnalyticsView_Invites.vue";
import AnalyticsView_Borderwall from "@/components/dashboard/analytics/AnalyticsView_Borderwall.vue";

import {
  DefaultAnalyticsTab,
  AnalyticsTab_Overview,
  AnalyticsTab_Retention,
  AnalyticsTab_UserActivity,
  AnalyticsTab_UserDemographics,
  AnalyticsTab_MessageActivity,
  AnalyticsTab_VoiceActivity,
  AnalyticsTab_RoleActivity,
  AnalyticsTab_Invites,
  AnalyticsTab_Borderwall,
} from "@/components/dashboard/analytics/analytics";

export default {
  components: {
    AnalyticsHeader,
    AnalyticsView_Overview,
    AnalyticsView_Retention,
    AnalyticsView_UserActivity,
    AnalyticsView_UserDemographics,
    AnalyticsView_MessageActivity,
    AnalyticsView_VoiceActivity,
    AnalyticsView_RoleActivity,
    AnalyticsView_Invites,
    AnalyticsView_Borderwall,
  },
  setup() {
    var tab = ref(DefaultAnalyticsTab);
    const endDate = new Date();
    const startDate = new Date(endDate);
    startDate.setDate(startDate.getDate() - 6);
    startDate.setHours(0, 0, 0, 0);
    const formatDate = (date) => [
      date.getFullYear(),
      String(date.getMonth() + 1).padStart(2, "0"),
      String(date.getDate()).padStart(2, "0"),
    ].join("-");
    // Views are conditionally mounted, so provide one shared range to preserve it across tab changes.
    const analyticsDateRange = reactive({
      preset: "last_7_days",
      startDate: formatDate(startDate),
      endDate: formatDate(endDate),
    });
    provide("analyticsDateRange", analyticsDateRange);

    return {
      tab,

      AnalyticsTab_Overview,
      AnalyticsTab_Retention,
      AnalyticsTab_UserActivity,
      AnalyticsTab_UserDemographics,
      AnalyticsTab_MessageActivity,
      AnalyticsTab_VoiceActivity,
      AnalyticsTab_RoleActivity,
      AnalyticsTab_Invites,
      AnalyticsTab_Borderwall,
    };
  },
};
</script>
