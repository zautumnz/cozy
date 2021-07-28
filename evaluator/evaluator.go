// Package evaluator contains the core of our interpreter, which walks
// the AST produced by the parser and evaluates the user-submitted program.
package evaluator

import (
	"context"
	"fmt"
	"io/ioutil"
	"math"
	"os"
	"regexp"
	"strings"

	"github.com/zacanger/cozy/ast"
	"github.com/zacanger/cozy/lexer"
	"github.com/zacanger/cozy/object"
	"github.com/zacanger/cozy/parser"
	"github.com/zacanger/cozy/token"
)

// pre-defined objects
var (
	TRUE  = &object.Boolean{Value: true}
	FALSE = &object.Boolean{Value: false}
	CTX   = context.Background()
)

// The built-in functions / standard-library methods are stored here.
var builtins = map[string]*object.Builtin{}

// Eval is our core function for evaluating nodes.
func Eval(node ast.Node, env *object.Environment) object.Object {
	return EvalContext(context.Background(), node, env)
}

// EvalContext is our core function for evaluating nodes.
// The context.Context provided can be used to cancel a running script instance.
func EvalContext(ctx context.Context, node ast.Node, env *object.Environment) object.Object {

	// We test our context at every iteration of our main-loop.
	select {
	case <-ctx.Done():
		return &object.Error{Message: ctx.Err().Error()}
	default:
		// noop
	}

	switch node := node.(type) {

	//Statements
	case *ast.Program:
		return evalProgram(node, env)
	case *ast.ExpressionStatement:
		return Eval(node.Expression, env)

	//Expressions
	case *ast.IntegerLiteral:
		return &object.Integer{Value: node.Value}
	case *ast.FloatLiteral:
		return &object.Float{Value: node.Value}
	case *ast.Boolean:
		return nativeBoolToBooleanObject(node.Value)
	case *ast.PrefixExpression:
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		return evalPrefixExpression(node.Operator, right)
	case *ast.PostfixExpression:
		return evalPostfixExpression(env, node.Operator, node)
	case *ast.InfixExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		right := Eval(node.Right, env)
		if isError(right) {
			return right
		}
		res := evalInfixExpression(node.Operator, left, right, env)
		if isError(res) {
			fmt.Printf("Error: %s\n", res.Inspect())
			os.Exit(1)
		}
		return res

	case *ast.BlockStatement:
		return evalBlockStatement(node, env)
	case *ast.IfExpression:
		return evalIfExpression(node, env)
	case *ast.WhileExpression:
		return evalWhileExpression(node, env)
	case *ast.TernaryExpression:
		return evalTernaryExpression(node, env)
	case *ast.ImportExpression:
		return evalImportExpression(node, env)
	case *ast.ForLoopExpression:
		return evalForLoopExpression(node, env)
	case *ast.ForeachStatement:
		return evalForeachExpression(node, env)
	case *ast.ReturnStatement:
		val := Eval(node.ReturnValue, env)
		if isError(val) {
			return val
		}
		return &object.ReturnValue{Value: val}
	case *ast.MutableStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.Set(node.Name.Value, val)
		return val
	case *ast.LetStatement:
		val := Eval(node.Value, env)
		if isError(val) {
			return val
		}
		env.SetLet(node.Name.Value, val)
		return val
	case *ast.Identifier:
		return evalIdentifier(node, env)
	case *ast.FunctionLiteral:
		params := node.Parameters
		body := node.Body
		defaults := node.Defaults
		docstring := node.DocString
		return &object.Function{Parameters: params, Env: env, Body: body, Defaults: defaults, DocString: docstring}
	case *ast.FunctionDefineLiteral:
		params := node.Parameters
		body := node.Body
		defaults := node.Defaults
		env.SetLet(node.TokenLiteral(), &object.Function{Parameters: params, Env: env, Body: body, Defaults: defaults})
		return &object.Boolean{Value: false}
	case *ast.CallExpression:
		if node.Function.TokenLiteral() == "quote" {
			return quote(node.Arguments[0], env)
		}

		function := Eval(node.Function, env)
		if isError(function) {
			return function
		}

		args := evalExpression(node.Arguments, env)

		// check for current args (...)
		if len(args) > 0 {
			firstArg, ok := args[0].(*object.Array)
			if ok && firstArg.IsCurrentArgs {
				newArgs := env.CurrentArgs
				args = append(newArgs, args[1:]...)
			}
		}

		if len(args) == 1 && isError(args[0]) {
			return args[0]
		}

		res := ApplyFunction(env, function, args)
		if isError(res) {
			fmt.Fprintf(os.Stderr, "Error calling `%s` : %s\n", node.Function, res.Inspect())
			os.Exit(1)
			return res
		}

		return res

	case *ast.ArrayLiteral:
		elements := evalExpression(node.Elements, env)
		if len(elements) == 1 && isError(elements[0]) {
			return elements[0]
		}
		return &object.Array{Elements: elements}
	case *ast.StringLiteral:
		return &object.String{Value: Interpolate(node.Value, env)}
	case *ast.CurrentArgsLiteral:
		return &object.Array{Token: node.Token, Elements: env.CurrentArgs, IsCurrentArgs: true}
	case *ast.RegexpLiteral:
		return &object.Regexp{Value: node.Value, Flags: node.Flags}
	case *ast.IndexExpression:
		left := Eval(node.Left, env)
		if isError(left) {
			return left
		}
		index := Eval(node.Index, env)
		if isError(index) {
			return index
		}
		return evalIndexExpression(left, index, env)
	case *ast.AssignStatement:
		return evalAssignStatement(node, env)
	case *ast.HashLiteral:
		return evalHashLiteral(node, env)
	}
	return nil
}

