<template>
    <div class="dashboard-container">
        <div class="dashboard-content">
            <div class="flex justify-end min-h-12">
                <AnalyticsDatePicker :modelValue="'last_7_days'" @update:modelValue="handleDateRangeUpdate"
                    :maxDate="new Date()" :defaultPreset="'last_7_days'" />
            </div>

            <div v-if="isDataError">
                <div class="mb-4">Data Error</div>
                <button @click="this.fetch">Retry</button>
            </div>
            <div v-else-if="!isDataFetched"
                class="w-full h-full absolute left-0 top-0 flex items-center justify-center">
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
                        <div>
                            <Line :data="GetDatasetsLine([
                                GetDatasetLine('#2F80ED', 'Members', data.current.guild_members).HasTimestamp().WithGradientBackgroundColor(),
                                GetDatasetLine('#eeeeee', 'Previous', data.previous.guild_members).HasTimestamp().WithBorderDash([5, 5]).WithBorderWidth(1),
                            ]).DontBeginAtZero()" :options="ChartOptions({
                                label: (context) => { return `${context.dataset.label}: ${context.parsed.y} members` }
                            })" />
                        </div>
                    </AnalyticsCardSlot>

                    <AnalyticsCardSlot name="Join/Leave">
                        <div>
                            <Line :data="GetDatasetsLine([
                                GetDatasetLine('#2F80ED', 'Sum', data.current.guild_net).HasTimestamp(),
                                GetDatasetBar('#43B581', 'Joins', data.current.guild_joins).HasTimestamp().Ungroup(),
                                GetDatasetBar('#F04747', 'Leaves', data.current.guild_leaves).HasTimestamp().Ungroup(),
                            ])" :options="ChartOptions({
                                label: (context) => { return `${context.dataset.label}: ${Math.abs(context.parsed.y)} members` }
                            })" />
                        </div>
                    </AnalyticsCardSlot>
                </div>
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
    mounted() {
        this.fetch();
    },
};
</script>