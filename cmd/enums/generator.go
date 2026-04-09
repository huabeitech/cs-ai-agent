package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"golang.org/x/tools/go/packages"
)

const (
	enumsPkgName = "enums"
	enumsDir     = "internal/pkg/enums"
	outputPath   = "web/lib/generated/enums.ts"
)

type enumValueType string

const (
	enumValueTypeInt    enumValueType = "int"
	enumValueTypeString enumValueType = "string"
)

type constDef struct {
	Name      string
	ValueType enumValueType
	Value     any
}

type enumItem struct {
	Name  string
	Value any
	Label string
	Order int
}

type enumDef struct {
	Name      string
	ValueType enumValueType
	Items     []enumItem
}

func main() {
	defs, err := parseEnums(enumsDir)
	if err != nil {
		panic(err)
	}
	content := buildTSFile(defs)
	if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(outputPath, []byte(content), 0o644); err != nil {
		panic(err)
	}
}

func parseEnums(dir string) ([]enumDef, error) {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax,
		Dir:  ".",
	}, "./"+filepath.ToSlash(dir))
	if err != nil {
		return nil, err
	}
	if packages.PrintErrors(pkgs) > 0 {
		return nil, fmt.Errorf("failed to load package %s", dir)
	}

	var pkg *packages.Package
	for _, candidate := range pkgs {
		if candidate.Name == enumsPkgName {
			pkg = candidate
			break
		}
	}
	if pkg == nil {
		return nil, fmt.Errorf("package %s not found in %s", enumsPkgName, dir)
	}

	typeMap := make(map[string]enumValueType)
	constMap := make(map[string]constDef)
	orderMap := make(map[string]int)
	order := 0

	files := sortedPackageFiles(pkg)
	for _, file := range files {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok {
				continue
			}
			switch genDecl.Tok {
			case token.TYPE:
				readTypes(genDecl, typeMap)
			case token.CONST:
				readConsts(genDecl, typeMap, constMap, orderMap, &order)
			}
		}
	}

	var defs []enumDef
	for _, file := range files {
		for _, decl := range file.Decls {
			genDecl, ok := decl.(*ast.GenDecl)
			if !ok || genDecl.Tok != token.VAR {
				continue
			}
			items := parseLabelMaps(genDecl, typeMap, constMap, orderMap)
			defs = append(defs, items...)
		}
	}

	sort.Slice(defs, func(i, j int) bool {
		return defs[i].Name < defs[j].Name
	})
	return defs, nil
}

func sortedPackageFiles(pkg *packages.Package) []*ast.File {
	type fileWithPath struct {
		path string
		file *ast.File
	}

	pairs := make([]fileWithPath, 0, len(pkg.Syntax))
	for _, file := range pkg.Syntax {
		pairs = append(pairs, fileWithPath{
			path: pkg.Fset.Position(file.Package).Filename,
			file: file,
		})
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].path < pairs[j].path
	})

	results := make([]*ast.File, 0, len(pairs))
	for _, pair := range pairs {
		results = append(results, pair.file)
	}
	return results
}

func readTypes(genDecl *ast.GenDecl, typeMap map[string]enumValueType) {
	for _, spec := range genDecl.Specs {
		typeSpec, ok := spec.(*ast.TypeSpec)
		if !ok {
			continue
		}
		ident, ok := typeSpec.Type.(*ast.Ident)
		if !ok {
			continue
		}
		valueType, ok := parseValueType(ident.Name)
		if !ok {
			continue
		}
		typeMap[typeSpec.Name.Name] = valueType
	}
}

func readConsts(genDecl *ast.GenDecl, typeMap map[string]enumValueType, constMap map[string]constDef, orderMap map[string]int, order *int) {
	for _, spec := range genDecl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok {
			continue
		}

		explicitTypeName := ""
		explicitValueType := enumValueType("")
		if valueSpec.Type != nil {
			typeName := exprName(valueSpec.Type)
			if valueType, ok := lookupValueType(typeName, typeMap); ok {
				explicitTypeName = typeName
				explicitValueType = valueType
			}
		}

		for idx, name := range valueSpec.Names {
			valueExpr := expressionAt(valueSpec.Values, idx)
			if valueExpr == nil {
				continue
			}

			valueType := explicitValueType
			if valueType == "" {
				valueType = inferValueType(valueExpr)
			}
			if valueType == "" {
				continue
			}

			value, ok := parseLiteralValue(valueExpr, valueType)
			if !ok {
				continue
			}

			constMap[name.Name] = constDef{
				Name:      name.Name,
				ValueType: valueType,
				Value:     value,
			}
			orderMap[name.Name] = *order
			*order++
			_ = explicitTypeName
		}
	}
}

