<template>
  <footer class="bg-secondary-dark">
    <div class="lg:px-6 max-w-7xl mx-auto px-4 py-8 sm:px-6">
      <nav class="-mx-5 -my-2 flex flex-wrap" aria-label="Footer">
        <span class="px-5 py-2 text-xl font-bold text-white"> Welcomer </span>
        <div class="flex">
          <div v-for="item in footer" :key="item.name" class="px-5 py-2">
            <router-link :to="item.href" class="text-white hover:text-gray-300">
              {{ item.name }}
            </router-link>
          </div>
        </div>
      </nav>
    </div>
  </footer>
</template>

<script>
const footer = [
  {
    name: "Terms",
    href: "/terms",
  },
  {
    name: "Privacy",
    href: "/privacy",
  },
  {
    name: "Status",
    href: "/status",
  },
];

export default {
  setup() {
    return {
      footer,
    };
  },
};
</script>
