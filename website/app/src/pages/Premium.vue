<template>
  <div class="relative min-h-screen">
    <Header />

    <main>
      <div id="overview" class="relative bg-secondary">
        <div class="absolute inset-0" aria-hidden="true">
          <div class="absolute inset-y-0 right-0 w-1/2 bg-primary" />
        </div>
        <div class="relative mx-auto max-w-7xl lg:px-6 lg:grid lg:grid-cols-2">
          <div class="px-6 pt-6 pb-12 bg-secondary lg:pt-12 lg:px-0 lg:pr-6">
            <div class="max-w-lg mx-auto lg:mx-0">
              <h1 class="text-base font-semibold tracking-wide text-primary">
                Welcomer Pro
              </h1>
              <h2 class="text-3xl font-bold text-left text-white flex justify-center tracking-tight">
                Everything you need to boost your server's engagement
              </h2>
              <div class="mt-12 space-y-8">
                <div v-for="feature in features" :key="feature.name" class="relative">
                  <dt>
                    <div class="absolute flex items-center justify-center w-12 h-12 rounded-md bg-secondary-light">
                      <font-awesome-icon :icon="feature.icon" class="w-6 h-6 text-white" aria-hidden="true" />
                    </div>
                    <p class="ml-16 text-xl font-semibold leading-6 text-white">
                      {{ feature.name }}
                      <span
                        class="inline-flex items-center px-3 py-0.5 rounded-full text-sm font-medium bg-secondary-light"
                        v-if="feature.soon">
                        Coming Soon
                      </span>
                    </p>
                  </dt>
                  <dd class="mt-1 ml-16 text-base text-gray-300">
                    {{ feature.description }}
                  </dd>
                </div>
              </div>
            </div>
          </div>
          <div class="px-4 py-12 bg-primary sm:px-6 lg:bg-none lg:px-0 lg:pl-8 lg:flex lg:items-center lg:justify-end">
            <div class="w-full lg:max-w-lg mx-auto space-y-8 lg:mx-0">
              <div>
                <span class="sr-only">Price</span>
                <p class="relative">
                  <span class="flex flex-col text-center" v-if="isDataFetched">
                    <span class="text-5xl font-bold text-white">from {{ formatCurrency(this.currency,
                      this.getFromPrice()) }}</span>
                    <span class="mt-2 text-base font-medium text-gray-100">per month</span>
                  </span>
                </p>
              </div>
              <ul class="grid gap-0.5 rounded sm:grid-cols-2">
                <li v-for="item in checklist" :key="item"
                  class="flex items-center px-4 py-4 space-x-3 text-base text-white bg-opacity-50">
                  <CheckIcon class="w-6 h-6 text-white" aria-hidden="true" />
                  <span>{{ item }}</span>
                </li>
              </ul>
              <a href="#plans"
                class="flex items-center justify-center w-full px-8 py-4 text-lg font-medium leading-6 bg-white border border-transparent rounded-md text-primary hover:text-primary-dark hover:bg-gray-200 md:px-10">
                Get Welcomer Pro
              </a>
              <a href="#custom-backgrounds"
                class="block text-base font-medium text-center text-white hover:text-gray-300">
                Get Custom Backgrounds Only
              </a>
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white" id="plans">
        <div class="hero-preview">
          <div class="sm:flex sm:flex-col sm:align-center">
            <div class="prose-lg text-center">
              <!-- <div class="mb-16 bg-secondary-dark rounded-md p-2 text-white font-semibold prose-sm">Save 25% off a monthly and annual membership this black friday!</div> -->
              <h1 class="font-black leading-8 tracking-tight text-gray-900">
                Choose the plan you want
              </h1>
              <p class="text-lg text-gray-500 section-subtitle max-w-prose mx-auto">
                Get started with Welcomer Pro without any recurring payments.
              </p>
            </div>
            <div class="grid grid-cols-1 lg:grid-cols-6 gap-6 mt-8">
              <div class="col-span-1 lg:col-span-5 w-full lg:w-fit relative bg-gray-100 rounded-lg p-0.5 flex flex-wrap self-center shadow-sm">
                <button type="button" @click="selectDuration(durationMonthly)" :class="[
                  durationSelected === durationMonthly
                    ? 'bg-white border-gray-300 text-gray-900 shadow-sm'
                    : 'border-transparent text-gray-700',
                  'relative border rounded-md py-2 w-full text-sm font-medium whitespace-nowrap focus:outline-none focus:ring-2 focus:ring-primary focus:z-10 lg:w-auto lg:px-8',
                ]">
                  Monthly
                  <span
                    v-if="isMonthlyRecurring"
                    class="inline-flex items-center ml-2 px-2.5 py-0.5 rounded-full text-xs font-medium bg-patreon text-white">
                    Recurring
                  </span>
                </button>
                <button type="button" @click="selectDuration(durationBiAnnually)" :class="[
                  'ml-0.5',
                  durationSelected === durationBiAnnually
                    ? 'bg-white border-gray-300 text-gray-900 shadow-sm'
                    : 'border-transparent text-gray-700',
                  'relative border rounded-md py-2 w-full text-sm font-medium whitespace-nowrap focus:outline-none focus:ring-2 focus:ring-primary focus:z-10 lg:w-auto lg:px-8',
                ]">
                  Biannual
                  <span
                    class="inline-flex items-center ml-2 px-2.5 py-0.5 rounded-full text-xs font-medium bg-primary text-white">
                    20% off
                  </span>
                </button>
                <button type="button" @click="selectDuration(durationAnnually)" :class="[
                  'ml-0.5',
                  durationSelected === durationAnnually
                    ? 'bg-white border-gray-300 text-gray-900 shadow-sm'
                    : 'border-transparent text-gray-700',
                  'relative border rounded-md py-2 w-full text-sm font-medium whitespace-nowrap focus:outline-none focus:ring-2 focus:ring-primary focus:z-10 lg:w-auto lg:px-8',
                ]">
                  Yearly
                  <span
                    class="inline-flex items-center ml-2 px-2.5 py-0.5 rounded-full text-xs font-medium bg-primary text-white">
                    20% off
                  </span>
                </button>
                <button type="button" @click="selectDuration(durationPatreon)" :class="[
                  'ml-0.5',
                  durationSelected === durationPatreon
                    ? 'bg-white border-gray-300 text-gray-900 shadow-sm'
                    : 'border-transparent text-gray-700',
                  'relative border rounded-md py-2 w-full text-sm font-medium whitespace-nowrap focus:outline-none focus:ring-2 focus:ring-patreon focus:z-10 lg:w-auto lg:px-8',
                ]">
                  Patreon
                  <span
                    class="inline-flex items-center ml-2 px-2.5 py-0.5 rounded-full text-xs font-medium bg-patreon text-white">
                    Recurring
                  </span>
                </button>
              </div>
              <Menu as="div" class="relative inline-block col-span-1 text-right align-middle">
                <MenuButton
                  class="inline-flex w-fit justify-right gap-x-1.5 rounded-md bg-white px-3 py-2 text-sm font-semibold text-gray-900 shadow-sm ring-1 ring-inset ring-gray-300 hover:bg-gray-50">
                  {{ getCurrencySymbol(currency) + ' – ' + currency }}
                  <ChevronDownIcon class="-mr-1 h-5 w-5 text-gray-400" aria-hidden="true" />
                </MenuButton>

                <transition enter-active-class="transition ease-out duration-100"
                  enter-from-class="transform opacity-0 scale-95" enter-to-class="transform opacity-100 scale-100"
                  leave-active-class="transition ease-in duration-75" leave-from-class="transform opacity-100 scale-100"
                  leave-to-class="transform opacity-0 scale-95">
                  <MenuItems
                    class="absolute right-0 z-10 mt-2 w-24 origin-top-right rounded-md bg-white shadow-lg ring-1 ring-black ring-opacity-5 focus:outline-none">
                    <div class="py-1">
                      <MenuItem v-for="currency in currencies" :key="currency" as="template" v-slot="{ active }">
                        <div @click="selectCurrency(currency)"
                          :class="[(active || this.currency === currency) ? 'bg-gray-100 text-gray-900' : 'text-gray-700', 'block px-4 py-2 text-sm cursor-pointer']">
                          {{ getCurrencySymbol(currency) + ' – ' + currency }}
                        </div>
                      </MenuItem>
                    </div>
                  </MenuItems>
                </transition>
              </Menu>
            </div>

            <div class="grid grid-cols-1 lg:grid-cols-2 gap-4 mt-8">
              <div class="border-gray-300 border p-6 lg:p-12 rounded-lg shadow-sm bg-white text-gray-900 h-fit">
                <h3 class="text-2xl font-bold sm:text-3xl">
                  Welcomer Basic
                </h3>
                <p class="mt-4 text-sm leading-6 text-gray-600">Includes all the essentials for your discord server.</p>
                <p class="mt-6 flex items-baseline gap-x-1">
                  <span class="text-xl font-bold tracking-tight text-gray-900">Free</span>
                </p>
                
                <router-link :to="{ name: 'invite' }"><button type="button" class="border-gray-300 hover:bg-gray-300 text-gray-900 border flex items-center justify-center px-5 py-3 mt-8 text-base font-medium rounded-md cursor-pointer w-full">Invite Welcomer</button></router-link>
              </div>
              <div class="-order-1">
                <div class="border-primary bg-primary text-white border p-6 lg:p-12 rounded-lg shadow-sm h-fit">
                  <h3 class="text-2xl font-bold sm:text-3xl">
                    Welcomer Pro
                  </h3>
                  <p class="mt-4 text-sm leading-6">Unlock more Welcomer features. Aimed at emerging or
                    well-established communities.</p>
                  <p class="mt-4 flex items-baseline gap-x-1" v-if="isDataFetched">
                    <span class="text-xl font-bold tracking-tight">{{ formatCurrency(this.currency,
                      (this.getSKU(this.getRelativeSKU())?.costs[this.currency] /
                        this.getSKU(this.getRelativeSKU())?.month_count)) }}</span>
                    <span class="text-sm font-semibold leading-6">{{
                      this.getSKU(this.getRelativeSKU())?.month_count > 1 ? '/ month*' : '/ month' }}</span>
                  </p>

                  <p v-if="(this.durationSelected == durationMonthly && isMonthlyRecurring) || this.durationSelected == durationPatreon" class="text-sm font-medium leading-6"> 7 days free </p>
                  <p v-if="this.getSKU(this.getRelativeSKU())?.month_count > 1" class="text-sm font-medium leading-6">Billed as {{ formatCurrency(this.currency,
                      this.getSKU(this.getRelativeSKU())?.costs[this.currency]) }}
                  </p>


                  <button type="button" @click.prevent="selectSKU(this.getRelativeSKU())"
                    :class="['bg-white hover:bg-gray-200', 'flex items-center justify-center px-5 py-3 mt-8 text-base font-medium text-primary border border-transparent rounded-md cursor-pointer w-full']">
                    <loading-icon class="mr-3" v-if="isCreatePaymentInProgress" />{{ durationSelected == durationPatreon ? 'Become a Patron' : 'Get Started' }}</button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="bg-white" id="custom-backgrounds">
        <div class="hero-preview">
          <div class="prose-lg text-center">
            <h1 class="font-black leading-8 tracking-tight text-gray-900">
              Just want backgrounds?
            </h1>
          </div>
          <div class="pb-16 mt-8 bg-white sm:mt-12 sm:pb-20 lg:pb-28">
            <div class="relative">
              <div class="mx-auto border border-gray-300 rounded-lg shadow-sm lg:flex">
                <div class="flex-1 px-6 py-8 my-auto bg-white lg:p-12">
                  <h2 class="text-2xl font-bold text-gray-900 sm:text-3xl">
                    Custom Welcomer Backgrounds
                  </h2>
                  <p class="mt-6 text-base leading-7 text-gray-600">
                    Get unlimited custom Welcomer backgrounds on any
                    server you choose, no need for monthly commitments,
                    this plan lasts forever.
                  </p>
                </div>
                <div
                  class="px-6 py-8 text-center shadow-sm bg-secondary lg:flex-shrink-0 lg:flex lg:flex-col lg:justify-center lg:p-12">
                  <p class="text-lg font-medium leading-6 text-gray-100">
                    Pay once, own it forever
                  </p>
                  <div class="flex items-center justify-center mt-4 text-5xl font-bold text-white">
                    <span v-if="this.getSKU(skuCustomBackgrounds)">
                      {{ formatCurrency(this.currency, this.getSKU(skuCustomBackgrounds)?.costs[this.currency]) }}
                    </span>
                  </div>
                  <button type="button" @click.prevent="selectSKU(skuCustomBackgrounds)"
                    class="flex items-center justify-center px-5 py-3 mt-8 text-base font-medium text-white border border-transparent rounded-md cursor-pointer bg-secondary-light hover:bg-secondary-dark w-full">
                    <loading-icon class="mr-3" v-if="isCreatePaymentInProgress" />
                    Get Started
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>

      <div class="bg-primary" id="faqs">
        <div class="hero-features">
          <div class="lg:grid lg:grid-cols-3 lg:gap-8">
            <div>
              <h1 class="text-3xl font-bold text-white">
                Frequently asked questions
              </h1>
              <p class="mt-4 text-lg text-white">
                Can't find what you are looking for? Reach out to us on our
                <a class="text-white hover:text-gray-300 underline" href="/support">support server</a>.
              </p>
            </div>
            <div class="mt-12 lg:mt-0 lg:col-span-2">
              <dl class="space-y-10 faq-container">
                <Disclosure as="div" v-for="faq in faqs" :key="faq.question" v-slot="{ open }">
                  <dt class="text-lg">
                    <DisclosureButton class="flex items-start justify-between w-full text-left">
                      <span :class="[open ? 'font-bold' : '', 'text-white']">
                        {{ faq.question }}
                      </span>
                      <span class="flex items-center h-6 ml-6">
                        <ChevronDownIcon :class="[
                          open ? '-rotate-180' : 'rotate-0',
                          'h-6 w-6 transform',
                        ]" aria-hidden="true" />
                      </span>
                    </DisclosureButton>
                  </dt>
                  <DisclosurePanel as="dd" class="pr-12 mt-2">
                    <span class="text-base text-gray-100" v-html="marked(faq.answer, true)"></span>
                  </DisclosurePanel>
                </Disclosure>
              </dl>
            </div>
          </div>
        </div>
      </div>
    </main>

    <Toast />

    <Footer />
  </div>
