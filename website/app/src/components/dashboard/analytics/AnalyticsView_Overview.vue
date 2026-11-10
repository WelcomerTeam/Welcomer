<template>
    <div class="dashboard-container">
        <div class="dashboard-content">
            <div class="flex justify-end min-h-12">
                <AnalyticsDatePicker :modelValue="analyticsDateRange" @update:modelValue="handleDateRangeUpdate"
                    :maxDate="new Date()" :defaultPreset="'last_7_days'" />
            </div>

            <div v-if="isDataError">
                <div class="mb-4">Data Error</div>
                <button @click="this.fetch">Retry</button>
            </div>
            <div v-else-if="!isDataFetched"
                class="pointer-events-none w-full h-full absolute left-0 top-0 flex items-center justify-center">
                <LoadingIcon />
            </div>
            <template v-if="data && !isDataError">
                <div class="grid grid-cols-1 gap-4 mt-2 lg:grid-cols-3 mb-4">
                    <AnalyticsCard name="Total Guild Members" :amount="data.current.total_guild_members"
                        :previousAmount="data.previous.total_guild_members" icon="fa-user" />
                    <AnalyticsCard name="Members Joined" :amount="data.current.members_joined"
                        :previousAmount="data.previous.members_joined" />
                    <AnalyticsCard name="Members Left" :amount="data.current.members_left"
                        :previousAmount="data.previous.members_left" :isDecreaseGood="true" />
                </div>
                <div class="mb-4 space-y-4">
                    <AnalyticsCardSlot name="Guild Members">
                        <Line class="h-72" :data="GetDatasetsLine([
                            GetDatasetLine('#2F80ED', 'Members', data.current.guild_members).HasTimestamp().WithGradientBackgroundColor(),
                            GetDatasetLine('#eeeeee', 'Previous', data.previous.guild_members).HasTimestamp().WithBorderDash([5, 5]).WithBorderWidth(1),
                        ]).DontBeginAtZero()" :options="ChartOptions({
                            label: (context) => { return `${context.dataset.label}: ${context.parsed.y} members` }
                        })" />
                    </AnalyticsCardSlot>

                    <AnalyticsCardSlot name="Join/Leave">
                        <Line class="h-72" :data="GetDatasetsLine([
                            GetDatasetLine('#2F80ED', 'Sum', data.current.guild_net).HasTimestamp(),
                            GetDatasetBar('#43B581', 'Joins', data.current.guild_joins).HasTimestamp().Ungroup(),
                            GetDatasetBar('#F04747', 'Leaves', data.current.guild_leaves).HasTimestamp().Ungroup(),
                        ])" :options="ChartOptions({
                                label: (context) => { return `${context.dataset.label}: ${Math.abs(context.parsed.y)} members` }
                            })" />
                    </AnalyticsCardSlot>
                </div>

                <AnalyticsCardSlot v-if="data.benchmarks" name="Guild metrics vs benchmark">
                    <p class="mb-4 text-sm text-gray-600 dark:text-gray-300">
                        Comparison of this guild's metrics against the {{ benchmarkStatisticLabel }} of guilds with similar member counts.
                    </p>
                    <div class="grid grid-cols-1 gap-4 xl:grid-cols-2">
                        <div v-for="metric in data.benchmarks.metrics" :key="metric.key"
                            class="rounded border border-gray-200 p-3 dark:border-secondary-light">
                            <div class="mb-2 flex items-start justify-between gap-3">
                                <span class="font-medium">{{ metric.label }}</span>
                                <span class="text-right text-sm text-gray-600 dark:text-gray-300">
                                    Server: {{ formatBenchmarkValue(metric.server_value, metric.unit) }}
                                    <span class="mx-1">·</span>
                                    {{ benchmarkStatisticLabel }}: {{ formatBenchmarkValue(metric.benchmark_value, metric.unit) }}
                                </span>
                            </div>
                            <div class="space-y-2" role="img"
                                :aria-label="`${metric.label}. Server: ${formatBenchmarkValue(metric.server_value, metric.unit)}. ${benchmarkStatisticLabel}: ${formatBenchmarkValue(metric.benchmark_value, metric.unit)}.`">
                                <div class="flex items-center gap-2">
                                    <span class="w-16 shrink-0 text-xs">Server</span>
                                    <div class="h-2 flex-1 overflow-hidden rounded bg-gray-200 dark:bg-secondary-light">
                                        <div class="h-full rounded bg-blue-500"
                                            :style="{ width: benchmarkBarWidth(metric.server_value, metric) }"></div>
                                    </div>
                                </div>
                                <div class="flex items-center gap-2">
                                    <span class="w-16 shrink-0 text-xs">{{ data.benchmarks.statistic === "p90" ? "P90" : "Median" }}</span>
                                    <div class="h-2 flex-1 overflow-hidden rounded bg-gray-200 dark:bg-secondary-light">
                                        <div class="h-full rounded bg-green-500"
                                            :style="{ width: benchmarkBarWidth(metric.benchmark_value, metric) }"></div>
                                    </div>
                                </div>
                            </div>
                        </div>
                    </div>
                </AnalyticsCardSlot>
            </template>
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
    Tooltip,
    BarController,
    BarElement,
} from "chart.js";

import { Line } from "vue-chartjs";

import { ChartOptions, GetDatasetsLine, GetDatasetLine, GetDatasetBar } from "@/components/dashboard/analytics/analytics";

import analyticsView from "@/components/dashboard/analytics/analyticsView";

export default {
    mixins: [analyticsView],
    data() {
        return {
            analyticsEndpoint: "overview",
        };
    },
    components: { AnalyticsCard, AnalyticsCardSlot, AnalyticsCardChange, AnalyticsDatePicker, LoadingIcon, Line },
    setup() {
        ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, BarController, BarElement);
        return { ChartOptions, GetDatasetsLine, GetDatasetLine, GetDatasetBar };
    },
    computed: {
        benchmarkStatisticLabel() {
            return this.data?.benchmarks?.statistic === "p90" ? "90th percentile" : "Median";
        },
    },
    methods: {
        formatBenchmarkValue(value, unit) {
            if (value === null || value === undefined || !Number.isFinite(value)) return "-";
            if (unit === "percent") return `${value.toFixed(1)}%`;
            if (unit === "users") return Intl.NumberFormat().format(Math.round(value));
            if (unit === "messages_per_user") return value.toFixed(1);
            if (unit === "days") return `${(value / 365.25).toFixed(1)} years`;
            if (unit === "milliseconds") return this.formatDuration(value / 1000);
            if (unit === "seconds") return this.formatDuration(value);
            return Intl.NumberFormat().format(value);
        },
        formatDuration(seconds) {
            if (seconds < 60) return `${Math.round(seconds)}s`;
            if (seconds < 3600) return `${Math.round(seconds / 60)}m`;
            if (seconds < 86400) return `${(seconds / 3600).toFixed(1)}h`;
            return `${(seconds / 86400).toFixed(1)}d`;
        },
        benchmarkBarWidth(value, metric) {
            if (value === null || value === undefined || !Number.isFinite(value) || value <= 0) return "0%";
            const maxValue = Math.max(metric.server_value || 0, metric.benchmark_value || 0);
            if (maxValue <= 0) return "0%";
            return `${Math.min(100, (value / maxValue) * 100)}%`;
        },
    },
    mounted() {
        this.fetch();
    },
};
</script>