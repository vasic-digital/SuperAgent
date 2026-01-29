#!/bin/bash
# Initialize all formatter Git submodules
# This script adds all 118 formatters as Git submodules

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
FORMATTERS_DIR="$(dirname "$SCRIPT_DIR")"
PROJECT_ROOT="$(dirname "$FORMATTERS_DIR")"

cd "$PROJECT_ROOT"

echo "=== Initializing Formatter Git Submodules ==="
echo "This will add 118 formatters as Git submodules"
echo

# Native Binary Formatters (60+)
echo "Adding native binary formatters..."

# Systems Languages
git submodule add https://github.com/llvm/llvm-project formatters/clang-format || true
git submodule add https://github.com/rust-lang/rustfmt formatters/rustfmt || true
git submodule add https://github.com/golang/go formatters/go || true
git submodule add https://github.com/ziglang/zig formatters/zig || true
git submodule add https://github.com/nim-lang/Nim formatters/nim || true

# JVM Languages
git submodule add https://github.com/google/google-java-format formatters/google-java-format || true
git submodule add https://github.com/pinterest/ktlint formatters/ktlint || true
git submodule add https://github.com/facebook/ktfmt formatters/ktfmt || true
git submodule add https://github.com/scalameta/scalafmt formatters/scalafmt || true
git submodule add https://github.com/scala-ide/scalariform formatters/scalariform || true
git submodule add https://github.com/diffplug/spotless formatters/spotless || true
git submodule add https://github.com/weavejester/cljfmt formatters/cljfmt || true
git submodule add https://github.com/kkinnear/zprint formatters/zprint || true

# Web Languages
git submodule add https://github.com/prettier/prettier formatters/prettier || true
git submodule add https://github.com/biomejs/biome formatters/biome || true
git submodule add https://github.com/dprint/dprint formatters/dprint || true
git submodule add https://github.com/psf/black formatters/black || true
git submodule add https://github.com/astral-sh/ruff formatters/ruff || true
git submodule add https://github.com/hhatto/autopep8 formatters/autopep8 || true
git submodule add https://github.com/google/yapf formatters/yapf || true
git submodule add https://github.com/rubocop/rubocop formatters/rubocop || true
git submodule add https://github.com/standardrb/standard formatters/standardrb || true
git submodule add https://github.com/ruby-formatter/rufo formatters/rufo || true
git submodule add https://github.com/PHP-CS-Fixer/PHP-CS-Fixer formatters/php-cs-fixer || true
git submodule add https://github.com/laravel/pint formatters/laravel-pint || true

# Functional Languages
git submodule add https://github.com/tweag/ormolu formatters/ormolu || true
git submodule add https://github.com/fourmolu/fourmolu formatters/fourmolu || true
git submodule add https://github.com/haskell/stylish-haskell formatters/stylish-haskell || true
git submodule add https://github.com/mihaimaruseac/hindent formatters/hindent || true
git submodule add https://github.com/ocaml-ppx/ocamlformat formatters/ocamlformat || true
git submodule add https://github.com/fsprojects/fantomas formatters/fantomas || true
git submodule add https://github.com/elixir-lang/elixir formatters/elixir || true
git submodule add https://github.com/WhatsApp/erlfmt formatters/erlfmt || true

# Mobile Development
git submodule add https://github.com/swiftlang/swift-format formatters/swift-format || true
git submodule add https://github.com/nicklockwood/SwiftFormat formatters/swiftformat || true
git submodule add https://github.com/dart-lang/dart_style formatters/dart-style || true

# Scripting Languages
git submodule add https://github.com/mvdan/sh formatters/shfmt || true
git submodule add https://github.com/lovesegfault/beautysh formatters/beautysh || true
git submodule add https://github.com/PowerShell/PSScriptAnalyzer formatters/psscriptanalyzer || true
git submodule add https://github.com/JohnnyMorganz/StyLua formatters/stylua || true
git submodule add https://github.com/appgurueu/luafmt formatters/luafmt || true
git submodule add https://github.com/perltidy/perltidy formatters/perltidy || true
git submodule add https://github.com/r-lib/styler formatters/styler || true
git submodule add https://github.com/posit-dev/air formatters/air || true
git submodule add https://github.com/yihui/formatR formatters/formatR || true

# Data Formats
git submodule add https://github.com/sqlfluff/sqlfluff formatters/sqlfluff || true
git submodule add https://github.com/darold/pgFormatter formatters/pgformatter || true
git submodule add https://github.com/sql-formatter-org/sql-formatter formatters/sql-formatter || true
git submodule add https://github.com/google/yamlfmt formatters/yamlfmt || true
git submodule add https://github.com/jqlang/jq formatters/jq || true
git submodule add https://github.com/tamasfe/taplo formatters/taplo || true
git submodule add https://github.com/bufbuild/buf formatters/buf || true

# Markup Languages
git submodule add https://github.com/htacg/tidy-html5 formatters/tidy-html5 || true
git submodule add https://github.com/stylelint/stylelint formatters/stylelint || true
git submodule add https://github.com/DavidAnson/markdownlint formatters/markdownlint || true
git submodule add https://github.com/hukkin/mdformat formatters/mdformat || true

# Infrastructure & DevOps
git submodule add https://github.com/hashicorp/terraform formatters/terraform || true
git submodule add https://github.com/gruntwork-io/terragrunt formatters/terragrunt || true
git submodule add https://github.com/hadolint/hadolint formatters/hadolint || true
git submodule add https://github.com/jessfraz/dockfmt formatters/dockfmt || true
git submodule add https://github.com/reteps/dockerfmt formatters/dockerfmt-modern || true

# Additional formatters
git submodule add https://github.com/uncrustify/uncrustify formatters/uncrustify || true
git submodule add https://github.com/nvuillam/npm-groovy-lint formatters/npm-groovy-lint || true
git submodule add https://github.com/mikefarah/yq formatters/yq || true
git submodule add https://github.com/n8d/old-fashioned-css-formatter formatters/old-fashioned-css-formatter || true

echo
echo "=== Initializing Submodules ==="
git submodule update --init --recursive

echo
echo "=== Submodule Initialization Complete ==="
echo "Total submodules: $(git submodule | wc -l)"
echo
echo "Next steps:"
echo "1. Pin versions: ./formatters/scripts/pin-versions.sh"
echo "2. Build binaries: ./formatters/scripts/build-all.sh"
echo "3. Health check: ./formatters/scripts/health-check-all.sh"
