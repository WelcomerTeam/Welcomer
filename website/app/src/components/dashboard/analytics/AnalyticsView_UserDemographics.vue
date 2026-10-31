<template>
    <div class="dashboard-container">
        <div class="dashboard-content">
            <div v-if="isDataError" class="mb-4">
                <p class="mb-2">Unable to load member demographics.</p>
                <button class="text-blue-500 hover:underline" @click="fetch">Retry</button>
            </div>
            <div v-else-if="!isDataFetched"
                class="w-full h-full absolute left-0 top-0 flex items-center justify-center">
                <LoadingIcon />
            </div>
            <template v-else-if="data && !isDataError">
                <p class="text-sm text-gray-500 mb-4">Based on current guild members. Account creation dates are derived
                    from Discord user IDs.</p>
                <div class="grid grid-cols-1 gap-4 mt-2 lg:grid-cols-3 mb-4">
                    <AnalyticsCard name="Average Account Age" :amount="formatAge(data.current.average_account_age_days)"
                        icon="fa-user" />
                    <AnalyticsCard name="Median Account Age"
                        :amount="formatAge(data.current.median_account_age_days)" />
                    <AnalyticsCard name="90th Percentile Account Age"
                        :amount="formatAge(data.current.p90_account_age_days)" />
                </div>
                <div class="mb-4">
                    <AnalyticsCardSlot name="Account Age Distribution">
                        <Bar class="h-72" v-if="data.current.account_age_distribution.length" :data="GetDatasetsLine([
                            GetDatasetBar('#2F80ED', 'Members', data.current.account_age_distribution),
                        ])" :options="distributionOptions" />
                        <p v-else class="text-sm text-gray-500">No data available for this period.</p>
                    </AnalyticsCardSlot>
                </div>
                <div class="mb-4">
                    <AnalyticsCardSlot name="Account Age vs Join Date">
                        <div class="h-72">
                            <Scatter :data="scatterData(data.current.account_age_vs_join_date)"
                                :options="scatterOptions" />
                        </div>
                    </AnalyticsCardSlot>
                    <p class="text-xs text-gray-500 mt-2">The scatter plot samples at most 1,000 members; summary
                        statistics and age buckets use all returned current members.</p>
                </div>
            </template>
        </div>
    </div>
</template>

<script>
import { Chart as ChartJS, CategoryScale, LinearScale, PointElement, BarController, BarElement, ScatterController, Tooltip } from "chart.js";
import { Bar, Scatter } from "vue-chartjs";
import AnalyticsCard from "@/components/dashboard/analytics/AnalyticsCard.vue";
import AnalyticsCardSlot from "@/components/dashboard/analytics/AnalyticsCardSlot.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";
import analyticsView from "@/components/dashboard/analytics/analyticsView";
import { GetDatasetsLine, GetDatasetBar } from "@/components/dashboard/analytics/analytics";

export default {
    mixins: [analyticsView],
    data() {
        return {
            analyticsEndpoint: "user-demographics",
            distributionOptions: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    legend: { display: false },
                    tooltip: {
                        callbacks: {
                            label: (context) => `${context.parsed.y} members`,
                        },
                    },
                },
                scales: {
                    x: {
                        title: { display: true, text: "Account age (months)" },
                        ticks: {
                            autoSkip: true,
                            maxTicksLimit: 24,
                        },
                    },
                    y: {
                        beginAtZero: true,
                        title: { display: true, text: "Members" },
                    },
                },
            },
            scatterOptions: {
                responsive: true,
                maintainAspectRatio: false,
                plugins: {
                    tooltip: {
                        callbacks: {
                            label: (context) => `${new Date(context.raw.x).toLocaleDateString()}: ${context.raw.y.toLocaleString()} days old`,
                        },
                    },
                },
                scales: {
                    x: {
                        type: "linear",
                        title: { display: true, text: "Joined date" },
                        ticks: { callback: (value) => new Date(value).toLocaleDateString() },
                    },
                    y: {
                        title: { display: true, text: "Account age (days)" },
                    },
                },
            },
        };
    },
    components: { AnalyticsCard, AnalyticsCardSlot, LoadingIcon, Bar, Scatter },
    setup() {
        ChartJS.register(CategoryScale, LinearScale, PointElement, Tooltip, BarController, BarElement, ScatterController);
        return { GetDatasetsLine, GetDatasetBar };
    },
    methods: {
        formatAge(days) {
            days = Math.max(0, Number(days) || 0);
            if (days < 365) return `${days} days`;
            const years = Math.floor(days / 365);
            const months = Math.floor((days % 365) / 30);
            return months > 0 ? `${years}y ${months}m` : `${years} years`;
        },
        scatterData(points) {
            return {
                datasets: [{
                    label: "Members",
                    data: points.map((point) => ({
                        x: new Date(point.joined_at).getTime(),
                        y: point.account_age_days,
                    })),
                    backgroundColor: "#2F80ED",
                    pointRadius: 3,
                }],
            };
        },
    },
};
</script>
