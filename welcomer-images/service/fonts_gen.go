//go:build ignore

package main

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

func getFontFiles(folderPath string) ([]string, error) {
	var fontFiles []string

	err := filepath.Walk(folderPath, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(info.Name()))
			if ext == ".ttf" || ext == ".otf" {
				fontFiles = append(fontFiles, path)
			} else {
				fmt.Printf("Unrecognized file %s\n", path)
			}
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return fontFiles, nil
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
	_, _ = fmt.Fprint(assetsFile, "	_ \"embed\"\n")
	_, _ = fmt.Fprint(assetsFile, "	\"fmt\"\n\n")
	_, _ = fmt.Fprint(assetsFile, "	\"golang.org/x/image/font/opentype\"\n")
	_, _ = fmt.Fprint(assetsFile, "	\"golang.org/x/image/font/sfnt\"\n")
	_, _ = fmt.Fprint(assetsFile, ")\n\n")
}

func writeDecodeBlock(assetsFile *os.File, fontFiles map[string][]string) {
	_, _ = fmt.Fprint(assetsFile, "func mustDecodeFont(n string, src []byte) *sfnt.Font {\n")
	_, _ = fmt.Fprint(assetsFile, "	res, err := opentype.Parse(src)\n")
	_, _ = fmt.Fprint(assetsFile, "	if err != nil {\n")
	_, _ = fmt.Fprint(assetsFile, "		panic(fmt.Sprintf(\"opentype.Parse(%s): %v\", n, err.Error()))\n")
	_, _ = fmt.Fprint(assetsFile, "	}\n\n")
	_, _ = fmt.Fprint(assetsFile, "	return res\n")
	_, _ = fmt.Fprint(assetsFile, "}\n")
}

func writeEmbedStatements(assetsFile *os.File, fontFiles map[string][]string) {
	for _, files := range fontFiles {
		for _, file := range files {
			fileName, folderName := standardizeFileName(file)
			_, _ = fmt.Fprintf(assetsFile, "//go:embed %s\n", file)
			_, _ = fmt.Fprintf(assetsFile, "var %sFontBytes []byte\n", removeInvalidCharacters(folderName+fileName))
			_, _ = fmt.Fprintf(assetsFile, "var %sFont = &Font{\n", removeInvalidCharacters(folderName+fileName))
			_, _ = fmt.Fprintf(assetsFile, "	Font: mustDecodeFont(\"%sFont\", %sFontBytes),\n", folderName+fileName, removeInvalidCharacters(folderName+fileName))
			_, _ = fmt.Fprintf(assetsFile, "}\n\n")
		}
		fmt.Fprintf(assetsFile, "\n")
	}
}

func writeMapStatements(assetsFile *os.File, fontFiles map[string][]string) {
	for folderName, files := range fontFiles {
		_, _ = fmt.Fprintf(assetsFile, "var %s = map[string]*Font{\n", folderName)
		for _, file := range files {
			fileName, folderName := standardizeFileName(file)
			_, _ = fmt.Fprintf(assetsFile, "    \"%s\": %sFont,\n", strings.ToLower(fileName), removeInvalidCharacters(folderName+fileName))
		}
		fmt.Fprintf(assetsFile, "}\n\n")
	}
}

func removeInvalidCharacters(in string) string {
	return strings.ReplaceAll(in, "-", "")
}

func main() {
	var fontFiles map[string][]string
	fontFiles = map[string][]string{}

	for i, v := range os.Args {
		if i > 0 {
			files, err := getFontFiles(v)
			if err != nil {
				panic(err)
			}

			fontFiles[v] = files
		}
	}

	assetsFile, err := os.Create("fonts.go")
	if err != nil {
		panic(err)
	}

	defer assetsFile.Close()

	writeHeader(assetsFile)
	writeMapStatements(assetsFile, fontFiles)
	writeDecodeBlock(assetsFile, fontFiles)
	writeEmbedStatements(assetsFile, fontFiles)
}
