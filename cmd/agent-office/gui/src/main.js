// ─── State ────────────────────────────────────────────────────────────────────
let socket = null;
let totalTokens = 0;
let totalSteps = 0;
let runStartTime = null;
let currentRunID = null;
let costPer1k = 0; // loaded from /api/config
let editingAgentName = null; // null if creating, non-null if editing
let currentRunState = 'QUEUED';
let currentLanguage = 'en';

const translations = {
  en: {
    logo_subtitle: "Companion GUI",
    btn_help_title: "How to use this UI",
    btn_lang_toggle_title: "Switch Language / 切換語言",
    run_id_label: "Run ID:",
    run_state_label: "State:",
    tab_run: "Run Thread",
    tab_logs: "Logs",
    tab_agents: "Agents",
    logs_header: "Session Logs",
    category_workforce: "Workforce",
    category_execution: "Execution",
    category_costs_logs: "Costs & Logs",
    btn_sidebar_toggle_title: "Toggle Sidebar",
    run_thread_header: "Run Thread",
    init_notice: "WebSocket connected. Run thread is waiting to stream events...",
    telemetry_header: "Telemetry Dashboard",
    metric_steps: "Execution Steps",
    metric_tokens: "Total LLM Tokens",
    metric_latency: "Elapsed Latency",
    metric_stage: "Active Stage",
    metric_cost: "Est. Cost (USD)",
    metric_session_log: "Session Log",
    metric_active_agents: "Active Agents",
    metric_idle_agents: "Idle Agents",
    metric_total_agents: "Total Agents",
    btn_view_log_title: "View Session Log",
    controls_header: "Supervisor Interruption Control",
    btn_interrupt: "Interrupt Run",
    btn_abort: "Abort Run",
    agents_header: "Configured Agents",
    btn_add_agent: "Add Agent",
    agent_form_title_new: "New Agent",
    agent_form_title_edit: "Edit Agent",
    label_agent_name: "Name",
    label_agent_role: "Role",
    label_agent_color: "Theme Color",
    label_agent_avatar: "Avatar (Emoji or URL)",
    label_agent_provider: "AI Provider",
    label_agent_token: "API Key / Token",
    label_agent_model: "Model Name",
    label_agent_backstory: "Backstory",
    placeholder_agent_name: "e.g. qa-engineer",
    placeholder_agent_role: "e.g. QA Engineer",
    placeholder_agent_avatar: "e.g. 🤖 or image URL...",
    placeholder_agent_token: "Enter API Key to Login (Optional)",
    placeholder_agent_model: "e.g. gpt-4o, gemini-1.5-pro",
    placeholder_agent_backstory: "Describe this agent's background and expertise...",
    btn_submit_agent_edit: "Save Changes",
    btn_cancel_agent: "Cancel",
    btn_test_token: "Test API Key",
    th_name: "Name",
    th_role: "Role",
    th_backstory: "Backstory",
    th_actions: "Actions",
    agents_loading: "Loading agents...",
    agents_empty: "No agents configured. Click \"Add Agent\" to create one.",
    modal_log_title: "Session Log",
    help_modal_title: "How to use Companion GUI",
    help_step_1_title: "1. Setup your Team (Agents Tab)",
    help_step_1_desc: "Before running any tasks, go to the Agents tab and add your agents (e.g., planner, coder, reviewer). If no agents are configured, the run will fail immediately.",
    help_step_2_title: "2. Apply Changes (Terminal)",
    help_step_2_desc: "Changes to agents or models are saved to <code>agent-office.yaml</code>. To apply these changes and start a new run, you must restart the <code>.\\agent-office.exe gui</code> process in your terminal.",
    help_step_3_title: "3. Monitor the Run Thread",
    help_step_3_desc: "Watch the agents collaborate in real-time. The Telemetry Dashboard tracks your active token usage, estimated cost, and latency for the current session.",
    help_step_4_title: "4. Supervisor Interruption",
    help_step_4_desc: "If the agents go off track, click Interrupt Run. The process will pause, allowing you to type guidance into the feedback box. Click Resume Execution to inject your feedback and steer the agents.",
    step_loading: "Loading...",
    new_messages_badge: "New messages below ↓"
  },
  zh: {
    logo_subtitle: "協同面板",
    btn_help_title: "如何使用此界面",
    btn_lang_toggle_title: "切換語言 / Switch Language",
    run_id_label: "運行 ID:",
    run_state_label: "狀態:",
    tab_run: "運行線程",
    tab_logs: "日誌",
    tab_agents: "智能體配置",
    logs_header: "運行日誌",
    category_workforce: "團隊成員",
    category_execution: "執行狀態",
    category_costs_logs: "成本與日誌",
    btn_sidebar_toggle_title: "折疊/顯示側邊欄",
    run_thread_header: "運行線程",
    init_notice: "WebSocket 已連線。正在等待串流事件...",
    telemetry_header: "數據面板 (Telemetry)",
    metric_steps: "執行步驟",
    metric_tokens: "總 Token 消耗",
    metric_latency: "總消耗時長",
    metric_stage: "當前階段",
    metric_cost: "估算成本 (USD)",
    metric_session_log: "工作日誌",
    metric_active_agents: "使用中智能體",
    metric_idle_agents: "閒置中智能體",
    metric_total_agents: "智能體總數",
    btn_view_log_title: "查看工作日誌",
    controls_header: "督導中斷控制 (Supervisor)",
    btn_interrupt: "中斷運行",
    btn_abort: "終止運行",
    agents_header: "配置的智能體列表",
    btn_add_agent: "新增智能體",
    agent_form_title_new: "新增智能體",
    agent_form_title_edit: "編輯智能體",
    label_agent_name: "名稱",
    label_agent_role: "角色",
    label_agent_color: "主題顏色",
    label_agent_avatar: "頭像 (Emoji 或 圖片網址)",
    label_agent_provider: "AI 提供商",
    label_agent_token: "API 密鑰 / Token",
    label_agent_model: "模型名稱",
    label_agent_backstory: "背景設定",
    placeholder_agent_name: "例如 qa-engineer",
    placeholder_agent_role: "例如 測試工程師",
    placeholder_agent_avatar: "例如 🤖 或圖片網址...",
    placeholder_agent_token: "輸入 API 密鑰以登入 (選填)",
    placeholder_agent_model: "例如 gpt-4o, gemini-1.5-pro",
    placeholder_agent_backstory: "描述此智能體的背景背景與專業知識...",
    btn_submit_agent_edit: "儲存修改",
    btn_cancel_agent: "取消",
    btn_test_token: "測試密鑰",
    th_name: "名稱",
    th_role: "角色",
    th_backstory: "背景設定",
    th_actions: "操作",
    agents_loading: "正在載入智能體...",
    agents_empty: "未配置智能體。請點擊「新增智能體」進行創建。",
    modal_log_title: "工作日誌",
    help_modal_title: "如何使用協同面板",
    help_step_1_title: "1. 配置您的團隊 (智能體配置分頁)",
    help_step_1_desc: "在運行任何任務之前，請前往「智能體配置」分頁並添加您的智能體 (如 planner, coder, reviewer)。如果未配置任何智能體，運行將立即失敗。",
    help_step_2_title: "2. 應用配置 (終端機)",
    help_step_2_desc: "對智能體或模型的更改會儲存到 <code>agent-office.yaml</code>。要應用這些更改並開始新的運行，您必須在終端機中重啟 <code>.\\agent-office.exe gui</code> 進程。",
    help_step_3_title: "3. 監控運行線程",
    help_step_3_desc: "實時觀察智能體的協同討論。數據面板會跟蹤當前會話的 active token 使用量、估算成本和延遲時長。",
    help_step_4_title: "4. 督導中斷與引導",
    help_step_4_desc: "如果智能體討論偏離軌道，點擊「中斷運行」。進程將暫停，允許您在反饋框中輸入引導。點擊「恢復執行」以注入您的反饋並引導智能體。",
    step_loading: "載入中...",
    new_messages_badge: "下方有新訊息 ↓"
  }
};

