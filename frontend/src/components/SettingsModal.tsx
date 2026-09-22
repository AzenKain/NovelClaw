import React, { useState, useEffect, useCallback } from "react";
import { useTranslation } from "react-i18next";
import {
  Settings,
  X,
  Key,
  Globe,
  Cpu,
  Shield,
  Zap,
  CheckCircle2,
  AlertTriangle,
  Trash2,
  RefreshCw,
  Sparkles,
  Flame,
  ArrowRight,
  Plus,
  Download,
  Upload,
  Edit3,
  UserCheck,
  Smile,
  RotateCcw,
  Boxes,
  Layers,
  BookOpen,
  Search,
  FileCheck2,
  ShieldAlert,
  Check,
  ChevronUp,
  ChevronDown,
  Crown,
  Ban,
  Lightbulb,
  Bookmark,
  Info,
  DownloadCloud,
} from "lucide-react";
import { useAppStore, FallbackTier } from "@/store/useAppStore";
import * as dtos from "@bindings/novelclaw/internal/dtos/models";
import { Select } from "@/components/ui/Select";
import { toast } from "react-hot-toast";
import { showConfirmModal } from "@/store/confirmStore";
import {
  PRESETS,
  TRANSLATION_MODES,
  type ProviderPreset,
} from "@/constants/providers";

const MODEL_SUGGESTIONS = [
  "claude-3-5-sonnet-20241022",
  "gpt-4o",
  "gpt-4o-mini",
  "deepseek-chat",
  "deepseek-reasoner",
  "gemini-2.0-flash",
  "qwen2.5:14b",
];