</template>

<style>
.faq-container a {
  text-decoration: underline;
}

.faq-container code {
    background: rgba(0, 0, 0, .2);
    padding: 2px;
}
</style>

<script>
import { ref } from "vue";

import Header from "@/components/Header.vue";
import Footer from "@/components/Footer.vue";
import Toast from "@/components/dashboard/Toast.vue";
import LoadingIcon from "@/components/LoadingIcon.vue";

import { toHTML } from "@/components/discord-markdown";

import billingAPI from "@/api/billing";

import { getErrorToast } from "@/utilities";

import {
  Disclosure,
  DisclosureButton,
  DisclosurePanel,
  RadioGroup,
  RadioGroupDescription,
  RadioGroupLabel,
  RadioGroupOption,
  Menu,
  MenuButton,
  MenuItem,
  MenuItems,
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
    soon: false,
    description:
      "Sometimes you don't want to give users a role immediately. Use timeroles to give them roles automatically when the time comes, it could be 10 minutes or in a year.",
  },
  {
    name: "Dedicated Pro Bot",
    icon: "plug-circle-bolt",
    description: "Use Welcomer with its own pro-only bot account.",
  },
  {
    name: "Custom Bot",
    icon: "plug-circle-plus",
    soon: true,
    description:
      "Run Welcomer with it's own unique account with a customizable username and avatar, all with the same reliability and uptime."
  },
];