function applyTranslations() {
  const elements = document.querySelectorAll('[data-i18n]');
  elements.forEach(el => {
    const key = el.getAttribute('data-i18n');
    const trans = translations[currentLanguage][key];
    if (trans) {
      el.innerHTML = trans;
    }
  });

  const placeholders = document.querySelectorAll('[data-i18n-placeholder]');
  placeholders.forEach(el => {
    const key = el.getAttribute('data-i18n-placeholder');
    const trans = translations[currentLanguage][key];
    if (trans) {
      el.placeholder = trans;
    }
  });

  const titles = document.querySelectorAll('[data-i18n-title]');
  titles.forEach(el => {
    const key = el.getAttribute('data-i18n-title');
    const trans = translations[currentLanguage][key];
    if (trans) {
      el.title = trans;
    }
  });

  updateGuidanceAreaTexts();
  stepCounter.textContent = currentLanguage === 'zh' ? `${totalSteps} 步` : `${totalSteps} steps`;
}

function setupLangToggle() {
  const btn = document.getElementById('btn-lang-toggle');
  if (btn) {
    btn.addEventListener('click', () => {
      currentLanguage = currentLanguage === 'en' ? 'zh' : 'en';
      applyTranslations();
    });
  }
}

// ─── DOM refs ─────────────────────────────────────────────────────────────────
const statusDot    = document.getElementById('status-dot');
const statusText   = document.getElementById('status-text');
const stateVal     = document.getElementById('run-state-val');
const runIdVal     = document.getElementById('run-id-val');
const stepCounter  = document.getElementById('step-counter');
const threadView   = document.getElementById('thread-view');

// Telemetry
const metricSteps   = document.getElementById('metric-steps');
const metricTokens  = document.getElementById('metric-tokens');
const metricLatency = document.getElementById('metric-latency');
const metricStage   = document.getElementById('metric-stage');
const metricCost    = document.getElementById('metric-cost');
const metricActiveAgents = document.getElementById('metric-active-agents');
const metricIdleAgents   = document.getElementById('metric-idle-agents');
const metricTotalAgents  = document.getElementById('metric-total-agents');
const sessionLogPath = document.getElementById('session-log-path');
const btnViewLog    = document.getElementById('btn-view-log');

// Modal
const logModal      = document.getElementById('log-modal');
const btnCloseModal = document.getElementById('btn-close-modal');
const logContentPre = document.getElementById('log-content-pre');

const helpModal     = document.getElementById('help-modal');
const btnHelp       = document.getElementById('btn-help');
const btnCloseHelp  = document.getElementById('btn-close-help');

// Controls
const btnInterrupt    = document.getElementById('btn-interrupt');
const btnAbort        = document.getElementById('btn-abort');
const btnResume       = document.getElementById('btn-resume');
const guidanceSection = document.getElementById('guidance-section');
const guidanceInput   = document.getElementById('guidance-input');

// Tabs
const tabRun    = document.getElementById('tab-run');
const tabLogs   = document.getElementById('tab-logs');
const tabAgents = document.getElementById('tab-agents');
const viewRun   = document.getElementById('view-run');
const viewLogs  = document.getElementById('view-logs');
const viewAgents = document.getElementById('view-agents');
const tabLogContentPre = document.getElementById('tab-log-content-pre');

// Agents panel
const btnAddAgent    = document.getElementById('btn-add-agent');
const addAgentForm   = document.getElementById('add-agent-form');
const inputName      = document.getElementById('input-agent-name');
const inputRole      = document.getElementById('input-agent-role');
const inputProvider  = document.getElementById('input-agent-provider');
const inputToken     = document.getElementById('input-agent-token');
const inputModel     = document.getElementById('input-agent-model');
const inputBackstory = document.getElementById('input-agent-backstory');
const btnSubmitAgent = document.getElementById('btn-submit-agent');
const btnCancelAgent = document.getElementById('btn-cancel-agent');
const btnTestToken     = document.getElementById('btn-test-token');
const testTokenResult  = document.getElementById('test-token-result');
const agentsTbody    = document.getElementById('agents-tbody');
const agentFormError = document.getElementById('agent-form-error');

