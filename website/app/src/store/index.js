import { createStore, createLogger } from "vuex";

import dashboard from "./modules/dashboard";
import user from "./modules/user";
import toast from "./modules/toast";
import popups from "./modules/popups";

const debug = process.env.NODE_ENV !== "production";

export default createStore({
  modules: {
    dashboard,
    user,
    toast,
    popups,
  },
  strict: debug,
  plugins: debug ? [createLogger()] : [],
});