func parseLabelMaps(genDecl *ast.GenDecl, typeMap map[string]enumValueType, constMap map[string]constDef, orderMap map[string]int) []enumDef {
	var defs []enumDef
	for _, spec := range genDecl.Specs {
		valueSpec, ok := spec.(*ast.ValueSpec)
		if !ok || len(valueSpec.Values) != 1 {
			continue
		}

		compLit, ok := valueSpec.Values[0].(*ast.CompositeLit)
		if !ok {
			continue
		}

		mapType, ok := compLit.Type.(*ast.MapType)
		if !ok {
			continue
		}

		if exprName(mapType.Value) != "string" {
			continue
		}

		keyTypeName := exprName(mapType.Key)
		mapValueType, ok := lookupValueType(keyTypeName, typeMap)
		if !ok {
			continue
		}

		items := make([]enumItem, 0, len(compLit.Elts))
		constNames := make([]string, 0, len(compLit.Elts))
		for _, elt := range compLit.Elts {
			kv, ok := elt.(*ast.KeyValueExpr)
			if !ok {
				continue
			}

			constName := exprName(kv.Key)
			def, ok := constMap[constName]
			if !ok || def.ValueType != mapValueType {
				continue
			}

			label, ok := parseStringLiteral(kv.Value)
			if !ok {
				continue
			}

			items = append(items, enumItem{
				Value: def.Value,
				Label: label,
				Order: orderMap[constName],
			})
			constNames = append(constNames, constName)
		}

		if len(items) == 0 {
			continue
		}

		enumName := keyTypeName
		if keyTypeName == "int" || keyTypeName == "string" {
			enumName = exportedEnumName(valueSpec.Names[0].Name)
			if enumName == "" {
				enumName = commonIdentifierPrefix(constNames)
			}
		}
		if enumName == "" {
			continue
		}

		valid := true
		for idx, constName := range constNames {
			itemName := strings.TrimPrefix(constName, enumName)
			if itemName == "" {
				valid = false
				break
			}
			items[idx].Name = itemName
		}
		if !valid {
			continue
		}

		sort.Slice(items, func(i, j int) bool {
			return items[i].Order < items[j].Order
		})

		defs = append(defs, enumDef{
			Name:      enumName,
			ValueType: mapValueType,
			Items:     items,
		})
	}
	return defs
}

func expressionAt(values []ast.Expr, idx int) ast.Expr {
	if idx < len(values) {
		return values[idx]
	}
	return nil
}

func exprName(expr ast.Expr) string {
	switch v := expr.(type) {
	case *ast.Ident:
		return v.Name
	default:
		return ""
	}
}

func parseValueType(name string) (enumValueType, bool) {
	switch name {
	case "int":
		return enumValueTypeInt, true
	case "string":
		return enumValueTypeString, true
	default:
		return "", false
	}
}

func lookupValueType(name string, typeMap map[string]enumValueType) (enumValueType, bool) {
	if valueType, ok := parseValueType(name); ok {
		return valueType, true
	}
	valueType, ok := typeMap[name]
	return valueType, ok
}

func inferValueType(expr ast.Expr) enumValueType {
	switch value := expr.(type) {
	case *ast.BasicLit:
		switch value.Kind {
		case token.INT:
			return enumValueTypeInt
		case token.STRING:
			return enumValueTypeString
		}
	case *ast.UnaryExpr:
		if value.Op == token.SUB {
			return inferValueType(value.X)
		}
	}
	return ""
}

func parseLiteralValue(expr ast.Expr, valueType enumValueType) (any, bool) {
	switch value := expr.(type) {
	case *ast.BasicLit:
		switch valueType {
		case enumValueTypeInt:
			parsed, err := strconv.Atoi(value.Value)
			return parsed, err == nil
		case enumValueTypeString:
			parsed, err := strconv.Unquote(value.Value)
			return parsed, err == nil
		}
	case *ast.UnaryExpr:
		if valueType == enumValueTypeInt && value.Op == token.SUB {
			parsed, ok := parseLiteralValue(value.X, valueType)
			if !ok {
				return nil, false
			}
			return -parsed.(int), true
		}
	}
	return nil, false
}

