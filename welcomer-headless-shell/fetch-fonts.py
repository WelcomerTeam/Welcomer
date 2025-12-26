import os
import requests
import shutil

fonts = {
    "Balsamiq Sans": {
        "name": "Balsamiq Sans",
        "defaultWeight": "regular",
        "weights": {
            "regular": "400",
            "bold": "700",
        },
    },
    "Fredoka": {
        "name": "Fredoka",
        "defaultWeight": "regular",
        "weights": {
            "300": "300",
            "regular": "400",
            "500": "500",
            "600": "600",
            "bold": "700",
        },
    },
    "Inter": {
        "name": "Inter",
        "defaultWeight": "regular",
        "weights": {
            "100": "100",
            "200": "200",
            "300": "300",
            "regular": "400",
            "500": "500",
            "600": "600",
            "bold": "700",
            "800": "800",
            "900": "900",
        },
    },
    "Luckiest Guy": {
        "name": "Luckiest Guy",
        "defaultWeight": "regular",
        "weights": {
            "regular": "400",
        },
    },
    "Mada": {
        "name": "Mada",
        "defaultWeight": "regular",
        "weights": {
            "200": "200",
            "300": "300",
            "regular": "400",
            "500": "500",
            "600": "600",
            "bold": "700",
            "800": "800",
            "900": "900",
        },
    },
    "Nunito": {
        "name": "Nunito",
        "defaultWeight": "regular",
        "weights": {
            "200": "200",
            "300": "300",
            "regular": "400",
            "600": "600",
            "bold": "700",
            "800": "800",
            "900": "900",
            "1000": "1000",
        },
    },
    "Poppins": {
        "name": "Poppins",
        "defaultWeight": "regular",
        "weights": {
            "100": "100",
            "200": "200",
            "300": "300",
            "regular": "400",
            "500": "500",
            "600": "600",
            "bold": "700",
            "800": "800",
            "900": "900",
        },
    },
    "Raleway": {
        "name": "Raleway",
        "defaultWeight": "regular",
        "weights": {
            "100": "100",
            "200": "200",
            "300": "300",
            "regular": "400",
            "500": "500",
            "600": "600",
            "bold": "700",
            "800": "800",
            "900": "900",
        },
    },
}

# with open("fonts.css", "w") as fontFile:
for fontKey in fonts:
    font = fonts[fontKey]

    for weightKey in font["weights"]:
        weight = font["weights"][weightKey]

        print(fontKey, weightKey)

        url = (
            f"https://fonts.googleapis.com/css2?family={font['name'].replace(' ', '+')}:wght@"
            + weight
        )
        req = requests.get(url)
        data = req.text

        lines = [line.strip() for line in data.split("\n") if "url(" in line]
        files = {}

        for line in lines:
            url = line.split("url(")[1].split(")")[0]

            format = (line.split("format(")[1].split(")")[0].replace('"', "").replace("'", ""))

            file_name = "fonts/" + url.split("/")[-1].split("?")[0]

            if not os.path.exists(file_name):
                font_req = requests.get(url)
                print("Downloading font "+ fontKey+ " with weight "+ weightKey+ ": "+ str(font_req.status_code))

                with open(file_name, "wb") as f:
                    f.write(font_req.content)

            # fontFile.write(f"/* {fontKey} - {weightKey} */\n")
            # fontFile.write(data.replace(url, "%FONT_BASE%/" + file_name) + "\n\n")
