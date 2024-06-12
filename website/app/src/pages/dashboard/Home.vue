<template>
  <div>
    <div class="relative">
      <img v-if="$store.getters.getCurrentSelectedGuild?.banner" :src="`https://cdn.discordapp.com/banners/${$store.getters.getCurrentSelectedGuild?.id
        }/${$store.getters.getCurrentSelectedGuild?.banner}.${$store.getters.getCurrentSelectedGuild?.banner.startsWith('a_')
          ? 'gif'
          : 'png'
        }?size=1024`" class="w-full aspect-video object-cover max-h-64" />
      <div v-else :style="getBackgroundGradient(
        rgbIntToRGB(
          $store.getters.getCurrentSelectedGuild?.embedColour
            ? $store.getters.getCurrentSelectedGuild?.embedColour
            : getDefaultGuildColour(
              $store.getters.getCurrentSelectedGuild?.id
            )
        )
      )
        " class="w-full aspect-video object-cover max-h-64" />
    </div>
    <div class="dashboard-container">
      <div class="pb-14">
        <img v-if="$store.getters.getCurrentSelectedGuild"
          class="w-32 h-32 rounded-full translate -translate-y-24 border-8 border-white bg-white dark:border-secondary dark:bg-secondary absolute"
          :src="$store.getters.getCurrentSelectedGuild?.icon !== ''
              ? `https://cdn.discordapp.com/icons/${$store.getters.getCurrentSelectedGuild?.id}/${$store.getters.getCurrentSelectedGuild?.icon}.webp?size=128`
              : '/assets/discordServer.svg'
            " alt="Guild icon" />
      </div>
      <div class="dashboard-title-container">
        <div class="dashboard-title align-middle">
          <span v-if="$store.getters.guildHasWelcomerPro"
            class="mr-2 inline-flex items-center p-2 rounded-md text-xs font-medium bg-primary-light text-white">
            <font-awesome-icon icon="heart" />
          </span>
          <span v-else-if="$store.getters.guildHasCustomBackgrounds"
            class="mr-2 inline-flex items-center p-2 rounded-md text-xs font-medium bg-gray-500 text-white">
            <font-awesome-icon icon="heart" />
          </span>
          {{ $store.getters.getCurrentSelectedGuild?.name }}
        </div>
      </div>
      <div class="dashboard-content">
        <div class="grid grid-cols-1 gap-5 mt-2 sm:grid-cols-2 lg:grid-cols-3">
          <!-- Card -->
          <Card name="Server Members" icon="user-group" :amount="$store.getters.getCurrentSelectedGuild?.member_count" />
          <Card name="Text Channels" icon="user-group"
            :amount="filterTextChannels($store.getters.getGuildChannels).length" />
          <Card name="Voice Channels" icon="user-group" :amount="filterVoiceChannels($store.getters.getGuildChannels).length
            " />
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import Card from "@/components/dashboard/Card.vue";

export default {
  components: { Card },
  methods: {
    filterTextChannels(channels) {
      return channels.filter((channel) => channel.type === 0);
    },

    filterVoiceChannels(channels) {
      return channels.filter((channel) => channel.type === 2);
    },

    rgbIntToRGB(rgbInt, defaultValue) {
      return (
        "#" +
        (rgbInt == undefined ? defaultValue : rgbInt)
          .toString(16)
          .slice(-6)
          .padStart(6, "0")
      );
    },

    getBackgroundGradient(rgbValue) {
      return `background-color: ${rgbValue}; background-image: url("data:image/svg+xml,%3Csvg width='84' height='48' viewBox='0 0 84 48' xmlns='http://www.w3.org/2000/svg'%3E%3Cpath d='M0 0h12v6H0V0zm28 8h12v6H28V8zm14-8h12v6H42V0zm14 0h12v6H56V0zm0 8h12v6H56V8zM42 8h12v6H42V8zm0 16h12v6H42v-6zm14-8h12v6H56v-6zm14 0h12v6H70v-6zm0-16h12v6H70V0zM28 32h12v6H28v-6zM14 16h12v6H14v-6zM0 24h12v6H0v-6zm0 8h12v6H0v-6zm14 0h12v6H14v-6zm14 8h12v6H28v-6zm-14 0h12v6H14v-6zm28 0h12v6H42v-6zm14-8h12v6H56v-6zm0-8h12v6H56v-6zm14 8h12v6H70v-6zm0 8h12v6H70v-6zM14 24h12v6H14v-6zm14-8h12v6H28v-6zM14 8h12v6H14V8zM0 8h12v6H0V8z' fill='${rgbValue.replace(
        "#",
        "%23"
      )}' filter='brightness(0.5)' fill-opacity='0.4' fill-rule='evenodd'/%3E%3C/svg%3E");`;
    },

    getDefaultGuildColour(guildID) {
      var guildColours = [
        "#2F80ED",
        "#72DACE",
        "#FBC01B",
        "#202225",
        "#292B2F",
        "#2F3237",
        "#36393F",
        "#40444A",
      ];

      return guildColours[BigInt(guildID) % BigInt(guildColours.length)];
    },
  },
};
</script>
