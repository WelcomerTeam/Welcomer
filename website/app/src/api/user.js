import { doLogin, getRequest } from "./routes";

export default {
  getUser(callback, errorCallback) {
    getRequest(
      "/api/users/@me",
      (response) => {
        if (response.status === 401) {
          callback({ user: null });
        } else {
          response
            .json()
            .then((res) => {
              if (res.ok) {
                callback({ user: res.data });
              } else {
                errorCallback(res.error);
              }
            })
            .catch((error) => {
              errorCallback(error);
            });
        }
      },
      (error) => {
        errorCallback(error);
      }
    );
  },

  getGuilds(doRefresh, callback, errorCallback) {
    function cmp(a, b) {
      if (a > b) return +1;
      if (a < b) return -1;
      return 0;
    }

    getRequest(
      "/api/users/guilds" + (doRefresh ? "?refresh=1" : ""),
      (response) => {
        if (response.status === 401) {
          doLogin();
        } else {
          response
            .json()
            .then((res) => {
              if (res.ok) {
                let sortedGuilds = res.data
                  .filter((a) => a.has_elevation)
                  .sort((a, b) => {
                    return cmp(a.is_owner, b.is_owner) || cmp(a.has_welcomer, b.has_welcomer) || a.name.localeCompare(b.name)
                  });
                sortedGuilds.reverse();
                callback({ guilds: sortedGuilds });
              } else {
                errorCallback(res.error);
              }
            })
            .catch((error) => {
              errorCallback(error);
            });
        }
      },
      (error) => {
        errorCallback(error);
      }
    );
  },
};
