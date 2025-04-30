<template>
  <div class="relative min-h-screen">
    <Header />

    <main>
      <div id="overview" class="relative bg-secondary">
        <div class="px-6 py-12 bg-secondary w-full max-w-7xl mx-auto">
          <h1 class="text-3xl font-bold text-left text-white tracking-tight">
            Text Formatting
          </h1>
        </div>
      </div>

      <div class="bg-white text-neutral-900">
        <div class="hero-preview">
          <div class="px-4 mx-auto max-w-7xl sm:px-6">
            <div class="sm:flex sm:flex-col sm:align-center">
              <div class="prose-lg text-center">
                <span class="mt-3 text-lg section-subtitle max-w-prose mx-auto">
                  Welcomer now uses <code>{{ mustacheTags }}</code> for formatting variables in your welcome, leaver and
                  borderwall messages. This allows you to customise your messages with ease.
                </span>
              </div>
            </div>

            <table class="mt-8 w-full">
              <thead>
                <tr>
                  <th scope="col" class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold sm:pl-3">Name</th>
                  <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold">Description</th>
                  <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold">Example</th>
                </tr>
              </thead>
              <tbody>
                <template v-for="formattingTag in formattingTags" :key="formattingTag.name">
                  <tr class="border-t border-gray-200">
                    <th colspan="5" scope="colgroup"
                      class="bg-gray-50 py-2 pl-4 pr-3 text-left text-sm font-semibold sm:pl-3">{{ formattingTag.name }}
                    </th>
                  </tr>
                  <tr v-for="(value, id) in formattingTag.values" :key="value.name"
                    :class="[id === 0 ? 'border-gray-300' : 'border-gray-200', 'border-t']">
                    <td class="py-4 pl-4 pr-3 text-sm font-medium sm:pl-3">
                      <code class="cursor-copy group relative whitespace-nowrap" @click="copyTag(value)">
                      {{ value.name }}
                      <font-awesome-icon icon="fa-regular fa-copy" class="w-4 h-4 top-1 text-gray-400 absolute -left-6 group-hover:visible invisible" aria-hidden="true" />
                    </code>
                    </td>
                    <td class="px-3 py-4 text-sm" v-html="marked(value.description, true, true)"></td>
                    <td class="px-3 py-4 text-sm break-all" v-html="marked(value.example, true)"></td>
                  </tr>
                </template>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <div>
        <div class="bg-primary">
          <div class="hero-features">
            <p class="text-3xl font-bold text-left text-white tracking-tight">
              Examples
            </p>

            <table class="mt-8 w-full">
              <thead>
                <tr>
                  <th scope="col" class="py-3.5 pl-4 pr-3 text-left text-sm font-semibold sm:pl-3">Example
                  </th>
                  <th scope="col" class="px-3 py-3.5 text-left text-sm font-semibold">Result</th>
                </tr>
              </thead>
              <tbody>
                <tr v-for="(value, id) in textExamples" :key="value.name"
                  :class="[id === 0 ? 'border-gray-300' : 'border-gray-200', 'border-t']">
                  <td class="py-4 pl-4 pr-3 text-sm font-medium sm:pl-3"><kbd>{{ value.example }}</kbd>
                  </td>
                  <td class="px-3 py-4 text-sm" v-html="marked(value.result, true)"></td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
      </div>
    </main>

    <Toast />

    <Footer />
  </div>
</template>

<style lang="scss">
code {
  @apply bg-secondary-dark text-white px-2 py-1 rounded-md;
}
</style>

<script>
import Header from "@/components/Header.vue";
import Footer from "@/components/Footer.vue";
import Toast from "@/components/dashboard/Toast.vue";

import { marked } from "@/utilities";

