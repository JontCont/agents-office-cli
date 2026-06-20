// 2D Virtual Office Canvas Visualization Engine
(function() {
  let activeCharacters = [];
  let animationFrameId = null;
  let isAnimating = false;
  let activeInteraction = null; // Track { speakerName, targetName, startTime }

  // Load background map image
  const mapImg = new Image();
  mapImg.src = '/.agents/visual-assets/office_map.png';
  let isMapLoaded = false;
  mapImg.onload = () => {
    isMapLoaded = true;
  };

  // Fixed resolutions and coordinates mapping
  const COORDINATES = {
    working: [
      { x: 140, y: 140 }, // Desk 1 (Lead Strategist)
      { x: 260, y: 140 }, // Desk 2 (Technical Architect)
      { x: 140, y: 240 }  // Desk 3 (Critical Reviewer)
    ],
    discussing: [
      { x: 445, y: 220 }, // Seat 1 (left of table)
      { x: 530, y: 220 }, // Seat 2 (right of table)
      { x: 488, y: 190 }  // Seat 3 (above table)
    ],
    listening: [
      { x: 445, y: 220 },
      { x: 530, y: 220 },
      { x: 488, y: 190 }
    ],
    resting: [
      { x: 440, y: 70 },  // Spot 1 (販賣機旁)
      { x: 470, y: 70 },  // Spot 2 (廚房中央)
      { x: 500, y: 70 }   // Spot 3 (冰箱旁)
    ]
  };

  function getCoordinates(index, state) {
    const s = (state || '').toLowerCase();
    let coordsList = COORDINATES.resting;
    if (s === 'working') {
      coordsList = COORDINATES.working;
    } else if (s === 'discussing' || s === 'listening') {
      coordsList = COORDINATES.discussing;
    }
    const idx = index % coordsList.length;
    return coordsList[idx];
  }

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

    activeCharacters = [];
    activeInteraction = null;
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

    const canvas = document.getElementById('activity-canvas');
    if (!canvas) {
      console.error('Canvas element #activity-canvas not found.');
      return;
    }

    // Load agents
    try {
      const res = await fetch('/api/agents');
      if (!res.ok) throw new Error(`HTTP ${res.status}`);
      const data = await res.json();
      const agents = data.agents || [];

      if (agents.length === 0) {
        stage.innerHTML = `
          <canvas id="activity-canvas" width="709" height="282"></canvas>
          <div class="activity-empty">No agents configured. Please go to the Agents tab to create some.</div>
        `;
        return;
      }

      // Populate agent characters state
      for (let i = 0; i < agents.length; i++) {
        const agent = agents[i];
        const spawnSpot = getCoordinates(i, 'resting');

        const charState = {
          name: agent.name,
          role: agent.role,
          avatar: agent.avatar || getFallbackAvatar(agent.name),
          visualCharacter: agent.visual_character || '',
          state: 'Resting', // 'Resting', 'Working', 'Discussing', 'Listening'
          x: spawnSpot.x,
          y: spawnSpot.y,
          targetX: spawnSpot.x,
          targetY: spawnSpot.y,
          facing: 'right', // 'left' or 'right'
          animationState: 'idle',
          lastAnimationState: 'idle',
          spriteImg: null,
          spriteConfig: null,
          currentFrameIndex: 0,
          lastFrameTime: performance.now(),
          talkingTimeout: null,
          loadFailed: false
        };

        activeCharacters.push(charState);

        // Load asset if visual character is configured
        if (charState.visualCharacter) {
          loadCharacterAssets(charState);
        }
      }

      // Start animation loop
      isAnimating = true;
      animationFrameId = requestAnimationFrame(animate);

    } catch (e) {
      console.error('Failed to load agents in Activity view:', e);
      stage.innerHTML = `
        <canvas id="activity-canvas" width="709" height="282"></canvas>
        <div class="activity-error">Failed to load agents: ${safeHtml(e.message)}</div>
      `;
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
      };

      img.src = imgUrl;

    } catch (e) {
      console.error(`[Config Error] Failed to load visual character "${name}":`, e.message);
      charState.loadFailed = true;
    }
  }

  // Main animation update tick
  function animate(timestamp) {
    if (!isAnimating) return;

    const canvas = document.getElementById('activity-canvas');
    if (!canvas) {
      animationFrameId = requestAnimationFrame(animate);
      return;
    }
    const ctx = canvas.getContext('2d');

    // 1. Clear Canvas
    ctx.clearRect(0, 0, canvas.width, canvas.height);

    // 2. Draw Pixel-art Background Map
    if (isMapLoaded) {
      ctx.drawImage(mapImg, 0, 0, canvas.width, canvas.height);
    } else {
      ctx.fillStyle = '#1c1c2b';
      ctx.fillRect(0, 0, canvas.width, canvas.height);
    }

    // 3. Update coordinates interpolation
    for (let i = 0; i < activeCharacters.length; i++) {
      const char = activeCharacters[i];
      const target = getCoordinates(i, char.state);
      char.targetX = target.x;
      char.targetY = target.y;

      const dx = char.targetX - char.x;
      const dy = char.targetY - char.y;
      const dist = Math.sqrt(dx * dx + dy * dy);
      const speed = 2.0; // Smooth movement speed (pixels/frame)

      if (dist > speed) {
        char.x += (dx / dist) * speed;
        char.y += (dy / dist) * speed;
        if (char.animationState !== 'walking') {
          char.lastAnimationState = char.animationState;
          char.animationState = 'walking';
        }
        // Update horizontal flip orientation based on velocity
        char.facing = dx < 0 ? 'left' : 'right';
      } else {
        char.x = char.targetX;
        char.y = char.targetY;
        if (char.animationState === 'walking') {
          char.animationState = char.lastAnimationState || 'idle';
        }
      }
    }

    // 4. Draw Active Mentions Connecting Arc & Pulse
    if (activeInteraction) {
      const elapsed = timestamp - activeInteraction.startTime;
      if (elapsed < 3000) {
        const charA = activeCharacters.find(c => c.name.toLowerCase() === activeInteraction.speakerName.toLowerCase());
        const charB = activeCharacters.find(c => c.name.toLowerCase() === activeInteraction.targetName.toLowerCase());

        if (charA && charB) {
          ctx.save();

          const ptA = { x: charA.x, y: charA.y - 12 };
          const ptB = { x: charB.x, y: charB.y - 12 };

          const midX = (ptA.x + ptB.x) / 2;
          const dist = Math.abs(ptA.x - ptB.x);
          const midY = Math.min(ptA.y, ptB.y) - Math.max(20, dist * 0.15);

          ctx.beginPath();
          ctx.moveTo(ptA.x, ptA.y);
          ctx.quadraticCurveTo(midX, midY, ptB.x, ptB.y);

          // Glow shadow
          ctx.shadowColor = 'rgba(99, 102, 241, 0.8)';
          ctx.shadowBlur = 6;
          ctx.strokeStyle = 'rgba(99, 102, 241, 0.6)';
          ctx.lineWidth = 2.5;
          ctx.lineCap = 'round';
          ctx.stroke();

          // Draw moving pulse particle
          const speed = 0.0015;
          const t = (elapsed * speed) % 1.0;
          const getQuadPoint = (p0, p1, p2, t) => {
            const x = (1 - t) * (1 - t) * p0.x + 2 * (1 - t) * t * p1.x + t * t * p2.x;
            const y = (1 - t) * (1 - t) * p0.y + 2 * (1 - t) * t * p1.y + t * t * p2.y;
            return { x, y };
          };

          const pulsePt = getQuadPoint(ptA, { x: midX, y: midY }, ptB, t);
          ctx.beginPath();
          ctx.arc(pulsePt.x, pulsePt.y, 4, 0, Math.PI * 2);
          ctx.fillStyle = '#818cf8';
          ctx.shadowColor = '#6366f1';
          ctx.shadowBlur = 10;
          ctx.fill();

          ctx.restore();
        }
      } else {
        activeInteraction = null;
      }
    }

    // 5. Render Character Sprites/Emojis & Labels
    const charSize = 32;
    for (const char of activeCharacters) {
      ctx.save();

      const drawX = char.x - charSize / 2;
      const drawY = char.y - charSize;

      // Flip sprite rendering if facing left
      if (char.facing === 'left') {
        ctx.translate(char.x, char.y);
        ctx.scale(-1, 1);
        ctx.translate(-char.x, -char.y);
      }

      // Draw sprite frame or emoji
      if (char.loadFailed || !char.spriteImg || !char.spriteConfig) {
        // Draw Fallback Emoji
        ctx.restore();
        ctx.save();
        ctx.font = '20px sans-serif';
        ctx.textAlign = 'center';
        ctx.textBaseline = 'middle';
        ctx.fillText(char.avatar, char.x, char.y - charSize / 2);
      } else {
        const config = char.spriteConfig;
        const state = char.animationState === 'walking' ? 'walk' : char.animationState;
        const anim = config.animations[state] || config.animations['idle'];

        if (anim && anim.frames && anim.frames.length > 0) {
          const fps = anim.fps || 8;
          const interval = 1000 / fps;
          const elapsed = timestamp - char.lastFrameTime;

          if (elapsed > interval) {
            char.currentFrameIndex = (char.currentFrameIndex + 1) % anim.frames.length;
            char.lastFrameTime = timestamp - (elapsed % interval);
          }

          const frameIndex = parseInt(anim.frames[char.currentFrameIndex], 10);
          const srcX = frameIndex * config.frameWidth;

          ctx.drawImage(
            char.spriteImg,
            srcX, 0,
            config.frameWidth, config.frameHeight,
            drawX, drawY,
            charSize, charSize
          );
        } else {
          ctx.drawImage(
            char.spriteImg,
            0, 0,
            config.frameWidth, config.frameHeight,
            drawX, drawY,
            charSize, charSize
          );
        }
      }
      ctx.restore();

      // Render Label Card Box above head
      ctx.save();
      const labelY = char.y - charSize - 20;

      ctx.font = 'bold 9px Outfit, sans-serif';
      const nameWidth = ctx.measureText(char.name).width;
      ctx.font = '600 7px Outfit, sans-serif';
      const roleWidth = ctx.measureText(`${char.role} - ${char.state}`).width;

      const maxW = Math.max(nameWidth, roleWidth) + 12;
      const boxH = 22;
      const boxX = char.x - maxW / 2;
      const boxY = labelY - 4;

      // Card Background (dark glassmorphism)
      ctx.fillStyle = 'rgba(11, 15, 25, 0.85)';
      ctx.strokeStyle = 'rgba(99, 102, 241, 0.35)';
      ctx.lineWidth = 1;
      ctx.beginPath();
      if (ctx.roundRect) {
        ctx.roundRect(boxX, boxY, maxW, boxH, 4);
      } else {
        ctx.rect(boxX, boxY, maxW, boxH);
      }
      ctx.fill();
      ctx.stroke();

      // Draw Name Text
      ctx.font = 'bold 9px Outfit, sans-serif';
      ctx.fillStyle = '#ffffff';
      ctx.textAlign = 'center';
      ctx.textBaseline = 'top';
      ctx.fillText(char.name, char.x, boxY + 3);

      // Draw Role & State
      ctx.font = '600 7px Outfit, sans-serif';
      let statusColor = '#9ca3af'; // Resting
      const s = (char.state || '').toLowerCase();
      if (s === 'working') statusColor = '#c084fc';
      else if (s === 'discussing') statusColor = '#60a5fa';
      else if (s === 'listening') statusColor = '#a5b4fc';

      ctx.fillStyle = statusColor;
      ctx.fillText(`${char.role} (${char.state || 'Resting'})`, char.x, boxY + 13);
      ctx.restore();
    }

    animationFrameId = requestAnimationFrame(animate);
  }

  // WebSocket event handler for real-time speech and thinking animations
  function handleWorkforceEvent(evt) {
    if (evt.type === 'agent.thinking') {
      const sender = evt.sender;
      const char = activeCharacters.find(c => c.name.toLowerCase() === sender.toLowerCase());
      if (char) {
        if (char.talkingTimeout) {
          clearTimeout(char.talkingTimeout);
          char.talkingTimeout = null;
        }

        // Set to thinking
        char.state = 'Working';
        char.animationState = 'thinking';
        char.currentFrameIndex = 0;
        char.lastFrameTime = performance.now();

        // Others become listening
        for (const other of activeCharacters) {
          if (other !== char) {
            other.state = 'Listening';
            other.animationState = 'idle';
            other.facing = 'right';
          }
        }
      }
    } else if (evt.type === 'agent.speak') {
      const sender = evt.sender;
      const char = activeCharacters.find(c => c.name.toLowerCase() === sender.toLowerCase());
      if (char) {
        if (char.talkingTimeout) {
          clearTimeout(char.talkingTimeout);
          char.talkingTimeout = null;
        }

        // Set to speaking/discussing
        char.state = 'Discussing';
        char.animationState = 'talking';
        char.currentFrameIndex = 0;
        char.lastFrameTime = performance.now();

        // Others become listening
        for (const other of activeCharacters) {
          if (other !== char) {
            other.state = 'Listening';
            other.animationState = 'idle';
            other.facing = 'right';
          }
        }

        // Parse mentions for facing directions
        const content = evt.content || '';
        let targetChar = null;
        for (const target of activeCharacters) {
          if (target === char) continue;
          const targetName = target.name.toLowerCase();
          const normalizedTarget = targetName.replace(/[\s\-_]/g, '');
          const normalizedContent = content.toLowerCase().replace(/[\s\-_]/g, '');
          if (normalizedContent.includes('@' + normalizedTarget) || normalizedContent.includes(normalizedTarget)) {
            targetChar = target;
            break;
          }
        }

        if (targetChar) {
          // Adjust speaker facing direction
          if (targetChar.x < char.x) {
            char.facing = 'left';
          } else {
            char.facing = 'right';
          }

          activeInteraction = {
            speakerName: char.name,
            targetName: targetChar.name,
            startTime: performance.now()
          };
        } else {
          char.facing = 'right';
        }

        // Set timeout to return to listening/idle state after 3.5 seconds
        char.talkingTimeout = setTimeout(() => {
          char.state = 'Listening';
          char.animationState = 'idle';
          char.currentFrameIndex = 0;
          char.lastFrameTime = performance.now();
          char.talkingTimeout = null;
          char.facing = 'right';
          activeInteraction = null;
        }, 3500);
      }
    } else if (evt.type === 'state.change') {
      const state = evt.content;
      if (state === 'completed' || state === 'failed' || state === 'cancelled' || state === 'queued') {
        activeInteraction = null;
        for (const c of activeCharacters) {
          if (c.talkingTimeout) {
            clearTimeout(c.talkingTimeout);
            c.talkingTimeout = null;
          }
          c.state = 'Resting';
          c.animationState = 'idle';
          c.facing = 'right';
        }
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
