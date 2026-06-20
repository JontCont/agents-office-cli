## 1. Frontmatter Parsing Implementation

- [x] 1.1 Use YAML frontmatter in SKILL.md files: Modify `loadSkillPrompt()` in `cmd/agent-office/main.go` to return three values (content, inputProvider, outputProvider) instead of just content. Behavior: When a SKILL.md file is loaded, YAML frontmatter is parsed if present and provider fields are extracted; when frontmatter is missing or malformed, empty strings are returned for provider fields. Verification: Unit test `TestLoadSkillPrompt_WithFrontmatter` passes.

- [x] 1.2 Parse token provider from SKILL.md frontmatter: Implement frontmatter parsing logic using `gopkg.in/yaml.v3` to extract `input_token_provider` and `output_token_provider` fields from the beginning of SKILL.md content delimited by `---` markers, implementing separate input and output token attribution. Behavior: Valid YAML frontmatter is parsed and fields are extracted; malformed YAML logs a warning and returns empty provider strings; content without `---` delimiters is treated as having no frontmatter. Verification: Unit test `TestParseFrontmatter_ValidYAML`, `TestParseFrontmatter_Malformed`, and `TestParseFrontmatter_MissingDelimiters` all pass.

- [x] 1.3 Update static skill loading loop (line 1321-1325 in `cmd/agent-office/main.go`) to capture provider values from `loadSkillPrompt()` and track them for later event metadata inclusion. Behavior: Static skills with frontmatter have their provider values stored; skills without frontmatter use agent's provider. Verification: Manual test with agent configured with a static skill containing frontmatter shows provider values logged during skill loading.

- [x] 1.4 Update dynamic skill loading loop (line 1329-1334 in `cmd/agent-office/main.go`) to capture provider values from `loadSkillPrompt()` and track them for later event metadata inclusion. Behavior: Dynamic skills triggered by `/skill-name` syntax have their provider values stored; last loaded skill's providers are used if multiple skills are loaded. Verification: Manual test triggering `/test-skill` with frontmatter shows provider values logged.

## 2. Fallback Logic Implementation

- [x] 2.1 Fallback to agent-level provider when skill doesn't specify: Implement fallback logic that uses agent's `Provider` field when skill frontmatter doesn't specify `input_token_provider` or `output_token_provider`. Behavior: When a provider field is empty string after frontmatter parsing, it is replaced with the agent's configured provider from `activeAgent.Provider`. Verification: Unit test `TestProviderFallback_AgentDefault` passes; manual test with skill missing output provider shows agent provider used for output.

- [x] 2.2 Handle case where agent's provider is also empty by using empty string (no fallback beyond agent level). Behavior: If both skill frontmatter and agent provider are empty, the metadata fields remain empty strings. Verification: Unit test `TestProviderFallback_BothEmpty` passes.

## 3. Event Metadata Integration

- [x] 3.1 Add token attribution to event metadata and include token attribution in event metadata: Add `input_token_provider` and `output_token_provider` fields to the event metadata map at line 1393 in `cmd/agent-office/main.go` where token count is already recorded. Behavior: When an agent response event is created, the metadata contains the token provider attribution fields from the most recently loaded skill (or agent default). Verification: Integration test sends message to agent with skill, checks event JSON contains `"input_token_provider": "test-input"` and `"output_token_provider": "test-output"`.

- [x] 3.2 Include token attribution in event metadata for all event types: Ensure token attribution fields are included in all event types where token count is tracked (agent.speak events at minimum). Behavior: Every agent response event that includes `metadata["tokens"]` also includes `metadata["input_token_provider"]` and `metadata["output_token_provider"]`. Verification: Review event generation code paths and confirm all token-tracking events include provider fields; manual test inspecting SSE stream confirms fields present.

## 4. Backward Compatibility Verification

- [x] 4.1 Preserve backward compatibility: Test existing skills without frontmatter continue to load and execute without errors. Behavior: Skills created before this feature (plain markdown with no `---` delimiters) load successfully and use agent's provider for attribution. Verification: Manual test with an existing skill from `.agents/skills/` directory shows skill works unchanged; event metadata shows agent's provider.

- [x] 4.2 Preserve backward compatibility for skills with only content: Test skills with only content (no frontmatter) are treated entirely as skill content without any data loss. Behavior: The entire file content is returned as skill prompt when no frontmatter is detected. Verification: Manual test with a skill file starting with `# Skill: test` (no frontmatter) shows the entire file is included in the system prompt.

## 5. Error Handling and Edge Cases

- [x] 5.1 Implement defensive error handling for YAML parsing that logs warnings but doesn't break skill loading. Behavior: When `yaml.Unmarshal()` returns an error, a warning is logged to console/log file and the function continues as if no frontmatter exists. Verification: Manual test with skill containing `---\ninvalid yaml: [unclosed\n---` shows warning logged and skill loads with agent provider used.

- [x] 5.2 Handle empty frontmatter blocks (just `---\n---\n` with no fields) by treating as no frontmatter. Behavior: Empty frontmatter block results in empty provider strings, triggering agent provider fallback. Verification: Unit test `TestParseFrontmatter_EmptyBlock` passes.

- [x] 5.3 Handle frontmatter with unexpected fields gracefully (ignore unknown fields, only extract the two provider fields). Behavior: YAML frontmatter with fields other than `input_token_provider` and `output_token_provider` is parsed successfully, unknown fields are ignored. Verification: Unit test `TestParseFrontmatter_ExtraFields` with frontmatter containing `extra_field: value` passes and only provider fields are extracted.

## 6. Testing and Documentation

- [x] 6.1 Create test skill at `.agents/skills/test-token-attribution/SKILL.md` with complete frontmatter (`input_token_provider: test-input`, `output_token_provider: test-output`). Behavior: Test skill exists with valid frontmatter for integration testing. Verification: File exists at path and contains valid YAML frontmatter.

- [x] 6.2 Run end-to-end integration test following acceptance criteria from design.md (configure agent with test skill, send message, verify event metadata). Behavior: Complete workflow from skill configuration to event metadata verification works as specified. Verification: All six acceptance criteria from design.md lines 129-134 pass.

- [x] 6.3 Add example frontmatter format to a documentation file or comment in `cmd/agent-office/main.go` near `loadSkillPrompt()` function. Behavior: Developers can see the expected frontmatter format without reading external specs. Verification: Code comment or doc file exists showing the `---\ninput_token_provider: ...\n---` format.