const formattingTags = [
  {
    name: "User",
    values: [
      { name: "{{User.ID}}", description: "The user's id", example: "143090142360371200" },
      { name: "{{User.Name}}", description: "The user's global name or username with discriminator", example: "ImRock" },
      { name: "{{User.Username}}", description: "The user's username", example: "imrock" },
      { name: "{{User.Discriminator}}", description: "The user's discriminator", example: "0" },
      { name: "{{User.GlobalName}}", description: "The user's global name", example: "ImRock" },
      { name: "{{User.Mention}}", description: "Mentions the user", example: "<@143090142360371200>" },
      { name: "{{User.CreatedAt}}", description: "The user's creation date as relative time", example: "`8 years ago`" },
      { name: "{{User.JoinedAt}}", description: "The user's join date as relative time", example: "`7 years ago`" },
      { name: "{{User.Avatar}}", description: "The user's avatar as a URL", example: "https://cdn.discordapp.com/avatars/143090142360371200/a73420b217a77a77b17fb42fa7ecfbcc.png" },
      { name: "{{User.Bot}}", description: "Boolean to indicate the user is a bot", example: "false" },
      { name: "{{User.Pending}}", description: "Boolean to indicate the user is pending membership screening", example: "false" },
    ]
  },
  {
    name: "Guild",
    values: [
      { name: "{{Guild.ID}}", description: "The guild's id", example: "341685098468343822" },
      { name: "{{Guild.Name}}", description: "The guild's name", example: "Welcomer Support Guild" },
      { name: "{{Guild.Icon}}", description: "The guild's icon as a URL", example: "https://cdn.discordapp.com/icons/341685098468343822/09cfc7fe72945a7c04ec6d3ddd01767c.png" },
      { name: "{{Guild.Splash}}", description: "The guild's splash image as a URL", example: "" },
      { name: "{{Guild.Members}}", description: "The guild's member count", example: "7600" },
      { name: "{{Guild.Banner}}", description: "The guild's banner image as a URL", example: "" },
    ]
  },
  {
    name: "Invite",
    values: [
      { name: "{{Invite.Code}}", description: "The code of the invite", example: "UyUVCEcBU9" },
      { name: "{{Invite.Uses}}", description: "The number of times the invite has been used", example: "2724" },
      { name: "{{Invite.Inviter}}", description: "The user who created the invite", example: "Welcomer#5491" },
      { name: "{{Invite.ChannelID}}", description: "The ID of the channel the invite is for", example: "1234567890" },
      { name: "{{Invite.CreatedAt}}", description: "The creation date of the invite", example: "`3 months ago`" },
      { name: "{{Invite.ExpiresAt}}", description: "The expiration date of the invite", example: "`in 5 days`" },
      { name: "{{Invite.MaxAge}}", description: "The maximum age of the invite in seconds", example: "86400" },
      { name: "{{Invite.MaxUses}}", description: "The maximum number of times the invite can be used", example: "10000" },
      { name: "{{Invite.Temporary}}", description: "Boolean to indicate if the invite is temporary", example: "false" }
    ]
  },
  {
    name: "Functions",
    values: [
      { name: "{{Ordinal(int)}}", description: "Returns the ordinal (st, nd, rd, th) for an integer passed in. You can do `{{Ordinal(Guild.Members)}}` to display the member count.", example: "7600th" },
    ]
  }
]

const textExamples = [
  { example: "Welcome {{User.Mention}} to **{{Guild.Name}}**! You are the {{Ordinal(Guild.Members)}} member!", result: "Welcome <@143090142360371200> to **Welcomer Support Guild**! You are the 7600th member!" },
  { example: "{{#User.Bot}}This message shows if the user is a bot.{{/User.Bot}}{{^User.Bot}}This message shows if the user is not a bot.{{/User.Bot}}", result: "This message shows if the user is not a bot." },
]

export default {
  components: {
    Header,
    Footer,
    Toast,
  },
  setup() {
    return {
      mustacheTags: "{{Mustache.Tags}}",
      formattingTags,
      textExamples,
    };
  },
  methods: {
    marked(text, embed, skipFormatting) {
      return marked(text, embed, skipFormatting);
    },

    copyTag(formattingTag) {
      navigator.clipboard.writeText(formattingTag.name);

      this.$store.dispatch("createToast", {
        title: "Copied to clipboard",
        icon: "info",
        class: "text-blue-500 bg-blue-100",
      });
    }
  }
};
</script>
