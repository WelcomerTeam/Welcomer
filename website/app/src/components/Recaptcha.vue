<template>
  <div></div>
</template>

<script>
export default {
  props: {
    action: {
      type: String,
      default: 'submit',
      required: true
    },
  },
  data() {
    return {
      sitekey: '6Le9GXcpAAAAAEjWrdGSRTq-thN0X8-yNiz4dy71',
    };
  },

  mounted() {
    const elem = document.querySelector(`script[data-identifier="recaptcha-script"]`);

    if (!elem) {
      const script = document.createElement('script');
      script.src = `https://www.google.com/recaptcha/api.js?render=${this.sitekey}`;
      script.async = true;
      script.defer = true;
      script.setAttribute('data-identifier', 'recaptcha-script');
      document.head.append(script);
    }
  },

  beforeUnmount() {
    var elem = document.querySelector(`script[data-identifier="recaptcha-script"]`);
    if (elem) {
      elem.remove();
    }

    elem = document.querySelector(`.grecaptcha-badge`);
    if (elem) {
      elem.parentElement.remove();
    }

    elem = document.querySelector(`meta[http-equiv="origin-trial"]`);
    if (elem) {
      elem.remove();
    }
  },

  methods: {
    execute() {
      var action = this.$props.action;
      var sitekey = this.sitekey;
      var emit = this.$emit;

      grecaptcha.ready(function () {
        grecaptcha.execute(sitekey, { action: action }).then((token) => {
          emit('verify', token);
        });
      })
    },
    render() {
      this.execute();
    },
  },
};
</script>