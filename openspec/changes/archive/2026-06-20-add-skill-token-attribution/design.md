## Context

The agent-office-cli system loads skills from `.agents/skills/<skill-name>/SKILL.md` files. Skills can be either:
- **Static skills**: Configured in `agent-office.yaml` under each agent's `skills` field
- **Dynamic skills**: Triggered by `/skill-name` syntax in messages

Currently, skills are loaded as plain markdown files and their content is appended to the system prompt. There is no mechanism to track which external token providers should be credited when a skill is used.

The system already tracks token usage per agent response in the `Event` struct's metadata field (see `cmd/agent-office/main.go:1393`), but this only captures total tokens without attribution to specific providers.

## Goals / Non-Goals

**Goals:**

- Enable skills to declare token attribution through YAML frontmatter
- Track which provider should be credited for input tokens and output tokens separately
- Include token provider attribution in event metadata for analytics/accounting
- Maintain backward compatibility with skills that don't have frontmatter
- Use existing YAML parsing infrastructure (gopkg.in/yaml.v3 is already in use)

**Non-Goals:**

- Implementing actual payment routing or credit deduction mechanisms
- Validating provider names against a whitelist or registry
- Changing how API authentication tokens work (separate concern from attribution)
- Modifying the actual LLM provider used for inference (attribution is accounting metadata only)

## Decisions

### Use YAML frontmatter in SKILL.md files

**Decision**: Add optional YAML frontmatter to the beginning of SKILL.md files, delimited by `---` markers.

**Rationale**: 
- Keeps configuration co-located with skill content (single source of truth)
- Frontmatter is a well-established pattern (Jekyll, Hugo, Spectra artifacts all use it)
- Easy to parse with existing `gopkg.in/yaml.v3` library
- Backward compatible (skills without frontmatter continue to work)

**Alternatives considered**:
- Separate `config.yaml` per skill directory: Increases file count, easier for configuration and content to drift out of sync
- Global `skills.yaml` configuration file: Requires maintaining synchronization between skill names and config entries, prone to stale entries

**Format**:
```markdown
---
input_token_provider: rtk-ai/rtk
output_token_provider: Caveman
---

# Skill: example-skill

Skill content...
```

### Separate input and output token attribution

**Decision**: Track `input_token_provider` and `output_token_provider` as separate fields.

**Rationale**:
- Input tokens (prompt construction) and output tokens (LLM generation) can legitimately have different attribution targets
- Provides flexibility for hybrid skill scenarios (e.g., prompt engineering from one provider, response generation credited to another)
- Aligns with how LLM APIs report usage (separate input/output token counts)

**Alternatives considered**:
- Single `token_provider` field: Simpler but loses the ability to attribute input/output separately
- Array of providers: More complex, unclear how to map providers to token types

### Fallback to agent-level provider when skill doesn't specify

**Decision**: If a skill's frontmatter is missing or doesn't specify token providers, fall back to the agent's configured `Provider` field.

**Rationale**:
- Provides sensible defaults for existing skills and new skills without attribution needs
- Maintains current behavior where all tokens are attributed to the agent's provider
- Avoids requiring every skill to specify attribution

**Alternatives considered**:
- Leave attribution empty: Would lose tracking data for skills without explicit config
- Use a hardcoded default like "unknown": Less useful than agent's actual provider

### Add token attribution to event metadata

**Decision**: When generating agent response events, include `input_token_provider` and `output_token_provider` in the event's metadata map.

**Rationale**:
- Event metadata is already used for token tracking (`metadata["tokens"]` at line 1393)
- Keeps attribution data in the same location as token counts
- Enables downstream analytics/logging to correlate tokens with providers
- Non-breaking change (adding new metadata fields doesn't affect existing consumers)

**Alternatives considered**:
- Create new event type for token attribution: Overcomplicates the event stream
- Store in separate logging system: Fragments related data across multiple systems

## Implementation Contract

**Behavior**:
- When `loadSkillPrompt()` reads a SKILL.md file, it SHALL parse YAML frontmatter if present
- If frontmatter contains `input_token_provider` or `output_token_provider`, those values SHALL be returned alongside the skill content
- When an agent uses a skill (static or dynamic), the skill's token providers SHALL be included in the response event metadata
- Skills without frontmatter SHALL continue to work unchanged, using the agent's provider for attribution

**Interface / Data Shape**:

1. **Modified `loadSkillPrompt()` return signature**:
   - Currently: `func loadSkillPrompt(skillName string) string`
   - New: `func loadSkillPrompt(skillName string) (content string, inputProvider string, outputProvider string)`

2. **YAML frontmatter structure** in SKILL.md:
   ```yaml
   ---
   input_token_provider: <provider-identifier>
   output_token_provider: <provider-identifier>
   ---
   ```
   Both fields are optional. Provider identifiers are strings (e.g., "rtk-ai/rtk", "Caveman", "openrouter").

3. **Event metadata additions** (in Event.Metadata map):
   - `input_token_provider` (string): Provider credited for input tokens
   - `output_token_provider` (string): Provider credited for output tokens

**Failure Modes**:
- Invalid YAML frontmatter: Log warning, treat as skill without frontmatter (use agent provider)
- Missing file: Existing behavior (empty string return, silently skip)
- Frontmatter present but fields missing: Use agent provider for missing fields

**Acceptance Criteria**:
1. Create a test skill with frontmatter specifying `input_token_provider: "test-input"` and `output_token_provider: "test-output"`
2. Configure an agent with this skill in `agent-office.yaml`
3. Send a message triggering the agent to respond
4. Verify the response event's metadata contains `input_token_provider: "test-input"` and `output_token_provider: "test-output"`
5. Test a skill without frontmatter and verify it falls back to the agent's provider
6. Test dynamic skill loading (`/skill-name`) and verify token attribution works

**Scope Boundaries**:
- **In scope**: Parsing frontmatter, extracting provider fields, including in event metadata, fallback logic
- **Out of scope**: Validating provider names, implementing billing, changing actual LLM provider used for inference, UI display of token attribution

## Risks / Trade-offs

**[Risk]** Frontmatter parsing errors could break skill loading
→ **Mitigation**: Use defensive parsing with error recovery; malformed frontmatter falls back to treating the file as plain content

**[Risk]** Provider name typos in frontmatter could lead to incorrect attribution
→ **Mitigation**: Explicitly documented as out-of-scope; attribution is metadata-only and doesn't affect runtime behavior. Future enhancement could add validation.

**[Risk]** Changing `loadSkillPrompt()` signature requires updating all call sites
→ **Mitigation**: Limited blast radius (used in 2 places: static skill loading at line 1322, dynamic skill loading at line 1331). Can be addressed in a single commit.

**[Trade-off]** YAML frontmatter adds parsing complexity vs. simpler plain text
→ **Accepted**: The flexibility and co-location benefits outweigh the minimal parsing overhead

**[Trade-off]** No validation of provider names means invalid values could be stored
→ **Accepted**: This is metadata-only tracking, not runtime behavior. Invalid values don't break functionality, just attribution accuracy.
