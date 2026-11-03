<template>
    <div class="dashboard-container">
        <div class="dashboard-content">
            <div class="flex justify-end min-h-12">
                <AnalyticsDatePicker :modelValue="analyticsDateRange" @update:modelValue="handleDateRangeUpdate"
                    :defaultPreset="'last_7_days'" />
            </div>

            <div v-if="isDataError">
                <div class="mb-4">Data Error</div>
                <button @click="this.fetch">Retry</button>
            </div>
            <div v-else-if="!isDataFetched"
                class="w-full h-full absolute left-0 top-0 flex items-center justify-center">
                <LoadingIcon />
            </div>
            <div v-if="data && !isDataError">
                <div class="grid grid-cols-1 gap-4 mt-2 lg:grid-cols-3 mb-4">
                    <AnalyticsCard name="Members Joined" :amount="data.current.members_joined"
                        :previousAmount="data.previous.members_joined"
                        description="Count of recorded join events in this range." />
                    <AnalyticsCard name="Members Left" :amount="data.current.members_left"
                        :previousAmount="data.previous.members_left" icon="fa-user" :isDecreaseGood="true"
                        description="Count of recorded leave events in this range." />
                    <AnalyticsCard name="Retention Rate" :amount="getRetentionRate(data.current)"
                        :previousAmount="getRetentionComparison(data.current, data.previous)"
                        :amountSuffix="getRetentionSuffix(data.current)"
                        :description="getRetentionDescription(data.current)" />
                </div>
                <p class="text-xs text-gray-500 mb-4">
                    This is retention as of the selected range's end, not after a fixed time from joining. Members who
                    joined later had less time to leave, so changes in join timing can affect comparisons between ranges.
                </p>
                <div class="mb-4 space-y-4">
                    <AnalyticsCardSlot name="Days on Server">
                        <Bar class="h-72" v-if="data.current.time_on_server_before_leaving.length" :data="GetDatasetsLine([
                            GetDatasetBar('#2F80ED', 'Left', data.current.time_on_server_before_leaving),
                            GetDatasetBar('#27AE60', 'Still on server', data.current.time_on_server),
                        ])" :options="ChartOptions({
                            label: (context) => { return `${context.dataset.label} ${context.parsed.x == 0 ? 'on the first day' : 'after ' + context.parsed.x + ' day' + (context.parsed.x !== 1 ? 's' : '')}: ${context.parsed.y} members` }
                        }).WithLegend()" />
                        <p v-else class="text-sm text-gray-500">No data available for this period.</p>
                    </AnalyticsCardSlot>
                    <AnalyticsCardSlot name="Retention Cohorts">
                        <div class="overflow-x-auto">
                            <table class="w-full text-sm text-center border-collapse overflow-x-auto">
                                <thead>
                                    <tr>
                                        <th></th>
                                        <th class="p-2 border border-gray-700"
                                            v-for="(header, headerIndex) in data.current.retention_cohorts"
                                            :key="headerIndex">
                                            Left {{ getMonthLabel(headerIndex) }}
                                        </th>
                                    </tr>
                                </thead>
                                <tbody>
                                    <tr v-for="(row, rowIndex) in data.current.retention_cohorts" :key="rowIndex">
                                        <th class="p-2 border border-gray-700">
                                            Joined {{ getMonthLabel(data.current.retention_cohorts.length - rowIndex -
                                                1) }}
                                        </th>
                                        <td v-for="(cell, cellIndex) in row.splice(0, data.current.retention_cohorts.length - rowIndex)"
                                            :key="cellIndex" class="p-2 border border-gray-700"
                                            :style="{ backgroundColor: `rgba(88, 101, 242, ${cell / 100})` }"
                                            :title="`${cell.toFixed(1)}% of users who joined ${getMonthLabel(data.current.retention_cohorts.length - rowIndex - 1)} left ${getMonthLabel(cellIndex)}`">
                                            {{ cell.toFixed(1) }}%
                                        </td>
                                    </tr>
                                </tbody>
                            </table>
                        </div>
                        <span class="text-xs mt-2">
                            This graph shows the percentage of users who remain on the server before leaving. The x-axis
                            is relative to the end date of the selected time period and shows the information from the
                            previous 12 months.
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
    BarController,
    BarElement,
    Tooltip,
    Legend,
} from "chart.js";

import { Bar } from "vue-chartjs";
import { ChartOptions, GetDatasetsLine, GetDatasetBar } from "@/components/dashboard/analytics/analytics";
import analyticsView from "@/components/dashboard/analytics/analyticsView";

export default {
    mixins: [analyticsView],
    data() {
        return {
            analyticsEndpoint: "retention",
        };
    },
    components: { AnalyticsCard, AnalyticsCardSlot, AnalyticsCardChange, AnalyticsDatePicker, LoadingIcon, Bar },
    setup() {
        ChartJS.register(CategoryScale, LinearScale, BarController, BarElement, Tooltip, Legend);
        return { ChartOptions, GetDatasetsLine, GetDatasetBar };
    },
    methods: {
        getRetentionRate(current) {
            return current.retention_percentage == null ? "-" : current.retention_percentage;
        },
        getRetentionSuffix(current) {
            return current.retention_percentage == null ? "" : "%";
        },
        getRetentionComparison(current, previous) {
            if (current.retention_percentage == null || previous.retention_percentage == null) {
                return null;
            }
            return previous.retention_percentage;
        },
        getRetentionDescription(current) {
            const cohortSize = current.retention_cohort_size;
            const retainedMembers = current.retained_members;
            if (!Number.isInteger(cohortSize) || !Number.isInteger(retainedMembers)) {
                return "Unique members who joined during this range and had no recorded leave by the range end.";
            }
            if (cohortSize === 0) {
                return "No unique members joined during this range.";
            }
            return `${retainedMembers} of ${cohortSize} unique members who joined during this range had no recorded leave by the range end.`;
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