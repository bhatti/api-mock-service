// SPDX-License-Identifier: MIT

// Package speccompare detects breaking and non-breaking changes between two OpenAPI 3.x specs.
// Breaking changes are changes that could cause existing clients to break without a code change.
package speccompare

import (
	"fmt"
	"sort"
	"strings"

	"github.com/getkin/kin-openapi/openapi3"
)

// ChangeType classifies the kind of spec change.
type ChangeType string

const (
	ChangeTypeRemovedPath     ChangeType = "removed-path"
	ChangeTypeRemovedMethod   ChangeType = "removed-method"
	ChangeTypeAddedPath       ChangeType = "added-path"
	ChangeTypeAddedMethod     ChangeType = "added-method"
	ChangeTypeNewRequired     ChangeType = "new-required-param"
	ChangeTypeTypeChange      ChangeType = "type-change"
	ChangeTypeRemovedField    ChangeType = "removed-response-field"
	ChangeTypeAddedField      ChangeType = "added-response-field"
	ChangeTypeStatusChange    ChangeType = "status-code-change"
	ChangeTypeEnumNarrowed    ChangeType = "enum-narrowed"
	ChangeTypeFormatChange    ChangeType = "format-change"
	ChangeTypeDescriptionOnly ChangeType = "description-change"
)

// SpecChange describes a single change between base and head specs.
type SpecChange struct {
	Path       string     `json:"path"`
	Method     string     `json:"method,omitempty"`
	Field      string     `json:"field,omitempty"`
	ChangeType ChangeType `json:"changeType"`
	Before     string     `json:"before,omitempty"`
	After      string     `json:"after,omitempty"`
	// Severity: "breaking" | "non-breaking" | "informational"
	Severity string `json:"severity"`
}

// SpecDiffReport is the result of comparing two specs.
type SpecDiffReport struct {
	BreakingChanges    []SpecChange `json:"breakingChanges"`
	NonBreakingChanges []SpecChange `json:"nonBreakingChanges"`
	AddedPaths         []string     `json:"addedPaths"`
	RemovedPaths       []string     `json:"removedPaths"`
}

// HasBreakingChanges returns true when any breaking change is present.
func (r *SpecDiffReport) HasBreakingChanges() bool {
	return len(r.BreakingChanges) > 0
}

// Summary returns a one-line human-readable summary.
func (r *SpecDiffReport) Summary() string {
	return fmt.Sprintf("breaking=%d non-breaking=%d added-paths=%d removed-paths=%d",
		len(r.BreakingChanges), len(r.NonBreakingChanges),
		len(r.AddedPaths), len(r.RemovedPaths))
}

// Diff compares base (old) spec against head (new) spec and returns a SpecDiffReport.
// Pass nil for either to treat it as an empty spec.
func Diff(base, head *openapi3.T) *SpecDiffReport {
	report := &SpecDiffReport{}

	basePaths := pathsOf(base)
	headPaths := pathsOf(head)

	// Removed paths
	for path := range basePaths {
		if _, ok := headPaths[path]; !ok {
			report.RemovedPaths = append(report.RemovedPaths, path)
			report.BreakingChanges = append(report.BreakingChanges, SpecChange{
				Path:       path,
				ChangeType: ChangeTypeRemovedPath,
				Before:     path,
				Severity:   "breaking",
			})
		}
	}

	// Added paths
	for path := range headPaths {
		if _, ok := basePaths[path]; !ok {
			report.AddedPaths = append(report.AddedPaths, path)
			report.NonBreakingChanges = append(report.NonBreakingChanges, SpecChange{
				Path:       path,
				ChangeType: ChangeTypeAddedPath,
				After:      path,
				Severity:   "non-breaking",
			})
		}
	}

	// Compare shared paths
	for path, baseItem := range basePaths {
		headItem, ok := headPaths[path]
		if !ok {
			continue
		}
		diffPathItem(report, path, baseItem, headItem)
	}

	sort.Slice(report.BreakingChanges, func(i, j int) bool {
		return report.BreakingChanges[i].Path < report.BreakingChanges[j].Path
	})
	sort.Slice(report.NonBreakingChanges, func(i, j int) bool {
		return report.NonBreakingChanges[i].Path < report.NonBreakingChanges[j].Path
	})
	sort.Strings(report.AddedPaths)
	sort.Strings(report.RemovedPaths)

	return report
}

