<template>
  <div class="w-full">
    <div class="flex items-center gap-2 bg-gray-100 px-3 py-2 dark:bg-secondary-dark">
      <svg class="h-4 w-4 text-gray-500 dark:text-gray-300" fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2" viewBox="0 0 24 24" aria-hidden="true">
        <circle cx="11" cy="11" r="7"></circle>
        <line x1="21" y1="21" x2="16.65" y2="16.65"></line>
      </svg>
      <input v-model="searchTerm" type="text" :placeholder="searchPlaceholder" class="w-full bg-transparent text-sm outline-none placeholder-gray-400 dark:placeholder-gray-500 border-transparent" />
    </div>

    <div v-if="loading" class="py-6 text-center text-sm text-gray-500 dark:text-gray-300">
      Loading emoji library...
    </div>

    <div v-else class="p-2 h-auto max-h-96 space-y-4 overflow-y-auto pr-1">
      <div v-if="!displayedGroups.length" class="rounded-md border border-dashed border-gray-300 px-3 py-4 text-center text-sm text-gray-500 dark:border-secondary-light dark:text-gray-300">
        No emojis found.
      </div>

      <div v-for="group in displayedGroups" :key="group.key" class="space-y-2">
        <div class="flex items-center justify-between px-1 text-[11px] font-semibold uppercase tracking-wide text-gray-500 dark:text-gray-400">{{ group.label }}</div>
        <div class="flex flex-wrap gap-2 justify-center" role="list">
          <button v-for="emoji in group.items" :key="emoji.key" type="button" @click="selectEmoji(emoji)" class="flex h-10 w-10 items-center justify-center rounded-md hover:bg-gray-100 dark:hover:bg-secondary-light">
            <img :src="emoji.url" :alt="emoji.name" :title="emoji.name" loading="lazy" class="max-h-7 w-7" @error="handleImageError" />
            <span class="sr-only">{{ emoji.name }}</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { computed, ref } from "vue";

import Fuse from "fuse.js";
import emojiDataByGroup from "unicode-emoji-json/data-by-group.json";

export default {
	name: "DiscordEmojiPicker",

	props: {
		modelValue: {
			type: String,
			default: "",
		},
		customEmojis: {
			type: Array,
			default: () => [],
		},
		customGroupLabel: {
			type: String,
			default: "Custom",
		},
		twemojiCdn: {
			type: String,
			default: "https://twemoji.maxcdn.com/v/latest/svg",
		},
		searchPlaceholder: {
			type: String,
			default: "Search emojis...",
		},
	},

	emits: ["update:modelValue", "select"],

	setup(props, { emit }) {
		const searchTerm = ref("");
		const loading = ref(true);
		const allEmojis = ref([]);
		const fuse = ref(null);
		const groupOrder = Object.keys(emojiDataByGroup);

		const toTwemojiCode = (emojiChar) => Array.from(emojiChar)
		.map((part) => part.codePointAt(0))
		.map((code) => code.toString(16))
		.join("-");

		const baseEmojis = [];
		groupOrder.forEach((groupName) => {
			const emojis = emojiDataByGroup[groupName] || [];

			emojis.forEach((emoji, index) => {
				const code = toTwemojiCode(emoji.emoji);

				baseEmojis.push({
					key: `${groupName}-${index}-${code}`,
					name: emoji.name,
					emoji: emoji.emoji,
					group: emoji.group || groupName,
					subgroup: emoji.subgroup,
					url: `${props.twemojiCdn}/${code}.svg`,
					searchText: `${emoji.name} ${emoji.group || groupName} ${emoji.subgroup || ""}`.toLowerCase(),
				});
			});
		});

		allEmojis.value = baseEmojis;
		fuse.value = new Fuse(baseEmojis, {
			keys: ["name", "group", "subgroup"],
			threshold: 0.32,
		});

		loading.value = false;

		let isReducedMotion = false;
		if (window.matchMedia) {
			const mediaQuery = window.matchMedia("(prefers-reduced-motion: reduce)");
			isReducedMotion = mediaQuery.matches;
		}

		const customEmojiItems = computed(() =>
			props.customEmojis
				.filter((emoji) => emoji && emoji.id)
				.map((emoji, index) => {
					return {
						key: `custom-${emoji.id}-${index}`,
						name: emoji.name,
						id: emoji.id,
						url: `https://cdn.discordapp.com/emojis/${emoji.id}.${emoji.animated && !isReducedMotion ? "gif" : "webp"}?size=64`,
						isCustom: true,
						searchText: emoji.name.toLowerCase(),
					};
				})
		);

		const displayedGroups = computed(() => {
			if (loading.value) {
				return [];
			}

			const term = searchTerm.value.trim().toLowerCase();
			const standardPool = term && fuse.value ? fuse.value.search(term).map((result) => result.item) : allEmojis.value;

			const grouped = new Map();
			standardPool.forEach((emoji) => {
				const bucket = grouped.get(emoji.group) || [];
				bucket.push(emoji);
				grouped.set(emoji.group, bucket);
			});

			const standardGroups = groupOrder
				.map((groupName) => {
					const items = grouped.get(groupName) || [];
					return items.length
						? {
							key: `group-${groupName}`,
							label: groupName,
							items,
						}
						: null;
				})
				.filter(Boolean);

			const customMatches = customEmojiItems.value.filter((emoji) => !term || emoji.searchText.includes(term));

			const ordered = [];
			if (customMatches.length) {
				ordered.push({
					key: "custom",
					label: props.customGroupLabel,
					items: customMatches,
				});
			}

			ordered.push(...standardGroups);
			return ordered;
		});

		const selectEmoji = (emoji) => {
			const payload = emoji.isCustom
				? {
					type: "custom",
					name: emoji.name,
					value: emoji.id,
					url: emoji.url,
				}
				: {
					type: "unicode",
					name: emoji.name,
					value: emoji.emoji,
					url: emoji.url,
				};

			emit("update:modelValue", payload.value);
			emit("select", payload);
		};

		const handleImageError = (event) => {
			// Hide the button if the emoji image fails to load
			const button = event.target.closest('button');
			if (button) {
				button.style.display = 'none';
			}
		};

		return {
			searchTerm,
			loading,
			displayedGroups,
			selectEmoji,
			handleImageError,
		};
	},
};
</script>
