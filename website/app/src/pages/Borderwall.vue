<template>
  <div class="relative min-h-screen">
    <Header />

    <main>
      <div class="hero-preview">
        <div class="px-4 mx-auto max-w-7xl sm:px-6">
          <div class="sm:flex sm:flex-col sm:align-center">
            <div class="prose-lg text-center">
              <img src="/assets/peek.png" alt="" class="mx-auto w-24 h-24 mb-4 select-none" />
              <h1 class="font-black leading-8 tracking-tight">
                This server is protected by Borderwall
              </h1>
              <div v-if="this.$route.params.key == ''">
                Missing Key
              </div>
              <div v-else-if="this.isDataError">
                <div class="mb-4">Data Error</div>
                <button @click="this.fetchBorderwall">Retry</button>
              </div>
              <div v-else-if="!this.isDataFetched" class="flex py-5 w-full justify-center">
                <LoadingIcon />
              </div>
              <span v-else-if="this.isValidKey" class="mt-3 text-lg section-subtitle max-w-prose mx-auto">
                You are verifying for <b> {{ guildName }} </b>. Please verify below.
              </span>
              <span v-else class="mt-3 text-lg section-subtitle max-w-prose mx-auto">
                Your BorderWall link has expired or already been used.
              </span>
            </div>
          </div>
        </div>
      </div>

      <div class="pb-48 max-w-7xl mx-auto px-6 lg:px-7">
        <div v-if="this.isDataFetched && this.isValidKey"
          :class="['text-white px-6 py-8 rounded-lg p-4 mb-4 text-center shadow-sm transition-all duration-500 min-h-52 flex items-center justify-center', this.isCompleted ? 'bg-green-600' : 'bg-secondary dark:bg-secondary-dark']">
          <div v-if="this.isCompleted" class="text-center space-y-4">
            <font-awesome-icon icon="fa-sharp fa-light fa-badge-check" class="W-24 h-24" aria-hidden="" />
            <h2 class="text-2xl font-semibold">
              You have been verified. You can now close this tab.
            </h2>
          </div>
          <div v-else class="space-y-4">
            <button @click="execute" :disabled="!this.isDataFetched && !this.isValidKey"
              class="cta-button bg-primary hover:bg-primary-dark w-full max-w-xl">
              <LoadingIcon v-if="this.isExecuting" class="mr-3" />
              Verify
            </button>

            <recaptcha ref="recaptcha" action="borderwall" @verify="verify" />

            <p>
              This site is protected by reCAPTCHA and the Google
              <a href="https://policies.google.com/privacy" target="_blank"
                class="font-semibold underline hover:text-gray-300">Privacy Policy</a>
              and
              <a href="https://policies.google.com/terms" target="_blank"
                class="font-semibold underline hover:text-gray-300">Terms of Service</a>
              apply.
            </p>
          </div>
        </div>
        <div class="text-center font-semibold">
          Welcomer or Borderwall will never ask you to scan any QR codes. <a href="https://welcomer.gg/phishing"
            target="_blank" class="font-semibold underline hover:text-gray-700 dark:hover:text-gray-300">Learn More</a>.
        </div>
      </div>
    </main>

    <Toast />

    <div class="footer-anchor">
      <Footer />
    </div>
  </div>
</template>

<script>
import { ref } from "vue";

import Header from "@/components/Header.vue";
import Footer from "@/components/Footer.vue";
import Toast from "@/components/dashboard/Toast.vue";
import Recaptcha from "@/components/Recaptcha.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";

import dashboardAPI from "@/api/dashboard";

import { getErrorToast } from "@/utilities";

export default {
  components: {
    Header,
    Footer,
    Toast,
    Recaptcha,
    LoadingIcon,
  },
  setup() {
    let isDataFetched = ref(false);
    let isDataError = ref(false);

    let isCompleted = ref(false);
    let isExecuting = ref(false);
    let isValidKey = ref(false);
    let guildName = ref("");
    let response = ref(null);

    return {
      isDataFetched,
      isDataError,
      isCompleted,
      isExecuting,
      isValidKey,
      guildName,
      response,
    };
  },

  mounted() {
    if (this.$route.params.key != "") {
      this.fetchBorderwall();
    }
  },

  methods: {
    fetchBorderwall() {
      this.isDataFetched = false;
      this.isDataError = false;

      dashboardAPI.getBorderwall(
        this.$route.params.key,
        (response) => {
          this.isDataFetched = true;
          this.isValidKey = response.valid;
          this.guildName = response.guild_name;
        },
        (error) => {
          this.$store.dispatch("createToast", getErrorToast(error));

          this.isDataFetched = true;
          this.isDataError = true;
        }
      );
    },

    async getPlatformVersion() {
      try {
        let entropyValues = await navigator.userAgentData.getHighEntropyValues(["platformVersion"]);
        return entropyValues.platformVersion;
      } catch (error) {
        return undefined;
      }
    },

    async sendResponse() {
      this.isExecuting = true;

      dashboardAPI.submitBorderwall(
        this.$route.params.key,
        {
          response: this.response,
          platform_version: await this.getPlatformVersion(),
        },
        () => {
          this.isCompleted = true;
        },
        (error) => {
          this.$store.dispatch("createToast", getErrorToast(error));
          this.isExecuting = false;
        }
      );
    },

    execute() {
      this.isExecuting = true;
      this.$refs.recaptcha.execute();
    },

    verify(response) {
      this.response = response;
      if (this.isExecuting) {
        this.sendResponse();
      }
    },
  },
};
</script>