// diffPathItem compares methods on a single path.
func diffPathItem(report *SpecDiffReport, path string, base, head *openapi3.PathItem) {
	methods := []struct {
		name string
		b    *openapi3.Operation
		h    *openapi3.Operation
	}{
		{"GET", base.Get, head.Get},
		{"POST", base.Post, head.Post},
		{"PUT", base.Put, head.Put},
		{"DELETE", base.Delete, head.Delete},
		{"PATCH", base.Patch, head.Patch},
		{"OPTIONS", base.Options, head.Options},
		{"HEAD", base.Head, head.Head},
	}

	for _, m := range methods {
		if m.b == nil && m.h != nil {
			// Method added — non-breaking
			report.NonBreakingChanges = append(report.NonBreakingChanges, SpecChange{
				Path: path, Method: m.name,
				ChangeType: ChangeTypeAddedMethod, Severity: "non-breaking",
			})
			continue
		}
		if m.b != nil && m.h == nil {
			// Method removed — breaking
			report.BreakingChanges = append(report.BreakingChanges, SpecChange{
				Path: path, Method: m.name,
				ChangeType: ChangeTypeRemovedMethod, Severity: "breaking",
			})
			continue
		}
		if m.b != nil && m.h != nil {
			diffOperation(report, path, m.name, m.b, m.h)
		}
	}
}

// diffOperation compares a single operation between base and head.
func diffOperation(report *SpecDiffReport, path, method string, base, head *openapi3.Operation) {
	// Check required request parameters added or promoted to required
	baseParams := paramMap(base.Parameters)
	headParams := paramMap(head.Parameters)
	for pName, pRef := range headParams {
		if pRef.Value == nil {
			continue
		}
		if pRef.Value.Required {
			if baseRef, existed := baseParams[pName]; !existed {
				// Brand new required param — breaking
				report.BreakingChanges = append(report.BreakingChanges, SpecChange{
					Path: path, Method: method, Field: pName,
					ChangeType: ChangeTypeNewRequired,
					After:      fmt.Sprintf("required param %q added", pName),
					Severity:   "breaking",
				})
			} else if baseRef != nil && baseRef.Value != nil && !baseRef.Value.Required {
				// Existing optional param promoted to required — breaking
				report.BreakingChanges = append(report.BreakingChanges, SpecChange{
					Path: path, Method: method, Field: pName,
					ChangeType: ChangeTypeNewRequired,
					Before:     fmt.Sprintf("param %q was optional", pName),
					After:      fmt.Sprintf("param %q is now required", pName),
					Severity:   "breaking",
				})
			}
		}
	}

	// Check response status codes changed (success codes only: 2xx)
	baseStatuses := successStatuses(base)
	headStatuses := successStatuses(head)
	for status := range baseStatuses {
		if _, ok := headStatuses[status]; !ok {
			report.BreakingChanges = append(report.BreakingChanges, SpecChange{
				Path: path, Method: method,
				ChangeType: ChangeTypeStatusChange,
				Before:     status,
				Severity:   "breaking",
			})
		}
	}
	for status := range headStatuses {
		if _, ok := baseStatuses[status]; !ok {
			report.NonBreakingChanges = append(report.NonBreakingChanges, SpecChange{
				Path: path, Method: method,
				ChangeType: ChangeTypeStatusChange,
				After:      status,
				Severity:   "non-breaking",
			})
		}
	}

	// Compare response schemas for shared success status codes
	for status, baseRespRef := range base.Responses {
		if !strings.HasPrefix(status, "2") {
			continue
		}
		headRespRef, ok := head.Responses[status]
		if !ok || baseRespRef == nil || headRespRef == nil {
			continue
		}
		baseSchema := extractResponseSchema(baseRespRef)
		headSchema := extractResponseSchema(headRespRef)
		if baseSchema != nil && headSchema != nil {
			diffSchema(report, path, method, status, "", baseSchema, headSchema)
		}
	}

	// Compare request body schema
	if base.RequestBody != nil && head.RequestBody != nil {
		baseReqSchema := extractRequestBodySchema(base.RequestBody)
		headReqSchema := extractRequestBodySchema(head.RequestBody)
		if baseReqSchema != nil && headReqSchema != nil {
			diffSchema(report, path, method, "request", "", baseReqSchema, headReqSchema)
		}
	}
}

