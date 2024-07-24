<template>
  <div class="relative overflow-hidden mx-auto carousel-container top-10" aria-hidden="true">
    <div class="carousel-overlay"></div>
    <div class="absolute overflow-hidden left-0 top-0 carousel">
      <img src="/assets/examples/Frame0.png" class="carousel-item" aria-hidden="true" alt="" style="--color: #0b5dea" />
      <img src="/assets/examples/Frame1.png" class="carousel-item" aria-hidden="true" alt="" style="--color: #ed4245" />
      <img src="/assets/examples/Frame2.png" class="carousel-item" aria-hidden="true" alt="" style="--color: #7ac4c9" />
      <img src="/assets/examples/Frame3.png" class="carousel-item" aria-hidden="true" alt="" style="--color: #171c39" />
      <!-- Set 2 -->
      <img src="/assets/examples/Frame0.png" class="carousel-item" aria-hidden="true" alt="" style="--color: #0b5dea" />
      <img src="/assets/examples/Frame1.png" class="carousel-item" aria-hidden="true" alt="" style="--color: #ed4245" />
      <img src="/assets/examples/Frame2.png" class="carousel-item" aria-hidden="true" alt="" style="--color: #7ac4c9" />
      <img src="/assets/examples/Frame3.png" class="carousel-item" aria-hidden="true" alt="" style="--color: #171c39" />
      <!-- Set 3 -->
      <img src="/assets/examples/Frame0.png" class="carousel-item" aria-hidden="true" alt="" style="--color: #0b5dea" />
      <img src="/assets/examples/Frame1.png" class="carousel-item" aria-hidden="true" alt="" style="--color: #ed4245" />
      <img src="/assets/examples/Frame2.png" class="carousel-item" aria-hidden="true" alt="" style="--color: #7ac4c9" />
      <img src="/assets/examples/Frame3.png" class="carousel-item" aria-hidden="true" alt="" style="--color: #171c39" />
    </div>
  </div>
</template>

<style lang="scss">
:root {
  --carousel-item-width: 528px;
  --carousel-item-count: 4;
  --carousel-set-count: 3;
  --carousel-items-visible: 6;
}

.carousel-container {
  width: calc(var(--carousel-item-width) * var(--carousel-items-visible));
  max-width: 100%;
  height: 162px;
}

.carousel-overlay {
  width: 100%;
  height: 100%;
  pointer-events: none;
  background: linear-gradient(90deg,
      var(--background-tertiary) 0%,
      transparent 20%,
      transparent 80%,
      var(--background-tertiary) 100%);
  mask: linear-gradient(90deg, black 0%, white 20%, white 80%, black 100%);
  position: absolute;
  z-index: 1;
}

.carousel:hover {
  animation-play-state: paused;
}

.carousel {
  animation: moveSlideshow 20s linear infinite;
  width: calc(var(--carousel-item-width) * var(--carousel-item-count) * var(--carousel-set-count));
  height: 162px;
}

.carousel-item {
  @apply shadow-lg border-solid border-8 rounded-md h-40 inline-flex mr-4;
  background-color: var(--color);
  border-color: var(--color);
  background-position: 0 0;
  background-size: cover;
  width: 512px;
  height: 162px;
}

@keyframes moveSlideshow {
  100% {
    transform: translateX(calc(0px - (var(--carousel-item-width) * var(--carousel-item-count))));
  }
}
</style>
