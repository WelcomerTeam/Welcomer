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
