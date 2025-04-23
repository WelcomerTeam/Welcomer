<template>
  <div v-for="popup in $store.getters.getPopups" v-bind:key="popup.id">
    <Popup :open="true" @close="closePopup(popup.id, popup.closeFunction)" :showCloseButton="popup.showCloseButton">
      <template v-slot:title>
        {{ popup.title }}
      </template>
      
      <p v-if="popup.description !== ''" v-html="marked(popup.description, true)"></p>
      
      <div class="flex flex-col-reverse justify-start gap-2 sm:flex-row-reverse mt-4">
        <button type="button" class="cta-button bg-primary hover:bg-primary-dark" @click="continuePopup(popup.id, popup.continueFunction)" v-if="!popup.hideContinueButton">
          {{ popup.continueLabel ? popup.continueLabel : 'Continue' }}
        </button>
        <button type="button" class="focus:ring-primary focus:border-primary focus:outline-none border border-transparent font-semibold inline-flex items-center justify-center px-4 py-2 rounded-md shadow-sm text-sm text-black hover:bg-gray-200 dark:text-gray-50 dark:hover:bg-secondary-dark" @click="closePopup(popup.id, popup.closeFunction)" v-if="!popup.hideCancelButton">
          {{ popup.closeLabel ? popup.closeLabel : 'Cancel' }}
        </button>
      </div>
    </Popup>
  </div>
</template>

<script>
import Popup from "@/components/Popup.vue";

import { marked } from "@/utilities";

export default {
  components: {
    Popup,
  },
  methods: {
    marked(text, embed) {
      return marked(text, embed);
    },
    
    closePopup(popupID, popupCloseFunction) {
      this.$store.dispatch("removePopup", popupID);
      if (popupCloseFunction) {
        popupCloseFunction();
      }
    },
    
    continuePopup(popupID, popupContinueFunction) {
      this.$store.dispatch("removePopup", popupID);
      if (popupContinueFunction) {
        popupContinueFunction();
      }
    },
  }
}
</script>