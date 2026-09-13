<template>
    <div>
        <span v-if="$props.amount > $props.previousAmount" :class="[$props.isDecreaseGood ? 'text-trendDown' : 'text-trendUp', 'font-semibold']">
            <font-awesome-icon icon="arrow-trend-up" class="mr-1" />
            +{{ (($props.amount / $props.previousAmount * 100) - 100).toFixed(2) }}%
        </span>
        <span v-else-if="$props.amount < $props.previousAmount" :class="[$props.isDecreaseGood ? 'text-trendUp' : 'text-trendDown', 'font-semibold']">
            <font-awesome-icon icon="arrow-trend-down" class="mr-1" />
            -{{ (($props.previousAmount / $props.amount * 100) - 100).toFixed(2) }}%
        </span>
        <span v-else>
            No change
        </span>
        vs last period
        
    </div>
</template>

<script>
export default {
    props: {
        amount: {
            type: Number,
            required: true
        },
        previousAmount: {
            type: Number,
            required: true
        },
        isDecreaseGood: {
            type: Boolean,
            default: false
        }
    },
};
</script>