package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultSpecLocation = "http://localhost:4096/doc"

func main() {
	os.Exit(run())
}

func run() int {
	var configPath string
	var patchedSpecPath string

	flag.StringVar(&configPath, "config", "", "path to the oapi-codegen config file")
	flag.StringVar(&patchedSpecPath, "patched-spec", "", "optional path to write the patched OpenAPI document")
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Usage: go tool opencodegen [flags] [spec-url-or-path]\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Defaults:\n")
		fmt.Fprintf(flag.CommandLine.Output(), "  spec source: %s\n", defaultSpecLocation)
		fmt.Fprintf(flag.CommandLine.Output(), "  config:      <module-root>/codegen.config.yaml\n\n")
		fmt.Fprintf(flag.CommandLine.Output(), "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	if flag.NArg() > 1 {
		fmt.Fprintln(os.Stderr, "opencodegen: expected at most one positional argument for the spec location")
		flag.Usage()
		return 2
	}

	specLocation := defaultSpecLocation
	if flag.NArg() == 1 {
		specLocation = flag.Arg(0)
	}

	if err := generate(specLocation, configPath, patchedSpecPath); err != nil {
		fmt.Fprintln(os.Stderr, "opencodegen:", err)
		return 1
	}

	return 0
}

func generate(specLocation, configPath, patchedSpecPath string) error {
	moduleRoot, err := findModuleRoot()
	if err != nil {
		return err
	}

	if configPath == "" {
		configPath = filepath.Join(moduleRoot, "codegen.config.yaml")
	}

	configPath, err = filepath.Abs(configPath)
	if err != nil {
		return fmt.Errorf("resolve config path: %w", err)
	}

	if _, err := os.Stat(configPath); err != nil {
		return fmt.Errorf("read config file %q: %w", configPath, err)
	}

	rawDocument, err := readSpec(specLocation)
	if err != nil {
		return err
	}

	document, err := decodeSpec(rawDocument)
	if err != nil {
		return err
	}

	document = normalize(document)
	patchOpenAPI31(document)

	patchedPath, cleanup, err := writePatchedSpec(document, patchedSpecPath)
	if err != nil {
		return err
	}
	defer cleanup()

	command := exec.Command("go", "tool", "oapi-codegen", "-config", configPath, patchedPath)
	command.Dir = filepath.Dir(configPath)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	if err := command.Run(); err != nil {
		return fmt.Errorf("run oapi-codegen: %w", err)
	}

	if patchedSpecPath != "" {
		fmt.Fprintf(os.Stdout, "patched OpenAPI written to %s\n", patchedPath)
	}

	return nil
}

func findModuleRoot() (string, error) {
	workingDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	current := workingDir
	for {
		candidate := filepath.Join(current, "go.mod")
		if _, err := os.Stat(candidate); err == nil {
			return current, nil
		}

		parent := filepath.Dir(current)
		if parent == current {
			return "", errors.New("could not find go.mod in current directory or any parent")
		}

		current = parent
	}
}

func readSpec(location string) ([]byte, error) {
	if strings.HasPrefix(location, "http://") || strings.HasPrefix(location, "https://") {
		response, err := http.Get(location)
		if err != nil {
			return nil, fmt.Errorf("fetch spec %q: %w", location, err)
		}
		defer response.Body.Close()

		if response.StatusCode < 200 || response.StatusCode >= 300 {
			return nil, fmt.Errorf("fetch spec %q: unexpected HTTP status %s", location, response.Status)
		}

		body, err := io.ReadAll(response.Body)
		if err != nil {
			return nil, fmt.Errorf("read spec response %q: %w", location, err)
		}

		return body, nil
	}

	path, err := filepath.Abs(location)
	if err != nil {
		return nil, fmt.Errorf("resolve spec path %q: %w", location, err)
	}

	body, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read spec file %q: %w", path, err)
	}

	return body, nil
}

func decodeSpec(document []byte) (any, error) {
	var decoded any

	if err := json.Unmarshal(document, &decoded); err == nil {
		return decoded, nil
	}

	if err := yaml.Unmarshal(document, &decoded); err == nil {
		return decoded, nil
	}

	return nil, errors.New("spec was not valid JSON or YAML")
}