// ─── Startup ──────────────────────────────────────────────────────────────────
window.addEventListener('DOMContentLoaded', () => {
  loadConfig();
  initTotalAgents();
  connect();
  setupTabs();
  setupAgentForm();
  setupModal();
  setupLangToggle();
  setupSmartScroll();
  setupAccordions();
  setupSidebarToggle();
});

async function initTotalAgents() {
  try {
    const res = await fetch('/api/agents');
    if (!res.ok) return;
    const data = await res.json();
    const agents = data.agents || [];
    metricTotalAgents.textContent = agents.length;
    metricIdleAgents.textContent = agents.length;
    metricActiveAgents.textContent = 0;
  } catch (e) {
    console.warn('Could not load total agents:', e);
  }
}

// ─── Config ───────────────────────────────────────────────────────────────────
async function loadConfig() {
  try {
    const res = await fetch('/api/config');
    if (!res.ok) return;
    const data = await res.json();
    costPer1k = data.cost_per_1k || 0;
  } catch (e) {
    console.warn('Could not load /api/config:', e);
  }
}

// ─── WebSocket ────────────────────────────────────────────────────────────────
function connect() {
  const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const host = window.location.host || 'localhost:8080';
  const wsUrl = `${wsProtocol}//${host}/ws`;

  statusDot.className = 'dot dot-reconnecting';
  statusText.textContent = 'Connecting...';

  socket = new WebSocket(wsUrl);

  socket.onopen = () => {
    statusDot.className = 'dot dot-connected';
    statusText.textContent = 'Connected';
    console.log('WebSocket connection established.');
  };

  socket.onclose = () => {
    statusDot.className = 'dot dot-disconnected';
    statusText.textContent = 'Disconnected. Retrying...';
    disableAllControls();
    setTimeout(connect, 1500);
  };

  socket.onerror = (err) => {
    console.error('WebSocket error:', err);
  };

  socket.onmessage = (event) => {
    const data = JSON.parse(event.data);
    handleServerEvent(data);
  };
}

function sendCommand(type, message = '') {
  if (!socket || socket.readyState !== WebSocket.OPEN) return;
  const cmd = { type, run_id: currentRunID || '', message };
  socket.send(JSON.stringify(cmd));
}

// ─── Controls ─────────────────────────────────────────────────────────────────
btnInterrupt.addEventListener('click', () => sendCommand('run.interrupt'));
btnAbort.addEventListener('click', () => sendCommand('run.abort'));
btnResume.addEventListener('click', () => {
  const guidance = guidanceInput.value.trim();

  // Disable inputs and buttons
  btnResume.disabled = true;
  guidanceInput.disabled = true;
  btnResume.classList.add('btn-loading');

  const textEl = document.getElementById('btn-resume-text');
  const isQueued = currentRunState === 'QUEUED';
  if (textEl) {
    if (isQueued) {
      textEl.textContent = currentLanguage === 'zh' ? '啟動中...' : 'Starting...';
    } else {
      textEl.textContent = currentLanguage === 'zh' ? '恢復中...' : 'Resuming...';
    }
  }

  if (isQueued) {
    sendCommand('run.start', guidance);
  } else {
    sendCommand('run.resume', guidance);
  }
  guidanceInput.value = '';
});

guidanceInput.addEventListener('keydown', (e) => {
  if (e.key === 'Enter' && !e.shiftKey) {
    e.preventDefault();
    if (!btnResume.disabled) {
      btnResume.click();
    }
  }
});

function disableAllControls() {
  btnInterrupt.disabled = true;
  btnAbort.disabled = true;
  btnResume.disabled = true;
  guidanceInput.disabled = true;
  guidanceSection.classList.add('disabled');
}

// ─── Server Event Router ──────────────────────────────────────────────────────
function handleServerEvent(evt) {
  if (evt.run_id) {
    currentRunID = evt.run_id;
    runIdVal.textContent = evt.run_id;
  }

  switch (evt.type) {
    case 'state.change':
      if (evt.content === 'RUNNING' && currentRunState === 'QUEUED') {
        threadView.innerHTML = '';
        totalTokens = 0;
        totalSteps = 0;
        metricSteps.textContent = '0';
        metricTokens.textContent = '0';
        metricCost.textContent = '$0.0000';
        metricLatency.textContent = '0.0s';
      }
      updateRunState(evt.content);
      appendStateCapsule(evt.content);
      break;
    case 'agent.speak':
      const isAgent = evt.sender.toLowerCase() !== 'user' && evt.sender.toLowerCase() !== 'supervisor' && evt.sender.toLowerCase() !== 'human';
      if (isAgent) {
        totalSteps++;
        stepCounter.textContent = `${totalSteps} steps`;
        metricSteps.textContent = totalSteps;
      }

      if (evt.metadata) {
        if (evt.metadata.tokens) {
          totalTokens += parseInt(evt.metadata.tokens, 10) || 0;
          metricTokens.textContent = totalTokens;
          updateCost();
        }
        if (evt.metadata.stage) {
          metricStage.textContent = evt.metadata.stage;
        }
        if (evt.metadata.latency) {
          metricLatency.textContent = `${evt.metadata.latency}s`;
        }
        if (evt.metadata.active_agents !== undefined) {
          metricActiveAgents.textContent = evt.metadata.active_agents;
        }
        if (evt.metadata.idle_agents !== undefined) {
          metricIdleAgents.textContent = evt.metadata.idle_agents;
        }
        if (evt.metadata.total_agents !== undefined) {
          metricTotalAgents.textContent = evt.metadata.total_agents;
        }
      }
      const color = (evt.metadata && evt.metadata.color) ? evt.metadata.color : null;
      const avatar = (evt.metadata && evt.metadata.avatar) ? evt.metadata.avatar : null;
      const provider = (evt.metadata && evt.metadata.provider) ? evt.metadata.provider : null;
      const model = (evt.metadata && evt.metadata.model) ? evt.metadata.model : null;
      appendMessage(evt.sender, evt.content, 'msg-agent', evt.timestamp, color, avatar, provider, model);
      break;
    case 'tool.call':
      appendMessage(evt.sender, `Executing tool: ${evt.content}`, 'msg-tool-call', evt.timestamp);
      break;
    case 'tool.return':
      appendMessage(evt.sender, `Returned: ${evt.content}`, 'msg-tool-return', evt.timestamp);
      break;
    case 'system.log':
      appendSystemLog(evt.content, evt.timestamp);
      break;
  }
}

