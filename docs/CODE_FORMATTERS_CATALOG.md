# Complete Code Formatters Catalog

A comprehensive catalog of all major open-source code formatters for programming and scripting languages, organized by category with detailed metadata for integration planning.

**Last Updated**: 2026-01-29

---

## Table of Contents

- [Systems Languages](#systems-languages)
- [JVM Languages](#jvm-languages)
- [Web Languages](#web-languages)
- [Functional Languages](#functional-languages)
- [Mobile Development](#mobile-development)
- [Scripting Languages](#scripting-languages)
- [Data Formats](#data-formats)
- [Markup Languages](#markup-languages)
- [Infrastructure & DevOps](#infrastructure--devops)
- [Unified Formatters](#unified-formatters)
- [Summary Tables](#summary-tables)

---

## Systems Languages

### C / C++

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **clang-format** ⭐ | [llvm/llvm-project](https://github.com/llvm/llvm-project) | C, C++, Java, JavaScript, ObjectiveC, Protobuf | Standalone binary | Binary, apt/brew | YAML (.clang-format) | Fast | Easy | Apache 2.0 |
| **Artistic Style (astyle)** | [ArtisticStyle/ArtisticStyle](https://astyle.sourceforge.net/) | C, C++, C++/CLI, C#, ObjectiveC, Java | Standalone binary | Binary, apt/brew | CLI flags or config file | Fast | Easy | LGPL |
| **uncrustify** | [uncrustify/uncrustify](https://github.com/uncrustify/uncrustify) | C, C++, C#, ObjectiveC, D, Java, Pawn, VALA | Standalone binary | Binary, apt/brew | Custom config file | Medium | Medium | GPL 2.0 |

**Most Popular**: clang-format (official LLVM formatter, used by Google, comprehensive options)

**Key Differences**:
- clang-format: Uses GHC's parser, production-ready, comprehensive
- astyle: Highly customizable, multiple brace styles
- uncrustify: Most configurable options, can insert missing braces

### Rust

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **rustfmt** ⭐ | [rust-lang/rustfmt](https://github.com/rust-lang/rustfmt) | Rust | Standalone binary | cargo install | TOML (rustfmt.toml) | Fast | Easy | Apache 2.0 |

**Most Popular**: rustfmt (official Rust formatter, pretty-printer with reformatting)

**Key Features**: Comprehensive pretty-printer that reformats code extensively (unlike gofmt which only fixes whitespace)

### Go

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **gofmt** ⭐ | [golang/go](https://github.com/golang/go) | Go | Standalone binary (built-in) | Included with Go | None (opinionated) | Fast | Easy | BSD-3-Clause |
| **goimports** | [golang/tools](https://golang.org/x/tools/cmd/goimports) | Go | Standalone binary | go install | None | Fast | Easy | BSD-3-Clause |

**Most Popular**: gofmt (official Go formatter, pattern-based whitespace fixer)

**Key Differences**:
- gofmt: Whitespace-only, doesn't change line breaks
- goimports: Adds/removes imports, includes gofmt functionality

### Zig

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **zig fmt** ⭐ | [ziglang/zig](https://github.com/ziglang/zig) | Zig | Standalone binary (built-in) | Included with Zig | None (opinionated) | Fast | Easy | MIT |

**Most Popular**: zig fmt (official formatter, generates output from AST)

### Nim

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **nimpretty** ⭐ | [nim-lang/Nim](https://github.com/nim-lang/Nim) | Nim | Standalone binary (built-in) | Included with Nim | None | Medium | Easy | MIT |
| **nimpretty_t** | [ttytm/nimpretty_t](https://github.com/ttytm/nimpretty_t) | Nim | Wrapper | Binary | Config file | Medium | Easy | MIT |

**Most Popular**: nimpretty (official Nim formatter)

---

## JVM Languages

### Java

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **google-java-format** ⭐ | [google/google-java-format](https://github.com/google/google-java-format) | Java | Library/Binary | Binary, Maven, Gradle | None (opinionated) | Fast | Easy | Apache 2.0 |
| **Spotless** | [diffplug/spotless](https://github.com/diffplug/spotless) | Multi-language | Build plugin | Maven/Gradle plugin | Gradle/Maven config | Medium | Medium | Apache 2.0 |
| **Eclipse JDT** | [eclipse-jdt/eclipse.jdt.core](https://github.com/eclipse-jdt/eclipse.jdt.core) | Java | IDE-based | Eclipse plugin | Eclipse prefs | Medium | Medium | EPL 2.0 |

**Most Popular**: google-java-format (Google's standard, no configuration, Java 17+ required)

**Key Features**:
- google-java-format: Opinionated, zero config, enforces single format
- Spotless: Wrapper supporting multiple formatters (Google, Prettier, Eclipse)

### Kotlin

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **ktlint** ⭐ | [pinterest/ktlint](https://github.com/pinterest/ktlint) | Kotlin | Standalone binary | Binary, Gradle | .editorconfig | Fast | Easy | MIT |
| **ktfmt** | [facebook/ktfmt](https://github.com/facebook/ktfmt) | Kotlin | Library/Binary | Binary, Maven, Gradle | None (opinionated) | Fast | Easy | Apache 2.0 |
| **Spotless** | [diffplug/spotless](https://github.com/diffplug/spotless) | Multi-language | Build plugin | Maven/Gradle plugin | Gradle/Maven config | Medium | Medium | Apache 2.0 |

**Most Popular**: ktlint (Pinterest standard, customizable via .editorconfig)

**Key Differences**:
- ktlint: Customizable, style checker + formatter
- ktfmt: Based on google-java-format, opinionated, no configuration (JDK 11+)

### Scala

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **scalafmt** ⭐ | [scalameta/scalafmt](https://github.com/scalameta/scalafmt) | Scala | Standalone binary | Binary, sbt, Maven, Gradle | HOCON (.scalafmt.conf) | Fast | Easy | Apache 2.0 |
| **scalariform** | [scala-ide/scalariform](https://github.com/scala-ide/scalariform) | Scala | Library | sbt plugin | Scala code | Medium | Medium | MIT |

**Most Popular**: scalafmt (active development, Scala 3 support, version 3.10.3+ as of Jan 2026)

**Key Differences**:
- scalafmt: Modern, adds/removes newlines for consistency, Metals/IntelliJ integration
- scalariform: Older, preserves line breaks, less opinionated

### Groovy

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **npm-groovy-lint** ⭐ | [nvuillam/npm-groovy-lint](https://github.com/nvuillam/npm-groovy-lint) | Groovy, Jenkinsfile, Gradle | Node.js tool | npm install | Config file | Medium | Medium | GPL 3.0 |
| **Spotless (greclipse)** | [diffplug/spotless](https://github.com/diffplug/spotless) | Groovy | Build plugin | Maven/Gradle plugin | Gradle/Maven config | Medium | Hard | Apache 2.0 |

**Most Popular**: npm-groovy-lint (based on CodeNarc, better than Eclipse formatter)

**Note**: Requires Node.js 12+ and Java 17, has issues with Spock tests

### Clojure

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **cljfmt** ⭐ | [weavejester/cljfmt](https://github.com/weavejester/cljfmt) | Clojure, ClojureScript | Library/Binary | Leiningen/CLI | .cljfmt config | Fast | Easy | EPL 1.0 |
| **zprint** | [kkinnear/zprint](https://github.com/kkinnear/zprint) | Clojure, ClojureScript | Library/Binary | Binary, babashka | EDN config | Fast | Easy | MIT |

**Most Popular**: cljfmt (Clojure Style Guide defaults, Calva uses it)

**Key Differences**:
- cljfmt: Lightweight, preserves structure
- zprint: More aggressive reformatting, fast babashka support, 3x faster on Apple Silicon

---

## Web Languages

### JavaScript / TypeScript

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **Prettier** ⭐ | [prettier/prettier](https://github.com/prettier/prettier) | JS, TS, JSX, TSX, JSON, HTML, CSS, GraphQL, Markdown | Node.js tool | npm install | JSON, YAML, TOML | Medium | Easy | MIT |
| **Biome** | [biomejs/biome](https://github.com/biomejs/biome) | JS, TS, JSX, TSX, JSON, HTML, CSS, GraphQL | Standalone binary (Rust) | Binary, npm | JSON (biome.json) | Very Fast | Easy | MIT |
| **dprint** | [dprint/dprint](https://github.com/dprint/dprint) | Multi-language (pluggable) | Standalone binary (Rust) | Binary | JSON (dprint.json) | Very Fast | Easy | MIT |

**Most Popular**: Prettier (de facto standard, 97% compatible with Biome)

**Key Differences**:
- Prettier: Established standard, slower, extensive plugin ecosystem
- Biome: 97% Prettier-compatible, significantly faster (Rust), includes linter (340+ rules)
- dprint: Pluggable platform, can wrap Biome/Prettier/Oxc

**Performance**: Biome > dprint > Prettier (Biome claims 35x faster than Prettier)

### PHP

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **PHP-CS-Fixer** ⭐ | [PHP-CS-Fixer/PHP-CS-Fixer](https://github.com/PHP-CS-Fixer/PHP-CS-Fixer) | PHP | PHP library | composer | PHP config file | Medium | Medium | MIT |
| **Laravel Pint** | [laravel/pint](https://github.com/laravel/pint) | PHP | PHP library (built on PHP-CS-Fixer) | composer | JSON (pint.json) | Medium | Easy | MIT |

**Most Popular**: Laravel Pint (for Laravel projects, opinionated), PHP-CS-Fixer (general PHP)

**Key Features**:
- Laravel Pint: Built on PHP-CS-Fixer, opinionated minimalist style for Laravel
- PHP-CS-Fixer: Full-featured, all PHP CS Fixer rules available

### Python

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **Black** ⭐ | [psf/black](https://github.com/psf/black) | Python | Python package | pip install | TOML (pyproject.toml) | Medium | Easy | MIT |
| **Ruff** | [astral-sh/ruff](https://github.com/astral-sh/ruff) | Python | Standalone binary (Rust) | Binary, pip | TOML (pyproject.toml) | Very Fast | Easy | MIT |
| **autopep8** | [hhatto/autopep8](https://github.com/hhatto/autopep8) | Python | Python package | pip install | CLI flags | Medium | Easy | MIT |
| **YAPF** | [google/yapf](https://github.com/google/yapf) | Python | Python package | pip install | Config file | Slow | Easy | Apache 2.0 |

**Most Popular**: Black (uncompromising formatter, PEP 8 aligned, 2026 stable style released Jan 2026)

**Key Differences**:
- Black: Opinionated, minimal config, "stable style" released Jan 2026
- Ruff: 30x+ faster than Black (Rust), >99.9% Black-compatible, drop-in replacement
- autopep8: Preserves input style, only formats non-compliant code
- YAPF: Google's tool, highly configurable (pep8, Google, Facebook, Chromium styles)

**Performance**: Ruff (30x faster) > Black > autopep8 > YAPF (100x slower than Ruff)

### Ruby

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **RuboCop** ⭐ | [rubocop/rubocop](https://github.com/rubocop/rubocop) | Ruby | Ruby gem | gem install | YAML (.rubocop.yml) | Slow (2-5s for 3000 lines) | Easy | MIT |
| **StandardRB** | [standardrb/standard](https://github.com/standardrb/standard) | Ruby | Ruby gem (built on RuboCop) | gem install | None (opinionated) | Medium | Easy | MIT |
| **Rufo** | [ruby-formatter/rufo](https://github.com/ruby-formatter/rufo) | Ruby | Ruby gem | gem install | Minimal config | Fast (<1s for 3000 lines) | Easy | MIT |

**Most Popular**: RuboCop (comprehensive linter + formatter, Ruby Style Guide)

**Key Differences**:
- RuboCop: Full-featured, highly configurable, slower
- StandardRB: Opinionated RuboCop config, no decisions, monthly updates
- Rufo: Real formatter (not find-replace), significantly faster, minimal config, Ruby 3.0+

---

## Functional Languages

### Haskell

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **Ormolu** ⭐ | [tweag/ormolu](https://github.com/tweag/ormolu) | Haskell | Standalone binary | cabal/stack | None (opinionated) | Fast | Easy | BSD-3-Clause |
| **Fourmolu** | [fourmolu/fourmolu](https://github.com/fourmolu/fourmolu) | Haskell | Standalone binary | cabal/stack | YAML (.fourmolu.yaml) | Fast | Easy | BSD-3-Clause |
| **stylish-haskell** | [haskell/stylish-haskell](https://github.com/haskell/stylish-haskell) | Haskell | Standalone binary | cabal/stack | YAML | Medium | Easy | BSD-3-Clause |
| **hindent** | [mihaimaruseac/hindent](https://github.com/mihaimaruseac/hindent) | Haskell | Standalone binary | cabal/stack | Config file | Medium | Medium | BSD-3-Clause |

**Most Popular**: Ormolu (default in Haskell Language Server, uses GHC parser)

**Key Differences**:
- Ormolu: Opinionated, zero config, uses GHC parser (no parsing bugs)
- Fourmolu: Fork of Ormolu, configurable, 4-space indent, continually merges Ormolu
- stylish-haskell/hindent: Use haskell-src-exts (parsing bugs)

### OCaml

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **ocamlformat** ⭐ | [ocaml-ppx/ocamlformat](https://github.com/ocaml-ppx/ocamlformat) | OCaml | Standalone binary | opam install | Config file (.ocamlformat) | Fast | Easy | MIT |

**Most Popular**: ocamlformat (official, version 0.26.2+, OCaml 5.3 support)

**Key Features**: Opinionated defaults, fully customizable, formats comments/docstrings, RPC server

### F#

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **Fantomas** ⭐ | [fsprojects/fantomas](https://github.com/fsprojects/fantomas) | F# | .NET tool | dotnet tool | Config file | Medium | Easy | Apache 2.0 |

**Most Popular**: Fantomas (de facto F# formatter)

**Key Features**: Ensures correct indentation and consistent spacing

### Elixir

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **mix format** ⭐ | [elixir-lang/elixir](https://github.com/elixir-lang/elixir) | Elixir | Built-in (Elixir) | Included with Elixir | Elixir code (.formatter.exs) | Fast | Easy | Apache 2.0 |

**Most Popular**: mix format (official built-in formatter)

**Key Features**: Tidyverse style guide, reads .formatter.exs configuration

### Erlang

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **erlfmt** ⭐ | [WhatsApp/erlfmt](https://github.com/WhatsApp/erlfmt) | Erlang | Standalone binary | rebar3 plugin | None (opinionated) | Fast | Easy | Apache 2.0 |

**Most Popular**: erlfmt (WhatsApp's opinionated formatter, version 1.6.0 Jan 2025)

**Key Features**: Enforces max line length, rebar3 fmt task

---

## Mobile Development

### Swift

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **swift-format** ⭐ | [swiftlang/swift-format](https://github.com/swiftlang/swift-format) | Swift | Standalone binary | Included with Swift 6+/Xcode 16 | JSON (.swift-format) | Fast | Easy | Apache 2.0 |
| **SwiftFormat** | [nicklockwood/SwiftFormat](https://github.com/nicklockwood/SwiftFormat) | Swift | Standalone binary | Binary, brew | Config file | Fast | Easy | MIT |

**Most Popular**: swift-format (Apple official, included in Xcode 16, Swift 6+)

**Key Differences**:
- swift-format: Official Apple tool, use `swift format` (with space)
- SwiftFormat: Community tool by Nick Lockwood, designed as fixer not linter

### Dart (Flutter)

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **dart format** ⭐ | [dart-lang/dart_style](https://github.com/dart-lang/dart_style) | Dart | Built-in (Dart SDK) | Included with Dart SDK | analysis_options.yaml | Fast | Easy | BSD-3-Clause |

**Most Popular**: dart format (official, replaced dartfmt, part of unified dart tool)

**Key Features**:
- Dart 3.7 (Feb 2025): New formatter with trailing comma options
- Dart 3.8 (Flutter 3.32): Preserve trailing commas option
- Configurable page width in analysis_options.yaml

### Objective-C

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **clang-format** ⭐ | [llvm/llvm-project](https://github.com/llvm/llvm-project) | C, C++, Java, JavaScript, ObjectiveC, Protobuf | Standalone binary | Binary, apt/brew | YAML (.clang-format) | Fast | Easy | Apache 2.0 |
| **uncrustify** | [uncrustify/uncrustify](https://github.com/uncrustify/uncrustify) | C, C++, C#, ObjectiveC, D, Java | Standalone binary | Binary, apt/brew | Custom config file | Medium | Medium | GPL 2.0 |

**Most Popular**: clang-format (official LLVM formatter)

**Note**: Uncrustify more configurable for Objective-C, but clang-format has fewer issues

---

## Scripting Languages

### Bash / Shell

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **shfmt** ⭐ | [mvdan/sh](https://github.com/mvdan/sh) | POSIX, Bash, mksh | Standalone binary (Go) | Binary, brew | .editorconfig or CLI flags | Fast | Easy | BSD-3-Clause |
| **beautysh** | [lovesegfault/beautysh](https://github.com/lovesegfault/beautysh) | Bash | Python package | pip install | CLI flags | Medium | Easy | MIT |

**Most Popular**: shfmt (parser + formatter + interpreter, supports POSIX/Bash/mksh)

**Key Differences**:
- shfmt: Fast (Go), operates on .sh/.bash files, EditorConfig support, Google Style Guide (`-i 2 -ci`)
- beautysh: Python-based, @formatter:off/on comments, customizable

### PowerShell

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **PSScriptAnalyzer** ⭐ | [PowerShell/PSScriptAnalyzer](https://github.com/PowerShell/PSScriptAnalyzer) | PowerShell | PowerShell module | PowerShell Gallery | PSD1 config | Medium | Easy | MIT |

**Most Popular**: PSScriptAnalyzer (official Microsoft tool, linter + formatter)

**Key Features**: Invoke-Formatter cmdlet, VS Code integration, PowerShell best practices

### Lua

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **StyLua** ⭐ | [JohnnyMorganz/StyLua](https://github.com/JohnnyMorganz/StyLua) | Lua 5.1-5.4, LuaJIT, Luau, FiveM | Standalone binary (Rust) | Binary, cargo | TOML (.stylua.toml) | Fast | Easy | MPL 2.0 |
| **luafmt** | [appgurueu/luafmt](https://github.com/appgurueu/luafmt) | Lua 5.1 | Lua script | Copy script | Minimal | Fast | Easy | MIT |

**Most Popular**: StyLua (Prettier-inspired, Roblox Lua Style Guide, LSP support)

**Key Features**: Deterministic, supports multiple Lua versions, VS Code extension, language server

### Perl

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **perltidy** ⭐ | [perltidy/perltidy](https://github.com/perltidy/perltidy) | Perl | Perl script | cpan install | .perltidyrc | Medium | Easy | GPL 2.0 |

**Most Popular**: perltidy (official, version 20260109.01 released Jan 21, 2026)

**Key Features**: perlstyle(1) defaults, highly customizable, HTML output support

### R

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **styler** ⭐ | [r-lib/styler](https://github.com/r-lib/styler) | R | R package | R install | R code | Slow | Easy | MIT |
| **Air** | [posit-dev/air](https://github.com/posit-dev/air) | R | Standalone binary (Rust) | Binary | Minimal config | Very Fast (300x) | Easy | MIT |
| **formatR** | [yihui/formatR](https://github.com/yihui/formatR) | R | R package | R install | R code | Medium | Easy | GPL 2.0 |

**Most Popular**: styler (tidyverse style guide, rich features)

**Key Differences**:
- styler: Established, flexible, tidyverse syntax support (pipes %>%)
- Air: NEW (Feb 2025), 300x+ faster than styler, Rust-based, minimal config
- formatR: Older, basic formatting (spaces, indentation, assignment operators)

**Note**: Air is a game-changer for R formatting performance (2025 release)

---

## Data Formats

### SQL

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **SQLFluff** ⭐ | [sqlfluff/sqlfluff](https://github.com/sqlfluff/sqlfluff) | SQL (multi-dialect), dbt/Jinja | Python package | pip install | TOML/INI (.sqlfluff) | Medium | Easy | MIT |
| **pgFormatter** | [darold/pgFormatter](https://github.com/darold/pgFormatter) | PostgreSQL SQL | Perl script | cpan/binary | Config file | Medium | Easy | BSD-3-Clause |
| **sql-formatter** | [sql-formatter-org/sql-formatter](https://github.com/sql-formatter-org/sql-formatter) | SQL (multi-dialect) | Node.js library | npm install | JSON config | Fast | Easy | MIT |

**Most Popular**: SQLFluff (dialect-flexible, linter + formatter, dbt support)

**Key Differences**:
- SQLFluff: Most feature-rich, supports 20+ SQL dialects, Jinja templating, monthly releases
- pgFormatter: PostgreSQL-focused, SQL-92 to SQL-2011 keywords, CGI/console
- sql-formatter: Lightweight, multi-dialect, fast

### YAML

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **yamlfmt** ⭐ | [google/yamlfmt](https://github.com/google/yamlfmt) | YAML | Standalone binary (Go) | Binary | YAML config | Fast | Easy | Apache 2.0 |
| **Prettier** | [prettier/prettier](https://github.com/prettier/prettier) | YAML (+ many others) | Node.js tool | npm install | JSON/YAML config | Medium | Easy | MIT |
| **yq** | [mikefarah/yq](https://github.com/mikefarah/yq) | YAML, JSON, XML, CSV, TOML | Standalone binary (Go) | Binary | None | Fast | Easy | MIT |

**Most Popular**: yamlfmt (Google, opinionated, highly configurable, no npm/node required)

**Key Differences**:
- yamlfmt: Standalone Go binary, fast, opinionated
- Prettier: Multi-language, opinionated, few options
- yq: Processor not formatter, jq-like for YAML

### JSON

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **jq** ⭐ | [jqlang/jq](https://github.com/jqlang/jq) | JSON | Standalone binary | Binary, apt/brew | None | Fast | Easy | MIT |
| **Prettier** | [prettier/prettier](https://github.com/prettier/prettier) | JSON (+ many others) | Node.js tool | npm install | JSON/YAML config | Medium | Easy | MIT |

**Most Popular**: jq (sed for JSON, de facto standard for JSON processing)

### TOML

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **Taplo** ⭐ | [tamasfe/taplo](https://github.com/tamasfe/taplo) | TOML | Standalone binary (Rust) | Binary, cargo | TOML config | Fast | Easy | MIT |
| **prettier-plugin-toml** | [un-ts/prettier](https://github.com/un-ts/prettier) | TOML | Prettier plugin | npm install | Prettier config | Medium | Easy | MIT |

**Most Popular**: Taplo (versatile TOML toolkit, validator + formatter)

**Note**: prettier-plugin-toml wraps Taplo's printer

### XML

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **xmllint** ⭐ | [GNOME/libxml2](https://gitlab.gnome.org/GNOME/libxml2) | XML | Standalone binary (libxml2) | apt/brew (built-in) | Env vars | Fast | Easy | MIT |
| **tidy** | [htacg/tidy-html5](https://github.com/htacg/tidy-html5) | HTML, XML | Standalone binary | Binary, apt/brew | Config file | Medium | Easy | MIT |
| **xmlstarlet** | [xpath/xmlstarlet](https://sourceforge.net/projects/xmlstar/) | XML | Standalone binary | apt/brew | CLI flags | Fast | Medium | MIT |

**Most Popular**: xmllint (libxml2, available on most systems, simple)

**Key Differences**:
- xmllint: Simple, reliable, XMLLINT_INDENT env var for indent size
- tidy: Also fixes errors, cleans formatting
- xmlstarlet: Powerful but more complex

### GraphQL

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **Prettier** ⭐ | [prettier/prettier](https://github.com/prettier/prettier) | GraphQL (+ many others) | Node.js tool | npm install | JSON/YAML config | Medium | Easy | MIT |

**Most Popular**: Prettier (official GraphQL support since v1.5.0)

**Key Features**: Auto-detects .graphql/.gql files, /* GraphQL */ comment tags for template literals

### Protobuf

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **buf format** ⭐ | [bufbuild/buf](https://github.com/bufbuild/buf) | Protobuf | Standalone binary (Go) | Binary, brew | None (opinionated) | Fast | Easy | Apache 2.0 |
| **clang-format** | [llvm/llvm-project](https://github.com/llvm/llvm-project) | Protobuf (+ C/C++/Java/etc.) | Standalone binary | Binary, apt/brew | YAML (.clang-format) | Fast | Easy | Apache 2.0 |

**Most Popular**: buf format (official Buf CLI, zero config, one way to format .proto)

**Key Features**: buf format --write (rewrite in place), --diff (show diff), no config options

---

## Markup Languages

### HTML

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **Prettier** ⭐ | [prettier/prettier](https://github.com/prettier/prettier) | HTML (+ many others) | Node.js tool | npm install | JSON/YAML config | Medium | Easy | MIT |
| **tidy** | [htacg/tidy-html5](https://github.com/htacg/tidy-html5) | HTML, XML | Standalone binary | Binary, apt/brew | Config file | Medium | Easy | MIT |

**Most Popular**: Prettier (de facto standard for web development)

### CSS / SCSS / SASS

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **Prettier** ⭐ | [prettier/prettier](https://github.com/prettier/prettier) | CSS, SCSS, SASS, Less | Node.js tool | npm install | JSON/YAML config | Medium | Easy | MIT |
| **Stylelint** | [stylelint/stylelint](https://github.com/stylelint/stylelint) | CSS, SCSS, SASS, Less | Node.js tool | npm install | JSON config | Medium | Easy | MIT |
| **Old Fashioned CSS Formatter** | [n8d/old-fashioned-css-formatter](https://github.com/n8d/old-fashioned-css-formatter) | CSS | Node.js tool | npm install | Config file | Fast | Easy | MIT |

**Most Popular**: Prettier (formatting), Stylelint (linting, deprecated stylistic rules in v16)

**Key Differences**:
- Prettier: Whitespace/indentation, no property ordering
- Stylelint: Linter (rules enforcement), complements Prettier (use together)
- Old Fashioned CSS Formatter: NEW (2025), CSSComb successor, property ordering (concentric/idiomatic/alphabetical)

**Best Practice**: Use Prettier + Stylelint together (complementary tools)

### Markdown

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **Prettier** ⭐ | [prettier/prettier](https://github.com/prettier/prettier) | Markdown (+ many others) | Node.js tool | npm install | JSON/YAML config | Medium | Easy | MIT |
| **markdownlint** | [DavidAnson/markdownlint](https://github.com/DavidAnson/markdownlint) | Markdown | Node.js tool (linter) | npm install | JSON config | Fast | Easy | MIT |
| **mdformat** | [hukkin/mdformat](https://github.com/hukkin/mdformat) | Markdown | Python package | pip install | TOML config | Fast | Easy | MIT |

**Most Popular**: Prettier (markdown support since v1.8, Nov 2017)

**Key Differences**:
- Prettier: Formatter (whitespace, structure)
- markdownlint: Linter (style rules), compatible with Prettier defaults
- mdformat: CommonMark compliant, fixes Prettier's AST/HTML bugs

**Best Practice**: Use Prettier (formatter) + markdownlint (linter) together

---

## Infrastructure & DevOps

### Terraform / HCL

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **terraform fmt** ⭐ | [hashicorp/terraform](https://github.com/hashicorp/terraform) | HCL (Terraform) | Built-in (Go) | Included with Terraform | None (opinionated) | Fast | Easy | MPL 2.0 |
| **terragrunt hclfmt** | [gruntwork-io/terragrunt](https://github.com/gruntwork-io/terragrunt) | HCL (Terragrunt) | Built-in | Included with Terragrunt | None | Fast | Easy | MIT |

**Most Popular**: terraform fmt (official, canonical HCL style, intentionally opinionated)

**Key Features**:
- --check (verify without rewriting)
- -recursive (process subdirectories)
- Applies to .tf and .tfvars files
- Based on Terraform code style
- Zero configuration by design

### Dockerfile

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **hadolint** ⭐ | [hadolint/hadolint](https://github.com/hadolint/hadolint) | Dockerfile | Standalone binary (Haskell) | Binary, Docker | YAML config | Fast | Easy | GPL 3.0 |
| **dockfmt** | [jessfraz/dockfmt](https://github.com/jessfraz/dockfmt) | Dockerfile | Standalone binary (Go) | Binary | None | Fast | Easy | MIT |
| **dockerfmt (modern)** | [reteps/dockerfmt](https://github.com/reteps/dockerfmt) | Dockerfile | Standalone binary (Go) | Binary | None | Fast | Easy | MIT |

**Most Popular**: hadolint (linter, most popular, uses Shellcheck for RUN instructions)

**Key Differences**:
- hadolint: **Linter** (checks best practices, security, errors)
- dockfmt: **Formatter** (syntax formatting, like gofmt)
- Modern dockerfmt: Uses moby/buildkit parser, mvdan/sh for shell formatting

---

## Unified Formatters

These formatters support multiple languages through a single tool:

| Formatter | GitHub URL | Languages | Architecture | Installation | Config Format | Performance | Integration | License |
|-----------|------------|-----------|--------------|--------------|---------------|-------------|-------------|---------|
| **Prettier** ⭐ | [prettier/prettier](https://github.com/prettier/prettier) | JS, TS, JSX, TSX, JSON, HTML, CSS, SCSS, GraphQL, Markdown, YAML | Node.js tool | npm install | JSON, YAML, TOML | Medium | Easy | MIT |
| **dprint** | [dprint/dprint](https://github.com/dprint/dprint) | Pluggable (TS, JSON, Markdown, TOML, Dockerfile, Prettier, Biome, etc.) | Standalone binary (Rust) | Binary | JSON (dprint.json) | Very Fast | Easy | MIT |
| **EditorConfig** | [editorconfig/editorconfig](https://github.com/editorconfig/editorconfig) | All (basic formatting) | Editor plugin | Per-editor | INI (.editorconfig) | N/A | Easy | BSD/MIT |

**Most Popular**: Prettier (de facto standard for web development)

**Key Differences**:
- **Prettier**: Opinionated, extensive language support, slower (JavaScript)
- **dprint**: Pluggable platform (wraps formatters), very fast (Rust), flexible
- **EditorConfig**: NOT a formatter, defines basic editor settings (indent, EOL, charset)

**Integration Strategy**: EditorConfig + Prettier work together
- EditorConfig: Basic file-level settings (all editors understand)
- Prettier: Advanced code formatting (reads .editorconfig, can be overridden by .prettierrc)

---

## Summary Tables

### By Performance

| Performance Tier | Formatters |
|-----------------|------------|
| **Very Fast (Rust/Go)** | Ruff, Biome, dprint, StyLua, Air (R), shfmt, yamlfmt, Taplo, buf |
| **Fast** | gofmt, goimports, rustfmt, clang-format, zig fmt, google-java-format, ktlint, ktfmt, scalafmt, swift-format, dart format, xmllint, jq |
| **Medium** | Black, autopep8, RuboCop, StandardRB, Rufo, PHP-CS-Fixer, Prettier, perltidy, formatR, PSScriptAnalyzer, styler |
| **Slow** | YAPF, RuboCop (large files), styler |

### By Installation Method

| Method | Formatters |
|--------|------------|
| **Built-in** | gofmt, goimports, zig fmt, nimpretty, mix format, dart format, terraform fmt, swift-format (Swift 6+/Xcode 16) |
| **Binary** | clang-format, rustfmt, shfmt, buf, hadolint, dockfmt, Taplo, yamlfmt, jq, xmllint, Ruff, Biome, dprint, StyLua, Air |
| **Language Package Manager** | Black, autopep8, YAPF, Ruff (pip), RuboCop, Rufo (gem), perltidy (cpan), styler, formatR (R), erlfmt (rebar3) |
| **npm** | Prettier, Biome (also binary), dprint (also binary), npm-groovy-lint, sql-formatter |
| **Build Tool Plugin** | Spotless (Maven/Gradle), google-java-format (Maven/Gradle) |
| **Multiple Methods** | Most modern tools support binary + package manager |

### By Configuration Philosophy

| Philosophy | Formatters |
|------------|------------|
| **Zero Config (Opinionated)** | gofmt, goimports, rustfmt, zig fmt, Black, StandardRB, google-java-format, ktfmt, terraform fmt, buf, swift-format (Apple), erlfmt |
| **Minimal Config** | Prettier, Biome, StyLua, Rufo, dockfmt |
| **Highly Configurable** | clang-format, uncrustify, RuboCop, YAPF, Spotless, ktlint, scalafmt, Fourmolu, ocamlformat, yamlfmt, Taplo |

### By License

| License | Formatters |
|---------|------------|
| **MIT** | Prettier, Biome, Black, Ruff, autopep8, ktlint, RuboCop, StandardRB, Rufo, PHP-CS-Fixer, Laravel Pint, scalariform, SwiftFormat, StyLua, beautysh, PSScriptAnalyzer, styler, Air, SQLFluff, Taplo, jq, xmllint, tidy, dockfmt, markdownlint, mdformat, terragrunt |
| **Apache 2.0** | clang-format, google-java-format, Spotless, ktfmt, scalafmt, mix format, erlfmt, swift-format, ocamlformat, Fantomas, yamlfmt, buf |
| **BSD-3-Clause** | gofmt, goimports, Ormolu, Fourmolu, dart format, shfmt, pgFormatter |
| **GPL** | uncrustify (GPL 2.0), hadolint (GPL 3.0), npm-groovy-lint (GPL 3.0), perltidy (GPL 2.0), formatR (GPL 2.0) |
| **MPL 2.0** | StyLua, terraform fmt |
| **EPL** | cljfmt (EPL 1.0), Eclipse JDT (EPL 2.0) |

### Recommended Formatters by Use Case

| Use Case | Recommended Formatter | Alternative |
|----------|----------------------|-------------|
| **Web Full-Stack** | Prettier | Biome (faster) |
| **High Performance JS/TS** | Biome | dprint |
| **Python Modern** | Ruff | Black |
| **Python Traditional** | Black | autopep8 |
| **Systems Programming** | clang-format (C/C++), rustfmt (Rust) | - |
| **JVM Polyglot** | Spotless | Language-specific |
| **DevOps/IaC** | terraform fmt, yamlfmt, buf | Prettier |
| **Mobile (iOS)** | swift-format (Apple) | SwiftFormat |
| **Mobile (Flutter)** | dart format | - |
| **Shell Scripting** | shfmt | beautysh |
| **Data Engineering** | SQLFluff (SQL), yq (YAML/JSON) | - |
| **Multi-Language Project** | dprint (pluggable) | Prettier + language-specific |

---

## Integration Complexity Assessment

### Easy Integration (Drop-in)
- All built-in formatters (gofmt, zig fmt, mix format, dart format, terraform fmt)
- Single binary with zero config (rustfmt, Black, Prettier, yamlfmt, buf)
- Language-native tools (RuboCop, perltidy, styler)

### Medium Integration (Configuration Required)
- Tools needing config files (clang-format, ktlint, scalafmt, Spotless)
- Build tool plugins (Maven, Gradle)
- Tools with dependencies (npm-groovy-lint needs Node + Java)

### Hard Integration (Complex Setup)
- IDE-specific formatters (Eclipse JDT)
- Tools requiring wrapper plugins (Spotless with multiple formatters)
- Tools with native binary compilation requirements

---

## Sources

### Systems Languages
- [GitHub - awesome-code-formatters](https://github.com/rishirdua/awesome-code-formatters)
- [Clang-Format Style Options](https://clang.llvm.org/docs/ClangFormatStyleOptions.html)
- [GitHub - rust-lang/rustfmt](https://github.com/rust-lang/rustfmt)
- [Rust vs Go in 2026 — Bitfield Consulting](https://bitfieldconsulting.com/posts/rust-vs-go)

### JVM Languages
- [GitHub - google/google-java-format](https://github.com/google/google-java-format)
- [GitHub - diffplug/spotless](https://github.com/diffplug/spotless)
- [Keep Your Kotlin Code Spotless](https://dev.to/rockandnull/keep-your-kotlin-code-spotless-a-guide-to-ktlint-and-ktfmt-linters-4a8o)
- [GitHub - facebook/ktfmt](https://github.com/facebook/ktfmt)
- [Scalafmt · Code formatter for Scala](http://scalameta.org/scalafmt/)
- [GitHub - scalameta/scalafmt](https://github.com/scalameta/scalafmt)

### Web Languages
- [Biome, toolchain of the web](https://biomejs.dev/)
- [GitHub - dprint/dprint](https://github.com/dprint/dprint)
- [The Ruff Formatter](https://astral.sh/blog/the-ruff-formatter)
- [Python code formatters comparison](https://blog.frank-mich.com/python-code-formatters-comparison-black-autopep8-and-yapf/)
- [black · PyPI](https://pypi.org/project/black/)
- [Laravel Pint - Laravel 12.x](https://laravel.com/docs/12.x/pint)
- [GitHub - laravel/pint](https://github.com/laravel/pint)
- [GitHub - standardrb/standard](https://github.com/standardrb/standard)
- [GitHub - rubocop/rubocop](https://github.com/rubocop/rubocop)

### Functional Languages
- [GitHub - tweag/ormolu](https://github.com/tweag/ormolu)
- [Fourmolu](https://fourmolu.github.io/)
- [GitHub - ocaml-ppx/ocamlformat](https://github.com/ocaml-ppx/ocamlformat)
- [GitHub - WhatsApp/erlfmt](https://github.com/WhatsApp/erlfmt)
- [mix format — Mix v1.20.0-dev](https://hexdocs.pm/mix/main/Mix.Tasks.Format.html)

### Mobile Development
- [GitHub - swiftlang/swift-format](https://github.com/swiftlang/swift-format)
- [GitHub - nicklockwood/SwiftFormat](https://github.com/nicklockwood/SwiftFormat)
- [dart format](https://dart.dev/tools/dart-format)
- [How to Configure the Updated Code Formatter in Dart 3.8](https://codewithandrea.com/articles/updated-formatter-dart-3-8/)

### Scripting Languages
- [GitHub - mvdan/sh](https://github.com/mvdan/sh)
- [GitHub - lovesegfault/beautysh](https://github.com/lovesegfault/beautysh)
- [PSScriptAnalyzer module](https://learn.microsoft.com/en-us/powershell/utility-modules/psscriptanalyzer/overview)
- [GitHub - JohnnyMorganz/StyLua](https://github.com/JohnnyMorganz/StyLua)
- [GitHub - perltidy/perltidy](https://github.com/perltidy/perltidy)
- [Air, an extremely fast R formatter](https://tidyverse.org/blog/2025/02/air/)
- [Non-Invasive Pretty Printing of R Code · styler](https://styler.r-lib.org/)

### Data Formats
- [GitHub - sqlfluff/sqlfluff](https://github.com/sqlfluff/sqlfluff)
- [GitHub - darold/pgFormatter](https://github.com/darold/pgFormatter)
- [GitHub - mikefarah/yq](https://github.com/mikefarah/yq)
- [A Detailed Comparison of YAML Formatters](https://xkyle.com/A-Detailed-Comparison-of-YAML-Formatters/)
- [Taplo | A versatile TOML toolkit](https://taplo.tamasfe.dev/)
- [Prettier TOML - Visual Studio Marketplace](https://marketplace.visualstudio.com/items?itemName=bodil.prettier-toml)
- [How to Pretty-Print XML From Command Line?](https://www.tutorialspoint.com/how-to-pretty-print-xml-from-command-line)
- [GraphQL just got a whole lot "Prettier"!](https://www.apollographql.com/blog/announcement/graphql-just-got-a-whole-lot-prettier-7701d4675f42/)
- [Formatting your Protobuf files - Buf Docs](https://buf.build/docs/format/)

### Markup Languages
- [Old Fashioned CSS Formatter](https://dev.to/stfbauer/old-fashioned-css-formatter-a-modern-successor-to-csscomb-538e)
- [Prettier + Stylelint](https://css-tricks.com/prettier-stylelint-writing-clean-css-keeping-clean-code-two-tool-game/)
- [Prettier 1.8: Markdown Support](https://prettier.io/blog/2017/11/07/1.8.0.html)
- [Configuring Markdownlint Alongside Prettier](https://www.joshuakgoldberg.com/blog/configuring-markdownlint-alongside-prettier/)
- [GitHub - hukkin/mdformat](https://github.com/hukkin/mdformat)

### Infrastructure & DevOps
- [terraform fmt command reference](https://developer.hashicorp.com/terraform/cli/commands/fmt)
- [Dockerfile Linter](https://hadolint.github.io/hadolint/)
- [GitHub - hadolint/hadolint](https://github.com/hadolint/hadolint)
- [GitHub - jessfraz/dockfmt](https://github.com/jessfraz/dockfmt)
- [GitHub - reteps/dockerfmt](https://github.com/reteps/dockerfmt)

### Unified Formatters
- [dprint - Code Formatter](https://dprint.dev/)
- [Configuration - dprint](https://dprint.dev/config/)
- [Why .editorconfig Still Matters Even with Prettier Around](https://dev.to/leapcell/why-editorconfig-still-matters-even-with-prettier-around-2cda)
- [Configuration File · Prettier](https://prettier.io/docs/configuration)

---

**Document Version**: 1.0
**Generated**: 2026-01-29
**Total Formatters Cataloged**: 80+
**Languages Covered**: 50+
