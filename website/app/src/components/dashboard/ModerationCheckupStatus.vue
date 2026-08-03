<template>
    <div v-if="$props.moderationCheckupStatus == 1">
        <!-- Pending -->
        <p class="border ring-yellow-500 border-yellow-500 p-4 rounded-md shadow-sm">
            This message is currently being reviewed. It will be approved or rejected shortly.
        </p>
    </div>
    <div v-else-if="$props.moderationCheckupStatus == 2">
        <!-- Approved -->
    </div>
    <div v-else-if="$props.moderationCheckupStatus == 3">
        <!-- Rejected -->
        <p class="border ring-red-500 border-red-500 p-4 rounded-md shadow-sm">
            This message potentially violates our guidelines and will not be shown to users. Please consult our Acceptable Use policy <a href="/terms" target="_blank" class="underline">here</a>.
            If you believe this is a mistake, please let us know on our <a href="/support" target="_blank" class="underline">support server</a>.
        </p>
    </div>
</template>

<script>
export default {
    props: {
        moderationCheckupStatus: {
            type: Number,
            required: true
        }
    }
}
</script>