import React, { useState, useEffect, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { 
  Wand2, 
  Sparkles, 
  Check, 
  RefreshCw, 
  ShieldCheck, 
  Flame, 
  Heart, 
  Zap, 
  Smile, 
  X,
  AlertTriangle,
  Settings,
  Cpu,
  Plus,
  Trash2,
  Edit3,
  Lock,
  Bot,
  Palette,
  Loader2,
  BookOpen,
} from 'lucide-react';
import { useAppStore } from '@/store/useAppStore';
import { extractCleanNarrativeText } from '@/utils/readerHtml';
import { toast } from 'react-hot-toast';
import { showConfirmModal } from '@/store/confirmStore';

interface StyleProfile {
  id: string;
  name: string;
  description: string;
  prompt: string;
  iconName?: 'heart' | 'flame' | 'zap' | 'smile' | 'palette';
  isBuiltin?: boolean;
}

const getBuiltinPresets = (t: (k: string) => string): StyleProfile[] => [
  {
    id: 'romcom',
    name: t('styleScout.romcom'),
    description: t('styleScout.romcomDesc'),
    prompt: t('styleScout.romcomPrompt'),
    iconName: 'heart',
    isBuiltin: true,
  },
  {
    id: 'wuxia',
    name: t('styleScout.wuxia'),
    description: t('styleScout.wuxiaDesc'),
    prompt: t('styleScout.wuxiaPrompt'),
    iconName: 'flame',
    isBuiltin: true,
  },
  {
    id: 'hardboiled',
    name: t('styleScout.hardboiled'),
    description: t('styleScout.hardboiledDesc'),
    prompt: t('styleScout.hardboiledPrompt'),
    iconName: 'zap',
    isBuiltin: true,
  },
  {
    id: 'comedy',
    name: t('styleScout.comedy'),
    description: t('styleScout.comedyDesc'),
    prompt: t('styleScout.comedyPrompt'),
    iconName: 'smile',
    isBuiltin: true,
  },
];

const CUSTOM_STYLES_KEY = 'neko_custom_styles_store';

const DEFAULT_FALLBACK_RAW = `「ねえ、先輩……本当に私のこと、ただの後輩としか見てないんですか？」
夕暮れの教室。夕陽が彼女の長い髪を茜色に染めていた。いつもはからかってばかりの少女が、真剣な瞳でこちらを見つめている。
「え……あ、いや……」
言葉が喉に詰まる。心臓の鼓動が、静まり返った放課後の教室に響いているようだった。`;

/**
 * Chapter titles that never contain narrative prose (front matter, artwork,
 * indexes). Declared at module scope so the useMemo below stays a pure
 * computation — defining them inside the memo defeats React Compiler's
 * memoization preservation.
 */
const isSkipTitle = (title: string) => {
  const lower = (title || '').toLowerCase();
  return (
    lower.includes('cover') ||
    lower.includes('bìa') ||
    lower.includes('insert') ||
    lower.includes('mục lục') ||
    lower.includes('muc luc') ||
    lower.includes('table of contents') ||
    lower.includes('contents') ||
    lower.includes('content') ||
    lower.includes('目次') ||
    lower.includes('目录') ||
    lower.includes('toc') ||
    lower.includes('copyright') ||
    lower.includes('colophon') ||
    lower.includes('afterword') ||
    lower.includes('hậu ký') ||
    lower.includes('lời bạt') ||
    lower.includes('あとがき') ||
    lower.includes('illustration') ||
    lower.includes('gallery') ||
    lower.includes('pinup')
  );
};

/** True when a chapter's raw content is real story prose, not front matter. */
const isNarrativeProse = (raw: string, title?: string) => {
  if (title && isSkipTitle(title)) return false;
  const hasImages = raw.includes('<img') || raw.includes('<image') || raw.includes('<svg');
  const clean = extractCleanNarrativeText(raw);
  if (clean.length < 250) return false;
  if (hasImages && clean.length < 350) return false;
  const head = clean.slice(0, 100).toLowerCase();
  if (head.includes('contents') || head.includes('目次') || head.includes('mục lục') || head.includes('table of contents')) {
    return false;
  }
  return true;
};

/**
 * Pure selector: pick a clean narrative excerpt from the project's chapters.
 * Kept at module scope (and out of the component) so the component's useMemo
 * is a single call — that keeps React Compiler able to preserve the
 * memoization, which inline branching + DOMParser calls prevented.
 */
const pickNarrativeExcerpt = (
  selectedChapter: { raw_content?: string | null; title?: string } | null | undefined,
  chapters: { raw_content?: string | null; title?: string }[],
): string => {
  // 1. Prefer the selected chapter when it holds clean story prose.
  if (selectedChapter?.raw_content && isNarrativeProse(selectedChapter.raw_content, selectedChapter.title)) {
    return extractCleanNarrativeText(selectedChapter.raw_content).slice(0, 500);
  }
  // 2. Otherwise scan the chapter list for the first narrative chapter.
  for (const ch of chapters) {
    if (ch.raw_content && isNarrativeProse(ch.raw_content, ch.title)) {
      return extractCleanNarrativeText(ch.raw_content).slice(0, 500);
    }
  }
  return DEFAULT_FALLBACK_RAW;
};


export const StyleScoutModal: React.FC = () => {
  const { t } = useTranslation();

  const {
    isStyleScoutOpen,
    setStyleScoutOpen,
    selectedProject,
    selectedChapter,
    chapters,
    setStyleGuide,
    updateProjectStyle,
    setEnableR19,
    enableR19,
    fetchSoul,
    previewStyle,
    autoScoutStyle,
    setSettingsOpen,
  } = useAppStore();

  // Mode: AI Auto Scout (Default) vs Manual (Presets + Custom)
  const [scoutMode, setScoutMode] = useState<'manual' | 'ai'>('ai');

  // Ensure default is always AI Auto Scout when modal opens
  useEffect(() => {
    if (isStyleScoutOpen) {
      setScoutMode('ai');
    }
  }, [isStyleScoutOpen]);

  // Custom styles state
  const [customStyles, setCustomStyles] = useState<StyleProfile[]>(() => {
    try {
      const stored = localStorage.getItem(CUSTOM_STYLES_KEY);
      if (stored) return JSON.parse(stored);
    } catch {
      // ignore
    }
    return [];
  });

  // Active selected style in manual mode
  const [selectedStyleId, setSelectedStyleId] = useState<string>('romcom');

  // Custom Style Form (Add / Edit)
  const [isStyleFormOpen, setIsStyleFormOpen] = useState<boolean>(false);
  const [editingStyleId, setEditingStyleId] = useState<string | null>(null);
  const [formName, setFormName] = useState<string>('');
  const [formDesc, setFormDesc] = useState<string>('');
  const [formPrompt, setFormPrompt] = useState<string>('');

  // AI Auto Scout State
  const [isScoutingAI, setIsScoutingAI] = useState<boolean>(false);
  const [aiScoutedProfile, setAiScoutedProfile] = useState<{
    styleName: string;
    styleDescription: string;
    stylePrompt: string;
    analysisNotes: string;
    chaptersSampled: number;
  } | null>(null);
  const [isEditingAiProfile, setIsEditingAiProfile] = useState<boolean>(false);

  // Live preview state
  const [userCustomRaw, setUserCustomRaw] = useState<string | null>(null);
  const [isGenerating, setIsGenerating] = useState<boolean>(false);
  const [sampleTranslated, setSampleTranslated] = useState<string>('');
  const [latencyMs, setLatencyMs] = useState<number | null>(null);
  const [modelUsed, setModelUsed] = useState<string | null>(null);
  const [errorMessage, setErrorMessage] = useState<string | null>(null);

  // Persist custom styles to localStorage
  useEffect(() => {
    try {
      localStorage.setItem(CUSTOM_STYLES_KEY, JSON.stringify(customStyles));
    } catch {
      // ignore
    }
  }, [customStyles]);

  // Combine built-in presets and user custom styles
  const builtinPresets = useMemo(() => getBuiltinPresets(t), [t]);

  const allManualStyles = useMemo(() => {
    return [...builtinPresets, ...customStyles];
  }, [builtinPresets, customStyles]);

  // Determine current active style prompt
  const activeStylePrompt = useMemo(() => {
    if (scoutMode === 'ai' && aiScoutedProfile) {
      return aiScoutedProfile.stylePrompt;
    }
    const found = allManualStyles.find(s => s.id === selectedStyleId);
    return found ? found.prompt : (builtinPresets[0]?.prompt || '');
  }, [scoutMode, aiScoutedProfile, selectedStyleId, allManualStyles, builtinPresets]);

  // Clean narrative sample for the preview, derived by a pure module-level
  // selector so the memo body stays a single call.
  const initialNarrativeExcerpt = useMemo(
    () => pickNarrativeExcerpt(selectedChapter, chapters),
    [selectedChapter, chapters],
  );

  const sampleRaw = userCustomRaw !== null ? userCustomRaw : initialNarrativeExcerpt;

  if (!isStyleScoutOpen || !selectedProject) return null;

  // Run live translation preview with the currently active style
  const handleRunAIPreview = async (stylePromptOverride?: string) => {
    const rawToTranslate = sampleRaw.trim();
    if (!rawToTranslate) {
      setErrorMessage(t('styleScout.errSelectSample'));
      return;
    }

    const activePrompt = stylePromptOverride || activeStylePrompt;
    setIsGenerating(true);
    setErrorMessage(null);

    try {
      const res = await previewStyle({
        project_id: selectedProject.id,
        raw_content: rawToTranslate,
        style_prompt: activePrompt,
        source_lang: selectedProject.source_lang,
        target_lang: selectedProject.target_lang,
      });

      if (res && res.success) {
        setSampleTranslated(res.translated_content);
        setLatencyMs(res.latency_ms);
        setModelUsed(res.model_used);
      } else {
        setErrorMessage(res?.error || t('styleScout.errPreviewFailed'));
      }
    } catch (err: any) {
      setErrorMessage(err?.message || t('styleScout.errConnection'));
    } finally {
      setIsGenerating(false);
    }
  };

  // Run AI Auto Scout across 3 opening narrative chapters
  const handleStartAutoScout = async () => {
    setIsScoutingAI(true);
    setErrorMessage(null);

    try {
      const res = await autoScoutStyle({
        project_id: selectedProject.id,
        source_lang: selectedProject.source_lang,
        target_lang: selectedProject.target_lang,
      });

      if (res && res.success) {
        setAiScoutedProfile({
          styleName: res.style_name,
          styleDescription: res.style_description,
          stylePrompt: res.style_prompt,
          analysisNotes: res.analysis_notes,
          chaptersSampled: res.chapters_sampled || 3,
        });

        if (selectedProject) {
          await updateProjectStyle(selectedProject.id, res.style_name, res.style_prompt);
        }
        setStyleGuide(res.style_prompt);

        if (res.sample_translated) {
          setSampleTranslated(res.sample_translated);
          setLatencyMs(res.latency_ms);
          setModelUsed(res.model_used);
        } else {
          // Trigger live preview with the scouted style
          handleRunAIPreview(res.style_prompt);
        }
      } else {
        setErrorMessage(res?.error || t('styleScout.errScoutFailed'));
      }
    } catch (err: any) {
      setErrorMessage(err?.message || t('styleScout.errScoutActivate'));
    } finally {
      setIsScoutingAI(false);
    }
  };

  // Select a manual style (preset or custom)
  const handleSelectManualStyle = (styleId: string) => {
    setSelectedStyleId(styleId);
    const p = allManualStyles.find(x => x.id === styleId);
    if (p) {
      handleRunAIPreview(p.prompt);
    }
  };

  // Open Create Custom Style form
  const handleOpenCreateForm = () => {
    setEditingStyleId(null);
    setFormName('');
    setFormDesc('');
    setFormPrompt('');
    setIsStyleFormOpen(true);
  };

  // Open Edit Custom Style form
  const handleOpenEditForm = (style: StyleProfile) => {
    setEditingStyleId(style.id);
    setFormName(style.name);
    setFormDesc(style.description);
    setFormPrompt(style.prompt);
    setIsStyleFormOpen(true);
  };

  // Save Custom Style
  const handleSaveCustomStyle = () => {
    if (!formName.trim() || !formPrompt.trim()) {
      setErrorMessage(t('styleScout.errFillForm'));
      return;
    }

    if (editingStyleId) {
      // Update existing
      setCustomStyles(prev => prev.map(s => s.id === editingStyleId ? {
        ...s,
        name: formName.trim(),
        description: formDesc.trim() || t('styleScout.customStyleDefaultDesc'),
        prompt: formPrompt.trim(),
      } : s));
    } else {
      // Create new
      const newStyle: StyleProfile = {
        id: `custom_style_${Date.now()}`,
        name: formName.trim(),
        description: formDesc.trim() || t('styleScout.userCreatedStyleDesc'),
        prompt: formPrompt.trim(),
        iconName: 'palette',
        isBuiltin: false,
      };
      setCustomStyles(prev => [...prev, newStyle]);
      setSelectedStyleId(newStyle.id);
      handleRunAIPreview(newStyle.prompt);
    }

    setIsStyleFormOpen(false);
    setErrorMessage(null);
  };

  // Delete Custom Style
  const handleDeleteCustomStyle = async (styleId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    const confirmed = await showConfirmModal({
      title: t('styleScout.deleteModalTitle'),
      message: t('styleScout.confirmDelete'),
      confirmText: t('common.delete'),
      cancelText: t('common.cancel'),
      variant: 'danger',
    });
    if (confirmed) {
      setCustomStyles(prev => prev.filter(s => s.id !== styleId));
      if (selectedStyleId === styleId) {
        setSelectedStyleId('romcom');
      }
      toast.success(t('styleScout.deletedToast'));
    }
  };

  // Save AI Profile as Custom Style into library
  const handleSaveAiProfileAsCustom = () => {
    if (!aiScoutedProfile) return;
    const newStyle: StyleProfile = {
      id: `ai_scout_${Date.now()}`,
      name: aiScoutedProfile.styleName,
      description: aiScoutedProfile.styleDescription,
      prompt: aiScoutedProfile.stylePrompt,
      iconName: 'palette',
      isBuiltin: false,
    };
    setCustomStyles(prev => [...prev, newStyle]);
    toast.success(t('styleScout.savedToLibrary'));
  };

  // Confirm and apply style guide to project
  const handleConfirmStyle = async () => {
    const styleName = scoutMode === 'ai' && aiScoutedProfile 
      ? aiScoutedProfile.styleName 
      : allManualStyles.find(s => s.id === selectedStyleId)?.name || 'Custom Style';
    if (selectedProject) {
      await updateProjectStyle(selectedProject.id, styleName, activeStylePrompt);
    }
    setStyleGuide(activeStylePrompt);
    fetchSoul();
    setStyleScoutOpen(false);
    toast.success(t('styleScout.appliedToast'));
  };

  const renderStyleIcon = (iconName?: string) => {
    switch (iconName) {
      case 'heart': return <Heart className="w-4 h-4 text-rose-400" />;
      case 'flame': return <Flame className="w-4 h-4 text-amber-400" />;
      case 'zap': return <Zap className="w-4 h-4 text-cyan-400" />;
      case 'smile': return <Smile className="w-4 h-4 text-emerald-400" />;
      default: return <Palette className="w-4 h-4 text-indigo-400" />;
    }
  };

  return (
    // 1. Plain black overlay opacity without blur
    <div className="fixed inset-0 bg-black/75 flex items-center justify-center z-50 p-4">
      <div className="bg-[#0c101a] border border-white/10rounded-2xl max-w-4xl w-full h-[90vh] max-h-[90vh] shadow-2xl flex flex-col overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        
        {/* Modal Header */}
        <div className="flex items-center justify-between px-6 py-4 border-b border-white/8 bg-[#0e1422] shrink-0">
          <div className="flex items-center space-x-3">
            <div className="p-2 bg-indigo-500/10 text-indigo-400 rounded-lg border border-indigo-500/20">
              <Wand2 className="w-5 h-5" />
            </div>
            <div>
              <h2 className="text-sm font-bold text-slate-100 flex items-center space-x-2">
                <span>{t('styleScout.managementTitle')}</span>
                <span className="text-[10px] uppercase font-mono px-2 py-0.5 rounded bg-indigo-500/10 text-indigo-300 border border-indigo-500/20">
                  {t('styleScout.aiToneDirective')}
                </span>
              </h2>
              <p className="text-xs text-slate-400 mt-0.5">
                {t('styleScout.headerDesc')}
              </p>
            </div>
          </div>
          <button
            onClick={() => setStyleScoutOpen(false)}
            className="text-slate-400 hover:text-slate-200 p-1.5 rounded-lg hover:bg-white/6 transition"
          >
            <X className="w-5 h-5" />
          </button>
        </div>

        {/* Modal Body */}
        <div className="flex-1 min-h-0 overflow-y-auto px-6 py-5 space-y-5">
          
          {/* Mode Switcher: AI Auto Scout (Default) vs Manual */}
          <div className="bg-[#0a0d16] p-1 rounded-xl border border-white/10 flex items-center gap-1.5">
            <button
              type="button"
              onClick={() => setScoutMode('ai')}
              className={`flex-1 py-2 px-3 rounded-lg text-xs font-semibold flex items-center justify-center space-x-2 transition ${
                scoutMode === 'ai'
                  ? 'bg-indigo-600 text-white shadow-sm border border-indigo-400/30'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-white/3'
              }`}
            >
              <Bot className="w-3.5 h-3.5 text-indigo-300" />
              <span>{t('styleScout.aiModeTab')}</span>
            </button>
            <button
              type="button"
              onClick={() => setScoutMode('manual')}
              className={`flex-1 py-2 px-3 rounded-lg text-xs font-semibold flex items-center justify-center space-x-2 transition ${
                scoutMode === 'manual'
                  ? 'bg-indigo-600 text-white shadow-sm border border-indigo-400/30'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-white/3'
              }`}
            >
              <Palette className="w-3.5 h-3.5" />
              <span>{t('styleScout.manualModeTab')}</span>
            </button>
          </div>

          {/* SECTION 1: Manual Mode & Custom Styles */}
          {scoutMode === 'manual' ? (
            <div className="space-y-3">
              <div className="flex items-center justify-between">
                <span className="text-xs font-semibold uppercase tracking-wider text-slate-400 flex items-center space-x-1.5">
                  <Palette className="w-3.5 h-3.5 text-indigo-400" />
                  <span>{t('styleScout.stylesList')}</span>
                </span>
                <button
                  type="button"
                  onClick={handleOpenCreateForm}
                  className="flex items-center space-x-1.5 px-3 py-1 bg-indigo-600/20 hover:bg-indigo-600/30 text-indigo-300 rounded-lg text-xs font-semibold border border-indigo-500/30 transition"
                >
                  <Plus className="w-3.5 h-3.5" />
                  <span>{t('styleScout.createNewStyle')}</span>
                </button>
              </div>

              {/* Grid of Styles */}
              <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-2.5">
                {allManualStyles.map(p => {
                  const isSelected = selectedStyleId === p.id;
                  return (
                    <div
                      key={p.id}
                      onClick={() => handleSelectManualStyle(p.id)}
                      className={`p-3 rounded-xl border text-left transition flex flex-col justify-between cursor-pointer relative group ${
                        isSelected
                          ? 'border-indigo-500/60 bg-indigo-500/10 ring-1 ring-indigo-500/40 shadow-xs'
                          : 'border-white/8 bg-white/2 hover:border-white/20'
                      }`}
                    >
                      <div>
                        <div className="flex items-center justify-between mb-1.5">
                          <div className="flex items-center space-x-2">
                            {renderStyleIcon(p.iconName)}
                            <span className="text-xs font-semibold text-slate-200">{p.name}</span>
                          </div>
                          {!p.isBuiltin && (
                            <div className="flex items-center space-x-1 opacity-80 group-hover:opacity-100">
                              <button
                                type="button"
                                title={t('styleScout.edit')}
                                onClick={(e) => {
                                  e.stopPropagation();
                                  handleOpenEditForm(p);
                                }}
                                className="p-1 hover:text-indigo-300 text-slate-400 rounded hover:bg-white/10"
                              >
                                <Edit3 className="w-3 h-3" />
                              </button>
                              <button
                                type="button"
                                title={t('styleScout.delete')}
                                onClick={(e) => handleDeleteCustomStyle(p.id, e)}
                                className="p-1 hover:text-rose-400 text-slate-400 rounded hover:bg-white/10"
                              >
                                <Trash2 className="w-3 h-3" />
                              </button>
                            </div>
                          )}
                        </div>
                        <p className="text-[10.5px] text-slate-400 line-clamp-2 leading-relaxed">
                          {p.description}
                        </p>
                      </div>

                      <div className="mt-2.5 pt-2 border-t border-white/6 flex items-center justify-between">
                        <span className="text-[9px] font-mono text-slate-500 uppercase">
                          {p.isBuiltin ? t('styleScout.default') : t('styleScout.custom')}
                        </span>
                        {isSelected && (
                          <span className="text-[10px] font-semibold text-indigo-400 flex items-center space-x-1">
                            <Check className="w-3 h-3" />
                            <span>{t('styleScout.selected')}</span>
                          </span>
                        )}
                      </div>
                    </div>
                  );
                })}
              </div>

              {/* Inline modal/form to create or edit a style */}
              {isStyleFormOpen && (
                <div className="bg-[#101524] border border-indigo-500/40 rounded-xl p-4 space-y-3 animate-in fade-in duration-150">
                  <div className="flex items-center justify-between border-b border-white/10 pb-2">
                    <span className="text-xs font-bold text-indigo-200 flex items-center space-x-1.5">
                      <Plus className="w-3.5 h-3.5 text-indigo-400" />
                      <span>{editingStyleId ? t('styleScout.editStyleTitle') : t('styleScout.createStyleTitle')}</span>
                    </span>
                    <button
                      type="button"
                      onClick={() => setIsStyleFormOpen(false)}
                      className="text-slate-400 hover:text-slate-200 p-1"
                    >
                      <X className="w-4 h-4" />
                    </button>
                  </div>

                  <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                    <div>
                      <label className="block text-[11px] font-semibold text-slate-300 mb-1">
                        {t('styleScout.styleNameLabel')}
                      </label>
                      <input
                        type="text"
                        value={formName}
                        onChange={(e) => setFormName(e.target.value)}
                        placeholder={t('styleScout.styleNamePlaceholder')}
                        className="w-full bg-[#080c14] border border-white/10 focus:border-indigo-500 rounded-lg px-3 py-1.5 text-xs text-slate-100 font-sans focus:outline-none"
                      />
                    </div>
                    <div>
                      <label className="block text-[11px] font-semibold text-slate-300 mb-1">
                        {t('styleScout.shortDescLabel')}
                      </label>
                      <input
                        type="text"
                        value={formDesc}
                        onChange={(e) => setFormDesc(e.target.value)}
                        placeholder={t('styleScout.shortDescPlaceholder')}
                        className="w-full bg-[#080c14] border border-white/10 focus:border-indigo-500 rounded-lg px-3 py-1.5 text-xs text-slate-100 font-sans focus:outline-none"
                      />
                    </div>
                  </div>

                  <div>
                    <label className="block text-[11px] font-semibold text-slate-300 mb-1">
                      {t('styleScout.promptRulesLabel')}
                    </label>
                    <textarea
                      rows={3}
                      value={formPrompt}
                      onChange={(e) => setFormPrompt(e.target.value)}
                      placeholder={t('styleScout.promptRulesPlaceholder')}
                      className="w-full bg-[#080c14] border border-white/10 focus:border-indigo-500 rounded-lg p-2.5 text-xs text-slate-100 font-sans leading-relaxed focus:outline-none"
                    />
                  </div>

                  <div className="flex items-center justify-end space-x-2 pt-1">
                    <button
                      type="button"
                      onClick={() => setIsStyleFormOpen(false)}
                      className="px-3 py-1.5 bg-white/4 hover:bg-white/8 text-slate-300 rounded-lg text-xs font-medium border border-white/10 transition"
                    >
                      {t('common.cancel')}
                    </button>
                    <button
                      type="button"
                      onClick={handleSaveCustomStyle}
                      className="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold transition"
                    >
                      {t('styleScout.saveStyleBtn')}
                    </button>
                  </div>
                </div>
              )}
            </div>
          ) : (
            /* SECTION 2: AI Auto Scout Mode */
            <div className="space-y-4">
              {/* Visual Disabled Notice for Manual Styles */}
              <div className="p-3.5 bg-indigo-950/20 border border-indigo-500/20 rounded-xl flex items-center justify-between text-xs text-slate-300">
                <div className="flex items-center space-x-2">
                  <Lock className="w-4 h-4 text-indigo-400 shrink-0" />
                  <span>{t('styleScout.manualLockedNotice')}</span>
                </div>
                {!aiScoutedProfile && (
                  <button
                    type="button"
                    onClick={handleStartAutoScout}
                    disabled={isScoutingAI}
                    className="flex items-center space-x-1.5 px-4 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold transition shadow-sm border border-indigo-400/30 shrink-0 ml-3 disabled:opacity-50"
                  >
                    {isScoutingAI ? (
                      <Loader2 className="w-4 h-4 animate-spin text-white" />
                    ) : (
                      <Bot className="w-4 h-4" />
                    )}
                    <span>{isScoutingAI ? t('styleScout.aiScoutingBtn') : t('styleScout.aiScoutStartBtn')}</span>
                  </button>
                )}
              </div>

              {/* AI Scouting In-Progress Animation */}
              {isScoutingAI && (
                <div className="p-7 sm:p-8 rounded-xl bg-[#090d16] border border-indigo-500/30 flex flex-col items-center justify-center text-center space-y-4 shadow-inner">
                  <div className="relative flex items-center justify-center w-14 h-14 rounded-2xl bg-indigo-500/10 border border-indigo-500/25 shadow-md shadow-indigo-500/10">
                    <BookOpen className="w-7 h-7 text-indigo-400 animate-pulse" />
                    <div className="absolute -inset-1 rounded-2xl border border-indigo-400/20 animate-ping pointer-events-none opacity-40" />
                  </div>
                  <div className="space-y-1.5 flex flex-col items-center">
                    <div className="inline-flex items-center space-x-2 px-3 py-1 rounded-full bg-indigo-500/10 border border-indigo-500/20 text-indigo-300">
                      <Loader2 className="w-3.5 h-3.5 animate-spin text-indigo-400 shrink-0" />
                      <h4 className="text-xs font-bold text-indigo-200">
                        {t('styleScout.skimmingNotice')}
                      </h4>
                    </div>
                    <p className="text-xs text-slate-400 max-w-md leading-relaxed pt-1">
                      {t('styleScout.agentExtracting')}
                    </p>
                  </div>
                </div>
              )}

              {/* AI Generated Style Profile Display */}
              {aiScoutedProfile && (
                <div className="bg-[#0a0f1d] border border-indigo-500/40 rounded-xl p-5 sm:p-6 space-y-4 shadow-sm animate-in fade-in duration-200">
                  <div className="flex items-start justify-between">
                    <div>
                      <div className="flex items-center space-x-2">
                        <Sparkles className="w-4 h-4 text-amber-400" />
                        <span className="text-[10px] uppercase font-mono px-2 py-0.5 rounded bg-indigo-500/20 text-indigo-300 border border-indigo-500/30">
                          {t('styleScout.aiGeneratedBadge', { count: aiScoutedProfile.chaptersSampled })}
                        </span>
                      </div>
                      <h3 className="text-sm font-bold text-slate-100 mt-1.5 flex items-center space-x-2">
                        <span>{aiScoutedProfile.styleName}</span>
                      </h3>
                      <p className="text-xs text-slate-300 mt-0.5 leading-relaxed">
                        {aiScoutedProfile.styleDescription}
                      </p>
                    </div>

                    <div className="flex items-center space-x-2 shrink-0">
                      <button
                        type="button"
                        onClick={() => setIsEditingAiProfile(prev => !prev)}
                        className="flex items-center space-x-1 px-3 py-1.5 bg-white/4 hover:bg-white/8 text-slate-300 rounded-lg text-xs font-medium border border-white/10 transition"
                      >
                        <Edit3 className="w-3.5 h-3.5 text-indigo-400" />
                        <span>{isEditingAiProfile ? t('styleScout.closeEdit') : t('styleScout.editStyle')}</span>
                      </button>
                      <button
                        type="button"
                        onClick={handleSaveAiProfileAsCustom}
                        className="flex items-center space-x-1 px-3 py-1.5 bg-indigo-600/20 hover:bg-indigo-600/30 text-indigo-300 rounded-lg text-xs font-semibold border border-indigo-500/30 transition"
                        title={t('styleScout.saveToLibrary')}
                      >
                        <Plus className="w-3.5 h-3.5" />
                        <span>{t('styleScout.saveToLibrary')}</span>
                      </button>
                      <button
                        type="button"
                        onClick={handleStartAutoScout}
                        disabled={isScoutingAI}
                        className="flex items-center space-x-1 px-2.5 py-1.5 bg-white/2 hover:bg-white/6 text-slate-400 rounded-lg text-xs transition"
                        title={t('styleScout.reanalyze')}
                      >
                        <RefreshCw className="w-3.5 h-3.5" />
                      </button>
                    </div>
                  </div>

                  {/* Inline Editor for AI Profile */}
                  {isEditingAiProfile ? (
                    <div className="bg-[#080c14] border border-white/10 rounded-lg p-3 space-y-2.5">
                      <div>
                        <label className="block text-[10.5px] font-semibold text-slate-400 mb-1">
                          {t('styleScout.styleName')}
                        </label>
                        <input
                          type="text"
                          value={aiScoutedProfile.styleName}
                          onChange={(e) => setAiScoutedProfile(prev => prev ? { ...prev, styleName: e.target.value } : null)}
                          className="w-full bg-[#111726] border border-white/10 rounded px-2.5 py-1 text-xs text-slate-100"
                        />
                      </div>
                      <div>
                        <label className="block text-[10.5px] font-semibold text-slate-400 mb-1">
                          {t('styleScout.generalDesc')}
                        </label>
                        <input
                          type="text"
                          value={aiScoutedProfile.styleDescription}
                          onChange={(e) => setAiScoutedProfile(prev => prev ? { ...prev, styleDescription: e.target.value } : null)}
                          className="w-full bg-[#111726] border border-white/10 rounded px-2.5 py-1 text-xs text-slate-100"
                        />
                      </div>
                      <div>
                        <label className="block text-[10.5px] font-semibold text-slate-400 mb-1">
                          {t('styleScout.proposedPrompt')}
                        </label>
                        <textarea
                          rows={3}
                          value={aiScoutedProfile.stylePrompt}
                          onChange={(e) => setAiScoutedProfile(prev => prev ? { ...prev, stylePrompt: e.target.value } : null)}
                          className="w-full bg-[#111726] border border-white/10 rounded p-2 text-xs text-slate-100 leading-relaxed font-sans"
                        />
                      </div>
                    </div>
                  ) : (
                    <>
                      {/* Analysis Findings */}
                      {aiScoutedProfile.analysisNotes && (
                        <div className="bg-[#080c14] border border-white/6 rounded-lg p-3">
                          <div className="text-[10px] uppercase font-mono font-semibold text-indigo-300 mb-1">
                            {t('styleScout.insights3Chapters')}
                          </div>
                          <div className="text-xs text-slate-300 leading-relaxed whitespace-pre-wrap">
                            {aiScoutedProfile.analysisNotes}
                          </div>
                        </div>
                      )}

                      {/* Style Prompt Directives */}
                      <div className="bg-[#080c14] border border-white/6 rounded-lg p-3">
                        <div className="text-[10px] uppercase font-mono font-semibold text-slate-400 mb-1">
                          {t('styleScout.appliedRules')}
                        </div>
                        <p className="text-xs text-indigo-200/90 leading-relaxed italic">
                          "{aiScoutedProfile.stylePrompt}"
                        </p>
                      </div>
                    </>
                  )}
                </div>
              )}
            </div>
          )}

          {/* SECTION 3: Side-by-side Live AI Showcase Sample (clean sample text without HTML junk) */}
          <div className="bg-[#090d18] border border-white/8 rounded-xl p-5 space-y-4 shadow-sm">
            <div className="flex items-center justify-between">
              <span className="text-xs font-semibold uppercase tracking-wider text-slate-300 flex items-center space-x-2">
                <Sparkles className="w-4 h-4 text-indigo-400" />
                <span>{t('styleScout.liveAiTest')}</span>
              </span>

              <div className="flex items-center space-x-2">
                {latencyMs !== null && (
                  <span className="text-[10px] font-mono px-2.5 py-1 rounded-md bg-white/4 border border-white/8 text-slate-300 flex items-center space-x-1.5">
                    <Cpu className="w-3 h-3 text-indigo-400" />
                    <span>{modelUsed || 'AI'} • {latencyMs}ms</span>
                  </span>
                )}

                <button
                  type="button"
                  onClick={() => handleRunAIPreview()}
                  disabled={isGenerating}
                  className="flex items-center space-x-1.5 px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold border border-indigo-400/30 transition shadow-xs disabled:opacity-50"
                >
                  <RefreshCw className={`w-3 h-3 ${isGenerating ? 'animate-spin' : ''}`} />
                  <span>{isGenerating ? t('styleScout.aiTranslatingSample') : t('styleScout.retranslateSample')}</span>
                </button>
              </div>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
              {/* Left Column: Editable Clean Excerpt */}
              <div className="flex flex-col bg-[#060912] border border-white/8 rounded-xl p-4 space-y-2">
                <div className="flex items-center justify-between pb-1 border-b border-white/4">
                  <span className="text-[10px] text-slate-400 uppercase font-mono font-semibold">
                    {t('styleScout.cleanExcerpt')}
                  </span>
                  <span className="text-[9px] font-mono text-slate-500">
                    {t('styleScout.charCount', { count: sampleRaw.length })}
                  </span>
                </div>
                <textarea
                  rows={6}
                  value={sampleRaw}
                  onChange={(e) => setUserCustomRaw(e.target.value)}
                  placeholder={t('styleScout.sourcePlaceholder')}
                  className="w-full flex-1 bg-transparent text-xs text-slate-200 font-serif leading-relaxed focus:outline-none resize-none pt-1 px-0.5"
                />
              </div>

              {/* Right Column: Live AI Translation Result */}
              <div className="flex flex-col bg-[#0e1422] border border-indigo-500/25 rounded-xl p-4 space-y-2 relative min-h-40">
                <div className="flex items-center justify-between pb-1 border-b border-white/4">
                  <span className="text-[10px] text-indigo-300 uppercase font-mono font-semibold">
                    {t('styleScout.sampleTranslation')}
                  </span>
                  {isGenerating && (
                    <span className="text-[10px] font-mono text-indigo-400 animate-pulse">
                      {t('styleScout.processingGateway')}
                    </span>
                  )}
                </div>

                <div className="flex-1 text-xs text-slate-100 font-serif leading-relaxed whitespace-pre-wrap select-text overflow-y-auto max-h-37.5 pt-1 px-0.5">
                  {sampleTranslated ? (
                    sampleTranslated
                  ) : (
                    <div className="flex flex-col items-center justify-center h-full text-slate-500 py-6 text-center">
                      <Sparkles className="w-6 h-6 text-slate-600 mb-1.5" />
                      <p className="text-xs">{t('styleScout.clickToTest')}</p>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>

          {/* Error Banner with Quick Settings Link */}
          {errorMessage && (
            <div className="p-3 bg-rose-950/40 border border-rose-500/30 rounded-xl text-xs text-rose-300 flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <AlertTriangle className="w-4 h-4 text-rose-400 shrink-0" />
                <span>{errorMessage}</span>
              </div>
              <button
                type="button"
                onClick={() => {
                  setStyleScoutOpen(false);
                  setSettingsOpen(true);
                }}
                className="flex items-center space-x-1 px-2.5 py-1 bg-rose-900/50 hover:bg-rose-900 text-rose-200 border border-rose-500/40 rounded-lg text-xs font-semibold transition shrink-0 ml-2"
              >
                <Settings className="w-3 h-3" />
                <span>{t('styleScout.openLlmSettings')}</span>
              </button>
            </div>
          )}

          {/* R19 Shield Switch */}
          <div className="pt-1">
            <label className="flex items-center space-x-3 p-3 bg-white/2 border border-white/8 rounded-xl cursor-pointer hover:border-white/20 transition">
              <input
                type="checkbox"
                checked={enableR19}
                onChange={e => setEnableR19(e.target.checked)}
                className="w-4 h-4 rounded border-white/20 text-indigo-600 focus:ring-indigo-500"
              />
              <div className="flex items-center space-x-2">
                <ShieldCheck className="w-4 h-4 text-emerald-400 shrink-0" />
                <span className="text-xs font-semibold text-slate-200">{t('styleScout.r19Shield')}</span>
              </div>
            </label>
          </div>
        </div>

        {/* Modal Footer */}
        <div className="flex items-center justify-between px-6 py-3 border-t border-white/8 bg-[#0a0d16] shrink-0">
          <div className="text-[11px] text-slate-400 flex items-center space-x-2">
            <span>{t('styleScout.appliedStyle')}</span>
            <span className="text-indigo-300 font-semibold font-mono">
              {scoutMode === 'ai' && aiScoutedProfile 
                ? `[AI Auto]: ${aiScoutedProfile.styleName}`
                : allManualStyles.find(s => s.id === selectedStyleId)?.name || t('styleScout.default')}
            </span>
          </div>

          <div className="flex items-center space-x-2.5">
            <button
              type="button"
              onClick={() => setStyleScoutOpen(false)}
              className="px-4 py-2 bg-white/4 hover:bg-white/8 text-slate-300 rounded-lg text-xs font-medium border border-white/10 transition"
            >
              {t('common.cancel')}
            </button>
            <button
              type="button"
              onClick={handleConfirmStyle}
              className="flex items-center space-x-1.5 px-5 py-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold shadow-sm border border-indigo-400/30 transition"
            >
              <Check className="w-4 h-4" />
              <span>{t('styleScout.approveStyleBtn')}</span>
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};
