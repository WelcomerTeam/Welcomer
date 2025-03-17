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
    return `https://discord.com/oauth2/authorize?client_id=${bot_id}&scope=bot%20applications.commands&permissions=${BotPermissions}${guild_id ? '&guild_id='+guild_id : ''}`
}

export const OpenBotInvite = (bot_id, guild_id, callback) => {
    TryOpenURLInPopup(GetBotInvite(bot_id, guild_id), callback)
}

export const OpenPatreonLink = (callback) => {
    TryOpenURLInPopup("/patreon_link", callback)
}

export const TryOpenURLInPopup = (url, callback) => {
    const padding = 64

    const width = Math.min(550, window.outerWidth-(padding*2));
    const height = Math.min(800, window.outerHeight-(padding*2));
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