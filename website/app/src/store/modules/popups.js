// initial state
const state = () => ({
    popups: [],
  });
  
  // getters
  const getters = {
    getPopups: (state) => {
      return state.popups;
    },
    getPopups: (state) => {
      return state.popups;
    },
  };
  
  // actions
  const actions = {
    createPopup({ commit }, popup) {
      popup.id = new Date().getTime();
      commit("addPopup", popup);
    },
    removePopup({ commit }, popupID) {
      commit("removePopup", popupID);
    },
  };
  
  // mutations
  const mutations = {
    addPopup(state, popupObject) {
      state.popups.push(popupObject);
    },
    removePopup(state, popupID) {
      state.popups = state.popups.filter((popup) => popup.id !== popupID);
    },
  };
  
  export default {
    namespaced: false,
    state,
    getters,
    actions,
    mutations,
  };
  