const checklist = [
  "Custom Backgrounds",
  "Dedicated Pro Bot",
  "Flexible plans",
  "No Recurring Payments",
];

const faqs = [
  {
    question: "I have donated, now what?",
    answer: "When you have donated through PayPal and Discord, you should immediately receive your memberships. You can see these when doing `/membership list`, and will also autocomplete when doing `/membership add` on a server. Currently Patreon pledges will require a support ticket on our [support server](/support), however you will be able to soon link your Patreon to your Discord account on the Welcomer website. Currently managing memberships is only done through the membership commands, but memberships within the website will be coming soon.",
  },
  {
    question: "I have donated through Patreon but I have not received my membership.",
    answer: "Currently we cannot automatically link Patreon pledges to Discord accounts. Please join our [support server](/support) and open a ticket with your Patreon email and Discord ID, and we will manually add the membership to your account. Automatic linking will be coming soon.",
  },
  {
    question: "How can I automatically pay monthly for my membership with PayPal?",
    answer: "Currently we do not support recurring payments through PayPal, but this is planned. You can currently buy a month, 6 months or a year. If you would like to pay monthly, you can [pledge via our Patreon](/premium).",
  },
  {
    question: "How long do I keep custom backgrounds for?",
    answer: "Custom background memberships will last forever. There are a one-time payment, just make sure you do not remove your membership.",
  },
];