// eval block statement
func evalBlockStatement(block *ast.BlockStatement, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range block.Statements {
		result = Eval(statement, env)
		if result != nil {
			rt := result.Type()
			if rt == object.RETURN_VALUE_OBJ || rt == object.ERROR_OBJ {
				return result
			}
		}
	}
	return result
}

// EvalModule evaluates the named module and returns a *object.Module object
// This creates a whole new cozy instance (lexer, parser, env, and evaluator),
// which isn't ideal, but we also do this with calls to stdlib's `eval()` and
// when interpreting interpolated strings.
func EvalModule(name string) object.Object {
	filename := FindModule(name)
	if filename == "" {
		return newError("ImportError: no module named '%s'", name)
	}

	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return newError("IOError: error reading module '%s': %s", name, err)
	}

	l := lexer.New(string(b))
	p := parser.New(l)

	module := p.ParseProgram()
	if len(p.Errors()) != 0 {
		return newError("ParseError: %s", p.Errors())
	}

	env := object.NewEnvironment()
	macroEnv := object.NewEnvironment()
	DefineMacros(module, macroEnv)
	expanded := ExpandMacros(module, macroEnv)
	Eval(expanded, env)

	return env.ExportedHash()
}

func evalWhileExpression(we *ast.WhileExpression, env *object.Environment) object.Object {
	var result object.Object

	for {
		condition := Eval(we.Condition, env)
		if isError(condition) {
			return condition
		}

		if isTruthy(condition) {
			result = Eval(we.Consequence, env)
		} else {
			break
		}
	}

	if result != nil {
		return result
	}

	return &object.Boolean{Value: false}
}

func evalImportExpression(ie *ast.ImportExpression, env *object.Environment) object.Object {
	name := Eval(ie.Name, env)
	if isError(name) {
		return name
	}

	if s, ok := name.(*object.String); ok {
		attrs := EvalModule(s.Value)
		if isError(attrs) {
			return attrs
		}
		return &object.Module{Name: s.Value, Attrs: attrs}
	}
	return newError("ImportError: invalid import path '%s'", name)
}

// for performance, using single instance of boolean
func nativeBoolToBooleanObject(input bool) *object.Boolean {
	if input {
		return TRUE
	}
	return FALSE
}

// eval prefix expression
func evalPrefixExpression(operator string, right object.Object) object.Object {
	switch operator {
	case "!":
		return evalBangOperatorExpression(right)
	case "-":
		return evalMinusPrefixOperatorExpression(right)
	default:
		return newError("unknown operator: %s%s", operator, right.Type())
	}
}

func evalPostfixExpression(env *object.Environment, operator string, node *ast.PostfixExpression) object.Object {
	switch operator {
	case "++":
		val, ok := env.Get(node.Token.Literal)
		if !ok {
			return newError("%s is unknown", node.Token.Literal)
		}

		switch arg := val.(type) {
		case *object.Integer:
			v := arg.Value
			env.Set(node.Token.Literal, &object.Integer{Value: v + 1})
			return arg
		default:
			return newError("%s is not an int", node.Token.Literal)
		}
	case "--":
		val, ok := env.Get(node.Token.Literal)
		if !ok {
			return newError("%s is unknown", node.Token.Literal)
		}

		switch arg := val.(type) {
		case *object.Integer:
			v := arg.Value
			env.Set(node.Token.Literal, &object.Integer{Value: v - 1})
			return arg
		default:
			return newError("%s is not an int", node.Token.Literal)
		}

	default:
		return newError("unknown operator: %s", operator)
	}
}

