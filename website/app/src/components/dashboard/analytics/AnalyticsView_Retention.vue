<template>
    <div class="dashboard-container">
        <div class="dashboard-content">
            <div class="flex justify-end min-h-12">
                <AnalyticsDatePicker @update:modelValue="handleDateRangeUpdate" :defaultPreset="'last_7_days'" />
            </div>

            <div v-if="isDataError">
                <div class="mb-4">Data Error</div>
                <button @click="this.fetch">Retry</button>
            </div>
            <div v-else-if="!isDataFetched" class="w-full h-full absolute left-0 top-0 flex items-center justify-center">
                <LoadingIcon />
            </div>
            <div v-if="data && !isDataError">
                <div class="grid grid-cols-1 gap-4 mt-2 lg:grid-cols-3 mb-4">
                    <AnalyticsCard name="Members Joined" :amount="data.current.members_joined" :previousAmount="data.previous.members_joined" />
                    <AnalyticsCard name="Members Left" :amount="data.current.members_left" :previousAmount="data.previous.members_left" icon="fa-user" :isDecreaseGood="true" />
                    <AnalyticsCard name="Retention Rate" :amount="data.current.retention_percentage" :previousAmount="data.previous.retention_percentage" amountSuffix="%" />
                </div>
                <div class="mb-4 space-y-4">
                    <AnalyticsCardSlot name="Days on Server">
                            <Bar :data="GetDatasetsLine([
                                GetDatasetBar('#2F80ED', 'Left', data.current.time_on_server_before_leaving),
                                GetDatasetBar('#27AE60', 'Still on server', data.current.time_on_server),
                            ])" :options="ChartOptions({
                                label: (context) => { return `${context.dataset.label} ${context.parsed.x == 0 ? 'on the first day' : 'after ' + context.parsed.x + ' day' + (context.parsed.x !== 1 ? 's' : '')}: ${context.parsed.y} members` }
                            }).WithLegend()" />
                    </AnalyticsCardSlot>
                    <AnalyticsCardSlot name="Retention Cohorts">
                        <div class="overflow-x-auto">
                            <table class="w-full text-sm text-center border-collapse overflow-x-auto">
                                <thead>
                                    <tr>
                                        <th></th>
                                        <th class="p-2 border border-gray-700" v-for="(header, headerIndex) in data.current.retention_cohorts" :key="headerIndex">
                                            Left {{ getMonthLabel(headerIndex) }}
                                        </th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr v-for="(row, rowIndex) in data.current.retention_cohorts" :key="rowIndex">
                                        <th class="p-2 border border-gray-700">
                                            Joined {{ getMonthLabel(data.current.retention_cohorts.length - rowIndex - 1) }}
                                        </th>
                                        <td
                                            v-for="(cell, cellIndex) in row.splice(0, data.current.retention_cohorts.length - rowIndex)"
                                            :key="cellIndex"
                                            class="p-2 border border-gray-700"
                                            :style="{ backgroundColor: `rgba(88, 101, 242, ${cell / 100})` }"
                                            :title="`${cell.toFixed(1)}% of users who joined ${getMonthLabel(data.current.retention_cohorts.length - rowIndex - 1)} left ${getMonthLabel(cellIndex)}`"
                                        >
                                            {{ cell.toFixed(1) }}%
                                        </td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                        <span class="text-xs mt-2">
                            This graph shows the percentage of users who remain on the server before leaving. The x-axis is relative to the end date of the selected time period and shows the information from the previous 12 months.
                        </span>
                    </AnalyticsCardSlot>
                </div>
            </div>
        </div>
    </div>
</template>

<script>
import AnalyticsCard from "@/components/dashboard/analytics/AnalyticsCard.vue";
import AnalyticsCardSlot from "@/components/dashboard/analytics/AnalyticsCardSlot.vue";
import AnalyticsCardChange from "@/components/dashboard/analytics/AnalyticsCardChange.vue";
import AnalyticsDatePicker from "@/components/dashboard/analytics/AnalyticsDatePicker.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";

import {
    Chart as ChartJS,
    CategoryScale,
    LinearScale,
    PointElement,
    LineElement,
    BarController,
    BarElement,
    Tooltip,
    Filler,
} from "chart.js";

import { Bar } from "vue-chartjs";

import { ChartOptions, GetDatasetsLine, GetDatasetBar } from "@/components/dashboard/analytics/analytics";

import { ref } from "vue";

import dashboardAPI from "@/api/dashboard";
import endpoints from "@/api/endpoints";
import { getErrorToast } from "@/utilities";

export default {
    components: {
        AnalyticsCard,
        AnalyticsCardSlot,
        AnalyticsCardChange,
        AnalyticsDatePicker,
        Bar,
        LoadingIcon,
    },
    setup() {
        var startDate = ref(new Date(Date.now() - 7 * 24 * 60 * 60 * 1000)); // 7 days ago
        var endDate = ref(new Date());

        var isDataFetched = ref(false);
        var isDataError = ref(false);
        var data = ref(null);

        ChartJS.register(
            CategoryScale,
            LinearScale,
            PointElement,
            LineElement,
            Tooltip,
            Filler,
            BarController,
            BarElement
        );

        return {
            ChartOptions,
            GetDatasetsLine,
            GetDatasetBar,

            startDate,
            endDate,

            isDataFetched,
            isDataError,
            data,
        };
    },
    mounted() {
        this.fetch();
    },
    methods: {
        fetch() {
            this.isDataError = false;
            this.isDataFetched = false;

            dashboardAPI.getConfig(
                endpoints.EndpointGuildAnalytics(this.$route.params.guildID, "retention") + "?from=" + this.startDate.toISOString() + "&to=" + this.endDate.toISOString(),
                ({ config }) => {
                    this.isDataFetched = true;
                    this.isDataError = false;

                    this.data = config;

                    this.data.current.retention_cohorts.reverse();
                },
                (error) => {
                    this.$store.dispatch("createToast", getErrorToast(error));

                    this.isDataFetched = true;
                    this.isDataError = true;
                }
            );
        },

        handleDateRangeUpdate(newDateRange) {
            this.startDate = new Date(newDateRange.startDate);
            this.endDate = new Date(newDateRange.endDate);
            this.fetch();
        },

        getMonthLabel(monthNumber) {
            if (monthNumber === 0) {
                return "this month";
            } else if (monthNumber === 1) {
                return "last month";
            } else {
                return monthNumber + " months ago";
            }
        }
    },
};
</script>