// diffSchema recursively compares two schemas and appends findings to the report.
func diffSchema(report *SpecDiffReport, path, method, status, fieldPrefix string, base, head *openapi3.Schema) {
	baseType := schemaType(base)
	headType := schemaType(head)

	if baseType != headType && baseType != "" && headType != "" {
		field := fieldPrefix
		if field == "" {
			field = "(root)"
		}
		report.BreakingChanges = append(report.BreakingChanges, SpecChange{
			Path: path, Method: method, Field: field,
			ChangeType: ChangeTypeTypeChange,
			Before:     baseType, After: headType,
			Severity: "breaking",
		})
		return
	}

	// Format change
	if base.Format != head.Format && base.Format != "" {
		field := fieldPrefix
		if field == "" {
			field = "(root)"
		}
		report.BreakingChanges = append(report.BreakingChanges, SpecChange{
			Path: path, Method: method, Field: field,
			ChangeType: ChangeTypeFormatChange,
			Before:     base.Format, After: head.Format,
			Severity: "breaking",
		})
	}

	// Enum narrowing
	if len(base.Enum) > 0 && len(head.Enum) > 0 {
		headEnumSet := make(map[string]bool)
		for _, v := range head.Enum {
			headEnumSet[fmt.Sprintf("%v", v)] = true
		}
		for _, v := range base.Enum {
			if !headEnumSet[fmt.Sprintf("%v", v)] {
				field := fieldPrefix
				if field == "" {
					field = "(root)"
				}
				report.BreakingChanges = append(report.BreakingChanges, SpecChange{
					Path: path, Method: method, Field: field,
					ChangeType: ChangeTypeEnumNarrowed,
					Before:     fmt.Sprintf("%v", v),
					Severity:   "breaking",
				})
			}
		}
	}

	// Response field removal / addition in object properties
	if status != "request" {
		// Removed fields from response — breaking (clients may depend on them)
		for propName, baseProp := range base.Properties {
			if baseProp == nil || baseProp.Value == nil {
				continue
			}
			fieldName := join(fieldPrefix, propName)
			if _, ok := head.Properties[propName]; !ok {
				report.BreakingChanges = append(report.BreakingChanges, SpecChange{
					Path: path, Method: method, Field: fieldName,
					ChangeType: ChangeTypeRemovedField,
					Before:     propName,
					Severity:   "breaking",
				})
			} else if head.Properties[propName] != nil && head.Properties[propName].Value != nil {
				diffSchema(report, path, method, status, fieldName,
					baseProp.Value, head.Properties[propName].Value)
			}
		}
		// Added fields — non-breaking
		for propName := range head.Properties {
			if _, ok := base.Properties[propName]; !ok {
				fieldName := join(fieldPrefix, propName)
				report.NonBreakingChanges = append(report.NonBreakingChanges, SpecChange{
					Path: path, Method: method, Field: fieldName,
					ChangeType: ChangeTypeAddedField,
					After:      propName,
					Severity:   "non-breaking",
				})
			}
		}
	}
}

// --- helpers ---

func pathsOf(doc *openapi3.T) map[string]*openapi3.PathItem {
	if doc == nil || doc.Paths == nil {
		return map[string]*openapi3.PathItem{}
	}
	result := make(map[string]*openapi3.PathItem, len(doc.Paths))
	for k, v := range doc.Paths {
		result[k] = v
	}
	return result
}

func paramMap(params openapi3.Parameters) map[string]*openapi3.ParameterRef {
	m := make(map[string]*openapi3.ParameterRef, len(params))
	for _, p := range params {
		if p != nil && p.Value != nil {
			m[p.Value.Name] = p
		}
	}
	return m
}

func successStatuses(op *openapi3.Operation) map[string]bool {
	m := make(map[string]bool)
	for status := range op.Responses {
		if strings.HasPrefix(status, "2") {
			m[status] = true
		}
	}
	return m
}

func extractResponseSchema(ref *openapi3.ResponseRef) *openapi3.Schema {
	if ref == nil || ref.Value == nil {
		return nil
	}
	for _, contentRef := range ref.Value.Content {
		if contentRef != nil && contentRef.Schema != nil && contentRef.Schema.Value != nil {
			return contentRef.Schema.Value
		}
	}
	return nil
}

func extractRequestBodySchema(ref *openapi3.RequestBodyRef) *openapi3.Schema {
	if ref == nil || ref.Value == nil {
		return nil
	}
	for _, contentRef := range ref.Value.Content {
		if contentRef != nil && contentRef.Schema != nil && contentRef.Schema.Value != nil {
			return contentRef.Schema.Value
		}
	}
	return nil
}

func schemaType(s *openapi3.Schema) string {
	if s == nil {
		return ""
	}
	return s.Type
}

func join(prefix, name string) string {
	if prefix == "" {
		return name
	}
	return prefix + "." + name
}
