package gen

import (
	"parser"
	"strings"
)

type PyCompiler struct {
	currentClass    string
	currentFunction string
	header          string
	source          string
}

func (c *PyCompiler) ResolveType(node *parser.Node) string {
	if node.Name != "Type" {
		panic(node)
	}
	return node.Data()
}

var pyprimitives = map[string]string{
	"ge":             " >= ",
	"le":             " <= ",
	"eq":             " == ",
	"ne":             " != ",
	"lt":             " < ",
	"gt":             " > ",
	"and":            " and ",
	"or":             " or ",
	"plus":           " + ",
	"minus":          " - ",
	"not":            "not ",
	"MemberAccess":   ".",
	"BreakStatement": "break",
}

func (c *PyCompiler) Recurse(node *parser.Node) string {
	ret := ""
	var cf parser.CodeFormatter
	p := pyprimitives[node.Name]
	if p != "" {
		return p
	}

	switch node.Name {
	case "Identifier", "Float", "Integer":
		return node.Data()
	case "Boolean":
		if node.Data() == "true" {
			return "True"
		} else {
			return "False"
		}
	case "Text":
		return "\"" + node.Data() + "\""
	case "Comparison":
		a := c.Recurse(node.Children[0])
		cmp := c.Recurse(node.Children[1])
		b := c.Recurse(node.Children[2])
		return a + cmp + b
	case "While":
		return "while " + c.Recurse(node.Children[0]) + ":\n" + c.Recurse(node.Children[1])
	case "If":
		return "if " + c.Recurse(node.Children[0]) + ":\n" + c.Recurse(node.Children[1])
	case "ElseIf":
		return "el" + c.Recurse(node.Children[0])
	case "Else":
		return "else:" + c.Recurse(node.Children[0])
	case "For":
		variable := c.Recurse(node.Children[0])
		array := c.Recurse(node.Children[1])
		block := c.Recurse(node.Children[2])

		return "for " + variable + " in " + array + ":\n" + block
	case "NewStatement":
		return c.Recurse(node.Children[0]) + "()"
	case "ReturnStatement":
		return "return " + c.Recurse(node.Children[0])
	case "ArrayIndexing":
		return c.Recurse(node.Children[0]) + "[" + c.Recurse(node.Children[1]) + "]"
	case "ArraySlicing":
		id := c.Recurse(node.Children[0])
		a := ""
		b := ""
		if node.Children[1].Name != "colon" {
			a = c.Recurse(node.Children[1])
		}
		back := node.Children[len(node.Children)-1]
		if back.Name != "colon" {
			b = c.Recurse(back)
		}
		return id + "[" + a + ":" + b + "]"
	case "PostInc":
		return c.Recurse(node.Children[0]) + " += 1"
	case "PostDec":
		return c.Recurse(node.Children[0]) + " -= 1"
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
		cf.Add("class " + cn + ":\n")

		cf.Inc()
		for _, child := range node.Children[1:] {
			ret := c.Recurse(child)
			if child.Name == "VariableDeclaration" {
				ret += "\n"
			}
			cf.Add(ret)
		}
		cf.Dec()
		cf.Add("\n")
		c.currentClass = ""
		return cf.String()
	case "Assignment":
		return node.Children[0].Data() + " = " + c.Recurse(node.Children[1])
	case "PlusEquals":
		return node.Children[0].Data() + " += " + c.Recurse(node.Children[1])
	case "VariableDeclaration":
		ret := "" //c.ResolveType(node.Children[0]) + " "
		n1 := node.Children[1]
		if n1.Name == "Assignment" {
			ret += c.Recurse(n1)
		} else {
			ret += n1.Data()
		}
		return ret
	case "Block":
		cf.Inc()
		cf.Add("\t")
		for _, child := range node.Children {
			r := c.Recurse(child)
			if !strings.HasSuffix(r, "\n") {
				r += "\n"
			}
			cf.Add(r)
		}
		cf.Dec()
		return cf.String()
	case "Destructor":
		c.currentFunction = node.Children[0].Data()
		ret := "def __del__(self):\n"
		c.currentFunction = ""
		return ret + "" + c.Recurse(node.Children[1])
	case "FunctionDeclaration":
		c.currentFunction = node.Children[1].Data()
		ret := "def " + c.currentFunction + "("
		args := ""
		if c.currentClass != "" {
			args += "self"
		}
		for _, child := range node.Children[2 : len(node.Children)-1] {
			if args != "" {
				args += ", "
			}
			args += c.Recurse(child)
		}
		ret += args + "):\n"
		ret += c.Recurse(node.Children[len(node.Children)-1])
		c.currentFunction = ""
		return ret
	}
	for _, child := range node.Children {
		ret += c.Recurse(child)
	}
	return ret
}
