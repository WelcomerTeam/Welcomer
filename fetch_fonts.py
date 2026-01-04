import json
import logging
import requests
from pathlib import Path

# Configure logging
logging.basicConfig(level=logging.INFO, format='%(levelname)s: %(message)s')

# Constants
FONTS_JSON = "fonts.json"
ROOT_FOLDER = Path("welcomer-headless-shell/fonts")

def load_fonts():
    try:
        with open(FONTS_JSON, "r") as f:
            return json.load(f)
    except FileNotFoundError:
        logging.error(f"{FONTS_JSON} not found.")
        raise
    except json.JSONDecodeError:
        logging.error(f"Invalid JSON in {FONTS_JSON}.")
        raise

def download_font_if_needed(url, file_path):
    if file_path.exists():
        return
    try:
        response = requests.get(url)
        response.raise_for_status()
        with open(file_path, "wb") as f:
            f.write(response.content)
        logging.info(f"Downloaded {file_path.name}")
    except requests.RequestException as e:
        logging.error(f"Failed to download {url}: {e}")

def clean_old_fonts(used_files):
    for file_path in ROOT_FOLDER.iterdir():
        if file_path.is_file() and file_path.name not in used_files:
            file_path.unlink()
            logging.info(f"Removed old font file: {file_path.name}")

def main():
    ROOT_FOLDER.mkdir(parents=True, exist_ok=True)
    fonts = load_fonts()
    used_files = set()

    for font_key, font in fonts.items():
        for weight_key, weight in font["weights"].items():
            logging.info(f"Processing {font_key} with weight {weight_key}")
            url = f"https://fonts.googleapis.com/css2?family={font['name'].replace(' ', '+')}:wght@{weight}"
            try:
                response = requests.get(url)
                response.raise_for_status()
                data = response.text
                lines = [line.strip() for line in data.split("\n") if "url(" in line]
                for line in lines:
                    font_url = line.split("url(")[1].split(")")[0]
                    file_name = font_url.split("/")[-1].split("?")[0]
                    file_path = ROOT_FOLDER / file_name
                    download_font_if_needed(font_url, file_path)
                    used_files.add(file_name)
            except requests.RequestException as e:
                logging.error(f"Failed to fetch CSS for {font_key} {weight_key}: {e}")

    clean_old_fonts(used_files)

if __name__ == "__main__":
    main()
