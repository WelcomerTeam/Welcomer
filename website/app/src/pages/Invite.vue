<template>
</template>

<script>
import {
  GetBotInvite,
  PrimaryBotId
} from "@/constants";

export default {
  setup() {
    window.location.replace(GetBotInvite(PrimaryBotId, null));
  },
}
</script>