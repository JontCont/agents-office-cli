## Why

Skills are currently global resources shared by all agents, but there is no way to track which external providers should be credited for input and output tokens when a skill is used. This makes it impossible to properly attribute token usage for accounting, analytics, or billing purposes when skills leverage different token providers.

## What Changes

- Add YAML frontmatter support to `SKILL.md` files for specifying token attribution
- Parse `input_token_provider` and `output_token_provider` fields from skill frontmatter
- Include token provider attribution in event metadata when skills are used
- Fallback to agent-level provider when skill does not specify token providers

## Non-Goals

- This change does NOT implement actual payment routing or credit deduction
- This change does NOT validate provider names against a registry
- This change does NOT change how API authentication tokens work
- This change does NOT affect backward compatibility with existing skills (skills without frontmatter continue to work)

## Capabilities

### New Capabilities

- `skill-token-attribution`: Parse and track token provider attribution for skills through YAML frontmatter in SKILL.md files

### Modified Capabilities

(none)

## Impact

- Affected specs: `skill-token-attribution` (new)
- Affected code:
  - Modified: `cmd/agent-office/main.go` (loadSkillPrompt function, event metadata generation)
  - New: (none - using existing YAML parsing libraries)
  - Removed: (none)