func parseStringLiteral(expr ast.Expr) (string, bool) {
	basicLit, ok := expr.(*ast.BasicLit)
	if !ok || basicLit.Kind != token.STRING {
		return "", false
	}
	value, err := strconv.Unquote(basicLit.Value)
	return value, err == nil
}

func commonIdentifierPrefix(names []string) string {
	if len(names) == 0 {
		return ""
	}

	common := splitIdentifier(names[0])
	for _, name := range names[1:] {
		tokens := splitIdentifier(name)
		limit := min(len(common), len(tokens))
		idx := 0
		for idx < limit && common[idx] == tokens[idx] {
			idx++
		}
		common = common[:idx]
		if len(common) == 0 {
			return ""
		}
	}
	return strings.Join(common, "")
}

func exportedEnumName(varName string) string {
	base := strings.TrimSuffix(varName, "LabelMap")
	if base == "" {
		return ""
	}

	tokens := splitIdentifier(base)
	if len(tokens) == 0 {
		return ""
	}

	var result strings.Builder
	for _, token := range tokens {
		if token == "" {
			continue
		}
		if isAllLower(token) && len(token) <= 3 {
			result.WriteString(strings.ToUpper(token))
			continue
		}
		result.WriteString(strings.ToUpper(token[:1]))
		result.WriteString(token[1:])
	}
	return result.String()
}

func splitIdentifier(value string) []string {
	if value == "" {
		return nil
	}

	runes := []rune(value)
	parts := make([]string, 0, 4)
	start := 0
	for idx := 1; idx < len(runes); idx++ {
		prev := runes[idx-1]
		curr := runes[idx]

		if isBoundary(runes, idx, prev, curr) {
			parts = append(parts, string(runes[start:idx]))
			start = idx
		}
	}
	parts = append(parts, string(runes[start:]))
	return parts
}

func isBoundary(runes []rune, idx int, prev rune, curr rune) bool {
	if isLower(prev) && isUpper(curr) {
		return true
	}
	if isUpper(prev) && isUpper(curr) && idx+1 < len(runes) && isLower(runes[idx+1]) {
		return true
	}
	if isDigit(prev) != isDigit(curr) {
		return true
	}
	return false
}

func isLower(r rune) bool {
	return r >= 'a' && r <= 'z'
}

func isUpper(r rune) bool {
	return r >= 'A' && r <= 'Z'
}

func isDigit(r rune) bool {
	return r >= '0' && r <= '9'
}

func isAllLower(value string) bool {
	for _, r := range value {
		if !isLower(r) {
			return false
		}
	}
	return true
}

func buildTSFile(defs []enumDef) string {
	var buf bytes.Buffer
	buf.WriteString("// Code generated by cmd/enums/generator.go. DO NOT EDIT.\n\n")
	for _, def := range defs {
		buf.WriteString(renderEnum(def))
		buf.WriteString("\n")
		buf.WriteString(renderLabels(def))
		buf.WriteString("\n\n")
	}
	return strings.TrimRight(buf.String(), "\n") + "\n"
}

func renderEnum(def enumDef) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("export enum %s {", def.Name))
	for _, item := range def.Items {
		lines = append(lines, fmt.Sprintf("  %s = %s,", item.Name, formatTSValue(item.Value, def.ValueType)))
	}
	lines = append(lines, "}")
	return strings.Join(lines, "\n")
}

func renderLabels(def enumDef) string {
	var lines []string
	lines = append(lines, fmt.Sprintf("export const %sLabels: Record<%s, string> = {", def.Name, def.Name))
	for _, item := range def.Items {
		lines = append(lines, fmt.Sprintf("  [%s.%s]: %s,", def.Name, item.Name, strconv.Quote(item.Label)))
	}
	lines = append(lines, "}")
	return strings.Join(lines, "\n")
}

func formatTSValue(value any, valueType enumValueType) string {
	switch valueType {
	case enumValueTypeString:
		return strconv.Quote(value.(string))
	case enumValueTypeInt:
		return fmt.Sprint(value.(int))
	default:
		panic(fmt.Sprintf("unsupported enum value type: %s", valueType))
	}
}

func min(a int, b int) int {
	if a < b {
		return a
	}
	return b
}
