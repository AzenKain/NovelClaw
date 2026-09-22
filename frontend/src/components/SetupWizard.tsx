import React, { useState, useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Globe,
  Cpu,
  Shield,
  Sparkles,
  ChevronRight,
  ChevronLeft,
  Check,
  AlertTriangle,
  Zap,
} from 'lucide-react';
import { useAppStore, FallbackTier } from '@/store/useAppStore';
import { PRESETS, TRANSLATION_MODES, DEFAULT_TRANSLATION_MODE, type ProviderPreset } from '@/constants/providers';
import * as dtos from '@bindings/novelclaw/internal/dtos/models';
import i18n from '@/i18n';
import { toast } from 'react-hot-toast';


const LANGUAGES = [
  { code: 'en', label: 'English', flag: '🇺🇸' },
  { code: 'vi', label: 'Tiếng Việt', flag: '🇻🇳' },
  { code: 'ja', label: '日本語', flag: '🇯🇵' },
  { code: 'zh', label: '中文', flag: '🇨🇳' },
  { code: 'ko', label: '한국어', flag: '🇰🇷' },
];

const STEPS = ['language', 'provider', 'fallback', 'preferences'] as const;
type WizardStep = (typeof STEPS)[number];

const STEP_ICONS: Record<WizardStep, React.ReactNode> = {
  language: <Globe className="w-4 h-4" />,
  provider: <Cpu className="w-4 h-4" />,
  fallback: <Shield className="w-4 h-4" />,
  preferences: <Sparkles className="w-4 h-4" />,
};

const STEP_KEYS: Record<WizardStep, string> = {
  language: 'wizard.stepLanguage',
  provider: 'wizard.stepProvider',
  fallback: 'wizard.stepFallback',
  preferences: 'wizard.stepPreferences',
};

