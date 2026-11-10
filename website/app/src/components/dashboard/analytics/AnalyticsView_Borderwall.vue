<template>
  <div class="dashboard-container">
    <div class="dashboard-content">
      <div class="flex justify-end min-h-12">
        <AnalyticsDatePicker :modelValue="analyticsDateRange" @update:modelValue="handleDateRangeUpdate" :maxDate="todayDate"
          :defaultPreset="'last_7_days'" />
      </div>
      <div v-if="isDataError" class="mb-4">
        <p class="mb-2">Unable to load Borderwall analytics.</p>
        <button class="text-blue-500 hover:underline" @click="fetch">Retry</button>
      </div>
      <div v-else-if="!isDataFetched" class="pointer-events-none w-full h-full absolute left-0 top-0 flex items-center justify-center">
        <LoadingIcon />
      </div>
      <template v-else-if="data && !isDataError">
        <div class="grid grid-cols-1 gap-4 mt-2 lg:grid-cols-4 mb-4">
          <AnalyticsCard name="Total Requests" :amount="data.current.total_requests"
            :previousAmount="data.previous.total_requests" icon="fa-shield" />
          <AnalyticsCard name="Verified Requests" :amount="data.current.verified_requests"
            :previousAmount="data.previous.verified_requests" />
          <AnalyticsCard name="Pass Rate" :amount="`${data.current.pass_rate.toFixed(1)}%`"
            :comparisonAmount="data.current.pass_rate" :previousAmount="data.previous.pass_rate" />
          <AnalyticsCard name="Average Time to Verify"
            :amount="formatDuration(data.current.average_verification_time_ms)"
            :comparisonAmount="data.current.average_verification_time_ms"
            :previousAmount="data.previous.average_verification_time_ms" />
        </div>
        <div class="mb-4">
          <AnalyticsCardSlot name="Requests Tracked">
            <Line class="h-72" v-if="data.current.requests_by_time.length" :data="GetDatasetsLine([
              GetDatasetLine('#2F80ED', 'Requests', data.current.requests_by_time).HasTimestamp(),
              GetDatasetLine('#27AE60', 'Verified', data.current.verified_by_time).HasTimestamp(),
            ])"
              :options="ChartOptions({ label: (context) => `${context.dataset.label}: ${context.parsed.y}` }).WithLegend()" />
            <p v-else class="text-sm text-gray-500">No requests were recorded in this period.</p>
          </AnalyticsCardSlot>
        </div>
        <div class="grid grid-cols-1 gap-4 lg:grid-cols-2 mb-4">
          <AnalyticsCardSlot name="Top Countries">
            <Doughnut v-if="data.current.top_countries.length" :data="doughnutData(data.current.top_countries)"
              :options="doughnutOptions" />
            <p v-else class="text-sm text-gray-500">No requests were recorded in this period.</p>
          </AnalyticsCardSlot>
          <AnalyticsCardSlot name="Top Operating Systems">
            <Doughnut v-if="data.current.top_devices.length" :data="doughnutData(data.current.top_devices)"
              :options="doughnutOptions" />
            <p v-else class="text-sm text-gray-500">No requests were recorded in this period.</p>
          </AnalyticsCardSlot>
        </div>
      </template>
    </div>
  </div>
</template>

<script>
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, Tooltip, ArcElement, DoughnutController, Legend } from "chart.js";
import { Doughnut, Line } from "vue-chartjs";
import AnalyticsCard from "@/components/dashboard/analytics/AnalyticsCard.vue";
import AnalyticsCardSlot from "@/components/dashboard/analytics/AnalyticsCardSlot.vue";
import AnalyticsDatePicker from "@/components/dashboard/analytics/AnalyticsDatePicker.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";
import analyticsView from "@/components/dashboard/analytics/analyticsView";
import { ChartOptions, GetDatasetsLine, GetDatasetLine } from "@/components/dashboard/analytics/analytics";

const chartColours = ["#5865F2", "#27AE60", "#F04747", "#F1C40F", "#2F80ED", "#9B59B6", "#16A085", "#E67E22", "#95A5A6", "#34495E"];

export default {
  mixins: [analyticsView],
  data() {
    return {
      analyticsEndpoint: "borderwall",
      doughnutOptions: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: { legend: { position: "bottom" } },
      },
    };
  },
  components: { AnalyticsCard, AnalyticsCardSlot, AnalyticsDatePicker, LoadingIcon, Doughnut, Line },
  setup() {
    ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, ArcElement, DoughnutController, Legend);
    return { ChartOptions, GetDatasetsLine, GetDatasetLine };
  },
  methods: {
    doughnutData(items) {
      return {
        labels: items.map((item) => item[0]),
        datasets: [{
          data: items.map((item) => item[1]),
          backgroundColor: items.map((_, index) => chartColours[index % chartColours.length]),
          borderWidth: 0,
        }],
      };
    },
    formatDuration(milliseconds) {
      if (milliseconds < 1000) return `${milliseconds} ms`;
      if (milliseconds < 60000) return `${(milliseconds / 1000).toFixed(1)} s`;
      return `${Math.floor(milliseconds / 60000)}m ${Math.floor((milliseconds % 60000) / 1000)}s`;
    },
  },
};
</script>
