# Repository Guidelines

## Project Structure & Module Organization
This repository is a chapter-based Go tutorial for building AI agents.
- `ch01`-`ch09`: runnable chapter implementations (each chapter evolves capabilities).
- `ch10`: placeholder for upcoming web/SSE chapter.
- `shared/`: common utilities (config, env loading, shared types, MCP helpers).
- `openspec/`: OpenSpec configuration.
- `.github/workflows/test-build.yml`: CI checks used for PR validation.

Follow the existing chapter pattern (`prompt.go`, `agent.go`, `main/main.go`, optional `tool/`, `context/`, `memory/`, `storage/`) when adding new functionality.

## Build, Test, and Development Commands
Use Go 1.25+.
- `cp .env.example .env`: initialize local configuration.
- `go run ./ch01/main --stream -q "hello"`: run a chapter entrypoint.
- `go test -v ./ch08/...`: run tests for one chapter.
- `go test -v ./...`: run all tests in the module.
- `go build -v ./ch09/main`: build a chapter binary (CI builds chapter mains).
- `gofmt -l .`: check formatting (matches CI).
- `gofmt -w <paths>`: apply formatting before commit.

## Coding Style & Naming Conventions
- Follow standard Go formatting and idioms (`gofmt` is required; tabs for indentation).
- Keep package names short, lowercase, and domain-focused (`context`, `tool`, `memory`).
- Exported identifiers: `PascalCase`; internal helpers: `camelCase`.
- Prefer small, composable files over large mixed-responsibility files.

## Testing Guidelines
- Place tests in `*_test.go` next to the code under test.
- Current tests focus on context policy modules (e.g., `ch05/context/policy_test.go`).
- Use descriptive test names like `Test<Feature>_<Scenario>`.
- For changes, add or update tests in the same chapter/package and run `go test -v ./<chapter>/...`.

## Commit & Pull Request Guidelines
- Keep commits scoped and imperative (patterns in history: `implement ch08`, `refactor ...`, `fix ...`).
- Optional prefixes are acceptable when useful (`feat:`, `fix:`, `refactor:`).
- PRs should include:
  - what changed and why,
  - affected chapters/packages,
  - local verification commands/results,
  - screenshots or terminal snippets for TUI-visible changes.

## Security & Configuration Tips
- Never commit secrets; keep keys in `.env` only.
- Use `.env.example` as the template for required variables (`OPENAI_BASE_URL`, `OPENAI_API_KEY`, `OPENAI_MODEL`).

## Learning Notes Style
- Learning notes should be kept in standalone chapter files such as `ch01/LEARNING_QA.md`, not appended to chapter `README.md`.
- Each chapter note should start with a learning guide section, then a Q&A section.
- The learning guide should include:
  - what to learn in the chapter,
  - what to run or operate,
  - what knowledge should be mastered after finishing,
  - extension reading summary,
  - what to focus on in the extension reading.
- Extension reading should keep original source links and include additional takeaways from reading the source material, not just a restatement of the local code.
- Q&A entries should record the user's questions and the assistant's answers, but should not record code change logs.
- Each Q&A entry should use this structure:
  - `一句总结`
  - `详细回答`
- Question titles may be lightly normalized into study-note phrasing, but must preserve the original meaning.
- Especially important entries should use a `[重点]` prefix in the title. Do not rely on color styling.
- Formatting rules for learning notes:
  - Prefer a single-level bullet list with complete sentences.
  - Bullet subpoints are allowed only when the content is clearly an enumeration of parallel items.
  - Do not overuse nested bullets.
  - Bullet headings should use meaningful labels, not vague labels like “第一步” or “第二步”.
  - Bullet headings should be bold.
  - Bullet headings and content should not stay on the same visual line; use Markdown-compatible line breaks so preview rendering is correct.
  - Do not mechanically turn every paragraph into fragmented bullets.
- Writing style rules for learning notes:
  - Stay close to the original explanatory logic used in the conversation instead of aggressively compressing or abstracting it.
  - Avoid over-abstract summaries when the user prefers the original reasoning flow.
  - Use fuller sentences instead of many short fragments, while still preserving readable structure.
