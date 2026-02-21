import { createStore, createLogger } from "vuex";

import dashboard from "./modules/dashboard";
import popups from "./modules/popups";
import toast from "./modules/toast";
import user from "./modules/user";

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
