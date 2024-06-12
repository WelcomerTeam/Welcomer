<template></template>

<script>
export default {
  props: {
    unsavedChanges: {
      type: Boolean,
      required: true,
    }
  },

  beforeRouteLeave() {
    return !this.confirmStayInDirtyForm();
  },

  beforeDestroy() {
    window.removeEventListener("beforeunload", this.beforeWindowUnload);
    window.removeEventListener("mouseup", this.mouseUpHandler);
  },

  mounted() {
    window.addEventListener("beforeunload", this.beforeWindowUnload);
    window.addEventListener("mouseup", this.mouseUpHandler);
  },

  methods: {
    onValueUpdate() {
      this.$props.unsavedChanges = true;
    },

    confirmStayInDirtyForm() {
      return this.$props.unsavedChanges && !this.confirmLeave();
    },

    confirmLeave() {
      return window.confirm(
        "You have unsaved changes! Are you sure you want to leave?"
      );
    },

    beforeWindowUnload(e) {
      if (this.confirmStayInDirtyForm()) {
        e.preventDefault();
        e.returnValue = "";
      }
    },
  },
};
</script>

<style></style>