export const SetupWizard: React.FC = () => {
  const { t } = useTranslation();
  const {
    saveLLMConfig,
    testLLMConnection,
    fetchLLMConfigs,
    llmConfigs,
    setFallbackChain,
    setTranslationMode,
    setThinkingEffort,
    completeSetupWizard,
  } = useAppStore();

  const [currentStep, setCurrentStep] = useState(0);
  const [selectedLang, setSelectedLang] = useState(i18n.language || 'en');

  // Provider form
  const [selectedPreset, setSelectedPreset] = useState<ProviderPreset>(PRESETS[0]);
  const [apiKey, setApiKey] = useState('');
  const [apiUrl, setApiUrl] = useState(PRESETS[0].apiUrl);
  const [modelName, setModelName] = useState(PRESETS[0].placeholderModel);
  const [isTesting, setIsTesting] = useState(false);
  const [isSaving, setIsSaving] = useState(false);
  const [testResult, setTestResult] = useState<dtos.TestConnectionResponse | null>(null);
  const [savedConfigs, setSavedConfigs] = useState<string[]>([]);

  // Preferences
  const [translationMode, setLocalTranslationMode] = useState(DEFAULT_TRANSLATION_MODE);
  const [thinkingEffort, setLocalThinkingEffort] = useState<'off' | 'low' | 'medium' | 'high' | 'xhigh' | 'max'>('medium');

  const step = STEPS[currentStep];

  const handleLanguageChange = useCallback((lang: string) => {
    setSelectedLang(lang);
    i18n.changeLanguage(lang);
    localStorage.setItem('neko_app_lang', lang);
    localStorage.setItem('i18nextLng', lang);
  }, []);

  const handleSelectPreset = useCallback((preset: ProviderPreset) => {
    setSelectedPreset(preset);
    setApiUrl(preset.apiUrl);
    setModelName(preset.placeholderModel);
    setTestResult(null);
    setApiKey('');
  }, []);

  const handleTestConnection = useCallback(async () => {
    setIsTesting(true);
    setTestResult(null);
    try {
      const res = await testLLMConnection({
        api_url: apiUrl.trim(),
        token: apiKey.trim(),
        model_name: modelName.trim(),
        config_id: '',
      });
      if (res) setTestResult(res);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Connection failed';
      setTestResult({ success: false, latency_ms: 0, message, model_found: undefined });
    } finally {
      setIsTesting(false);
    }
  }, [apiUrl, apiKey, modelName, testLLMConnection]);

  const handleSaveProvider = useCallback(async () => {
    if (!modelName.trim()) return;
    setIsSaving(true);
    try {
      const configId = `cfg_${selectedPreset.providerName}_${Date.now()}`;
      await saveLLMConfig({
        id: configId,
        provider_name: selectedPreset.providerName,
        api_url: apiUrl.trim(),
        token: apiKey.trim(),
        model_name: modelName.trim(),
        is_active: true,
        is_default: savedConfigs.length === 0,
        rate_limit_rpm: selectedPreset.rpm,
        timeout_seconds: selectedPreset.timeout,
        reasoning_effort: 'off',
        max_retries: 3,
        custom_headers: {},
      });
      await fetchLLMConfigs();
      setSavedConfigs((prev) => [...prev, selectedPreset.name]);
      toast.success(t('wizard.providerSaved'));
      setApiKey('');
      setTestResult(null);
    } catch (err: unknown) {
      const message = err instanceof Error ? err.message : 'Save failed';
      toast.error(message);
    } finally {
      setIsSaving(false);
    }
  }, [selectedPreset, apiUrl, apiKey, modelName, savedConfigs.length, saveLLMConfig, fetchLLMConfigs, t]);

  const handleNext = useCallback(() => {
    // Validate provider step
    if (step === 'provider' && savedConfigs.length === 0 && llmConfigs.length === 0) {
      toast.error(t('wizard.providerAtLeastOne'));
      return;
    }

    if (currentStep < STEPS.length - 1) {
      // When moving from provider to fallback, build the chain from the
      // providers ACTUALLY saved in the database (never invent tiers).
      if (step === 'provider') {
        const configs = useAppStore.getState().llmConfigs;
        if (configs.length > 0) {
          const chain: FallbackTier[] = configs.map((cfg, i) => ({
            id: `tier_${cfg.id}`,
            name: i === 0 ? t('settings.tierPrimary') : i === 1 ? t('settings.tierSecondary') : t('settings.tierTertiary'),
            configId: cfg.id,
            providerName: cfg.provider_name,
            modelName: cfg.model_name || '',
            maxRetries: cfg.max_retries ?? 3,
            timeoutSeconds: cfg.timeout_seconds ?? 120,
            reasoningEffort: (cfg.reasoning_effort as FallbackTier['reasoningEffort']) ?? 'off',
            triggerCondition: '429_5xx_timeout',
            enabled: true,
          }));
          setFallbackChain(chain);
        }
      }
      setCurrentStep((prev) => prev + 1);
    }
  }, [currentStep, step, savedConfigs.length, llmConfigs.length, setFallbackChain, t]);

  const handleBack = useCallback(() => {
    if (currentStep > 0) setCurrentStep((prev) => prev - 1);
  }, [currentStep]);

  const handleFinish = useCallback(() => {
    // Save preferences (store setters persist to localStorage/DB as needed).
    setTranslationMode(translationMode);
    setThinkingEffort(thinkingEffort);
    completeSetupWizard();
    toast.success(t('wizard.finish'));
  }, [translationMode, thinkingEffort, setTranslationMode, setThinkingEffort, completeSetupWizard, t]);

  const isLastStep = currentStep === STEPS.length - 1;
  const fallbackChain = useAppStore((s) => s.fallbackChain);

  return (
    <div className="fixed inset-0 z-9998 flex items-center justify-center bg-[#090d16]">
      {/* Ambient */}
      <div className="absolute inset-0 overflow-hidden pointer-events-none">
        <div className="absolute top-1/2 left-1/2 -translate-x-1/2 -translate-y-1/2 w-200 h-200 rounded-full bg-indigo-500/3 blur-[150px]" />
      </div>

      {/* Wizard Card */}
      <div className="relative w-full max-w-2xl mx-4 flex flex-col bg-[#0c101a] rounded-2xl border border-white/6 shadow-2xl overflow-hidden">
        {/* Header */}
        <div className="px-8 pt-8 pb-4">
          <div className="flex items-center gap-3 mb-1">
            <img src="/novelclaw-logo.png" alt="NovelClaw" className="w-8 h-8 rounded-lg shadow-md shadow-indigo-500/20" />
            <h1 className="text-xl font-semibold text-white tracking-tight">
              {t('wizard.title')}
            </h1>
          </div>
          <p className="text-sm text-white/40 ml-11">
            {t('wizard.subtitle')}
          </p>
        </div>

        {/* Step Indicator */}
        <div className="px-8 py-4">
          <div className="flex items-center gap-1">
            {STEPS.map((s, i) => (
              <React.Fragment key={s}>
                <button
                  onClick={() => {
                    if (i < currentStep) setCurrentStep(i);
                  }}
                  className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg text-xs font-medium transition-all duration-200 ${
                    i === currentStep
                      ? 'bg-indigo-500/15 text-indigo-400 border border-indigo-500/30'
                      : i < currentStep
                      ? 'bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 cursor-pointer hover:bg-emerald-500/15'
                      : 'bg-white/3 text-white/30 border border-white/6'
                  }`}
                >
                  {i < currentStep ? (
                    <Check className="w-3.5 h-3.5" />
                  ) : (
                    STEP_ICONS[s]
                  )}
                  <span className="hidden sm:inline">{t(STEP_KEYS[s])}</span>
                </button>
                {i < STEPS.length - 1 && (
                  <div className={`flex-1 h-px mx-1 ${i < currentStep ? 'bg-emerald-500/30' : 'bg-white/6'}`} />
                )}
              </React.Fragment>
            ))}
          </div>
          <p className="text-xs text-white/30 mt-2">
            {t('wizard.step', { current: currentStep + 1, total: STEPS.length })}
          </p>
        </div>

        {/* Step Content */}
        <div className="px-8 py-4 flex-1 min-h-85 max-h-120 overflow-y-auto custom-scrollbar">
          {step === 'language' && (
            <LanguageStep
              selectedLang={selectedLang}
              onSelect={handleLanguageChange}
            />
          )}
          {step === 'provider' && (
            <ProviderStep
              presets={PRESETS}
              selectedPreset={selectedPreset}
              onSelectPreset={handleSelectPreset}
              apiKey={apiKey}
              onApiKeyChange={setApiKey}
              apiUrl={apiUrl}
              onApiUrlChange={setApiUrl}
              modelName={modelName}
              onModelNameChange={setModelName}
              isTesting={isTesting}
              isSaving={isSaving}
              testResult={testResult}
              savedConfigs={savedConfigs}
              onTest={handleTestConnection}
              onSave={handleSaveProvider}
            />
          )}
          {step === 'fallback' && (
            <FallbackStep fallbackChain={fallbackChain} />
          )}
          {step === 'preferences' && (
            <PreferencesStep
              translationMode={translationMode}
              onModeChange={setLocalTranslationMode}
              thinkingEffort={thinkingEffort}
              onEffortChange={setLocalThinkingEffort}
            />
          )}
        </div>

        {/* Footer */}
        <div className="px-8 py-5 border-t border-white/6 flex items-center justify-between">
          <button
            onClick={handleBack}
            disabled={currentStep === 0}
            className={`flex items-center gap-1.5 px-4 py-2 rounded-lg text-sm font-medium transition-all duration-200 ${
              currentStep === 0
                ? 'text-white/20 cursor-not-allowed'
                : 'text-white/60 hover:text-white hover:bg-white/6'
            }`}
          >
            <ChevronLeft className="w-4 h-4" />
            {t('wizard.back')}
          </button>

          <div className="flex items-center gap-3">
            {step !== 'language' && !isLastStep && (
              <button
                onClick={() => setCurrentStep((prev) => prev + 1)}
                className="px-4 py-2 rounded-lg text-sm text-white/40 hover:text-white/60 transition-colors"
              >
                {t('wizard.skip')}
              </button>
            )}
            <button
              onClick={isLastStep ? handleFinish : handleNext}
              className="flex items-center gap-1.5 px-5 py-2.5 rounded-lg text-sm font-medium bg-indigo-600 hover:bg-indigo-500 text-white transition-all duration-200 shadow-lg shadow-indigo-500/20"
            >
              {isLastStep ? t('wizard.finish') : t('wizard.next')}
              {!isLastStep && <ChevronRight className="w-4 h-4" />}
              {isLastStep && <Zap className="w-4 h-4" />}
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

/* ──────────────────────── Step Components ──────────────────────── */

const LanguageStep: React.FC<{
  selectedLang: string;
  onSelect: (lang: string) => void;
}> = ({ selectedLang, onSelect }) => {
  const { t } = useTranslation();
  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-lg font-semibold text-white">{t('wizard.langTitle')}</h2>
        <p className="text-sm text-white/50 mt-1">{t('wizard.langDescription')}</p>
      </div>
      <div className="grid grid-cols-1 gap-2">
        {LANGUAGES.map((lang) => (
          <button
            key={lang.code}
            onClick={() => onSelect(lang.code)}
            className={`flex items-center gap-4 px-4 py-3.5 rounded-xl border transition-all duration-200 text-left ${
              selectedLang === lang.code
                ? 'bg-indigo-500/10 border-indigo-500/30 ring-1 ring-indigo-500/20'
                : 'bg-white/2 border-white/6 hover:bg-white/4 hover:border-white/10'
            }`}
          >
            <span className="text-2xl">{lang.flag}</span>
            <div className="flex-1">
              <span className={`text-sm font-medium ${selectedLang === lang.code ? 'text-indigo-300' : 'text-white/80'}`}>
                {lang.label}
              </span>
              <span className="text-xs text-white/30 ml-2 uppercase">{lang.code}</span>
            </div>
            {selectedLang === lang.code && (
              <Check className="w-5 h-5 text-indigo-400" />
            )}
          </button>
        ))}
      </div>
    </div>
  );
};

const ProviderStep: React.FC<{
  presets: ProviderPreset[];
  selectedPreset: ProviderPreset;
  onSelectPreset: (p: ProviderPreset) => void;
  apiKey: string;
  onApiKeyChange: (v: string) => void;
  apiUrl: string;
  onApiUrlChange: (v: string) => void;
  modelName: string;
  onModelNameChange: (v: string) => void;
  isTesting: boolean;
  isSaving: boolean;
  testResult: dtos.TestConnectionResponse | null;
  savedConfigs: string[];
  onTest: () => void;
  onSave: () => void;
}> = ({
  presets,
  selectedPreset,
  onSelectPreset,
  apiKey,
  onApiKeyChange,
  apiUrl,
  onApiUrlChange,
  modelName,
  onModelNameChange,
  isTesting,
  isSaving,
  testResult,
  savedConfigs,
  onTest,
  onSave,
}) => {
  const { t } = useTranslation();
  return (
    <div className="space-y-5">
      <div>
        <h2 className="text-lg font-semibold text-white">{t('wizard.providerTitle')}</h2>
        <p className="text-sm text-white/50 mt-1">{t('wizard.providerDescription')}</p>
      </div>

      {/* Saved indicator */}
      {savedConfigs.length > 0 && (
        <div className="flex items-center gap-2 px-3 py-2 rounded-lg bg-emerald-500/10 border border-emerald-500/20">
          <Check className="w-4 h-4 text-emerald-400" />
          <span className="text-xs text-emerald-300">
            {savedConfigs.join(', ')}
          </span>
        </div>
      )}

      {/* Preset selector */}
      <div>
        <label className="text-xs text-white/50 font-medium mb-2 block">{t('wizard.providerSelect')}</label>
        <div className="flex flex-wrap gap-2">
          {presets.map((p) => (
            <button
              key={p.id}
              onClick={() => onSelectPreset(p)}
              title={p.hintKey ? t(p.hintKey, p.hint) : p.hint}
              className={`px-3 py-1.5 rounded-lg text-xs font-medium border transition-all duration-200 ${
                selectedPreset.id === p.id
                  ? 'bg-indigo-500/15 border-indigo-500/30 text-indigo-300'
                  : 'bg-white/3 border-white/6 text-white/60 hover:bg-white/6 hover:text-white/80'
              }`}
            >
              {p.nameKey ? t(p.nameKey, p.name) : p.name}
            </button>
          ))}
        </div>
      </div>

      {/* Form */}
      <div className="space-y-3">
        <div>
          <label className="text-xs text-white/50 font-medium mb-1 block">{t('wizard.providerApiUrl')}</label>
          <input
            value={apiUrl}
            onChange={(e) => onApiUrlChange(e.target.value)}
            className="w-full px-3 py-2 rounded-lg bg-white/4 border border-white/8 text-sm text-white placeholder-white/25 focus:outline-none focus:ring-1 focus:ring-indigo-500/50 focus:border-indigo-500/30 transition-all"
          />
        </div>
        <div>
          <label className="text-xs text-white/50 font-medium mb-1 block">{t('wizard.providerApiKey')}</label>
          <input
            type="password"
            value={apiKey}
            onChange={(e) => onApiKeyChange(e.target.value)}
            placeholder={t('wizard.providerApiKeyPlaceholder')}
            className="w-full px-3 py-2 rounded-lg bg-white/4 border border-white/8 text-sm text-white placeholder-white/25 focus:outline-none focus:ring-1 focus:ring-indigo-500/50 focus:border-indigo-500/30 transition-all"
          />
        </div>
        <div>
          <label className="text-xs text-white/50 font-medium mb-1 block">{t('wizard.providerModel')}</label>
          <input
            value={modelName}
            onChange={(e) => onModelNameChange(e.target.value)}
            placeholder={selectedPreset.placeholderModelKey ? t(selectedPreset.placeholderModelKey, selectedPreset.placeholderModel) : selectedPreset.placeholderModel}
            className="w-full px-3 py-2 rounded-lg bg-white/4 border border-white/8 text-sm text-white placeholder-white/25 focus:outline-none focus:ring-1 focus:ring-indigo-500/50 focus:border-indigo-500/30 transition-all"
          />
        </div>
      </div>

      {/* Test result */}
      {testResult && (
        <div className={`flex items-center gap-2 px-3 py-2 rounded-lg text-xs ${
          testResult.success
            ? 'bg-emerald-500/10 border border-emerald-500/20 text-emerald-300'
            : 'bg-red-500/10 border border-red-500/20 text-red-300'
        }`}>
          {testResult.success ? <Check className="w-3.5 h-3.5" /> : <AlertTriangle className="w-3.5 h-3.5" />}
          {testResult.success ? t('wizard.providerTestSuccess') : (testResult.message || t('wizard.providerTestFail'))}
          {testResult.success && testResult.latency_ms > 0 && (
            <span className="text-white/30 ml-auto">{testResult.latency_ms}ms</span>
          )}
        </div>
      )}

      {/* Actions */}
      <div className="flex items-center gap-2">
        <button
          onClick={onTest}
          disabled={isTesting || !modelName.trim()}
          className="flex items-center gap-1.5 px-3 py-2 rounded-lg text-xs font-medium bg-white/6 text-white/70 hover:bg-white/10 hover:text-white disabled:opacity-40 disabled:cursor-not-allowed transition-all"
        >
          {isTesting ? (
            <div className="w-3.5 h-3.5 rounded-full border-2 border-white/20 border-t-white animate-spin" />
          ) : (
            <Zap className="w-3.5 h-3.5" />
          )}
          {t('wizard.providerTest')}
        </button>
        <button
          onClick={onSave}
          disabled={isSaving || !modelName.trim()}
          className="flex items-center gap-1.5 px-4 py-2 rounded-lg text-xs font-medium bg-indigo-600 hover:bg-indigo-500 text-white disabled:opacity-40 disabled:cursor-not-allowed transition-all shadow-lg shadow-indigo-500/20"
        >
          {isSaving ? (
            <div className="w-3.5 h-3.5 rounded-full border-2 border-white/20 border-t-white animate-spin" />
          ) : (
            <Check className="w-3.5 h-3.5" />
          )}
          {t('common.save')}
        </button>
      </div>
    </div>
  );
};

const FallbackStep: React.FC<{
  fallbackChain: FallbackTier[];
}> = ({ fallbackChain }) => {
  const { t } = useTranslation();
  return (
    <div className="space-y-4">
      <div>
        <h2 className="text-lg font-semibold text-white">{t('wizard.fallbackTitle')}</h2>
        <p className="text-sm text-white/50 mt-1">{t('wizard.fallbackDescription')}</p>
      </div>

      {fallbackChain.length > 0 ? (
        <>
          <div className="flex items-center gap-2 px-3 py-2 rounded-lg bg-indigo-500/10 border border-indigo-500/20">
            <Check className="w-4 h-4 text-indigo-400" />
            <span className="text-xs text-indigo-300">{t('wizard.fallbackAutoGenerated')}</span>
          </div>

          <div className="space-y-2">
            {fallbackChain.map((tier, i) => (
              <div
                key={tier.id}
                className="flex items-center gap-3 px-4 py-3 rounded-xl bg-white/3 border border-white/6"
              >
                <div className="flex items-center justify-center w-6 h-6 rounded-full bg-indigo-500/20 text-indigo-300 text-xs font-bold">
                  {i + 1}
                </div>
                <div className="flex-1 min-w-0">
                  <div className="text-sm text-white font-medium truncate">{tier.name}</div>
                  <div className="text-xs text-white/40 truncate">
                    {tier.providerName} · {tier.modelName}
                  </div>
                </div>
                <div className="flex items-center gap-1.5">
                  <span className="text-[10px] text-white/30">
                    {tier.maxRetries}× retry
                  </span>
                  <span className={`w-2 h-2 rounded-full ${tier.enabled ? 'bg-emerald-400' : 'bg-white/20'}`} />
                </div>
              </div>
            ))}
          </div>
        </>
      ) : (
        <div className="flex flex-col items-center justify-center py-12 text-white/30">
          <Shield className="w-8 h-8 mb-3 opacity-40" />
          <p className="text-sm">{t('wizard.providerAtLeastOne')}</p>
        </div>
      )}
    </div>
  );
};

const EFFORT_LEVELS = ['off', 'low', 'medium', 'high', 'xhigh', 'max'] as const;

const PreferencesStep: React.FC<{
  translationMode: string;
  onModeChange: (mode: string) => void;
  thinkingEffort: 'off' | 'low' | 'medium' | 'high' | 'xhigh' | 'max';
  onEffortChange: (effort: 'off' | 'low' | 'medium' | 'high' | 'xhigh' | 'max') => void;
}> = ({ translationMode, onModeChange, thinkingEffort, onEffortChange }) => {
  const { t } = useTranslation();
  return (
    <div className="space-y-6">
      <div>
        <h2 className="text-lg font-semibold text-white">{t('wizard.prefsTitle')}</h2>
        <p className="text-sm text-white/50 mt-1">{t('wizard.prefsDescription')}</p>
      </div>

      {/* Translation Mode — same 4 modes as Settings (shared constant) */}
      <div className="space-y-2">
        <label className="text-xs text-white/50 font-medium">{t('wizard.prefsMode')}</label>
        <div className="grid grid-cols-2 gap-2">
          {TRANSLATION_MODES.map((mode) => {
            const active = translationMode === mode.id;
            return (
              <button
                key={mode.id}
                onClick={() => onModeChange(mode.id)}
                className={`flex flex-col items-start gap-1 px-4 py-3 rounded-xl border transition-all duration-200 text-left ${
                  active
                    ? 'bg-indigo-500/10 border-indigo-500/30 ring-1 ring-indigo-500/20'
                    : 'bg-white/2 border-white/6 hover:bg-white/4'
                }`}
              >
                <span className={`text-sm font-medium ${active ? 'text-indigo-300' : 'text-white/70'}`}>
                  {mode.labelKey ? t(mode.labelKey) : mode.label}
                </span>
                <span className="text-[10px] text-white/40 uppercase tracking-wider font-mono">{mode.badge}</span>
              </button>
            );
          })}
        </div>
        <p className="text-xs text-white/30">{t('wizard.prefsModeDesc')}</p>
      </div>

      {/* Thinking Effort */}
      <div className="space-y-3">
        <label className="text-xs text-white/50 font-medium">{t('wizard.prefsThinking')}</label>
        <div className="flex items-center gap-1">
          {EFFORT_LEVELS.map((level) => (
            <button
              key={level}
              onClick={() => onEffortChange(level)}
              className={`flex-1 py-2 rounded-lg text-xs font-medium transition-all duration-200 ${
                thinkingEffort === level
                  ? 'bg-indigo-500/20 text-indigo-300 border border-indigo-500/30'
                  : 'bg-white/3 text-white/40 border border-white/6 hover:bg-white/6 hover:text-white/60'
              }`}
            >
              {level === 'xhigh' ? 'X-High' : level.charAt(0).toUpperCase() + level.slice(1)}
            </button>
          ))}
        </div>
        <p className="text-xs text-white/30">{t('wizard.prefsThinkingDesc')}</p>
      </div>
    </div>
  );
};