// ─── Cost Estimation ──────────────────────────────────────────────────────────
function updateCost() {
  const estCost = (totalTokens / 1000) * costPer1k;
  metricCost.textContent = `$${estCost.toFixed(4)}`;
}

// ─── Session Log ──────────────────────────────────────────────────────────────
async function fetchAndShowSessionLog() {
  try {
    const res = await fetch('/api/session/latest');
    if (!res.ok) return;
    const data = await res.json();
    if (data.path) {
      sessionLogPath.textContent = `Saved: ${data.path}`;
      btnViewLog.classList.remove('hidden');
    }
  } catch (e) {
    console.warn('Could not fetch session log:', e);
  }
}

// ─── Modal ────────────────────────────────────────────────────────────────────
function setupModal() {
  btnViewLog.addEventListener('click', () => {
    switchTab('logs');
  });

  btnCloseModal.addEventListener('click', () => {
    logModal.classList.add('hidden');
  });

  // Close when clicking overlay
  logModal.addEventListener('click', (e) => {
    if (e.target === logModal) {
      logModal.classList.add('hidden');
    }
  });

  // Help Modal
  if(btnHelp) {
    btnHelp.addEventListener('click', () => {
      helpModal.classList.remove('hidden');
    });
  }
  if(btnCloseHelp) {
    btnCloseHelp.addEventListener('click', () => {
      helpModal.classList.add('hidden');
    });
  }
  if(helpModal) {
    helpModal.addEventListener('click', (e) => {
      if (e.target === helpModal) {
        helpModal.classList.add('hidden');
      }
    });
  }
}

// ─── Run State ────────────────────────────────────────────────────────────────
function updateGuidanceAreaTexts() {
  const isZh = currentLanguage === 'zh';
  const labelEl = document.getElementById('guidance-label');
  const inputEl = document.getElementById('guidance-input');
  const btnTextEl = document.getElementById('btn-resume-text');

  if (!labelEl || !inputEl || !btnTextEl) return;

  if (currentRunState === 'QUEUED') {
    labelEl.textContent = isZh ? '任務描述 / 提示詞' : 'Task Prompt';
    inputEl.placeholder = isZh ? '輸入任務提示詞以啟動工作流...' : 'Type a task prompt to launch the workforce...';
    btnTextEl.textContent = isZh ? '啟動任務' : 'Start Task';
  } else {
    labelEl.textContent = isZh ? '恢復引導與反饋' : 'Resume Guidance & Feedback';
    inputEl.placeholder = isZh ? '輸入引導反饋以注入線程...' : 'Type supervisor guidance to inject into the thread...';
    btnTextEl.textContent = isZh ? '恢復執行' : 'Resume Execution';
  }
}

function updateRunState(state) {
  currentRunState = state;
  stateVal.textContent = state;
  stateVal.className = 'value-state';

  if (btnResume) {
    btnResume.classList.remove('btn-loading');
  }

  switch (state) {
    case 'QUEUED':
      stateVal.classList.add('state-queued');
      btnInterrupt.disabled = true;
      btnAbort.disabled = true;
      btnResume.disabled = false;
      guidanceInput.disabled = false;
      guidanceSection.classList.remove('disabled');
      updateGuidanceAreaTexts();
      break;
    case 'RUNNING':
      stateVal.classList.add('state-running');
      btnInterrupt.disabled = false;
      btnAbort.disabled = false;
      btnResume.disabled = true;
      guidanceInput.disabled = true;
      guidanceSection.classList.add('disabled');
      break;
    case 'INTERRUPTING':
      stateVal.classList.add('state-interrupting');
      btnInterrupt.disabled = true;
      btnAbort.disabled = false;
      btnResume.disabled = true;
      guidanceInput.disabled = true;
      guidanceSection.classList.add('disabled');
      break;
    case 'INTERRUPTED':
      stateVal.classList.add('state-interrupted');
      btnInterrupt.disabled = true;
      btnAbort.disabled = false;
      btnResume.disabled = false;
      guidanceInput.disabled = false;
      guidanceSection.classList.remove('disabled');
      guidanceInput.focus();
      updateGuidanceAreaTexts();
      break;
    case 'RESUMING':
      stateVal.classList.add('state-resuming');
      btnInterrupt.disabled = true;
      btnAbort.disabled = true;
      btnResume.disabled = true;
      guidanceInput.disabled = true;
      guidanceSection.classList.add('disabled');
      break;
    case 'COMPLETED':
    case 'CANCELLED':
      stateVal.classList.add(state === 'COMPLETED' ? 'state-completed' : 'state-cancelled');
      disableAllControls();
      // Fetch session log after a short delay for the file to be written
      setTimeout(fetchAndShowSessionLog, 800);
      break;
    case 'FAILED':
      stateVal.classList.add('state-failed');
      disableAllControls();
      if (document.querySelectorAll('.message-bubble, .system-message').length === 0) {
        appendSystemLog('Error: No agents configured in workspace. Please add agents first.', Math.floor(Date.now() / 1000));
      }
      break;
  }

  if (state === 'QUEUED' || state === 'COMPLETED' || state === 'CANCELLED' || state === 'FAILED') {
    if (metricActiveAgents && metricIdleAgents && metricTotalAgents) {
      metricActiveAgents.textContent = '0';
      metricIdleAgents.textContent = metricTotalAgents.textContent;
    }
  }
}

// ─── Tab switching ────────────────────────────────────────────────────────────
function setupTabs() {
  tabRun.addEventListener('click', () => switchTab('run'));
  tabLogs.addEventListener('click', () => switchTab('logs'));
  tabAgents.addEventListener('click', () => switchTab('agents'));
}

