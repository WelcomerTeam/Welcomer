import dashboardAPI from "@/api/dashboard";

// initial state
const state = () => ({
  selectedGuild: null,
  guild: null,
  guildHasWelcomer: false,

  guildHasWelcomerPro: false, // TODO
  guildHasCustomBackgrounds: false, // TODO

  isLoadingGuild: false,

  guildChannels: [],
  guildChannelsPacked: [],
  guildRoles: [],
  guildEmojis: [],

  isLoadingGuildMembers: false, // TODO
  guildMemberResults: {}, // TODO
});

// getters
const getters = {
  isLoadingGuild: (state) => {
    return state.isLoadingGuild;
  },

  isLoadingGuildMembers: (state) => {
    return state.isLoadingGuildMembers;
  },

  guildHasWelcomerPro: (state) => {
    return state.guildHasWelcomerPro;
  },

  guildHasCustomBackgrounds: (state) => {
    return state.guildHasCustomBackgrounds;
  },

  guildHasWelcomer: (state) => {
    return state.guildHasWelcomer;
  },

  getCurrentSelectedGuild: (state) => {
    return state.guild;
  },

  getSelectedGuildID: (state) => {
    return state.selectedGuild;
  },

  getGuildChannels: (state) => {
    return state.guildChannels;
  },

  getPackedGuildChannels: (state) => {
    return state.guildChannelsPacked;
  },

  getGuildRoles: (state) => {
    return state.guildRoles;
  },

  getAssignableGuildRoles: (state) => {
    var roles = state.guildRoles.filter((role) => role.id !== state.guild.id);
    roles.sort((a, b) => a.position - b.position);
    return roles;

    // return state.guildRoles.filter((role) => role.is_assignable && role.id !== state.guild.id);
  },

  getGuildEmojis: (state) => {
    return state.guildEmojis;
  },

  getGuildMemberResults: (state) => {
    return state.guildMemberResults;
  },

  getGuildChannelById: (state) => (channelID) => {
    return state.guildChannels.find((channel) => channel.id == channelID);
  },

  getGuildRoleById: (state) => (roleID) => {
    return state.guildRoles.find((role) => role.id == roleID);
  },

  getGuildEmojiById: (state) => (emojiID) => {
    return state.guildEmojis.find((emoji) => emoji.id == emojiID);
  },

  getGuildMemberById: (state) => (guildMemberID) => {
    return state.guildMemberResults[guildMemberID];
  },
};

// actions
const actions = {
  fillGuild({ dispatch, state }) {
    if (state.selectedGuild != state.guild?.id && state.selectedGuild) {
      dispatch("fetchGuild");
    }
  },

  fetchGuild({ commit, state }) {
    if (state.selectedGuild == undefined) {
      commit("setGuild", { guild: null, hasWelcomer: false });
    } else {
      commit("loadingGuild");
      dashboardAPI.getGuild(
        state.selectedGuild,
        ({ guild, hasWelcomer }) => {
          commit("setGuild", { guild, hasWelcomer });
        },
        () => commit("setGuild", { guild: null, hasWelcomer: false })
      );
    }
  },

  fetchGuildMembersByQuery({ commit, state }, query) {
    commit("loadingGuildMembers");
    dashboardAPI.fetchGuildMembers(
      query,
      state.selectedGuild,
      (guildMembers) => {
        commit("setGuildMemberResults", guildMembers);
      },
      () => {
        commit("setGuildMemberResults", []);
      }
    );
  },
};

// mutations
const mutations = {
  setGuild(state, { guild, hasWelcomer }) {
    state.guild = guild?.guild;
    state.guildHasWelcomerPro = guild?.has_welcomer_pro;
    state.guildHasCustomBackgrounds = guild?.has_custom_backgrounds;
    state.guildChannels = guild?.guild?.channels || [];
    state.guildChannelsPacked = packGuildChannels(state.guildChannels);
    state.guildRoles = guild?.guild?.roles || [];

    state.guildRoles.sort((a, b) => b.position - a.position);

    state.guildEmojis = guild?.guild?.emojis || [];
    state.isLoadingGuild = false;
    state.guildHasWelcomer = hasWelcomer;
  },

  setSelectedGuild(state, guildID) {
    state.selectedGuild = guildID;
    this.dispatch("fillGuild");
  },

  setGuildMemberResults(state, guildMembers) {
    guildMembers.forEach((member) => {
      state.guildMemberResults[member.id] = member;
    });
    state.isLoadingGuildMembers = false;
  },

  loadingGuildMembers(state) {
    state.isLoadingGuildMembers = true;
  },

  loadingGuild(state) {
    state.isLoadingGuild = true;

    state.guild = {};
    state.guild.id = state.selectedGuild;
  },
};

function packGuildChannels(channels) {
  let categories = {};
  let emptyCategory = { channels: [] };

  channels
    .filter((channel) => channel.type === 4)
    .map((channel) => {
      channel = JSON.parse(JSON.stringify(channel));
      channel.channels = [];
      categories[channel.id] = channel;
    });

  channels
    .filter((channel) => channel.type !== 4)
    .map((channel) => {
      if (categories[channel.parent_id]) {
        categories[channel.parent_id].channels.push(channel);
      } else {
        emptyCategory.channels.push(channel);
      }
    });

  return [emptyCategory, ...Object.values(categories)];
}

export default {
  namespaced: false,
  state,
  getters,
  actions,
  mutations,
};
