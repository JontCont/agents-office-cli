## ADDED Requirements

### Requirement: Parse token provider from SKILL.md frontmatter

The system SHALL parse YAML frontmatter from SKILL.md files to extract token provider attribution fields. The frontmatter block SHALL be delimited by `---` markers at the beginning of the file.

#### Scenario: Valid frontmatter with both providers

- **WHEN** a SKILL.md file contains valid YAML frontmatter with `input_token_provider` and `output_token_provider` fields
- **THEN** the system returns both provider values along with the skill content

##### Example: Complete frontmatter

- **GIVEN** SKILL.md contains:
  ```
  ---
  input_token_provider: rtk-ai/rtk
  output_token_provider: Caveman
  ---
  
  # Skill content
  ```
- **WHEN** the skill is loaded
- **THEN** `input_token_provider` is "rtk-ai/rtk" AND `output_token_provider` is "Caveman"

#### Scenario: Frontmatter with only input provider

- **WHEN** frontmatter contains only `input_token_provider`
- **THEN** the system returns the input provider value AND uses agent's provider for output

#### Scenario: Frontmatter with only output provider

- **WHEN** frontmatter contains only `output_token_provider`
- **THEN** the system returns the output provider value AND uses agent's provider for input

#### Scenario: No frontmatter present

- **WHEN** SKILL.md has no YAML frontmatter block
- **THEN** the system uses agent's provider for both input and output attribution

#### Scenario: Malformed frontmatter

- **WHEN** SKILL.md has malformed YAML in the frontmatter block
- **THEN** the system logs a warning AND treats the file as having no frontmatter

##### Example: Malformed YAML cases

| Input | Behavior | Notes |
|-------|----------|-------|
| `---\ninvalid yaml: [unclosed\n---` | Use agent provider | Invalid syntax |
| `---\n---\n# Content` | Use agent provider | Empty frontmatter |
| `input_token_provider: foo\n# Content` | Use agent provider | Missing delimiters |

### Requirement: Include token attribution in event metadata

The system SHALL include token provider attribution in the metadata of agent response events when skills are used.

#### Scenario: Static skill with attribution

- **WHEN** an agent uses a static skill (configured in agent's `skills` field) that has token provider frontmatter
- **THEN** the response event metadata contains `input_token_provider` and `output_token_provider` fields matching the skill's frontmatter

#### Scenario: Dynamic skill with attribution

- **WHEN** an agent uses a dynamic skill (triggered by `/skill-name` in message) that has token provider frontmatter
- **THEN** the response event metadata contains `input_token_provider` and `output_token_provider` fields matching the skill's frontmatter

#### Scenario: Multiple skills used in one response

- **WHEN** an agent uses multiple skills in a single response (e.g., one static + one dynamic)
- **THEN** the event metadata reflects the token attribution from the last skill loaded

##### Example: Two skills with different providers

- **GIVEN** static skill A has `output_token_provider: ProviderA` AND dynamic skill B has `output_token_provider: ProviderB`
- **WHEN** agent uses both skills in one response
- **THEN** event metadata contains `output_token_provider: ProviderB` (last skill wins)

#### Scenario: Skill without frontmatter fallback

- **WHEN** an agent uses a skill that has no token provider frontmatter
- **THEN** the event metadata contains the agent's configured `Provider` value for both input and output

##### Example: Agent provider fallback

- **GIVEN** agent configured with `provider: openrouter` AND skill has no frontmatter
- **WHEN** skill is used
- **THEN** event metadata contains `input_token_provider: openrouter` AND `output_token_provider: openrouter`

### Requirement: Preserve backward compatibility

The system SHALL continue to load and execute skills that do not have YAML frontmatter without any change in behavior.

#### Scenario: Existing skills without frontmatter

- **WHEN** a skill created before this feature has no frontmatter
- **THEN** the skill loads successfully AND uses agent's provider for attribution AND skill content is unchanged

#### Scenario: Skills with only content

- **WHEN** SKILL.md contains only markdown content with no `---` delimiters
- **THEN** the entire file is treated as skill content AND uses agent's provider for attribution