func evalBangOperatorExpression(right object.Object) object.Object {
	switch right {
	case TRUE:
		return FALSE
	case FALSE:
		return TRUE
	default:
		return FALSE
	}
}

func evalMinusPrefixOperatorExpression(right object.Object) object.Object {
	switch obj := right.(type) {
	case *object.Integer:
		return &object.Integer{Value: -obj.Value}
	case *object.Float:
		return &object.Float{Value: -obj.Value}
	default:
		return newError("unknown operator: -%s", right.Type())
	}
}

func evalInfixExpression(operator string, left, right object.Object, env *object.Environment) object.Object {
	switch {
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalIntegerInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.FLOAT_OBJ:
		return evalFloatInfixExpression(operator, left, right)
	case left.Type() == object.FLOAT_OBJ && right.Type() == object.INTEGER_OBJ:
		return evalFloatIntegerInfixExpression(operator, left, right)
	case left.Type() == object.INTEGER_OBJ && right.Type() == object.FLOAT_OBJ:
		return evalIntegerFloatInfixExpression(operator, left, right)
	case left.Type() == object.STRING_OBJ && right.Type() == object.STRING_OBJ:
		return evalStringInfixExpression(operator, left, right)
	case operator == "&&":
		return nativeBoolToBooleanObject(objectToNativeBoolean(left) && objectToNativeBoolean(right))
	case operator == "||":
		return nativeBoolToBooleanObject(objectToNativeBoolean(left) || objectToNativeBoolean(right))
	case operator == "!~":
		return notMatches(left, right)
	case operator == "~=":
		return matches(left, right, env)

	case operator == "==":
		return nativeBoolToBooleanObject(left == right)

	case operator == "!=":
		return nativeBoolToBooleanObject(left != right)
	case left.Type() == object.BOOLEAN_OBJ && right.Type() == object.BOOLEAN_OBJ:
		return evalBooleanInfixExpression(operator, left, right)
	case left.Type() != right.Type():
		return newError("type mismatch: %s %s %s",
			left.Type(), operator, right.Type())
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func matches(left, right object.Object, env *object.Environment) object.Object {
	str := left.Inspect()

	if right.Type() != object.REGEXP_OBJ {
		return newError("regexp required for regexp-match, given %s", right.Type())
	}

	val := right.(*object.Regexp).Value
	if right.(*object.Regexp).Flags != "" {
		val = "(?" + right.(*object.Regexp).Flags + ")" + val
	}

	// Compile the regular expression.
	r, err := regexp.Compile(val)

	// Ensure it compiled
	if err != nil {
		return newError("error compiling regexp '%s': %s", right.Inspect(), err)
	}

	res := r.FindStringSubmatch(str)

	// Do we have any captures?
	if len(res) > 1 {
		for i := 1; i < len(res); i++ {
			env.Set(fmt.Sprintf("$%d", i), &object.String{Value: res[i]})
		}
	}

	// Test if it matched
	if len(res) > 0 {
		return TRUE
	}

	return FALSE
}

func notMatches(left, right object.Object) object.Object {
	str := left.Inspect()

	if right.Type() != object.REGEXP_OBJ {
		return newError("regexp required for regexp-match, given %s", right.Type())
	}

	val := right.(*object.Regexp).Value
	if right.(*object.Regexp).Flags != "" {
		val = "(?" + right.(*object.Regexp).Flags + ")" + val
	}

	// Compile the regular expression.
	r, err := regexp.Compile(val)

	// Ensure it compiled
	if err != nil {
		return newError("error compiling regexp '%s': %s", right.Inspect(), err)
	}

	// Test if it matched
	if r.MatchString(str) {
		return FALSE
	}

	return TRUE
}

// boolean operations
func evalBooleanInfixExpression(operator string, left, right object.Object) object.Object {
	// convert the bools to strings.
	l := &object.String{Value: string(left.Inspect())}
	r := &object.String{Value: string(right.Inspect())}

	switch operator {
	case "<":
		return evalStringInfixExpression(operator, l, r)
	case "<=":
		return evalStringInfixExpression(operator, l, r)
	case ">":
		return evalStringInfixExpression(operator, l, r)
	case ">=":
		return evalStringInfixExpression(operator, l, r)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Integer).Value
	rightVal := right.(*object.Integer).Value
	switch operator {
	case "+":
		return &object.Integer{Value: leftVal + rightVal}
	case "+=":
		return &object.Integer{Value: leftVal + rightVal}
	case "%":
		return &object.Integer{Value: leftVal % rightVal}
	case "**":
		return &object.Integer{Value: int64(math.Pow(float64(leftVal), float64(rightVal)))}
	case "-":
		return &object.Integer{Value: leftVal - rightVal}
	case "-=":
		return &object.Integer{Value: leftVal - rightVal}
	case "*":
		return &object.Integer{Value: leftVal * rightVal}
	case "*=":
		return &object.Integer{Value: leftVal * rightVal}
	case "/":
		return &object.Integer{Value: leftVal / rightVal}
	case "/=":
		return &object.Integer{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	case "..":
		len := int(rightVal-leftVal) + 1
		array := make([]object.Object, len)
		i := 0
		for i < len {
			array[i] = &object.Integer{Value: leftVal}
			leftVal++
			i++
		}
		return &object.Array{Elements: array}
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}
func evalFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Float).Value
	rightVal := right.(*object.Float).Value
	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "+=":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "-=":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "*=":
		return &object.Float{Value: leftVal * rightVal}
	case "**":
		return &object.Float{Value: math.Pow(leftVal, rightVal)}
	case "/":
		return &object.Float{Value: leftVal / rightVal}
	case "/=":
		return &object.Float{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalFloatIntegerInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := left.(*object.Float).Value
	rightVal := float64(right.(*object.Integer).Value)
	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "+=":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "-=":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "*=":
		return &object.Float{Value: leftVal * rightVal}
	case "**":
		return &object.Float{Value: math.Pow(leftVal, rightVal)}
	case "/":
		return &object.Float{Value: leftVal / rightVal}
	case "/=":
		return &object.Float{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalIntegerFloatInfixExpression(operator string, left, right object.Object) object.Object {
	leftVal := float64(left.(*object.Integer).Value)
	rightVal := right.(*object.Float).Value
	switch operator {
	case "+":
		return &object.Float{Value: leftVal + rightVal}
	case "+=":
		return &object.Float{Value: leftVal + rightVal}
	case "-":
		return &object.Float{Value: leftVal - rightVal}
	case "-=":
		return &object.Float{Value: leftVal - rightVal}
	case "*":
		return &object.Float{Value: leftVal * rightVal}
	case "*=":
		return &object.Float{Value: leftVal * rightVal}
	case "**":
		return &object.Float{Value: math.Pow(leftVal, rightVal)}
	case "/":
		return &object.Float{Value: leftVal / rightVal}
	case "/=":
		return &object.Float{Value: leftVal / rightVal}
	case "<":
		return nativeBoolToBooleanObject(leftVal < rightVal)
	case "<=":
		return nativeBoolToBooleanObject(leftVal <= rightVal)
	case ">":
		return nativeBoolToBooleanObject(leftVal > rightVal)
	case ">=":
		return nativeBoolToBooleanObject(leftVal >= rightVal)
	case "==":
		return nativeBoolToBooleanObject(leftVal == rightVal)
	case "!=":
		return nativeBoolToBooleanObject(leftVal != rightVal)
	default:
		return newError("unknown operator: %s %s %s",
			left.Type(), operator, right.Type())
	}
}

func evalStringInfixExpression(operator string, left, right object.Object) object.Object {
	l := left.(*object.String)
	r := right.(*object.String)

	switch operator {
	case "==":
		return nativeBoolToBooleanObject(l.Value == r.Value)
	case "!=":
		return nativeBoolToBooleanObject(l.Value != r.Value)
	case ">=":
		return nativeBoolToBooleanObject(l.Value >= r.Value)
	case ">":
		return nativeBoolToBooleanObject(l.Value > r.Value)
	case "<=":
		return nativeBoolToBooleanObject(l.Value <= r.Value)
	case "<":
		return nativeBoolToBooleanObject(l.Value < r.Value)
	case "+":
		return &object.String{Value: l.Value + r.Value}
	case "+=":
		return &object.String{Value: l.Value + r.Value}
	}

	return newError("unknown operator: %s %s %s",
		left.Type(), operator, right.Type())
}

// evalIfExpression handles an `if` expression, running the block
// if the condition matches, and running any optional else block
// otherwise.
func evalIfExpression(ie *ast.IfExpression, env *object.Environment) object.Object {
	// Create an environment for handling regexps
	var permit []string
	i := 1
	for i < 32 {
		permit = append(permit, fmt.Sprintf("$%d", i))
		i++
	}
	nEnv := object.NewTemporaryScope(env, permit)
	condition := Eval(ie.Condition, nEnv)
	if isError(condition) {
		return condition
	}
	if isTruthy(condition) {
		return Eval(ie.Consequence, nEnv)
	} else if ie.Alternative != nil {
		return Eval(ie.Alternative, nEnv)
	} else {
		return &object.Boolean{Value: false}
	}
}

// evalTernaryExpression handles a ternary-expression. If the condition
// is true we return the contents of evaluating the true-branch, otherwise
// the false-branch. (Unlike an `if` statement we know that we always have
// an alternative/false branch.)
func evalTernaryExpression(te *ast.TernaryExpression, env *object.Environment) object.Object {

	condition := Eval(te.Condition, env)
	if isError(condition) {
		return condition
	}

	if isTruthy(condition) {
		return Eval(te.IfTrue, env)
	}
	return Eval(te.IfFalse, env)
}

func evalAssignStatement(a *ast.AssignStatement, env *object.Environment) (val object.Object) {
	evaluated := Eval(a.Value, env)
	if isError(evaluated) {
		return evaluated
	}

	// An assignment is generally:
	//    variable = value
	// But we cheat and reuse the implementation for:
	//    i += 4
	// In this case we record the "operator" as "+="
	switch a.Operator {
	case "+=":
		// Get the current value
		current, ok := env.Get(a.Name.String())
		if !ok {
			return newError("%s is unknown", a.Name.String())
		}

		res := evalInfixExpression("+=", current, evaluated, env)
		if isError(res) {
			fmt.Printf("Error handling += %s\n", res.Inspect())
			return res
		}

		env.Set(a.Name.String(), res)
		return res

	case "-=":

		// Get the current value
		current, ok := env.Get(a.Name.String())
		if !ok {
			return newError("%s is unknown", a.Name.String())
		}

		res := evalInfixExpression("-=", current, evaluated, env)
		if isError(res) {
			fmt.Printf("Error handling -= %s\n", res.Inspect())
			return res
		}

		env.Set(a.Name.String(), res)
		return res

	case "*=":
		// Get the current value
		current, ok := env.Get(a.Name.String())
		if !ok {
			return newError("%s is unknown", a.Name.String())
		}

		res := evalInfixExpression("*=", current, evaluated, env)
		if isError(res) {
			fmt.Printf("Error handling *= %s\n", res.Inspect())
			return res
		}

		env.Set(a.Name.String(), res)
		return res

	case "/=":

		// Get the current value
		current, ok := env.Get(a.Name.String())
		if !ok {
			return newError("%s is unknown", a.Name.String())
		}

		res := evalInfixExpression("/=", current, evaluated, env)
		if isError(res) {
			fmt.Printf("Error handling /= %s\n", res.Inspect())
			return res
		}

		env.Set(a.Name.String(), res)
		return res

	case "=":
		_, ok := env.Get(a.Name.String())
		if !ok {
			fmt.Printf("Setting unknown variable '%s' is an error!\n", a.Name.String())
			os.Exit(1)
		}

		env.Set(a.Name.String(), evaluated)
	}
	return evaluated
}

func evalForLoopExpression(fle *ast.ForLoopExpression, env *object.Environment) object.Object {
	rt := &object.Boolean{Value: true}
	for {
		condition := Eval(fle.Condition, env)
		if isError(condition) {
			return condition
		}
		if isTruthy(condition) {
			rt := Eval(fle.Consequence, env)
			if !isError(rt) && (rt.Type() == object.RETURN_VALUE_OBJ || rt.Type() == object.ERROR_OBJ) {
				return rt
			}
		} else {
			break
		}
	}
	return rt
}

// handle "for x [,y] in .."
func evalForeachExpression(fle *ast.ForeachStatement, env *object.Environment) object.Object {

	// expression
	val := Eval(fle.Value, env)

	helper, ok := val.(object.Iterable)
	if !ok {
		return newError("%s object doesn't implement the Iterable interface", val.Type())
	}

	// The one/two values we're going to permit
	var permit []string
	permit = append(permit, fle.Ident)
	if fle.Index != "" {
		permit = append(permit, fle.Index)
	}

	// Create a new environment for the block
	// This will allow writing EVERYTHING to the parent scope,
	// except the two variables named in the permit-array
	child := object.NewTemporaryScope(env, permit)

	// Reset the state of any previous iteration.
	helper.Reset()

	// Get the initial values.
	ret, idx, ok := helper.Next()

	for ok {

		// Set the index + name
		child.Set(fle.Ident, ret)

		idxName := fle.Index
		if idxName != "" {
			child.Set(fle.Index, idx)
		}

		// Eval the block
		rt := Eval(fle.Body, child)

		// If we got an error/return then we handle it.
		if !isError(rt) && (rt.Type() == object.RETURN_VALUE_OBJ || rt.Type() == object.ERROR_OBJ) {
			return rt
		}

		// Loop again
		ret, idx, ok = helper.Next()
	}

	return &object.Boolean{Value: true}
}

func isTruthy(obj object.Object) bool {
	switch obj {
	case TRUE:
		return true
	case FALSE:
		return false
	default:
		return true
	}
}

func evalProgram(program *ast.Program, env *object.Environment) object.Object {
	var result object.Object
	for _, statement := range program.Statements {
		result = Eval(statement, env)
		switch result := result.(type) {
		case *object.ReturnValue:
			return result.Value
		case *object.Error:
			return result
		}
	}
	return result
}

func newError(format string, a ...interface{}) *object.Error {
	message := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, message+"\n")
	return &object.Error{Message: message}
}

func isError(obj object.Object) bool {
	if obj != nil {
		return obj.Type() == object.ERROR_OBJ
	}
	return false
}

func evalIdentifier(node *ast.Identifier, env *object.Environment) object.Object {
	if val, ok := env.Get(node.Value); ok {
		return val
	}
	if builtin, ok := builtins[node.Value]; ok {
		return builtin
	}
	fmt.Fprintf(os.Stderr, "identifier not found: %s\n", node.Value)
	os.Exit(1)
	return newError("identifier not found: " + node.Value)
}

func evalExpression(exps []ast.Expression, env *object.Environment) []object.Object {
	var result []object.Object
	for _, e := range exps {
		evaluated := Eval(e, env)
		if isError(evaluated) {
			return []object.Object{evaluated}
		}
		result = append(result, evaluated)
	}
	return result
}

// Split a line of text into tokens, but keep anything "quoted"
// together..
// So this input:
//   /bin/sh -c "ls /etc"
// Would give output of the form:
//   /bin/sh
//   -c
//   ls /etc
func splitCommand(input string) []string {

	// This does the split into an array
	r := regexp.MustCompile(`[^\s"']+|"([^"]*)"|'([^']*)`)
	res := r.FindAllString(input, -1)

	// However the resulting pieces might be quoted.
	// So we have to remove them, if present.
	var result []string
	for _, e := range res {
		result = append(result, trimQuotes(e, '"'))
	}
	return result
}

// Remove balanced characters around a string.
func trimQuotes(in string, c byte) string {
	if len(in) >= 2 {
		if in[0] == c && in[len(in)-1] == c {
			return in[1 : len(in)-1]
		}
	}
	return in
}

func evalIndexExpression(left, index object.Object, env *object.Environment) object.Object {
	switch {
	case left.Type() == object.ARRAY_OBJ:
		return evalArrayIndexExpression(left, index, env)
	case left.Type() == object.HASH_OBJ:
		return evalHashIndexExpression(left, index, env)
	case left.Type() == object.STRING_OBJ:
		return evalStringIndexExpression(left, index, env)
	case left.Type() == object.MODULE_OBJ:
		return evalModuleIndexExpression(left, index, env)
	default:
		if fn, ok := objectGetMethod(left, index, env); ok {
			return fn
		}
		return newError("index operator not support:%s", left.Type())

	}
}

func evalModuleIndexExpression(module, index object.Object, env *object.Environment) object.Object {
	moduleObject := module.(*object.Module)
	return evalHashIndexExpression(moduleObject.Attrs, index, env)
}

func evalArrayIndexExpression(array, index object.Object, env *object.Environment) object.Object {
	arrayObject := array.(*object.Array)
	switch t := index.(type) {
	case *object.Integer:
		idx := t.Value
		max := int64(len(arrayObject.Elements) - 1)
		if idx < 0 || idx > max {
			return &object.Error{Message: "Indexing failed on array, out of bounds"}
		}
		return arrayObject.Elements[idx]
	default:
		if fn, ok := objectGetMethod(array, index, env); ok {
			return fn
		}
		return &object.Error{Message: "Indexing on array failed"}
	}
}

func evalHashIndexExpression(hash, index object.Object, env *object.Environment) object.Object {
	hashObject := hash.(*object.Hash)
	key, ok := index.(object.Hashable)
	if !ok {
		return newError("unusable as hash key: %s", index.Type())
	}
	pair, ok := hashObject.Pairs[key.HashKey()]
	if !ok {
		var fn object.Object
		if fn, ok = objectGetMethod(hash, index, env); ok {
			return fn
		}
		return &object.Error{Message: "Indexing not possible on this object"}
	}
	return pair.Value
}

func evalStringIndexExpression(input, index object.Object, env *object.Environment) object.Object {
	str := input.(*object.String).Value
	switch t := index.(type) {
	case *object.Integer:
		idx := t.Value
		max := int64(len(str))
		if idx < 0 || idx > max {
			return &object.Error{Message: "Indexing not possible on this object"}
		}

		// Get the characters as an array of runes
		chars := []rune(str)

		// Now index
		ret := chars[idx]

		// And return as a string.
		return &object.String{Value: string(ret)}
	default:
		if fn, ok := objectGetMethod(input, index, env); ok {
			return fn
		}

		return &object.Error{Message: "Indexing not possible on this object"}
	}
}

func evalHashLiteral(node *ast.HashLiteral, env *object.Environment) object.Object {
	pairs := make(map[object.HashKey]object.HashPair)
	for keyNode, valueNode := range node.Pairs {
		key := Eval(keyNode, env)
		if isError(key) {
			return key
		}
		hashKey, ok := key.(object.Hashable)
		if !ok {
			return newError("unusable as hash key: %s", key.Type())
		}
		value := Eval(valueNode, env)
		if isError(value) {
			return value
		}
		hashed := hashKey.HashKey()
		pairs[hashed] = object.HashPair{Key: key, Value: value}

	}
	return &object.Hash{Pairs: pairs}

}

// ApplyFunction applies a function in an environment
func ApplyFunction(env *object.Environment, fn object.Object, args []object.Object) object.Object {
	switch fn := fn.(type) {
	case *object.Function:
		extendEnv := extendFunctionEnv(fn, args)
		evaluated := Eval(fn.Body, extendEnv)
		return upwrapReturnValue(evaluated)
	case *object.Builtin:
		return fn.Fn(env, args...)
	default:
		return newError("not a function: %s", fn.Type())
	}
}

func extendFunctionEnv(fn *object.Function, args []object.Object) *object.Environment {
	env := object.NewEnclosedEnvironment(fn.Env, args)

	// Set the defaults
	for key, val := range fn.Defaults {
		env.Set(key, Eval(val, env))
	}
	for paramIdx, param := range fn.Parameters {
		if paramIdx < len(args) {
			env.Set(param.Value, args[paramIdx])
		}
	}
	return env
}

func upwrapReturnValue(obj object.Object) object.Object {
	if returnValue, ok := obj.(*object.ReturnValue); ok {
		return returnValue.Value
	}
	return obj
}

// RegisterBuiltin registers a built-in function. This is used to register
// our "standard library" functions.
func RegisterBuiltin(name string, fun object.BuiltinFunction) {
	builtins[name] = &object.Builtin{Fn: fun}
}

// SetContext lets you configure a context, which is helpful if you wish to
// cause execution to timeout after a given period, for example.
func SetContext(ctx context.Context) {
	CTX = ctx
}

func objectGetMethod(o, key object.Object, env *object.Environment) (ret object.Object, ok bool) {
	switch k := key.(type) {
	case *object.String:
		var fn object.BuiltinFunction
		if fn = o.GetMethod(k.Value); fn != nil {
			return &object.Builtin{Fn: fn}, true
		}

		// If we reach this point then the invokation didn't
		// succeed, that probably means that the function wasn't
		// implemented in go.
		// So now we want to look for it in cozy, and we have
		// enough details to find the appropriate function.
		//  * We have the object involved.
		//  * We have the type of that object.
		//  * We have the name of the function.
		//  * We have the arguments.
		//
		// We'll use the type + name to lookup the (global) function
		// to invoke. For example in this case we'll invoke
		// `string.foo()` - because the type of the object we're
		// invoking-against is string:
		//  "zac".foo();
		// For this case we'll be looking for `array.foo()`.
		//   let a = [ 1, 2, 3 ];
		//   print(a.foo());
		// As a final fall-back we'll look for "object.foo()"
		// if "array.foo()" isn't defined.
		attempts := []string{}
		if _, ok = object.SystemTypesMap[o.Type()]; ok {
			attempts = append(attempts, strings.ToLower(string(o.Type())))
		} else {
			attempts = append(attempts, string(o.Type()))
		}
		attempts = append(attempts, "object")

		// Look for "$type.name", or "object.name"
		for _, prefix := range attempts {

			// What we're attempting to execute.
			name := prefix + "." + k.Value

			// Try to find that function in our environment.
			if val, ok := env.Get(name); ok {
				if fn, ok := val.(*object.Function); ok {
					copyFn := *fn
					emptyArgs := make([]object.Object, 0)
					copyFn.Env = object.NewEnclosedEnvironment(fn.Env, emptyArgs)
					copyFn.Env.Set("self", o)
					return &copyFn, true
				}
				return val, true
			}
		}
	}
	return nil, false
}

func objectToNativeBoolean(o object.Object) bool {
	if r, ok := o.(*object.ReturnValue); ok {
		o = r.Value
	}
	switch obj := o.(type) {
	case *object.Boolean:
		return obj.Value
	case *object.String:
		return obj.Value != ""
	case *object.Regexp:
		return obj.Value != ""
	case *object.Integer:
		if obj.Value == 0 {
			return false
		}
		return true
	case *object.Float:
		if obj.Value == 0.0 {
			return false
		}
		return true
	case *object.Array:
		if len(obj.Elements) == 0 {
			return false
		}
		return true
	case *object.Hash:
		if len(obj.Pairs) == 0 {
			return false
		}
		return true
	default:
		return true
	}
}

// Macro-related code
func DefineMacros(program *ast.Program, env *object.Environment) {
	definitions := []int{}

	for i, statement := range program.Statements {
		if isMacroDefinition(statement) {
			addMacro(statement, env)
			definitions = append(definitions, i)
		}
	}

	for i := len(definitions) - 1; i >= 0; i = i - 1 {
		definitionIndex := definitions[i]
		program.Statements = append(
			program.Statements[:definitionIndex],
			program.Statements[definitionIndex+1:]...,
		)
	}
}

func isMacroDefinition(node ast.Statement) bool {
	letStatement, ok := node.(*ast.LetStatement)
	if !ok {
		return false
	}

	_, ok = letStatement.Value.(*ast.MacroLiteral)
	return ok
}

func addMacro(stmt ast.Statement, env *object.Environment) {
	letStatement, _ := stmt.(*ast.LetStatement)
	macroLiteral, _ := letStatement.Value.(*ast.MacroLiteral)

	macro := &object.Macro{
		Parameters: macroLiteral.Parameters,
		Env:        env,
		Body:       macroLiteral.Body,
	}

	env.SetLet(letStatement.Name.Value, macro)
}

func ExpandMacros(program ast.Node, env *object.Environment) ast.Node {
	return ast.Modify(program, func(node ast.Node) ast.Node {
		callExpression, ok := node.(*ast.CallExpression)
		if !ok {
			return node
		}

		macro, ok := isMacroCall(callExpression, env)
		if !ok {
			return node
		}

		args := quoteArgs(callExpression)
		evalEnv := extendMacroEnv(macro, args)

		evaluated := Eval(macro.Body, evalEnv)

		quote, ok := evaluated.(*object.Quote)
		if !ok {
			panic("we only support returning AST-nodes from macros")
		}

		return quote.Node
	})
}

func isMacroCall(
	exp *ast.CallExpression,
	env *object.Environment,
) (*object.Macro, bool) {
	identifier, ok := exp.Function.(*ast.Identifier)
	if !ok {
		return nil, false
	}

	obj, ok := env.Get(identifier.Value)
	if !ok {
		return nil, false
	}

	macro, ok := obj.(*object.Macro)
	if !ok {
		return nil, false
	}

	return macro, true
}

func quoteArgs(exp *ast.CallExpression) []*object.Quote {
	args := []*object.Quote{}

	for _, a := range exp.Arguments {
		args = append(args, &object.Quote{Node: a})
	}

	return args
}

func extendMacroEnv(
	macro *object.Macro,
	args []*object.Quote,
) *object.Environment {
	emptyArgs := make([]object.Object, 0)
	extended := object.NewEnclosedEnvironment(macro.Env, emptyArgs)

	for paramIdx, param := range macro.Parameters {
		extended.Set(param.Value, args[paramIdx])
	}

	return extended
}

func quote(node ast.Node, env *object.Environment) object.Object {
	node = evalUnquoteCalls(node, env)
	return &object.Quote{Node: node}
}

func evalUnquoteCalls(quoted ast.Node, env *object.Environment) ast.Node {
	return ast.Modify(quoted, func(node ast.Node) ast.Node {
		if !isUnquoteCall(node) {
			return node
		}

		call, ok := node.(*ast.CallExpression)
		if !ok {
			return node
		}

		if len(call.Arguments) != 1 {
			return node
		}

		unquoted := Eval(call.Arguments[0], env)
		return convertObjectToASTNode(unquoted)
	})
}

func convertObjectToASTNode(obj object.Object) ast.Node {
	switch obj := obj.(type) {
	case *object.Integer:
		t := token.Token{
			Type:    token.INT,
			Literal: fmt.Sprintf("%d", obj.Value),
		}
		return &ast.IntegerLiteral{Token: t, Value: obj.Value}
	case *object.Boolean:
		var t token.Token
		if obj.Value {
			t = token.Token{Type: token.TRUE, Literal: "true"}
		} else {
			t = token.Token{Type: token.FALSE, Literal: "false"}
		}
		return &ast.Boolean{Token: t, Value: obj.Value}
	case *object.Quote:
		return obj.Node
	default:
		return nil
	}
}

func isUnquoteCall(node ast.Node) bool {
	callExpression, ok := node.(*ast.CallExpression)
	if !ok {
		return false
	}

	return callExpression.Function.TokenLiteral() == "unquote"
}
