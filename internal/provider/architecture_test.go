// Copyright (c) Josh Archer
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestArchitecture_ResourcesUseAPIClientOnly enforces the Single Egress Rule:
// Terraform resources must only communicate through the configured Seerr APIClient
// (r.client.Request) and must NEVER make out-of-band direct HTTP calls to downstream
// services or external hosts.
func TestArchitecture_ResourcesUseAPIClientOnly(t *testing.T) {
	files, err := filepath.Glob("resource_*.go")
	if err != nil {
		t.Fatalf("failed to glob resource files: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no resource_*.go files found")
	}

	fset := token.NewFileSet()

	forbiddenSelectors := map[string][]string{
		"http": {
			"Get",
			"Post",
			"Head",
			"PostForm",
			"NewRequest",
			"NewRequestWithContext",
			"DefaultClient",
			"DefaultTransport",
		},
	}

	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}

		src, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("failed to read file %s: %v", file, err)
		}

		node, err := parser.ParseFile(fset, file, src, parser.AllErrors)
		if err != nil {
			t.Fatalf("failed to parse file %s: %v", file, err)
		}

		ast.Inspect(node, func(n ast.Node) bool {
			// Check for forbidden selector expressions like http.Get, http.NewRequestWithContext, etc.
			sel, ok := n.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			ident, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}

			if forbiddenMethods, found := forbiddenSelectors[ident.Name]; found {
				for _, forbidden := range forbiddenMethods {
					if sel.Sel.Name == forbidden {
						pos := fset.Position(sel.Pos())
						t.Errorf(
							"%s:%d: forbidden direct HTTP call '%s.%s' in resource implementation.\n"+
								"Architectural Rule: Resources must only make HTTP requests via r.client.Request() "+
								"to the configured Seerr base URL.",
							file, pos.Line, ident.Name, sel.Sel.Name,
						)
					}
				}
			}

			return true
		})
	}
}
