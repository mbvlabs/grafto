package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run cmd/generate/main.go <type> <name>")
		fmt.Println("Example: go run cmd/generate/main.go view User")
		os.Exit(1)
	}

	generateType := os.Args[1]
	resourceName := os.Args[2]

	switch generateType {
	case "view":
		if err := generateView(resourceName); err != nil {
			fmt.Printf("Error generating view: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Generated view for %s\n", resourceName)
	default:
		fmt.Printf("Unknown generate type: %s\n", generateType)
		os.Exit(1)
	}
}

func generateView(resourceName string) error {
	pluralName := strings.ToLower(resourceName) + "s"
	viewPath := filepath.Join("views", pluralName+".templ")

	if _, err := os.Stat(viewPath); err == nil {
		return fmt.Errorf("file %s already exists", viewPath)
	}

	content, err := generateViewContent(resourceName, pluralName)
	if err != nil {
		return fmt.Errorf("failed to generate content: %w", err)
	}

	if err := os.WriteFile(viewPath, []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to write file: %w", err)
	}

	return nil
}

type TemplateData struct {
	ResourceName string
	PluralName   string
}

func generateViewContent(resourceName, pluralName string) (string, error) {
	tmplPath := filepath.Join("cmd", "generate", "templates", "view.templ")
	tmplContent, err := os.ReadFile(tmplPath)
	if err != nil {
		return "", fmt.Errorf("failed to read template file: %w", err)
	}

	tmpl, err := template.New("view").Parse(string(tmplContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse template: %w", err)
	}

	data := TemplateData{
		ResourceName: resourceName,
		PluralName:   pluralName,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("failed to execute template: %w", err)
	}

	return buf.String(), nil
}
