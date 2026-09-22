import React, { useState, useEffect, useRef, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'react-hot-toast';
import { 
  Play, 
  Square,
  Layers,
  PauseCircle, 
  StopCircle, 
  Edit3, 
  Save, 
  Send, 
  Sparkles, 
  Zap, 
  Search, 
  Cpu, 
  DollarSign, 
  ChevronDown,
  ChevronUp,
  PanelRightClose,
  MessageSquare,
  Languages,
  ArrowRight,
  BookOpen,
  Code2,
  FolderPlus,
  RotateCcw,
  Plus,
  BookPlus,
  Check,
  Columns2,
  AlertTriangle,
  Loader2,
  X,
  Share2,
  Trash2,
  Maximize2,
  Minimize2,
  Brain,
  User,
  Timer,
  Users,
  Globe,
  HelpCircle,
  Wrench,
  ShieldCheck,
  BookMarked,
  Wand2,
  Bot,
  CheckCircle2,
} from 'lucide-react';
import * as FSService from '@bindings/novelclaw/internal/services/fsservice';
import { useAppStore } from '@/store/useAppStore';
import { Select } from '@/components/ui/Select';
import { sanitizeReaderHtml } from '@/utils/readerHtml';
import { showConfirmModal } from '@/store/confirmStore';

const slashCommands = [
  { cmd: '/compact', descKey: 'cowork.cmdCompactDesc', syntax: '/compact' },
  { cmd: '/clear', descKey: 'cowork.cmdClearDesc', syntax: '/clear' },
  { cmd: '/term', descKey: 'cowork.cmdTermDesc', syntax: '/term <source> = <target> [category]' },
  { cmd: '/char', descKey: 'cowork.cmdCharDesc', syntax: '/char <A> -> <B>: <address>' },
  { cmd: '/steer', descKey: 'cowork.cmdSteerDesc', syntax: '/steer <instruction>' },
  { cmd: '/soul', descKey: 'cowork.cmdSoulDesc', syntax: '/soul <soul_id>' },
  { cmd: '/rollback', descKey: 'cowork.cmdRollbackDesc', syntax: '/rollback [chapter]' },
  { cmd: '/pause', descKey: 'cowork.cmdPauseDesc', syntax: '/pause' },
  { cmd: '/resume', descKey: 'cowork.cmdResumeDesc', syntax: '/resume' },
  { cmd: '/abort', descKey: 'cowork.cmdAbortDesc', syntax: '/abort' },
  { cmd: '/learn', descKey: 'cowork.cmdLearnDesc', syntax: '/learn [notes]' },
  { cmd: '/help', descKey: 'cowork.cmdHelpDesc', syntax: '/help' },
];

const ToolCallStepCard: React.FC<{
  msg: any;
  isExpanded: boolean;
  onToggle: () => void;
}> = ({ msg, isExpanded, onToggle }) => {
  const { t } = useTranslation();
  let toolName = '';
  let input: any = null;
  let output: any = null;

  if (msg.action_call_json) {
    try {
      const parsed = JSON.parse(msg.action_call_json);
      toolName = parsed.tool_name || '';
      input = parsed.input;
      output = parsed.output;
    } catch {
      // fallback
    }
  }

  if (!toolName) {
    // Matches both the legacy Vietnamese label and the current English one.
    toolName = msg.content?.replace(/^(Thực thi công cụ|Executing tool):\s*/i, '') || 'Tool';
  }

  let title = t('tools.toolExecution', { toolName });
  let icon = <Wrench className="w-3.5 h-3.5 text-indigo-400 shrink-0" />;

  if (toolName === 'lookup_character_relation' || toolName.includes('character')) {
    title = t('tools.charRelation');
    icon = <Users className="w-3.5 h-3.5 text-cyan-400 shrink-0" />;
  } else if (toolName === 'lookup_world_lore' || toolName.includes('world') || toolName.includes('lore')) {
    title = t('tools.worldLore');
    icon = <BookOpen className="w-3.5 h-3.5 text-amber-400 shrink-0" />;
  } else if (toolName === 'search_book_context' || toolName.includes('search')) {
    title = t('tools.searchContext');
    icon = <Search className="w-3.5 h-3.5 text-indigo-400 shrink-0" />;
  } else if (toolName === 'web_lookup' || toolName.includes('web')) {
    title = t('tools.webLookup');
    icon = <Globe className="w-3.5 h-3.5 text-blue-400 shrink-0" />;
  } else if (toolName === 'ask_human_coworker' || toolName.includes('human')) {
    title = t('tools.askHuman');
    icon = <HelpCircle className="w-3.5 h-3.5 text-rose-400 shrink-0" />;
  } else if (toolName === 'shadow_critic_hotpatch' || toolName.includes('critic') || toolName === 'shadow_critic_review') {
    title = t('tools.shadowCritic');
    icon = <ShieldCheck className="w-3.5 h-3.5 text-emerald-400 shrink-0" />;
  } else if (toolName === 'match_glossary_terms' || toolName.includes('glossary')) {
    title = t('tools.glossaryContext');
    icon = <BookMarked className="w-3.5 h-3.5 text-amber-400 shrink-0" />;
  }

  return (
    <div className="w-full max-w-[94%] bg-[#080c14] border border-white/8 hover:border-indigo-500/30 rounded-xl overflow-hidden shadow-sm transition animate-in fade-in duration-150">
      <div 
        onClick={onToggle}
        className="w-full px-3 py-2 flex items-center justify-between cursor-pointer bg-white/2 hover:bg-white/5 transition select-none gap-2"
        title={toolName}
      >
        <div className="flex items-center space-x-2 min-w-0 flex-1">
          {icon}
          <span className="font-semibold text-xs text-slate-200 truncate">{title}</span>
          <span className="text-[10px] font-mono text-slate-500 hidden 2xl:inline truncate max-w-30">({toolName})</span>
        </div>
        <div className="flex items-center space-x-1.5 shrink-0">
          <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 whitespace-nowrap">
            <Check className="w-2.5 h-2.5 shrink-0" /> {t('batchImport.done')}
          </span>
          {isExpanded ? <ChevronUp className="w-3 h-3 text-slate-400 shrink-0" /> : <ChevronDown className="w-3 h-3 text-slate-400 shrink-0" />}
        </div>
      </div>
      {isExpanded && (
        <div className="p-3 bg-[#060810] border-t border-white/6 space-y-2 text-[11px] font-mono">
          <div className="text-[10px] text-slate-400 font-mono flex items-center gap-1.5 pb-1 border-b border-white/4">
            <Wrench className="w-3 h-3 text-indigo-400 shrink-0" />
            <span>Tool: <strong className="text-slate-200">{toolName}</strong></span>
          </div>
          {input && (
            <div>
              <div className="text-[10px] font-semibold text-slate-500 uppercase tracking-wider mb-1">{t('cowork.inputParams')}:</div>
              <div className="p-2 rounded bg-black/40 text-slate-300 whitespace-pre-wrap break-all border border-white/4">
                {typeof input === 'object' ? JSON.stringify(input, null, 2) : String(input)}
              </div>
            </div>
          )}
          {output && (
            <div>
              <div className="text-[10px] font-semibold text-slate-500 uppercase tracking-wider mb-1">{t('cowork.executionResult')}:</div>
              <div className="p-2 rounded bg-black/40 text-emerald-300/90 whitespace-pre-wrap break-all border border-white/4 max-h-48 overflow-y-auto">
                {typeof output === 'object' ? JSON.stringify(output, null, 2) : String(output)}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

const TranslationSummaryCard: React.FC<{ msg: any }> = ({ msg }) => {
  const { t } = useTranslation();
  let stats: any = null;
  if (msg.action_call_json) {
    try {
      stats = JSON.parse(msg.action_call_json);
    } catch {
      // fallback
    }
  }

  const lines = msg.content ? msg.content.split('\n') : [];
  const rawHeader = lines[0] || t('cowork.summaryHeader');
  const headerText = rawHeader.replace(/^\*\*|\*\*$/g, '').replace(/^#+\s*/, '').trim();
  const detailText = lines.slice(1).join('\n').trim();

  return (
    <div className="w-full max-w-[94%] p-3.5 rounded-xl bg-linear-to-b from-indigo-950/30 to-[#090d18] border border-indigo-500/30 shadow-lg text-xs space-y-3 animate-in fade-in duration-200">
      <div className="flex items-center justify-between gap-2">
        <div className="flex items-center space-x-2 min-w-0 flex-1" title={headerText}>
          <Sparkles className="w-4 h-4 text-amber-400 shrink-0" />
          <span className="font-bold text-slate-100 text-sm truncate">{headerText}</span>
        </div>
        <span className="inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 shrink-0 whitespace-nowrap">
          ✓ {t('batchImport.done')}
        </span>
      </div>

      {stats && (
        <div className="grid grid-cols-2 gap-2 pt-1">
          <div className="p-2.5 rounded-lg bg-white/3 border border-white/6 text-center min-w-0">
            <div className="text-[10px] text-slate-400 truncate">{t('cowork.translatedRunes')}</div>
            <div className="text-xs font-bold text-indigo-300 font-mono mt-0.5 truncate">{stats.total_runes?.toLocaleString() || 0} runes</div>
          </div>
          {(() => {
            const tokPerSec = stats.tokens_per_second || (stats.duration_ms ? (stats.total_tokens / (stats.duration_ms / 1000)) : 0);
            const runesPerSec = stats.runes_per_second || stats.runes_per_sec || 0;
            const isInstant = (stats.duration_ms && stats.duration_ms < 200) || runesPerSec > 5000 || tokPerSec > 2000;
            return (
              <div className="p-2.5 rounded-lg bg-white/3 border border-white/6 text-center min-w-0">
                <div className="text-[10px] text-slate-400 truncate">{t('cowork.translationSpeed')}</div>
                <div className="text-xs font-bold text-cyan-300 font-mono mt-0.5 truncate">
                  {isInstant ? 'Cached / Fast' : tokPerSec > 0 ? `${tokPerSec.toFixed(1)} tok/s` : runesPerSec > 0 ? `${runesPerSec.toFixed(1)} r/s` : '—'}
                </div>
                {!isInstant && tokPerSec > 0 && runesPerSec > 0 && (
                  <div className="text-[9.5px] text-slate-500 font-mono mt-0.5 truncate">
                    ({runesPerSec.toFixed(1)} r/s)
                  </div>
                )}
              </div>
            );
          })()}
          <div className="p-2.5 rounded-lg bg-white/3 border border-white/6 text-center min-w-0">
            <div className="text-[10px] text-slate-400 truncate">{t('cowork.shadowCritic')}</div>
            <div className="text-xs font-bold text-amber-300 font-mono mt-0.5 truncate">{stats.revised_count || 0} {t('cowork.revisions')}</div>
          </div>
          <div className="p-2.5 rounded-lg bg-white/3 border border-white/6 text-center min-w-0">
            <div className="text-[10px] text-slate-400 truncate">{t('cowork.tokens')}</div>
            <div className="text-xs font-bold text-emerald-300 font-mono mt-0.5 truncate">{stats.total_tokens?.toLocaleString() || 0}</div>
          </div>
        </div>
      )}

      {detailText && (
        <div className="text-[11px] text-slate-300 leading-relaxed font-sans pt-1 border-t border-white/6 whitespace-pre-wrap">
          {detailText}
        </div>
      )}
    </div>
  );
};

export const CoworkerWorkspace: React.FC = () => {
  const { t } = useTranslation();

  const {
    selectedProject,
    chapters,
    selectedChapter,
    selectChapter,
    isTranslating,
    telemetry,
    thoughtText,
    activeTool,
    translationMode,
    setTranslationMode,
    startTranslation,
    triggerSoftStop,
    triggerHardAbort,
    updateChapterTranslation,
    updateProjectLanguages,
    rollbackChapter,
    appendBookToProject,
    pendingAskHuman,
    setPendingAskHuman,
    soul,
    availableSouls,
    setActiveSoul,
    entities,
    relations,
    autoScanEntities,
    setActiveTab,
    novelclawThreads,
    selectedThread,
    selectNovelClawThread,
    createSubThread,
    deleteNovelClawThread,
    novelclawMessages,
    isNovelClawThinking,
    novelclawThinkingText,
    novelclawActiveTool,
    sendNovelClawMessage,
    clearNovelClawThread,
    compactNovelClawThread,
    autoCompactThreshold,
    selectedVolume,
    setSelectedVolume,
    decisionMode,
    setDecisionMode,
    resolveAskHuman,
    styleGuide,
    setStyleGuide,
    autoScoutStyle,
    setStyleScoutOpen,
    batchProgress,
    startBatchTranslation,
    stopBatchTranslation,
    thinkingEffort,
    setThinkingEffort,
  } = useAppStore();

  const [isRollingBack, setIsRollingBack] = useState(false);
  const [isInlineEditing, setIsInlineEditing] = useState(false);
  const [editableContent, setEditableContent] = useState(selectedChapter?.translated_content || '');
  const [prevChapterId, setPrevChapterId] = useState(selectedChapter?.id);
  const [rawViewMode, setRawViewMode] = useState<'rendered' | 'code'>('rendered');
  const [translatedViewMode, setTranslatedViewMode] = useState<'rendered' | 'code'>('rendered');
  const [chatInput, setChatInput] = useState('');
  const [isSoulPanelOpen, setIsSoulPanelOpen] = useState(true);
  const [paneLayout, setPaneLayout] = useState<'split' | 'target' | 'source'>('split');
  const chatBottomRef = useRef<HTMLDivElement>(null);
  const translatedContentRef = useRef<HTMLDivElement>(null);

  // Batch Translation Scope & Modal States
  const [translateScope, setTranslateScope] = useState<'chapter' | 'volume' | 'series'>('chapter');
  const [isBatchModalOpen, setIsBatchModalOpen] = useState(false);
  const [batchTargetVolume, setBatchTargetVolume] = useState<string>('');
  const [batchSkipCompleted, setBatchSkipCompleted] = useState(true);

  // NovelClaw Console State
  const [isCreatingSubThread, setIsCreatingSubThread] = useState(false);
  const [subThreadTitle, setSubThreadTitle] = useState('');
  const [subThreadVol, setSubThreadVol] = useState(0);
  const [subThreadChap, setSubThreadChap] = useState(0);
  const [expandAllThinking, setExpandAllThinking] = useState(false);
  const [expandedThinkingIds, setExpandedThinkingIds] = useState<Record<string, boolean>>({});
  const [liveThinkingExpanded, setLiveThinkingExpanded] = useState(false);
  const [expandedToolIds, setExpandedToolIds] = useState<Record<string, boolean>>({});
  // HITL countdown timer (Stage 7)
  const [hitlCountdown, setHitlCountdown] = useState<number | null>(null);
  const [hitlCustomInput, setHitlCustomInput] = useState('');

  useEffect(() => {
    if (!pendingAskHuman || !pendingAskHuman.createdAt || !pendingAskHuman.timeoutSeconds) {
      setHitlCountdown(null);
      return;
    }
    const calcRemaining = () => {
      const elapsed = Math.floor((Date.now() - pendingAskHuman.createdAt!) / 1000);
      return Math.max(0, (pendingAskHuman.timeoutSeconds || 60) - elapsed);
    };
    setHitlCountdown(calcRemaining());
    const interval = setInterval(() => {
      const remaining = calcRemaining();
      setHitlCountdown(remaining);
      if (remaining <= 0) {
        clearInterval(interval);
      }
    }, 1000);
    return () => clearInterval(interval);
  }, [pendingAskHuman]);

  const filteredSlashCommands = useMemo(() => {
    if (!chatInput.startsWith('/')) return [];
    const query = chatInput.slice(1).toLowerCase();
    return slashCommands.filter(c => c.cmd.slice(1).startsWith(query));
  }, [chatInput]);

  useEffect(() => {
    const handleKeyDown = (e: KeyboardEvent) => {
      if ((e.ctrlKey || e.metaKey) && e.key.toLowerCase() === 'o') {
        e.preventDefault();
        setExpandAllThinking(prev => !prev);
      }
    };
    window.addEventListener('keydown', handleKeyDown);
    return () => window.removeEventListener('keydown', handleKeyDown);
  }, []);

  useEffect(() => {
    chatBottomRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [novelclawMessages, isNovelClawThinking, novelclawThinkingText]);

  // Volume segmentation and management
  const volumes = useMemo(() => {
    const set = new Set<string>();
    chapters.forEach(ch => {
      const m = ch.title.match(/^\[(.*?)\]/);
      if (m && m[1]) set.add(m[1]);
    });
    return Array.from(set);
  }, [chapters]);

  const filteredChapters = useMemo(() => {
    if (selectedVolume === 'all') return chapters;
    return chapters.filter(c => c.title.startsWith(`[${selectedVolume}]`));
  }, [chapters, selectedVolume]);

  const [isAddingVolume, setIsAddingVolume] = useState(false);
  const [newVolPath, setNewVolPath] = useState('');
  const [newVolLabel, setNewVolLabel] = useState('');
  const [isAppendingVol, setIsAppendingVol] = useState(false);

  const handlePickAndAddVolume = async () => {
    try {
      const path = await FSService.PickBookFile();
      if (path) {
        setNewVolPath(path);
        const clean = path.replace(/\\/g, '/');
        const filename = clean.substring(clean.lastIndexOf('/') + 1);
        const base = filename.substring(0, filename.lastIndexOf('.')) || filename;
        setNewVolLabel(base);
        setIsAddingVolume(true);
      }
    } catch (err) {
      console.error('Pick book failed:', err);
    }
  };

  const handleConfirmAddVolume = async () => {
    if (!newVolPath || !selectedProject) return;
    setIsAppendingVol(true);
    try {
      await appendBookToProject(selectedProject.id, newVolPath, newVolLabel.trim() || undefined);
      setIsAddingVolume(false);
      setNewVolPath('');
      setNewVolLabel('');
      toast.success(t('cowork.addVolumeSuccess'));
    } catch (err: any) {
      toast.error(t('cowork.addVolumeError') + (err?.message || err));
    } finally {
      setIsAppendingVol(false);
    }
  };

  // Current chapter volume info & bounds
  const currentChapterVolumeInfo = useMemo(() => {
    if (!selectedChapter) {
      return { volTag: '', volStart: 1, volEnd: 1, chapterCount: 0 };
    }
    const m = selectedChapter.title.match(/^\[(.*?)\]/);
    const volTag = m ? m[1] : '';
    if (!volTag) {
      return { volTag: '', volStart: 1, volEnd: chapters.length || 1, chapterCount: chapters.length };
    }

    const volChapters = chapters.filter(c => c.title.startsWith(`[${volTag}]`));
    let start = Number.MAX_SAFE_INTEGER;
    let end = 0;
    volChapters.forEach(c => {
      if (c.chapter_index < start) start = c.chapter_index;
      if (c.chapter_index > end) end = c.chapter_index;
    });

    return {
      volTag,
      volStart: start === Number.MAX_SAFE_INTEGER ? selectedChapter.chapter_index : start,
      volEnd: end === 0 ? selectedChapter.chapter_index : end,
      chapterCount: volChapters.length,
    };
  }, [selectedChapter, chapters]);

  // Graph readiness evaluation for current chapter / volume
  const chapterGraphReadiness = useMemo(() => {
    const { volTag, volStart, volEnd } = currentChapterVolumeInfo;
    const totalRelations = relations.length;
    const totalEntities = entities.length;

    if (volTag) {
      const volumeRelations = relations.filter(r => r.since_chapter >= volStart && r.since_chapter <= volEnd);
      const isGraphReady = volumeRelations.length > 0;
      const isStyleReady = Boolean(styleGuide && styleGuide.trim().length > 0);
      const isFullyReady = isGraphReady && isStyleReady;
      return {
        isReady: isFullyReady,
        isGraphReady,
        isStyleReady,
        hasVolumeTag: true,
        volTag,
        volStart,
        volEnd,
        volumeRelationsCount: volumeRelations.length,
        totalRelations,
        totalEntities,
        badgeText: isFullyReady
          ? t('cowork.volReadyBadge', { volTag })
          : !isGraphReady && !isStyleReady
          ? t('cowork.volNoGraphStyleBadge', { volTag })
          : !isGraphReady
          ? t('cowork.volNoGraphBadge', { volTag })
          : t('cowork.noStyleBadge'),
        message: isFullyReady
          ? t('cowork.volReadyMsg', { volTag })
          : t('cowork.volNotReadyMsg', { volTag }),
      };
    }

    // Untagged chapter
    const isGraphReady = totalRelations > 0;
    const isStyleReady = Boolean(styleGuide && styleGuide.trim().length > 0);
    const isFullyReady = isGraphReady && isStyleReady;
    return {
      isReady: isFullyReady,
      isGraphReady,
      isStyleReady,
      hasVolumeTag: false,
      volTag: '',
      volStart: 1,
      volEnd: chapters.length || 1,
      volumeRelationsCount: totalRelations,
      totalRelations,
      totalEntities,
      badgeText: isFullyReady
        ? t('cowork.readyL2StyleBadge')
        : !isGraphReady && !isStyleReady
        ? t('cowork.unscannedL2StyleBadge')
        : !isGraphReady
        ? t('cowork.noL2Badge')
        : t('cowork.noStyleBadge'),
      message: isFullyReady
        ? t('cowork.l2StyleReadyMsg')
        : t('cowork.l2StyleNotReadyMsg'),
    };
  }, [currentChapterVolumeInfo, relations, entities, chapters.length, styleGuide, t]);

  // Pre-translate interceptor states
  const [isPreTranslateModalOpen, setIsPreTranslateModalOpen] = useState(false);
  const [isPreScanning, setIsPreScanning] = useState(false);
  const [dismissedVolumes, setDismissedVolumes] = useState<Set<string>>(new Set());
  const [dontAskAgainForThisVol, setDontAskAgainForThisVol] = useState(false);

  // Handle Translate Button Click (with Readiness Interceptor)
  const handleTranslateClick = async () => {
    if (isTranslating || isPreScanning) return;
    if (!selectedChapter) {
      toast.error(t('cowork.selectChapterFirst'));
      return;
    }

    const { volTag } = currentChapterVolumeInfo;
    const isDismissed = volTag ? dismissedVolumes.has(volTag) : dismissedVolumes.has('__unscoped__');

    // Intercept if graph or style is not ready and user has not dismissed for this volume
    if ((!chapterGraphReadiness.isGraphReady || !chapterGraphReadiness.isStyleReady) && !isDismissed) {
      setDontAskAgainForThisVol(false);
      setIsPreTranslateModalOpen(true);
      return;
    }

    await startTranslation();
  };

  const activeVolTag = useMemo(() => {
    if (selectedVolume !== 'all') return selectedVolume;
    if (currentChapterVolumeInfo.volTag) return currentChapterVolumeInfo.volTag;
    if (volumes.length > 0) return volumes[0];
    return '';
  }, [selectedVolume, currentChapterVolumeInfo.volTag, volumes]);

  const effectiveBatchVolume = batchTargetVolume || activeVolTag;

  const targetBatchChapters = useMemo(() => {
    let list = [...chapters].sort((a, b) => a.chapter_index - b.chapter_index);
    if (translateScope === 'volume') {
      const vol = effectiveBatchVolume;
      if (vol) {
        list = list.filter(c => c.title.startsWith(`[${vol}]`));
      }
    }
    return list;
  }, [chapters, translateScope, effectiveBatchVolume]);

  const batchStats = useMemo(() => {
    const total = targetBatchChapters.length;
    const completed = targetBatchChapters.filter(
      c => c.status === 'completed' && Boolean(c.translated_content && c.translated_content.trim().length > 0)
    ).length;
    const pending = total - completed;
    return { total, completed, pending };
  }, [targetBatchChapters]);

  const handleActionClick = () => {
    if (isTranslating || isPreScanning) return;
    if (translateScope === 'chapter') {
      handleTranslateClick();
    } else {
      if (translateScope === 'volume' && !batchTargetVolume) {
        setBatchTargetVolume(activeVolTag);
      }
      setIsBatchModalOpen(true);
    }
  };

  const handleConfirmStartBatch = async () => {
    setIsBatchModalOpen(false);
    try {
      const vol = translateScope === 'volume' ? effectiveBatchVolume : undefined;
      await startBatchTranslation({
        scope: translateScope as 'volume' | 'series',
        targetVolume: vol,
        skipCompleted: batchSkipCompleted,
      });
      toast.success(t('cowork.batchCompletedToast'));
    } catch (err: any) {
      console.error('Batch translation error:', err);
      toast.error(t('cowork.batchTranslationError') + (err?.message || err));
    }
  };

  const handleStopBatch = async () => {
    if (batchProgress?.isStopping) return;
    try {
      await stopBatchTranslation();
      toast.success(t('cowork.batchStoppedToast'));
    } catch (err: any) {
      toast.error(t('cowork.stopError') + (err?.message || err));
    }
  };

  // 1-Click Scan Volume & Style, then Seamlessly Auto-Translate
  const handleScanAndTranslate = async () => {
    if (isPreScanning || isTranslating || !selectedChapter || !selectedProject) return;

    setIsPreScanning(true);
    const { volTag } = currentChapterVolumeInfo;
    const { isGraphReady, isStyleReady } = chapterGraphReadiness;

    try {
      // 1. Auto Scout Style if missing
      if (!isStyleReady) {
        toast.loading(t('cowork.skimming3ChaptersToast'), { id: 'prescan-toast' });
        const scoutRes = await autoScoutStyle({
          project_id: selectedProject.id,
          source_lang: selectedProject.source_lang,
          target_lang: selectedProject.target_lang,
        });
        if (scoutRes && scoutRes.success) {
          setStyleGuide(scoutRes.style_prompt);
          toast.success(t('cowork.styleFormedToast', { styleName: scoutRes.style_name }), { id: 'prescan-toast' });
        }
      }

      // 2. Auto Scan Character Graph if missing
      if (!isGraphReady) {
        toast.loading(t('cowork.scanningRelationsToast', { volTag: volTag ? `[${volTag}]` : '' }), { id: 'prescan-toast' });
        const res = await autoScanEntities(
          selectedChapter.chapter_index,
          volTag ? 'volume' : 'all',
          volTag
        );
        if (res && res.success) {
          toast.success(
            t('cowork.extractedEntitiesToast', { entities: res.entities_found, relations: res.relations_found }),
            { id: 'prescan-toast' }
          );
        }
      }

      setIsPreTranslateModalOpen(false);
      toast.success(t('cowork.dataReadyStartToast'), { id: 'prescan-toast' });
      // Start translation immediately without extra clicks!
      await startTranslation();
    } catch (err: any) {
      console.error('Pre-scan error:', err);
      toast.error(t('cowork.prescanErrorToast', { err: err?.message || err }), { id: 'prescan-toast' });
    } finally {
      setIsPreScanning(false);
    }
  };

  // Proceed with old/empty graph
  const handleProceedWithoutScan = async () => {
    const { volTag } = currentChapterVolumeInfo;
    if (dontAskAgainForThisVol) {
      setDismissedVolumes(prev => new Set(prev).add(volTag || '__unscoped__'));
    }
    setIsPreTranslateModalOpen(false);
    await startTranslation();
  };

  if (selectedChapter && selectedChapter.id !== prevChapterId) {
    setPrevChapterId(selectedChapter.id);
    setEditableContent(selectedChapter.translated_content || '');
  }

  // Auto-scroll translated content pane to bottom as new chunks stream in
  useEffect(() => {
    if (isTranslating && translatedContentRef.current) {
      translatedContentRef.current.scrollTop = translatedContentRef.current.scrollHeight;
    }
  }, [selectedChapter?.translated_content, isTranslating]);

  // Turn off inline editing automatically whenever translation starts
  useEffect(() => {
    if (isTranslating) {
      setIsInlineEditing(false);
    }
  }, [isTranslating]);

  // Sync editableContent when selectedChapter's content updates (if not actively editing)
  useEffect(() => {
    if (!isInlineEditing && selectedChapter) {
      setEditableContent(selectedChapter.translated_content || '');
    }
  }, [selectedChapter, isInlineEditing]);

  const handleSaveInlineEdit = async () => {
    if (!selectedChapter) return;
    try {
      await updateChapterTranslation(selectedChapter.id, editableContent);
      setIsInlineEditing(false);
      toast.success(t('cowork.saveTranslationSuccess'));
    } catch (err: any) {
      toast.error(t('cowork.saveTranslationError') + (err?.message || err));
    }
  };

  const handleRollback = async () => {
    if (!selectedChapter) return;
    const confirmed = await showConfirmModal({
      title: t('cowork.rollbackTitle'),
      message: t('cowork.rollbackConfirm', { index: selectedChapter.chapter_index }),
      confirmText: t('cowork.rollbackCheckpoint'),
      cancelText: t('common.cancel'),
      variant: 'warning',
    });
    if (!confirmed) return;
    setIsRollingBack(true);
    try {
      await rollbackChapter();
      toast.success(t('cowork.rollbackSuccess', { index: selectedChapter.chapter_index }));
    } catch (err: any) {
      toast.error(t('cowork.rollbackError', { err: err?.message || err }));
    } finally {
      setIsRollingBack(false);
    }
  };

  const handleSendChat = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!chatInput.trim()) return;
    const text = chatInput.trim();
    setChatInput('');
    await sendNovelClawMessage(text);
  };

  const handleConfirmCreateSubThread = async () => {
    if (!selectedProject) return;
    try {
      await createSubThread({
        project_id: selectedProject.id,
        title: subThreadTitle.trim() || t('cowork.newSubchatDefault'),
        volume_index: Number(subThreadVol) || 0,
        chapter_index: Number(subThreadChap) || 0,
      });
      setIsCreatingSubThread(false);
      setSubThreadTitle('');
      setSubThreadVol(0);
      setSubThreadChap(0);
      toast.success(t('cowork.createSubchatSuccess'));
    } catch (err: any) {
      toast.error(t('cowork.createSubchatError') + (err?.message || err));
    }
  };

  const activeTokens = selectedThread?.token_count || 0;
  const maxTokens = autoCompactThreshold || 250000;
  const tokenPercent = Math.min(100, Math.round((activeTokens / maxTokens) * 100));

  if (!selectedProject) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center p-8 text-center bg-[#0a0d14]">
        <div className="w-16 h-16 rounded-2xl bg-indigo-500/10 border border-indigo-500/20 flex items-center justify-center text-indigo-400 mb-4">
          <BookOpen className="w-8 h-8" />
        </div>
        <h2 className="text-base font-bold text-slate-100">{t('cowork.noProjectTitle', 'Select a project to start translating')}</h2>
        <p className="text-xs text-slate-400 mt-1.5 max-w-md leading-relaxed">
          {t('cowork.noProjectDesc', 'Import a new book (EPUB, TXT, DOCX, MOBI, etc.) or choose a project from the list to start the translation workflow with NovelClaw.')}
        </p>
        <button
          onClick={() => useAppStore.getState().setImportOpen(true)}
          className="mt-5 flex items-center space-x-2 px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-sm transition active:scale-[0.98]"
        >
          <FolderPlus className="w-4 h-4" />
          <span>{t('cowork.importNow', 'Import Book Now')}</span>
        </button>
      </div>
    );
  }

  return (
    <div className="flex-1 flex h-full min-h-0 overflow-hidden bg-[#0a0d14] relative">
      {/* LANE 1: Real-time Translation & Thoughts */}
      <div className="flex-1 flex flex-col min-w-0 bg-[#0c1018] overflow-hidden">
        {/* Tier 1: Project Navigation & Breadcrumb Context (Rigid Single Row, Zero Overlap) */}
        <div className="h-10 px-2.5 sm:px-3 md:px-4 border-b border-white/8 flex items-center justify-between bg-[#0e131f] shrink-0 gap-2 sm:gap-3.5 flex-nowrap overflow-hidden">
          {/* Left Cluster: Full Breadcrumb Navigation (Never overflows or paints on right cluster) */}
          <div className="flex items-center gap-1.5 sm:gap-2 min-w-0 flex-1 overflow-hidden">
            <span className="hidden xl:inline-flex h-7 px-2 items-center rounded-md bg-white/4 border border-white/8 text-[10px] font-mono font-medium text-slate-400 uppercase whitespace-nowrap shrink-0">
              {t('cowork.lane1Badge')}
            </span>

            {/* Volume Selector (if multi-volume project exists) */}
            {volumes.length > 0 && (
              <Select
                value={selectedVolume}
                onChange={(e) => {
                  const newVol = e.target.value;
                  setSelectedVolume(newVol);
                  if (newVol !== 'all') {
                     const firstCh = chapters.find(c => c.title.startsWith(`[${newVol}]`));
                    if (firstCh) selectChapter(firstCh);
                  }
                }}
                className="w-20 sm:w-24 md:w-28 h-7 truncate text-xs font-semibold text-indigo-300 border-indigo-500/30"
                containerClassName="shrink-0 max-w-[100px] sm:max-w-none"
                title={t('cowork.filterVolumeTooltip')}
              >
                <option value="all">{t('cowork.allVolumes', { count: volumes.length })}</option>
                {volumes.map(vol => (
                  <option key={vol} value={vol}>
                    {vol}
                  </option>
                ))}
              </Select>
            )}

            {/* Quick Add Volume Button (Next to Volume Selector) */}
            <button
              onClick={handlePickAndAddVolume}
              className="h-7 w-7 flex items-center justify-center rounded-md text-xs font-medium text-slate-300 hover:text-white bg-[#111726] hover:bg-white/8 border border-white/10 transition shrink-0"
              title={t('cowork.addVolumeTitle')}
            >
              <Plus className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
            </button>

            {/* Chapter Selector (Expansive, Flex-1, Min-w-0 To Never Overlap Sibling) */}
            <Select
              value={selectedChapter?.id || ''}
              onChange={(e) => {
                const ch = chapters.find(c => c.id === e.target.value);
                if (ch) selectChapter(ch);
              }}
              className="w-full h-7 text-xs font-medium text-slate-100"
              containerClassName="flex-1 min-w-0 max-w-[560px]"
              title={selectedChapter ? selectedChapter.title : t('cowork.selectChapterToTranslate')}
            >
              {selectedVolume !== 'all' ? (
                filteredChapters.map(ch => {
                  const cleanTitle = ch.title.replace(/^\[(.*?)\]\s*/, '');
                  const statusSuffix = ch.status ? ` (${ch.status})` : '';
                  return (
                    <option key={ch.id} value={ch.id}>
                      {cleanTitle}{statusSuffix}
                    </option>
                  );
                })
              ) : volumes.length > 0 ? (
                volumes.map(vol => (
                  <optgroup key={vol} label={t('cowork.volumeOptgroup', { vol })}>
                    {chapters
                      .filter(ch => ch.title.startsWith(`[${vol}]`))
                      .map(ch => {
                        const cleanTitle = ch.title.replace(/^\[(.*?)\]\s*/, '');
                        const statusSuffix = ch.status ? ` (${ch.status})` : '';
                        return (
                          <option key={ch.id} value={ch.id}>
                            {cleanTitle}{statusSuffix}
                          </option>
                        );
                      })}
                  </optgroup>
                ))
              ) : (
                chapters.map(ch => {
                  const statusSuffix = ch.status ? ` (${ch.status})` : '';
                  return (
                    <option key={ch.id} value={ch.id}>
                      {ch.title}{statusSuffix}
                    </option>
                  );
                })
              )}
            </Select>
          </div>

          {/* Right Cluster: View Mode & Language Flow */}
          <div className="flex items-center gap-1.5 sm:gap-2 shrink-0">
            {/* Pane View Switcher (Adaptive: full labels on xl+, compact badges on mobile/tablet) */}
            <div className="h-7 flex items-center p-0.5 bg-[#111726] border border-white/10 rounded-md shrink-0 text-xs">
              <button
                type="button"
                onClick={() => setPaneLayout('split')}
                className={`h-full px-1.5 sm:px-2 rounded flex items-center gap-1 text-[11px] font-medium transition whitespace-nowrap ${
                  paneLayout === 'split'
                    ? 'bg-white/10 text-white font-semibold shadow-xs'
                    : 'text-slate-400 hover:text-slate-200'
                }`}
                title={t('cowork.splitView')}
              >
                <Columns2 className="w-3.5 h-3.5 shrink-0" />
                <span className="hidden xl:inline">{t('cowork.splitView')}</span>
              </button>
              <button
                type="button"
                onClick={() => setPaneLayout('target')}
                className={`h-full px-1.5 sm:px-2 rounded flex items-center gap-1 text-[11px] font-medium transition whitespace-nowrap ${
                  paneLayout === 'target'
                    ? 'bg-white/10 text-white font-semibold shadow-xs'
                    : 'text-slate-400 hover:text-slate-200'
                }`}
                title={t('cowork.targetOnly')}
              >
                <span className="hidden xl:inline">{t('cowork.targetOnly')}</span>
                <span className="xl:hidden text-[10px] font-bold text-indigo-300">VI</span>
              </button>
              <button
                type="button"
                onClick={() => setPaneLayout('source')}
                className={`h-full px-1.5 sm:px-2 rounded flex items-center gap-1 text-[11px] font-medium transition whitespace-nowrap ${
                  paneLayout === 'source'
                    ? 'bg-white/10 text-white font-semibold shadow-xs'
                    : 'text-slate-400 hover:text-slate-200'
                }`}
                title={t('cowork.sourceOnly')}
              >
                <span className="hidden xl:inline">{t('cowork.sourceOnly')}</span>
                <span className="xl:hidden text-[10px] font-bold text-slate-300">JA</span>
              </button>
            </div>

            {/* Language Direction Setup (Single unified pill, NO box-in-box) */}
            <div 
              className="h-7 flex items-center gap-1 px-1.5 sm:px-2 bg-[#111726] border border-white/10 hover:border-white/20 rounded-md text-xs shrink-0 transition"
              title={t('cowork.langDirectionTitle')}
            >
              <Languages className="w-3.5 h-3.5 text-indigo-400 shrink-0 hidden sm:inline" />
              <select
                value={selectedProject.source_lang || 'ja'}
                onChange={(e) => updateProjectLanguages(e.target.value, selectedProject.target_lang || 'en')}
                className="bg-transparent border-0 text-xs font-mono font-bold text-slate-200 cursor-pointer focus:outline-none p-0 pr-0.5 [&>option]:bg-[#111726] [&>option]:text-slate-200"
                title={t('cowork.sourceLangTooltip')}
              >
                <option value="ja">JA</option>
                <option value="en">EN</option>
                <option value="zh">ZH</option>
                <option value="ko">KO</option>
                <option value="vi">VI</option>
              </select>
              <ArrowRight className="w-3 h-3 text-slate-500 shrink-0" />
              <select
                value={selectedProject.target_lang || 'en'}
                onChange={(e) => updateProjectLanguages(selectedProject.source_lang || 'ja', e.target.value)}
                className="bg-transparent border-0 text-xs font-mono font-bold text-indigo-300 cursor-pointer focus:outline-none p-0 pr-0.5 [&>option]:bg-[#111726] [&>option]:text-slate-200"
                title={t('cowork.targetLangTooltip')}
              >
                <option value="en">EN</option>
                <option value="vi">VI</option>
                <option value="zh">ZH</option>
                <option value="ja">JA</option>
                <option value="ko">KO</option>
              </select>
            </div>
          </div>
        </div>

        {/* Tier 2: Execution CTA, Engine Mode & Telemetry (Rigid Single Row) */}
        <div className="h-9 px-2.5 sm:px-3 md:px-4 bg-[#0a0e19] border-b border-white/6 flex items-center justify-between text-xs shrink-0 gap-2 overflow-hidden flex-nowrap">
          {/* Left Cluster: Execution CTAs & Translation Engine */}
          <div className="flex items-center gap-1 sm:gap-1.5 shrink-0">
            {batchProgress ? (
              <>
                {/* Stop Batch Translation Button */}
                <button
                  type="button"
                  onClick={handleStopBatch}
                  disabled={batchProgress.isStopping}
                  className="h-7 flex items-center space-x-1.5 px-2.5 sm:px-3 rounded-md text-xs font-semibold whitespace-nowrap shrink-0 transition bg-rose-600 hover:bg-rose-500 text-white shadow-xs border border-rose-400/30 active:scale-[0.98] disabled:opacity-60 disabled:cursor-not-allowed"
                  title={t('cowork.batchStopBtn')}
                >
                  {batchProgress.isStopping ? (
                    <Loader2 className="w-3 h-3 animate-spin shrink-0" />
                  ) : (
                    <Square className="w-3 h-3 fill-current shrink-0" />
                  )}
                  <span>
                    {batchProgress.isStopping
                      ? t('cowork.batchStoppingBtn')
                      : t('cowork.batchStopBtn')}
                  </span>
                </button>

                {/* Animated Batch Progress Indicator */}
                <div 
                  className="h-7 flex items-center space-x-1.5 px-2.5 rounded-md bg-indigo-500/15 border border-indigo-500/30 text-indigo-200 text-xs shrink-0 font-medium"
                  title={t('cowork.translatingChapterTitle', { title: batchProgress.currentChapterTitle })}
                >
                  <span className="w-2 h-2 rounded-full bg-indigo-400 animate-ping shrink-0" />
                  <span className="font-mono tabular-nums text-indigo-300">
                    {t('cowork.batchProgressStatus', {
                      current: Math.min(batchProgress.completedChapters + 1, batchProgress.totalChapters),
                      total: batchProgress.totalChapters,
                    })}
                  </span>
                  {batchProgress.volumeName && (
                    <span className="hidden xl:inline text-indigo-400 font-mono text-[11px]">
                      [{batchProgress.volumeName}]
                    </span>
                  )}
                </div>
              </>
            ) : (
              <>
                {/* Scope Selector */}
                <div 
                  className="flex items-center h-7 rounded-md bg-[#111726] border border-white/10 p-0.5 shrink-0"
                  title={t('cowork.translateScope')}
                >
                  <Layers className="w-3.5 h-3.5 text-indigo-400 ml-1.5 shrink-0" />
                  <select
                    value={translateScope}
                    onChange={(e) => {
                      const sc = e.target.value as 'chapter' | 'volume' | 'series';
                      setTranslateScope(sc);
                      if (sc === 'volume') {
                        setBatchTargetVolume(activeVolTag);
                      }
                    }}
                    disabled={isTranslating || isPreScanning}
                    className="bg-transparent text-slate-200 text-xs font-medium focus:outline-none cursor-pointer pl-1 pr-1.5 py-0 border-0 disabled:opacity-50 disabled:cursor-not-allowed [&>option]:bg-[#111726] [&>option]:text-slate-200"
                  >
                    <option value="chapter">{t('cowork.scopeChapter')}</option>
                    <option value="volume">
                      {t('cowork.scopeVolume')}{activeVolTag ? ` (${activeVolTag})` : ''}
                    </option>
                    <option value="series">{t('cowork.scopeSeries')}</option>
                  </select>
                </div>

                {/* Main Action Translate Button */}
                <button
                  type="button"
                  onClick={handleActionClick}
                  disabled={isTranslating || isPreScanning}
                  className={`h-7 flex items-center space-x-1.5 px-2.5 sm:px-3 rounded-md text-xs font-semibold whitespace-nowrap shrink-0 transition ${
                    isTranslating || isPreScanning
                      ? 'bg-white/5 text-slate-400 border border-white/10 cursor-not-allowed'
                      : 'bg-indigo-600 hover:bg-indigo-500 text-white shadow-xs border border-indigo-400/30 active:scale-[0.98]'
                  }`}
                  title={
                    translateScope === 'chapter' 
                      ? t('cowork.scopeChapterLabel') 
                      : translateScope === 'volume' 
                      ? t('cowork.scopeVolumeLabel', { volume: activeVolTag }) 
                      : t('cowork.scopeSeriesLabel')
                  }
                >
                  {isPreScanning ? (
                    <Loader2 className="w-3 h-3 animate-spin shrink-0 text-amber-300" />
                  ) : (
                    <Play className="w-3 h-3 fill-current shrink-0" />
                  )}
                  <span>
                    {isTranslating
                      ? t('cowork.translating')
                      : isPreScanning
                      ? t('cowork.scanning')
                      : translateScope === 'chapter'
                      ? t('projects.translateNow')
                      : translateScope === 'volume'
                      ? `${t('cowork.scopeVolume')}${activeVolTag ? ` [${activeVolTag}]` : ''}`
                      : t('cowork.scopeSeries')}
                  </span>
                </button>
              </>
            )}

            {/* Translation Engine Mode Picker */}
            <Select
              value={translationMode}
              onChange={(e) => setTranslationMode(e.target.value)}
              className="w-24 sm:w-32 md:w-36 h-7 truncate text-xs text-slate-300"
              containerClassName="shrink-0 max-w-[120px] sm:max-w-none"
              title={t('cowork.translationModeTooltip')}
            >
              <option value="concurrent_dual_agent">Dual-Agent (Critic)</option>
              <option value="hierarchical_3pass">Hierarchical 3-Pass</option>
              <option value="single_pass">Single Pass</option>
              <option value="swarm_arc_parallel">Swarm Arc</option>
            </Select>

            {/* Rollback to Checkpoint Button */}
            <button
              onClick={handleRollback}
              disabled={isTranslating || isRollingBack || !selectedChapter}
              className="h-7 flex items-center space-x-1 px-2 sm:px-2.5 rounded-md text-xs font-medium text-slate-300 hover:text-white bg-[#111726] hover:bg-white/8 border border-white/10 transition disabled:opacity-40 disabled:cursor-not-allowed shrink-0"
              title={t('cowork.rollbackTooltip')}
            >
              <RotateCcw className={`w-3 h-3 text-amber-400 shrink-0 ${isRollingBack ? 'animate-spin' : ''}`} />
              <span className="hidden xl:inline">{t('cowork.rollbackCheckpoint')}</span>
            </button>

            {/* Toggle Lane 2 Sidebar Button */}
            <button
              onClick={() => setIsSoulPanelOpen(!isSoulPanelOpen)}
              className={`h-7 px-2 sm:px-2.5 rounded-md border text-xs transition whitespace-nowrap shrink-0 flex items-center space-x-1 ${
                isSoulPanelOpen
                  ? 'bg-indigo-500/15 border-indigo-500/30 text-indigo-300 hover:bg-indigo-500/25'
                  : 'bg-[#111726] border-white/10 text-slate-300 hover:text-white hover:bg-white/8'
              }`}
              title={isSoulPanelOpen ? t('cowork.collapseChat') : t('cowork.openChat')}
            >
              {isSoulPanelOpen ? (
                <PanelRightClose className="w-3.5 h-3.5 shrink-0" />
              ) : (
                <>
                  <MessageSquare className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
                  <span className="text-[11px] font-medium hidden xl:inline">NovelClaw</span>
                </>
              )}
            </button>

            {/* Graph Readiness Indicator / 1-Click Scan Trigger */}
            {chapterGraphReadiness.isReady ? (
              <div
                className="hidden lg:flex items-center space-x-1.5 h-7 px-2 rounded-md text-[11px] font-medium bg-emerald-500/10 text-emerald-300 border border-emerald-500/20 whitespace-nowrap shrink-0 cursor-default"
                title={chapterGraphReadiness.message}
              >
                <Check className="w-3 h-3 text-emerald-400 shrink-0" />
                <span className="hidden 2xl:inline truncate max-w-32.5">{chapterGraphReadiness.badgeText}</span>
              </div>
            ) : (
              <button
                onClick={() => setIsPreTranslateModalOpen(true)}
                className="flex items-center space-x-1.5 h-7 px-2 rounded-md text-[11px] font-semibold bg-amber-500/15 text-amber-300 border border-amber-500/30 hover:bg-amber-500/25 transition whitespace-nowrap shrink-0"
                title={chapterGraphReadiness.message}
              >
                <AlertTriangle className="w-3 h-3 text-amber-400 shrink-0 animate-pulse" />
                <span className="hidden 2xl:inline truncate max-w-35">{chapterGraphReadiness.badgeText}</span>
              </button>
            )}
          </div>

          {/* Right Cluster: Active Tool & Fallback Telemetry if Panel Closed */}
          <div className="flex items-center gap-2 sm:gap-2.5 shrink-0 min-w-0">
            {activeTool && (
              <div className="flex items-center space-x-1 px-2 py-0.5 rounded-full bg-indigo-500/10 border border-indigo-500/20 text-indigo-300 text-[10px] whitespace-nowrap shrink-0">
                <Search className="w-2.5 h-2.5 shrink-0" />
                <span className="max-w-40 sm:max-w-55 truncate">{activeTool}</span>
              </div>
            )}

            {/* Fallback Telemetry Metrics shown ONLY when Soul panel is closed */}
            {!isSoulPanelOpen && (
              <div className="flex items-center gap-2 text-slate-400 text-[11px] font-mono shrink-0">
                <div className="flex items-center space-x-1 whitespace-nowrap" title={t('cowork.speed')}>
                  <Zap className="w-3 h-3 text-amber-400/80 shrink-0" />
                  <strong className="text-slate-200">
                    {telemetry ? (
                      (() => {
                        const tokPerSec = telemetry.tokens_per_second || (telemetry.duration_ms ? (telemetry.total_tokens / (telemetry.duration_ms / 1000)) : 0);
                        return tokPerSec > 0 ? (
                          <>
                            <span className="text-amber-300">{tokPerSec.toFixed(1)}</span>
                            <span className="text-slate-400 text-[10px] ml-0.5">tok/s</span>
                          </>
                        ) : (
                          `${telemetry.runes_per_second.toFixed(1)}r/s`
                        );
                      })()
                    ) : '--'}
                  </strong>
                </div>

                <div className="flex items-center space-x-1 whitespace-nowrap" title={t('cowork.tokens')}>
                  <Cpu className="w-3 h-3 text-indigo-400/80 shrink-0" />
                  <strong className="text-slate-200">
                    {telemetry ? telemetry.total_tokens.toLocaleString() : '--'}
                  </strong>
                </div>

                <div className="flex items-center space-x-1 whitespace-nowrap" title={t('cowork.cost')}>
                  <DollarSign className="w-3 h-3 text-emerald-400 shrink-0" />
                  <strong className="text-emerald-400">
                    {telemetry ? `$${telemetry.estimated_cost_usd.toFixed(4)}` : '$0.00'}
                  </strong>
                </div>
              </div>
            )}
          </div>
        </div>

        {/* Human Coworker Clarification Banner (HITL Stage 7) */}
        {pendingAskHuman && (
          <div className="bg-amber-500/10 border-b border-amber-500/20 px-4 py-3 flex flex-col gap-3 animate-in slide-in-from-top duration-200 shrink-0">
            <div className="flex items-start justify-between gap-3">
              <div className="flex items-start gap-2.5">
                <div className="w-7 h-7 rounded-lg bg-amber-500/20 text-amber-400 flex items-center justify-center shrink-0 border border-amber-500/30">
                  <Sparkles className="w-4 h-4" />
                </div>
                <div>
                  <div className="text-xs font-bold text-amber-200 flex items-center gap-2">
                    <span>{t('cowork.hitlTitle')}</span>
                    {hitlCountdown !== null && hitlCountdown > 0 && (
                      <span className="inline-flex items-center gap-1 px-1.5 py-0.5 rounded-md bg-amber-500/20 border border-amber-500/30 text-[10px] font-mono text-amber-300 tabular-nums">
                        <Timer className="w-3 h-3" />
                        {hitlCountdown}s
                      </span>
                    )}
                    {hitlCountdown !== null && hitlCountdown <= 0 && (
                      <span className="text-[10px] text-slate-400 italic">{t('cowork.hitlTimeoutExpired')}</span>
                    )}
                  </div>
                  <div className="text-xs text-slate-200 mt-0.5 font-medium">{pendingAskHuman.question}</div>
                  {pendingAskHuman.contextSnippet && (
                    <div className="text-[11px] text-slate-400 italic mt-0.5">
                      {t('cowork.contextLabel')}: &quot;{pendingAskHuman.contextSnippet}&quot;
                    </div>
                  )}
                </div>
              </div>

              <button
                onClick={async () => {
                  const fallback = pendingAskHuman.options?.[0] || 'Default';
                  await resolveAskHuman(pendingAskHuman.jobId || '', fallback);
                  setPendingAskHuman(null);
                }}
                className="p-1 text-slate-400 hover:text-white transition shrink-0"
                title={t('cowork.skipHitlTooltip')}
              >
                <X className="w-3.5 h-3.5" />
              </button>
            </div>

            <div className="flex flex-col sm:flex-row items-start sm:items-center gap-2 pl-9">
              {/* Option buttons */}
              <div className="flex items-center gap-2 flex-wrap">
                {pendingAskHuman.options && pendingAskHuman.options.length > 0 ? (
                  pendingAskHuman.options.map((opt, idx) => (
                    <button
                      key={idx}
                      onClick={async () => {
                        await resolveAskHuman(pendingAskHuman.jobId || '', opt);
                        sendNovelClawMessage(`[HITL] ${t('cowork.iChoose')}: ${opt} — ${t('cowork.forQuestion')}: "${pendingAskHuman.question}"`);
                      }}
                      className="px-3 py-1.5 bg-amber-500/20 hover:bg-amber-500/30 text-amber-200 border border-amber-500/30 rounded-lg text-xs font-semibold transition hover:scale-[1.02] active:scale-95"
                    >
                      {opt}
                    </button>
                  ))
                ) : (
                  <button
                    onClick={async () => {
                      await resolveAskHuman(pendingAskHuman.jobId || '', t('cowork.agreeRecommendation'));
                      sendNovelClawMessage(`[HITL] ${t('cowork.agreeRecommendation')} — ${t('cowork.forQuestion')}: "${pendingAskHuman.question}"`);
                    }}
                    className="px-3 py-1.5 bg-amber-500/20 hover:bg-amber-500/30 text-amber-200 border border-amber-500/30 rounded-lg text-xs font-semibold transition"
                  >
                    {t('cowork.agreeRecommendation')}
                  </button>
                )}
              </div>

              {/* Custom input */}
              <div className="flex items-center gap-1.5 flex-1 min-w-0 w-full sm:w-auto">
                <input
                  type="text"
                  value={hitlCustomInput}
                  onChange={(e) => setHitlCustomInput(e.target.value)}
                  onKeyDown={async (e) => {
                    if (e.key === 'Enter' && hitlCustomInput.trim()) {
                      await resolveAskHuman(pendingAskHuman.jobId || '', hitlCustomInput.trim());
                      sendNovelClawMessage(`[HITL] ${t('cowork.customInstruction')}: "${hitlCustomInput.trim()}" — ${t('cowork.forQuestion')}: "${pendingAskHuman.question}"`);
                      setHitlCustomInput('');
                    }
                  }}
                  placeholder={t('cowork.customInstructionPlaceholder')}
                  className="flex-1 min-w-0 h-7 px-2.5 bg-[#111726] border border-white/10 hover:border-amber-500/30 focus:border-amber-500/50 rounded-lg text-xs text-slate-200 placeholder:text-slate-500 focus:outline-none focus:ring-1 focus:ring-amber-500/20 transition"
                />
                <button
                  onClick={async () => {
                    if (hitlCustomInput.trim()) {
                      await resolveAskHuman(pendingAskHuman.jobId || '', hitlCustomInput.trim());
                      sendNovelClawMessage(`[HITL] ${t('cowork.customInstruction')}: "${hitlCustomInput.trim()}" — ${t('cowork.forQuestion')}: "${pendingAskHuman.question}"`);
                      setHitlCustomInput('');
                    }
                  }}
                  disabled={!hitlCustomInput.trim()}
                  className="h-7 px-2.5 bg-amber-500/20 hover:bg-amber-500/30 text-amber-300 border border-amber-500/30 rounded-lg text-xs font-semibold transition disabled:opacity-40 disabled:cursor-not-allowed shrink-0"
                >
                  <Send className="w-3 h-3" />
                </button>
              </div>
            </div>

            {hitlCountdown !== null && hitlCountdown > 0 && (
              <div className="pl-9 text-[10px] text-slate-500">
                {t('cowork.hitlAutoSelectCountdown', { count: hitlCountdown })}
              </div>
            )}
          </div>
        )}

        {/* Main Content Pane (Dual Pane: Raw vs Translated) */}
        <div className="flex-1 flex flex-col md:flex-row overflow-hidden min-h-0">
          {/* Raw Text Column */}
          {(paneLayout === 'split' || paneLayout === 'source') && (
            <div className={`flex-1 ${paneLayout === 'split' ? 'border-b md:border-b-0 md:border-r border-white/8' : ''} p-3 sm:p-4 md:p-5 overflow-hidden flex flex-col bg-[#0b0f18] min-w-0`}>
              <div className="flex items-center justify-between pb-2 mb-2 border-b border-white/6 shrink-0 min-w-0 gap-2">
                <div className="flex items-center space-x-1.5 min-w-0">
                  <span className="text-xs font-semibold uppercase tracking-wider text-slate-300 truncate">
                    {t('cowork.sourceText')}
                  </span>
                  <span className="px-1.5 py-0.5 rounded text-[10px] font-mono font-bold bg-white/6 text-slate-300 border border-white/10 uppercase shrink-0" title={`${t('cowork.sourceText')}: ${(selectedProject.source_lang || 'ja').toUpperCase()}`}>
                    {(selectedProject.source_lang || 'ja').toUpperCase()}
                  </span>
                </div>

                <div className="flex items-center space-x-1 shrink-0">
                  <button
                    onClick={() => setRawViewMode(rawViewMode === 'rendered' ? 'code' : 'rendered')}
                    className="flex items-center space-x-1 px-2 py-0.5 rounded text-[11px] font-medium bg-white/4 hover:bg-white/8 text-slate-300 border border-white/10 transition shrink-0"
                    title={rawViewMode === 'rendered' ? t('cowork.viewRawCode') : t('cowork.viewBookPage')}
                  >
                    {rawViewMode === 'rendered' ? (
                      <>
                        <Code2 className="w-3 h-3 text-slate-400 shrink-0" />
                        <span className="hidden sm:inline">{t('cowork.sourceCode')}</span>
                      </>
                    ) : (
                      <>
                        <BookOpen className="w-3 h-3 text-indigo-400 shrink-0" />
                        <span className="hidden sm:inline">{t('cowork.bookPage')}</span>
                      </>
                    )}
                  </button>
                </div>
              </div>

              {rawViewMode === 'rendered' ? (
                <div 
                  className="reader-content flex-1 overflow-y-auto overflow-x-hidden pr-1 text-[13.5px] leading-7 font-serif text-slate-200 select-text"
                  dangerouslySetInnerHTML={{ 
                    __html: sanitizeReaderHtml(selectedChapter?.raw_content || '', selectedProject?.id, selectedChapter?.content_path) || `<p class="text-slate-500 italic">(${t('cowork.noContent')})</p>` 
                  }}
                />
              ) : (
                <div className="flex-1 overflow-y-auto overflow-x-hidden text-xs font-mono text-slate-300 leading-6 whitespace-pre-wrap select-text bg-[#070a10] p-3 rounded-lg border border-white/5">
                  {selectedChapter?.raw_content || t('cowork.noContent')}
                </div>
              )}
            </div>
          )}

          {/* Translated Text Column */}
          {(paneLayout === 'split' || paneLayout === 'target') && (
            <div className="flex-1 p-3 sm:p-4 md:p-5 overflow-hidden flex flex-col bg-[#0d121c] min-w-0">
              <div className="flex items-center justify-between pb-2 mb-2 border-b border-white/6 shrink-0 min-w-0 gap-2">
                <div className="flex items-center space-x-1.5 min-w-0">
                  <span className="text-xs font-semibold uppercase tracking-wider text-indigo-300 truncate">
                    {t('cowork.translatedText')}
                  </span>
                  <span className="px-1.5 py-0.5 rounded text-[10px] font-mono font-bold bg-indigo-500/10 text-indigo-300 border border-indigo-500/20 uppercase shrink-0" title={`${t('cowork.translatedText')}: ${(selectedProject.target_lang || 'en').toUpperCase()}`}>
                    {(selectedProject.target_lang || 'en').toUpperCase()}
                  </span>
                </div>

                <div className="flex items-center space-x-1.5 shrink-0">
                  {!isInlineEditing && (
                    <button
                      onClick={() => setTranslatedViewMode(translatedViewMode === 'rendered' ? 'code' : 'rendered')}
                      className="flex items-center space-x-1 px-2 py-0.5 rounded text-[11px] font-medium bg-white/4 hover:bg-white/8 text-slate-300 border border-white/10 transition shrink-0"
                      title={translatedViewMode === 'rendered' ? t('cowork.viewRawCode') : t('cowork.viewBookPage')}
                    >
                      {translatedViewMode === 'rendered' ? (
                        <>
                          <Code2 className="w-3 h-3 text-slate-400 shrink-0" />
                          <span className="hidden sm:inline">{t('cowork.sourceCode')}</span>
                        </>
                      ) : (
                        <>
                          <BookOpen className="w-3 h-3 text-indigo-400 shrink-0" />
                          <span className="hidden sm:inline">{t('cowork.bookPage')}</span>
                        </>
                      )}
                    </button>
                  )}

                  {isInlineEditing ? (
                    <button
                      onClick={handleSaveInlineEdit}
                      className="flex items-center space-x-1 px-2 py-0.5 bg-emerald-600 hover:bg-emerald-500 text-white rounded text-[11px] font-medium transition whitespace-nowrap shrink-0 shadow-xs"
                    >
                      <Save className="w-3 h-3 shrink-0" />
                      <span>{t('cowork.saveChanges')}</span>
                    </button>
                  ) : (
                    <button
                      onClick={() => !isTranslating && setIsInlineEditing(true)}
                      disabled={isTranslating}
                      className="flex items-center space-x-1 px-2 py-0.5 bg-white/4 hover:bg-white/8 text-slate-300 border border-white/10 rounded text-[11px] font-medium transition whitespace-nowrap shrink-0 disabled:opacity-40 disabled:cursor-not-allowed"
                      title={isTranslating ? t('cowork.cannotEditWhileTranslating') : t('cowork.inlineEdit')}
                    >
                      <Edit3 className="w-3 h-3 shrink-0" />
                      <span className="hidden sm:inline">{t('cowork.inlineEdit')}</span>
                    </button>
                  )}
                </div>
              </div>

              {isInlineEditing ? (
                <textarea
                  className="flex-1 w-full p-3.5 bg-[#090d16] border border-indigo-500/40 rounded-lg text-[13.5px] text-slate-100 font-serif leading-7 focus:outline-none focus:ring-1 focus:ring-indigo-500/30 resize-none overflow-y-auto overflow-x-hidden"
                  value={editableContent}
                  onChange={(e) => setEditableContent(e.target.value)}
                />
              ) : translatedViewMode === 'rendered' ? (
                selectedChapter?.translated_content ? (
                  <div 
                    ref={translatedContentRef}
                    className="reader-content flex-1 overflow-y-auto overflow-x-hidden pr-1 text-[13.5px] leading-7 font-serif text-slate-100 select-text relative"
                  >
                    <div dangerouslySetInnerHTML={{ __html: sanitizeReaderHtml(selectedChapter.translated_content, selectedProject?.id, selectedChapter?.content_path) }} />
                    {isTranslating && (
                      <div className="flex items-center space-x-2 my-3 px-3 py-1.5 rounded-md bg-indigo-500/10 border border-indigo-500/20 text-indigo-300 text-xs font-sans animate-pulse select-none">
                        <Loader2 className="w-3.5 h-3.5 animate-spin text-indigo-400 shrink-0" />
                        <span>{t('cowork.streamingSegments')}</span>
                      </div>
                    )}
                  </div>
                ) : isTranslating ? (
                  <div className="flex-1 flex flex-col items-center justify-center text-slate-400 text-xs space-y-2 p-6">
                    <Loader2 className="w-6 h-6 animate-spin text-indigo-400" />
                    <span className="text-slate-200 font-semibold">{t('cowork.initializingTranslation')}</span>
                    <span className="text-[11px] text-slate-400 text-center max-w-xs leading-relaxed">
                      {t('cowork.streamingDirectNotice')}
                    </span>
                  </div>
                ) : (
                  <div className="flex-1 flex items-center justify-center text-slate-500 italic text-sm">
                    {t('cowork.noActiveStream')}
                  </div>
                )
              ) : (
                <div 
                  ref={translatedContentRef}
                  className="flex-1 overflow-y-auto overflow-x-hidden text-xs font-mono text-slate-200 leading-6 whitespace-pre-wrap select-text bg-[#070a10] p-3 rounded-lg border border-white/5"
                >
                  {selectedChapter?.translated_content || (
                    isTranslating ? (
                      <span className="text-indigo-400 flex items-center space-x-1.5 animate-pulse font-sans">
                        <Loader2 className="w-3.5 h-3.5 animate-spin" />
                        <span>{t('cowork.streamingRawSource')}</span>
                      </span>
                    ) : (
                      <span className="text-slate-500 italic">{t('cowork.noActiveStream')}</span>
                    )
                  )}
                  {isTranslating && selectedChapter?.translated_content && (
                    <div className="mt-2 text-indigo-400 flex items-center space-x-1.5 animate-pulse font-sans">
                      <Loader2 className="w-3 h-3 animate-spin" />
                      <span>{t('cowork.streamingNextSegment')}</span>
                    </div>
                  )}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* LANE 2: NovelClaw Agentic Console (Claude Code Style Activity Stream & Omni-App Control) */}
      {isSoulPanelOpen && (
        <div className="fixed inset-y-0 right-0 z-40 w-full sm:w-96 md:relative md:inset-auto md:w-80 lg:w-96 xl:w-105 flex flex-col bg-[#0b0f19] border-l border-white/8 shadow-2xl md:shadow-none shrink-0 overflow-hidden text-slate-200">
          {/* Two-Tier Structured Header: Tier 1 (Identity & Telemetry) + Tier 2 (Thread Strip) */}
          <div className="shrink-0 flex flex-col border-b border-white/8 bg-[#0d121f]">
            {/* TIER 1: Assistant Identity, Persona Switcher & Global Stream Actions */}
            <div className="py-2.5 px-3 flex items-center justify-between gap-2 border-b border-white/6 bg-[#0e1320] select-none">
              {/* Left: Avatar + Two-line identity (Name & Selector on line 1, Archetype on line 2) */}
              <div className="flex items-center space-x-2.5 min-w-0 flex-1">
                <div className="relative shrink-0">
                  <div className="w-8 h-8 rounded-lg bg-linear-to-br from-indigo-500/20 to-purple-500/20 border border-indigo-500/30 flex items-center justify-center text-base shadow-xs">
                    {soul?.avatar || '🐾'}
                  </div>
                  <span
                    className={`absolute -bottom-0.5 -right-0.5 w-2.5 h-2.5 rounded-full border-2 border-[#0e1320] ${
                      isNovelClawThinking
                        ? 'bg-indigo-400 animate-pulse'
                        : isTranslating
                        ? 'bg-emerald-400 animate-pulse'
                        : 'bg-emerald-500'
                    }`}
                    title={
                      isNovelClawThinking
                        ? t('cowork.thinkingStatus')
                        : isTranslating
                        ? t('cowork.translatingStatus')
                        : t('cowork.readyStatus')
                    }
                  />
                </div>

                {/* Vertical Column: Name & Selector on line 1, Archetype on line 2 */}
                <div className="min-w-0 flex-1 flex flex-col justify-center">
                  {/* Line 1: Soul Name + Persona Switcher */}
                  <div className="flex items-center space-x-2 min-w-0">
                    <span className="text-xs font-bold text-slate-100 truncate tracking-tight">
                      {soul?.name || 'NovelClaw'}
                    </span>
                    {availableSouls && availableSouls.length > 1 && (
                      <select
                        value={soul?.id || ''}
                        onChange={(e) => setActiveSoul(e.target.value)}
                        className="bg-white/4 hover:bg-white/8 text-[10px] text-amber-300/90 rounded border border-amber-500/20 px-1.5 py-0.5 max-w-26.25 truncate focus:outline-none focus:border-amber-500 transition cursor-pointer"
                        title={t('cowork.switchSoulPersona')}
                      >
                        {availableSouls.map((s) => (
                          <option key={s.id} value={s.id} className="bg-[#121827] text-slate-200">
                            {s.avatar} {s.name}
                          </option>
                        ))}
                      </select>
                    )}
                  </div>

                  {/* Line 2: Archetype badge with ample width */}
                  <div className="flex items-center mt-1 min-w-0">
                    {soul?.archetype && (
                      <span
                        className="text-[9px] px-1.5 py-0.5 rounded bg-indigo-500/10 text-indigo-300 border border-indigo-500/20 font-mono truncate max-w-55"
                        title={soul.archetype}
                      >
                        {soul.archetype}
                      </span>
                    )}
                  </div>
                </div>
              </div>

              {/* Right: Quick Stream Actions (Expand Thinking, Clear Thread) */}
              <div className="flex items-center space-x-1 shrink-0">
                {/* Toggle Expand Thinking (Ctrl+O) */}
                <button
                  onClick={() => setExpandAllThinking((prev) => !prev)}
                  className={`p-1.5 rounded-lg transition ${
                    expandAllThinking
                      ? 'bg-indigo-600/20 text-indigo-300 border border-indigo-500/30'
                      : 'text-slate-400 hover:text-white hover:bg-white/6'
                  }`}
                  title={t('cowork.toggleAllThinkingTooltip')}
                >
                  {expandAllThinking ? <Minimize2 className="w-3.5 h-3.5" /> : <Maximize2 className="w-3.5 h-3.5" />}
                </button>

                {/* Clear Thread */}
                <button
                  onClick={async () => {
                    const confirmed = await showConfirmModal({
                      title: t('cowork.clearChatTitle'),
                      message: t('cowork.clearChatConfirm'),
                      confirmText: t('common.delete'),
                      cancelText: t('common.cancel'),
                      variant: 'warning',
                    });
                    if (confirmed) {
                      clearNovelClawThread();
                    }
                  }}
                  className="p-1.5 text-slate-400 hover:text-rose-300 hover:bg-white/6 rounded-lg transition"
                  title={t('cowork.clearChatTooltip')}
                >
                  <Trash2 className="w-3.5 h-3.5" />
                </button>

                {/* Close Drawer Button on Mobile */}
                <button
                  onClick={() => setIsSoulPanelOpen(false)}
                  className="p-1.5 text-slate-400 hover:text-white hover:bg-white/6 rounded-lg transition md:hidden"
                  title={t('cowork.closeChat')}
                >
                  <X className="w-3.5 h-3.5" />
                </button>
              </div>
            </div>

            {/* TIER 2: Thread Navigation Strip & Active Translation Controls */}
            <div className="h-8 px-3 flex items-center justify-between gap-2 bg-[#080c14] select-none">
              {/* Left: Thread Icon + Selector + New Sub-Chat Button */}
              <div className="flex items-center space-x-1.5 min-w-0 flex-1">
                <MessageSquare className="w-3 h-3 text-indigo-400 shrink-0" />
                <select
                  value={selectedThread?.id || ''}
                  onChange={(e) => {
                    const found = novelclawThreads.find((t) => t.id === e.target.value);
                    if (found) selectNovelClawThread(found);
                  }}
                  className="bg-transparent text-[11px] font-medium text-slate-300 hover:text-white border-0 py-0.5 pr-2 truncate max-w-45 sm:max-w-55 focus:outline-none cursor-pointer"
                  title={selectedThread ? (selectedThread.is_main ? `★ ${selectedThread.title}` : selectedThread.title) : t('cowork.selectThread')}
                >
                  {novelclawThreads.map((t) => (
                    <option key={t.id} value={t.id} className="bg-[#121827] text-slate-200">
                      {t.is_main ? `★ ${t.title}` : t.title}
                    </option>
                  ))}
                </select>

                <button
                  onClick={() => setIsCreatingSubThread(true)}
                  className="px-1.5 py-0.5 hover:bg-white/8 text-slate-400 hover:text-indigo-300 rounded transition text-[10px] flex items-center space-x-1 shrink-0 font-medium"
                  title={t('cowork.newSubchatTooltip')}
                >
                  <Plus className="w-3 h-3" />
                  <span className="hidden sm:inline">{t('cowork.newSubchat')}</span>
                </button>
              </div>

              {/* Right: Delete Sub-chat (if not main) & Contextual Stop/Abort (if Translating) */}
              <div className="flex items-center space-x-1 shrink-0">
                {selectedThread && !selectedThread.is_main && (
                  <button
                    onClick={async () => {
                      const confirmed = await showConfirmModal({
                        title: t('cowork.deleteSubchatTitle'),
                        message: t('cowork.deleteSubchatConfirm', { title: selectedThread.title }),
                        confirmText: t('cowork.deleteSubchat'),
                        cancelText: t('common.cancel'),
                        variant: 'danger',
                      });
                      if (confirmed) {
                        deleteNovelClawThread(selectedThread.id);
                      }
                    }}
                    className="px-1.5 py-0.5 hover:bg-rose-500/20 text-slate-400 hover:text-rose-400 rounded transition text-[10px] flex items-center space-x-1 border border-rose-500/10 hover:border-rose-500/30"
                    title={t('cowork.deleteSubchatTooltip')}
                  >
                    <Trash2 className="w-2.5 h-2.5 text-rose-400" />
                    <span className="text-[10px] text-rose-400/90">{t('cowork.deleteSubchat')}</span>
                  </button>
                )}

                {/* Translation Runtime Status & Emergency Controls (Only visible when translating) */}
                {isTranslating && (
                  <div className="flex items-center space-x-1 shrink-0 pl-1 border-l border-white/10">
                    <span className="inline-flex items-center px-1.5 py-0.5 rounded text-[9px] font-medium bg-emerald-500/10 text-emerald-300 border border-emerald-500/20 animate-pulse">
                      {t('cowork.translating')}
                    </span>
                    <button
                      onClick={() => triggerSoftStop()}
                      className="p-0.5 bg-amber-500/10 hover:bg-amber-500/20 text-amber-300 border border-amber-500/20 rounded transition"
                      title={t('cowork.softStopTooltip')}
                    >
                      <PauseCircle className="w-3 h-3" />
                    </button>
                    <button
                      onClick={() => triggerHardAbort()}
                      className="p-0.5 bg-rose-500/10 hover:bg-rose-500/20 text-rose-300 border border-rose-500/20 rounded transition"
                      title={t('cowork.hardAbortTooltip')}
                    >
                      <StopCircle className="w-3 h-3" />
                    </button>
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Activity Stream & Chat History */}
          <div className="flex-1 p-3 md:p-4 overflow-y-auto space-y-3 font-sans">
            {novelclawMessages.length === 0 ? (
              <div className="h-full flex flex-col items-center justify-center text-center p-4">
                <div className="w-12 h-12 rounded-2xl bg-indigo-500/10 border border-indigo-500/20 flex items-center justify-center text-xl mb-3">
                  {soul?.avatar || '🐾'}
                </div>
                <h4 className="text-xs font-semibold text-slate-200">
                  {soul?.name ? t('cowork.assistantGreetingNamed', { name: soul.name }) : t('cowork.welcomeConsole')}
                </h4>
                <p className="text-[11px] text-slate-400 mt-1 max-w-70 leading-relaxed">
                  {soul?.greeting || t('cowork.defaultAssistantDesc')}
                </p>
                <div className="mt-3 space-y-1.5 w-full text-left">
                  {[
                    t('cowork.promptSuggestion1'),
                    t('cowork.promptSuggestion2'),
                    t('cowork.promptSuggestion3'),
                    t('cowork.promptSuggestion4'),
                    t('cowork.promptSuggestion5'),
                  ].map((sample, idx) => (
                    <button
                      key={idx}
                      onClick={() => setChatInput(sample)}
                      className="w-full text-left p-2 rounded-lg bg-white/3 hover:bg-indigo-500/10 border border-white/6 hover:border-indigo-500/30 text-[10.5px] text-slate-300 hover:text-indigo-200 transition line-clamp-2 flex items-center gap-1.5"
                    >
                      <MessageSquare className="w-3 h-3 text-indigo-400/70 shrink-0" />
                      <span className="truncate">{sample}</span>
                    </button>
                  ))}
                </div>
              </div>
            ) : (
              novelclawMessages.map(msg => {
                const isUser = msg.sender === 'user' || msg.role === 'user';
                const isExpanded = expandAllThinking || expandedThinkingIds[msg.id];

                // Parse action call JSON if available
                let actions: any[] = [];
                if (msg.action_call_json) {
                  try {
                    const parsed = JSON.parse(msg.action_call_json);
                    actions = Array.isArray(parsed) ? parsed : [parsed];
                  } catch {
                    // ignore JSON parse errors
                  }
                }

                return (
                  <div
                    key={msg.id}
                    className={`flex flex-col ${isUser ? 'items-end' : 'items-start'} space-y-1.5`}
                  >
                    <div className="flex items-center space-x-1.5 text-[10px] text-slate-500 px-1">
                      <span>{isUser ? t('cowork.userSender') : `${soul?.avatar ? soul.avatar + ' ' : ''}${soul?.name || 'NovelClaw'}`}</span>
                      <span>•</span>
                      <span>{new Date(msg.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}</span>
                      {msg.token_count > 0 && (
                        <span className="font-mono text-[9px] text-slate-500">({msg.token_count.toLocaleString()} tok)</span>
                      )}
                    </div>

                    {/* Reasoning Process (Compact single-line by default, expands on click or Ctrl+O) */}
                    {msg.thinking_content && (
                      <div className="w-full max-w-[94%] bg-[#080c14] border border-purple-500/20 rounded-xl overflow-hidden shadow-inner">
                        <button
                          type="button"
                          onClick={() => {
                            setExpandedThinkingIds(prev => ({
                              ...prev,
                              [msg.id]: !isExpanded,
                            }));
                          }}
                          className={`w-full px-3 py-1.5 flex items-center justify-between text-[11px] font-mono transition select-none ${
                            isExpanded
                              ? 'bg-purple-500/10 text-purple-200 border-b border-purple-500/15'
                              : 'bg-purple-500/5 text-purple-300 hover:bg-purple-500/10'
                          }`}
                        >
                          <span className="flex items-center space-x-2">
                            <Sparkles className="w-3.5 h-3.5 text-purple-400 shrink-0" />
                            <span className="italic font-medium">{t('cowork.reasoningProcess')}</span>
                            <span className="text-[10px] text-purple-400/60">
                              ({msg.thinking_content.split('\n').filter(Boolean).length} {t('cowork.linesCount')})
                            </span>
                          </span>
                          <span className="text-[10px] text-purple-400/80 flex items-center space-x-1">
                            <span>{isExpanded ? t('common.collapse') : t('common.expand')}</span>
                            {isExpanded ? <ChevronUp className="w-3 h-3" /> : <ChevronDown className="w-3 h-3" />}
                          </span>
                        </button>
                        {isExpanded && (
                          <div className="p-3 text-[11px] font-mono text-slate-300 italic leading-relaxed whitespace-pre-wrap max-h-80 overflow-y-auto bg-[#060810]">
                            {msg.thinking_content}
                          </div>
                        )}
                      </div>
                    )}

                    {/* 🛠️ Claude Code / Cowork Tool Call Step Card */}
                    {msg.step_type === 'tool_call' ? (
                      <ToolCallStepCard
                        msg={msg}
                        isExpanded={expandedToolIds[msg.id] ?? false}
                        onToggle={() => setExpandedToolIds(prev => ({ ...prev, [msg.id]: !prev[msg.id] }))}
                      />
                    ) : null}

                    {/* ⚡ Action Dispatched Badges & Interactive Cards (Always visible when present) */}
                    {actions.length > 0 && (
                      <div className="w-full max-w-[94%] space-y-1.5">
                        {actions.map((act, actIdx) => (
                          <div
                            key={actIdx}
                            className="p-2.5 rounded-xl bg-[#090d18] border border-indigo-500/30 text-xs shadow-sm flex flex-col space-y-1.5"
                          >
                            <div className="flex items-center justify-between">
                              <span className="flex items-center space-x-1.5 font-semibold text-indigo-300 text-[11px]">
                                <Zap className="w-3.5 h-3.5 text-amber-400" />
                                <span>[Action Dispatched]: {act.action}</span>
                              </span>
                              <span className="text-[9px] font-mono px-1.5 py-0.2 rounded bg-indigo-500/10 text-indigo-300 border border-indigo-500/20 uppercase">
                                {act.category || 'studio'}
                              </span>
                            </div>
                            <p className="text-[11px] text-slate-300 leading-snug">
                              └ {act.description}
                            </p>
                            <div className="pt-0.5 flex items-center space-x-2">
                              {(act.category === 'graph' || act.action?.includes('character') || act.action?.includes('relation')) && (
                                <button
                                  type="button"
                                  onClick={() => setActiveTab('graph')}
                                  className="text-[10px] font-medium text-indigo-400 hover:text-indigo-300 bg-indigo-500/10 hover:bg-indigo-500/20 px-2 py-0.5 rounded border border-indigo-500/20 transition flex items-center space-x-1"
                                >
                                  <span>{t('cowork.viewInL2Graph')}</span>
                                  <ArrowRight className="w-2.5 h-2.5" />
                                </button>
                              )}
                              {(act.category === 'glossary' || act.action?.includes('glossary')) && (
                                <button
                                  type="button"
                                  onClick={() => setActiveTab('glossary')}
                                  className="text-[10px] font-medium text-indigo-400 hover:text-indigo-300 bg-indigo-500/10 hover:bg-indigo-500/20 px-2 py-0.5 rounded border border-indigo-500/20 transition flex items-center space-x-1"
                                >
                                  <span>{t('cowork.viewInL3Glossary')}</span>
                                  <ArrowRight className="w-2.5 h-2.5" />
                                </button>
                              )}
                              {act.action === 'switch_tab' && act.data?.tab && (
                                <button
                                  type="button"
                                  onClick={() => setActiveTab(act.data.tab)}
                                  className="text-[10px] font-medium text-indigo-400 hover:text-indigo-300 bg-indigo-500/10 hover:bg-indigo-500/20 px-2 py-0.5 rounded border border-indigo-500/20 transition flex items-center space-x-1"
                                >
                                  <span>{t('cowork.openTabNamed', { tab: act.data.tab })}</span>
                                  <ArrowRight className="w-2.5 h-2.5" />
                                </button>
                              )}
                            </div>
                          </div>
                        ))}
                      </div>
                    )}

                    {/* 💬 Message Content Bubble / Specialized Card */}
                    {msg.step_type === 'step_status' ? (
                      /* 📍 Milestone Step Status Pill */
                      <div className="w-full max-w-[94%] px-3 py-2 rounded-xl bg-white/2 border border-white/6 flex items-center space-x-2.5 text-xs text-slate-300 shadow-sm animate-in fade-in duration-150">
                        <div className="w-2 h-2 rounded-full bg-indigo-400 animate-pulse shrink-0" />
                        <span className="font-mono text-[11px] leading-relaxed text-indigo-200/90 flex-1">{msg.content}</span>
                        <span className="text-[9px] font-mono text-slate-500 shrink-0">
                          {new Date(msg.created_at).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
                        </span>
                      </div>
                    ) : msg.step_type === 'translation_summary' ? (
                      /* 📊 Translation Summary Card */
                      <TranslationSummaryCard msg={msg} />
                    ) : msg.step_type === 'compact_summary' ? (
                      /* Compact Distillation Message Banner */
                      <div className="w-full max-w-[94%] p-3 rounded-xl bg-[#091512] border border-emerald-500/30 text-emerald-200 text-xs leading-relaxed shadow-sm">
                        <div className="flex items-center space-x-1.5 font-semibold text-emerald-400 mb-1.5">
                          <Brain className="w-3.5 h-3.5" />
                          <span>{t('cowork.distilledContext')}</span>
                        </div>
                        <div className="font-mono text-[11px] whitespace-pre-wrap opacity-90 leading-5">
                          {msg.content}
                        </div>
                      </div>
                    ) : msg.step_type === 'compact_notice' ? (
                      <div className="w-full text-center py-1">
                        <span className="text-[10.5px] text-indigo-300/80 bg-indigo-500/10 border border-indigo-500/20 px-2.5 py-1 rounded-full">
                          {msg.content}
                        </span>
                      </div>
                    ) : msg.step_type === 'tool_call' ? null : (
                      /* Standard Message Bubble */
                      msg.content ? (
                        <div
                          className={`max-w-[90%] rounded-xl p-3 text-xs leading-relaxed whitespace-pre-wrap select-text ${
                            isUser
                              ? 'bg-indigo-600/20 text-indigo-100 border border-indigo-500/30 rounded-br-xs'
                              : 'bg-white/4 text-slate-200 border border-white/8 rounded-bl-xs'
                          }`}
                        >
                          {msg.content}
                        </div>
                      ) : null
                    )}
                  </div>
                );
              })
            )}

            {/* Live Thinking / Active Tool Streaming Indicator (Compact by default, click to expand live stream) */}
            {(isNovelClawThinking || isTranslating) && (
              <div className="flex flex-col items-start space-y-1.5 animate-in fade-in duration-150 w-full">
                <div className="flex items-center space-x-1.5 text-[10px] text-slate-500 px-1">
                  <span>{soul?.name || 'NovelClaw'}</span>
                  <span>•</span>
                  <span>{t('cowork.thinkingStatus')}</span>
                </div>
                <div className="w-full max-w-[94%] bg-[#080c14] border border-indigo-500/30 rounded-xl overflow-hidden shadow-inner">
                  <div
                    onClick={() => setLiveThinkingExpanded(prev => !prev)}
                    className="w-full px-3 py-2 flex items-center justify-between cursor-pointer bg-indigo-500/5 hover:bg-indigo-500/10 transition select-none gap-2"
                  >
                    <div className="flex items-center space-x-2 min-w-0 flex-1">
                      <Loader2 className="w-3.5 h-3.5 animate-spin text-indigo-400 shrink-0" />
                      <span className="italic text-[11px] font-mono font-medium text-indigo-200 shrink-0">{t('cowork.thinkingStatus')}</span>
                      {novelclawActiveTool && (
                        <span 
                          className="text-[10px] bg-indigo-500/20 text-indigo-300 border border-indigo-500/30 px-1.5 py-0.5 rounded truncate font-mono max-w-32.5 sm:max-w-42.5 shrink min-w-0"
                          title={novelclawActiveTool}
                        >
                          ● {novelclawActiveTool}
                        </span>
                      )}
                    </div>
                    <div className="flex items-center space-x-1 text-[10px] text-indigo-400/80 shrink-0">
                      <span>{liveThinkingExpanded ? t('common.collapse') : t('cowork.viewDetails')}</span>
                      {liveThinkingExpanded ? <ChevronUp className="w-3 h-3 shrink-0" /> : <ChevronDown className="w-3 h-3 shrink-0" />}
                    </div>
                  </div>
                  {liveThinkingExpanded && (
                    <div className="p-3 text-[11px] font-mono text-slate-300 italic whitespace-pre-wrap leading-relaxed max-h-60 overflow-y-auto bg-[#060910] border-t border-indigo-500/20">
                      {novelclawThinkingText || thoughtText || t('cowork.defaultThinkingText')}
                    </div>
                  )}
                </div>
              </div>
            )}

            <div ref={chatBottomRef} />
          </div>

          {/* Smart Input Box & Slash Command Autocomplete */}
          <div className="p-3 border-t border-white/8 bg-[#0c101a] shrink-0 relative">
            {/* Slash Command Suggestions Popover */}
            {filteredSlashCommands.length > 0 && (
              <div className="absolute bottom-full left-3 right-3 mb-2 bg-[#0e1322] border border-indigo-500/30 rounded-xl shadow-2xl overflow-hidden z-20 max-h-56 overflow-y-auto">
                <div className="px-3 py-1.5 border-b border-white/6 text-[10px] font-mono text-slate-400 flex items-center justify-between bg-[#090d18]">
                  <span>{t('cowork.slashSuggestionsTitle')}</span>
                  <span>{t('cowork.slashSuggestionsHint')}</span>
                </div>
                <div className="p-1">
                  {filteredSlashCommands.map((item) => (
                    <button
                      key={item.cmd}
                      type="button"
                      onClick={() => {
                        setChatInput(item.cmd + ' ');
                      }}
                      className="w-full text-left px-2.5 py-1.5 rounded-lg hover:bg-indigo-600/20 transition flex items-center justify-between text-xs"
                    >
                      <div className="flex items-center space-x-2">
                        <span className="font-mono font-semibold text-indigo-300">{item.cmd}</span>
                        <span className="text-[11px] text-slate-300">{t(item.descKey)}</span>
                      </div>
                      <span className="text-[9.5px] font-mono text-slate-500 hidden sm:inline">{item.syntax}</span>
                    </button>
                  ))}
                </div>
              </div>
            )}

            {/* Minimalist Live Translation Telemetry (Ultra-thin single-line status) */}
            {telemetry && (() => {
              const tokPerSec = telemetry.tokens_per_second || (telemetry.duration_ms ? (telemetry.total_tokens / (telemetry.duration_ms / 1000)) : 0);
              return (
                <div className="flex items-center justify-between px-1 pb-1.5 text-[10px] font-mono text-slate-400 select-none whitespace-nowrap">
                  <span title={t('cowork.speed')} className="flex items-center">
                    <span className="text-amber-300 font-semibold">
                      {tokPerSec > 0 ? tokPerSec.toFixed(1) : telemetry.runes_per_second.toFixed(1)}
                    </span>
                    <span className="text-slate-400 ml-0.5 font-medium">tok/s</span>
                    {telemetry.runes_per_second > 0 && (
                      <span className="text-slate-500 ml-1 font-normal">({telemetry.runes_per_second.toFixed(1)} r/s)</span>
                    )}
                  </span>
                  <span className="text-slate-700">•</span>
                  <span title={t('cowork.tokens')}>
                    <span className="text-slate-200 font-semibold">{telemetry.total_tokens.toLocaleString()}</span>
                    <span className="text-slate-500 ml-0.5">tok</span>
                  </span>
                  <span className="text-slate-700">•</span>
                  <span title={t('cowork.cost')} className="text-emerald-400/90 font-semibold">
                    ${telemetry.estimated_cost_usd.toFixed(4)}
                  </span>
                  {(telemetry.reasoning_tokens || 0) > 0 && (
                    <>
                      <span className="text-slate-700">•</span>
                      <span title={t('cowork.reasoningTokens')} className="text-purple-300 font-semibold">
                        {telemetry.reasoning_tokens?.toLocaleString()} <span className="text-slate-500 ml-0.5">think</span>
                      </span>
                    </>
                  )}
                  {telemetry.revised_count > 0 && (
                    <>
                      <span className="text-slate-700">•</span>
                      <span title={t('cowork.shadowCriticRevisions')}>
                        <span className="text-indigo-300 font-semibold">{telemetry.revised_count}</span>
                        <span className="text-slate-500 ml-0.5">rev</span>
                      </span>
                    </>
                  )}
                </div>
              );
            })()}

            {/* Context Telemetry & Mode Bar (Placed right above the chat input) */}
            <div className="flex items-center justify-between pb-2 px-0.5 text-[10px] text-slate-400 select-none gap-1.5 flex-wrap">
              {/* Token Context Button */}
              <button
                type="button"
                onClick={() => compactNovelClawThread()}
                className="flex items-center space-x-1.5 px-2 py-0.5 rounded-md bg-white/4 hover:bg-indigo-500/10 border border-white/10 hover:border-indigo-500/30 text-slate-300 hover:text-indigo-200 transition font-mono shrink-0"
                title={t('cowork.compactContextTooltip')}
              >
                <Brain className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
                <span>Context: {(activeTokens / 1000).toFixed(1)}k / {((autoCompactThreshold || 250000) / 1000).toFixed(0)}k ({tokenPercent}%)</span>
              </button>

              <div className="flex items-center space-x-1.5 shrink-0">
                {/* Thinking Effort Quick Selector */}
                <div
                  className="flex items-center space-x-1 px-1.5 py-0.5 rounded-md bg-white/4 border border-white/10 text-slate-300 font-mono"
                  title={t('settings.thinkingEffort')}
                >
                  <span className={`text-[9.5px] font-semibold ${thinkingEffort === 'off' ? 'text-slate-500' : 'text-purple-400'}`}>🧠</span>
                  <select
                    value={thinkingEffort}
                    onChange={(e) => setThinkingEffort(e.target.value as any)}
                    className="bg-transparent text-[10px] text-slate-200 focus:outline-none cursor-pointer pr-1"
                  >
                    <option value="off" className="bg-[#0e1322] text-slate-300">Think: Off</option>
                    <option value="low" className="bg-[#0e1322] text-slate-300">Think: Low</option>
                    <option value="medium" className="bg-[#0e1322] text-slate-300">Think: Med</option>
                    <option value="high" className="bg-[#0e1322] text-slate-300">Think: High</option>
                    <option value="xhigh" className="bg-[#0e1322] text-slate-300">Think: XHigh</option>
                    <option value="max" className="bg-[#0e1322] text-slate-300">Think: Max</option>
                  </select>
                </div>

                {/* Decision Mode Toggle: Auto Process vs Always Review */}
                <button
                  type="button"
                  onClick={() => setDecisionMode(decisionMode === 'manual' ? 'auto' : 'manual')}
                  className={`flex items-center space-x-1 px-2 py-0.5 rounded-md border text-[10px] font-medium transition cursor-pointer shrink-0 ${
                    decisionMode === 'auto'
                      ? 'bg-emerald-500/15 border-emerald-500/30 text-emerald-300 hover:bg-emerald-500/25'
                      : 'bg-amber-500/10 border-amber-500/30 text-amber-300 hover:bg-amber-500/20'
                  }`}
                  title={decisionMode === 'auto' ? t('cowork.autoProcessDesc') : t('cowork.alwaysReviewDesc')}
                >
                  {decisionMode === 'auto' ? (
                    <>
                      <Zap className="w-3 h-3 text-emerald-400 shrink-0" />
                      <span>Auto Process</span>
                    </>
                  ) : (
                    <>
                      <User className="w-3 h-3 text-amber-400 shrink-0" />
                      <span>Always Review</span>
                    </>
                  )}
                </button>
              </div>
            </div>

            <form onSubmit={handleSendChat}>
              <div className="flex items-center space-x-2">
                <div className="relative flex-1">
                  <input
                    type="text"
                    className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-3 py-2 text-xs text-slate-100 placeholder-slate-500 focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                    placeholder={t('cowork.inputPlaceholder')}
                    value={chatInput}
                    onChange={(e) => setChatInput(e.target.value)}
                  />
                  {chatInput.startsWith('/') && (
                    <span className="absolute right-2 top-2 text-[10px] font-mono text-indigo-400">
                      /cmd
                    </span>
                  )}
                </div>
                <button
                  type="submit"
                  disabled={!chatInput.trim() || isNovelClawThinking}
                  className="p-2 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg transition shadow-xs shrink-0 disabled:opacity-40"
                  title={t('cowork.sendTooltip')}
                >
                  <Send className="w-3.5 h-3.5" />
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Add Volume Modal Dialog */}
      {isAddingVolume && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
          <div className="bg-[#0e1320] border border-white/10rounded-2xl max-w-md w-full p-5 shadow-2xl animate-in fade-in zoom-in-95 duration-150">
            <div className="flex items-center justify-between pb-3 border-b border-white/8">
              <h3 className="text-sm font-semibold text-slate-100 flex items-center space-x-2">
                <BookPlus className="w-4 h-4 text-indigo-400" />
                <span>{t('cowork.addVolumeTitle')}</span>
              </h3>
              <button
                onClick={() => setIsAddingVolume(false)}
                className="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-white/6 transition"
              >
                ✕
              </button>
            </div>

            <div className="mt-4 space-y-3">
              <div>
                <label className="text-xs text-slate-400 mb-1 block">{t('cowork.selectedBookFile')}</label>
                <div className="p-2.5 rounded-lg bg-black/40 border border-white/8 text-xs font-mono text-slate-300 break-all">
                  {newVolPath}
                </div>
              </div>

              <div>
                <label className="text-xs text-slate-400 mb-1 block">{t('cowork.volumeTagLabel')}</label>
                <input
                  type="text"
                  value={newVolLabel}
                  onChange={(e) => setNewVolLabel(e.target.value)}
                  placeholder={t('cowork.volumeTagPlaceholder')}
                  className="w-full bg-[#111726] border border-white/10 rounded-lg px-3 py-2 text-xs text-slate-100 focus:outline-none focus:border-indigo-500"
                />
                <p className="text-[11px] text-slate-500 mt-1">
                  {t('cowork.volumeTagDesc')}
                </p>
              </div>

              <div className="flex items-center justify-end space-x-2 pt-2">
                <button
                  onClick={() => setIsAddingVolume(false)}
                  disabled={isAppendingVol}
                  className="px-3 py-1.5 rounded-lg text-xs font-medium text-slate-400 hover:text-white bg-white/4 hover:bg-white/8 transition"
                >
                  {t('common.cancel')}
                </button>
                <button
                  onClick={handleConfirmAddVolume}
                  disabled={isAppendingVol || !newVolPath}
                  className="px-4 py-1.5 rounded-lg text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-500 transition disabled:opacity-50 flex items-center space-x-1.5 shadow-sm"
                >
                  {isAppendingVol ? (
                    <span>{t('cowork.loadingVolume')}</span>
                  ) : (
                    <>
                      <Check className="w-3.5 h-3.5" />
                      <span>{t('cowork.confirmAddVolume')}</span>
                    </>
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Create Sub-chat Thread Modal Dialog */}
      {isCreatingSubThread && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4 animate-in fade-in duration-150">
          <div className="bg-[#0e1320] border border-white/10rounded-2xl max-w-md w-full p-5 shadow-2xl">
            <div className="flex items-center justify-between pb-3 border-b border-white/8">
              <h3 className="text-sm font-semibold text-slate-100 flex items-center space-x-2">
                <Plus className="w-4 h-4 text-indigo-400" />
                <span>{t('cowork.createSubchatTitle')}</span>
              </h3>
              <button
                onClick={() => setIsCreatingSubThread(false)}
                className="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-white/6 transition"
              >
                ✕
              </button>
            </div>

            <div className="mt-4 space-y-3">
              <div>
                <label className="text-xs text-slate-400 mb-1 block">{t('cowork.subchatTitleLabel')}</label>
                <input
                  type="text"
                  value={subThreadTitle}
                  onChange={(e) => setSubThreadTitle(e.target.value)}
                  placeholder={t('cowork.subchatTitlePlaceholder')}
                  className="w-full bg-[#111726] border border-white/10 rounded-lg px-3 py-2 text-xs text-slate-100 focus:outline-none focus:border-indigo-500"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="text-xs text-slate-400 mb-1 block">{t('cowork.volumeIndexLabel')}</label>
                  <input
                    type="number"
                    value={subThreadVol}
                    onChange={(e) => setSubThreadVol(Number(e.target.value))}
                    min={0}
                    className="w-full bg-[#111726] border border-white/10 rounded-lg px-3 py-2 text-xs text-slate-100 focus:outline-none focus:border-indigo-500"
                  />
                  <p className="text-[10px] text-slate-500 mt-1">{t('cowork.volumeIndexHint')}</p>
                </div>

                <div>
                  <label className="text-xs text-slate-400 mb-1 block">{t('cowork.chapterIndexLabel')}</label>
                  <input
                    type="number"
                    value={subThreadChap}
                    onChange={(e) => setSubThreadChap(Number(e.target.value))}
                    min={0}
                    className="w-full bg-[#111726] border border-white/10 rounded-lg px-3 py-2 text-xs text-slate-100 focus:outline-none focus:border-indigo-500"
                  />
                  <p className="text-[10px] text-slate-500 mt-1">{t('cowork.chapterIndexHint')}</p>
                </div>
              </div>

              <div className="flex items-center justify-end space-x-2 pt-2">
                <button
                  onClick={() => setIsCreatingSubThread(false)}
                  className="px-3 py-1.5 rounded-lg text-xs font-medium text-slate-400 hover:text-white bg-white/4 hover:bg-white/8 transition"
                >
                  {t('common.cancel')}
                </button>
                <button
                  onClick={handleConfirmCreateSubThread}
                  className="px-4 py-1.5 rounded-lg text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-500 transition flex items-center space-x-1.5 shadow-sm"
                >
                  <Check className="w-3.5 h-3.5" />
                  <span>{t('cowork.confirmCreateSubchat')}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Pre-Translation Graph Readiness Interceptor Modal */}
      {isPreTranslateModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/60 flex items-center justify-center p-4 animate-in fade-in duration-150">
          <div className="bg-[#0c101c] border border-white/10 rounded-xl shadow-2xl max-w-md w-full overflow-hidden text-slate-200">
            {/* Header */}
            <div className="px-5 py-3.5 border-b border-white/8 flex items-center justify-between bg-linear-to-r from-amber-500/15 via-indigo-500/5 to-transparent">
              <div className="flex items-center space-x-2.5 min-w-0">
                <div className="w-7 h-7 rounded-lg bg-amber-500/20 border border-amber-500/30 flex items-center justify-center shrink-0">
                  <AlertTriangle className="w-4 h-4 text-amber-400" />
                </div>
                <div className="min-w-0">
                  <h3 className="text-xs font-semibold text-slate-100 truncate">
                    {t('cowork.preTranslationCheck')}
                  </h3>
                  <p className="text-[10px] text-slate-400 truncate">
                    {t('cowork.preTranslationSubtitle')}
                  </p>
                </div>
              </div>
              <button
                onClick={() => !isPreScanning && setIsPreTranslateModalOpen(false)}
                disabled={isPreScanning}
                className="p-1 rounded-md text-slate-400 hover:text-white hover:bg-white/8 transition disabled:opacity-40"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {/* Body */}
            <div className="p-4 space-y-3.5 text-xs">
              {/* Readiness Checklist */}
              <div className="space-y-2">
                {/* 1. Character Graph Status */}
                {chapterGraphReadiness.isGraphReady ? (
                  <div className="flex items-center space-x-2.5 p-2.5 rounded-lg bg-emerald-500/8 border border-emerald-500/20 text-emerald-300">
                    <CheckCircle2 className="w-4 h-4 shrink-0 text-emerald-400" />
                    <span className="text-[11px] leading-snug">
                      <strong>{t('cowork.graphL2Label')}</strong> {t('cowork.graphReadyCount', { count: chapterGraphReadiness.volumeRelationsCount || chapterGraphReadiness.totalRelations })}
                    </span>
                  </div>
                ) : (
                  <div className="flex items-start space-x-2.5 p-2.5 rounded-lg bg-amber-500/8 border border-amber-500/20 text-amber-200">
                    <AlertTriangle className="w-4 h-4 shrink-0 text-amber-400 mt-0.5" />
                    <div className="text-[11px] leading-snug">
                      <strong>{t('cowork.graphL2Label')}</strong> {t('cowork.graphNotScannedVolume', { volTag: currentChapterVolumeInfo.volTag ? `[${currentChapterVolumeInfo.volTag}]` : '' })}
                    </div>
                  </div>
                )}

                {/* 2. Style Scout Status */}
                {chapterGraphReadiness.isStyleReady ? (
                  <div className="flex items-center space-x-2.5 p-2.5 rounded-lg bg-emerald-500/8 border border-emerald-500/20 text-emerald-300">
                    <CheckCircle2 className="w-4 h-4 shrink-0 text-emerald-400" />
                    <span className="text-[11px] leading-snug">
                      <strong>{t('cowork.styleLabel')}</strong> {selectedProject?.style_name ? `${selectedProject.style_name} • ${t('cowork.styleReadyDesc')}` : t('cowork.styleReadyDesc')}
                    </span>
                  </div>
                ) : (
                  <div className="flex items-start space-x-2.5 p-2.5 rounded-lg bg-indigo-500/8 border border-indigo-500/20 text-indigo-200">
                    <Bot className="w-4 h-4 shrink-0 text-indigo-400 mt-0.5" />
                    <div className="text-[11px] leading-snug">
                      <strong>{t('cowork.styleAiLabel')}</strong> {t('cowork.styleNotFormedDesc')}
                    </div>
                  </div>
                )}
              </div>

              {/* Dismiss Checkbox */}
              {currentChapterVolumeInfo.volTag && (
                <label className="flex items-center space-x-2 text-[11px] text-slate-400 cursor-pointer select-none pt-1">
                  <input
                    type="checkbox"
                    checked={dontAskAgainForThisVol}
                    onChange={(e) => setDontAskAgainForThisVol(e.target.checked)}
                    className="rounded border-white/20 bg-black/40 text-indigo-500 focus:ring-0 focus:ring-offset-0 cursor-pointer"
                  />
                  <span>{t('cowork.dontAskAgainVol', { volTag: currentChapterVolumeInfo.volTag })}</span>
                </label>
              )}

              {/* Action Links */}
              <div className="flex items-center justify-between pt-1 text-[11px] border-t border-white/6">
                <button
                  type="button"
                  onClick={() => {
                    setIsPreTranslateModalOpen(false);
                    setActiveTab('graph');
                  }}
                  className="text-indigo-400 hover:text-indigo-300 flex items-center space-x-1 transition underline-offset-2 hover:underline"
                >
                  <Share2 className="w-3 h-3" />
                  <span>{t('cowork.openL2GraphTab')}</span>
                </button>

                <button
                  type="button"
                  onClick={() => {
                    setIsPreTranslateModalOpen(false);
                    setStyleScoutOpen(true);
                  }}
                  className="text-indigo-400 hover:text-indigo-300 flex items-center space-x-1 transition underline-offset-2 hover:underline"
                >
                  <Wand2 className="w-3 h-3" />
                  <span>{t('cowork.openStyleScoutModal')}</span>
                </button>
              </div>
            </div>

            {/* Modal Actions */}
            <div className="px-4 py-3 border-t border-white/8 bg-[#090d16] flex flex-col gap-2">
              <button
                onClick={handleScanAndTranslate}
                disabled={isPreScanning}
                className="w-full h-8 px-4 rounded-lg text-xs font-semibold text-white bg-linear-to-r from-indigo-600 to-indigo-500 hover:from-indigo-500 hover:to-indigo-400 transition shadow-sm border border-indigo-400/30 flex items-center justify-center space-x-2 disabled:opacity-50 active:scale-[0.98]"
              >
                {isPreScanning ? (
                  <>
                    <Loader2 className="w-3.5 h-3.5 animate-spin text-amber-300" />
                    <span>{t('cowork.scanningAndPreparing')}</span>
                  </>
                ) : (
                  <>
                    <Sparkles className="w-3.5 h-3.5 text-amber-300" />
                    <span>
                      {!chapterGraphReadiness.isGraphReady && !chapterGraphReadiness.isStyleReady
                        ? t('cowork.scanBothAndTranslate')
                        : !chapterGraphReadiness.isStyleReady
                        ? t('cowork.scanStyleAndTranslate')
                        : t('cowork.scanGraphAndTranslate')}
                    </span>
                  </>
                )}
              </button>

              <div className="flex items-center justify-end space-x-2 pt-0.5">
                <button
                  onClick={() => setIsPreTranslateModalOpen(false)}
                  disabled={isPreScanning}
                  className="px-3 py-1.5 rounded-lg text-xs text-slate-400 hover:text-white bg-transparent hover:bg-white/6 transition disabled:opacity-40"
                >
                  {t('common.cancel')}
                </button>
                <button
                  onClick={handleProceedWithoutScan}
                  disabled={isPreScanning}
                  className="px-3 py-1.5 rounded-lg text-xs font-medium text-amber-400 hover:text-amber-300 bg-amber-500/10 hover:bg-amber-500/20 border border-amber-500/20 transition disabled:opacity-40"
                  title={t('cowork.proceedWithoutScanTooltip')}
                >
                  {t('cowork.proceedWithoutScanBtn')}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Batch Translation Confirmation Modal */}
      {isBatchModalOpen && (
        <div className="fixed inset-0 z-50 bg-black/60 flex items-center justify-center p-4 animate-in fade-in duration-150">
          <div className="bg-[#0c101c] border border-white/10 rounded-2xl shadow-2xl max-w-lg w-full overflow-hidden text-slate-200 flex flex-col">
            {/* Header */}
            <div className="px-5 py-4 border-b border-white/8 flex items-center justify-between bg-linear-to-r from-indigo-500/15 via-purple-500/10 to-transparent">
              <div className="flex items-center space-x-2.5 min-w-0">
                <div className="w-8 h-8 rounded-xl bg-indigo-500/20 border border-indigo-500/30 flex items-center justify-center shrink-0">
                  <Layers className="w-4 h-4 text-indigo-400" />
                </div>
                <div className="min-w-0">
                  <h3 className="text-sm font-semibold text-slate-100 truncate">
                    {t('cowork.batchModalTitle')}
                  </h3>
                  <p className="text-xs text-slate-400 truncate">
                    {t('cowork.batchModalSubtitle')}
                  </p>
                </div>
              </div>
              <button
                type="button"
                onClick={() => setIsBatchModalOpen(false)}
                className="p-1 rounded-md text-slate-400 hover:text-white hover:bg-white/8 transition"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {/* Body */}
            <div className="p-5 space-y-4 text-xs">
              {/* Scope Switcher Pill */}
              <div className="space-y-1.5">
                <label className="text-slate-400 font-medium">
                  {t('cowork.batchTargetScope')}
                </label>
                <div className="grid grid-cols-2 gap-2 p-1 bg-[#111726] border border-white/10 rounded-lg">
                  <button
                    type="button"
                    onClick={() => {
                      setTranslateScope('volume');
                      if (!batchTargetVolume) setBatchTargetVolume(activeVolTag);
                    }}
                    className={`py-1.5 px-3 rounded-md text-xs font-semibold transition flex items-center justify-center space-x-1.5 ${
                      translateScope === 'volume'
                        ? 'bg-indigo-600 text-white shadow-xs'
                        : 'text-slate-400 hover:text-slate-200'
                    }`}
                  >
                    <BookOpen className="w-3.5 h-3.5 shrink-0" />
                    <span>{t('cowork.scopeVolume')}</span>
                  </button>
                  <button
                    type="button"
                    onClick={() => setTranslateScope('series')}
                    className={`py-1.5 px-3 rounded-md text-xs font-semibold transition flex items-center justify-center space-x-1.5 ${
                      translateScope === 'series'
                        ? 'bg-indigo-600 text-white shadow-xs'
                        : 'text-slate-400 hover:text-slate-200'
                    }`}
                  >
                    <Globe className="w-3.5 h-3.5 shrink-0" />
                    <span>{t('cowork.scopeSeries')}</span>
                  </button>
                </div>
              </div>

              {/* Volume Selection (if volume scope) */}
              {translateScope === 'volume' && volumes.length > 0 && (
                <div className="space-y-1.5">
                  <label className="text-slate-400 font-medium">
                    {t('cowork.batchSelectVolume')}
                  </label>
                  <select
                    value={effectiveBatchVolume}
                    onChange={(e) => setBatchTargetVolume(e.target.value)}
                    className="w-full h-8 px-2.5 rounded-lg bg-[#111726] border border-white/10 text-slate-200 text-xs font-medium focus:outline-none focus:border-indigo-500/50 cursor-pointer"
                  >
                    {volumes.map((vol) => {
                      const count = chapters.filter(c => c.title.startsWith(`[${vol}]`)).length;
                      return (
                        <option key={vol} value={vol}>
                          {vol} ({count} {t('projects.chapter')})
                        </option>
                      );
                    })}
                  </select>
                </div>
              )}

              {/* Statistics Breakdown */}
              <div className="grid grid-cols-3 gap-2.5 p-3 rounded-xl bg-[#111726] border border-white/8">
                <div className="text-center">
                  <div className="text-[11px] text-slate-400 font-medium">{t('cowork.batchTotalChapters')}</div>
                  <div className="text-base font-bold font-mono text-slate-100 mt-0.5">
                    {batchStats.total}
                  </div>
                </div>
                <div className="text-center border-x border-white/8">
                  <div className="text-[11px] text-emerald-400/90 font-medium">{t('cowork.batchCompletedChapters')}</div>
                  <div className="text-base font-bold font-mono text-emerald-400 mt-0.5">
                    {batchStats.completed}
                  </div>
                </div>
                <div className="text-center">
                  <div className="text-[11px] text-amber-400/90 font-medium">{t('cowork.batchPendingChapters')}</div>
                  <div className="text-base font-bold font-mono text-amber-400 mt-0.5">
                    {batchStats.pending}
                  </div>
                </div>
              </div>

              {/* Skip Completed Checkbox */}
              <label className="flex items-start space-x-2.5 p-2.5 rounded-lg bg-white/2 border border-white/6 hover:bg-white/4 cursor-pointer transition select-none">
                <input
                  type="checkbox"
                  checked={batchSkipCompleted}
                  onChange={(e) => setBatchSkipCompleted(e.target.checked)}
                  className="rounded border-white/20 bg-black/40 text-indigo-500 focus:ring-0 focus:ring-offset-0 cursor-pointer mt-0.5"
                />
                <div className="space-y-0.5">
                  <div className="font-semibold text-slate-200">
                    {t('cowork.batchSkipCompleted')}
                  </div>
                  <div className="text-[11px] text-slate-400">
                    {t('cowork.batchSkipCompletedDesc')}
                  </div>
                </div>
              </label>

              {/* Warning if all completed and skipping */}
              {batchStats.total > 0 && batchStats.pending === 0 && batchSkipCompleted && (
                <div className="p-3 rounded-lg bg-amber-500/10 border border-amber-500/20 text-amber-300 flex items-start space-x-2 text-[11px] leading-relaxed">
                  <AlertTriangle className="w-4 h-4 shrink-0 text-amber-400 mt-0.5" />
                  <span>{t('cowork.batchAllCompletedWarning')}</span>
                </div>
              )}

              {/* Live Preview Information Callout */}
              <div className="p-3 rounded-lg bg-indigo-500/10 border border-indigo-500/20 text-indigo-200 flex items-start space-x-2 text-[11px] leading-relaxed">
                <Zap className="w-4 h-4 shrink-0 text-amber-400 mt-0.5" />
                <span>{t('cowork.batchLivePreviewNotice')}</span>
              </div>
            </div>

            {/* Footer */}
            <div className="px-5 py-3.5 border-t border-white/8 flex items-center justify-end space-x-2 bg-[#090d16]">
              <button
                type="button"
                onClick={() => setIsBatchModalOpen(false)}
                className="px-3.5 py-1.5 rounded-lg text-xs text-slate-400 hover:text-white bg-transparent hover:bg-white/6 transition"
              >
                {t('common.cancel')}
              </button>
              <button
                type="button"
                onClick={handleConfirmStartBatch}
                disabled={batchStats.total === 0 || (batchStats.pending === 0 && batchSkipCompleted)}
                className="h-8 px-4 rounded-lg text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 disabled:cursor-not-allowed transition shadow-sm border border-indigo-400/30 flex items-center space-x-1.5 active:scale-[0.98]"
              >
                <Play className="w-3.5 h-3.5 fill-current shrink-0" />
                <span>{t('cowork.batchStartBtn')}</span>
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
