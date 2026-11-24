import { createLogger, createStore } from "vuex";

import dashboard from "./modules/dashboard";
import popups from "./modules/popups";
import toast from "./modules/toast";
import user from "./modules/user";

const debug = process.env.NODE_ENV !== "production";

export function createAppStore() {
  return createStore({
    modules: {
      dashboard,
      user,
      toast,
      popups,
    },
    strict: debug,
    plugins: debug ? [createLogger()] : [],
  });
}
