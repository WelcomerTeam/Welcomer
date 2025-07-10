const dummyMembers = [
  {
    id: 330416853971107840,
    name: "Welcomer",
    display_name: "Welcomer#5491",
    discriminator: "5491",
    avatar: "5f65708fd35ee3844a463d5bf9fe7828",
    bot: true,
  },
  {
    id: 689205223306297416,
    name: "Lukef",
    display_name: "Lukef#3842",
    discriminator: "3842",
    avatar: "a_bc7930302db3239b4b31c430793eaeb4",
    bot: false,
  },
  {
    id: 209692588792479744,
    name: "Lowy",
    display_name: "Lowy#2001",
    discriminator: "2001",
    avatar: "a_2f775ae4d0449b35b18ca4374d1244dd",
    bot: false,
  },
  {
    id: 774938059342479372,
    name: "Senpai Legend",
    display_name: "Senpai Legend#0001",
    discriminator: "0001",
    avatar: "a_a39932eac94ae45b5868f41d9670e081",
    bot: false,
  },
  {
    id: 157589756442574848,
    name: "Bergin",
    display_name: "Bergin#2077",
    discriminator: "2077",
    avatar: "e4037b1f022697bfcd29ad2b3b9c4988",
    bot: false,
  },
  {
    id: 660139603960922115,
    name: "ReafilL",
    display_name: "ReafilL#8684",
    discriminator: "8684",
    avatar: "3e508d5972a4a02f78e072dd4fb6a418",
    bot: false,
  },
  {
    id: 749469480865759283,
    name: "anne_",
    display_name: "anne_#7432",
    discriminator: "7432",
    avatar: "67dcf6d87e02b7db74aa1dd19d22a099",
    bot: false,
  },
  {
    id: 855599077542723604,
    name: "hiraginoyuki",
    display_name: "hiraginoyuki#1284",
    discriminator: "1284",
    avatar: "5bc8db40cb86ed904693dbce219870c3",
    bot: false,
  },
  {
    id: 852719537849630792,
    name: "Reo",
    display_name: "Reo#6699",
    discriminator: "6699",
    avatar: "c3258b8cf95c4c38fed3bb1a1bcc4e66",
    bot: false,
  },
  {
    id: 253960412117205002,
    name: "AmberChief23",
    display_name: "AmberChief23#2513",
    discriminator: "2513",
    avatar: "0cc5d84580414f0684fd2223657103a8",
    bot: false,
  },
];

import Fuse from "fuse.js";

const dummyMemberSearch = new Fuse(dummyMembers, {
  keys: ["name", "id", "display_name"],
  threshold: 0.2,
  useExtendedSearch: true,
});

import { doLogin, doRequest, getRequest } from "./routes";

export default {
  getGuild(guildID, callback, errorCallback) {
    getRequest(
      "/api/guild/" + guildID,
      (response) => {
        if (response.status === 401) {
          doLogin();
        } else if (response.status == 403) {
          callback({ guild: null, hasWelcomer: false });
        } else {
          response
            .json()
            .then((res) => {
              if (res.ok) {
                callback({ guild: res.data, hasWelcomer: true });
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

  getStatus(callback, errorCallback) {
    getRequest(
      "/api/status",
      (response) => {
        if (response.status === 401) {
          doLogin();
        } else {
          response
            .json()
            .then((res) => {
              if (res.ok) {
                callback(res.data);
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

  fetchGuildMembers(query, guildID, callback, errorCallback) {
    if (query == "") {
      callback(this.dummyMembers);
    } else {
      callback(
        dummyMemberSearch.search(query).map((v) => {
          return v.item;
        })
      );
    }
  },

  getConfig(endpoint, callback, errorCallback) {
    return this.doAPICall("GET", endpoint, undefined, undefined, callback, errorCallback);
  },

  doPost(endpoint, data, files, callback, errorCallback) {
    return this.doAPICall("POST", endpoint, data, files, callback, errorCallback);
  },

  doAPICall(method, endpoint, data, files, callback, errorCallback) {
    return doRequest(
      method,
      endpoint,
      data,
      files,
      (response) => {
        if (response.status === 401) {
          doLogin();
        } else if (response.status == 403) {
          callback({ config: null });
        } else {
          response
            .json()
            .then((res) => {
              if (res.ok) {
                callback({ config: res.data });
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

  getBorderwall(borderwallId, callback, errorCallback) {
    getRequest(
      "/api/borderwall/" + borderwallId,
      (response) => {
        if (response.status === 401) {
          doLogin();
        } else {
          response
            .json()
            .then((res) => {
              if (res.ok) {
                callback({ code: res.code, data: res.data });
              } else {
                errorCallback({ code: res.code, error: res.error });
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

  submitBorderwall(borderwallId, response, callback, errorCallback) {
    doRequest(
      "POST",
      "/api/borderwall/" + borderwallId,
      response,
      null,
      (response) => {
        if (response.status === 401) {
          doLogin();
        } else {
          response
            .json()
            .then((res) => {
              if (res.ok) {
                callback(res.data);
              } else {
                console.debug("Ok is false", res);
                errorCallback(res.error);
              }
            })
            .catch((error) => {
              console.debug("Caught error", error);
              errorCallback(error);
            });
        }
      },
      (error) => {
        console.debug("doRequest error", error);
        errorCallback(error);
      }
    );
  },
};
