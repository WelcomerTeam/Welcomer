import { doLogin, doRequest, getRequest } from "./routes";

export default {
  getSKUs(callback, errorCallback) {
    getRequest(
      "/api/billing/skus",
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

  createPayment(sku, currency, guild_id, callback, errorCallback) {
    doRequest(
      "POST",
      "/api/billing/payments",
      { sku: sku, currency: currency, guild_id: guild_id },
      null,
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
                callback({ url: res.data.url });
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
    )
  },
}