function switchTab(tab) {
  tabRun.classList.remove('tab-btn--active');
  tabLogs.classList.remove('tab-btn--active');
  tabAgents.classList.remove('tab-btn--active');
  viewRun.classList.remove('tab-view--active');
  viewLogs.classList.remove('tab-view--active');
  viewAgents.classList.remove('tab-view--active');

  if (tab === 'run') {
    tabRun.classList.add('tab-btn--active');
    viewRun.classList.add('tab-view--active');
  } else if (tab === 'logs') {
    tabLogs.classList.add('tab-btn--active');
    viewLogs.classList.add('tab-view--active');
    loadLatestSessionLogContent();
  } else if (tab === 'agents') {
    tabAgents.classList.add('tab-btn--active');
    viewAgents.classList.add('tab-view--active');
    loadAgents();
  }
}

// ─── Agents Panel ─────────────────────────────────────────────────────────────
function setupAgentForm() {
  const drawerBackdrop = document.getElementById('drawer-backdrop');
  if (drawerBackdrop) {
    drawerBackdrop.addEventListener('click', () => {
      hideAgentForm();
    });
  }

  btnAddAgent.addEventListener('click', () => {
    editingAgentName = null;
    document.getElementById('agent-form-title').textContent = currentLanguage === 'zh' ? '新增智能體' : 'New Agent';
    btnSubmitAgent.textContent = currentLanguage === 'zh' ? '創建智能體' : 'Create Agent';
    addAgentForm.classList.remove('hidden');
    addAgentForm.classList.add('slide-in');
    if (drawerBackdrop) drawerBackdrop.classList.remove('hidden');
    btnAddAgent.disabled = true;
    inputName.focus();
  });

  btnCancelAgent.addEventListener('click', () => {
    hideAgentForm();
  });

  btnSubmitAgent.addEventListener('click', () => {
    submitAgent();
  });

  inputProvider.addEventListener('change', () => {
    const provider = inputProvider.value;
    const defaultModels = {
      openai: 'gpt-4o',
      gemini: 'gemini-1.5-flash',
      anthropic: 'claude-3-haiku',
      openrouter: 'google/gemini-flash-1.5'
    };
    if (provider && defaultModels[provider]) {
      inputModel.value = defaultModels[provider];
    }
  });

  if (btnTestToken) {
    btnTestToken.addEventListener('click', async () => {
      const provider = inputProvider.value;
      const token = inputToken.value.trim();
      const model = inputModel.value.trim();

      testTokenResult.className = 'test-result';
      testTokenResult.classList.remove('hidden');
      testTokenResult.textContent = currentLanguage === 'zh' ? '正在測試連線...' : 'Testing connection...';
      testTokenResult.style.color = 'var(--text-secondary)';

      if (!provider) {
        testTokenResult.className = 'test-result error';
        testTokenResult.textContent = currentLanguage === 'zh' ? '請先選擇 AI 提供商。' : 'Please select an AI Provider first.';
        return;
      }
      if (!token) {
        testTokenResult.className = 'test-result error';
        testTokenResult.textContent = currentLanguage === 'zh' ? '請先輸入 API 密鑰。' : 'Please enter an API Key first.';
        return;
      }

      const originalText = btnTestToken.textContent;
      btnTestToken.textContent = currentLanguage === 'zh' ? '測試中...' : 'Testing...';
      btnTestToken.classList.add('btn-loading');
      btnTestToken.disabled = true;
      inputToken.disabled = true;
      inputProvider.disabled = true;
      inputModel.disabled = true;

      try {
        const res = await fetch('/api/agents/test', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ provider, token, model })
        });
        const data = await res.json();
        if (res.ok && data.success) {
          testTokenResult.className = 'test-result success';
          testTokenResult.textContent = currentLanguage === 'zh' ? '連線成功！密鑰有效。' : 'Connection successful! API Key is valid.';
        } else {
          testTokenResult.className = 'test-result error';
          testTokenResult.textContent = (currentLanguage === 'zh' ? '連線失敗: ' : 'Connection failed: ') + (data.error || 'Unauthorized');
        }
      } catch (e) {
        testTokenResult.className = 'test-result error';
        testTokenResult.textContent = (currentLanguage === 'zh' ? '連線錯誤: ' : 'Connection error: ') + e.message;
      } finally {
        btnTestToken.disabled = false;
        inputToken.disabled = false;
        inputProvider.disabled = false;
        inputModel.disabled = false;
        btnTestToken.classList.remove('btn-loading');
        btnTestToken.textContent = originalText;
      }
    });
  }

  // Event delegation for edit and delete buttons on the table body
  agentsTbody.addEventListener('click', (e) => {
    const btnEdit = e.target.closest('.btn-edit');
    const btnDelete = e.target.closest('.btn-delete');
    if (btnEdit) {
      const name = btnEdit.getAttribute('data-name');
      const role = btnEdit.getAttribute('data-role');
      const backstory = btnEdit.getAttribute('data-backstory');
      const color = btnEdit.getAttribute('data-color');
      const avatar = btnEdit.getAttribute('data-avatar');
      const provider = btnEdit.getAttribute('data-provider');
      const model = btnEdit.getAttribute('data-model');
      const token = btnEdit.getAttribute('data-token');
      startEditAgent(name, role, backstory, color, avatar, provider, model, token);
    } else if (btnDelete) {
      const name = btnDelete.getAttribute('data-name');
      deleteAgent(name);
    }
  });
}

function startEditAgent(name, role, backstory, color, avatar, provider, model, token) {
  editingAgentName = name;
  addAgentForm.classList.remove('hidden');
  addAgentForm.classList.add('slide-in');
  const drawerBackdrop = document.getElementById('drawer-backdrop');
  if (drawerBackdrop) drawerBackdrop.classList.remove('hidden');
  btnAddAgent.disabled = true;
  
  // Prefill fields
  inputName.value = name;
  inputRole.value = role;
  inputBackstory.value = backstory;
  inputProvider.value = provider || '';
  inputToken.value = token || '';
  inputModel.value = model || '';
  
  const selectColor = document.getElementById('input-agent-color');
  if (selectColor) {
    selectColor.value = color || '#6366f1';
  }
  const inputAvatar = document.getElementById('input-agent-avatar');
  if (inputAvatar) {
    inputAvatar.value = avatar || '';
  }
  
  // Update UI texts
  document.getElementById('agent-form-title').textContent = currentLanguage === 'zh' ? '編輯智能體' : 'Edit Agent';
  btnSubmitAgent.textContent = currentLanguage === 'zh' ? '儲存修改' : 'Save Changes';
  
  inputName.focus();
}