const durationPatreon = 0;
const durationMonthly = 1;
const durationBiAnnually = 2;
const durationAnnually = 3;

const skuCustomBackgrounds = "WEL/CBG";
const skuWelcomerPro = "WEL/1P1";
const skuWelcomerProBiAnnual = "WEL/1P6";
const skuWelcomerProAnnual = "WEL/1P12";

export default {
  components: {
    CheckIcon,
    ChevronDownIcon,
    Disclosure,
    DisclosureButton,
    DisclosurePanel,
    LoadingIcon,
    Footer,
    Header,
    Menu,
    MenuButton,
    MenuItem,
    MenuItems,
    RadioGroup,
    RadioGroupDescription,
    RadioGroupLabel,
    RadioGroupOption,
    Toast,
  },
  setup() {
    let durationSelected = ref(durationMonthly);
    let skus = ref([]);
    let currency = ref("USD");
    let currencies = ref([]);

    let isDataFetched = ref(false);
    let isDataError = ref(false);
    let isCreatePaymentInProgress = ref(false);
    let isMonthlyRecurring = ref(false);

    return {
      features,
      checklist,
      faqs,

      currencies,
      currency,
      durationSelected,
      skus,

      isDataFetched,
      isDataError,

      isCreatePaymentInProgress,
      isMonthlyRecurring,

      durationMonthly,
      durationBiAnnually,
      durationAnnually,
      durationPatreon,
      skuCustomBackgrounds,
      skuWelcomerPro,
      skuWelcomerProBiAnnual,
      skuWelcomerProAnnual,
    };
  },

  mounted() {
    var url = new URL(document.location);
    if (url.hash == "#success") {
      this.$store.dispatch("createToast", {
        title: "Your payment was successful, thank you for supporting us! 🎉 Check out the membership commands on your server to manage your memberships",
        icon: "heart",
        class: "text-white bg-primary",
        expiration: 60000,
    });
    }

    this.fetchSKUs();
  },

  methods: {
    marked(input, embed) {
      if (input) {
        return toHTML(input, {
          embed: embed,
          discordCallback: {
            user: function (user) {
              return `@${user.id}`;
            },
            channel: function (channel) {
              return `#${channel.id}`;
            },
            role: function (role) {
              return `@${role.id}`;
            },
            everyone: function () {
              return `@everyone`;
            },
            here: function () {
              return `@here`;
            },
          },
          cssModuleNames: {
            "d-emoji": "emoji",
          },
        });
      }
      return "";
    },
    getLocale() {
      return (navigator.languages && navigator.languages.length) ? navigator.languages[0] : navigator.language;
    },
    formatCurrency(currency, value) {
      return new Intl.NumberFormat(this.getLocale(), {
        style: "currency",
        currency: currency,
      }).format(value);
    },
    getCurrencySymbol(currency) {
      return new Intl.NumberFormat(this.getLocale(), {
        style: 'currency',
        currency: currency
      }).formatToParts().filter((i) => i.type == 'currency')[0].value;
    },
    getFromPrice() {
      let minimumPrice = Number.MAX_SAFE_INTEGER;
      this.skus.forEach(sku => {
        minimumPrice = Math.min(minimumPrice, sku.costs[this.currency]);
      });

      return minimumPrice == Number.MAX_SAFE_INTEGER ? 0 : minimumPrice;
    },
    getRelativeSKU() {
      if (this.durationSelected == durationMonthly) {
        return this.skuWelcomerPro;
      } else if (this.durationSelected == durationBiAnnually) {
        return this.skuWelcomerProBiAnnual;
      } else if (this.durationSelected == durationAnnually) {
        return this.skuWelcomerProAnnual;
      } else if (this.durationSelected == durationPatreon) {
        return this.skuWelcomerPro;
      }
    },
    getSKU(skuName) {
      return this.skus.find((sku) => sku.id === skuName);
    },
    fetchSKUs() {
      this.isDataFetched = false;
      this.isDataError = false;

      billingAPI.getSKUs(
        (response) => {
          this.isDataFetched = true;
          this.skus = response.skus;
          this.currency = response.default_currency;
          this.currencies = response.available_currencies;

          this.skus.forEach((sku) => {
            if (sku.id === this.skuWelcomerPro) {
              this.isMonthlyRecurring = sku.is_recurring && sku.paypal_subscription_id !== "";
            }
          });
        },
        (error) => {
          this.$store.dispatch("createToast", getErrorToast(error));

          this.isDataFetched = true;
          this.isDataError = true;
        }
      )
    },
    selectCurrency(currency) {
      this.currency = currency;
    },
    selectDuration(duration) {
      this.durationSelected = duration;
    },
    selectSKU(skuName) {
      const sku = this.getSKU(skuName);

      if (this.durationSelected === durationPatreon) {
        return window.open(
          `https://www.patreon.com/join/Welcomer/checkout?rid=${sku.patreon_checkout_id}`,
          "_blank"
        );
      }

      this.isCreatePaymentInProgress = true;

      billingAPI.createPayment(sku.id, this.currency, (response) => {
        this.isCreatePaymentInProgress = false;
        window.location.href = response.url;
      }, (error) => {
        this.isCreatePaymentInProgress = false;
        this.$store.dispatch("createToast", getErrorToast(error));
      });
    }
  }
};
</script>
