package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	defaultOutPath = "responses_ingress.go"
	envSchemaPath  = "EZ_CONTRACT_RESPONSES_SCHEMA"
)

type responsesSchemaDoc struct {
	ResponsesRequest struct {
		FieldPolicy responsesFieldPolicy `yaml:"x-ez-ingress-field-policy"`
	} `yaml:"ResponsesRequest"`
	ResponsesResponse struct {
		StreamEvents responsesStreamEvents `yaml:"x-ez-ingress-stream-events"`
	} `yaml:"ResponsesResponse"`
}

type responsesFieldPolicy struct {
	PassThrough []string `yaml:"pass_through"`
	Reject      []string `yaml:"reject"`
}

type responsesStreamEvents struct {
	OutputTextDelta     string `yaml:"output_text_delta"`
	OutputToolCallDelta string `yaml:"output_tool_call_delta"`
	Completed           string `yaml:"completed"`
	Error               string `yaml:"error"`
}

type responsesIngressContract struct {
	SchemaPath        string
	PassThroughFields []string
	RejectFields      []string
	Events            responsesStreamEvents
}

func main() {
	var schemaPath string
	var outPath string
	flag.StringVar(&schemaPath, "schema", "", "path to ez-contract responses schema")
	flag.StringVar(&outPath, "out", defaultOutPath, "output Go file path")
	flag.Parse()

	resolvedSchemaPath, err := resolveSchemaPath(schemaPath)
	if err != nil {
		fail(err)
	}
	contract, err := loadContract(resolvedSchemaPath)
	if err != nil {
		fail(err)
	}

	content, err := render(contract)
	if err != nil {
		fail(err)
	}
	if err := os.WriteFile(outPath, content, 0o644); err != nil {
		fail(fmt.Errorf("write output file %q: %w", outPath, err))
	}
}

func resolveSchemaPath(explicit string) (string, error) {
	candidates := make([]string, 0, 6)
	if v := strings.TrimSpace(explicit); v != "" {
		candidates = append(candidates, v)
	}
	if v := strings.TrimSpace(os.Getenv(envSchemaPath)); v != "" {
		candidates = append(candidates, v)
	}
	candidates = append(candidates,
		"../../ez-contract/schemas/responses/responses.yaml",
		"../ez-contract/schemas/responses/responses.yaml",
		"../../../ez-contract/schemas/responses/responses.yaml",
		"./ez-contract/schemas/responses/responses.yaml",
	)

	for _, candidate := range candidates {
		resolved := filepath.Clean(candidate)
		if _, err := os.Stat(resolved); err == nil {
			return resolved, nil
		}
	}

	return "", fmt.Errorf("responses schema not found; pass -schema or set %s", envSchemaPath)
}

func loadContract(schemaPath string) (*responsesIngressContract, error) {
	raw, err := os.ReadFile(schemaPath)
	if err != nil {
		return nil, fmt.Errorf("read schema %q: %w", schemaPath, err)
	}

	var doc responsesSchemaDoc
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, fmt.Errorf("parse yaml schema %q: %w", schemaPath, err)
	}

	pass, err := normalizeFields(doc.ResponsesRequest.FieldPolicy.PassThrough, "pass_through")
	if err != nil {
		return nil, err
	}
	reject, err := normalizeFields(doc.ResponsesRequest.FieldPolicy.Reject, "reject")
	if err != nil {
		return nil, err
	}
	if err := ensureNoOverlap(pass, reject); err != nil {
		return nil, err
	}

	events, err := normalizeEvents(doc.ResponsesResponse.StreamEvents)
	if err != nil {
		return nil, err
	}

	return &responsesIngressContract{
		SchemaPath:        schemaPathForComment(schemaPath),
		PassThroughFields: pass,
		RejectFields:      reject,
		Events:            events,
	}, nil
}

func schemaPathForComment(path string) string {
	cleaned := filepath.Clean(path)
	if rel, err := filepath.Rel(".", cleaned); err == nil {
		return filepath.ToSlash(rel)
	}
	return filepath.ToSlash(cleaned)
}

func normalizeFields(fields []string, label string) ([]string, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("schema metadata %q is empty", label)
	}

	set := make(map[string]struct{}, len(fields))
	for _, field := range fields {
		normalized := strings.ToLower(strings.TrimSpace(field))
		if normalized == "" {
			return nil, fmt.Errorf("schema metadata %q contains empty field name", label)
		}
		set[normalized] = struct{}{}
	}
	out := make([]string, 0, len(set))
	for field := range set {
		out = append(out, field)
	}
	sort.Strings(out)
	return out, nil
}

func ensureNoOverlap(pass, reject []string) error {
	if len(pass) == 0 || len(reject) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(pass))
	for _, field := range pass {
		seen[field] = struct{}{}
	}
	for _, field := range reject {
		if _, ok := seen[field]; ok {
			return fmt.Errorf("field %q appears in both pass_through and reject lists", field)
		}
	}
	return nil
}

