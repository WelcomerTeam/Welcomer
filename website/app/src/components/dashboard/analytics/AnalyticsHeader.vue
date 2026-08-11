<template>
    <div class="flex flex-col sm:flex-row sm:flex-wrap bg-gray-200 dark:bg-secondary-dark">
        <div @click="selectTab(tabName)" v-for="tabName in AnalyticsTabs" :key="tabName"
            :class="['cursor-pointer border-b-2 px-6 py-4 flex-shrink-0', isActiveTab(tabName) ? 'border-primary text-black dark:text-white' : 'hover:border-primary hover:text-black hover:dark:text-white hover:bg-inherit dark:text-gray-400 border-transparent']">
            {{ tabName }}
        </div>
    </div>
</template>

<script>
import { AnalyticsTabs } from "@/constants";

export default {
    props: {
        tab: {
            type: String,
            required: true,
        },
    },
    setup() {
        return {
            AnalyticsTabs,
        };
    },
    emits: ["update:tab"],
    methods: {
        isActiveTab(tabName) {
            return this.tab === tabName;
        },

        selectTab(tabName) {
            this.$emit("update:tab", tabName);
        },
    },
};
</script>