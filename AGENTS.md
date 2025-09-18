# Repository Guidelines

## Project Structure & Module Organization
- `Workflow/chatgpt` and `Workflow/dalle` are executable JavaScript for Automation (JXA) entry points triggered by Alfred; keep shared helpers in these files to minimise duplication.
- `Workflow/info.plist` defines objects, connections, and version metadata—edit via Alfred where possible and lint with `plutil` before committing.
- `Workflow/images/about` stores gallery screenshots; update filenames in-place to avoid breaking references.
- Root-level `README.md`, `FAQ.md`, and `LICENSE` document behaviour and legal terms—sync any behavioural changes here alongside code updates.

## Build, Test, and Development Commands
- Export required variables (e.g. `export openai_api_key=...`, `export gpt_model=gpt-4o`) to mirror Alfred’s configuration when running locally.
- `./Workflow/chatgpt 'status check'` runs the ChatGPT script via its shebang; use this to sanity check prompts without packaging.
- `osascript Workflow/dalle 'sunset sketch'` exercises the image workflow from Terminal.
- `plutil -lint Workflow/info.plist` validates workflow metadata prior to release.

## Coding Style & Naming Conventions
- Use 2-space indentation, `const`/`let` over `var`, and `camelCase` for functions and helpers matching existing scripts.
- Rely on template literals for dynamic strings and early returns for Alfred routing decisions; add brief comments only for non-obvious macOS or Alfred quirks.
- Keep environment-variable keys in lowercase with underscores (`openai_api_key`) to align with Alfred configuration fields.

## Testing Guidelines
- No automated test harness exists; verify behaviour by running the workflow in Alfred and from the command line with representative prompts.
- Confirm streaming, history persistence, and error paths by toggling cache files in `~/Library/Application Support/Alfred/Workflow Data/com.openai.workflow`.
- Before publishing, clear cached chats/images and re-run to ensure a clean first-run experience.

## Commit & Pull Request Guidelines
- Follow the existing short, imperative commit style (`README.md: Fix unicode character`, `Update to 25.2`); include the touched component when helpful.
- In pull requests, summarise behaviour changes, list manual test prompts, link any tracking issues, and attach refreshed screenshots if UI output shifts.
- Bump workflow version metadata in `Workflow/info.plist` and note dependency or configuration changes in the PR description.

## Security & Configuration Tips
- Never commit API keys or cached chats; Alfred stores secrets in user preferences—confirm `.gitignore` still excludes exported archives.
- Document new configuration keys in `README.md` and provide safe defaults in the workflow variables to prevent empty-value crashes.
- When sharing builds, export via Alfred’s “Export…” option to ensure signatures and entitlements remain intact.
