<template>
  <div class="dashboard-container">
    <div class="dashboard-content">
      <div class="flex justify-end min-h-12">
        <AnalyticsDatePicker :modelValue="'last_7_days'" @update:modelValue="handleDateRangeUpdate" :maxDate="todayDate"
          :defaultPreset="'last_7_days'" />
      </div>
      <div v-if="isDataError" class="mb-4">
        <p class="mb-2">Unable to load voice activity.</p>
        <button class="text-blue-500 hover:underline" @click="fetch">Retry</button>
      </div>
      <div v-else-if="!isDataFetched && !isDataError"
        class="w-full h-full absolute left-0 top-0 flex items-center justify-center">
        <LoadingIcon />
      </div>
      <template v-else-if="data">
        <div class="grid grid-cols-1 gap-4 mt-2 lg:grid-cols-4 mb-4">
          <AnalyticsCard name="Voice Sessions" :amount="data.current.voice_sessions"
            :previousAmount="data.previous.voice_sessions" icon="fa-microphone-lines" />
          <AnalyticsCard name="Unique Users" :amount="data.current.unique_users"
            :previousAmount="data.previous.unique_users" />
          <AnalyticsCard name="Total Voice Time" :amount="formatDuration(data.current.total_voice_time_ms)"
            :comparisonAmount="data.current.total_voice_time_ms" :previousAmount="data.previous.total_voice_time_ms" />
          <AnalyticsCard name="Average Session" :amount="formatDuration(data.current.average_session_ms)"
            :comparisonAmount="data.current.average_session_ms" :previousAmount="data.previous.average_session_ms" />
        </div>
        <div class="mb-4">
          <AnalyticsCardSlot name="Voice Sessions">
            <Line class="h-72" v-if="data.current.voice_sessions_by_time.length" :data="GetDatasetsLine([
              GetDatasetLine('#2F80ED', 'Sessions', data.current.voice_sessions_by_time).HasTimestamp().WithGradientBackgroundColor(),
            ])" :options="voiceSessionsOptions" />
            <p v-else class="text-sm text-gray-500">No voice sessions were recorded in this period.</p>
          </AnalyticsCardSlot>
        </div>
        <div class="mb-4">
          <AnalyticsCardSlot name="Total Voice Time">
            <Line class="h-72" v-if="data.current.voice_time_by_time.length" :data="GetDatasetsLine([
              GetDatasetLine('#27AE60', 'Total Voice Time (seconds)', data.current.voice_time_by_time).HasTimestamp().WithGradientBackgroundColor(),
            ])" :options="voiceTimeOptions" />
            <p v-else class="text-sm text-gray-500">No voice time was recorded in this period.</p>
          </AnalyticsCardSlot>
        </div>
        <div class="mb-4">
          <AnalyticsCardSlot name="Top Voice Users">
            <Bar class="h-72" v-if="data.current.top_users.length" :data="GetDatasetsLine([
              GetDatasetBar('#5865F2', 'Voice time (seconds)', data.current.top_users),
            ])" :options="barOptions" />
            <p v-else class="text-sm text-gray-500">No completed voice sessions were recorded in this period.</p>
          </AnalyticsCardSlot>
        </div>
      </template>
    </div>
  </div>
</template>

<script>
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, LineElement, BarController, BarElement, Tooltip, Filler } from "chart.js";
import { Bar, Line } from "vue-chartjs";
import AnalyticsCard from "@/components/dashboard/analytics/AnalyticsCard.vue";
import AnalyticsCardSlot from "@/components/dashboard/analytics/AnalyticsCardSlot.vue";
import AnalyticsDatePicker from "@/components/dashboard/analytics/AnalyticsDatePicker.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";
import analyticsView from "@/components/dashboard/analytics/analyticsView";
import { GetDatasetsLine, GetDatasetLine, GetDatasetBar } from "@/components/dashboard/analytics/analytics";

export default {
  mixins: [analyticsView],
  data() {
    return {
      analyticsEndpoint: "voice-activity",
      voiceSessionsOptions: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { display: false },
          tooltip: {
            callbacks: {
              label: (context) => `${context.parsed.y} sessions`,
            },
          },
        },
        scales: {
          y: {
            beginAtZero: true,
            ticks: { precision: 0, stepSize: 1 },
          },
        },
      },
      voiceTimeOptions: {
        responsive: true,
        maintainAspectRatio: false,
        plugins: {
          legend: { display: false },
          tooltip: {
            callbacks: {
              label: (context) => `${(context.parsed.y / 60).toFixed(0)} minutes`,
            },
          },
        },
        scales: {
          y: {
            beginAtZero: true,
            title: { display: true, text: "Minutes" },
            ticks: {
              precision: 0,
              callback: (value) => `${(value / 60).toFixed(0)}m`,
            },
          },
        },
      },
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
    ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Filler, BarController, BarElement);
    return { GetDatasetsLine, GetDatasetLine, GetDatasetBar };
  },
  methods: {
    formatDuration(milliseconds) {
      let seconds = Math.floor(milliseconds / 1000);
      const days = Math.floor(seconds / 86400);
      seconds %= 86400;
      const hours = Math.floor(seconds / 3600);
      seconds %= 3600;
      const minutes = Math.floor(seconds / 60);
      if (days > 0) return `${days}d ${hours}h`;
      if (hours > 0) return `${hours}h ${minutes}m`;
      if (minutes > 0) return `${minutes}m`;
      return `${seconds}s`;
    },
  },
};
</script>
