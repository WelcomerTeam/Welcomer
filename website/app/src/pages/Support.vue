<template>
  <div class="relative min-h-screen">
    <Header />

    <main>
      <div id="overview" class="relative bg-secondary">
        <div class="px-6 py-12 bg-secondary w-full max-w-7xl mx-auto">
          <h1 class="text-3xl font-bold text-left text-white tracking-tight">
            Support
          </h1>
        </div>
      </div>

      <div id="plans">
        <div class="bg-white text-neutral-900">
          <div class="hero-preview">
            <div class="px-4 pt-8 mx-auto max-w-7xl sm:px-6">
              <div class="sm:flex sm:flex-col sm:align-center">
                <div class="prose-lg text-center">
                  <h1
                    class="font-black leading-8 tracking-tight text-gray-900"
                  >
                    Title
                  </h1>
                  <span
                    class="mt-3 text-lg text-gray-500 section-subtitle max-w-prose mx-auto"
                  >
                    Subheading
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div>
        <div class="bg-donate">
          <div class="hero-features">
            <div class="mx-4 my-12 lg:grid lg:grid-cols-3 lg:gap-8">
              Hello World
            </div>
          </div>
        </div>
      </div>
    </main>

    <div class="footer-anchor">
      <Footer />
    </div>
  </div>
</template>

<script>
import Header from "@/components/Header.vue";
import Footer from "@/components/Footer.vue";

import { ref } from "vue";
import {
  Disclosure,
  DisclosureButton,
  DisclosurePanel,
  RadioGroup,
  RadioGroupDescription,
  RadioGroupLabel,
  RadioGroupOption,
} from "@headlessui/vue";

import { CheckIcon } from "@heroicons/vue/solid";
import { ChevronDownIcon } from "@heroicons/vue/outline";

const features = [
  {
    name: "Animated Welcomer Backgrounds",
    icon: "photo-film",
    description:
      "Show off your awesome animated backgrounds to users who join, whatever it is. Except when it's rickroll...",
  },
  {
    name: "Time Roles",
    icon: "user-clock",
    description:
      "Sometimes you don't want to give users a role immediately. Use timeroles to give them roles automatically when the time comes, it could be 10 minutes or in a year.",
  },
  {
    name: "Dedicated Donator Bot",
    icon: "plug-circle-bolt",
    description: "Run Welcomer on its own donator-only bot account.",
  },
  {
    name: "Whitelabelled Bot",
    icon: "plug-circle-plus",
    soon: true,
    description:
      "Run Welcomer on its own unique bot account with a fully customisable username and avatar, with the same uptime and reliability.",
  },
];

const checklist = [
  "Custom Backgrounds",
  "Dedicated Donator Bot",
  "Flexible plans",
  "No recurring payments",
];

// Data below will be fetched from API

const currency = "£";

const durations = [
  {
    name: "Monthly",
    months: 1,
    multiplier: 1,
  },
  {
    name: "Annually",
    months: 12,
    multiplier: 0.8,
  },
  {
    name: "Patreon",
    months: 1,
    multiplier: 1,
    isPatreon: true,
  },
];

const plans = [
  {
    name: "Welcomer x1",
    price: 5,
    footer: "Welcomer Pro for 1 server.",
    patreonPrice: 5,
    patreonCheckout: 3744919,
  },
  {
    name: "Welcomer x3",
    price: 10,
    footer: "Welcomer Pro for 3 servers.",
    patreonPrice: 10,
    patreonCheckout: 3744921,
  },
  {
    name: "Welcomer x5",
    price: 15,
    footer: "Welcomer Pro for 5 servers.",
    patreonPrice: 15,
    patreonCheckout: 3744926,
  },
];

const faqs = [
  {
    question: "TODO",
    answer: "TODO",
  },
];

const customBackgroundPrice = 5;
const fromPrice = 5;

export default {
  components: {
    Disclosure,
    DisclosureButton,
    DisclosurePanel,
    RadioGroup,
    RadioGroupDescription,
    RadioGroupLabel,
    RadioGroupOption,

    CheckIcon,
    ChevronDownIcon,

    Header,
    Footer,
  },
  setup() {
    const durationSelected = ref(durations[0]);
    const planSelected = ref(plans[0]);

    return {
      features,
      checklist,

      durationSelected,
      durations,

      planSelected,
      plans,

      customBackgroundPrice,
      currency,
      fromPrice,

      faqs,
    };
  },
  methods: {
    selectPlan(plan) {
      this.planSelected = plan;
      this.handleClick();
    },
    selectDuration(duration) {
      this.durationSelected = duration;
    },
    handleCustomBackgroundClick() {
      alert("handle cbg");
      // Open the donate page
    },

    handleClick() {
      if (this.durationSelected.isPatreon) {
        return window.open(
          `https://www.patreon.com/join/Welcomer/checkout?rid=${this.planSelected.patreonCheckout}`,
          "_blank"
        );
      }

      alert(`handle ${this.durationSelected.name} ${this.planSelected.name}`);
      // Open the donate page
    },
  },
};
</script>
