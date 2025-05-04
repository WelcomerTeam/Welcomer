<template>
  <div class="relative min-h-screen">
    <Header />

    <main>
      <div id="overview" class="relative bg-secondary text-white">
        <div class="px-6 py-12 w-full max-w-7xl mx-auto">
          <h1 class="text-3xl font-bold text-left tracking-tight">
            Frequently Asked Questions
          </h1>
        </div>
      </div>

      <div class="pb-32">
        <div class="hero-preview">
          <div class="px-4 mx-auto max-w-7xl sm:px-6 space-y-8">
            <p class="mt-4 text-lg p-8 border border-gray-300 rounded-lg shadow-sm">
              Can't find what you are looking for? Reach out to us on our
              <a class="underline" href="/support">support server</a>.
            </p>

            <ul class="mb-8 gap-y-1">
              <li v-for="faq in faqs" :key="faq.question"><a :href="'#' + getAnchor(faq.question)" class="text-primary underline font-bold block">{{ faq.question }}</a></li>
            </ul>
            <div class="faq-container space-y-8">
              <div v-for="faq in faqs" :key="faq.question" :id="getAnchor(faq.question)">
                <h2 class="font-semibold leading-8 tracking-tight">
                  <a :href="'#' + getAnchor(faq.question)">{{ faq.question }}</a>
                </h2>
                <span class="mt-3 text-lg section-subtitle max-w-prose mx-auto" v-html="marked(faq.answer, true, true)"></span>
              </div>
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
import Header from "@/components/Header.vue";
import Footer from "@/components/Footer.vue";

import { marked } from "@/utilities";

const faqs = [
  {
    question: "How can I add Welcomer to my server?",
    answer: "You can invite Welcomer to your server [here](/invite).",
  },
  {
    question: "My server is not showing up on the dashboard.",
    answer: "The dashboard will show all the guilds that the current logged in uer is in, even if Welcomer is not currently in it. Please make sure you are logged in as the correct user, and try refreshing the guild list.",
  },

  {
    question: "I have donated, now what?",
    answer: "When you have donated through PayPal and Discord, you should immediately receive your memberships. You can see these when doing `/membership list`, and will also autocomplete when doing `/membership add` on a server.",
  },
  {
    question: "I have donated through Patreon but I have not received my membership.",
    answer: "If you have not received your membership, you have to make sure your Discord account is linked to your Patreon account. You can do this by going to the [Patreon settings](https://www.patreon.com/settings/apps) and linking your Discord account. You can also link your Patreon account in the **Memberships** section on the dashboard.",
  },
  {
    question: "How can I automatically pay monthly for my membership with PayPal?",
    answer: "We now support monthly subscription plans. You can view the plans on [our premium page](/premium). We currently support monthly subscriptions, Patreon and Discord subscriptions. 6 month and 12 month plans do not automatically renew.",
  },
  {
    question: "How long do I keep custom backgrounds for?",
    answer: "Custom background memberships will last forever. There are a one-time payment, just make sure you do not remove your membership.",
  },

  {
    question: "How can I include the name of the user who joined in the welcome message?",
    answer: "You can use `{{User.Name}}` to show the name that is displayed for users.",
  },
  {
    question: "How can I include the member count in the welcome message?",
    answer: "You can use `{{Guild.Members}}` which will show as `374`. Use `{{Ordinal(Guild.Members)}}` to show as `374th`. [See all the formatting tags here](/formatting).",
  },
  {
    question: "How can I test the welcome message?",
    answer: "You can test your welcome messages via `/welcomer test`. When creating messages through the dashboard, we try to make sure it will display exactly how it shows in Discord, but you can always test it to be sure.",
  },
  {
    question: "How can I upload a custom background?",
    answer: "You can upload a custom background by first making sure you have a **Welcomer Pro** or **Custom Backgrounds** membership added to your server. On the dashboard go to the Welcomer tab and under **Welcomer Image Background**, make sure you select the **Custom** tab. If you have an active membership, this should let you upload a custom background or select a previously uploaded one.",
  },

  {
    question: "Why are my roles or channels not showing up in dropdowns?",
    answer: "The dashboard will only show channels it can message in or roles that it can assign (if applicable). If you would like to let use assign a role to a user that is not showing in the dropdown, make sure the top role that Welcomer has is above the roles you would like to assign.",
  },
];

export default {
  components: {
    Header,
    Footer,
  },
  setup() {
    return {
      faqs,
    };
  },
  methods: {
    marked(text, embed, skipFormatting) {
        return marked(text, embed, skipFormatting);
    },

    getAnchor(title) {
      return title.toLowerCase().replace(/ /g, "-");
    }
  },
};
</script>
