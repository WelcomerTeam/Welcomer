
export const AnalyticsTab_Overview = "Overview";
export const AnalyticsTab_Retention = "Retention";
export const AnalyticsTab_UserActivity = "User Activity";
export const AnalyticsTab_UserDemographics = "User Demographics";
export const AnalyticsTab_MessageActivity = "Message Activity";
export const AnalyticsTab_VoiceActivity = "Voice Activity";
export const AnalyticsTab_RoleActivity = "Role Activity";
export const AnalyticsTab_Invites = "Invites";
export const AnalyticsTab_Borderwall = "Borderwall";

export const DefaultAnalyticsTab = AnalyticsTab_Overview;

export const AnalyticsTabs = [
    AnalyticsTab_Overview,
    AnalyticsTab_Retention,
    AnalyticsTab_UserActivity,
    AnalyticsTab_UserDemographics,
    AnalyticsTab_MessageActivity,
    AnalyticsTab_VoiceActivity,
    AnalyticsTab_RoleActivity,
    AnalyticsTab_Invites,
    AnalyticsTab_Borderwall
];

export const ChartOptions = (callbacks) => {
    return {
        responsive: true,
        maintainAspectRatio: false,

        animation: {
            duration: 700,
            easing: 'easeOutQuart',
        },

        layout: {
            padding: 10,
        },

        interaction: {
            mode: 'index',
            intersect: false,
        },

        plugins: {
            legend: {
                display: false,
            },

            tooltip: {
                enabled: true,

                backgroundColor: getComputedStyle(document.body).getPropertyValue('background-color'),
                titleColor: getComputedStyle(document.body).getPropertyValue('color'),
                bodyColor: getComputedStyle(document.body).getPropertyValue('color'),

                padding: 10,
                cornerRadius: 8,

                displayColors: false,

                callbacks: callbacks,
            },
        },
    }
}

export const GetDatasetsLine = (datasets) => {
    if (datasets.length === 0) {
        throw new Error("At least one dataset is required.");
    }

    let valueZeroLabelCount = datasets[0].labels.length;

    for (let i = 1; i < datasets.length; i++) {
        if (datasets[i].labels.length > valueZeroLabelCount) {
            throw new Error("All datasets must have the same number of labels.");
        }
    }

    return {
        labels: datasets[0].labels,
        datasets: datasets,


        DontBeginAtZero() {
            return {
                ...this,
                beginAtZero: false,
            }
        }
    }
}

export const GetDatasetBar = (colour, label, data) => {
    return {
        type: 'bar',
        label: label,
        labels: data.map((item) => item[0]),
        data: data.map((item) => item[1]),

        borderColor: colour,
        backgroundColor: colour,

        borderWidth: 1,
        borderRadius: 4,

        maxBarThickness: 20,

        HasTimestamp() {
            var hasHours = this.labels.some(label => {
                const date = new Date(label);
                return date.getHours() !== 0 || date.getMinutes() !== 0;
            });

            this.labels = this.labels.map(label =>
                new Date(label).toLocaleString(undefined, {
                    day: 'numeric',
                    month: 'short',
                    hour: hasHours ? '2-digit' : undefined,
                    minute: hasHours ? '2-digit' : undefined
                })
            );

            return this;
        },

        WithBackgroundColor(color) {
            return {
                ...this,
                backgroundColor: color,
            }
        },

        Ungroup() {
            return {
                ...this,
                grouped: false,
            }
        },
    }
}

export const GetDatasetLine = (colour, label, data) => {
    return {
        type : 'line',
        label: label,
        labels: data.map((item) => item[0]),
        data: data.map((item) => item[1]),

        pointRadius: 0,
        pointHoverRadius: 0,

        tension: 0.4,

        borderColor: colour,
        backgroundColor: "transparent",
        fill: true,

        HasTimestamp() {
            var hasHours = this.labels.some(label => {
                const date = new Date(label);
                return date.getHours() !== 0 || date.getMinutes() !== 0;
            });

            this.labels = this.labels.map(label =>
                new Date(label).toLocaleString(undefined, {
                    day: 'numeric',
                    month: 'short',
                    hour: hasHours ? '2-digit' : undefined,
                    minute: hasHours ? '2-digit' : undefined
                })
            );

            return this;
        },

        WithBorderDash(dash) {
            return {
                ...this,
                borderDash: dash,
            }
        },

        WithBorderWidth(width) {
            return {
                ...this,
                borderWidth: width,
            }
        },

        WithBackgroundColor(color) {
            return {
                ...this,
                backgroundColor: color,
            }
        },

        WithGradientBackgroundColor() {
            return {
                ...this,
                backgroundColor: (context) => {
                    const { chart } = context;
                    const { ctx, chartArea } = chart;

                    if (!chartArea) return undefined;

                    const gradient = ctx.createLinearGradient(0, chartArea.top, 0, chartArea.bottom);

                    gradient.addColorStop(0, `${this.borderColor}4D`);
                    gradient.addColorStop(1, `${this.borderColor}00`);

                    return gradient;
                },
            }
        },
    }
}