<template>
  <div class="dashboard-container">
    <div class="dashboard-content">
      <div class="flex justify-end min-h-12">
        <AnalyticsDatePicker :modelValue="'last_7_days'" @update:modelValue="handleDateRangeUpdate" :maxDate="todayDate"
          :defaultPreset="'last_7_days'" />
      </div>
      <div v-if="isDataError" class="mb-4">
        <p class="mb-2">Unable to load invite analytics.</p>
        <button class="text-blue-500 hover:underline" @click="fetch">Retry</button>
      </div>
      <div v-else-if="!isDataFetched" class="w-full h-full absolute left-0 top-0 flex items-center justify-center">
        <LoadingIcon />
      </div>
      <template v-else-if="data && !isDataError">
        <div class="grid grid-cols-1 gap-4 mt-2 lg:grid-cols-4 mb-4">
          <AnalyticsCard name="Tracked Invites" :amount="data.current.total_invites" icon="fa-user-plus" />
          <AnalyticsCard name="Invite Joins" :amount="data.current.invite_joins"
            :previousAmount="data.previous.invite_joins" />
          <AnalyticsCard name="Tracked Joins" :amount="data.current.tracked_joins"
            :previousAmount="data.previous.tracked_joins" />
          <AnalyticsCard name="Untracked Joins" :amount="data.current.untracked_joins"
            :previousAmount="data.previous.untracked_joins" />
        </div>
        <div class="grid grid-cols-1 gap-4 mt-2 lg:grid-cols-1 mb-4">
          <AnalyticsCard name="Tracked Joins per Saved Invite"
            :amount="data.current.average_joins_per_invite.toFixed(1)"
            :comparisonAmount="data.current.average_joins_per_invite"
            :previousAmount="data.previous.average_joins_per_invite" />
        </div>
        <div class="mb-4">
          <AnalyticsCardSlot name="Invites Tracked">
            <div class="h-72">
              <Line :data="GetDatasetsLine([
                GetDatasetLine('#2F80ED', 'All joins', data.current.joins_by_time).HasTimestamp(),
                GetDatasetLine('#27AE60', 'Invite tracked', data.current.tracked_joins_by_time).HasTimestamp().WithBorderDash([5, 5]),
              ])"
                :options="ChartOptions({ label: (context) => `${context.dataset.label}: ${context.parsed.y} joins` }).WithLegend()" />
            </div>
          </AnalyticsCardSlot>
        </div>
        <div class="mb-4">
          <AnalyticsCardSlot name="Top Invites by Tracked Joins">
            <div v-if="data.current.top_invites.length" class="h-72">
              <Bar :data="GetDatasetsLine([
                GetDatasetBar('#5865F2', 'Joins', data.current.top_invites),
              ])" :options="barOptions" />
            </div>
            <p v-else class="text-sm text-gray-500">No invite-attributed joins were recorded in this period.</p>
          </AnalyticsCardSlot>
        </div>
        <div class="mb-4">
          <AnalyticsCardSlot name="Top Invite Creators">
            <div v-if="data.current.top_creators.length" class="h-72">
              <Bar :data="GetDatasetsLine([
                GetDatasetBar('#27AE60', 'Tracked joins', data.current.top_creators),
              ])" :options="barOptions" />
            </div>
            <p v-else class="text-sm text-gray-500">No invite-attributed joins were recorded in this period.</p>
          </AnalyticsCardSlot>
        </div>
        <p class="text-xs text-gray-500 mt-3">
          Invite totals count saved invite snapshots. Invite attribution is captured for new joins when the gateway can
          identify the used invite; historical joins from before tracking was enabled are not backfilled. Discord does
          not provide historical expiration or use-limit details here, so active, inactive, and unused invite counts are
          unavailable.
        </p>
      </template>
    </div>
  </div>
</template>

<script>
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, BarController, BarElement, Tooltip } from "chart.js";
import { Bar, Line } from "vue-chartjs";
import AnalyticsCard from "@/components/dashboard/analytics/AnalyticsCard.vue";
import AnalyticsCardSlot from "@/components/dashboard/analytics/AnalyticsCardSlot.vue";
import AnalyticsDatePicker from "@/components/dashboard/analytics/AnalyticsDatePicker.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";
import analyticsView from "@/components/dashboard/analytics/analyticsView";
import { ChartOptions, GetDatasetsLine, GetDatasetLine, GetDatasetBar } from "@/components/dashboard/analytics/analytics";

export default {
  mixins: [analyticsView],
  data() {
    return {
      analyticsEndpoint: "invites",
      barOptions: {
        responsive: true,
        maintainAspectRatio: false,
        indexAxis: "y",
        plugins: { legend: { display: false } },
        scales: { x: { beginAtZero: true } },
      },
    };
  },
  components: { AnalyticsCard, AnalyticsCardSlot, AnalyticsDatePicker, LoadingIcon, Bar, Line },
  setup() {
    ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, BarController, BarElement);
    return { ChartOptions, GetDatasetsLine, GetDatasetLine, GetDatasetBar };
  },
};
</script>
