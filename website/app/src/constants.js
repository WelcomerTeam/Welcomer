export const LinkSupportServer = "https://discord.gg/UyUVCEcBU9";
export const LinkYoutubeChannel = "";
export const LinkPhishing = "https://discord.com/safety/common-scams-what-to-look-out-for";

export const PrimaryBotId = "330416853971107840";
export const DonatorBotId = "498519480985583636";

export const Toggle_ShowFeaturesOnPrimaryNavigation = false;
export const Toggle_ShowFeaturesOnDashboard = false;

export const BotPermissions = 399397809407; // Temporary catch-all permissions

export const PlatformTypePaypal = "paypal";
export const PlatformTypePatreon = "patreon";
export const PlatformTypeStripe = "stripe";
export const PlatformTypePaypalSubscription = "paypal_subscription";
export const PlatformTypeDiscord = "discord";

export const GetBotInvite = (bot_id, guild_id) => {
    return `https://discord.com/oauth2/authorize?client_id=${bot_id}&scope=bot%20applications.commands&permissions=${BotPermissions}${guild_id ? '&guild_id=' + guild_id : ''}`
}

export const OpenBotInvite = (bot_id, guild_id, callback) => {
    TryOpenURLInPopup(GetBotInvite(bot_id, guild_id), callback)
}

export const OpenPatreonLink = (callback) => {
    TryOpenURLInPopup("/patreon_link", callback)
}

export const TryOpenURLInPopup = (url, callback) => {
    const padding = 64

    const width = Math.min(550, window.outerWidth - (padding * 2));
    const height = Math.min(800, window.outerHeight - (padding * 2));
    const left = window.screenX + (window.outerWidth - width) / 2;
    const top = window.screenY + (window.outerHeight - height) / 2;

    var popup;

    popup = window.open(url, "_blank", `popup=1, width=${width}, height=${height}, left=${left}, top=${top}`);
    if (!popup) {
        popup = window.open(url, "_blank");
        if (!popup) {
            popup = window.open(url);
            if (!popup) {
                console.error(`Failed to open URL: ${url}`);
            }
        }
    }

    if (popup && callback) {
        const interval = setInterval(() => {
            if (popup.closed) {
                clearInterval(interval);
                callback();
            }
        }, 500);
    }

    return popup
}

export const NavigationFeatures = [
    {
        name: "Welcome Images",
        href: "/features#welcomer",
        description: "Welcome users to your servers with customizable images",
        icon: "image",
    },
    // {
    //     name: "Reaction Roles",
    //     href: "/features#reactionroles",
    //     description: "Allow users to control what roles they receive",
    //     icon: "face-laugh",
    // },
    // {
    //     name: "Moderation",
    //     href: "/features#moderation",
    //     description:
    //         "Easily moderate your guilds and have easy access to who has done what",
    //     icon: "user-shield",
    // },
    // {
    //     name: "Logging",
    //     href: "/features#logging",
    //     description:
    //         "Have easy access to all interactions with your guild both online and in guilds",
    //     icon: "boxes-packing",
    // },
    {
        name: "Temporary Channels",
        href: "/features#tempchannels",
        description: "Allow users to make temporary voice channels in your server",
        icon: "microphone-lines",
    },
    // {
    //     name: "Guild Analytics",
    //     href: "/features#analytics",
    //     description: "View information about your server such as user joins",
    //     icon: "chart-line",
    // },
    {
        name: "Borderwall",
        href: "/features#borderwall",
        description:
            "Secure your server by making them manually verify their identity",
        icon: "door-open",
    },
];

export const NavigationResources = [
    {
        name: "Status",
        href: "/status",
        description: "View the current status of the bot",
        icon: "heart-pulse",
    },
    {
        name: "Support Server",
        href: "/support",
        description:
            "Join our support server for extra support, make new suggestions and more",
        icon: "life-ring",
    },
    {
        name: "FAQ",
        href: "/faq",
        description: "Check out our FAQ, your question may already be answered",
        icon: "person-circle-question",
    },
    // {
    //   name: "Video Tutorials",
    //   href: "/tutorials",
    //   description:
    //     "View some of our video tutorials to get a better idea of how to setup the bot",
    //   icon: ["fab", "youtube"],
    // },
    {
        name: "Welcome Image Backgrounds",
        href: "/backgrounds",
        description:
            "View our list of image backgrounds you can use with welcome images",
        icon: "images",
    },
    // {
    //     name: "Custom Embed Builder",
    //     href: "/builder",
    //     description: "View our custom embed builder to see how embeds may look",
    //     icon: "tachograph-digital",
    // },
    {
        name: "Text Formatting",
        href: "/formatting",
        description:
            "View how to format your text with information about the user and more",
        icon: "paint-roller",
    },
];

export const FAQs = [
    {
        "title": "ABC123",
        "list": {
            "question": "answer"
        },
    },
    {
        "title": "Markdown Support",
        "list": {
            "question": "**Hello World** We need image support (Test)[Here]. ![alt text](https://cdn.discordapp.com/icons/341685098468343822/09cfc7fe72945a7c04ec6d3ddd01767c.webp?size=128)"
        }
    }
]

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

        scales: {
            x: {
                grid: {
                    display: false,
                },

                border: {
                    display: false,
                },

                ticks: {
                    color: getComputedStyle(document.body).getPropertyValue('color'),
                    font: {
                        size: 12,
                    },
                },
            },

            y: {
                display: false,
                beginAtZero: true,
            },
        },
    }
}

export const GetDatasetLine = (label, data) => {
    return {
        labels: data.map((item => item[0])),
        datasets: [
            {
                label: label,
                data: data.map((item) => item[1]),

                pointRadius: 0,
                pointHoverRadius: 0,

                tension: 0.4,

                borderColor: "#2F80ED",
                backgroundColor: (context) => {
                    const { chart } = context;
                    const { ctx, chartArea } = chart;

                    if (!chartArea) return undefined;

                    const gradient = ctx.createLinearGradient(0, chartArea.top, 0, chartArea.bottom);

                    gradient.addColorStop(0, `#2F80ED4D`);
                    gradient.addColorStop(1, `#2F80ED00`);

                    return gradient;
                },
                fill: true,
            },
        ],
    }
}