package main

import (
	"bytes"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"text/template"
)

// binaryGeneratedMkTemplate is the template for the file binary.generated.mk, use binaryGeneratedMkTemplateOptions as the context.
const binaryGeneratedMkTemplate = `# Generated by ./cmd/generate-makefile. DO NOT EDIT.

{{ .Content -}}

.PHONY: clean-binary
clean-binary:
{{- range .Recipes }}
	rm -f {{ .OutputPath }}
{{- end }}
`

// binaryRecipeTemplate is the template for one target, use binaryRecipeOptions as the context.
const binaryRecipeTemplate = `.PHONY: {{ .TargetName }}
{{ .TargetName }}: SHELL:=$(RUN_IN_BUILD_SHELL)
{{ .TargetName }}: image-build-env {{ StringsJoin .DependencyTargets " " }} ## {{ .Comment }}
{{- if .UseCGO }}
	$(CGO) build -ldflags "$(LDFLAGS)" -tags "${BUILD_TAGS}" -o {{ .OutputPath }} {{ .SourcePath }}
{{- else }}
	$(GO) build -ldflags "$(LDFLAGS)" -tags "${BUILD_TAGS}" -o {{ .OutputPath }} {{ .SourcePath }}
{{- end }}

`

type binaryRecipeOptions struct {
	// TargetName is the name of the makefile target, also it is the output path of the binary.
	TargetName string
	// SourcePath is the path to the source file.
	SourcePath string
	// OutputPath is the path to the output file.
	OutputPath string
	// UseCGO introduces a CGO_ENABLED=1 environment variable to the build command.
	UseCGO bool
	// DependencyTargets are the targets that this target depends on.
	DependencyTargets []string
	// Comment is the comment for the target, do not need to include the leading `##`
	Comment string
}

type binaryGeneratedMkOptions struct {
	Recipes []binaryRecipeOptions
	Content string
}

func renderBinaryGeneratedMk() error {
	targetFile, err := os.OpenFile("binary.generated.mk", os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return errors.Wrap(err, "open file binary.generated.mk")
	}
	recipeTemplate, err := template.New("binary.generated.mk recipe").Funcs(defaultFuncMap).Parse(binaryRecipeTemplate)
	if err != nil {
		return errors.Wrap(err, "parse binary.generated.mk recipe template")
	}

	var buffer bytes.Buffer
	for _, recipe := range recipes {
		err := recipeTemplate.Execute(&buffer, recipe)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("render recipe in binary.generated.mk, recipe: %s", recipe.TargetName))
		}
	}
	binaryTemplate, err := template.New("binary.generated.mk").Parse(binaryGeneratedMkTemplate)
	if err != nil {
		return errors.Wrap(err, "parse binary.generated.mk template")
	}
	err = binaryTemplate.Execute(targetFile, binaryGeneratedMkOptions{
		Recipes: recipes,
		Content: buffer.String(),
	})
	if err != nil {
		return errors.Wrap(err, "render binary.generated.mk")
	}
	return nil
}

// recipes is the list of recipes to generate, edit here to build new binaries.
var recipes = []binaryRecipeOptions{
	{
		TargetName:        "chaos-controller-manager",
		SourcePath:        "cmd/chaos-controller-manager/main.go",
		OutputPath:        "images/chaos-mesh/bin/chaos-controller-manager",
		UseCGO:            false,
		DependencyTargets: nil,
		Comment:           "Build binary chaos-controller-manager",
	}, {
		TargetName: "chaos-daemon",
		SourcePath: "cmd/chaos-daemon/main.go",
		OutputPath: "images/chaos-daemon/bin/chaos-daemon",
		UseCGO:     true,
		DependencyTargets: []string{
			"pkg/time/fakeclock/fake_clock_gettime.o",
			"pkg/time/fakeclock/fake_gettimeofday.o",
		},
		Comment: "Build binary chaos-daemon",
	}, {
		TargetName: "chaos-dashboard",
		SourcePath: "cmd/chaos-dashboard/main.go",
		OutputPath: "images/chaos-dashboard/bin/chaos-dashboard",
		UseCGO:     true,
		DependencyTargets: []string{
			"ui",
		},
		Comment: "Build binary chaos-dashboard",
	}, {
		TargetName:        "cdh",
		SourcePath:        "cmd/chaos-daemon-helper/main.go",
		OutputPath:        "images/chaos-daemon/bin/chaos-daemon-helper",
		UseCGO:            true,
		DependencyTargets: nil,
		Comment:           "Build binary chaos-daemon-helper",
	},
}