async function deleteAgent(name) {
  const confirmMsg = currentLanguage === 'zh' ? `確定要刪除智能體 "${name}" 嗎？` : `Are you sure you want to delete agent "${name}"?`;
  if (!confirm(confirmMsg)) {
    return;
  }
  try {
    const res = await fetch(`/api/agents?name=${encodeURIComponent(name)}`, {
      method: 'DELETE',
    });
    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || `HTTP ${res.status}`);
    }
    await loadAgents();
  } catch (e) {
    const errMsg = currentLanguage === 'zh' ? `刪除智能體失敗: ${e.message}` : `Failed to delete agent: ${e.message}`;
    alert(errMsg);
  }
}

function hideAgentForm() {
  // Blur any active element inside the form before sliding it out
  // to prevent the browser from automatically scrolling the page horizontally
  if (document.activeElement && addAgentForm.contains(document.activeElement)) {
    document.activeElement.blur();
  }

  addAgentForm.classList.add('hidden');
  addAgentForm.classList.remove('slide-in');
  const drawerBackdrop = document.getElementById('drawer-backdrop');
  if (drawerBackdrop) drawerBackdrop.classList.add('hidden');
  agentFormError.classList.add('hidden');
  agentFormError.textContent = '';
  inputName.value = '';
  inputRole.value = '';
  inputProvider.value = '';
  inputToken.value = '';
  inputModel.value = '';
  inputBackstory.value = '';
  const selectColor = document.getElementById('input-agent-color');
  if (selectColor) {
    selectColor.value = '#6366f1';
  }
  const inputAvatar = document.getElementById('input-agent-avatar');
  if (inputAvatar) {
    inputAvatar.value = '';
  }
  btnAddAgent.disabled = false;
  
  if (testTokenResult) {
    testTokenResult.classList.add('hidden');
    testTokenResult.textContent = '';
  }

  // Reset UI texts
  document.getElementById('agent-form-title').textContent = currentLanguage === 'zh' ? '新增智能體' : 'New Agent';
  btnSubmitAgent.textContent = currentLanguage === 'zh' ? '創建智能體' : 'Create Agent';
  editingAgentName = null;

  // Enforce resetting any horizontal scroll shifts
  setTimeout(() => {
    window.scrollTo({ left: 0 });
    document.body.scrollLeft = 0;
    document.documentElement.scrollLeft = 0;
    const appContainer = document.querySelector('.app-container');
    if (appContainer) appContainer.scrollLeft = 0;
  }, 100);
}

async function loadAgents() {
  const loadingMsg = currentLanguage === 'zh' ? '正在載入智能體...' : 'Loading agents...';
  agentsTbody.innerHTML = `<div class="agents-empty">${loadingMsg}</div>`;
  try {
    const res = await fetch('/api/agents');
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    const data = await res.json();
    renderAgentsTable(data.agents || []);
  } catch (e) {
    const errMsg = currentLanguage === 'zh' ? '載入智能體失敗: ' + e.message : 'Failed to load agents: ' + e.message;
    agentsTbody.innerHTML = `<div class="agents-empty">${errMsg}</div>`;
  }
}

function renderAgentsTable(agents) {
  if (!agents || agents.length === 0) {
    const emptyMsg = currentLanguage === 'zh' ? '未配置智能體。請點擊「新增智能體」進行創建。' : 'No agents configured. Click "Add Agent" to create one.';
    agentsTbody.innerHTML = `<div class="agents-empty">${emptyMsg}</div>`;
    return;
  }
  const editTitle = currentLanguage === 'zh' ? '編輯智能體' : 'Edit Agent';
  const deleteTitle = currentLanguage === 'zh' ? '刪除智能體' : 'Delete Agent';
  agentsTbody.innerHTML = agents.map(a => {
    const activeColor = a.color || '#6366f1';
    const glowColor = hexToRgba(activeColor, 0.15);
    const borderGlow = hexToRgba(activeColor, 0.3);
    const activeAvatar = a.avatar || getDefaultAvatar(a.name);
    const isUrl = activeAvatar && (activeAvatar.startsWith('http://') || activeAvatar.startsWith('https://') || activeAvatar.startsWith('/') || activeAvatar.startsWith('./'));
    const avatarHtml = isUrl 
      ? `<img src="${escapeHtml(activeAvatar)}" alt="${escapeHtml(a.name)}">`
      : `<span>${escapeHtml(activeAvatar)}</span>`;
      
    const providerBadge = a.provider ? `<span class="agent-card-provider-badge">${escapeHtml(a.provider)}${a.model ? ` / ${escapeHtml(a.model)}` : ''}</span>` : '';
    
    return `
    <div class="agent-card" style="--agent-color: ${activeColor}; --agent-glow: ${glowColor}; --agent-border: ${borderGlow};">
      <div class="agent-card-header">
        <div class="agent-card-avatar" style="border-color: ${activeColor}; box-shadow: 0 0 10px ${glowColor};">
          ${avatarHtml}
        </div>
        <div class="agent-card-meta">
          <div class="agent-card-name">${escapeHtml(a.name)}</div>
          <div class="agent-card-role" style="background: ${hexToRgba(activeColor, 0.12)}; border: 1px solid ${borderGlow}; color: ${activeColor};">${escapeHtml(a.role)}</div>
        </div>
        <div class="agent-card-actions">
          <button class="btn-icon btn-edit" title="${editTitle}" data-name="${escapeHtml(a.name)}" data-role="${escapeHtml(a.role)}" data-backstory="${escapeHtml(a.backstory || '')}" data-color="${escapeHtml(a.color || '')}" data-avatar="${escapeHtml(a.avatar || '')}" data-provider="${escapeHtml(a.provider || '')}" data-model="${escapeHtml(a.model || '')}" data-token="${escapeHtml(a.token || '')}">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M11 4H4a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2v-7"></path><path d="M18.5 2.5a2.121 2.121 0 1 1 3 3L12 15l-4 1 1-4 9.5-9.5z"></path></svg>
          </button>
          <button class="btn-icon btn-delete" title="${deleteTitle}" data-name="${escapeHtml(a.name)}" style="color: var(--danger-color); margin-left: 0.25rem;">
            <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="3 6 5 6 21 6"></polyline><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"></path><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line></svg>
          </button>
        </div>
      </div>
      <div class="agent-card-body">
        <p class="agent-card-backstory">${escapeHtml(a.backstory || '—')}</p>
      </div>
      <div class="agent-card-footer">
        <div class="agent-card-provider">
          <span class="provider-status-dot" style="background-color: ${activeColor}; box-shadow: 0 0 6px ${activeColor};"></span>
          ${providerBadge}
        </div>
      </div>
    </div>
    `;
  }).join('');
}