export const SettingsModal: React.FC = () => {
  const { t } = useTranslation();

  const {
    isSettingsOpen,
    setSettingsOpen,
    llmConfigs,
    defaultLLMConfig,
    fallbackChain,
    setFallbackChain,
    fetchLLMConfigs,
    saveLLMConfig,
    deleteLLMConfig,
    setDefaultLLMConfig,
    testLLMConnection,
    translationMode,
    setTranslationMode,
    enableR19,
    setEnableR19,
    enableHotPatch,
    setEnableHotPatch,
    enableAgenticRAG,
    setEnableAgenticRAG,
    maxToolIterations,
    setMaxToolIterations,
    styleGuide,
    setStyleGuide,
    decisionMode,
    setDecisionMode,
    autoCompactThreshold,
    setAutoCompactThreshold,
    soul,
    availableSouls,
    fetchAvailableSouls,
    setActiveSoul,
    saveCustomSoul,
    deleteCustomSoul,
    restoreDefaultSouls,
    importSoul,
    exportSoul,
    appInfo,
    checkForUpdates,
    selectedProject,
    skills,
    skillPresets,
    activeSkillsTokenWeight,
    isSkillsLoading,
    fetchSkills,
    fetchSkillPresets,
    toggleSkill,
    updateSkillCustomRules,
    resetSkillToDefault,
    resetSkillsToDefault,
    applyCapabilityPreset,
  } = useAppStore();

  const [activeTab, setActiveTab] = useState<
    "providers" | "fallback" | "pipeline" | "soul" | "skills" | "about"
  >("providers");

  const handleCheckForUpdates = async () => {
    if (isCheckingUpdate) return;
    setIsCheckingUpdate(true);
    setUpdateFeedback(null);
    try {
      await checkForUpdates();
      setUpdateFeedback("update.done"); // updater window takes over
    } catch {
      // The updater window itself surfaces the failure; we only flip the
      // inline hint so the button does not look stuck.
      setUpdateFeedback("update.error");
    } finally {
      setIsCheckingUpdate(false);
    }
  };

  // Fallback Chain Management State
  const [localFallbackChain, setLocalFallbackChain] = useState<FallbackTier[]>(
    [],
  );
  const [isFallbackDirty, setIsFallbackDirty] = useState(false);
  const [fallbackSuccess, setFallbackSuccess] = useState<string | null>(null);

  // Skills Management State
  const [editingSkill, setEditingSkill] = useState<dtos.SkillDTO | null>(null);
  const [customRulesDraft, setCustomRulesDraft] = useState<string>("");
  const [isSavingRules, setIsSavingRules] = useState(false);
  const [skillFeedback, setSkillFeedback] = useState<string | null>(null);

  // Soul Workshop State
  const [editingSoul, setEditingSoul] = useState<dtos.SoulDTO | null>(null);
  const [isCreatingSoul, setIsCreatingSoul] = useState(false);
  const [soulImportText, setSoulImportText] = useState("");
  const [showSoulImportModal, setShowSoulImportModal] = useState(false);
  const [soulFeedback, setSoulFeedback] = useState<string | null>(null);
  const [isSavingSoul, setIsSavingSoul] = useState(false);

  // Form State
  const [configId, setConfigId] = useState<string>("");
  const [providerName, setProviderName] = useState("openrouter");
  const [apiUrl, setApiUrl] = useState("https://openrouter.ai/api/v1");
  const [token, setToken] = useState("");
  const [modelName, setModelName] = useState("");
  const [modelPlaceholder, setModelPlaceholder] = useState(
    "vd: deepseek/deepseek-chat, gpt-4o, gemini-2.0-flash...",
  );
  const [rateLimitRpm, setRateLimitRpm] = useState(60);
  const [timeoutSeconds, setTimeoutSeconds] = useState(120);
  const [reasoningEffort, setReasoningEffort] = useState<
    "off" | "low" | "medium" | "high" | "xhigh" | "max"
  >("off");
  const [isDefault, setIsDefault] = useState(true);

  // Update / About state
  const [isCheckingUpdate, setIsCheckingUpdate] = useState(false);
  const [updateFeedback, setUpdateFeedback] = useState<string | null>(null);

  // Status & Feedback
  const [isSaving, setIsSaving] = useState(false);
  const [isTesting, setIsTesting] = useState(false);
  const [testResult, setTestResult] =
    useState<dtos.TestConnectionResponse | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  useEffect(() => {
    if (isSettingsOpen) {
      fetchLLMConfigs();
      const projId = selectedProject?.id || "default";
      fetchSkills(projId);
      fetchSkillPresets();
    }
  }, [
    isSettingsOpen,
    fetchLLMConfigs,
    fetchSkills,
    fetchSkillPresets,
    selectedProject,
  ]);

  const generateDefaultChain = useCallback((): FallbackTier[] => {
    if (llmConfigs && llmConfigs.length > 0) {
      const defCfg = defaultLLMConfig || llmConfigs[0];
      const otherConfigs = llmConfigs.filter((c) => c.id !== defCfg.id);

      const chain: FallbackTier[] = [
        {
          id: `tier_primary_${Date.now()}`,
          name: t("settings.tierPrimary"),
          configId: defCfg.id,
          providerName: defCfg.provider_name,
          modelName: defCfg.model_name || "deepseek-chat",
          maxRetries: defCfg.max_retries || 3,
          timeoutSeconds: defCfg.timeout_seconds || 120,
          triggerCondition: "429_5xx_timeout",
          enabled: true,
        },
      ];

      if (otherConfigs.length > 0) {
        const secCfg = otherConfigs[0];
        chain.push({
          id: `tier_secondary_${Date.now() + 1}`,
          name: t("settings.tierSecondary"),
          configId: secCfg.id,
          providerName: secCfg.provider_name,
          modelName: secCfg.model_name || "gpt-4o-mini",
          maxRetries: secCfg.max_retries || 3,
          timeoutSeconds: secCfg.timeout_seconds || 90,
          triggerCondition: "429_5xx_timeout",
          enabled: true,
        });
      } else {
        chain.push({
          id: `tier_secondary_${Date.now() + 1}`,
          name: t("settings.tierSecondary"),
          configId: "",
          providerName: "openrouter",
          modelName: "deepseek/deepseek-chat",
          maxRetries: 3,
          timeoutSeconds: 90,
          triggerCondition: "429_5xx_timeout",
          enabled: true,
        });
      }

      if (otherConfigs.length > 1) {
        const tertCfg = otherConfigs[1];
        chain.push({
          id: `tier_tertiary_${Date.now() + 2}`,
          name: t("settings.tierTertiary"),
          configId: tertCfg.id,
          providerName: tertCfg.provider_name,
          modelName: tertCfg.model_name || "qwen2.5:14b",
          maxRetries: tertCfg.max_retries || 2,
          timeoutSeconds: tertCfg.timeout_seconds || 60,
          triggerCondition: "429_5xx_timeout",
          enabled: true,
        });
      } else {
        chain.push({
          id: `tier_tertiary_${Date.now() + 2}`,
          name: t("settings.tierTertiary"),
          configId: "",
          providerName: "ollama",
          modelName: "qwen2.5:14b",
          maxRetries: 2,
          timeoutSeconds: 180,
          triggerCondition: "429_5xx_timeout",
          enabled: true,
        });
      }

      return chain;
    }

    return [
      {
        id: `tier_1_${Date.now()}`,
        name: t("settings.tierPrimary"),
        configId: "",
        providerName: "openrouter",
        modelName: "deepseek/deepseek-chat",
        maxRetries: 3,
        timeoutSeconds: 120,
        triggerCondition: "429_5xx_timeout",
        enabled: true,
      },
      {
        id: `tier_2_${Date.now() + 1}`,
        name: t("settings.tierSecondary"),
        configId: "",
        providerName: "openai",
        modelName: "gpt-4o-mini",
        maxRetries: 3,
        timeoutSeconds: 90,
        triggerCondition: "429_5xx_timeout",
        enabled: true,
      },
      {
        id: `tier_3_${Date.now() + 2}`,
        name: t("settings.tierTertiary"),
        configId: "",
        providerName: "ollama",
        modelName: "qwen2.5:14b",
        maxRetries: 2,
        timeoutSeconds: 180,
        triggerCondition: "429_5xx_timeout",
        enabled: true,
      },
    ];
  }, [llmConfigs, defaultLLMConfig, t]);

  useEffect(() => {
    if (isSettingsOpen) {
      if (fallbackChain && fallbackChain.length > 0) {
        setLocalFallbackChain(fallbackChain);
      } else {
        const defChain = generateDefaultChain();
        setLocalFallbackChain(defChain);
        setFallbackChain(defChain);
      }
    }
  }, [isSettingsOpen, fallbackChain, generateDefaultChain, setFallbackChain]);

  const handleMoveTierUp = (index: number) => {
    if (index <= 0) return;
    setLocalFallbackChain((prev) => {
      const next = [...prev];
      const temp = next[index - 1];
      next[index - 1] = next[index];
      next[index] = temp;
      return next;
    });
    setIsFallbackDirty(true);
  };

  const handleMoveTierDown = (index: number) => {
    if (index >= localFallbackChain.length - 1) return;
    setLocalFallbackChain((prev) => {
      const next = [...prev];
      const temp = next[index + 1];
      next[index + 1] = next[index];
      next[index] = temp;
      return next;
    });
    setIsFallbackDirty(true);
  };

  const handleAddTier = () => {
    const newTierNumber = localFallbackChain.length + 1;
    const defaultProvider = llmConfigs[0]?.provider_name || "openrouter";
    const defaultCfgId = llmConfigs[0]?.id || "";
    const defaultModel = llmConfigs[0]?.model_name || "deepseek-chat";
    const newTier: FallbackTier = {
      id: `tier_${Date.now()}`,
      name:
        newTierNumber === 1
          ? t("settings.tierPrimary")
          : newTierNumber === 2
            ? t("settings.tierSecondary")
            : newTierNumber === 3
              ? t("settings.tierTertiary")
              : `${t("settings.tierBackup")} ${newTierNumber}`,
      configId: defaultCfgId,
      providerName: defaultProvider,
      modelName: defaultModel,
      maxRetries: 3,
      timeoutSeconds: 120,
      triggerCondition: "429_5xx_timeout",
      enabled: true,
    };
    setLocalFallbackChain((prev) => [...prev, newTier]);
    setIsFallbackDirty(true);
  };

  const handleDeleteTier = (index: number) => {
    if (localFallbackChain.length <= 1) {
      toast.error(t("settings.keepAtLeastOneTier"));
      return;
    }
    setLocalFallbackChain((prev) => prev.filter((_, i) => i !== index));
    setIsFallbackDirty(true);
  };

  const handleToggleTier = (index: number) => {
    setLocalFallbackChain((prev) =>
      prev.map((tier, i) =>
        i === index ? { ...tier, enabled: !tier.enabled } : tier,
      ),
    );
    setIsFallbackDirty(true);
  };

  const handleUpdateTier = (index: number, updates: Partial<FallbackTier>) => {
    setLocalFallbackChain((prev) =>
      prev.map((tier, i) => (i === index ? { ...tier, ...updates } : tier)),
    );
    setIsFallbackDirty(true);
  };

  const handleSelectProviderForTier = (index: number, provVal: string) => {
    const matchedCfg = llmConfigs.find(
      (c) => c.id === provVal || c.provider_name === provVal,
    );
    if (matchedCfg) {
      handleUpdateTier(index, {
        configId: matchedCfg.id,
        providerName: matchedCfg.provider_name,
        modelName: matchedCfg.model_name || localFallbackChain[index].modelName,
        timeoutSeconds:
          matchedCfg.timeout_seconds ||
          localFallbackChain[index].timeoutSeconds,
        maxRetries:
          matchedCfg.max_retries || localFallbackChain[index].maxRetries,
      });
    } else {
      const preset = PRESETS.find(
        (p) => p.providerName === provVal || p.id === provVal,
      );
      const suggestedModel = preset?.placeholderModel
        ? preset.placeholderModel.split(",")[0].replace("vd:", "").trim()
        : localFallbackChain[index].modelName;
      handleUpdateTier(index, {
        configId: "",
        providerName: provVal,
        modelName: suggestedModel,
      });
    }
  };

  const handleApplyFallbackPreset = (
    presetKey: "high_perf" | "cost_saving" | "offline_safe",
  ) => {
    let newTiers: FallbackTier[] = [];
    if (presetKey === "high_perf") {
      newTiers = [
        {
          id: `tier_1_${Date.now()}`,
          name: t("settings.tierPrimary"),
          configId:
            llmConfigs.find((c) => c.provider_name === "anthropic")?.id ||
            llmConfigs[0]?.id ||
            "",
          providerName: "anthropic",
          modelName: "claude-3-5-sonnet-20241022",
          maxRetries: 3,
          timeoutSeconds: 120,
          triggerCondition: "429_5xx_timeout",
          enabled: true,
        },
        {
          id: `tier_2_${Date.now() + 1}`,
          name: t("settings.tierSecondary"),
          configId:
            llmConfigs.find((c) => c.provider_name === "openai")?.id ||
            llmConfigs[0]?.id ||
            "",
          providerName: "openai",
          modelName: "gpt-4o",
          maxRetries: 3,
          timeoutSeconds: 90,
          triggerCondition: "429_5xx_timeout",
          enabled: true,
        },
        {
          id: `tier_3_${Date.now() + 2}`,
          name: t("settings.tierTertiary"),
          configId:
            llmConfigs.find((c) => c.provider_name === "deepseek")?.id ||
            llmConfigs[0]?.id ||
            "",
          providerName: "deepseek",
          modelName: "deepseek-chat",
          maxRetries: 2,
          timeoutSeconds: 60,
          triggerCondition: "429_5xx_timeout",
          enabled: true,
        },
      ];
    } else if (presetKey === "cost_saving") {
      newTiers = [
        {
          id: `tier_1_${Date.now()}`,
          name: t("settings.tierPrimary"),
          configId:
            llmConfigs.find((c) => c.provider_name === "deepseek")?.id ||
            llmConfigs[0]?.id ||
            "",
          providerName: "deepseek",
          modelName: "deepseek-chat",
          maxRetries: 3,
          timeoutSeconds: 90,
          triggerCondition: "429_5xx_timeout",
          enabled: true,
        },
        {
          id: `tier_2_${Date.now() + 1}`,
          name: t("settings.tierSecondary"),
          configId:
            llmConfigs.find((c) => c.provider_name === "openrouter")?.id ||
            llmConfigs[0]?.id ||
            "",
          providerName: "openrouter",
          modelName: "qwen/qwen-2.5-72b-instruct",
          maxRetries: 2,
          timeoutSeconds: 90,
          triggerCondition: "429_5xx_timeout",
          enabled: true,
        },
        {
          id: `tier_3_${Date.now() + 2}`,
          name: t("settings.tierTertiary"),
          configId:
            llmConfigs.find((c) => c.provider_name === "openai")?.id ||
            llmConfigs[0]?.id ||
            "",
          providerName: "openai",
          modelName: "gpt-4o-mini",
          maxRetries: 2,
          timeoutSeconds: 60,
          triggerCondition: "429_5xx_timeout",
          enabled: true,
        },
      ];
    } else if (presetKey === "offline_safe") {
      newTiers = [
        {
          id: `tier_1_${Date.now()}`,
          name: t("settings.tierPrimary"),
          configId: defaultLLMConfig?.id || llmConfigs[0]?.id || "",
          providerName: defaultLLMConfig?.provider_name || "openrouter",
          modelName: defaultLLMConfig?.model_name || "deepseek/deepseek-chat",
          maxRetries: 3,
          timeoutSeconds: 60,
          triggerCondition: "any_error",
          enabled: true,
        },
        {
          id: `tier_2_${Date.now() + 1}`,
          name: t("settings.tierSecondary"),
          configId:
            llmConfigs.find((c) => c.provider_name === "ollama")?.id || "",
          providerName: "ollama",
          modelName: "qwen2.5:14b",
          maxRetries: 2,
          timeoutSeconds: 180,
          triggerCondition: "429_5xx_timeout",
          enabled: true,
        },
      ];
    }
    setLocalFallbackChain(newTiers);
    setIsFallbackDirty(true);
    toast.success(t("settings.appliedFallbackPreset"));
  };

  const handleResetFallbackChain = () => {
    const defChain = generateDefaultChain();
    setLocalFallbackChain(defChain);
    setIsFallbackDirty(true);
    toast.success(t("settings.resetDefaultChain"));
  };

  const handleSaveFallbackChain = () => {
    setFallbackChain(localFallbackChain);
    setIsFallbackDirty(false);
    setFallbackSuccess(t("settings.fallbackSaved"));
    toast.success(t("settings.fallbackSaved"));
    setTimeout(() => setFallbackSuccess(null), 3500);
  };

  const handleApplySkillPreset = async (presetId: string) => {
    try {
      const projId = selectedProject?.id || "default";
      await applyCapabilityPreset(presetId, projId);
      setSkillFeedback(
        t("settings.skillPresetApplied", { preset: presetId.toUpperCase() }),
      );
      toast.success(
        t("settings.skillPresetApplied", { preset: presetId.toUpperCase() }),
      );
      setTimeout(() => setSkillFeedback(null), 3500);
    } catch (err: any) {
      toast.error(
        t("settings.skillPresetError", { error: err.message || err }),
      );
    }
  };

  const handleToggleSkill = async (
    skillId: string,
    currentEnabled: boolean,
  ) => {
    try {
      const projId = selectedProject?.id || "default";
      await toggleSkill(skillId, !currentEnabled, projId);
    } catch (err: any) {
      toast.error(
        t("settings.skillToggleError", { error: err.message || err }),
      );
    }
  };

  const handleResetSkill = async (skillId: string) => {
    try {
      const projId = selectedProject?.id || "default";
      await resetSkillToDefault(skillId, projId);
      if (editingSkill && editingSkill.id === skillId) {
        setCustomRulesDraft("");
      }
      setSkillFeedback(t("settings.skillResetSuccess"));
      toast.success(t("settings.skillResetSuccess"));
      setTimeout(() => setSkillFeedback(null), 3000);
    } catch (err: any) {
      toast.error(t("settings.skillResetError", { error: err.message || err }));
    }
  };

  const handleSaveCustomRules = async () => {
    if (!editingSkill) return;
    setIsSavingRules(true);
    try {
      const projId = selectedProject?.id || "default";
      await updateSkillCustomRules(editingSkill.id, customRulesDraft, projId);
      const skillDisplayName = t(
        `settings.skillsList.${editingSkill.id}.name`,
        { defaultValue: editingSkill.name },
      );
      setSkillFeedback(
        t("settings.skillUpdatedRules", { name: skillDisplayName }),
      );
      toast.success(
        t("settings.skillUpdatedRules", { name: skillDisplayName }),
      );
      setTimeout(() => setSkillFeedback(null), 3500);
      setEditingSkill(null);
    } catch (err: any) {
      toast.error(
        t("settings.skillUpdateError", { error: err.message || err }),
      );
    } finally {
      setIsSavingRules(false);
    }
  };

  if (!isSettingsOpen) return null;

  const handleApplyPreset = (p: ProviderPreset) => {
    setProviderName(p.providerName);
    setApiUrl(p.apiUrl);
    setModelPlaceholder(
      p.placeholderModelKey
        ? t(p.placeholderModelKey, p.placeholderModel)
        : p.placeholderModel,
    );
    setRateLimitRpm(p.rpm);
    setTimeoutSeconds(p.timeout);
    setTestResult(null);
    setErrorMessage(null);
  };

  const handleEditConfig = (cfg: dtos.LLMConfigDTO) => {
    setConfigId(cfg.id);
    setProviderName(cfg.provider_name);
    setApiUrl(cfg.api_url);
    setToken("");
    setModelName(cfg.model_name);
    setRateLimitRpm(cfg.rate_limit_rpm || 60);
    setTimeoutSeconds(cfg.timeout_seconds || 120);
    setReasoningEffort((cfg.reasoning_effort as any) || "off");
    setIsDefault(cfg.is_default);
    setTestResult(null);
    setErrorMessage(null);
    setSuccessMessage(null);
  };

  const handleResetForm = () => {
    setConfigId("");
    setProviderName("openrouter");
    setApiUrl("https://openrouter.ai/api/v1");
    setToken("");
    setModelName("");
    setModelPlaceholder(
      "vd: deepseek/deepseek-chat, gpt-4o, gemini-2.0-flash...",
    );
    setRateLimitRpm(60);
    setTimeoutSeconds(120);
    setReasoningEffort("off");
    setIsDefault(llmConfigs.length === 0);
    setTestResult(null);
    setErrorMessage(null);
    setSuccessMessage(null);
  };

  const handleTestConnection = async () => {
    setIsTesting(true);
    setTestResult(null);
    setErrorMessage(null);
    try {
      const res = await testLLMConnection({
        api_url: apiUrl.trim(),
        token: token.trim(),
        model_name: modelName.trim(),
        config_id: configId || undefined,
      });
      if (res) {
        setTestResult(res);
      }
    } catch (err: any) {
      setTestResult({
        success: false,
        latency_ms: 0,
        message: err?.message || t("settings.connectionError"),
        model_found: undefined,
      });
    } finally {
      setIsTesting(false);
    }
  };

  const handleSaveSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    setIsSaving(true);
    setErrorMessage(null);
    setSuccessMessage(null);

    try {
      const targetId = configId || `cfg_${providerName}_${Date.now()}`;
      await saveLLMConfig({
        id: targetId,
        provider_name: providerName.trim(),
        api_url: apiUrl.trim(),
        token: token.trim(),
        model_name: modelName.trim(),
        is_active: true,
        is_default: isDefault,
        rate_limit_rpm: rateLimitRpm,
        timeout_seconds: timeoutSeconds,
        reasoning_effort: reasoningEffort,
        max_retries: 3,
        custom_headers: {},
      });

      setSuccessMessage(t("settings.saveSuccess"));
      handleResetForm();
    } catch (err: any) {
      setErrorMessage(err?.message || t("settings.saveError"));
    } finally {
      setIsSaving(false);
    }
  };

  const handleDelete = async (id: string) => {
    const confirmed = await showConfirmModal({
      title: t("settings.deleteLLMConfig"),
      message: t("settings.deleteLLMConfigConfirm"),
      confirmText: t("common.delete"),
      cancelText: t("common.cancel"),
      variant: "danger",
    });
    if (!confirmed) return;
    try {
      await deleteLLMConfig(id);
      if (configId === id) handleResetForm();
    } catch (err: any) {
      setErrorMessage(err?.message || t("settings.deleteError"));
    }
  };

  const handleSetDefault = async (id: string) => {
    try {
      await setDefaultLLMConfig(id);
    } catch (err: any) {
      setErrorMessage(err?.message || t("settings.setDefaultError"));
    }
  };

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
      <div className="bg-[#0c101a] border border-white/10 rounded-2xl max-w-4xl w-full h-[90vh] max-h-[90vh] shadow-2xl flex flex-col overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        {/* Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-white/8 bg-[#0e1422] shrink-0">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-indigo-500/10 text-indigo-400 rounded-lg border border-indigo-500/20">
              <Settings className="w-5 h-5" />
            </div>
            <div>
              <h2 className="text-sm font-bold text-slate-100 flex items-center space-x-2">
                <span>{t("settings.title")}</span>
                <span className="text-[10px] uppercase font-mono px-2 py-0.5 rounded bg-white/4 text-indigo-300 border border-indigo-500/20">
                  AES-256 Vault
                </span>
              </h2>
              <p className="text-xs text-slate-400">{t("settings.subtitle")}</p>
            </div>
          </div>
          <button
            onClick={() => setSettingsOpen(false)}
            className="text-slate-400 hover:text-slate-200 p-1.5 rounded-lg hover:bg-white/6 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Tab Navigation — six tabs in a balanced 3×2 grid so the row that
            holds About is as full as the first one (a 5-column grid left a
            single orphaned cell). */}
        <div className="border-b border-white/8 bg-[#0a0d16] px-2 sm:px-4 select-none shrink-0">
          <div className="grid grid-cols-3 gap-1">
            <button
              onClick={() => setActiveTab("providers")}
              className={`flex items-center justify-center space-x-2 py-3 px-2 text-xs font-semibold border-b-2 transition truncate ${
                activeTab === "providers"
                  ? "border-indigo-500 text-indigo-300 bg-indigo-500/8"
                  : "border-transparent text-slate-400 hover:text-slate-200 hover:bg-white/2"
              }`}
              title={t("settings.tabProviders")}
            >
              <Cpu className="w-4 h-4 shrink-0" />
              <span className="truncate">{t("settings.tabProviders")}</span>
            </button>
            <button
              onClick={() => setActiveTab("fallback")}
              className={`flex items-center justify-center space-x-2 py-3 px-2 text-xs font-semibold border-b-2 transition truncate ${
                activeTab === "fallback"
                  ? "border-indigo-500 text-indigo-300 bg-indigo-500/8"
                  : "border-transparent text-slate-400 hover:text-slate-200 hover:bg-white/2"
              }`}
              title={t("settings.tabFallback")}
            >
              <Zap className="w-4 h-4 shrink-0" />
              <span className="truncate">{t("settings.tabFallback")}</span>
            </button>
            <button
              onClick={() => setActiveTab("pipeline")}
              className={`flex items-center justify-center space-x-2 py-3 px-2 text-xs font-semibold border-b-2 transition truncate ${
                activeTab === "pipeline"
                  ? "border-indigo-500 text-indigo-300 bg-indigo-500/8"
                  : "border-transparent text-slate-400 hover:text-slate-200 hover:bg-white/2"
              }`}
              title={t("settings.tabPreferences")}
            >
              <Sparkles className="w-4 h-4 shrink-0" />
              <span className="truncate">{t("settings.tabPreferences")}</span>
            </button>
            <button
              onClick={() => {
                setActiveTab("soul");
                fetchAvailableSouls();
              }}
              className={`flex items-center justify-center space-x-2 py-3 px-2 text-xs font-semibold border-b-2 transition truncate ${
                activeTab === "soul"
                  ? "border-indigo-500 text-indigo-300 bg-indigo-500/8"
                  : "border-transparent text-slate-400 hover:text-slate-200 hover:bg-white/2"
              }`}
              title={t("settings.tabSouls")}
            >
              <Smile className="w-4 h-4 shrink-0" />
              <span className="truncate">{t("settings.tabSouls")}</span>
            </button>
            <button
              onClick={() => {
                setActiveTab("skills");
                const projId = selectedProject?.id || "default";
                fetchSkills(projId);
                fetchSkillPresets();
              }}
              className={`flex items-center justify-center space-x-2 py-3 px-2 text-xs font-semibold border-b-2 transition truncate ${
                activeTab === "skills"
                  ? "border-indigo-500 text-indigo-300 bg-indigo-500/8"
                  : "border-transparent text-slate-400 hover:text-slate-200 hover:bg-white/2"
              }`}
              title={t("settings.tabSkills")}
            >
              <Boxes className="w-4 h-4 shrink-0" />
              <span className="truncate">{t("settings.tabSkills")}</span>
            </button>
            <button
              onClick={() => setActiveTab("about")}
              className={`flex items-center justify-center space-x-2 py-3 px-2 text-xs font-semibold border-b-2 transition truncate ${
                activeTab === "about"
                  ? "border-indigo-500 text-indigo-300 bg-indigo-500/8"
                  : "border-transparent text-slate-400 hover:text-slate-200 hover:bg-white/2"
              }`}
              title={t("settings.tabAbout")}
            >
              <Info className="w-4 h-4 shrink-0" />
              <span className="truncate">{t("settings.tabAbout")}</span>
            </button>
          </div>
        </div>

        {/* Modal Body */}
        <div className="flex-1 min-h-0 overflow-y-auto p-4 sm:p-6 space-y-6">
          {activeTab === "providers" && (
            <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
              {/* Left Column: Form & Presets (7 cols) */}
              <div className="lg:col-span-7 space-y-4">
                {/* Presets */}
                <div>
                  <label className="block text-[11px] font-semibold text-slate-400 uppercase tracking-wider mb-2">
                    {t("settings.quickPresets")}
                  </label>
                  <div className="flex flex-wrap gap-1.5">
                    {PRESETS.map((p) => (
                      <button
                        key={p.id}
                        type="button"
                        onClick={() => handleApplyPreset(p)}
                        title={p.hintKey ? t(p.hintKey, p.hint) : p.hint}
                        className={`text-[11px] px-2.5 py-1 rounded-lg border font-medium transition ${
                          providerName === p.providerName
                            ? "bg-indigo-600 text-white border-indigo-400/40 shadow-xs"
                            : "bg-white/3 text-slate-300 border-white/8 hover:bg-white/[0.07]"
                        }`}
                      >
                        {p.nameKey ? t(p.nameKey, p.name) : p.name}
                      </button>
                    ))}
                  </div>
                </div>

                {/* Form */}
                <form
                  onSubmit={handleSaveSubmit}
                  className="bg-[#101522] border border-white/8 rounded-xl p-4 space-y-3.5"
                >
                  <div className="flex items-center justify-between pb-2 border-b border-white/6">
                    <span className="text-xs font-semibold text-slate-200 flex items-center space-x-1.5">
                      <Key className="w-3.5 h-3.5 text-indigo-400" />
                      <span>
                        {configId
                          ? t("settings.editProvider")
                          : t("settings.addProvider")}
                      </span>
                    </span>
                    {configId && (
                      <button
                        type="button"
                        onClick={handleResetForm}
                        className="text-[11px] text-slate-400 hover:text-slate-200 underline"
                      >
                        {t("settings.newConfig")}
                      </button>
                    )}
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    <div>
                      <label className="block text-slate-400 text-[11px] font-medium mb-1">
                        {t("settings.providerName")}:
                      </label>
                      <input
                        type="text"
                        placeholder="openrouter, deepseek, openai, vilao, custom..."
                        className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-2.5 py-1.5 text-xs text-slate-100 font-mono focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                        value={providerName}
                        onChange={(e) => setProviderName(e.target.value)}
                        required
                      />
                    </div>

                    <div>
                      <label className="block text-slate-400 text-[11px] font-medium mb-1 items-center justify-between">
                        <span>{t("settings.modelName")}:</span>
                        <span className="text-[10px] text-amber-400/90 font-sans ml-1">
                          * {t("settings.selfSpecified")}
                        </span>
                      </label>
                      <input
                        type="text"
                        placeholder={modelPlaceholder}
                        className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-2.5 py-1.5 text-xs text-slate-100 font-mono focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                        value={modelName}
                        onChange={(e) => setModelName(e.target.value)}
                        required
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block text-slate-400 text-[11px] font-medium mb-1 items-center justify-between">
                      <span>{t("settings.apiUrl")}:</span>
                      <span className="text-[10px] text-slate-500 ml-1">
                        OpenAI / Gemini / ViLao
                      </span>
                    </label>
                    <input
                      type="url"
                      placeholder="https://openrouter.ai/api/v1"
                      className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-2.5 py-1.5 text-xs text-slate-100 font-mono focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                      value={apiUrl}
                      onChange={(e) => setApiUrl(e.target.value)}
                      required
                    />
                  </div>

                  <div>
                    <label className="text-slate-400 text-[11px] font-medium mb-1 flex items-center justify-between">
                      <span className="flex items-center space-x-1">
                        <Shield className="w-3 h-3 text-emerald-400" />
                        <span>{t("settings.apiKey")}:</span>
                      </span>
                      {configId && !token && (
                        <span className="text-[10px] text-emerald-400 font-sans">
                          ({t("settings.existingKeyNotice")})
                        </span>
                      )}
                    </label>
                    <input
                      type="password"
                      placeholder={
                        configId
                          ? "••••••••••••••••••••••••"
                          : "sk-or-v1-... or sk-..."
                      }
                      className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-2.5 py-1.5 text-xs text-slate-100 font-mono focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                      value={token}
                      onChange={(e) => setToken(e.target.value)}
                    />
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    <div>
                      <label className="block text-slate-400 text-[11px] font-medium mb-1">
                        {t("settings.rateLimit")}:
                      </label>
                      <input
                        type="number"
                        min="1"
                        max="1000"
                        className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-2.5 py-1.5 text-xs text-slate-100 font-mono focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                        value={rateLimitRpm}
                        onChange={(e) =>
                          setRateLimitRpm(parseInt(e.target.value) || 60)
                        }
                      />
                    </div>

                    <div>
                      <label className="block text-slate-400 text-[11px] font-medium mb-1">
                        {t("settings.timeout")}:
                      </label>
                      <input
                        type="number"
                        min="10"
                        max="600"
                        className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-2.5 py-1.5 text-xs text-slate-100 font-mono focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                        value={timeoutSeconds}
                        onChange={(e) =>
                          setTimeoutSeconds(parseInt(e.target.value) || 120)
                        }
                      />
                    </div>
                  </div>

                  <div>
                    <div className="flex items-center justify-between mb-1">
                      <label className="text-slate-400 text-[11px] font-medium">
                        {t("settings.thinkingEffort")}:
                      </label>
                      <span className="text-[10px] text-indigo-400 font-normal font-mono">
                        {reasoningEffort === "off"
                          ? t("settings.effortOffDesc")
                          : t("settings.effortActiveDesc")}
                      </span>
                    </div>
                    <Select
                      value={reasoningEffort}
                      onChange={(e) =>
                        setReasoningEffort(e.target.value as any)
                      }
                      className="w-full text-xs font-sans"
                    >
                      <option value="off">{t("settings.effortOff")}</option>
                      <option value="low">{t("settings.effortLow")}</option>
                      <option value="medium">
                        {t("settings.effortMedium")}
                      </option>
                      <option value="high">{t("settings.effortHigh")}</option>
                      <option value="xhigh">{t("settings.effortXHigh")}</option>
                      <option value="max">{t("settings.effortMax")}</option>
                    </Select>
                  </div>

                  <div className="flex items-center space-x-2 pt-1">
                    <input
                      type="checkbox"
                      id="isDefaultCheckbox"
                      checked={isDefault}
                      onChange={(e) => setIsDefault(e.target.checked)}
                      className="w-3.5 h-3.5 rounded border-white/20 text-indigo-600 focus:ring-indigo-500"
                    />
                    <label
                      htmlFor="isDefaultCheckbox"
                      className="text-xs text-slate-300 font-medium cursor-pointer"
                    >
                      {t("settings.isDefault")}
                    </label>
                  </div>

                  {/* Feedback Messages */}
                  {testResult && (
                    <div
                      className={`p-2.5 rounded-lg border text-xs flex items-start space-x-2 ${
                        testResult.success
                          ? "bg-emerald-950/40 border-emerald-500/30 text-emerald-300"
                          : "bg-rose-950/40 border-rose-500/30 text-rose-300"
                      }`}
                    >
                      {testResult.success ? (
                        <CheckCircle2 className="w-4 h-4 text-emerald-400 shrink-0 mt-0.5" />
                      ) : (
                        <AlertTriangle className="w-4 h-4 text-rose-400 shrink-0 mt-0.5" />
                      )}
                      <div className="flex-1">
                        <div className="font-semibold flex items-center justify-between">
                          <span>{testResult.message}</span>
                          {testResult.latency_ms > 0 && (
                            <span className="font-mono text-[10px] px-1.5 py-0.2 bg-black/40 rounded">
                              {testResult.latency_ms} ms
                            </span>
                          )}
                        </div>
                        {testResult.model_found && (
                          <div className="text-[10px] opacity-80 mt-0.5 font-mono">
                            {t("settings.verifiedModel")}:{" "}
                            {testResult.model_found}
                          </div>
                        )}
                      </div>
                    </div>
                  )}

                  {errorMessage && (
                    <div className="p-2.5 bg-rose-950/40 border border-rose-500/30 rounded-lg text-rose-300 text-xs">
                      {errorMessage}
                    </div>
                  )}

                  {successMessage && (
                    <div className="p-2.5 bg-emerald-950/40 border border-emerald-500/30 rounded-lg text-emerald-300 text-xs">
                      {successMessage}
                    </div>
                  )}

                  {/* Action Buttons */}
                  <div className="flex items-center justify-between pt-2 border-t border-white/6">
                    <button
                      type="button"
                      disabled={isTesting}
                      onClick={handleTestConnection}
                      className="px-3.5 py-1.5 text-xs font-semibold bg-white/4 hover:bg-white/8 text-cyan-300 border border-cyan-500/30 rounded-lg flex items-center space-x-1.5 transition disabled:opacity-50"
                    >
                      {isTesting ? (
                        <>
                          <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                          <span>{t("settings.testing")}</span>
                        </>
                      ) : (
                        <>
                          <Zap className="w-3.5 h-3.5 text-cyan-400" />
                          <span>{t("settings.testConnection")}</span>
                        </>
                      )}
                    </button>

                    <button
                      type="submit"
                      disabled={isSaving}
                      className="px-4 py-1.5 text-xs font-semibold bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg shadow-sm border border-indigo-400/30 transition disabled:opacity-50"
                    >
                      {isSaving ? t("common.saving") : t("settings.saveConfig")}
                    </button>
                  </div>
                </form>
              </div>

              {/* Right Column: Configured Providers List (5 cols) */}
              <div className="lg:col-span-5 space-y-3">
                <div className="flex items-center justify-between">
                  <h4 className="text-xs font-semibold text-slate-300 uppercase tracking-wider">
                    {t("settings.configuredProviders")} ({llmConfigs.length})
                  </h4>
                  <button
                    onClick={fetchLLMConfigs}
                    className="text-slate-400 hover:text-slate-200 p-1 rounded-lg hover:bg-white/6 transition"
                    title="Refresh"
                  >
                    <RefreshCw className="w-3.5 h-3.5" />
                  </button>
                </div>

                {llmConfigs.length === 0 ? (
                  <div className="p-6 bg-white/2 border border-dashed border-white/8 rounded-xl text-center text-xs text-slate-400">
                    <Globe className="w-8 h-8 text-slate-600 mx-auto mb-2" />
                    <p>{t("settings.noConfigs")}</p>
                  </div>
                ) : (
                  <div className="space-y-2.5 max-h-120 overflow-y-auto pr-1">
                    {llmConfigs.map((cfg) => {
                      const isSelectedDefault =
                        cfg.is_default || defaultLLMConfig?.id === cfg.id;
                      return (
                        <div
                          key={cfg.id}
                          className={`p-3.5 rounded-xl border transition relative ${
                            isSelectedDefault
                              ? "bg-indigo-500/8 border-indigo-500/40 shadow-sm"
                              : "bg-white/2 border-white/8 hover:border-white/20"
                          }`}
                        >
                          <div className="flex items-center justify-between mb-1.5">
                            <div className="flex items-center space-x-2">
                              <span className="font-semibold text-xs text-slate-100 uppercase font-mono">
                                {cfg.provider_name}
                              </span>
                              {isSelectedDefault && (
                                <span className="px-1.5 py-0.5 text-[9px] font-semibold bg-amber-500/15 text-amber-300 border border-amber-500/30 rounded-md flex items-center space-x-0.5">
                                  <Flame className="w-2.5 h-2.5" />
                                  <span>{t("settings.defaultBadge")}</span>
                                </span>
                              )}
                            </div>
                            <div className="flex items-center space-x-1">
                              <button
                                onClick={() => handleEditConfig(cfg)}
                                className="px-2 py-0.5 text-[10px] font-medium bg-white/4 hover:bg-white/8 text-slate-300 rounded border border-white/10 transition"
                              >
                                {t("common.edit")}
                              </button>
                              <button
                                onClick={() => handleDelete(cfg.id)}
                                className="p-1 text-slate-400 hover:text-rose-400 rounded hover:bg-rose-950/40 transition"
                                title={t("settings.deleteConfig")}
                              >
                                <Trash2 className="w-3.5 h-3.5" />
                              </button>
                            </div>
                          </div>

                          <div className="text-[11px] font-mono text-indigo-300 truncate mb-1">
                            {cfg.model_name}
                          </div>

                          <div className="text-[10px] text-slate-400 truncate mb-2">
                            {cfg.api_url}
                          </div>

                          <div className="flex items-center justify-between pt-2 border-t border-white/6 text-[10px] text-slate-400 gap-1 flex-wrap">
                            <span className="font-mono">
                              RPM: {cfg.rate_limit_rpm || 60}
                            </span>
                            <span className="font-mono">
                              Timeout: {cfg.timeout_seconds || 120}s
                            </span>
                            <span
                              className={`font-mono px-1.5 py-0.2 rounded text-[9.5px] ${cfg.reasoning_effort && cfg.reasoning_effort !== "off" ? "bg-purple-500/15 text-purple-300 border border-purple-500/30" : "bg-white/4 text-slate-500"}`}
                            >
                              Think: {cfg.reasoning_effort || "off"}
                            </span>
                            {!isSelectedDefault && (
                              <button
                                onClick={() => handleSetDefault(cfg.id)}
                                className="text-indigo-400 hover:text-indigo-300 font-semibold underline ml-auto"
                              >
                                {t("settings.setDefault")}
                              </button>
                            )}
                          </div>
                        </div>
                      );
                    })}
                  </div>
                )}
              </div>
            </div>
          )}

          {activeTab === "fallback" && (
            <div className="space-y-6">
              {/* Header Banner & Strategy Presets */}
              <div className="bg-[#101522] border border-white/8 rounded-xl p-5 space-y-4">
                <div className="flex flex-col md:flex-row md:items-center justify-between gap-3">
                  <div>
                    <h3 className="text-sm font-semibold text-slate-200 mb-1 flex items-center space-x-2">
                      <Zap className="w-4 h-4 text-amber-400" />
                      <span>{t("settings.fallbackTitle")}</span>
                    </h3>
                    <p className="text-xs text-slate-400 leading-relaxed max-w-3xl">
                      {t("settings.fallbackSubtitle")}
                    </p>
                  </div>

                  {/* Save Button in Header */}
                  <div className="shrink-0 flex items-center gap-2">
                    {fallbackSuccess && (
                      <span className="text-xs text-emerald-400 font-medium flex items-center gap-1">
                        <Check className="w-3.5 h-3.5" />
                        {fallbackSuccess}
                      </span>
                    )}
                    <button
                      type="button"
                      onClick={handleSaveFallbackChain}
                      className={`px-3.5 py-1.5 rounded-lg text-xs font-semibold flex items-center gap-1.5 transition-all shadow-sm ${
                        isFallbackDirty
                          ? "bg-indigo-600 hover:bg-indigo-500 text-white shadow-indigo-500/25 ring-1 ring-indigo-400"
                          : "bg-white/8 hover:bg-white/12 text-slate-200 border border-white/10"
                      }`}
                    >
                      <Check className="w-3.5 h-3.5" />
                      <span>{t("settings.saveFallbackChain")}</span>
                    </button>
                  </div>
                </div>

                {/* Quick Strategy Presets */}
                <div className="pt-3 border-t border-white/6 flex flex-wrap items-center gap-2">
                  <span className="text-[11px] font-medium text-slate-400 mr-1">
                    {t("settings.fallbackPresets")}
                  </span>
                  <button
                    type="button"
                    onClick={() => handleApplyFallbackPreset("high_perf")}
                    className="px-2.5 py-1 rounded-lg bg-indigo-500/10 hover:bg-indigo-500/20 text-indigo-300 border border-indigo-500/20 text-[11px] font-medium transition-all"
                  >
                    {t("settings.presetHighPerf")}
                  </button>
                  <button
                    type="button"
                    onClick={() => handleApplyFallbackPreset("cost_saving")}
                    className="px-2.5 py-1 rounded-lg bg-emerald-500/10 hover:bg-emerald-500/20 text-emerald-300 border border-emerald-500/20 text-[11px] font-medium transition-all"
                  >
                    {t("settings.presetCostSaving")}
                  </button>
                  <button
                    type="button"
                    onClick={() => handleApplyFallbackPreset("offline_safe")}
                    className="px-2.5 py-1 rounded-lg bg-cyan-500/10 hover:bg-cyan-500/20 text-cyan-300 border border-cyan-500/20 text-[11px] font-medium transition-all"
                  >
                    {t("settings.presetOfflineSafety")}
                  </button>
                </div>

                {/* Live Dynamic Flow Diagram */}
                <div className="p-4 bg-[#090d16] rounded-xl border border-white/6 space-y-2">
                  <div className="text-[10px] font-semibold text-slate-400 uppercase tracking-wider flex items-center gap-1.5">
                    <Zap className="w-3 h-3 text-amber-400" />
                    <span>{t("settings.fallbackFlowPreview")}</span>
                  </div>

                  <div className="flex flex-wrap items-center gap-2.5 text-xs pt-1">
                    {localFallbackChain.map((tier, idx) => {
                      const isLast = idx === localFallbackChain.length - 1;
                      const tierBadgeStyle =
                        idx === 0
                          ? "border-indigo-500/40 bg-indigo-500/10 text-indigo-300"
                          : idx === 1
                            ? "border-cyan-500/40 bg-cyan-500/10 text-cyan-300"
                            : idx === 2
                              ? "border-amber-500/40 bg-amber-500/10 text-amber-300"
                              : "border-emerald-500/40 bg-emerald-500/10 text-emerald-300";

                      return (
                        <React.Fragment key={tier.id}>
                          <div
                            className={`flex-1 min-w-40 p-3 rounded-lg border transition-all ${
                              tier.enabled
                                ? "bg-[#111726] border-white/10 shadow-xs"
                                : "bg-[#0a0d16] border-white/5 opacity-50"
                            }`}
                          >
                            <div className="flex items-center justify-between gap-1.5 mb-1.5">
                              <span
                                className={`text-[10px] px-1.5 py-0.5 rounded font-mono font-bold border ${tierBadgeStyle}`}
                              >
                                {tier.name ||
                                  `${t("settings.tierLabel")} ${idx + 1}`}
                              </span>
                              {!tier.enabled && (
                                <span className="text-[9px] text-slate-500 font-mono bg-white/4 px-1.5 py-0.5 rounded">
                                  {t("settings.tierDisabled")}
                                </span>
                              )}
                            </div>
                            <div className="font-semibold text-slate-200 capitalize text-xs truncate">
                              {tier.providerName}
                            </div>
                            <div
                              className="text-[10px] text-slate-400 font-mono truncate mt-0.5"
                              title={tier.modelName}
                            >
                              {tier.modelName || t("settings.noModelSelected")}
                            </div>
                            <div className="flex items-center gap-2 text-[10px] text-slate-500 mt-2 pt-1.5 border-t border-white/4 font-mono">
                              <span>{tier.maxRetries}x retry</span>
                              <span>•</span>
                              <span>{tier.timeoutSeconds}s</span>
                            </div>
                          </div>

                          {!isLast && (
                            <div className="shrink-0 flex items-center justify-center px-1 text-slate-500">
                              <div className="flex flex-col items-center gap-0.5">
                                <ArrowRight className="w-3.5 h-3.5 text-slate-400" />
                                <span className="text-[9px] text-amber-400 font-mono px-1.5 py-0.5 bg-amber-500/10 rounded border border-amber-500/20 whitespace-nowrap">
                                  {tier.triggerCondition === "429_only"
                                    ? "429"
                                    : tier.triggerCondition === "any_error"
                                      ? "All Err"
                                      : "429 / 5xx"}
                                </span>
                              </div>
                            </div>
                          )}
                        </React.Fragment>
                      );
                    })}
                  </div>
                </div>
              </div>

              {/* Tiers Configuration List */}
              <div className="bg-[#101522] border border-white/8 rounded-xl p-5 space-y-4">
                <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-2">
                  <div>
                    <h3 className="text-sm font-semibold text-slate-200 flex items-center space-x-2">
                      <Layers className="w-4 h-4 text-indigo-400" />
                      <span>{t("settings.fallbackTiersConfig")}</span>
                    </h3>
                    <p className="text-xs text-slate-400 mt-0.5">
                      {t("settings.fallbackDesc")}
                    </p>
                  </div>

                  <div className="flex items-center gap-2 shrink-0">
                    <button
                      type="button"
                      onClick={handleResetFallbackChain}
                      className="px-2.5 py-1.5 rounded-lg bg-white/4 hover:bg-white/8 text-slate-400 hover:text-slate-200 border border-white/10 text-xs flex items-center gap-1.5 transition-all"
                      title={t("settings.resetDefaultChain")}
                    >
                      <RotateCcw className="w-3.5 h-3.5" />
                      <span className="hidden sm:inline">
                        {t("settings.resetDefaultChain")}
                      </span>
                    </button>
                    <button
                      type="button"
                      onClick={handleAddTier}
                      className="px-3 py-1.5 rounded-lg bg-indigo-600/20 hover:bg-indigo-600/30 text-indigo-300 border border-indigo-500/30 text-xs font-medium flex items-center gap-1.5 transition-all"
                    >
                      <Plus className="w-3.5 h-3.5" />
                      <span>{t("settings.addTier")}</span>
                    </button>
                  </div>
                </div>

                {/* Tier Cards */}
                <div className="space-y-3 pt-1">
                  {localFallbackChain.map((tier, index) => {
                    const isFirst = index === 0;
                    const isLast = index === localFallbackChain.length - 1;

                    return (
                      <div
                        key={tier.id}
                        className={`p-4 rounded-xl border transition-all ${
                          tier.enabled
                            ? "bg-[#0a0e1a] border-white/8 hover:border-white/15"
                            : "bg-[#090d16] border-white/4 opacity-60"
                        }`}
                      >
                        {/* Tier Header: Order controls, Name, Enable/Disable, Delete */}
                        <div className="flex items-center justify-between gap-3 mb-3 pb-2.5 border-b border-white/6">
                          <div className="flex items-center gap-2">
                            {/* Move Up / Down Buttons */}
                            <div className="flex items-center space-x-1">
                              <button
                                type="button"
                                disabled={isFirst}
                                onClick={() => handleMoveTierUp(index)}
                                className="p-1 rounded bg-white/4 hover:bg-white/8 text-slate-400 hover:text-slate-200 disabled:opacity-20 disabled:cursor-not-allowed border border-white/6 transition-all"
                                title={t("settings.moveUp")}
                              >
                                <ChevronUp className="w-3.5 h-3.5" />
                              </button>
                              <button
                                type="button"
                                disabled={isLast}
                                onClick={() => handleMoveTierDown(index)}
                                className="p-1 rounded bg-white/4 hover:bg-white/8 text-slate-400 hover:text-slate-200 disabled:opacity-20 disabled:cursor-not-allowed border border-white/6 transition-all"
                                title={t("settings.moveDown")}
                              >
                                <ChevronDown className="w-3.5 h-3.5" />
                              </button>
                            </div>

                            {/* Tier Badge */}
                            <span className="text-[11px] font-mono font-bold px-2 py-0.5 rounded bg-indigo-500/10 text-indigo-300 border border-indigo-500/20">
                              {t("settings.tierLabel")} {index + 1}
                            </span>

                            {/* Tier Name input */}
                            <input
                              type="text"
                              value={tier.name}
                              onChange={(e) =>
                                handleUpdateTier(index, {
                                  name: e.target.value,
                                })
                              }
                              placeholder={`${t("settings.tierLevel")} ${index + 1}`}
                              className="h-7 px-2 text-xs bg-transparent border border-transparent hover:border-white/10 focus:border-indigo-500 focus:bg-[#111726] rounded text-slate-200 font-medium focus:outline-none w-36 sm:w-48 transition-all"
                            />
                          </div>

                          <div className="flex items-center gap-2">
                            {/* Toggle Enabled */}
                            <button
                              type="button"
                              onClick={() => handleToggleTier(index)}
                              className={`px-2 py-0.5 rounded text-[11px] font-medium border flex items-center gap-1.5 transition-all ${
                                tier.enabled
                                  ? "bg-emerald-500/10 text-emerald-300 border-emerald-500/30"
                                  : "bg-white/4 text-slate-500 border-white/10 hover:text-slate-400"
                              }`}
                              title={
                                tier.enabled
                                  ? t("settings.disableTier")
                                  : t("settings.enableTier")
                              }
                            >
                              <div
                                className={`w-1.5 h-1.5 rounded-full ${
                                  tier.enabled
                                    ? "bg-emerald-400"
                                    : "bg-slate-500"
                                }`}
                              />
                              <span>
                                {tier.enabled
                                  ? t("settings.activeBadge")
                                  : t("settings.tierDisabled")}
                              </span>
                            </button>

                            {/* Delete Tier */}
                            <button
                              type="button"
                              disabled={localFallbackChain.length <= 1}
                              onClick={() => handleDeleteTier(index)}
                              className="p-1 rounded bg-rose-500/10 hover:bg-rose-500/20 text-rose-400 disabled:opacity-20 disabled:cursor-not-allowed border border-rose-500/20 transition-all"
                              title={t("settings.removeTier")}
                            >
                              <Trash2 className="w-3.5 h-3.5" />
                            </button>
                          </div>
                        </div>

                        {/* Tier Inputs Grid */}
                        <div className="space-y-3 text-xs">
                          {/* Row 1: Provider Selection & Model Identifier */}
                          <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                            {/* Provider Select */}
                            <div>
                              <label className="block text-slate-400 text-[11px] font-medium mb-1">
                                {t("settings.providerName")}
                              </label>
                              <Select
                                value={tier.configId || tier.providerName}
                                onChange={(e) =>
                                  handleSelectProviderForTier(
                                    index,
                                    e.target.value,
                                  )
                                }
                                className="w-full h-8 px-2.5 text-xs font-sans"
                                containerClassName="w-full"
                              >
                                {llmConfigs.length > 0 && (
                                  <optgroup
                                    label={t("settings.savedConfigsInStudio")}
                                  >
                                    {llmConfigs.map((cfg) => (
                                      <option key={cfg.id} value={cfg.id}>
                                        ★ {cfg.provider_name.toUpperCase()} (
                                        {cfg.model_name || "No model"})
                                        {cfg.is_default
                                          ? ` [${t("settings.defaultBadge")}]`
                                          : ""}
                                      </option>
                                    ))}
                                  </optgroup>
                                )}
                                <optgroup label={t("settings.sampleProviders")}>
                                  {PRESETS.map((p) => (
                                    <option key={p.id} value={p.providerName}>
                                      {p.nameKey
                                        ? t(p.nameKey, p.name)
                                        : p.name}
                                    </option>
                                  ))}
                                </optgroup>
                              </Select>
                            </div>

                            {/* Model Name Customizer */}
                            <div>
                              <label className="block text-slate-400 text-[11px] font-medium mb-1">
                                {t("settings.customModel")}
                              </label>
                              <input
                                type="text"
                                value={tier.modelName}
                                onChange={(e) =>
                                  handleUpdateTier(index, {
                                    modelName: e.target.value,
                                  })
                                }
                                placeholder={t("settings.modelPlaceholder")}
                                className="w-full h-8 px-2.5 text-xs bg-[#090d16] border border-white/10 rounded-lg text-slate-200 font-mono focus:border-indigo-500 focus:outline-none"
                              />

                              {/* Quick Model Suggestion Chips */}
                              <div className="flex flex-wrap items-center gap-1.5 mt-1.5">
                                <span className="text-[10px] text-slate-500 font-mono">
                                  {t("settings.suggestions")}
                                </span>
                                {MODEL_SUGGESTIONS.map((sug) => (
                                  <button
                                    key={sug}
                                    type="button"
                                    onClick={() =>
                                      handleUpdateTier(index, {
                                        modelName: sug,
                                      })
                                    }
                                    className={`text-[10px] px-1.5 py-0.5 rounded font-mono border transition-all ${
                                      tier.modelName === sug
                                        ? "bg-indigo-600/30 text-indigo-300 border-indigo-500/50 font-semibold"
                                        : "bg-white/3 text-slate-400 border-white/6 hover:bg-white/8 hover:text-slate-200"
                                    }`}
                                  >
                                    {sug}
                                  </button>
                                ))}
                              </div>
                            </div>
                          </div>

                          {/* Row 2: Trigger Condition, Thinking Effort, Retries, Timeout */}
                          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-3 pt-1 border-t border-white/4">
                            {/* Trigger Condition */}
                            <div>
                              <label className="block text-slate-400 text-[11px] font-medium mb-1">
                                {t("settings.triggerCondition")}
                              </label>
                              <Select
                                value={tier.triggerCondition}
                                onChange={(e) =>
                                  handleUpdateTier(index, {
                                    triggerCondition: e.target.value as
                                      | "429_5xx_timeout"
                                      | "429_only"
                                      | "any_error",
                                  })
                                }
                                className="w-full h-8 px-2.5 text-xs font-sans"
                                containerClassName="w-full"
                              >
                                <option value="429_5xx_timeout">
                                  {t("settings.triggerAll")}
                                </option>
                                <option value="429_only">
                                  {t("settings.trigger429Only")}
                                </option>
                                <option value="any_error">
                                  {t("settings.triggerAnyError")}
                                </option>
                              </Select>
                            </div>

                            {/* Thinking Effort */}
                            <div>
                              <label className="block text-slate-400 text-[11px] font-medium mb-1">
                                {t("settings.thinkingEffort")}
                              </label>
                              <Select
                                value={tier.reasoningEffort || "off"}
                                onChange={(e) =>
                                  handleUpdateTier(index, {
                                    reasoningEffort: e.target.value as any,
                                  })
                                }
                                className="w-full h-8 px-2.5 text-xs font-sans"
                                containerClassName="w-full"
                              >
                                <option value="off">
                                  {t("settings.effortOff")}
                                </option>
                                <option value="low">
                                  {t("settings.effortLow")}
                                </option>
                                <option value="medium">
                                  {t("settings.effortMedium")}
                                </option>
                                <option value="high">
                                  {t("settings.effortHigh")}
                                </option>
                                <option value="xhigh">
                                  {t("settings.effortXHigh")}
                                </option>
                                <option value="max">
                                  {t("settings.effortMax")}
                                </option>
                              </Select>
                            </div>

                            {/* Max Retries */}
                            <div>
                              <label className="block text-slate-400 text-[11px] font-medium mb-1">
                                {t("settings.retries")}
                              </label>
                              <input
                                type="number"
                                min={1}
                                max={5}
                                value={tier.maxRetries}
                                onChange={(e) =>
                                  handleUpdateTier(index, {
                                    maxRetries: Math.max(
                                      1,
                                      parseInt(e.target.value) || 1,
                                    ),
                                  })
                                }
                                className="w-full h-8 px-2.5 text-xs bg-[#090d16] border border-white/10 rounded-lg text-slate-200 font-mono focus:border-indigo-500 focus:outline-none"
                              />
                            </div>

                            {/* Timeout Seconds */}
                            <div>
                              <label className="block text-slate-400 text-[11px] font-medium mb-1">
                                {t("settings.timeout")}
                              </label>
                              <input
                                type="number"
                                min={15}
                                max={300}
                                step={15}
                                value={tier.timeoutSeconds}
                                onChange={(e) =>
                                  handleUpdateTier(index, {
                                    timeoutSeconds: Math.max(
                                      15,
                                      parseInt(e.target.value) || 60,
                                    ),
                                  })
                                }
                                className="w-full h-8 px-2.5 text-xs bg-[#090d16] border border-white/10 rounded-lg text-slate-200 font-mono focus:border-indigo-500 focus:outline-none"
                              />
                            </div>
                          </div>
                        </div>
                      </div>
                    );
                  })}
                </div>

                {/* Bottom Save & Add Action Row */}
                <div className="pt-3 border-t border-white/6 flex flex-col sm:flex-row items-center justify-between gap-3">
                  <div className="text-xs text-slate-400 flex items-center gap-2">
                    {isFallbackDirty ? (
                      <span className="text-amber-400 flex items-center gap-1 font-mono text-[11px]">
                        <AlertTriangle className="w-3.5 h-3.5" />
                        {t("settings.unsavedFallbackChanges")}
                      </span>
                    ) : (
                      <span className="text-slate-500 font-mono text-[11px]">
                        {t("settings.fallbackSynced")}
                      </span>
                    )}
                  </div>

                  <div className="flex items-center gap-2 w-full sm:w-auto">
                    <button
                      type="button"
                      onClick={handleAddTier}
                      className="flex-1 sm:flex-none px-3 py-2 rounded-lg bg-white/4 hover:bg-white/8 text-slate-300 border border-white/10 text-xs font-medium flex items-center justify-center gap-1.5 transition-all"
                    >
                      <Plus className="w-3.5 h-3.5" />
                      <span>{t("settings.addTier")}</span>
                    </button>

                    <button
                      type="button"
                      onClick={handleSaveFallbackChain}
                      className="flex-1 sm:flex-none px-4 py-2 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold flex items-center justify-center gap-1.5 transition-all shadow-sm shadow-indigo-500/25"
                    >
                      <Check className="w-3.5 h-3.5" />
                      <span>{t("settings.saveFallbackChain")}</span>
                    </button>
                  </div>
                </div>
              </div>

              {/* Resilience & Deduplication Info Cards */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4 text-xs">
                <div className="p-4 bg-[#101522] border border-white/8 rounded-xl">
                  <h4 className="font-semibold text-slate-200 mb-1 flex items-center space-x-1.5">
                    <Shield className="w-3.5 h-3.5 text-emerald-400" />
                    <span>{t("settings.backoffTitle")}</span>
                  </h4>
                  <p className="text-slate-400 leading-relaxed text-[11px]">
                    {t("settings.backoffDesc")}
                  </p>
                </div>

                <div className="p-4 bg-[#101522] border border-white/8 rounded-xl">
                  <h4 className="font-semibold text-slate-200 mb-1 flex items-center space-x-1.5">
                    <Sparkles className="w-3.5 h-3.5 text-indigo-400" />
                    <span>{t("settings.dedupTitle")}</span>
                  </h4>
                  <p className="text-slate-400 leading-relaxed text-[11px]">
                    {t("settings.dedupDesc")}
                  </p>
                </div>
              </div>
            </div>
          )}

          {activeTab === "pipeline" && (
            <div className="space-y-5 text-xs">
              <div className="bg-[#101522] border border-white/8 rounded-xl p-5 space-y-4">
                <h3 className="text-sm font-semibold text-slate-200 flex items-center space-x-2">
                  <Sparkles className="w-4 h-4 text-indigo-400" />
                  <span>{t("settings.pipelineTitle")}</span>
                </h3>

                <div>
                  <label className="block text-slate-400 text-[11px] font-medium mb-1">
                    {t("settings.defaultModeLabel")}
                  </label>
                  <Select
                    value={translationMode}
                    onChange={(e) => setTranslationMode(e.target.value)}
                    className="w-full h-9 px-3 text-xs"
                    containerClassName="w-full"
                  >
                    {TRANSLATION_MODES.map((mode) => (
                      <option key={mode.id} value={mode.id}>
                        {mode.labelKey ? t(mode.labelKey) : mode.label}
                      </option>
                    ))}
                  </Select>
                </div>

                <div className="space-y-2.5 pt-2">
                  <label className="flex items-start space-x-2.5 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={enableHotPatch}
                      onChange={(e) => setEnableHotPatch(e.target.checked)}
                      className="w-4 h-4 rounded border-white/20 text-indigo-600 focus:ring-indigo-500 mt-0.5"
                    />
                    <div>
                      <span className="font-semibold text-slate-200">
                        {t("settings.shadowCriticTitle")}
                      </span>
                      <p className="text-[11px] text-slate-400">
                        {t("settings.shadowCriticDesc")}
                      </p>
                    </div>
                  </label>

                  <label className="flex items-start space-x-2.5 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={enableR19}
                      onChange={(e) => setEnableR19(e.target.checked)}
                      className="w-4 h-4 rounded border-white/20 text-indigo-600 focus:ring-indigo-500 mt-0.5"
                    />
                    <div>
                      <span className="font-semibold text-slate-200">
                        {t("settings.r19ShieldTitle")}
                      </span>
                      <p className="text-[11px] text-slate-400">
                        {t("settings.r19ShieldDesc")}
                      </p>
                    </div>
                  </label>

                  <label className="flex items-start space-x-2.5 cursor-pointer">
                    <input
                      type="checkbox"
                      checked={enableAgenticRAG}
                      onChange={(e) => setEnableAgenticRAG(e.target.checked)}
                      className="w-4 h-4 rounded border-white/20 text-indigo-600 focus:ring-indigo-500 mt-0.5"
                    />
                    <div>
                      <span className="font-semibold text-slate-200">
                        {t("settings.agenticRAGTitle")}
                      </span>
                      <p className="text-[11px] text-slate-400">
                        {t("settings.agenticRAGDesc")}
                      </p>
                    </div>
                  </label>

                  {enableAgenticRAG && (
                    <div className="pl-6 pt-1 flex items-center gap-3">
                      <span className="text-xs text-slate-300 font-medium">
                        {t("settings.maxIterationsLabel")}
                      </span>
                      <input
                        type="number"
                        min={1}
                        max={4}
                        value={maxToolIterations}
                        onChange={(e) =>
                          setMaxToolIterations(
                            Math.max(
                              1,
                              Math.min(4, parseInt(e.target.value) || 2),
                            ),
                          )
                        }
                        className="w-14 h-7 px-2 bg-[#111726] border border-white/10 rounded text-xs text-white text-center focus:outline-none focus:border-indigo-500"
                      />
                      <span className="text-[11px] text-slate-500">
                        {t("settings.recommendedQueries")}
                      </span>
                    </div>
                  )}
                </div>

                <div className="pt-2">
                  <label className="block text-slate-400 text-[11px] font-medium mb-1">
                    {t("settings.globalStyleGuideLabel")}
                  </label>
                  <textarea
                    rows={3}
                    placeholder={t("settings.globalStyleGuidePlaceholder")}
                    value={styleGuide}
                    onChange={(e) => setStyleGuide(e.target.value)}
                    className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg p-3 text-xs text-slate-200 focus:outline-none focus:ring-1 focus:ring-indigo-500/30 font-sans resize-none transition"
                  />
                </div>

                {/* Decision Mode (HITL Stage 7) */}
                <div className="pt-3 border-t border-white/8 space-y-3">
                  <h4 className="text-xs font-semibold text-amber-300 flex items-center space-x-1.5">
                    <span>⚖️</span>
                    <span>{t("settings.decisionModeTitle")}</span>
                  </h4>
                  <p className="text-[11px] text-slate-400">
                    {t("settings.decisionModeDesc")}
                  </p>
                  <div className="space-y-2">
                    <label
                      className={`flex items-start space-x-2.5 cursor-pointer p-2.5 rounded-lg border transition ${
                        decisionMode === "manual"
                          ? "bg-amber-500/10 border-amber-500/30"
                          : "bg-transparent border-white/6 hover:border-white/12"
                      }`}
                      onClick={() => setDecisionMode("manual")}
                    >
                      <input
                        type="radio"
                        name="decision_mode"
                        value="manual"
                        checked={decisionMode === "manual"}
                        onChange={() => setDecisionMode("manual")}
                        className="w-4 h-4 text-amber-500 focus:ring-amber-500 mt-0.5"
                      />
                      <div>
                        <span className="font-semibold text-slate-200">
                          {t("settings.alwaysReviewTitle")}
                        </span>
                        <p className="text-[11px] text-slate-400 mt-0.5">
                          {t("settings.alwaysReviewDesc")}
                        </p>
                      </div>
                    </label>
                    <label
                      className={`flex items-start space-x-2.5 cursor-pointer p-2.5 rounded-lg border transition ${
                        decisionMode === "auto"
                          ? "bg-emerald-500/10 border-emerald-500/30"
                          : "bg-transparent border-white/6 hover:border-white/12"
                      }`}
                      onClick={() => setDecisionMode("auto")}
                    >
                      <input
                        type="radio"
                        name="decision_mode"
                        value="auto"
                        checked={decisionMode === "auto"}
                        onChange={() => setDecisionMode("auto")}
                        className="w-4 h-4 text-emerald-500 focus:ring-emerald-500 mt-0.5"
                      />
                      <div>
                        <span className="font-semibold text-slate-200">
                          {t("settings.autoProcessTitle")}
                        </span>
                        <p className="text-[11px] text-slate-400 mt-0.5">
                          {t("settings.autoProcessDesc")}
                        </p>
                      </div>
                    </label>
                  </div>
                </div>

                {/* NovelClaw Agentic Chat Settings */}
                <div className="pt-3 border-t border-white/8 space-y-3">
                  <h4 className="text-xs font-semibold text-indigo-300 flex items-center space-x-1.5">
                    <span>🐾</span>
                    <span>{t("settings.novelclawConsoleTitle")}</span>
                  </h4>

                  <div>
                    <div className="flex items-center justify-between mb-1">
                      <label className="text-slate-400 text-[11px] font-medium">
                        {t("settings.autoCompactThresholdLabel")}
                      </label>
                      <span className="font-mono text-indigo-400 text-xs font-semibold">
                        {(autoCompactThreshold || 250000).toLocaleString()}{" "}
                        tokens
                      </span>
                    </div>
                    <input
                      type="range"
                      min={50000}
                      max={500000}
                      step={25000}
                      value={autoCompactThreshold || 250000}
                      onChange={(e) =>
                        setAutoCompactThreshold(Number(e.target.value))
                      }
                      className="w-full accent-indigo-500 cursor-pointer"
                    />
                    <div className="flex justify-between text-[10px] text-slate-500 font-mono mt-0.5">
                      <span>{t("settings.compactEarly")}</span>
                      <span>{t("settings.compactDefault")}</span>
                      <span>{t("settings.compactLarge")}</span>
                    </div>
                  </div>
                </div>
              </div>
            </div>
          )}

          {activeTab === "soul" && (
            <div className="space-y-6">
              {/* Top Action Bar */}
              <div className="flex flex-col sm:flex-row sm:items-center justify-between gap-3 p-4 rounded-xl bg-[#0f1422] border border-white/8">
                <div>
                  <h3 className="text-sm font-semibold text-slate-100 flex items-center space-x-2">
                    <span className="text-lg">🎭</span>
                    <span>{t("settings.soulsTitle")}</span>
                  </h3>
                  <p className="text-xs text-slate-400 mt-0.5">
                    {t("settings.soulsSubtitle")}
                  </p>
                </div>
                <div className="flex items-center space-x-2 shrink-0">
                  <button
                    onClick={() => {
                      setEditingSoul({
                        id: `soul_${Date.now()}`,
                        name: t("settings.defaultNewSoulName"),
                        avatar: "🌸",
                        title: t("settings.defaultNewSoulRole"),
                        archetype: t("settings.defaultNewSoulArchetype"),
                        description: t("settings.defaultNewSoulDesc"),
                        greeting: t("settings.defaultNewSoulGreeting"),
                        on_confused: t("settings.defaultNewSoulConfused"),
                        on_success: t("settings.defaultNewSoulSuccess"),
                        system_tone: "gentle, literary, expressive",
                        current_mood: "idle",
                        system_prompt_addon: `# Persona Rules\n- Style: Gentle, literary, expressive\n- Tone: Respectful companion`,
                        is_default: false,
                      });
                      setIsCreatingSoul(true);
                      setSoulFeedback(null);
                    }}
                    className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow transition"
                  >
                    <Plus className="w-3.5 h-3.5" />
                    <span>{t("settings.createSoul")}</span>
                  </button>
                  <button
                    onClick={() => setShowSoulImportModal(true)}
                    className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-white/6 hover:bg-white/12 text-slate-300 text-xs font-semibold border border-white/10 transition"
                  >
                    <Upload className="w-3.5 h-3.5" />
                    <span>{t("settings.importSoul")}</span>
                  </button>
                  <button
                    onClick={async () => {
                      const confirmed = await showConfirmModal({
                        title: t("settings.restoreDefaultSoulsConfirmTitle"),
                        message: t("settings.restoreDefaultSoulsConfirm"),
                        confirmText: t("settings.restoreDefaultSouls"),
                        cancelText: t("common.cancel"),
                        variant: "warning",
                      });
                      if (confirmed) {
                        try {
                          await restoreDefaultSouls();
                          setSoulFeedback(
                            t("settings.restoreDefaultSoulsSuccess"),
                          );
                        } catch (err: any) {
                          setSoulFeedback(
                            t("settings.restoreDefaultSoulsError", {
                              error: err.message,
                            }),
                          );
                        }
                      }
                    }}
                    className="flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-white/6 hover:bg-white/12 text-slate-300 text-xs font-semibold border border-white/10 transition"
                    title={t("settings.restoreDefaultSoulsTooltip")}
                  >
                    <RotateCcw className="w-3.5 h-3.5" />
                    <span>{t("settings.restoreDefaultSouls")}</span>
                  </button>
                </div>
              </div>

              {soulFeedback && (
                <div className="p-3 rounded-lg bg-indigo-500/10 border border-indigo-500/30 text-xs text-indigo-300 flex items-center justify-between">
                  <span>{soulFeedback}</span>
                  <button
                    onClick={() => setSoulFeedback(null)}
                    className="text-slate-400 hover:text-white"
                  >
                    <X className="w-3.5 h-3.5" />
                  </button>
                </div>
              )}

              {/* Main Content: Left Cards Grid (5 cols), Right Editor/Preview (7 cols) */}
              <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                {/* Left: Soul List Cards */}
                <div className="lg:col-span-5 space-y-3">
                  <div className="text-xs font-semibold text-slate-400 uppercase tracking-wider px-1">
                    {t("settings.soulsListTitle")} (
                    {availableSouls?.length || 0})
                  </div>
                  <div className="space-y-2.5 max-h-137.5 overflow-y-auto pr-1">
                    {availableSouls?.map((s) => {
                      const isActive = soul?.id === s.id;
                      const isSelected = editingSoul?.id === s.id;
                      return (
                        <div
                          key={s.id}
                          className={`p-3.5 rounded-xl border transition cursor-pointer flex flex-col justify-between space-y-2.5 ${
                            isActive
                              ? "bg-indigo-500/10 border-indigo-500/40 ring-1 ring-indigo-500/30"
                              : isSelected
                                ? "bg-white/5 border-white/20"
                                : "bg-[#0f1422] border-white/6 hover:border-white/15"
                          }`}
                          onClick={() => {
                            setEditingSoul(s);
                            setIsCreatingSoul(false);
                            setSoulFeedback(null);
                          }}
                        >
                          <div className="flex items-start justify-between gap-2">
                            <div className="flex items-center space-x-2.5 min-w-0">
                              <span className="text-2xl shrink-0 p-1.5 rounded-lg bg-white/4 border border-white/10">
                                {s.avatar || "🐾"}
                              </span>
                              <div className="min-w-0">
                                <div className="flex items-center space-x-2">
                                  <h4 className="text-xs font-bold text-slate-100 truncate">
                                    {s.name}
                                  </h4>
                                  {isActive && (
                                    <span className="text-[9px] px-1.5 py-0.5 rounded bg-emerald-500/20 text-emerald-300 border border-emerald-500/30 font-semibold shrink-0">
                                      {t("settings.activeSoulBadge")}
                                    </span>
                                  )}
                                  {s.is_default && (
                                    <span className="text-[9px] px-1.5 py-0.5 rounded bg-white/6 text-slate-400 font-mono shrink-0">
                                      {t("settings.defaultSoulBadge")}
                                    </span>
                                  )}
                                </div>
                                <p className="text-[11px] text-indigo-300 font-medium truncate mt-0.5">
                                  {s.title || s.archetype}
                                </p>
                              </div>
                            </div>

                            <div
                              className="flex items-center space-x-1 shrink-0"
                              onClick={(e) => e.stopPropagation()}
                            >
                              {!isActive && (
                                <button
                                  onClick={async () => {
                                    await setActiveSoul(s.id);
                                    setSoulFeedback(
                                      t("settings.switchedSoulFeedback", {
                                        name: s.name,
                                      }),
                                    );
                                  }}
                                  className="px-2 py-1 rounded bg-indigo-500/20 hover:bg-indigo-500/30 text-indigo-200 border border-indigo-500/30 text-[10px] font-semibold transition flex items-center space-x-1"
                                  title={t("settings.selectSoulTooltip")}
                                >
                                  <UserCheck className="w-3 h-3" />
                                  <span>{t("settings.selectSoul")}</span>
                                </button>
                              )}
                              <button
                                onClick={async () => {
                                  try {
                                    const md = await exportSoul(s.id);
                                    navigator.clipboard.writeText(md);
                                    setSoulFeedback(
                                      t("settings.exportedSoulClipboard", {
                                        name: s.name,
                                      }),
                                    );
                                  } catch (err: any) {
                                    setSoulFeedback(
                                      t("settings.restoreDefaultSoulsError", {
                                        error: err.message,
                                      }),
                                    );
                                  }
                                }}
                                className="p-1 text-slate-400 hover:text-white rounded hover:bg-white/8 transition"
                                title={t("settings.exportSoulTooltip")}
                              >
                                <Download className="w-3.5 h-3.5" />
                              </button>
                              {!s.is_default &&
                                s.id !== "soul_neko_assistant" &&
                                s.id !== "soul_tieu_mai_wuxia" && (
                                  <button
                                    onClick={async () => {
                                      const confirmed = await showConfirmModal({
                                        title: t("settings.deleteSoulTitle"),
                                        message: t(
                                          "settings.deleteSoulConfirm",
                                          { name: s.name },
                                        ),
                                        confirmText: t("common.delete"),
                                        cancelText: t("common.cancel"),
                                        variant: "danger",
                                      });
                                      if (confirmed) {
                                        await deleteCustomSoul(s.id);
                                        if (editingSoul?.id === s.id) {
                                          setEditingSoul(null);
                                        }
                                        setSoulFeedback(
                                          t("settings.deletedSoulFeedback", {
                                            name: s.name,
                                          }),
                                        );
                                      }
                                    }}
                                    className="p-1 text-slate-400 hover:text-rose-400 rounded hover:bg-rose-500/10 transition"
                                    title={t("settings.deleteSoulTooltip")}
                                  >
                                    <Trash2 className="w-3.5 h-3.5" />
                                  </button>
                                )}
                            </div>
                          </div>
                          <p className="text-[11px] text-slate-400 line-clamp-2 italic">
                            &quot;{s.description || s.greeting}&quot;
                          </p>
                        </div>
                      );
                    })}
                  </div>
                </div>

                {/* Right: Form Editor & Live Preview */}
                <div className="lg:col-span-7 space-y-4">
                  {editingSoul ? (
                    <div className="p-4 rounded-xl bg-[#0f1422] border border-white/8 space-y-4">
                      <div className="flex items-center justify-between border-b border-white/8 pb-3">
                        <div className="flex items-center space-x-2">
                          <Edit3 className="w-4 h-4 text-indigo-400" />
                          <h4 className="text-xs font-bold text-slate-100">
                            {isCreatingSoul
                              ? t("settings.createNewSoulTitle")
                              : `${t("settings.editSoulTitle")}: ${editingSoul.name}`}
                          </h4>
                        </div>
                        <div className="flex items-center space-x-2">
                          <button
                            onClick={async () => {
                              setIsSavingSoul(true);
                              try {
                                await saveCustomSoul(editingSoul);
                                setSoulFeedback(
                                  t("settings.savedSoulFeedback", {
                                    name: editingSoul.name,
                                  }),
                                );
                                setIsCreatingSoul(false);
                              } catch (err: any) {
                                setSoulFeedback(err.message);
                              } finally {
                                setIsSavingSoul(false);
                              }
                            }}
                            disabled={
                              isSavingSoul ||
                              !editingSoul.name.trim() ||
                              !editingSoul.id.trim() ||
                              !/^[a-zA-Z0-9_-]+$/.test(editingSoul.id.trim())
                            }
                            className="px-3 py-1.5 rounded-lg bg-emerald-600 hover:bg-emerald-500 text-white text-xs font-semibold transition disabled:opacity-50"
                          >
                            {isSavingSoul
                              ? t("settings.savingSoul")
                              : t("settings.saveSoul")}
                          </button>
                          {soul?.id !== editingSoul.id && (
                            <button
                              onClick={async () => {
                                await saveCustomSoul(editingSoul);
                                await setActiveSoul(editingSoul.id);
                                setSoulFeedback(
                                  t("settings.savedAndActivatedSoulFeedback", {
                                    name: editingSoul.name,
                                  }),
                                );
                              }}
                              className="px-3 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold transition"
                            >
                              {t("settings.activateSoulNow")}
                            </button>
                          )}
                        </div>
                      </div>

                      {/* Avatar Quick Picker + Name */}
                      <div className="grid grid-cols-1 sm:grid-cols-12 gap-3">
                        <div className="sm:col-span-3 space-y-1">
                          <label className="text-[11px] text-slate-400 font-medium">
                            {t("settings.soulAvatarLabel")}:
                          </label>
                          <div className="flex items-center space-x-1.5">
                            <input
                              type="text"
                              value={editingSoul.avatar}
                              onChange={(e) =>
                                setEditingSoul({
                                  ...editingSoul,
                                  avatar: e.target.value,
                                })
                              }
                              className="w-12 h-9 text-center text-xl bg-[#121827] border border-white/10 rounded-lg focus:outline-none focus:border-indigo-500"
                              title={t("settings.soulAvatarTitle")}
                            />
                            <div className="flex flex-wrap gap-1 max-w-30">
                              {[
                                "🐱",
                                "⚔️",
                                "🌸",
                                "🦊",
                                "🧙‍♂️",
                                "🤖",
                                "👑",
                                "🐉",
                              ].map((em) => (
                                <button
                                  key={em}
                                  type="button"
                                  onClick={() =>
                                    setEditingSoul({
                                      ...editingSoul,
                                      avatar: em,
                                    })
                                  }
                                  className="w-5 h-5 rounded hover:bg-white/10 text-xs flex items-center justify-center transition"
                                >
                                  {em}
                                </button>
                              ))}
                            </div>
                          </div>
                        </div>

                        <div className="sm:col-span-5 space-y-1">
                          <label className="text-[11px] text-slate-400 font-medium">
                            {t("settings.soulNameLabel")} (*):
                          </label>
                          <input
                            type="text"
                            value={editingSoul.name}
                            onChange={(e) =>
                              setEditingSoul({
                                ...editingSoul,
                                name: e.target.value,
                              })
                            }
                            className="w-full h-9 px-3 text-xs bg-[#121827] border border-white/10 rounded-lg text-slate-200 focus:outline-none focus:border-indigo-500"
                            placeholder={t("settings.soulNamePlaceholder")}
                          />
                        </div>

                        <div className="sm:col-span-4 space-y-1">
                          <label className="text-[11px] text-slate-400 font-medium">
                            {t("settings.soulIdLabel")} (*):
                          </label>
                          <input
                            type="text"
                            value={editingSoul.id}
                            disabled={!isCreatingSoul}
                            onChange={(e) =>
                              setEditingSoul({
                                ...editingSoul,
                                id: e.target.value
                                  .toLowerCase()
                                  .replace(/[^a-z0-9_-]/g, "_"),
                              })
                            }
                            className="w-full h-9 px-3 text-xs bg-[#121827] border border-white/10 rounded-lg text-slate-200 focus:outline-none focus:border-indigo-500 disabled:opacity-60 font-mono"
                            placeholder="soul_wuxia_mai"
                          />
                          {editingSoul.id &&
                            !/^[a-zA-Z0-9_-]+$/.test(editingSoul.id) && (
                              <span className="text-[10px] text-rose-400">
                                {t("settings.soulIdHint")}
                              </span>
                            )}
                        </div>
                      </div>

                      {/* Title & Archetype */}
                      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                        <div className="space-y-1">
                          <label className="text-[11px] text-slate-400 font-medium">
                            {t("settings.soulRoleLabel")}:
                          </label>
                          <input
                            type="text"
                            value={editingSoul.title}
                            onChange={(e) =>
                              setEditingSoul({
                                ...editingSoul,
                                title: e.target.value,
                              })
                            }
                            className="w-full h-9 px-3 text-xs bg-[#121827] border border-white/10 rounded-lg text-slate-200 focus:outline-none focus:border-indigo-500"
                            placeholder={t("settings.soulRolePlaceholder")}
                          />
                        </div>
                        <div className="space-y-1">
                          <label className="text-[11px] text-slate-400 font-medium">
                            {t("settings.soulArchetypeLabel")}:
                          </label>
                          <input
                            type="text"
                            value={editingSoul.archetype}
                            onChange={(e) =>
                              setEditingSoul({
                                ...editingSoul,
                                archetype: e.target.value,
                              })
                            }
                            className="w-full h-9 px-3 text-xs bg-[#121827] border border-white/10 rounded-lg text-slate-200 focus:outline-none focus:border-indigo-500"
                            placeholder={t("settings.soulArchetypePlaceholder")}
                          />
                        </div>
                      </div>

                      {/* Description & Tone */}
                      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                        <div className="space-y-1">
                          <label className="text-[11px] text-slate-400 font-medium">
                            {t("settings.soulDescLabel")}:
                          </label>
                          <input
                            type="text"
                            value={editingSoul.description}
                            onChange={(e) =>
                              setEditingSoul({
                                ...editingSoul,
                                description: e.target.value,
                              })
                            }
                            className="w-full h-9 px-3 text-xs bg-[#121827] border border-white/10 rounded-lg text-slate-200 focus:outline-none focus:border-indigo-500"
                            placeholder={t("settings.soulDescPlaceholder")}
                          />
                        </div>
                        <div className="space-y-1">
                          <label className="text-[11px] text-slate-400 font-medium">
                            {t("settings.soulToneLabel")}:
                          </label>
                          <input
                            type="text"
                            value={editingSoul.system_tone}
                            onChange={(e) =>
                              setEditingSoul({
                                ...editingSoul,
                                system_tone: e.target.value,
                              })
                            }
                            className="w-full h-9 px-3 text-xs bg-[#121827] border border-white/10 rounded-lg text-slate-200 focus:outline-none focus:border-indigo-500 font-mono"
                            placeholder="e.g. lively, supportive, sharp"
                          />
                        </div>
                      </div>

                      {/* Interactive Dialogues: Greeting, OnConfused, OnSuccess */}
                      <div className="space-y-3 pt-2 border-t border-white/6">
                        <h5 className="text-[11px] font-bold text-amber-300 uppercase tracking-wider">
                          {t("settings.soulInteractiveDialogue")}
                        </h5>
                        <div className="space-y-2">
                          <div className="space-y-1">
                            <label className="text-[11px] text-slate-400 font-medium">
                              {t("settings.soulGreetingLabel")}:
                            </label>
                            <input
                              type="text"
                              value={editingSoul.greeting}
                              onChange={(e) =>
                                setEditingSoul({
                                  ...editingSoul,
                                  greeting: e.target.value,
                                })
                              }
                              className="w-full h-8 px-3 text-xs bg-[#121827] border border-white/10 rounded-lg text-slate-200 focus:outline-none focus:border-indigo-500"
                              placeholder={t(
                                "settings.soulGreetingPlaceholder",
                              )}
                            />
                          </div>
                          <div className="space-y-1">
                            <label className="text-[11px] text-slate-400 font-medium">
                              {t("settings.soulConfusedLabel")}:
                            </label>
                            <input
                              type="text"
                              value={editingSoul.on_confused}
                              onChange={(e) =>
                                setEditingSoul({
                                  ...editingSoul,
                                  on_confused: e.target.value,
                                })
                              }
                              className="w-full h-8 px-3 text-xs bg-[#121827] border border-white/10 rounded-lg text-slate-200 focus:outline-none focus:border-indigo-500"
                              placeholder={t(
                                "settings.soulConfusedPlaceholder",
                              )}
                            />
                          </div>
                          <div className="space-y-1">
                            <label className="text-[11px] text-slate-400 font-medium">
                              {t("settings.soulSuccessLabel")}:
                            </label>
                            <input
                              type="text"
                              value={editingSoul.on_success}
                              onChange={(e) =>
                                setEditingSoul({
                                  ...editingSoul,
                                  on_success: e.target.value,
                                })
                              }
                              className="w-full h-8 px-3 text-xs bg-[#121827] border border-white/10 rounded-lg text-slate-200 focus:outline-none focus:border-indigo-500"
                              placeholder={t("settings.soulSuccessPlaceholder")}
                            />
                          </div>
                        </div>
                      </div>

                      {/* Markdown SystemPromptAddon Editor */}
                      <div className="space-y-1 pt-2 border-t border-white/6">
                        <div className="flex items-center justify-between">
                          <label className="text-[11px] text-indigo-300 font-bold uppercase tracking-wider">
                            {t("settings.soulPromptRulesLabel")}
                          </label>
                          <span className="text-[10px] text-slate-500">
                            {t("settings.soulPromptRulesHint")}
                          </span>
                        </div>
                        <textarea
                          rows={6}
                          value={editingSoul.system_prompt_addon}
                          onChange={(e) =>
                            setEditingSoul({
                              ...editingSoul,
                              system_prompt_addon: e.target.value,
                            })
                          }
                          className="w-full bg-[#121827] border border-white/10 rounded-lg p-3 text-xs text-slate-200 focus:outline-none focus:border-indigo-500 font-mono resize-y transition leading-relaxed"
                          placeholder={t("settings.soulPromptRulesPlaceholder")}
                        />
                      </div>
                    </div>
                  ) : (
                    <div className="p-8 rounded-xl bg-[#0f1422] border border-white/8 text-center space-y-3">
                      <span className="text-4xl block">✨</span>
                      <h4 className="text-sm font-semibold text-slate-200">
                        {t("settings.soulSelectPlaceholder")}
                      </h4>
                      <p className="text-xs text-slate-400 max-w-sm mx-auto">
                        {t("settings.soulSelectDesc")}
                      </p>
                    </div>
                  )}
                </div>
              </div>

              {/* Soul Import Modal */}
              {showSoulImportModal && (
                <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/60 animate-in fade-in duration-150">
                  <div className="bg-[#111625] border border-white/15 rounded-2xl w-full max-w-lg p-5 space-y-4 shadow-2xl">
                    <div className="flex items-center justify-between border-b border-white/8 pb-3">
                      <h4 className="text-sm font-bold text-slate-100 flex items-center space-x-2">
                        <Upload className="w-4 h-4 text-indigo-400" />
                        <span>{t("settings.importSoulModalTitle")}</span>
                      </h4>
                      <button
                        onClick={() => setShowSoulImportModal(false)}
                        className="text-slate-400 hover:text-white"
                      >
                        <X className="w-4 h-4" />
                      </button>
                    </div>
                    <p className="text-xs text-slate-400">
                      {t("settings.importSoulModalDesc")}
                    </p>
                    <textarea
                      rows={10}
                      value={soulImportText}
                      onChange={(e) => setSoulImportText(e.target.value)}
                      placeholder={t("settings.importSoulModalPlaceholder")}
                      className="w-full bg-[#0b0e18] border border-white/10 rounded-xl p-3 text-xs text-slate-200 focus:outline-none focus:border-indigo-500 font-mono resize-none"
                    />
                    <div className="flex items-center justify-end space-x-2 pt-2">
                      <button
                        onClick={() => setShowSoulImportModal(false)}
                        className="px-3 py-1.5 bg-white/6 hover:bg-white/10 text-slate-300 rounded-lg text-xs font-semibold transition"
                      >
                        {t("common.cancel")}
                      </button>
                      <button
                        onClick={async () => {
                          if (!soulImportText.trim()) return;
                          try {
                            const imported = await importSoul(soulImportText);
                            if (imported) {
                              setEditingSoul(imported);
                              setSoulFeedback(
                                t("settings.importSoulSuccess", {
                                  name: imported.name,
                                }),
                              );
                              toast.success(
                                t("settings.importSoulSuccess", {
                                  name: imported.name,
                                }),
                              );
                            }
                            setShowSoulImportModal(false);
                            setSoulImportText("");
                          } catch (err: any) {
                            toast.error(
                              t("settings.importSoulError", {
                                error: err.message || err,
                              }),
                            );
                          }
                        }}
                        disabled={!soulImportText.trim()}
                        className="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold transition disabled:opacity-50"
                      >
                        {t("settings.importSoulBtn")}
                      </button>
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}

          {activeTab === "skills" &&
            (() => {
              const activeEnabledIds = skills
                .filter((s) => s.is_enabled)
                .map((s) => s.id);
              const isPresetActive = (
                presetId: string,
                fallbackCount: number,
                fallbackIds: string[],
              ) => {
                const preset = skillPresets.find((p) => p.id === presetId);
                if (
                  preset &&
                  preset.active_skill_ids &&
                  preset.active_skill_ids.length > 0
                ) {
                  if (
                    preset.active_skill_ids.length !== activeEnabledIds.length
                  )
                    return false;
                  return preset.active_skill_ids.every((id) =>
                    activeEnabledIds.includes(id),
                  );
                }
                if (activeEnabledIds.length !== fallbackCount) return false;
                return fallbackIds.every((id) => activeEnabledIds.includes(id));
              };

              const isEcoActive = isPresetActive("eco", 2, [
                "skill_literary_translator",
                "skill_foreign_sanitizer",
              ]);
              const isStandardActive = isPresetActive("standard", 5, [
                "skill_style_scout",
                "skill_entity_extraction",
                "skill_literary_translator",
                "skill_agentic_researcher",
                "skill_foreign_sanitizer",
              ]);
              const isPublishingActive = isPresetActive("publishing", 6, [
                "skill_style_scout",
                "skill_entity_extraction",
                "skill_literary_translator",
                "skill_shadow_critic",
                "skill_agentic_researcher",
                "skill_foreign_sanitizer",
              ]);

              const ecoTokens =
                skillPresets.find((p) => p.id === "eco")
                  ?.total_estimated_tokens ?? 470;
              const stdTokens =
                skillPresets.find((p) => p.id === "standard")
                  ?.total_estimated_tokens ?? 1600;
              const pubTokens =
                skillPresets.find((p) => p.id === "publishing")
                  ?.total_estimated_tokens ?? 2450;

              return (
                <div className="space-y-6">
                  {/* Header & Token Meter */}
                  <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 p-4 rounded-xl bg-linear-to-r from-indigo-950/40 via-purple-950/20 to-slate-900/60 border border-indigo-500/20">
                    <div className="space-y-1">
                      <div className="flex items-center space-x-2">
                        <Boxes className="w-5 h-5 text-indigo-400" />
                        <h3 className="text-sm font-semibold text-slate-100">
                          {t("settings.skillsRegistryTitle")}
                        </h3>
                      </div>
                      <p className="text-xs text-slate-400 max-w-2xl leading-relaxed">
                        {t("settings.skillsRegistrySubtitle")}
                      </p>
                    </div>
                    <div className="flex items-center shrink-0 space-x-3 bg-black/40 px-3.5 py-2.5 rounded-xl border border-white/10">
                      <Cpu className="w-4 h-4 text-indigo-400 shrink-0" />
                      <div>
                        <div className="text-[10px] text-slate-400 font-medium uppercase tracking-wider">
                          {t("settings.activePromptTokens")}
                        </div>
                        <div className="text-sm font-bold font-mono flex items-center space-x-1.5">
                          <span
                            className={
                              activeSkillsTokenWeight > 1200
                                ? "text-purple-400"
                                : activeSkillsTokenWeight > 500
                                  ? "text-amber-400"
                                  : "text-emerald-400"
                            }
                          >
                            ~{activeSkillsTokenWeight}
                          </span>
                          <span className="text-[11px] text-slate-400 font-normal">
                            {t("settings.tokensPerCall")}
                          </span>
                        </div>
                      </div>
                    </div>
                  </div>

                  {/* Feedback Notice */}
                  {skillFeedback && (
                    <div className="flex items-center space-x-2 p-3 bg-indigo-500/10 border border-indigo-500/30 text-indigo-200 text-xs rounded-xl">
                      <CheckCircle2 className="w-4 h-4 text-indigo-400 shrink-0" />
                      <span>{skillFeedback}</span>
                    </div>
                  )}

                  {/* Capability Presets Banner (1-Click Presets) */}
                  <div className="bg-[#101522] border border-white/8 rounded-xl p-4 space-y-3">
                    <div className="flex items-center justify-between">
                      <div className="flex items-center space-x-2">
                        <Zap className="w-4 h-4 text-amber-400" />
                        <span className="text-xs font-semibold text-slate-200 uppercase tracking-wider">
                          {t("settings.presetsTitle")}
                        </span>
                      </div>
                      <div className="flex items-center gap-2">
                        <span className="text-[11px] text-slate-400">
                          {t("settings.currentProject")}:{" "}
                          <span className="text-indigo-300 font-mono font-medium">
                            {selectedProject?.title ||
                              t("settings.globalDefault")}
                          </span>
                        </span>
                        <button
                          type="button"
                          onClick={async () => {
                            const confirmed = await showConfirmModal({
                              title: t("settings.restoreFactorySkills"),
                              message: t(
                                "settings.restoreFactorySkillsConfirm",
                              ),
                              confirmText: t("settings.restoreFactorySkills"),
                              cancelText: t("common.cancel"),
                              variant: "warning",
                            });
                            if (confirmed) {
                              await resetSkillsToDefault(selectedProject?.id);
                              setSkillFeedback(t("settings.skillResetSuccess"));
                            }
                          }}
                          className="px-2.5 py-1 bg-rose-500/10 hover:bg-rose-500/20 text-rose-300 border border-rose-500/25 rounded-md text-[11px] font-medium flex items-center gap-1 transition-colors"
                          title={t("settings.restoreFactorySkillsTooltip")}
                        >
                          <RotateCcw className="w-3 h-3" />
                          <span>{t("settings.restoreFactorySkillsBtn")}</span>
                        </button>
                      </div>
                    </div>

                    <div className="grid grid-cols-1 sm:grid-cols-3 gap-3">
                      {/* Preset Eco */}
                      <button
                        type="button"
                        onClick={() => handleApplySkillPreset("eco")}
                        className={`p-3 rounded-lg border text-left transition flex flex-col justify-between group ${
                          isEcoActive
                            ? "border-emerald-500 bg-emerald-500/15 ring-2 ring-emerald-500/40 shadow-sm"
                            : "border-emerald-500/20 bg-emerald-500/4 hover:bg-emerald-500/8 hover:border-emerald-500/40"
                        }`}
                      >
                        <div>
                          <div className="flex items-center justify-between mb-1">
                            <span className="text-xs font-bold text-emerald-300 flex items-center space-x-1.5">
                              <Zap className="w-3.5 h-3.5 text-emerald-400" />
                              <span>{t("settings.presetEcoTitle")}</span>
                            </span>
                            <div className="flex items-center space-x-1">
                              {isEcoActive && (
                                <span className="text-[10px] px-1.5 py-0.5 rounded bg-emerald-500 text-slate-950 font-bold flex items-center space-x-0.5">
                                  <Check className="w-2.5 h-2.5" />
                                  <span>{t("settings.presetEcoBadge")}</span>
                                </span>
                              )}
                              <span className="text-[10px] px-1.5 py-0.5 rounded bg-emerald-500/20 text-emerald-300 font-mono">
                                ~{ecoTokens} tok
                              </span>
                            </div>
                          </div>
                          <p className="text-[11px] text-slate-400 leading-snug">
                            {t("settings.presetEcoDesc")}
                          </p>
                        </div>
                        <div className="text-[10px] text-emerald-400 font-medium mt-2 group-hover:underline flex items-center space-x-1">
                          <span>
                            {isEcoActive
                              ? t("settings.presetActiveBadge")
                              : t("settings.applyEco")}
                          </span>
                          {!isEcoActive && <ArrowRight className="w-3 h-3" />}
                        </div>
                      </button>

                      {/* Preset Standard */}
                      <button
                        type="button"
                        onClick={() => handleApplySkillPreset("standard")}
                        className={`p-3 rounded-lg border text-left transition flex flex-col justify-between group ${
                          isStandardActive
                            ? "border-indigo-500 bg-indigo-500/15 ring-2 ring-indigo-500/40 shadow-sm"
                            : "border-indigo-500/20 bg-indigo-500/4 hover:bg-indigo-500/8 hover:border-indigo-500/40"
                        }`}
                      >
                        <div>
                          <div className="flex items-center justify-between mb-1">
                            <span className="text-xs font-bold text-indigo-300 flex items-center space-x-1.5">
                              <span>⚖️</span>
                              <span>{t("settings.presetStandardTitle")}</span>
                            </span>
                            <div className="flex items-center space-x-1">
                              {isStandardActive && (
                                <span className="text-[10px] px-1.5 py-0.5 rounded bg-indigo-500 text-slate-950 font-bold flex items-center space-x-0.5">
                                  <Check className="w-2.5 h-2.5" />
                                  <span>
                                    {t("settings.presetStandardBadge")}
                                  </span>
                                </span>
                              )}
                              <span className="text-[10px] px-1.5 py-0.5 rounded bg-indigo-500/20 text-indigo-300 font-mono">
                                ~{stdTokens} tok
                              </span>
                            </div>
                          </div>
                          <p className="text-[11px] text-slate-400 leading-snug">
                            {t("settings.presetStandardDesc")}
                          </p>
                        </div>
                        <div className="text-[10px] text-indigo-400 font-medium mt-2 group-hover:underline flex items-center space-x-1">
                          <span>
                            {isStandardActive
                              ? t("settings.presetActiveBadge")
                              : t("settings.applyStandard")}
                          </span>
                          {!isStandardActive && (
                            <ArrowRight className="w-3 h-3" />
                          )}
                        </div>
                      </button>

                      {/* Preset Publishing */}
                      <button
                        type="button"
                        onClick={() => handleApplySkillPreset("publishing")}
                        className={`p-3 rounded-lg border text-left transition flex flex-col justify-between group ${
                          isPublishingActive
                            ? "border-purple-500 bg-purple-500/15 ring-2 ring-purple-500/40 shadow-sm"
                            : "border-purple-500/20 bg-purple-500/4 hover:bg-purple-500/8 hover:border-purple-500/40"
                        }`}
                      >
                        <div>
                          <div className="flex items-center justify-between mb-1">
                            <span className="text-xs font-bold text-purple-300 flex items-center space-x-1.5">
                              <Crown className="w-3.5 h-3.5 text-purple-400" />
                              <span>{t("settings.presetPublishingTitle")}</span>
                            </span>
                            <div className="flex items-center space-x-1">
                              {isPublishingActive && (
                                <span className="text-[10px] px-1.5 py-0.5 rounded bg-purple-500 text-slate-950 font-bold flex items-center space-x-0.5">
                                  <Check className="w-2.5 h-2.5" />
                                  <span>
                                    {t("settings.presetPublishingBadge")}
                                  </span>
                                </span>
                              )}
                              <span className="text-[10px] px-1.5 py-0.5 rounded bg-purple-500/20 text-purple-300 font-mono">
                                ~{pubTokens} tok
                              </span>
                            </div>
                          </div>
                          <p className="text-[11px] text-slate-400 leading-snug">
                            {t("settings.presetPublishingDesc")}
                          </p>
                        </div>
                        <div className="text-[10px] text-purple-400 font-medium mt-2 group-hover:underline flex items-center space-x-1">
                          <span>
                            {isPublishingActive
                              ? t("settings.presetActiveBadge")
                              : t("settings.applyPublishing")}
                          </span>
                          {!isPublishingActive && (
                            <ArrowRight className="w-3 h-3" />
                          )}
                        </div>
                      </button>
                    </div>
                  </div>

                  {/* Skills Grid */}
                  <div className="space-y-3">
                    <div className="flex items-center justify-between">
                      <h4 className="text-xs font-semibold text-slate-300 uppercase tracking-wider flex items-center space-x-2">
                        <Layers className="w-4 h-4 text-indigo-400" />
                        <span>
                          {t("settings.skillsListTitle", {
                            count: skills.length,
                          })}
                        </span>
                      </h4>
                      {isSkillsLoading && (
                        <span className="text-[11px] text-slate-400 flex items-center space-x-1">
                          <RefreshCw className="w-3 h-3 animate-spin text-indigo-400" />
                          <span>{t("settings.syncing")}</span>
                        </span>
                      )}
                    </div>

                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      {skills.map((s) => {
                        const isEnabled = s.is_enabled;
                        const isLight =
                          s.weight_badge === "light" ||
                          s.weight_badge === "low";
                        const isMedium = s.weight_badge === "medium";
                        const weightColor = isLight
                          ? "bg-emerald-500/10 text-emerald-400 border-emerald-500/20"
                          : isMedium
                            ? "bg-amber-500/10 text-amber-400 border-amber-500/20"
                            : "bg-purple-500/10 text-purple-400 border-purple-500/20";

                        const weightLabel = isLight
                          ? t("settings.ultraLight")
                          : isMedium
                            ? t("settings.medium")
                            : t("settings.deep");

                        const renderSkillIcon = () => {
                          switch (s.id) {
                            case "skill_style_scout":
                              return (
                                <Sparkles className="w-4 h-4 text-purple-400" />
                              );
                            case "skill_entity_extraction":
                              return <Boxes className="w-4 h-4 text-sky-400" />;
                            case "skill_literary_translator":
                              return (
                                <BookOpen className="w-4 h-4 text-indigo-400" />
                              );
                            case "skill_shadow_critic":
                              return (
                                <ShieldAlert className="w-4 h-4 text-rose-400" />
                              );
                            case "skill_agentic_researcher":
                              return (
                                <Search className="w-4 h-4 text-amber-400" />
                              );
                            case "skill_foreign_sanitizer":
                              return (
                                <FileCheck2 className="w-4 h-4 text-emerald-400" />
                              );
                            default:
                              return (
                                <Zap className="w-4 h-4 text-indigo-400" />
                              );
                          }
                        };

                        return (
                          <div
                            key={s.id}
                            className={`rounded-xl border p-5 flex flex-col justify-between transition ${
                              isEnabled
                                ? "bg-[#101522] border-white/12 shadow-sm"
                                : "bg-[#0d101a]/70 border-white/4 opacity-75"
                            }`}
                          >
                            <div>
                              {/* Card Top Row */}
                              <div className="flex items-start justify-between gap-3">
                                <div className="flex items-start space-x-3 min-w-0 flex-1">
                                  <div className="w-8 h-8 rounded-lg bg-white/5 border border-white/10 flex items-center justify-center shrink-0 mt-0.5">
                                    {renderSkillIcon()}
                                  </div>
                                  <div>
                                    <div className="flex items-center space-x-2">
                                      <h5 className="text-xs font-semibold text-slate-100">
                                        {t(`settings.skillsList.${s.id}.name`, {
                                          defaultValue: s.name,
                                        })}
                                      </h5>
                                    </div>
                                    <div className="text-[10px] font-mono text-slate-500">
                                      {s.id}
                                    </div>
                                  </div>
                                </div>

                                {/* Toggle Switch */}
                                <button
                                  type="button"
                                  role="switch"
                                  aria-checked={isEnabled}
                                  onClick={() =>
                                    handleToggleSkill(s.id, isEnabled)
                                  }
                                  className={`relative inline-flex h-5 w-10 shrink-0 cursor-pointer rounded-full border-2 border-transparent transition-colors duration-200 ease-in-out focus:outline-none ${
                                    isEnabled ? "bg-indigo-600" : "bg-slate-700"
                                  }`}
                                >
                                  <span
                                    className={`pointer-events-none inline-block h-4 w-4 transform rounded-full bg-white shadow ring-0 transition duration-200 ease-in-out ${
                                      isEnabled
                                        ? "translate-x-5"
                                        : "translate-x-0"
                                    }`}
                                  />
                                </button>
                              </div>

                              {/* Badges Row */}
                              <div className="flex flex-wrap items-center gap-1.5 mt-3">
                                <span className="text-[10px] px-2 py-0.5 rounded-md bg-white/5 text-slate-300 font-medium border border-white/6">
                                  {s.category}
                                </span>
                                <span
                                  className={`text-[10px] px-2 py-0.5 rounded-md border font-medium font-mono ${weightColor}`}
                                >
                                  {weightLabel} (~{s.estimated_token_weight}{" "}
                                  tok)
                                </span>
                                {s.custom_rules && s.custom_rules.trim() && (
                                  <span className="text-[10px] px-1.5 py-0.5 rounded bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 flex items-center space-x-1">
                                    <Sparkles className="w-3 h-3 text-indigo-400" />
                                    <span>{t("settings.hasCustomRules")}</span>
                                  </span>
                                )}
                              </div>

                              {/* Description */}
                              <p className="text-xs text-slate-300 mt-2.5 leading-relaxed">
                                {t(`settings.skillsList.${s.id}.desc`, {
                                  defaultValue: s.description,
                                })}
                              </p>

                              {/* Rules summary */}
                              <div className="flex flex-wrap items-center gap-3 text-[11px] text-slate-400 mt-2.5">
                                {s.prohibited_rules &&
                                  s.prohibited_rules.length > 0 && (
                                    <span className="text-rose-400/90 flex items-center space-x-1">
                                      <Ban className="w-3 h-3 text-rose-400" />
                                      <span>
                                        {t("settings.prohibitedRulesCount", {
                                          count: s.prohibited_rules.length,
                                        })}
                                      </span>
                                    </span>
                                  )}
                                {s.few_shots && s.few_shots.length > 0 && (
                                  <span className="text-sky-400/90 flex items-center space-x-1">
                                    <Lightbulb className="w-3 h-3 text-sky-400" />
                                    <span>
                                      {t("settings.fewShotsCount", {
                                        count: s.few_shots.length,
                                      })}
                                    </span>
                                  </span>
                                )}
                              </div>
                            </div>

                            {/* Card Bottom Actions */}
                            <div className="flex items-center justify-between pt-3.5 mt-3.5 border-t border-white/6">
                              <button
                                type="button"
                                onClick={() => {
                                  setEditingSkill(s);
                                  setCustomRulesDraft(s.custom_rules || "");
                                }}
                                className="text-xs text-indigo-400 hover:text-indigo-300 flex items-center space-x-1 font-medium hover:underline"
                              >
                                <Edit3 className="w-3.5 h-3.5" />
                                <span>{t("settings.customizePromptBtn")}</span>
                              </button>

                              {(s.custom_rules || !isEnabled) && (
                                <button
                                  type="button"
                                  onClick={() => handleResetSkill(s.id)}
                                  title={t("settings.resetDefaultSkillTooltip")}
                                  className="text-[11px] text-slate-500 hover:text-slate-300 flex items-center space-x-1 transition"
                                >
                                  <RotateCcw className="w-3 h-3" />
                                  <span>
                                    {t("settings.resetDefaultSkillBtn")}
                                  </span>
                                </button>
                              )}
                            </div>
                          </div>
                        );
                      })}
                    </div>
                  </div>

                  {/* Custom Rules Modal Dialog */}
                  {editingSkill && (
                    <div className="fixed inset-0 z-50 bg-black/60 flex items-center justify-center p-4">
                      <div className="bg-[#101522] border border-white/10 rounded-2xl w-full max-w-2xl max-h-[85vh] flex flex-col shadow-2xl overflow-hidden animate-in fade-in duration-200">
                        {/* Dialog Header */}
                        <div className="flex items-center justify-between px-5 py-3.5 border-b border-white/8 bg-[#0b0e18]">
                          <div className="flex items-center space-x-2">
                            <Boxes className="w-4 h-4 text-indigo-400" />
                            <div>
                              <h4 className="text-xs font-bold text-slate-100">
                                {t("settings.editSkillTitle")}:{" "}
                                {t(
                                  `settings.skillsList.${editingSkill.id}.name`,
                                  { defaultValue: editingSkill.name },
                                )}
                              </h4>
                              <div className="text-[10px] font-mono text-slate-500">
                                {editingSkill.id}
                              </div>
                            </div>
                          </div>
                          <button
                            onClick={() => setEditingSkill(null)}
                            className="text-slate-400 hover:text-slate-200 p-1 rounded-lg hover:bg-white/6"
                          >
                            <X className="w-4 h-4" />
                          </button>
                        </div>

                        {/* Dialog Body */}
                        <div className="p-5 overflow-y-auto space-y-4">
                          {/* Built-in Prompt Preview */}
                          <div>
                            <label className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider mb-1 flex items-center space-x-1.5">
                              <BookOpen className="w-3.5 h-3.5 text-slate-400" />
                              <span>{t("settings.systemPromptOriginal")}</span>
                            </label>
                            <pre className="text-[11px] font-mono bg-[#0b0e18] border border-white/6 rounded-xl p-3 text-slate-300 overflow-x-auto max-h-36 whitespace-pre-wrap leading-relaxed">
                              {editingSkill.system_prompt_template}
                            </pre>
                          </div>

                          {/* Built-in Prohibited Rules */}
                          {editingSkill.prohibited_rules &&
                            editingSkill.prohibited_rules.length > 0 && (
                              <div>
                                <label className="text-[11px] font-semibold text-rose-400 uppercase tracking-wider mb-1.5 flex items-center space-x-1.5">
                                  <ShieldAlert className="w-3.5 h-3.5 text-rose-400" />
                                  <span>
                                    {t("settings.builtinRules")} (
                                    {editingSkill.prohibited_rules.length}):
                                  </span>
                                </label>
                                <div className="space-y-1 bg-rose-500/4 border border-rose-500/20 rounded-xl p-3">
                                  {editingSkill.prohibited_rules.map(
                                    (rule, idx) => (
                                      <div
                                        key={idx}
                                        className="text-xs text-rose-300 flex items-start space-x-2"
                                      >
                                        <span className="text-rose-400 mt-0.5">
                                          •
                                        </span>
                                        <span className="leading-snug">
                                          {rule}
                                        </span>
                                      </div>
                                    ),
                                  )}
                                </div>
                              </div>
                            )}

                          {/* Few-Shot Examples (if any) */}
                          {editingSkill.few_shots &&
                            editingSkill.few_shots.length > 0 && (
                              <div>
                                <label className="text-[11px] font-semibold text-sky-400 uppercase tracking-wider mb-1.5 flex items-center space-x-1.5">
                                  <Sparkles className="w-3.5 h-3.5 text-sky-400" />
                                  <span>{t("settings.fewShotsExamples")}</span>
                                </label>
                                <div className="space-y-2 max-h-40 overflow-y-auto pr-1">
                                  {editingSkill.few_shots.map((fs, idx) => (
                                    <div
                                      key={idx}
                                      className="bg-[#0b0e18] border border-white/6 rounded-xl p-2.5 text-xs space-y-1"
                                    >
                                      {fs.note && (
                                        <div className="text-[10px] text-amber-300/80 font-medium flex items-center gap-1">
                                          <Bookmark className="w-3 h-3 text-amber-400/80 shrink-0" />
                                          <span>{fs.note}</span>
                                        </div>
                                      )}
                                      <div className="text-slate-400 font-mono text-[11px]">
                                        <span className="text-slate-500 font-sans">
                                          {t("settings.fewShotInput")}:
                                        </span>{" "}
                                        {fs.input}
                                      </div>
                                      <div className="text-emerald-300 font-mono text-[11px]">
                                        <span className="text-slate-500 font-sans">
                                          {t("settings.fewShotOutput")}:
                                        </span>{" "}
                                        {fs.output}
                                      </div>
                                    </div>
                                  ))}
                                </div>
                              </div>
                            )}

                          {/* Custom User Rules Input */}
                          <div>
                            <div className="flex items-center justify-between mb-1.5">
                              <label className="text-[11px] font-semibold text-indigo-300 uppercase tracking-wider flex items-center space-x-1.5">
                                <Edit3 className="w-3.5 h-3.5 text-indigo-400" />
                                <span>{t("settings.projectCustomRules")}</span>
                              </label>
                              <span className="text-[10px] text-slate-500">
                                {t("settings.injectPromptNotice")}
                              </span>
                            </div>
                            <textarea
                              rows={6}
                              value={customRulesDraft}
                              onChange={(e) =>
                                setCustomRulesDraft(e.target.value)
                              }
                              placeholder={t("settings.customRulesPlaceholder")}
                              className="w-full bg-[#0b0e18] border border-white/10 focus:border-indigo-500 rounded-xl p-3 text-xs text-slate-200 focus:outline-none font-mono resize-none leading-relaxed transition"
                            />
                          </div>
                        </div>

                        {/* Dialog Footer */}
                        <div className="flex items-center justify-between px-5 py-3 border-t border-white/8 bg-[#0b0e18]">
                          <button
                            type="button"
                            onClick={() => handleResetSkill(editingSkill.id)}
                            className="px-3 py-1.5 bg-white/4 hover:bg-white/8 text-slate-400 hover:text-slate-200 rounded-lg text-xs font-medium border border-white/10 transition flex items-center space-x-1.5"
                          >
                            <RotateCcw className="w-3.5 h-3.5" />
                            <span>{t("settings.resetDefaultRules")}</span>
                          </button>

                          <div className="flex items-center space-x-2">
                            <button
                              type="button"
                              onClick={() => setEditingSkill(null)}
                              className="px-3.5 py-1.5 bg-white/6 hover:bg-white/10 text-slate-300 rounded-lg text-xs font-semibold transition"
                            >
                              {t("common.cancel")}
                            </button>
                            <button
                              type="button"
                              onClick={handleSaveCustomRules}
                              disabled={isSavingRules}
                              className="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold transition flex items-center space-x-1.5 disabled:opacity-50"
                            >
                              {isSavingRules ? (
                                <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                              ) : (
                                <CheckCircle2 className="w-3.5 h-3.5" />
                              )}
                              <span>{t("settings.saveRules")}</span>
                            </button>
                          </div>
                        </div>
                      </div>
                    </div>
                  )}
                </div>
              );
            })()}
          {activeTab === "about" && (
            <div className="max-w-xl mx-auto">
              <div className="bg-[#101522] border border-white/8 rounded-xl p-2 space-y-2 text-center">
                <div className="flex flex-col items-center gap-1.5">
                  <img
                    src="/novelclaw-logo.png"
                    alt="NovelClaw"
                    className="w-16 h-16 rounded-2xl shadow-lg shadow-indigo-500/20 bg-[#0a0d16] p-1"
                  />
                  <h3 className="text-base font-semibold text-slate-100">
                    {t("appName")}
                  </h3>
                  <p className="text-[11px] text-slate-400">{t("tagline")}</p>
                </div>

                <div className="grid grid-cols-2 gap-2 text-left">
                  <div className="bg-[#0d1220] border border-white/6 rounded-lg px-3 py-2.5">
                    <div className="text-[10px] text-slate-500 uppercase tracking-wider mb-0.5">
                      {t("update.version")}
                    </div>
                    <div className="text-sm font-mono font-semibold text-indigo-300">
                      {appInfo?.version || t("update.unknown")}
                    </div>
                  </div>
                  <div className="bg-[#0d1220] border border-white/6 rounded-lg px-3 py-2.5">
                    <div className="text-[10px] text-slate-500 uppercase tracking-wider mb-0.5">
                      {t("update.platform")}
                    </div>
                    <div className="text-sm font-mono font-semibold text-slate-200">
                      {appInfo?.platform || t("update.unknown")}
                    </div>
                  </div>
                  <div className="col-span-2 bg-[#0d1220] border border-white/6 rounded-lg px-3 py-2.5">
                    <div className="text-[10px] text-slate-500 uppercase tracking-wider mb-0.5">
                      {t("update.repository")}
                    </div>
                    <div className="text-xs font-mono text-slate-300 truncate">
                      {appInfo?.repo || t("update.unknown")}
                    </div>
                  </div>
                  <div className="col-span-2 bg-[#0d1220] border border-white/6 rounded-lg px-3 py-2.5">
                    <div className="text-[10px] text-slate-500 uppercase tracking-wider mb-0.5">
                      {t("update.dataDir")}
                    </div>
                    <div className="text-xs font-mono text-slate-300 break-all">
                      {appInfo?.data_dir || t("update.unknown")}
                    </div>
                  </div>
                </div>

                <button
                  type="button"
                  onClick={handleCheckForUpdates}
                  disabled={isCheckingUpdate}
                  className="w-full flex items-center justify-center space-x-2 px-4 py-2.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold transition disabled:opacity-50 disabled:cursor-not-allowed"
                >
                  {isCheckingUpdate ? (
                    <RefreshCw className="w-4 h-4 animate-spin" />
                  ) : (
                    <DownloadCloud className="w-4 h-4" />
                  )}
                  <span>
                    {isCheckingUpdate
                      ? t("update.checking")
                      : t("update.checkForUpdates")}
                  </span>
                </button>

                {updateFeedback && (
                  <p className="text-[11px] text-slate-400">
                    {t(updateFeedback)}
                  </p>
                )}
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-end px-6 py-3 border-t border-white/8 bg-[#0a0d16] shrink-0">
          <button
            type="button"
            onClick={() => setSettingsOpen(false)}
            className="px-4 py-2 bg-white/4 hover:bg-white/8 text-slate-300 rounded-lg text-xs font-semibold border border-white/10 transition"
          >
            {t("common.close")}
          </button>
        </div>
      </div>
    </div>
  );
};
