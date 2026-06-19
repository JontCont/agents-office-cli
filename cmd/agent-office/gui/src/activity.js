// Sprite Animation and Activity Tab visualizer
(function() {
  let activeCharacters = [];
  let animationFrameId = null;
  let isAnimating = false;

  // Check Canvas support
  function isCanvasSupported() {
    const elem = document.createElement('canvas');
    return !!(elem.getContext && elem.getContext('2d'));
  }

  // Helper to escape HTML safely
  function safeHtml(str) {
    if (!str) return '';
    return str
      .replace(/&/g, '&amp;')
      .replace(/</g, '&lt;')
      .replace(/>/g, '&gt;')
      .replace(/"/g, '&quot;')
      .replace(/'/g, '&#039;');
  }

  // Get default emoji avatar based on agent name
  function getFallbackAvatar(name) {
    const nameLower = (name || '').toLowerCase();
    if (nameLower.includes('code') || nameLower.includes('developer') || nameLower.includes('program')) {
      return '💻';
    }
    if (nameLower.includes('review') || nameLower.includes('audit') || nameLower.includes('test')) {
      return '🔍';
    }
    if (nameLower.includes('plan') || nameLower.includes('strat') || nameLower.includes('lead')) {
      return '📋';
    }
    if (nameLower.includes('architect') || nameLower.includes('design')) {
      return '📐';
    }
    return '🤖';
  }

  // Initialize Activity View
  async function initializeActivityView() {
    const stage = document.querySelector('.activity-stage');
    if (!stage) return;

    // Reset current state
    stage.innerHTML = '';
    activeCharacters = [];
    if (animationFrameId) {
      cancelAnimationFrame(animationFrameId);
      animationFrameId = null;
    }
    isAnimating = false;

    // Canvas support check
    if (!isCanvasSupported()) {
      stage.innerHTML = `<div class="activity-error">Canvas not supported. Please use a modern browser.</div>`;
      console.error('Canvas API is not available on this browser.');
      return;
    }

    // Load agents
    try {
      const res = await fetch('/api/agents');
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      const agents = data.agents || [];

      if (agents.length === 0) {
        stage.innerHTML = `<div class="activity-empty">No agents configured. Please go to the Agents tab to create some.</div>`;
        return;
      }

      // Populate stage and initialize anim states
      for (const agent of agents) {
        const charContainer = document.createElement('div');
        charContainer.className = 'agent-character';
        charContainer.setAttribute('data-agent', agent.name);

        const canvas = document.createElement('canvas');
        canvas.width = 64;
        canvas.height = 64;

        const labelDiv = document.createElement('div');
        labelDiv.className = 'agent-character-label';
        
        const nameDiv = document.createElement('div');
        nameDiv.className = 'agent-character-name';
        nameDiv.textContent = agent.name;

        const roleDiv = document.createElement('div');
        roleDiv.className = 'agent-character-role';
        roleDiv.textContent = agent.role;

        labelDiv.appendChild(nameDiv);
        labelDiv.appendChild(roleDiv);
        charContainer.appendChild(canvas);
        charContainer.appendChild(labelDiv);
        stage.appendChild(charContainer);

        const charState = {
          name: agent.name,
          role: agent.role,
          avatar: agent.avatar || getFallbackAvatar(agent.name),
          visualCharacter: agent.visual_character || '',
          canvas: canvas,
          ctx: canvas.getContext('2d'),
          animationState: 'idle',
          spriteImg: null,
          spriteConfig: null,
          currentFrameIndex: 0,
          lastFrameTime: 0,
          talkingTimeout: null,
          loadFailed: false
        };

        activeCharacters.push(charState);

        // Load asset if visual character is configured
        if (charState.visualCharacter) {
          loadCharacterAssets(charState);
        } else {
          // Draw initial emoji immediately
          drawEmojiFallback(charState);
        }
      }

      // Start animation loop
      isAnimating = true;
      animationFrameId = requestAnimationFrame(animate);

    } catch (e) {
      console.error('Failed to load agents in Activity view:', e);
      stage.innerHTML = `<div class="activity-error">Failed to load agents: ${safeHtml(e.message)}</div>`;
    }
  }

  // Load Character Assets: config.json and sprite.png
  async function loadCharacterAssets(charState) {
    const name = charState.visualCharacter;
    const configUrl = `/.agents/visual-assets/characters/${name}/config.json`;

    try {
      const configRes = await fetch(configUrl);
      if (!configRes.ok) {
        throw new Error(`Failed to load config.json (HTTP ${configRes.status})`);
      }
      const configData = await configRes.json();
      
      // Basic validation
      if (!configData.spriteSheet || !configData.animations) {
        throw new Error(`config.json is missing required fields (spriteSheet/animations)`);
      }

      const imgUrl = `/.agents/visual-assets/characters/${name}/${configData.spriteSheet}`;
      const img = new Image();
      
      img.onload = () => {
        charState.spriteConfig = configData;
        charState.spriteImg = img;
        charState.currentFrameIndex = 0;
        charState.lastFrameTime = performance.now();
      };

      img.onerror = () => {
        console.error(`[Sprite Error] Failed to load image asset: ${imgUrl} for visual character "${name}"`);
        charState.loadFailed = true;
        drawEmojiFallback(charState);
      };

      img.src = imgUrl;

    } catch (e) {
      console.error(`[Config Error] Failed to load visual character "${name}":`, e.message);
      charState.loadFailed = true;
      drawEmojiFallback(charState);
    }
  }

  // Draw Emoji Fallback
  function drawEmojiFallback(charState) {
    const ctx = charState.ctx;
    const canvas = charState.canvas;
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    ctx.save();
    // Render emoji avatar
    ctx.font = '36px sans-serif';
    ctx.textAlign = 'center';
    ctx.textBaseline = 'middle';
    ctx.fillText(charState.avatar, canvas.width / 2, canvas.height / 2);
    ctx.restore();
  }

  // Main animation update tick
  function animate(timestamp) {
    if (!isAnimating) return;

    for (const char of activeCharacters) {
      if (char.loadFailed || !char.spriteImg || !char.spriteConfig) {
        // Static emoji fallback redraw (to prevent blank canvas during theme shifts)
        drawEmojiFallback(char);
        continue;
      }

      const config = char.spriteConfig;
      const state = char.animationState;
      const anim = config.animations[state] || config.animations['idle'];
      
      if (!anim || !anim.frames || anim.frames.length === 0) {
        // Fallback to static first frame if animation frames not defined
        drawStaticFirstFrame(char);
        continue;
      }

      const fps = anim.fps || 8;
      const interval = 1000 / fps;
      const elapsed = timestamp - char.lastFrameTime;

      if (elapsed > interval) {
        char.currentFrameIndex++;
        const shouldLoop = anim.loop !== false;
        
        if (char.currentFrameIndex >= anim.frames.length) {
          if (shouldLoop) {
            char.currentFrameIndex = 0;
          } else {
            char.currentFrameIndex = anim.frames.length - 1;
          }
        }
        char.lastFrameTime = timestamp - (elapsed % interval);
      }

      // Draw the current sprite frame
      drawSpriteFrame(char, anim.frames[char.currentFrameIndex]);
    }

    animationFrameId = requestAnimationFrame(animate);
  }

  // Draw Sprite Frame
  function drawSpriteFrame(char, frameVal) {
    const ctx = char.ctx;
    const canvas = char.canvas;
    const config = char.spriteConfig;
    const img = char.spriteImg;

    ctx.clearRect(0, 0, canvas.width, canvas.height);

    // Frame validation
    const frameIndex = parseInt(frameVal, 10);
    const totalPossibleFrames = Math.floor(img.width / config.frameWidth);

    if (isNaN(frameIndex) || frameIndex < 0 || frameIndex >= totalPossibleFrames) {
      console.warn(`[Animation Warning] Invalid frame index ${frameVal} for visual character "${char.visualCharacter}" (max frames: ${totalPossibleFrames}). Drawing static first frame.`);
      ctx.drawImage(img, 0, 0, config.frameWidth, config.frameHeight, 0, 0, canvas.width, canvas.height);
      return;
    }

    const srcX = frameIndex * config.frameWidth;
    ctx.drawImage(
      img,
      srcX, 0,
      config.frameWidth, config.frameHeight,
      0, 0,
      canvas.width, canvas.height
    );
  }

  // Draw Static First Frame
  function drawStaticFirstFrame(char) {
    const ctx = char.ctx;
    const canvas = char.canvas;
    const config = char.spriteConfig;
    const img = char.spriteImg;

    ctx.clearRect(0, 0, canvas.width, canvas.height);
    ctx.drawImage(
      img,
      0, 0,
      config.frameWidth, config.frameHeight,
      0, 0,
      canvas.width, canvas.height
    );
  }

  // WebSocket event handler for real-time speech animations
  function handleWorkforceEvent(evt) {
    if (evt.type === 'agent.speak') {
      const sender = evt.sender;
      // Match characters case-insensitively
      const char = activeCharacters.find(c => c.name.toLowerCase() === sender.toLowerCase());
      if (char) {
        // Clear any existing talking timeout
        if (char.talkingTimeout) {
          clearTimeout(char.talkingTimeout);
          char.talkingTimeout = null;
        }

        // Switch state to talking
        if (char.animationState !== 'talking') {
          char.animationState = 'talking';
          char.currentFrameIndex = 0;
          char.lastFrameTime = performance.now();
        }

        // Set timeout to return to idle after 3 seconds of silence
        char.talkingTimeout = setTimeout(() => {
          char.animationState = 'idle';
          char.currentFrameIndex = 0;
          char.lastFrameTime = performance.now();
          char.talkingTimeout = null;
        }, 3000);
      }
    }
  }

  // Bind custom event listener
  document.addEventListener('workforce-event', (e) => {
    if (e.detail) {
      handleWorkforceEvent(e.detail);
    }
  });

  // Expose function globally
  window.initializeActivityView = initializeActivityView;

})();
