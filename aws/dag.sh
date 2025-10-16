#!/bin/bash

# Script to parse cdk.out/manifest.json and display stack dependencies as a DAG

set -e

MANIFEST_FILE="cdk.out/manifest.json"

if [ ! -f "$MANIFEST_FILE" ]; then
    echo "Error: $MANIFEST_FILE not found. Run 'cdk synth' first."
    exit 1
fi

# Extract stack dependencies using jq
# Filter only CloudFormation stacks (not assets)
stacks=$(jq -r '.artifacts | to_entries[] | select(.value.type == "aws:cloudformation:stack") | .key' "$MANIFEST_FILE")

echo "CDK Stack Dependency Graph"
echo "=========================="
echo ""

# Build dependency map
declare -A stack_deps

for stack in $stacks; do
    deps=$(jq -r --arg stack "$stack" \
        '.artifacts[$stack].dependencies[]? // empty | select(. | endswith(".assets") | not)' \
        "$MANIFEST_FILE" | tr '\n' ' ')
    stack_deps[$stack]="$deps"
done

# Print as directed graph showing edges
echo "Stack Dependencies (A → B means A depends on B):"
echo ""

# Find all edges and group by source
has_deps=false
for stack in $stacks; do
    deps="${stack_deps[$stack]}"
    if [ -n "$deps" ]; then
        has_deps=true
        echo "  $stack"
        for dep in $deps; do
            echo "    └─→ $dep"
        done
        echo ""
    fi
done

if [ "$has_deps" = "false" ]; then
    echo "  (No dependencies found)"
    echo ""
fi

# Show stacks with no dependencies
echo "Base Stacks (no dependencies):"
echo ""
base_found=false
for stack in $stacks; do
    if [ -z "${stack_deps[$stack]}" ]; then
        echo "  • $stack"
        base_found=true
    fi
done
if [ "$base_found" = "false" ]; then
    echo "  (None - all stacks depend on something)"
fi

echo ""

# Show reverse dependencies (what depends on each stack)
echo "Reverse Dependencies (stacks that depend on each):"
echo ""

for stack in $stacks; do
    # Find what depends on this stack
    dependents=()
    for s in $stacks; do
        deps="${stack_deps[$s]}"
        for d in $deps; do
            if [ "$d" = "$stack" ]; then
                dependents+=("$s")
                break
            fi
        done
    done

    if [ ${#dependents[@]} -gt 0 ]; then
        echo "  $stack ← required by:"
        for dependent in "${dependents[@]}"; do
            echo "    • $dependent"
        done
        echo ""
    fi
done

echo ""

echo "Deployment Order"
echo "================"
echo ""

# Topological sort to determine deployment order
declare -A visited
declare -A in_progress
deployment_order=()

function visit() {
    local node=$1

    if [ "${visited[$node]}" = "true" ]; then
        return
    fi

    if [ "${in_progress[$node]}" = "true" ]; then
        echo "Error: Circular dependency detected at $node"
        exit 1
    fi

    in_progress[$node]="true"

    # Visit dependencies first
    local deps=$(jq -r --arg stack "$node" \
        '.artifacts[$stack].dependencies[]? // empty | select(. | endswith(".assets") | not)' \
        "$MANIFEST_FILE")

    for dep in $deps; do
        visit "$dep"
    done

    in_progress[$node]="false"
    visited[$node]="true"
    deployment_order+=("$node")
}

# Visit all stacks
for stack in $stacks; do
    visit "$stack"
done

# Print deployment order
order_num=1
for stack in "${deployment_order[@]}"; do
    echo "$order_num. $stack"
    ((order_num++))
done