func normalize(value any) any {
	switch typed := value.(type) {
	case map[string]any:
		normalized := make(map[string]any, len(typed))
		for key, entry := range typed {
			normalized[key] = normalize(entry)
		}
		return normalized
	case map[any]any:
		normalized := make(map[string]any, len(typed))
		for key, entry := range typed {
			normalized[fmt.Sprint(key)] = normalize(entry)
		}
		return normalized
	case []any:
		normalized := make([]any, len(typed))
		for index, entry := range typed {
			normalized[index] = normalize(entry)
		}
		return normalized
	default:
		return typed
	}
}

func patchOpenAPI31(value any) {
	switch typed := value.(type) {
	case map[string]any:
		for key, entry := range typed {
			patchOpenAPI31(entry)
			typed[key] = entry
		}

		patchExclusiveBound(typed, "exclusiveMinimum", "minimum")
		patchExclusiveBound(typed, "exclusiveMaximum", "maximum")
		patchNullableTypeArray(typed)
		patchNullableSchemaUnion(typed, "anyOf")
		patchNullableSchemaUnion(typed, "oneOf")

		if version, ok := typed["openapi"].(string); ok && strings.HasPrefix(version, "3.1") {
			typed["openapi"] = "3.0.3"
		}
	case []any:
		for _, entry := range typed {
			patchOpenAPI31(entry)
		}
	}
}

func patchExclusiveBound(schema map[string]any, exclusiveKey, boundKey string) {
	value, ok := schema[exclusiveKey]
	if !ok {
		return
	}

	if !isNumeric(value) {
		return
	}

	if _, exists := schema[boundKey]; !exists {
		schema[boundKey] = value
	}
	schema[exclusiveKey] = true
}

func patchNullableTypeArray(schema map[string]any) {
	rawTypes, ok := schema["type"].([]any)
	if !ok || len(rawTypes) != 2 {
		return
	}

	var nonNull string
	hasNull := false

	for _, rawType := range rawTypes {
		typeName, ok := rawType.(string)
		if !ok {
			return
		}

		if typeName == "null" {
			hasNull = true
			continue
		}

		if nonNull != "" {
			return
		}

		nonNull = typeName
	}

	if !hasNull || nonNull == "" {
		return
	}

	schema["type"] = nonNull
	schema["nullable"] = true
}

func patchNullableSchemaUnion(schema map[string]any, field string) {
	rawVariants, ok := schema[field].([]any)
	if !ok || len(rawVariants) != 2 {
		return
	}

	var nonNullSchema map[string]any
	hasNullSchema := false

	for _, rawVariant := range rawVariants {
		variant, ok := rawVariant.(map[string]any)
		if !ok {
			return
		}

		if isNullSchema(variant) {
			hasNullSchema = true
			continue
		}

		if nonNullSchema != nil {
			return
		}

		nonNullSchema = variant
	}

	if !hasNullSchema || nonNullSchema == nil {
		return
	}

	delete(schema, field)
	for key, value := range nonNullSchema {
		if _, exists := schema[key]; !exists {
			schema[key] = value
		}
	}
	schema["nullable"] = true
}

func isNullSchema(schema map[string]any) bool {
	if len(schema) != 1 {
		return false
	}

	typeName, ok := schema["type"].(string)
	return ok && typeName == "null"
}

func isNumeric(value any) bool {
	switch value.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64,
		json.Number:
		return true
	default:
		return false
	}
}

func writePatchedSpec(document any, destination string) (string, func(), error) {
	encoded, err := json.MarshalIndent(document, "", "  ")
	if err != nil {
		return "", nil, fmt.Errorf("encode patched spec: %w", err)
	}

	if destination != "" {
		path, err := filepath.Abs(destination)
		if err != nil {
			return "", nil, fmt.Errorf("resolve patched spec path %q: %w", destination, err)
		}

		if err := os.WriteFile(path, encoded, 0o644); err != nil {
			return "", nil, fmt.Errorf("write patched spec %q: %w", path, err)
		}

		return path, func() {}, nil
	}

	file, err := os.CreateTemp("", "opencodegen-*.json")
	if err != nil {
		return "", nil, fmt.Errorf("create temp spec file: %w", err)
	}

	if _, err := file.Write(encoded); err != nil {
		file.Close()
		os.Remove(file.Name())
		return "", nil, fmt.Errorf("write temp spec file: %w", err)
	}

	if err := file.Close(); err != nil {
		os.Remove(file.Name())
		return "", nil, fmt.Errorf("close temp spec file: %w", err)
	}

	return file.Name(), func() {
		_ = os.Remove(file.Name())
	}, nil
}
