<template>
  <div class="dashboard-container">
    <div class="dashboard-content">
      <div class="flex justify-end min-h-12">
        <AnalyticsDatePicker :modelValue="'last_7_days'" @update:modelValue="handleDateRangeUpdate" :maxDate="todayDate"
          :defaultPreset="'last_7_days'" />
      </div>
      <div v-if="isDataError" class="mb-4">
        <p class="mb-2">Unable to load message activity.</p>
        <button class="text-blue-500 hover:underline" @click="fetch">Retry</button>
      </div>
      <div v-else-if="!isDataFetched" class="w-full h-full absolute left-0 top-0 flex items-center justify-center">
        <LoadingIcon />
      </div>
      <template v-else-if="data && !isDataError">
        <div class="grid grid-cols-1 gap-4 mt-2 lg:grid-cols-3 mb-4">
          <AnalyticsCard name="Messages Sent" :amount="data.current.total_messages"
            :previousAmount="data.previous.total_messages" icon="fa-chart-line" />
          <AnalyticsCard name="Unique Users" :amount="data.current.unique_users"
            :previousAmount="data.previous.unique_users" />
          <AnalyticsCard name="Messages per User" :amount="data.current.messages_per_user"
            :previousAmount="data.previous.messages_per_user" />
        </div>
        <div class="mb-4">
          <AnalyticsCardSlot name="Messages Sent">
            <Line class="h-72" v-if="data.current.messages_by_time.length" :data="GetDatasetsLine([
              GetDatasetLine('#2F80ED', 'Messages', data.current.messages_by_time).HasTimestamp().WithGradientBackgroundColor(),
            ])" :options="ChartOptions({ label: (context) => `${context.parsed.y} messages` })" />
            <p v-else class="text-sm text-gray-500">No messages were recorded in this period.</p>
          </AnalyticsCardSlot>
        </div>
        <div class="mb-4">
          <AnalyticsCardSlot name="Unique Users by Period">
            <Line class="h-72" v-if="data.current.unique_users_by_time.length" :data="GetDatasetsLine([
              GetDatasetLine('#27AE60', 'Unique users', data.current.unique_users_by_time).HasTimestamp().WithGradientBackgroundColor(),
            ])" :options="ChartOptions({ label: (context) => `${context.parsed.y} unique users` })" />
            <p v-else class="text-sm text-gray-500">No unique users were recorded in this period.</p>
          </AnalyticsCardSlot>
        </div>
        <div class="mb-4">
          <AnalyticsCardSlot name="Messages by Channel">
            <Bar class="h-72" v-if="data.current.messages_by_channel.length" :data="GetDatasetsLine([
              GetDatasetBar('#5865F2', 'Messages', data.current.messages_by_channel),
            ])" :options="barOptions" />
            <p v-else class="text-sm text-gray-500">No messages were recorded in this period.</p>
          </AnalyticsCardSlot>
        </div>
        <div class="mb-4">
          <AnalyticsCardSlot name="Message Heatmap">
            <div class="overflow-x-auto">
              <table class="w-full text-xs text-center">
                <thead>
                  <tr>
                    <th class="p-1"></th>
                    <th v-for="hour in hours" :key="hour" class="p-1 font-normal">{{ hour }}</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="(day, dayIndex) in days" :key="day">
                    <th class="p-2 text-left font-normal">{{ day }}</th>
                    <td v-for="hour in hours" :key="hour" class="p-1">
                      <span class="block min-w-5 h-6 rounded"
                        :style="{ backgroundColor: heatmapColor(cellCount(dayIndex, hour)) }"
                        :title="`${day}, ${hour}:00 - ${cellCount(dayIndex, hour)} messages`"></span>
                    </td>
                  </tr>
                </tbody>
              </table>
            </div>
          </AnalyticsCardSlot>
        </div>
        <div class="mb-4">
          <AnalyticsCardSlot name="Top Message Authors">
            <Bar class="h-72" v-if="data.current.top_authors.length" :data="GetDatasetsLine([
              GetDatasetBar('#27AE60', 'Messages', data.current.top_authors),
            ])" :options="barOptions" />
            <p v-else class="text-sm text-gray-500">No messages were recorded in this period.</p>
          </AnalyticsCardSlot>
        </div>
        <div class="mb-4">
          <AnalyticsCardSlot name="Top Contributors">
            <Bar class="h-72" v-if="data.current.top_contributors.length" :data="GetDatasetsLine([
              GetDatasetBar('#F1C40F', 'Active hours', data.current.top_contributors),
            ])" :options="barOptions" />
            <p v-else class="text-sm text-gray-500">No messages were recorded in this period.</p>
          </AnalyticsCardSlot>
          <p class="text-xs text-gray-500 mt-2">Each hour with at least one message counts once for a user.</p>
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
import { ChartOptions, GetDatasetsLine, GetDatasetLine, GetDatasetBar } from "@/components/dashboard/analytics/analytics";

export default {
  mixins: [analyticsView],
  data() {
    return {
      analyticsEndpoint: "message-activity",
      days: ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"],
      hours: Array.from({ length: 24 }, (_, hour) => hour),
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
    return { ChartOptions, GetDatasetsLine, GetDatasetLine, GetDatasetBar };
  },
  methods: {
    cellCount(day, hour) {
      const cell = this.data?.current?.message_heatmap?.find((point) => point.day_of_week === day && point.hour_of_day === hour);
      return cell?.message_count ?? 0;
    },
    heatmapColor(count) {
      const maximum = Math.max(1, ...(this.data?.current?.message_heatmap ?? []).map((point) => point.message_count));
      const alpha = count === 0 ? 0.04 : 0.12 + (count / maximum) * 0.78;
      return `rgba(47, 128, 237, ${alpha})`;
    },
  },
};
</script>
