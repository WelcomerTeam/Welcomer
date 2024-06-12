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