func normalizeEvents(events responsesStreamEvents) (responsesStreamEvents, error) {
	var out responsesStreamEvents
	out.OutputTextDelta = strings.TrimSpace(events.OutputTextDelta)
	out.OutputToolCallDelta = strings.TrimSpace(events.OutputToolCallDelta)
	out.Completed = strings.TrimSpace(events.Completed)
	out.Error = strings.TrimSpace(events.Error)

	switch {
	case out.OutputTextDelta == "":
		return responsesStreamEvents{}, errors.New("missing output_text_delta stream event")
	case out.OutputToolCallDelta == "":
		return responsesStreamEvents{}, errors.New("missing output_tool_call_delta stream event")
	case out.Completed == "":
		return responsesStreamEvents{}, errors.New("missing completed stream event")
	case out.Error == "":
		return responsesStreamEvents{}, errors.New("missing error stream event")
	}
	return out, nil
}

func render(contract *responsesIngressContract) ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("// Code generated by go generate; DO NOT EDIT.\n")
	fmt.Fprintf(&buf, "// Source: %s\n\n", contract.SchemaPath)
	buf.WriteString("package contract\n\n")
	buf.WriteString("import (\n")
	buf.WriteString("\t\"fmt\"\n")
	buf.WriteString("\t\"sort\"\n")
	buf.WriteString("\t\"strings\"\n")
	buf.WriteString(")\n\n")
	buf.WriteString("const (\n")
	fmt.Fprintf(&buf, "\tResponsesEventOutputTextDelta     = %q\n", contract.Events.OutputTextDelta)
	fmt.Fprintf(&buf, "\tResponsesEventOutputToolCallDelta = %q\n", contract.Events.OutputToolCallDelta)
	fmt.Fprintf(&buf, "\tResponsesEventCompleted           = %q\n", contract.Events.Completed)
	fmt.Fprintf(&buf, "\tResponsesEventError               = %q\n", contract.Events.Error)
	buf.WriteString(")\n\n")
	buf.WriteString("type ResponsesFieldDecision string\n\n")
	buf.WriteString("const (\n")
	buf.WriteString("\tResponsesFieldDecisionPassThrough ResponsesFieldDecision = \"pass_through\"\n")
	buf.WriteString("\tResponsesFieldDecisionReject      ResponsesFieldDecision = \"reject\"\n")
	buf.WriteString("\tResponsesFieldDecisionDegrade     ResponsesFieldDecision = \"degrade\"\n")
	buf.WriteString(")\n\n")

	writeFieldMap(&buf, "responsesPassThroughRequestFields", contract.PassThroughFields)
	buf.WriteString("\n")
	writeFieldMap(&buf, "responsesRejectedRequestFields", contract.RejectFields)
	buf.WriteString("\n")

	buf.WriteString("func ResponsesRequestFieldDecision(field string) ResponsesFieldDecision {\n")
	buf.WriteString("\tnormalized := strings.ToLower(strings.TrimSpace(field))\n")
	buf.WriteString("\tif normalized == \"\" {\n")
	buf.WriteString("\t\treturn ResponsesFieldDecisionDegrade\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tif _, ok := responsesPassThroughRequestFields[normalized]; ok {\n")
	buf.WriteString("\t\treturn ResponsesFieldDecisionPassThrough\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tif _, ok := responsesRejectedRequestFields[normalized]; ok {\n")
	buf.WriteString("\t\treturn ResponsesFieldDecisionReject\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\treturn ResponsesFieldDecisionDegrade\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func ValidateResponsesRequestFields(payload map[string]any) error {\n")
	buf.WriteString("\tfor key := range payload {\n")
	buf.WriteString("\t\tif ResponsesRequestFieldDecision(key) == ResponsesFieldDecisionReject {\n")
	buf.WriteString("\t\t\treturn fmt.Errorf(\"unsupported field %q\", strings.ToLower(strings.TrimSpace(key)))\n")
	buf.WriteString("\t\t}\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\treturn nil\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func ResponsesPassThroughRequestFields() []string {\n")
	buf.WriteString("\treturn sortedResponseFields(responsesPassThroughRequestFields)\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func ResponsesRejectedRequestFields() []string {\n")
	buf.WriteString("\treturn sortedResponseFields(responsesRejectedRequestFields)\n")
	buf.WriteString("}\n\n")

	buf.WriteString("func sortedResponseFields(src map[string]struct{}) []string {\n")
	buf.WriteString("\tif len(src) == 0 {\n")
	buf.WriteString("\t\treturn nil\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tout := make([]string, 0, len(src))\n")
	buf.WriteString("\tfor field := range src {\n")
	buf.WriteString("\t\tout = append(out, field)\n")
	buf.WriteString("\t}\n")
	buf.WriteString("\tsort.Strings(out)\n")
	buf.WriteString("\treturn out\n")
	buf.WriteString("}\n")

	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return nil, fmt.Errorf("format generated source: %w", err)
	}
	return formatted, nil
}

func writeFieldMap(buf *bytes.Buffer, mapName string, fields []string) {
	fmt.Fprintf(buf, "var %s = map[string]struct{}{\n", mapName)
	for _, field := range fields {
		fmt.Fprintf(buf, "\t%q: {},\n", field)
	}
	buf.WriteString("}\n")
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "genresponsesingress: %v\n", err)
	os.Exit(1)
}
