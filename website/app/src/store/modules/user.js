import userAPI from "@/api/user";

// initial state
const state = () => ({
  isLoggedIn: false,
  user: null,
  guilds: [],

  isLoadingUser: false,
  isLoadingGuilds: false,
});

// getters
const getters = {
  isLoadingUser: (state) => {
    return state.isLoadingUser;
  },

  isLoadingGuilds: (state) => {
    return state.isLoadingGuilds;
  },

  isLoggedIn: (state) => {
    return state.isLoggedIn;
  },

  getCurrentUser: (state) => {
    return state.user;
  },

  getGuilds: (state) => {
    return state.guilds;
  },
};

// actions
const actions = {
  fetchCurrentUser({ commit }) {
    commit("loadingUser");
    userAPI.getUser(
      ({ user }) => {
        commit("setCurrentUser", user);
      },
      () => {
        commit("setCurrentUser", undefined);
      }
    );
  },

  refreshGuilds({ commit }) {
    commit("loadingGuilds");
    userAPI.getGuilds(
      true,
      ({ guilds }) => {
        commit("setGuilds", guilds);
      },
      () => {
        commit("setGuilds", undefined);
      }
    );
  },

  fetchGuilds({ commit }) {
    commit("loadingGuilds");
    userAPI.getGuilds(
      false,
      ({ guilds }) => {
        commit("setGuilds", guilds);
      },
      () => {
        commit("setGuilds", undefined);
      }
    );
  },
};

// mutations
const mutations = {
  setCurrentUser(state, userObject) {
    state.isLoggedIn = userObject != undefined;
    state.user = userObject;
    state.isLoadingUser = false;
  },

  setGuilds(state, guilds) {
    state.guilds = guilds;
    state.isLoadingGuilds = false;
  },

  loadingUser(state) {
    state.isLoadingUser = true;
  },

  loadingGuilds(state) {
    state.isLoadingGuilds = true;
  },
};

export default {
  namespaced: false,
  state,
  getters,
  actions,
  mutations,
};