async function submitAgent() {
  const name = inputName.value.trim();
  const role = inputRole.value.trim();
  const backstory = inputBackstory.value.trim();
  const selectColor = document.getElementById('input-agent-color');
  const color = selectColor ? selectColor.value : '#6366f1';
  const inputAvatar = document.getElementById('input-agent-avatar');
  const avatar = inputAvatar ? inputAvatar.value.trim() : '';
  const provider = inputProvider.value;
  const token = inputToken.value.trim();
  const model = inputModel.value.trim();

  agentFormError.classList.add('hidden');

  if (!name || !role) {
    agentFormError.textContent = currentLanguage === 'zh' ? '名稱與角色為必填項。' : 'Name and Role are required.';
    agentFormError.classList.remove('hidden');
    return;
  }

  btnSubmitAgent.disabled = true;
  btnSubmitAgent.classList.add('btn-loading');
  btnSubmitAgent.textContent = editingAgentName 
    ? (currentLanguage === 'zh' ? '儲存中...' : 'Saving...') 
    : (currentLanguage === 'zh' ? '創建中...' : 'Creating...');

  // Disable all form inputs
  inputName.disabled = true;
  inputRole.disabled = true;
  inputProvider.disabled = true;
  inputToken.disabled = true;
  inputModel.disabled = true;
  inputBackstory.disabled = true;
  btnCancelAgent.disabled = true;
  if (btnTestToken) btnTestToken.disabled = true;

  try {
    let url = '/api/agents';
    let method = 'POST';
    let payload = { name, role, backstory, color, avatar, provider, token, model };
    
    if (editingAgentName) {
      method = 'PUT';
      payload.originalName = editingAgentName;
    }

    const res = await fetch(url, {
      method: method,
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(payload),
    });

    if (!res.ok) {
      const err = await res.json();
      throw new Error(err.error || `HTTP ${res.status}`);
    }

    hideAgentForm();
    await loadAgents();
  } catch (e) {
    agentFormError.textContent = (currentLanguage === 'zh' ? '錯誤: ' : 'Error: ') + e.message;
    agentFormError.classList.remove('hidden');
  } finally {
    btnSubmitAgent.disabled = false;
    btnSubmitAgent.classList.remove('btn-loading');
    btnSubmitAgent.textContent = editingAgentName 
      ? (currentLanguage === 'zh' ? '儲存修改' : 'Save Changes') 
      : (currentLanguage === 'zh' ? '創建智能體' : 'Create Agent');

    // Re-enable form inputs
    inputName.disabled = false;
    inputRole.disabled = false;
    inputProvider.disabled = false;
    inputToken.disabled = false;
    inputModel.disabled = false;
    inputBackstory.disabled = false;
    btnCancelAgent.disabled = false;
    if (btnTestToken) btnTestToken.disabled = false;
  }
}

// ─── Message Rendering ────────────────────────────────────────────────────────
function formatTime(timestamp) {
  if (!timestamp) return '';
  const date = new Date(timestamp * 1000);
  return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit', second: '2-digit' });
}

const BEAUTIFUL_COLORS = [
  '#6366f1', // Indigo
  '#a855f7', // Purple
  '#3b82f6', // Blue
  '#10b981', // Emerald
  '#f59e0b', // Amber
  '#f43f5e', // Rose
  '#0d9488'  // Teal
];

function getColorForName(name) {
  if (!name) return BEAUTIFUL_COLORS[0];
  let hash = 0;
  for (let i = 0; i < name.length; i++) {
    hash = name.charCodeAt(i) + ((hash << 5) - hash);
  }
  const index = Math.abs(hash) % BEAUTIFUL_COLORS.length;
  return BEAUTIFUL_COLORS[index];
}

function getDefaultAvatar(name) {
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
  if (nameLower.includes('supervisor') || nameLower.includes('user') || nameLower.includes('human')) {
    return '👤';
  }
  return '🤖';
}

function hexToRgba(hex, alpha) {
  if (!hex) return '';
  hex = hex.replace('#', '');
  if (hex.length === 3) {
    hex = hex.split('').map(char => char + char).join('');
  }
  const r = parseInt(hex.substring(0, 2), 16);
  const g = parseInt(hex.substring(2, 4), 16);
  const b = parseInt(hex.substring(4, 6), 16);
  return `rgba(${r}, ${g}, ${b}, ${alpha})`;
}

