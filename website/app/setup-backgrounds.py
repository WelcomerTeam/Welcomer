import shutil
import os
import sys
from PIL import Image, ImageOps

if len(sys.argv) < 3:
    print("setup-backgrounds.py <png folder source> <webp folder destination>")
    sys.exit(1)

# Copy the backgrounds.json file to the src directory
source_file = "../../welcomer-images/service/backgrounds.json"
destination_file = "src/backgrounds.json"

shutil.copy(source_file, destination_file)

print(f'Copied {source_file} to {destination_file}')

# Define the directory where the PNG files are located
source_directory = sys.argv[1]
destination_directory = sys.argv[2]

output = []

# Loop through all files in the directory
for filename in os.listdir(source_directory):
    if filename.endswith('.png'):
        path = os.path.join(destination_directory, f'{filename[:-4]}.webp')

        if os.path.exists(path):
            print(f'{path} already exists, skipping')
            continue

        # Open the PNG file
        image = Image.open(os.path.join(source_directory, filename))

        new_size = (1000, 300)

        # Resize the image to fit within 1000x300
        image = ImageOps.fit(image, new_size, method=Image.LANCZOS, centering=(0.5, 0.5))

        # Save the image as WEBP with quality 95
        image.save(path, quality=95)

        print(f'Converted {filename} to {path}')
