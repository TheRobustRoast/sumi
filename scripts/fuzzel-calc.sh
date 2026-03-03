#!/usr/bin/env bash
# ╔══════════════════════════════════════════════════════════════╗
# ║  Fuzzel Calculator                                          ║
# ║  Type a math expression, result copies to clipboard         ║
# ║  Uses python3 for evaluation (safe subset via ast)          ║
# ╚══════════════════════════════════════════════════════════════╝

RESULT=$(echo "" | fuzzel --dmenu \
    --prompt "calc: " \
    --width 25 \
    --lines 0)

[[ -z "$RESULT" ]] && exit 0

# Evaluate safely using Python's ast.literal_eval for simple math
# Falls back to bc for more complex expressions
ANSWER=$(python3 -c "
import ast, operator, math

ops = {
    ast.Add: operator.add, ast.Sub: operator.sub,
    ast.Mult: operator.mul, ast.Div: operator.truediv,
    ast.Pow: operator.pow, ast.Mod: operator.mod,
    ast.USub: operator.neg, ast.UAdd: operator.pos,
}

def safe_eval(node):
    if isinstance(node, ast.Constant):
        return node.value
    elif isinstance(node, ast.BinOp):
        return ops[type(node.op)](safe_eval(node.left), safe_eval(node.right))
    elif isinstance(node, ast.UnaryOp):
        return ops[type(node.op)](safe_eval(node.operand))
    elif isinstance(node, ast.Call) and isinstance(node.func, ast.Name):
        func = getattr(math, node.func.id, None)
        if func:
            args = [safe_eval(a) for a in node.args]
            return func(*args)
    raise ValueError('unsafe')

try:
    expr = '$RESULT'
    tree = ast.parse(expr, mode='eval')
    result = safe_eval(tree.body)
    print(result)
except:
    print('error')
" 2>/dev/null)

if [[ "$ANSWER" == "error" ]] || [[ -z "$ANSWER" ]]; then
    # Fallback to bc
    ANSWER=$(echo "$RESULT" | bc -l 2>/dev/null)
fi

if [[ -n "$ANSWER" ]] && [[ "$ANSWER" != "error" ]]; then
    echo -n "$ANSWER" | wl-copy
    notify-send -a sumi -t 2000 "Calculator" "$RESULT = $ANSWER (copied)"
else
    notify-send -a sumi -t 2000 "Calculator" "Invalid expression: $RESULT"
fi
