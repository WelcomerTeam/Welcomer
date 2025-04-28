import { toHTML } from "@/components/discord-markdown";
import store from "@/store";

export function getHexColor(number) {
    return "#" + (number >>> 0).toString(16).slice(-6);
}

export function navigateToErrors() {
    let error = document.querySelector(".errors");
    
    if (error) {
        error.scrollIntoView({
            behavior: "smooth",
            block: "center",
            inline: "center",
        });    
    } else {
        console.warn(
            "No error to scroll into view. Is there a missing error message?"
        );
    }
}

export function isValidJson(json) {
    try {
        JSON.parse(json);
    } catch (e) {
        return false;
    }
    return true;
}

export function getSuccessToast() {
    return {
        title: "Changes saved.",
        icon: "check",
        class: "text-green-500 bg-green-100",
    }
}

export function getValidationToast() {
    return {
        title: "Please fix any errors before submitting",
        icon: "xmark",
        class: "text-red-500 bg-red-100",
    }
}

export function getErrorToast(error) {
    return {
        title: error,
        icon: "xmark",
        class: "text-red-500 bg-red-100",
    }
}

export function getRolePermissionListAsString(permissions) {
    const nameMap = {
        0x0000000000000002: "Kick Members",
        0x0000000000000004: "Ban Members",
        0x0000000000000008: "Administrator",
        0x0000000000000010: "Manage Channels",
        0x0000000000000020: "Manage Server",
        0x0000000000002000: "Manage Messages",
        0x0000000010000000: "Manage Roles",
        0x0000000020000000: "Manage Webhooks",
        0x0000000040000000: "Manage Emojis",
        0x0000000400000000: "Manage Threads",
        0x0000010000000000: "Moderate Members",
    };
    
    var roleNames = [];
    
    for (const [permission, name] of Object.entries(nameMap)) {
        if (permissions & permission) {
            roleNames.push(name);
        }
    }
    
    if (roleNames.length === 0) {
        return "None";
    }
    
    return roleNames.join(", ");
}

export function ordinal(number) {
    console.log(number);
    const suffixes = ["th", "st", "nd", "rd"];

    return number.toString() + suffixes[(number % 100 >= 11 && number % 100 <= 13) ? 0 : (number % 10 < 4 ? number % 10 : 0)];
}

export function formatText(text) {
    if (!text || !text.includes("{{")) {
        return text;
    }
    
    const rules = {
        "{{User.ID}}": store.getters.getCurrentUser?.id,
        "{{User.Name}}": store.getters.getCurrentUser?.global_name,
        "{{User.Username}}": store.getters.getCurrentUser?.username,
        "{{User.Discriminator}}": store.getters.getCurrentUser?.discriminator,
        "{{User.GlobalName}}": store.getters.getCurrentUser?.global_name,
        "{{User.Mention}}": `<@${store.getters.getCurrentUser?.id}>`,
        "{{User.Avatar}}": store.getters.getCurrentUser?.avatar
            ? `https://cdn.discordapp.com/avatars/${store.getters.getCurrentUser?.id}/${store.getters.getCurrentUser?.avatar}.png`
            : `https://cdn.discordapp.com/embed/avatars/${(store.getters.getCurrentUser?.id >> 22) % 6}.png`,
        "{{User.Bot}}": "False",
        "{{User.Pending}}": "False",
        "{{Guild.ID}}": store.getters.getCurrentSelectedGuild?.id,
        "{{Guild.Name}}": store.getters.getCurrentSelectedGuild?.name,
        "{{Guild.Icon}}": `https://cdn.discordapp.com/icons/${store.getters.getCurrentSelectedGuild?.id}/${store.getters.getCurrentSelectedGuild?.icon}.png`,
        "{{Guild.Splash}}": `https://cdn.discordapp.com/splashes/${store.getters.getCurrentSelectedGuild?.id}/${store.getters.getCurrentSelectedGuild?.splash}.png`,
        "{{Guild.Members}}": store.getters.getCurrentSelectedGuild?.member_count,
        "{{Ordinal(Guild.Members)}}": ordinal(store.getters.getCurrentSelectedGuild?.member_count),
        "{{Guild.Banner}}": `https://cdn.discordapp.com/banners/${store.getters.getCurrentSelectedGuild?.id}/${store.getters.getCurrentSelectedGuild?.banner}.png`,
        "{{Invite.Code}}": "Unknown",
        "{{Invite.Uses}}": "0",
        "{{Invite.Inviter}}": "Unknown",
        "{{Invite.ChannelID}}": "0",
        "{{Invite.CreatedAt}}": "0",
        "{{Invite.ExpiresAt}}": "0",
        "{{Invite.MaxAge}}": "0",
        "{{Invite.MaxUses}}": "0",
        "{{Invite.Temporary}}": "False",
    };

    for (const [key, value] of Object.entries(rules)) {
        text = text.replace(key, value);
    }

    return text;
}

export function marked(input, embed) {
    if (input) {
        return toHTML(formatText(input), {
            embed: embed,
            discordCallback: {
                user: function (user) {
                    if (user.id == "143090142360371200") {
                        return `@ImRock`;
                    }

                    if (store.getters.getCurrentUser?.id == user.id) {
                        return `@${store.getters.getCurrentUser.global_name}`;
                    }

                    return `@${user.id}`;
                },
                channel: function (channel) {
                    var channelName = store.getters.getGuildChannels.find((c) => c.id == channel.id)?.name;

                    if (channelName) {
                        return `#${channelName}`;
                    }

                    console.warn(`Channel ${channel.id} not found in guild channels. Channels: ${store.getters.getGuildChannels}`);

                    return `#${channel.id}`;
                },
                role: function (role) {
                    var roleName = store.getters.getGuildRoles.find((r) => r.id == role.id)?.name;

                    if (roleName) {
                        return `@${roleName}`;
                    }

                    console.warn(`Role ${role.id} not found in guild roles. Roles: ${store.getters.getGuildRoles}`);

                    return `@${role.id}`;
                },
                everyone: function () {
                    return `@everyone`;
                },
                here: function () {
                    return `@here`;
                },
            },
            cssModuleNames: {
                "d-emoji": "emoji",
            },
        });
    }
    return "";
}