<template>
  <div class="relative min-h-screen bg-secondary">
    <Header />
    <HoistHeading />
    <main>
      <div class="mx-auto my-4 lg:relative max-w-7xl">
        <div class="hero">
          <div class="px-6 my-auto lg:w-1/2 sm:px-8 xl:pr-16">
            <h1 class="hero-title">The go-to bot for your discord server</h1>
            <p class="hero-subtitle">
              Providing tools to let you manage your server. Join the more than
              <b>521,473</b> Discord guilds using Welcomer.
            </p>
            <div class="hero-buttons">
              <div class="hero-primary">
                <router-link :to="{ name: 'invite' }">
                  Invite Welcomer
                </router-link>
              </div>
              <div class="hero-secondary">
                <a href="#features">Features</a>
              </div>
            </div>
          </div>
        </div>
        <div class="hero-accompany">
          <div class="flex w-full h-full">
            <DiscordEmbed class="max-w-full m-auto" :isBot="true" :showTimestamp="true" :isDark="true" :embeds="[
              {
                description: `Hey **ImRock**, you are the 256th member on **Welcomer Support Guild**`,
                image: {
                  url: '/assets/welcomerImage.svg',
                },
              },
            ]" />
          </div>
        </div>
      </div>

      <div class="bg-white text-secondary">
        <div class="hero-preview">
          <div class="pb-12 prose-lg text-center">
            <h1 class="text-lg font-black leading-8 tracking-tight text-gray-900">
              Elevate Your User Engagement
            </h1>
            <span class="text-lg text-gray-700 max-w-3xl mx-auto section-subtitle">Explore the reasons why Welcomer is the go-to choice for empowering Discord guilds. Packed with numerous features and boasting a user-friendly dashboard, you have the flexibility to make changes whenever you desire.
            </span>
          </div>

          <div v-for="item in previews" :key="item.name" class="hero-preview-item">
            <div class="my-auto text-center sm:text-left">
              <div>
                <h2 class="mb-4 text-3xl font-black leading-tight text-gray-900">
                  {{ item.name }}
                </h2>
                <span class="text-gray-700">{{ item.description }}
                  <router-link :to="item.href" v-if="item.href"
                    class="text-primary hover:text-primary-dark font-semibold">
                    {{ item.label }}
                  </router-link>
                </span>
              </div>
            </div>
            <div class="my-auto">
              <img :src="item.src" class="px-4 mt-4 sm:mt-0" />
            </div>
          </div>
        </div>
      </div>

      <div class="bg-primary">
        <div class="hero-features">
          <div class="hero-features-item">
            <div v-for="item in features" :key="item.name"
              class="p-4 mx-0 my-3 text-center rounded-lg sm:text-left sm:mx-2 hover:bg-primary-light">
              <h2 class="text-xl font-medium text-white">{{ item.name }}</h2>
              <p>{{ item.description }}</p>
            </div>
          </div>
        </div>
      </div>

      <div class="bg-secondary">
        <div class="px-4 py-12 mx-auto text-center max-w-prose sm:px-6 md:py-16 lg:px-7 lg:py-20">
          <h2 class="text-2xl font-bold tracking-tight text-gray-900 sm:text-4xl">
            <span class="block text-white">Ready to get started?</span>
            <span class="block text-primary">Invite Welcomer today, it's free.</span>
          </h2>
          <p class="mt-5 text-lg leading-6 text-gray-400">
            Elevate Moderation, Enhance Engagement: Welcomer is your key to a thriving community. Start today!
          </p>
          <router-link :to="{ name: 'dashboard.guilds' }">
            <button type="button" class="w-32 mt-4 cta-button bg-primary hover:bg-primary-dark">
              Dashboard
            </button>
          </router-link>
        </div>
      </div>
    </main>

    <Footer />
  </div>
</template>

<script>
import Header from "@/components/Header.vue";
import Footer from "@/components/Footer.vue";
import DiscordEmbed from "@/components/DiscordEmbed.vue";
import HoistHeading from "@/components/hoist/HoistHeading.vue";

const previews = [
  {
    name: "Empower your Welcome Messages",
    description: "Tailor your welcome messages with Welcomer, offering personalized customization for text, direct messages, and images, ensuring a unique and engaging experience for every Discord user.",
    src: "/assets/placeholder.png",
  },
  {
    name: "Reinforce Your Server's Defenses with Borderwall",
    description:
      "With Borderwall enabled, ensure only authentic users can send messages by presenting a challenge upon joining, guaranteeing a secure and spam-resistant community.",
    src: "/assets/placeholder.png",
  },
  // {
  //   name: "Increase user engagement",
  //   description:
  //     "Reward users with XP for conversing on your server and let them compete against each other with a public leaderboard.",
  //   src: "/assets/placeholder.png",
  // },
  // {
  //   name: "Automate your server",
  //   description:
  //     "Automation allows you to create tasks that execute on events in a zero-code environment. Want to give roles to anybody who sends “pog”? You can do that without any knowledge of programming.",
  //   src: "/assets/placeholder.png",
  // },
];

const features = [
  // AutoReminders
  {
    name: "AutoRoles",
    description: "Automatically give users roles when they join your server.",
  },
  {
    name: "Borderwall",
    description: "Secure your server from automated accounts with a challenge when joining.",
  },
  {
    name: "FreeRoles",
    description: "Allow users to easily give and remove certain roles from themselves through a command.",
  },
  // InviteRoles
  {
    name: "Leaver",
    description: "Send a message when users leave your server, wishing them farewell.",
  },
  // LevelRoles
  // Lockdown
  // Polls
  // ReactionRoles
  {
    name: "Rules",
    description: "Provide a list of rules for users to see, and send them when a user joins your server.",
  },
  // Starboard
  // StickyRoles
  // Suggestions
  {
    name: "TempChannels",
    description: "Allow users to create temporary voice channels, freeing up space.",
  },
  {
    name: "TimeRoles",
    description: "Reward users for staying in your server for a period of time, with special roles.",
  },
  {
    name: "Welcomer",
    description: "Welcome new users to your servers to your server with fancy images and text or send them a direct message.",
  },
  // XP
];

export default {
  components: {
    Header,
    Footer,
    DiscordEmbed,
    HoistHeading,
  },
  setup() {
    return {
      previews,
      features,
    };
  },
};
</script>
