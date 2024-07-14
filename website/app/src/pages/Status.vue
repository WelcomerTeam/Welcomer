<template>
  <div class="relative min-h-screen">
    <Header />

    <main>
      <div class="relative bg-secondary text-white px-6 py-12 w-full max-w-7xl mx-auto">
        <h1 class="text-3xl font-bold text-left tracking-tight">
          Status
        </h1>
      </div>

      <div class="bg-white text-neutral-900 pb-32">
        <div class="hero-preview">
          <div class="px-4 mx-auto max-w-7xl sm:px-6">
            <div v-if="this.isDataError">
              <div class="mb-4">Data Error</div>
              <button @click="this.fetchStatus">Retry</button>
            </div>
            <div v-else-if="!this.isDataFetched" class="flex py-5 w-full justify-center">
              <LoadingIcon />
            </div>
            <div v-else v-for="manager in status" :key="manager.id">
              <p class="text-xl font-semibold text-left tracking-tight text-gray-900">
                {{ manager.name }}
              </p>
              <p>
                <b>Guilds:</b> {{ getTotalGuildsForManager(manager).toLocaleString() }}<br />
                <b>Available Shards:</b> {{ getAvailableShardsForManager(manager).toLocaleString() }}/{{ manager.shards.length.toLocaleString() }}
              </p>  
              <div class="flex gap-2 flex-wrap mb-16 mt-4">
                <button :class="['w-10 h-10 flex rounded-md items-center justify-center text-sm font-bold focus:ring-2 relative group', getStyleForShard(shard)]" v-for="shard in manager.shards" :key="shard.shard_id">
                  <span>{{ shard.shard_id }}</span>
                  <div class="hidden group-hover:block group-focus:block absolute w-36 p-4 rounded-md bg-secondary text-white z-10 -translate-x-1/2 left-1/2 top-full translate-y-1 text-xs">
                    Guilds: {{ shard.guilds.toLocaleString() }}<br/>
                    Latency: {{ shard.latency.toLocaleString() }}ms<br/>
                    Uptime: {{ formatSeconds(shard.uptime) }}
                  </div>
                </button>
              </div>
            </div>
            <div class="flex flex-wrap gap-2 mt-8">
              <span :class="['font-bold text-sm px-2 py-1 rounded-md', getStyleForShard({ status: 0 })]">Idle</span>
              <span :class="['font-bold text-sm px-2 py-1 rounded-md', getStyleForShard({ status: 1 })]">Connecting</span>
              <span :class="['font-bold text-sm px-2 py-1 rounded-md', getStyleForShard({ status: 2 })]">Connected</span>
              <span :class="['font-bold text-sm px-2 py-1 rounded-md', getStyleForShard({ status: 3 })]">Ready</span>
              <span :class="['font-bold text-sm px-2 py-1 rounded-md', getStyleForShard({ status: 4 })]">Reconnecting</span>
              <span :class="['font-bold text-sm px-2 py-1 rounded-md', getStyleForShard({ status: 5 })]">Closing</span>
              <span :class="['font-bold text-sm px-2 py-1 rounded-md', getStyleForShard({ status: 6 })]">Closed</span>
              <span :class="['font-bold text-sm px-2 py-1 rounded-md', getStyleForShard({ status: 7 })]">Erroring</span>
            </div>
          </div>
        </div>
      </div>
    </main>

    <Toast />

    <div class="footer-anchor">
      <Footer />
    </div>
  </div>
</template>

<style lang="scss"></style>

<script>
import { ref } from "vue";

import Header from "@/components/Header.vue";
import Footer from "@/components/Footer.vue";
import Toast from "@/components/dashboard/Toast.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";

import { getErrorToast } from "@/utilities";

import dashboardAPI from "@/api/dashboard";

export default {
  components: {
    Header,
    Footer,
    Toast,
    LoadingIcon,
  },
  setup() {
    let isDataFetched = ref(false);
    let isDataError = ref(false);
    let status = ref({});
    let interval = 0;

    return {
      isDataFetched,
      isDataError,
      status,
      interval
    };
  },
  mounted() {
    this.fetchStatus(true);

    this.interval = setInterval(() => {
      this.fetchStatus(false);
    }, 10000);
  },
  unmounted() {
    clearInterval(this.interval);
  },
  methods: {
    fetchStatus(force) {
      if (!force) {
        // Check if we don't have focus. If prevent and false, do not fetch the status.
        try {
          if (!document.hasFocus()) {
            return
          }
        } catch (e) {
          return
        }
      }

      // this.isDataFetched = false;
      this.isDataError = false;

      dashboardAPI.getStatus(
        ({ managers }) => {
          this.status = managers
            .filter(manager => manager.shards.length > 0)
            .sort((a, b) => b.shards.length - a.shards.length);
          this.isDataFetched = true;
          this.isDataError = false;
        },
        (error) => {
          this.$store.dispatch("createToast", getErrorToast(error));

          this.isDataFetched = true;
          this.isDataError = true;
        }
      )
    },

    formatSeconds(seconds) {
      const minutes = Math.floor(seconds / 60);
      const hours = Math.floor(minutes / 60);
      const days = Math.floor(hours / 24);

      if (days > 0) {
        return `${days} day${days > 1 ? 's' : ''}`;
      } else if (hours > 0) {
        return `${hours} hour${hours > 1 ? 's' : ''}`;
      } else if (minutes > 0) {
        return `${minutes} minute${minutes > 1 ? 's' : ''}`;
      } else {
        return `${seconds} second${seconds > 1 ? 's' : ''}`;
      }
    },

    getTotalGuildsForManager(manager) {
      return manager.shards.reduce((acc, shard) => acc + shard.guilds, 0);
    },

    getAvailableShardsForManager(manager) {
      return manager.shards.filter(shard => shard.status === 3 || shard.status === 4).length;
    },

    getStyleForShard(shard) {
      if (shard.latency < 0) {
        shard.status = 4;
      }

      switch (shard.status) {
        case 1: // connecting
          return "bg-fuchsia-100 text-fuchsia-700 ring-fuchsia-700"
        case 2: // connected
          return "bg-fuchsia-200 text-fuchsia-800 ring-fuchsia-800"
        case 3: // ready
          return "bg-emerald-200 text-emerald-800 ring-emerald-800"
        case 4: // reconnecting
          return "bg-emerald-100 text-emerald-700 ring-emerald-700"
        case 5: // closing
          return "bg-amber-200 text-amber-800 ring-amber-800"
        case 6: // closed
          return "bg-amber-100 text-amber-700 ring-amber-700"
        case 7: // erroring
          return "bg-red-200 text-red-800 ring-red-800"
        default:
          return "bg-gray-200 text-gray-800 ring-gray-800"
      }
    }
  },
};
</script>
