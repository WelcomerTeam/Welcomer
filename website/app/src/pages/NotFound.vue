<template>
  <div class="relative min-h-screen">
    <Header />

    <main>
      <div class="p-8 mx-12 my-12 text-center text-white rounded-md bg-secondary-dark">
        <img class="mx-auto" src="/assets/finding.png" width="64" />
        <p class="my-4 text-xl">
          I can't seem to find this page. Maybe you went to the wrong place?
        </p>
        <a class="underline cursor-pointer" @click.prevent="handleBack('/')">Go back</a>
      </div>
    </main>

    <div class="footer-anchor">
      <Footer />
    </div>
  </div>
</template>

<script>
import Header from "@/components/Header.vue";
import Footer from "@/components/Footer.vue";

export default {
  components: {
    Header,
    Footer,
  },

  data() {
    return {
      fromRoute: null,
    }
  },

  beforeRouteEnter(to, from, next) {
    next((vm) => {
      vm.fromRoute = from;
    });
  },

  methods: {
    handleBack(fallback) {
      if (!this.fromRoute?.name) {
        this.$router.push(fallback);
      } else {
        this.$router.back();
      }
    },
  },
};
</script>
