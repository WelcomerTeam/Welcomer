<template>
    <div class="dashboard-container">
        <div class="dashboard-content">
            <div class="grid grid-cols-1 gap-4 mt-2 lg:grid-cols-3 mb-4">
                <AnalyticsCard name="Total Guild Members" :amount="9564" :previousAmount="9542" icon="fa-user" />
                <AnalyticsCard name="Members Joined" :amount="123" :previousAmount="123" />
                <AnalyticsCard name="Members Left" :amount="123" />
            </div>
            <div class="mb-4">
                <!-- <AnalyticsCard name="Total Messages Sent" :amount="123456" :previousAmount="123456" icon="fa-message" /> -->
                <AnalyticsCardSlot name="Test">
                    <Line :data="GetDatasetLine('Members', [
                        ['A', 1],
                        ['B',6],
                        ['C',3],
                    ])" :options="ChartOptions({
                        label: (context) => { return `${context.parsed.y} members` }
                    })" />
                </AnalyticsCardSlot>
            </div>
        </div>
    </div>
</template>

<script>
import AnalyticsCard from "@/components/dashboard/analytics/AnalyticsCard.vue";
import AnalyticsCardSlot from "@/components/dashboard/analytics/AnalyticsCardSlot.vue";

import {
    Chart as ChartJS,
    CategoryScale,
    LinearScale,
    PointElement,
    LineElement,
    Tooltip,
    Filler,
} from "chart.js";

import { Line } from "vue-chartjs";

import { ChartOptions, GetDatasetLine } from "@/constants";

export default {
    components: {
        AnalyticsCard,
        AnalyticsCardSlot,
        Line,
    },
    setup() {
        ChartJS.register(CategoryScale, LinearScale, PointElement, LineElement, Tooltip, Filler);

        return {
            ChartOptions,
            GetDatasetLine,
        };
    },
};
</script>