function appendMessage(sender, content, typeClass, timestamp, color, avatar, provider, model) {
  const isAgent = typeClass === 'msg-agent';
  const nameLower = (sender || '').toLowerCase();
  const isSupervisor = nameLower.includes('supervisor') || nameLower.includes('user') || nameLower.includes('human') || typeClass === 'msg-supervisor';

  // Determine actual color
  let activeColor = color;
  if (!activeColor) {
    if (isSupervisor) {
      activeColor = '#ec4899'; // Pink-ish for supervisor
    } else if (isAgent) {
      activeColor = getColorForName(sender);
    }
  }

  // Determine actual avatar
  let activeAvatar = avatar;
  if (!activeAvatar) {
    if (typeClass === 'msg-tool-call') {
      activeAvatar = '🛠️';
    } else if (typeClass === 'msg-tool-return') {
      activeAvatar = '📥';
    } else {
      activeAvatar = getDefaultAvatar(sender);
    }
  }

  // Create message row container
  const row = document.createElement('div');
  row.className = 'message-row';
  if (isSupervisor) {
    row.classList.add('message-row--supervisor');
  } else {
    row.classList.add('message-row--agent');
  }

  // Create avatar container
  const avatarDiv = document.createElement('div');
  avatarDiv.className = 'message-avatar';
  if (activeColor) {
    avatarDiv.style.borderColor = activeColor;
    avatarDiv.style.boxShadow = `0 0 8px ${hexToRgba(activeColor, 0.3)}`;
  }

  const isUrl = activeAvatar && (activeAvatar.startsWith('http://') || activeAvatar.startsWith('https://') || activeAvatar.startsWith('/') || activeAvatar.startsWith('./'));
  if (isUrl) {
    const img = document.createElement('img');
    img.src = activeAvatar;
    img.alt = sender;
    avatarDiv.appendChild(img);
  } else {
    const span = document.createElement('span');
    span.textContent = activeAvatar;
    avatarDiv.appendChild(span);
  }

  // Create bubble container
  const bubble = document.createElement('div');
  bubble.className = `message-bubble ${typeClass}`;
  
  if (activeColor) {
    bubble.style.background = `linear-gradient(135deg, ${hexToRgba(activeColor, 0.15)} 0%, ${hexToRgba(activeColor, 0.05)} 100%)`;
    bubble.style.borderColor = hexToRgba(activeColor, 0.3);
    bubble.style.borderStyle = 'solid';
    bubble.style.borderWidth = '1px';
  }

  // Create header
  const header = document.createElement('div');
  header.className = 'msg-header';
  if (activeColor) {
    header.style.color = activeColor;
  }

  const senderSpan = document.createElement('span');
  senderSpan.textContent = sender;
  header.appendChild(senderSpan);

  if (provider) {
    const modelTag = document.createElement('span');
    modelTag.className = 'msg-model-tag';
    modelTag.textContent = ` [${provider}${model ? ` / ${model}` : ''}]`;
    modelTag.style.opacity = '0.5';
    modelTag.style.fontSize = '0.75rem';
    modelTag.style.marginLeft = '0.5rem';
    header.appendChild(modelTag);
  }

  const timeSpan = document.createElement('span');
  timeSpan.className = 'msg-time';
  timeSpan.textContent = formatTime(timestamp);
  header.appendChild(timeSpan);

  // Create body
  const body = document.createElement('div');
  body.className = 'msg-body';
  body.textContent = content;

  bubble.appendChild(header);
  bubble.appendChild(body);

  row.appendChild(avatarDiv);
  row.appendChild(bubble);

  // Smart scroll check
  const isNearBottom = (threadView.scrollHeight - threadView.scrollTop - threadView.clientHeight) <= 100;
  threadView.appendChild(row);
  
  if (isNearBottom || isSupervisor) {
    scrollToBottom();
  } else {
    showNewMessagesIndicator();
  }
}

function showNewMessagesIndicator() {
  const indicator = document.getElementById('new-messages-indicator');
  if (indicator) {
    indicator.classList.remove('hidden');
  }
}

function setupSmartScroll() {
  const indicator = document.getElementById('new-messages-indicator');
  if (indicator) {
    indicator.addEventListener('click', () => {
      scrollToBottom();
      indicator.classList.add('hidden');
    });
  }

  threadView.addEventListener('scroll', () => {
    const isNearBottom = (threadView.scrollHeight - threadView.scrollTop - threadView.clientHeight) <= 50;
    if (isNearBottom && indicator) {
      indicator.classList.add('hidden');
    }
  });
}

function appendStateCapsule(state) {
  const isNearBottom = (threadView.scrollHeight - threadView.scrollTop - threadView.clientHeight) <= 100;
  const capsule = document.createElement('div');
  capsule.className = 'status-capsule';
  capsule.textContent = `Run State transitioned to: ${state}`;
  threadView.appendChild(capsule);
  
  if (isNearBottom) {
    scrollToBottom();
  } else {
    showNewMessagesIndicator();
  }
}

function appendSystemLog(message, timestamp) {
  const isNearBottom = (threadView.scrollHeight - threadView.scrollTop - threadView.clientHeight) <= 100;
  const logDiv = document.createElement('div');
  logDiv.className = 'system-message';
  logDiv.textContent = `[SYSTEM LOG - ${formatTime(timestamp)}] ${message}`;
  threadView.appendChild(logDiv);
  
  if (isNearBottom) {
    scrollToBottom();
  } else {
    showNewMessagesIndicator();
  }
}

function scrollToBottom() {
  threadView.scrollTop = threadView.scrollHeight;
}

function escapeHtml(str) {
  const d = document.createElement('div');
  d.appendChild(document.createTextNode(str));
  return d.innerHTML;
}

// ─── Telemetry Accordion & Collapsible Sidebar ─────────────────────────────────
function setupAccordions() {
  const headers = document.querySelectorAll('.accordion-header');
  headers.forEach(header => {
    header.addEventListener('click', () => {
      const targetId = header.getAttribute('data-target');
      const content = document.getElementById(targetId);
      if (content) {
        const isCollapsed = content.classList.toggle('collapsed');
        header.classList.toggle('collapsed', isCollapsed);
      }
    });
  });
}

function setupSidebarToggle() {
  const btn = document.getElementById('btn-sidebar-toggle');
  const appContent = document.querySelector('.app-content');
  if (btn && appContent) {
    btn.addEventListener('click', () => {
      appContent.classList.toggle('sidebar-collapsed');
    });
  }
}

async function loadLatestSessionLogContent() {
  if (!tabLogContentPre) return;
  tabLogContentPre.textContent = currentLanguage === 'zh' ? '正在載入日誌...' : 'Loading logs...';
  try {
    const res = await fetch('/api/session/latest/content');
    if (!res.ok) throw new Error('Failed to load');
    const json = await res.json();
    tabLogContentPre.textContent = JSON.stringify(json, null, 2);
  } catch (e) {
    tabLogContentPre.textContent = currentLanguage === 'zh' ? '載入日誌內容失敗。' : 'Error loading log content.';
  }
}
