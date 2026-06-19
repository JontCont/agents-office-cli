# Visual Assets for Agents

This directory manages the visual assets (sprite sheets and configurations) for the workspace agent visualizer companion GUI.

## Directory Structure

Visual assets must follow this structure:

```
.agents/visual-assets/
├── README.md
└── characters/
    ├── <character_name>/
    │   ├── sprite.png
    │   └── config.json
    └── ...
```

- Each subdirectory inside `characters/` represents a single visual character (e.g. `engineer`, `strategist`).
- The directory name will be displayed in the companion GUI Agents preference settings dropdown.

## Sprite Sheet Format (`sprite.png`)

- **Frame Dimensions**: Each frame must be `64x64` pixels (or other fixed dimensions configured in `config.json`).
- **Layout**: Frames must be arranged in a single horizontal strip (e.g., for an 11-frame animation of `64x64` pixels, the dimensions of `sprite.png` should be `704x64` pixels).
- **Background**: Transparent background is highly recommended for compatibility with dark/light UI theme styles.

## Configuration Format (`config.json`)

Each character directory must include a `config.json` file. Here is an example config:

```json
{
  "name": "engineer",
  "spriteSheet": "sprite.png",
  "frameWidth": 64,
  "frameHeight": 64,
  "animations": {
    "idle": { "frames": [0, 1, 2, 3], "fps": 8, "loop": true },
    "talking": { "frames": [4, 5, 6, 7], "fps": 12, "loop": true },
    "thinking": { "frames": [8, 9, 10], "fps": 6, "loop": true }
  }
}
```

- `name`: Display name of the character.
- `spriteSheet`: Name of the image file (relative to the directory containing `config.json`).
- `frameWidth`: Width of each individual animation frame (pixels).
- `frameHeight`: Height of each individual animation frame (pixels).
- `animations`: A dictionary of animations, supporting:
  - `idle`: Played when the agent is idle.
  - `talking`: Played when the agent is actively speaking (websocket `agent.speak` event).
  - `thinking`: Played when the agent is performing thoughts/computations.

Each animation definition must contain:
- `frames`: An array of 0-based horizontal frame indices.
- `fps`: Speed of frame transitions (frames per second).
- `loop`: Whether the animation repeats continuously.

## Step-by-Step Instructions: Adding a Custom Character

1. Create a new subdirectory inside `.agents/visual-assets/characters/`, e.g., `.agents/visual-assets/characters/custom-robot/`.
2. Generate or place your sprite sheet image as a horizontal PNG layout in that folder (e.g., `sprite.png`).
3. Create a `config.json` file in that folder mapping your animation frame indices for `idle`, `talking`, and `thinking` states.
4. Refresh or reload your Companion GUI browser window.
5. In the **Agents** tab, edit an agent and select your new custom character from the **Visual Character** dropdown, then click Save.
6. Open the **Activity** tab to see your agent animate!
