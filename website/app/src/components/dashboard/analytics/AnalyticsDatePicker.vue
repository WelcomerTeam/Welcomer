<template>
    <Popover as="div" v-slot="{ open }" class="relative">
        <div class="border-gray-300 dark:border-secondary-light">
                <PopoverButton :class="[
                    'bg-white dark:bg-secondary',
                    'relative w-full cursor-default rounded-md border border-gray-300 py-2 pl-3 pr-10 text-left text-sm shadow-sm focus:border-primary focus:outline-none focus:ring-1 focus:ring-primary dark:border-secondary-light',
                ]">
                    <span>{{ formattedRange }}</span>
                    <span class="absolute inset-y-0 right-0 flex items-center pr-2 pointer-events-none">
                        <ChevronDownIcon :class="[
                            open ? 'transform rotate-180' : '',
                            'h-5 w-5 text-gray-400 transition-all duration-100',
                        ]" aria-hidden="true" />
                    </span>
                </PopoverButton>

                <transition
                    :show="open"
                    leave-active-class="transition duration-100 ease-in"
                    leave-from-class="opacity-100"
                    leave-to-class="opacity-0"
                >
                    <PopoverPanel class="absolute right-0 z-10 mt-1 w-[min(100vw-2rem,38rem)] rounded-md border border-gray-300 bg-white p-4 text-sm shadow-lg dark:border-secondary-light dark:bg-secondary">
                        <div class="grid gap-5 sm:grid-cols-[minmax(0,1fr)_10rem]">
                            <section aria-label="Choose a custom date range">
                                <div class="mb-3 flex items-center justify-between">
                                    <button
                                        type="button"
                                        class="flex h-8 w-8 items-center justify-center rounded-md text-gray-500 hover:bg-gray-100 hover:text-gray-900 focus:outline-none focus:ring-2 focus:ring-primary dark:hover:bg-secondary-light dark:hover:text-white"
                                        aria-label="Previous month"
                                        @click="changeMonth(-1)"
                                    >
                                        <ChevronLeftIcon class="h-5 w-5" aria-hidden="true" />
                                    </button>
                                    <span class="font-medium text-gray-900 dark:text-white">{{ monthLabel }}</span>
                                    <button
                                        type="button"
                                        class="flex h-8 w-8 items-center justify-center rounded-md text-gray-500 hover:bg-gray-100 hover:text-gray-900 focus:outline-none focus:ring-2 focus:ring-primary dark:hover:bg-secondary-light dark:hover:text-white"
                                        aria-label="Next month"
                                        @click="changeMonth(1)"
                                    >
                                        <ChevronRightIcon class="h-5 w-5" aria-hidden="true" />
                                    </button>
                                </div>

                                <div class="grid grid-cols-7 text-center text-xs font-medium text-gray-500 dark:text-gray-400">
                                    <span v-for="weekday in weekdays" :key="weekday" class="py-1">{{ weekday }}</span>
                                </div>
                                <div class="grid grid-cols-7">
                                    <button
                                        v-for="day in calendarDays"
                                        :key="day.value"
                                        type="button"
                                        :disabled="isDateDisabled(day.value)"
                                        :aria-label="formatLongDate(day.value)"
                                        :aria-pressed="isSelected(day.value)"
                                        :class="[
                                            'relative flex h-9 items-center justify-center text-sm focus:z-10 focus:outline-none focus:ring-2 focus:ring-primary',
                                            day.isCurrentMonth ? 'text-gray-900 dark:text-white' : 'text-gray-400 dark:text-gray-500',
                                            isDateDisabled(day.value) ? 'cursor-not-allowed opacity-40' : '',
                                            isRangeMiddle(day.value) ? 'bg-primary/15 dark:bg-primary/25' : '',
                                            isRangeEdge(day.value) ? 'bg-primary font-semibold text-white' : (isDateDisabled(day.value) ? '' : 'hover:bg-gray-100 dark:hover:bg-secondary-light'),
                                            isRangeStart(day.value) ? 'rounded-l-md' : '',
                                            isRangeEnd(day.value) ? 'rounded-r-md' : '',
                                            isSingleDayRange(day.value) ? 'rounded-md' : '',
                                        ]"
                                        @click="selectDate(day.value)"
                                    >
                                        {{ day.day }}
                                    </button>
                                </div>
                                <p class="mt-3 text-xs text-gray-500 dark:text-gray-400">
                                    {{ awaitingRangeEnd ? "Choose an end date" : "Choose a start date, then an end date" }}
                                </p>
                            </section>

                            <section class="border-t border-gray-200 pt-4 dark:border-secondary-light sm:border-l sm:border-t-0 sm:pl-5 sm:pt-0" aria-label="Date range presets">
                                <p class="mb-2 text-xs font-medium uppercase text-gray-500 dark:text-gray-400">Presets</p>
                                <div class="space-y-1">
                                    <button
                                        v-for="option in presetOptions"
                                        :key="option.value"
                                        type="button"
                                        :class="[
                                            'w-full rounded-md px-3 py-2 text-left text-sm transition-colors focus:outline-none focus:ring-2 focus:ring-primary',
                                            selectedPreset === option.value
                                                ? 'bg-primary font-medium text-white'
                                                : 'text-gray-700 hover:bg-gray-100 dark:text-gray-200 dark:hover:bg-secondary-light',
                                        ]"
                                        @click="selectPreset(option.value)"
                                    >
                                        {{ option.label }}
                                    </button>
                                </div>
                            </section>
                        </div>
                    </PopoverPanel>
                </transition>
        </div>
    </Popover>
