<template>
    <div>
        <label v-if="$slots.default" class="block mb-1 text-neutral-500 text-xs font-medium">
            <slot></slot>
        </label>
        <input :value="modelValue"
            @blur="onblur($event.target.value)"
            @keydown.enter="onenter($event)"
            @keydown.up="onarrow($event)"
            @keydown.down="onarrow($event)"
            :placeholder="placeholder"
            class="border rounded w-full h-9 py-2 px-3 bg-secondary-dark border-secondary-light"/>
    </div>
</template>

<script>
export default {
    name: 'InputCalculator',
    props: {
        modelValue: {
            type: [String, Number],
            default: ''
        },
        type: {
            type: String,
            default: 'text'
        },
        placeholder: {
            type: String,
            default: ''
        },
        min: {
            type: Number,
            default: null
        },
        max: {
            type: Number,
            default: null
        },
        minPercentage: {
            type: Number,
            default: null
        },
        maxPercentage: {
            type: Number,
            default: null
        },
        step: {
            type: Number,
            default: 1
        },
        forceStringOutput: {
            type: Boolean,
            default: false
        }
    },

    methods: {
        cleanValue(value) {
            // if type is number remove non digits, operators, decimal point and basic operators
            if (this.type === 'number') {
                if (typeof value !== 'string') {
                    value = String(value);
                }
                value = value.replace(/[^0-9+\-*/%.() ]/g, '');
            }
            return value;
        },
        update(value) {
            value = this.cleanValue(value);
            const expr = String(value).trim();

            // allow only digits, whitespace, parentheses, decimal point and basic operators
            if (/^[0-9\s()+\-*/%.]+$/.test(expr) && /[+\-*/%]/.test(expr)) {
                try {
                    // evaluate safely-ish by limiting allowed characters and using Function in strict mode
                    const result = Function(`"use strict"; return (${expr});`)();
                    if (typeof result === 'number' && isFinite(result)) {
                        value = result;
                    }
                } catch (e) {
                    // leave value as-is on error
                }
            }

            if (this.type === 'number' && !String(value).endsWith('%')) {
                const num = parseFloat(value);

                if (this.step % 1 === 0) {
                    value = isNaN(num) ? value : Math.round(num);
                } else {
                    value = isNaN(num) ? value : num;
                }
            }


            if (String(value).endsWith("%")) {
                value = String(value).slice(0, -1);
                if (this.minPercentage !== null) {
                    value = Math.max(this.minPercentage, value);
                }
                if (this.maxPercentage !== null) {
                    value = Math.min(this.maxPercentage, value);
                }
                value = String(value) + "%";
            } else {
                if (this.min !== null) {
                    value = Math.max(this.min, value);
                }
                if (this.max !== null) {
                    value = Math.min(this.max, value);
                }
            }

            if (this.forceStringOutput) {
                value = String(value);
            }

            this.$emit('update:modelValue', value);

            return value
        },
        onblur(value) {
            this.update(value);
        },
        onenter(event) {
            event.target.blur();
        },
        onarrow(event) {
            let value = this.cleanValue(event.target.value);
            if (this.type == "number") {
                if (value !== event.target.value) return;

                let isPercentage = String(value).endsWith("%");
                if (isPercentage) {
                    value = String(value).slice(0, -1);
                }

                let newValue = Number(value) + Number(event.key === "ArrowUp" ? this.step : -this.step);
                newValue = Math.round((newValue + Number.EPSILON) * 100000) / 100000;
                
                if (isPercentage) {
                    newValue = String(newValue) + "%";
                }
                
                event.target.value = this.update(newValue);
                event.target.selectionEnd = event.target.value.length; // move cursor to end
            }
        },
    }
}
</script>