/**
 * Shared provider presets and translation modes.
 *
 * Single source of truth for BOTH the first-run SetupWizard and the
 * SettingsModal. Keeping them separate previously caused the wizard to
 * offer only 6 providers and 2 translation modes while Settings offered 9
 * providers and 4 modes.
 */

export interface ProviderPreset {
  id: string;
  name: string;
  nameKey?: string;
  providerName: string;
  apiUrl: string;
  placeholderModel: string;
  placeholderModelKey?: string;
  rpm: number;
  timeout: number;
  hint: string;
  hintKey?: string;
}

export const PRESETS: ProviderPreset[] = [
  {
    id: 'openrouter',
    name: 'OpenRouter',
    providerName: 'openrouter',
    apiUrl: 'https://openrouter.ai/api/v1',
    placeholderModel: 'e.g. deepseek/deepseek-chat, google/gemini-2.0-flash-001, anthropic/claude-3.5-sonnet...',
    rpm: 60,
    timeout: 120,
    hint: 'Multi-provider aggregated gateway',
    hintKey: 'settings.presetHintOpenRouter',
  },
  {
    id: 'openai',
    name: 'OpenAI (Native)',
    providerName: 'openai',
    apiUrl: 'https://api.openai.com/v1',
    placeholderModel: 'e.g. gpt-4o, gpt-4o-mini, o3-mini...',
    rpm: 60,
    timeout: 120,
    hint: 'Official OpenAI endpoint',
    hintKey: 'settings.presetHintOpenAI',
  },
  {
    id: 'gemini',
    name: 'Google Gemini (Native)',
    providerName: 'gemini',
    apiUrl: 'https://generativelanguage.googleapis.com/v1beta',
    placeholderModel: 'e.g. gemini-2.0-flash, gemini-1.5-pro...',
    rpm: 60,
    timeout: 120,
    hint: 'Google Generative Language REST API',
  },
  {
    id: 'anthropic',
    name: 'Claude (Anthropic)',
    providerName: 'anthropic',
    apiUrl: 'https://api.anthropic.com/v1',
    placeholderModel: 'e.g. claude-3-5-sonnet-20241022, claude-3-7-sonnet...',
    rpm: 50,
    timeout: 120,
    hint: 'Anthropic Messages API Native',
  },
  {
    id: 'deepseek',
    name: 'DeepSeek Official',
    providerName: 'deepseek',
    apiUrl: 'https://api.deepseek.com/v1',
    placeholderModel: 'e.g. deepseek-chat, deepseek-reasoner...',
    rpm: 60,
    timeout: 120,
    hint: 'Official DeepSeek API',
    hintKey: 'settings.presetHintDeepSeek',
  },
  {
    id: 'vilao',
    name: 'ViLao AI',
    providerName: 'vilao',
    apiUrl: 'https://api.vilao.ai',
    placeholderModel: 'e.g. vilao-v1, vilao-7b-novel...',
    rpm: 60,
    timeout: 120,
    hint: 'ViLao dedicated translation API (https://api.vilao.ai)',
    hintKey: 'settings.presetHintViLao',
  },
  {
    id: 'ollama',
    name: 'Ollama (Local)',
    providerName: 'ollama',
    apiUrl: 'http://localhost:11434/v1',
    placeholderModel: 'e.g. qwen2.5:14b, llama3.3:70b...',
    rpm: 120,
    timeout: 180,
    hint: 'Run local offline models on your machine',
    hintKey: 'settings.presetHintOllama',
  },
  {
    id: 'custom_openai',
    name: 'OpenAI Compatible (Custom)',
    nameKey: 'settings.presetOpenAICustom',
    providerName: 'openai_compatible',
    apiUrl: '',
    placeholderModel: 'Enter model on your server (vLLM, LMStudio, Groq, Together, Mistral...)',
    placeholderModelKey: 'settings.presetPlaceholderOpenAICustom',
    rpm: 60,
    timeout: 120,
    hint: 'Compatible with all standard OpenAI API servers',
    hintKey: 'settings.presetHintOpenAICustom',
  },
  {
    id: 'custom_google',
    name: 'Google API (Custom)',
    nameKey: 'settings.presetGoogleCustom',
    providerName: 'google_compatible',
    apiUrl: 'https://generativelanguage.googleapis.com/v1beta',
    placeholderModel: 'Enter Google Gemini model name...',
    placeholderModelKey: 'settings.presetPlaceholderGoogleCustom',
    rpm: 60,
    timeout: 120,
    hint: 'Custom endpoint compatible with Google Generative Language',
    hintKey: 'settings.presetHintGoogleCustom',
  },
  {
    id: 'custom_other',
    name: 'Other (Custom Provider)',
    nameKey: 'settings.presetOtherCustom',
    providerName: 'custom',
    apiUrl: '',
    placeholderModel: 'Enter any model name...',
    placeholderModelKey: 'settings.presetPlaceholderOtherCustom',
    rpm: 60,
    timeout: 120,
    hint: 'Freely configure URL and custom provider headers',
    hintKey: 'settings.presetHintOtherCustom',
  },
];

/**
 * Translation pipeline modes. These MUST stay in sync with the backend
 * engine's accepted mode values.
 */
export interface TranslationModeOption {
  id: string;
  labelKey?: string;
  label?: string;
  descriptionKey?: string;
  badge: string;
}

export const TRANSLATION_MODES: TranslationModeOption[] = [
  {
    id: 'concurrent_dual_agent',
    labelKey: 'settings.modeDualAgent',
    label: 'Dual-Agent (Critic)',
    badge: 'Balanced',
  },
  {
    id: 'hierarchical_3pass',
    label: 'Hierarchical 3-Pass (Draft → Critic → Polish)',
    badge: 'Deep',
  },
  {
    id: 'single_pass',
    labelKey: 'settings.modeSinglePass',
    label: 'Single Pass',
    badge: 'Fast',
  },
  {
    id: 'swarm_arc_parallel',
    labelKey: 'settings.modeSwarmArc',
    label: 'Swarm Arc',
    badge: 'Volume',
  },
];

export const DEFAULT_TRANSLATION_MODE = 'concurrent_dual_agent';
