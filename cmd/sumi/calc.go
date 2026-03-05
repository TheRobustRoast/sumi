package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"math"
	"os/exec"
	"strconv"

	"github.com/charmbracelet/huh"
	"github.com/spf13/cobra"

	"sumi/internal/theme"
)

func calcCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "calc [expression]",
		Short: "Quick calculator (result copied to clipboard)",
		Long:  "Evaluate a math expression and copy the result to clipboard.\nSupports: +, -, *, /, %, parentheses, sqrt, sin, cos, tan, log, pow, pi, e.",
		Example: `  sumi calc "2+3*4"
  sumi calc "sqrt(144)"
  sumi calc "pow(2,10)"`,
		RunE: func(cmd *cobra.Command, args []string) error {
			var expr string
			if len(args) > 0 {
				expr = args[0]
			} else {
				form := huh.NewForm(
					huh.NewGroup(
						huh.NewInput().
							Title("calc").
							Description("math expression (e.g. 2+3*4, sqrt(144))").
							Value(&expr),
					),
				)
				if err := form.Run(); err != nil {
					return nil
				}
			}
			if expr == "" {
				return nil
			}

			result, err := safeEval(expr)
			if err != nil {
				return fmt.Errorf("calc: %w", err)
			}

			// Format nicely — drop trailing zeros
			text := strconv.FormatFloat(result, 'f', -1, 64)

			// Copy to clipboard
			c := exec.Command("wl-copy", text)
			c.Run() //nolint:errcheck

			fmt.Println(theme.Ok(fmt.Sprintf("%s = %s (copied)", expr, text)))

			exec.Command("notify-send", "-a", "sumi", "-t", "2000",
				"Calculator", fmt.Sprintf("%s = %s (copied)", expr, text)).Run() //nolint:errcheck

			return nil
		},
	}
}

// safeEval evaluates a math expression safely using Go's AST parser.
// Supports: +, -, *, /, parentheses, and math functions (sqrt, sin, cos, etc.)
func safeEval(expr string) (float64, error) {
	// Wrap in a fake Go expression context
	node, err := parser.ParseExpr(expr)
	if err != nil {
		return 0, fmt.Errorf("invalid expression: %s", expr)
	}
	return evalNode(node)
}

func evalNode(node ast.Expr) (float64, error) {
	switch n := node.(type) {
	case *ast.BasicLit:
		switch n.Kind {
		case token.INT, token.FLOAT:
			return strconv.ParseFloat(n.Value, 64)
		}
		return 0, fmt.Errorf("unsupported literal: %s", n.Value)

	case *ast.ParenExpr:
		return evalNode(n.X)

	case *ast.UnaryExpr:
		x, err := evalNode(n.X)
		if err != nil {
			return 0, err
		}
		switch n.Op {
		case token.SUB:
			return -x, nil
		case token.ADD:
			return x, nil
		}
		return 0, fmt.Errorf("unsupported unary op: %s", n.Op)

	case *ast.BinaryExpr:
		left, err := evalNode(n.X)
		if err != nil {
			return 0, err
		}
		right, err := evalNode(n.Y)
		if err != nil {
			return 0, err
		}
		switch n.Op {
		case token.ADD:
			return left + right, nil
		case token.SUB:
			return left - right, nil
		case token.MUL:
			return left * right, nil
		case token.QUO:
			if right == 0 {
				return 0, fmt.Errorf("division by zero")
			}
			return left / right, nil
		case token.REM:
			if right == 0 {
				return 0, fmt.Errorf("modulo by zero")
			}
			return math.Mod(left, right), nil
		}
		return 0, fmt.Errorf("unsupported binary op: %s", n.Op)

	case *ast.CallExpr:
		ident, ok := n.Fun.(*ast.Ident)
		if !ok {
			return 0, fmt.Errorf("unsupported function call")
		}
		var args []float64
		for _, arg := range n.Args {
			v, err := evalNode(arg)
			if err != nil {
				return 0, err
			}
			args = append(args, v)
		}
		return evalFunc(ident.Name, args)

	case *ast.Ident:
		switch n.Name {
		case "pi", "PI":
			return math.Pi, nil
		case "e", "E":
			return math.E, nil
		}
		return 0, fmt.Errorf("unknown variable: %s", n.Name)
	}

	return 0, fmt.Errorf("unsupported expression type")
}

func evalFunc(name string, args []float64) (float64, error) {
	if len(args) < 1 {
		return 0, fmt.Errorf("%s: need at least 1 argument", name)
	}
	switch name {
	case "sqrt":
		return math.Sqrt(args[0]), nil
	case "abs":
		return math.Abs(args[0]), nil
	case "sin":
		return math.Sin(args[0]), nil
	case "cos":
		return math.Cos(args[0]), nil
	case "tan":
		return math.Tan(args[0]), nil
	case "log", "ln":
		return math.Log(args[0]), nil
	case "log2":
		return math.Log2(args[0]), nil
	case "log10":
		return math.Log10(args[0]), nil
	case "ceil":
		return math.Ceil(args[0]), nil
	case "floor":
		return math.Floor(args[0]), nil
	case "round":
		return math.Round(args[0]), nil
	case "pow":
		if len(args) < 2 {
			return 0, fmt.Errorf("pow needs 2 arguments")
		}
		return math.Pow(args[0], args[1]), nil
	case "max":
		if len(args) < 2 {
			return 0, fmt.Errorf("max needs 2 arguments")
		}
		return math.Max(args[0], args[1]), nil
	case "min":
		if len(args) < 2 {
			return 0, fmt.Errorf("min needs 2 arguments")
		}
		return math.Min(args[0], args[1]), nil
	}
	return 0, fmt.Errorf("unknown function: %s", name)
}
