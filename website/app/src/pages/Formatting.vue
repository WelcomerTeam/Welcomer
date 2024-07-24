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
                      class="bg-gray-50 py-2 pl-4 pr-3 text-left text-sm font-semibold sm:pl-3">{{
                        formattingTag.name }}</th>
                  </tr>
                  <tr v-for="(value, id) in formattingTag.values" :key="value.name"
                    :class="[id === 0 ? 'border-gray-300' : 'border-gray-200', 'border-t']">
                    <td class="py-4 pl-4 pr-3 text-sm font-medium sm:pl-3"><code>{{ value.name }}</code>
                    </td>
                    <td class="px-3 py-4 text-sm" v-html="marked(value.description, true)"></td>
                    <td class="px-3 py-4 text-sm break-all" v-html="marked(value.example, true)"></td>
                  </tr>
                </template>
              </tbody>
            </table>
          </div>
        </div>
      </div>

      <div>
        <div class="bg-donate">
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

import { toHTML } from "@/components/discord-markdown";

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
  },
  setup() {
    return {
      mustacheTags: "{{Mustache.Tags}}",
      formattingTags,
      textExamples,
    };
  },
  methods: {
    marked(input, embed) {
      if (input) {
        return toHTML(input, {
          embed: embed,
          discordCallback: {
            user: function (user) {
              if (user.id == 143090142360371200) {
                return `@ImRock`;
              }

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
  }
};
</script>
