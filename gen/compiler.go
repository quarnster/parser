package gen

import (
	"parser"
	"strings"
)

type Compiler struct {
	currentClass    string
	currentFunction string
	header          string
	source          string
}

func (c *Compiler) ResolveType(node *parser.Node) string {
	if node.Name != "Type" {
		panic(node)
	}
	return node.Data()
}

var primitives = map[string]string{
	"ge":             " >= ",
	"le":             " <= ",
	"eq":             " == ",
	"ne":             " != ",
	"lt":             " < ",
	"gt":             " > ",
	"and":            " && ",
	"or":             " || ",
	"plus":           " + ",
	"PlusEquals":     " += ",
	"minus":          " - ",
	"not":            "!",
	"MemberAccess":   ".",
	"BreakStatement": "break",
}

func (c *Compiler) Recurse(node *parser.Node) string {
	ret := ""
	var cf parser.CodeFormatter
	p := primitives[node.Name]
	if p != "" {
		return p
	}

	switch node.Name {
	case "Identifier", "Boolean", "Float", "Integer":
		return node.Data()
	case "Text":
		return "\"" + node.Data() + "\""
	case "Comparison":
		a := c.Recurse(node.Children[0])
		cmp := c.Recurse(node.Children[1])
		b := c.Recurse(node.Children[2])
		return a + cmp + b
	case "While":
		return "while (" + c.Recurse(node.Children[0]) + ")\n" + c.Recurse(node.Children[1])
	case "If":
		return "if (" + c.Recurse(node.Children[0]) + ")\n" + c.Recurse(node.Children[1])
	case "ElseIf":
		ret += "else "
	case "Else":
		ret += "else\n"
	case "ReturnStatement":
		return "return " + c.Recurse(node.Children[0])
	case "ArrayIndexing":
		return c.Recurse(node.Children[0]) + "[" + c.Recurse(node.Children[1]) + "]"
	case "PostInc":
		return c.Recurse(node.Children[0]) + "++"
	case "PostDec":
		return c.Recurse(node.Children[0]) + "--"
	case "FunctionCall":
		ret = node.Children[0].Data()
		args := ""
		for _, child := range node.Children[1:] {
			if args != "" {
				args += ", "
			}
			args += c.Recurse(child)
		}

		return ret + "(" + args + ")"
	case "Class":
		cn := node.Children[0].Data()
		c.currentClass = cn
		cf.Add("class " + cn + "\n{\npublic:\n")
		cf.Inc()
		for _, child := range node.Children {
			ret := c.Recurse(child)
			if child.Name == "VariableDeclaration" {
				ret += ";\n"
			}
			cf.Add(ret)
		}
		cf.Dec()
		cf.Add("};\n")
		c.currentClass = ""
		return cf.String()
	case "Assignment":
		return node.Children[0].Data() + " = " + c.Recurse(node.Children[1])
	case "VariableDeclaration":
		ret := c.ResolveType(node.Children[0]) + " "
		n1 := node.Children[1]
		if n1.Name == "Assignment" {
			ret += c.Recurse(n1)
		} else {
			ret += n1.Data()
		}
		return ret
	case "Block":
		cf.Add("{\n")
		cf.Inc()
		for _, child := range node.Children {
			r := c.Recurse(child)
			if !strings.HasSuffix(r, "}\n") {
				r += ";\n"
			}
			cf.Add(r)
		}
		cf.Dec()
		cf.Add("}\n")
		return cf.String()
	case "Destructor":
		c.currentFunction = node.Children[0].Data()
		ret := "~" + c.currentFunction
		c.currentFunction = ""
		return ret + "() {}"
	case "FunctionDeclaration":
		c.currentFunction = node.Children[1].Data()
		ret := c.ResolveType(node.Children[0]) + " " + c.currentFunction + "("
		args := ""
		for _, child := range node.Children[2 : len(node.Children)-1] {
			if args != "" {
				args += ", "
			}
			args += c.Recurse(child)
		}
		ret += args + ")\n"
		ret += c.Recurse(node.Children[len(node.Children)-1])
		c.currentFunction = ""
		return ret
	}
	for _, child := range node.Children {
		ret += c.Recurse(child)
	}
	return ret
}