</template>

<script>
import { Popover, PopoverButton, PopoverPanel } from "@headlessui/vue";
import { ChevronDownIcon, ChevronLeftIcon, ChevronRightIcon } from "@heroicons/vue/solid";

export default {
    components: {
        ChevronDownIcon,
        ChevronLeftIcon,
        ChevronRightIcon,
        Popover,
        PopoverButton,
        PopoverPanel,
    },
    props: {
        modelValue: {
            type: [String, Object],
            required: true,
        },
        defaultPreset: {
            type: String,
            default: "last_7_days",
        },
        minDate: {
            type: String,
            default: null,
        },
        maxDate: {
            type: String,
            default: null,
        },
    },
    emits: ["update:modelValue"],
    data() {
        const today = new Date();

        return {
            selectedPreset: this.defaultPreset || "last_7_days",
            customStart: "",
            customEnd: "",
            awaitingRangeEnd: false,
            visibleMonth: new Date(today.getFullYear(), today.getMonth(), 1),
        };
    },
    computed: {
        presetOptions() {
            return [
                { value: "last_7_days", label: "Last 7 days" },
                { value: "last_30_days", label: "Last 30 days" },
                { value: "this_week", label: "This week" },
                { value: "last_week", label: "Last week" },
                { value: "this_month", label: "This month" },
                { value: "last_month", label: "Last month" },
                { value: "custom", label: "Custom range" },
            ];
        },
        weekdays() {
            return ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
        },
        monthLabel() {
            return new Intl.DateTimeFormat(undefined, {
                month: "long",
                year: "numeric",
            }).format(this.visibleMonth);
        },
        calendarDays() {
            const firstDay = new Date(this.visibleMonth.getFullYear(), this.visibleMonth.getMonth(), 1);
            const start = this.addDays(firstDay, -((firstDay.getDay() + 6) % 7));

            return Array.from({ length: 42 }, (_, index) => {
                const date = this.addDays(start, index);

                return {
                    day: date.getDate(),
                    value: this.formatDate(date),
                    isCurrentMonth: date.getMonth() === this.visibleMonth.getMonth(),
                };
            });
        },
        formattedRange() {
            const range = this.normalizeModelValue({
                preset: this.selectedPreset,
                startDate: this.customStart,
                endDate: this.customEnd,
            });
            
            return `${this.formatLongDate(range.startDate)} — ${this.formatLongDate(range.endDate)}`;
        },
    },
    watch: {
        modelValue: {
            immediate: true,
            handler(value) {
                const range = this.normalizeModelValue(value);
                this.selectedPreset = range.preset;
                this.customStart = range.startDate;
                this.customEnd = range.endDate;
                this.awaitingRangeEnd = false;

                const endDate = this.toDate(range.endDate);
                if (endDate) {
                    this.visibleMonth = new Date(endDate.getFullYear(), endDate.getMonth(), 1);
                }
            },
        },
    },
    methods: {
        formatDate(date) {
            const year = date.getFullYear();
            const month = String(date.getMonth() + 1).padStart(2, "0");
            const day = String(date.getDate()).padStart(2, "0");
            
            return `${year}-${month}-${day}`;
        },
        formatLongDate(value) {
            if (!value) {
                return "Select a range";
            }
            
            const date = new Date(`${value}T00:00:00`);
            return new Intl.DateTimeFormat(undefined, {
                year: "numeric",
                month: "short",
                day: "numeric",
            }).format(date);
        },
        toDate(value) {
            if (!value) {
                return null;
            }
            
            if (value instanceof Date) {
                return new Date(value.getFullYear(), value.getMonth(), value.getDate());
            }
            
            const date = new Date(`${value}T00:00:00`);
            return Number.isNaN(date.getTime())
            ? null
            : new Date(date.getFullYear(), date.getMonth(), date.getDate());
        },
        addDays(date, days) {
            const next = new Date(date);
            next.setDate(next.getDate() + days);
            return next;
        },
        startOfWeek(date) {
            const day = date.getDay();
            const diff = (day + 6) % 7; // Monday as start of week
            return this.addDays(date, -diff);
        },
        endOfWeek(date) {
            return this.addDays(this.startOfWeek(date), 6);
        },
        startOfMonth(date) {
            return new Date(date.getFullYear(), date.getMonth(), 1);
        },
        endOfMonth(date) {
            return new Date(date.getFullYear(), date.getMonth() + 1, 0);
        },
        changeMonth(offset) {
            this.visibleMonth = new Date(this.visibleMonth.getFullYear(), this.visibleMonth.getMonth() + offset, 1);
        },
        buildPresetRange(preset) {
            const today = new Date();
            const currentDay = new Date(today.getFullYear(), today.getMonth(), today.getDate());
            
            switch (preset) {
                case "last_7_days":
                return {
                    preset,
                    startDate: this.formatDate(this.addDays(currentDay, -6)),
                    endDate: this.formatDate(currentDay),
                };
                case "last_30_days":
                return {
                    preset,
                    startDate: this.formatDate(this.addDays(currentDay, -29)),
                    endDate: this.formatDate(currentDay),
                };
                case "this_week":
                return {
                    preset,
                    startDate: this.formatDate(this.startOfWeek(currentDay)),
                    endDate: this.formatDate(currentDay),
                };
                case "last_week": {
                    const start = this.addDays(this.startOfWeek(currentDay), -7);
                    return {
                        preset,
                        startDate: this.formatDate(start),
                        endDate: this.formatDate(this.addDays(start, 6)),
                    };
                }
                case "this_month":
                return {
                    preset,
                    startDate: this.formatDate(this.startOfMonth(currentDay)),
                    endDate: this.formatDate(currentDay),
                };
                case "last_month": {
                    const lastMonthEnd = new Date(currentDay.getFullYear(), currentDay.getMonth(), 0);
                    return {
                        preset,
                        startDate: this.formatDate(this.startOfMonth(lastMonthEnd)),
                        endDate: this.formatDate(this.endOfMonth(lastMonthEnd)),
                    };
                }
                default:
                return {
                    preset: "custom",
                    startDate: this.formatDate(this.addDays(currentDay, -6)),
                    endDate: this.formatDate(currentDay),
                };
            }
        },
        rangesEqual(a, b) {
            return a.startDate === b.startDate && a.endDate === b.endDate;
        },
        identifyPreset(range) {
            const presets = this.presetOptions
                .filter((option) => option.value !== "custom")
                .map((option) => option.value);

            for (const preset of presets) {
                const presetRange = this.buildPresetRange(preset);
                if (this.rangesEqual(presetRange, range)) {
                    return preset;
                }
            }
            
            return "custom";
        },
        normalizeModelValue(value) {
            if (typeof value === "string") {
                const presetMatch = this.presetOptions.some((option) => option.value === value);
                if (presetMatch && value !== "custom") {
                    return this.buildPresetRange(value);
                }
                
                try {
                    const parsed = JSON.parse(value);
                    if (parsed && parsed.startDate && parsed.endDate) {
                        const range = {
                            preset: parsed.preset || "custom",
                            startDate: parsed.startDate,
                            endDate: parsed.endDate,
                        };
                        return {
                            ...range,
                            preset: this.identifyPreset(range),
                        };
                    }
                } catch {
                    // fall through
                }
            }
            
            if (value && typeof value === "object") {
                const startDate = value.startDate || value.start;
                const endDate = value.endDate || value.end;
                
                if (startDate && endDate) {
                    const range = {
                        preset: value.preset || "custom",
                        startDate,
                        endDate,
                    };
                    
                    return {
                        ...range,
                        preset: this.identifyPreset(range),
                    };
                }
            }
            
            return this.buildPresetRange("last_7_days");
        },
        emitRange(range) {
            this.$emit("update:modelValue", range);
        },
        selectPreset(preset) {
            if (preset === "custom") {
                this.selectedPreset = "custom";
                this.awaitingRangeEnd = false;
                return;
            }

            const range = this.buildPresetRange(preset);
            this.selectedPreset = preset;
            this.customStart = range.startDate;
            this.customEnd = range.endDate;
            this.awaitingRangeEnd = false;
            this.emitRange(range);
        },
        isDateDisabled(value) {
            const date = new Date(`${value}T00:00:00`);
            return Boolean((this.minDate && date < this.minDate) || (this.maxDate && date > this.maxDate));
        },
        selectDate(value) {
            if (this.isDateDisabled(value)) {
                return;
            }

            if (!this.awaitingRangeEnd) {
                this.selectedPreset = "custom";
                this.customStart = value;
                this.customEnd = "";
                this.awaitingRangeEnd = true;
                return;
            }

            const [startDate, endDate] = value < this.customStart
                ? [value, this.customStart]
                : [this.customStart, value];

            this.customStart = startDate;
            this.customEnd = endDate;
            this.awaitingRangeEnd = false;
            this.emitRange({ preset: "custom", startDate, endDate });
        },
        isRangeStart(value) {
            return Boolean(this.customStart && value === this.customStart);
        },
        isRangeEnd(value) {
            return Boolean(this.customEnd && value === this.customEnd);
        },
        isSingleDayRange(value) {
            return this.isRangeStart(value) && this.isRangeEnd(value);
        },
        isRangeEdge(value) {
            return this.isRangeStart(value) || this.isRangeEnd(value);
        },
        isRangeMiddle(value) {
            return Boolean(this.customStart && this.customEnd && value > this.customStart && value < this.customEnd);
        },
        isSelected(value) {
            return this.isRangeEdge(value) || this.isRangeMiddle(value);
        },
    },
};
</script>
