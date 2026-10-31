<template>
  <div class="dashboard-container">
    <div class="dashboard-content">
      <div class="flex justify-end min-h-12">
        <AnalyticsDatePicker :modelValue="'last_7_days'" @update:modelValue="handleDateRangeUpdate" :maxDate="todayDate"
          :defaultPreset="'last_7_days'" />
      </div>

      <div v-if="isDataError" class="mb-4">
        <p class="mb-2">Unable to load user activity.</p>
        <button class="text-blue-500 hover:underline" @click="fetch">Retry</button>
      </div>
      <div v-else-if="!isDataFetched" class="w-full h-full absolute left-0 top-0 flex items-center justify-center">
        <LoadingIcon />
      </div>
      <template v-else-if="data && !isDataError">
        <div class="grid grid-cols-1 gap-4 mt-2 lg:grid-cols-3 mb-4">
          <AnalyticsCard name="Daily Active Users" :amount="data.current.daily_active"
            :previousAmount="data.previous.daily_active" icon="fa-user" />
          <AnalyticsCard name="Weekly Active Users" :amount="data.current.weekly_active"
            :previousAmount="data.previous.weekly_active" />
          <AnalyticsCard name="Monthly Active Users" :amount="data.current.monthly_active"
            :previousAmount="data.previous.monthly_active" />
        </div>
        <AnalyticsCardSlot name="Unique User Activity" class="mb-4">
            <Line class="h-72" v-if="data.current.unique_user_activity.length" :data="GetDatasetsLine([
              GetDatasetLine('#2F80ED', 'Active users', data.current.unique_user_activity).HasTimestamp().WithGradientBackgroundColor(),
            ])" :options="ChartOptions({ label: (context) => `${context.parsed.y} active users` })" />
            <p v-else class="text-sm text-gray-500">No data available for this period.</p>
        </AnalyticsCardSlot>

        <div class="grid grid-cols-1 gap-4 mt-4 lg:grid-cols-3 mb-4">
          <AnalyticsCard name="Total Members" :amount="data.current.total_members" icon="fa-user" />
          <AnalyticsCard name="Active Members" :amount="data.current.active_members" />
          <AnalyticsCard name="Dormant Members" :amount="data.current.dormant_members" :isDecreaseGood="true" />
        </div>
        <AnalyticsCardSlot name="Message Activity" class="mb-4">
            <Line class="h-72" v-if="data.current.total_user_activity.length" :data="GetDatasetsLine([
              GetDatasetLine('#2F80ED', 'Messages', data.current.total_user_activity).HasTimestamp().WithGradientBackgroundColor(),
            ])" :options="ChartOptions({ label: (context) => `${context.parsed.y} messages` })" />
            <p v-else class="text-sm text-gray-500">No data available for this period.</p>
        </AnalyticsCardSlot>
        <div class="grid grid-cols-1 gap-4 mt-4 lg:grid-cols-3 mb-2">
          <AnalyticsCard name="Joined Members Who Messaged" :amount="data.current.time_to_first_message"
            :previousAmount="data.previous.time_to_first_message" />
          <AnalyticsCard name="Average Time to First Message"
            :amount="formatDuration(data.current.average_time_to_first_message)"
            :comparisonAmount="data.current.average_time_to_first_message"
            :previousAmount="data.previous.average_time_to_first_message" />
          <AnalyticsCard name="Median Time to First Message"
            :amount="formatDuration(data.current.median_time_to_first_message)"
            :comparisonAmount="data.current.median_time_to_first_message"
            :previousAmount="data.previous.median_time_to_first_message" />
        </div>
        <p class="text-xs text-gray-500 mb-4">
          Onboarding times include members who joined during the selected period and sent a message during that period.
        </p>
      </template>
    </div>
  </div>
</template>

<script>
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Filler } from "chart.js";
import { Line } from "vue-chartjs";
import AnalyticsCard from "@/components/dashboard/analytics/AnalyticsCard.vue";
import AnalyticsCardSlot from "@/components/dashboard/analytics/AnalyticsCardSlot.vue";
import AnalyticsDatePicker from "@/components/dashboard/analytics/AnalyticsDatePicker.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";
import analyticsView from "@/components/dashboard/analytics/analyticsView";
import { ChartOptions, GetDatasetsLine, GetDatasetLine } from "@/components/dashboard/analytics/analytics";

export default {
  mixins: [analyticsView],
  data() {
    return { analyticsEndpoint: "user-activity" };
  },
  components: { AnalyticsCard, AnalyticsCardSlot, AnalyticsDatePicker, LoadingIcon, Line },
  setup() {
    ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Filler);
    return { ChartOptions, GetDatasetsLine, GetDatasetLine };
  },
  methods: {
    formatDuration(seconds) {
      if (seconds >= 86400) return `${Math.floor(seconds / 86400)}d ${Math.floor((seconds % 86400) / 3600)}h`;
      if (seconds >= 3600) return `${Math.floor(seconds / 3600)}h ${Math.floor((seconds % 3600) / 60)}m`;
      if (seconds >= 60) return `${Math.floor(seconds / 60)}m ${seconds % 60}s`;
      return `${seconds}s`;
    },
  },
};
</script>
