//go:build ignore

package main

import (
	"fmt"
	_ "image/jpeg"
	_ "image/png"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func getImageFiles(folderPath string) ([]string, error) {
	var imageFiles []string

	err := filepath.Walk(folderPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(info.Name()))
			if ext == ".png" || ext == ".jpg" || ext == ".jpeg" || ext == ".gif" {
				imageFiles = append(imageFiles, path)
			} else {
				fmt.Printf("Unrecognized file %s\n", path)
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return imageFiles, nil
}

func standardizeFileName(filePath string) (string, string) {
	folder := filepath.Dir(filePath)
	base := filepath.Base(filePath)
	name := strings.TrimSuffix(base, filepath.Ext(base))
	parts := strings.Split(name, "_")

	for i, part := range parts {
		parts[i] = strings.Title(part)
	}

	standardizedName := strings.Join(parts, "")
	return standardizedName, folder
}

func writeHeader(assetsFile *os.File) {
	_, _ = fmt.Fprint(assetsFile, "package service\n\n")
	_, _ = fmt.Fprint(assetsFile, "import (\n")
	_, _ = fmt.Fprint(assetsFile, "	\"bytes\"\n")
	_, _ = fmt.Fprint(assetsFile, "	_ \"embed\"\n")
	_, _ = fmt.Fprint(assetsFile, "	\"fmt\"\n")
	_, _ = fmt.Fprint(assetsFile, "	\"image\"\n")
	_, _ = fmt.Fprint(assetsFile, ")\n\n")
}

func writeDecodeBlock(assetsFile *os.File, imageFiles map[string][]string) {
	_, _ = fmt.Fprint(assetsFile, "func mustDecodeBytes(n string, src []byte) image.Image {\n")
	_, _ = fmt.Fprint(assetsFile, "	res, _, err := image.Decode(bytes.NewBuffer(src))\n")
	_, _ = fmt.Fprint(assetsFile, "	if err != nil {\n")
	_, _ = fmt.Fprint(assetsFile, "		panic(fmt.Sprintf(\"image.Decode(%s): %v\", n, err.Error()))\n")
	_, _ = fmt.Fprint(assetsFile, "	}\n\n")
	_, _ = fmt.Fprint(assetsFile, "	return res\n")
	_, _ = fmt.Fprint(assetsFile, "}\n\n")
}

func writeEmbedStatements(assetsFile *os.File, imageFiles map[string][]string) {
	for _, files := range imageFiles {
		for _, file := range files {
			fileName, folderName := standardizeFileName(file)
			_, _ = fmt.Fprintf(assetsFile, "//go:embed %s\n", file)
			_, _ = fmt.Fprintf(assetsFile, "var %sImageBytes []byte\n", removeInvalidCharacters(folderName+fileName))
			_, _ = fmt.Fprintf(assetsFile, "var %sImage = mustDecodeBytes(\"%sImage\", %sImageBytes)\n\n", removeInvalidCharacters(folderName+fileName), folderName+fileName, removeInvalidCharacters(folderName+fileName))
		}
		fmt.Fprintf(assetsFile, "\n")
	}
}

func writeMapStatements(assetsFile *os.File, imageFiles map[string][]string) {
	for folderName, files := range imageFiles {
		_, _ = fmt.Fprintf(assetsFile, "var %s = map[string]image.Image{\n", folderName)
		for _, file := range files {
			fileName, folderName := standardizeFileName(file)
			_, _ = fmt.Fprintf(assetsFile, "    \"%s\": %sImage,\n", strings.ToLower(fileName), removeInvalidCharacters(folderName+fileName))
		}
		fmt.Fprintf(assetsFile, "}\n\n")
	}
}

func removeInvalidCharacters(in string) string {
	return strings.ReplaceAll(in, "-", "")
}

func main() {
	var imageFiles map[string][]string
	imageFiles = map[string][]string{}

	for i, v := range os.Args {
		if i > 0 {
			files, err := getImageFiles(v)
			if err != nil {
				panic(err)
			}

			imageFiles[v] = files
		}
	}

	assetsFile, err := os.Create("assets.go")
	if err != nil {
		panic(err)
	}

	defer assetsFile.Close()

	writeHeader(assetsFile)
	writeMapStatements(assetsFile, imageFiles)
	writeDecodeBlock(assetsFile, imageFiles)
	writeEmbedStatements(assetsFile, imageFiles)
}
