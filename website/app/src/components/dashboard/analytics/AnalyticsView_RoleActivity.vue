<template>
  <div class="dashboard-container">
    <div class="dashboard-content">
      <div class="flex justify-end min-h-12">
        <AnalyticsDatePicker :modelValue="'last_7_days'" @update:modelValue="handleDateRangeUpdate" :maxDate="todayDate"
          :defaultPreset="'last_7_days'" />
      </div>
      <div v-if="isDataError" class="mb-4">
        <p class="mb-2">Unable to load role activity.</p>
        <button class="text-blue-500 hover:underline" @click="fetch">Retry</button>
      </div>
      <div v-else-if="!isDataFetched" class="w-full h-full absolute left-0 top-0 flex items-center justify-center">
        <LoadingIcon />
      </div>
      <template v-else-if="data && !isDataError">
        <div class="grid grid-cols-1 gap-4 mt-2 lg:grid-cols-2 mb-4">
          <AnalyticsCard name="Recorded Role Changes" :amount="totalChanges" :previousAmount="previousTotalChanges"
            icon="fa-user-group" />
          <AnalyticsCard name="Roles Changed" :amount="data.current.role_changes.length" />
        </div>
        <AnalyticsCardSlot name="Roles Given by Module and Role" class="mb-4">
          <div v-if="data.current.role_changes.length" class="h-72">
            <Bar :data="GetDatasetsLine([
              GetDatasetBar('#5865F2', 'Role changes', data.current.role_changes),
            ])" :options="barOptions" />
          </div>
          <p v-else class="text-sm text-gray-500">No autorole, freerole, reaction-role, or timerole grants were recorded
            in this period.</p>
        </AnalyticsCardSlot>
        <p class="text-xs text-gray-500 mt-3">
          Bars are grouped by the Welcomer module that granted each role: autoroles, freeroles, reaction roles, or
          timeroles. Only grants recorded after role analytics logging was deployed are included.
        </p>
      </template>
    </div>
  </div>
</template>

<script>
import { Chart as ChartJS, CategoryScale, LinearScale, BarController, BarElement, Tooltip } from "chart.js";
import { Bar } from "vue-chartjs";
import AnalyticsCard from "@/components/dashboard/analytics/AnalyticsCard.vue";
import AnalyticsCardSlot from "@/components/dashboard/analytics/AnalyticsCardSlot.vue";
import AnalyticsDatePicker from "@/components/dashboard/analytics/AnalyticsDatePicker.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";
import analyticsView from "@/components/dashboard/analytics/analyticsView";
import { GetDatasetsLine, GetDatasetBar } from "@/components/dashboard/analytics/analytics";

export default {
  mixins: [analyticsView],
  data() {
    return {
      analyticsEndpoint: "role-activity",
      barOptions: {
        responsive: true,
        maintainAspectRatio: false,
        indexAxis: "y",
        plugins: { legend: { display: false } },
        scales: { x: { beginAtZero: true } },
      },
    };
  },
  components: { AnalyticsCard, AnalyticsCardSlot, AnalyticsDatePicker, LoadingIcon, Bar },
  setup() {
    ChartJS.register(CategoryScale, LinearScale, BarController, BarElement, Tooltip);
    return { GetDatasetsLine, GetDatasetBar };
  },
  computed: {
    totalChanges() {
      return (this.data?.current?.role_changes ?? []).reduce((total, item) => total + item[1], 0);
    },
    previousTotalChanges() {
      return (this.data?.previous?.role_changes ?? []).reduce((total, item) => total + item[1], 0);
    },
  },
};
</script>
