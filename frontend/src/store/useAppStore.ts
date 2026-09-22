import { create } from 'zustand';
import * as dtos from '@bindings/novelclaw/internal/dtos/models';
import * as ProjectService from '@bindings/novelclaw/internal/services/projectservice';
import * as SoulService from '@bindings/novelclaw/internal/services/soulservice';
import * as TranslationService from '@bindings/novelclaw/internal/services/translationservice';
import * as GraphService from '@bindings/novelclaw/internal/services/graphservice';
import * as LLMService from '@bindings/novelclaw/internal/services/llmservice';
import * as GlossaryService from '@bindings/novelclaw/internal/services/glossaryservice';
import * as NovelClawService from '@bindings/novelclaw/internal/services/novelclawservice';
import * as SkillService from '@bindings/novelclaw/internal/services/skillservice';
import * as WorldBibleService from '@bindings/novelclaw/internal/services/worldbibleservice';
import * as AppMetaService from '@bindings/novelclaw/pkg/appmeta/updateservice';
import * as appmetaModels from '@bindings/novelclaw/pkg/appmeta/models';
import i18n from '@/i18n';

export interface BatchTranslationProgress {
  scope: 'volume' | 'series';
  totalChapters: number;
  completedChapters: number;
  currentChapterIndex: number;
  currentChapterTitle: string;
  volumeName?: string;
  isStopping?: boolean;
}

export interface FallbackTier {
  id: string;
  name: string;
  configId: string;
  providerName: string;
  modelName: string;
  maxRetries: number;
  timeoutSeconds: number;
  triggerCondition: '429_5xx_timeout' | '429_only' | 'any_error';
  enabled: boolean;
  reasoningEffort?: 'off' | 'low' | 'medium' | 'high' | 'xhigh' | 'max';
}

interface AppState {
  // App Lifecycle
  isAppLoading: boolean;
  isFirstRun: boolean;
  loadingMessage: string;
  setAppLoading: (loading: boolean) => void;
  setFirstRun: (firstRun: boolean) => void;
  setLoadingMessage: (msg: string) => void;
  completeSetupWizard: () => void;
  initializeApp: () => Promise<void>;

  // App info & self-update
  appInfo: appmetaModels.AppInfo | null;
  fetchAppInfo: () => Promise<void>;
  checkForUpdates: () => Promise<void>;
  checkForUpdatesSilent: () => Promise<void>;

  // Navigation & Modals
  activeTab: 'workspace' | 'graph' | 'glossary' | 'world_bible' | 'benchmark';
  isStyleScoutOpen: boolean;
  isExportOpen: boolean;
  isImportOpen: boolean;
  isSettingsOpen: boolean;
  setActiveTab: (tab: 'workspace' | 'graph' | 'glossary' | 'world_bible' | 'benchmark') => void;
  setStyleScoutOpen: (open: boolean) => void;
  setExportOpen: (open: boolean) => void;
  setImportOpen: (open: boolean) => void;
  setSettingsOpen: (open: boolean) => void;

  // Settings & LLM Configs
  llmConfigs: dtos.LLMConfigDTO[];
  defaultLLMConfig: dtos.LLMConfigDTO | null;
  fallbackChain: FallbackTier[];
  thinkingEffort: 'off' | 'low' | 'medium' | 'high' | 'xhigh' | 'max';
  setThinkingEffort: (effort: 'off' | 'low' | 'medium' | 'high' | 'xhigh' | 'max') => void;
  setFallbackChain: (chain: FallbackTier[]) => void;
  reconcileFallbackChain: (configs: dtos.LLMConfigDTO[]) => void;
  fetchLLMConfigs: () => Promise<void>;
  saveLLMConfig: (req: dtos.SaveLLMConfigRequest) => Promise<void>;
  deleteLLMConfig: (id: string) => Promise<void>;
  setDefaultLLMConfig: (id: string) => Promise<void>;
  testLLMConnection: (req: dtos.TestConnectionRequest) => Promise<dtos.TestConnectionResponse | null>;
  batchImportBooks: (req: dtos.BatchImportBookRequest) => Promise<dtos.BatchImportResultDTO | null>;
  previewStyle: (req: dtos.PreviewStyleRequest) => Promise<dtos.PreviewStyleResponse | null>;
  autoScoutStyle: (req: dtos.AutoScoutStyleRequest) => Promise<dtos.AutoScoutStyleResponse | null>;

  // Projects & Chapters
  isProjectManagerOpen: boolean;
  setProjectManagerOpen: (open: boolean) => void;
  projects: dtos.ProjectDTO[];
  selectedProject: dtos.ProjectDTO | null;
  chapters: dtos.ChapterDTO[];
  selectedChapter: dtos.ChapterDTO | null;
  fetchProjects: () => Promise<void>;
  selectProject: (p: dtos.ProjectDTO) => void;
  fetchChapters: () => Promise<void>;
  selectChapter: (ch: dtos.ChapterDTO) => void;
  updateChapterTranslation: (id: string, text: string) => Promise<void>;
  updateProjectLanguages: (sourceLang: string, targetLang: string) => Promise<void>;
  updateProjectStyle: (projectId: string, styleName: string, styleGuide: string) => Promise<void>;
  deleteProject: (id: string) => Promise<void>;
  renameProject: (id: string, newTitle: string) => Promise<void>;
  inheritKnowledge: (sourceProjectID: string, targetProjectID: string) => Promise<dtos.InheritKnowledgeResultDTO | null>;
  rollbackChapter: () => Promise<void>;
  appendBookToProject: (projectId: string, filePath: string, volName?: string) => Promise<dtos.ProjectDTO | null>;
  selectedVolume: string;
  setSelectedVolume: (volume: string) => void;

  // Coworker Workspace & Translation
  isTranslating: boolean;
  telemetry: dtos.ExecutionTelemetryDTO | null;
  thoughtText: string;
  isThoughtOpen: boolean;
  activeTool: string | null;
  translationMode: string;
  enableR19: boolean;
  enableHotPatch: boolean;
  enableAgenticRAG: boolean;
  maxToolIterations: number;
  decisionMode: string;
  styleGuide: string;
  setThoughtOpen: (open: boolean) => void;
  setTranslationMode: (mode: string) => void;
  setEnableR19: (enabled: boolean) => void;
  setEnableHotPatch: (enabled: boolean) => void;
  setEnableAgenticRAG: (enabled: boolean) => void;
  setMaxToolIterations: (iterations: number) => void;
  setDecisionMode: (mode: string) => void;
  setStyleGuide: (guide: string) => void;
  startTranslation: (overrideChapter?: dtos.ChapterDTO) => Promise<dtos.ExecutionTelemetryDTO | null>;
  triggerSoftStop: () => Promise<void>;
  triggerHardAbort: () => Promise<void>;

  // Batch Translation
  batchProgress: BatchTranslationProgress | null;
  stopBatchRequested: boolean;
  setStopBatchRequested: (requested: boolean) => void;
  startBatchTranslation: (options: {
    scope: 'volume' | 'series';
    targetVolume?: string;
    skipCompleted?: boolean;
  }) => Promise<void>;
  stopBatchTranslation: () => Promise<void>;

  // Soul & NovelClaw Companion Chat
  soul: dtos.SoulDTO | null;
  availableSouls: dtos.SoulDTO[];
  fetchSoul: () => Promise<void>;
  fetchAvailableSouls: () => Promise<void>;
  setActiveSoul: (soulId: string, projectId?: string) => Promise<void>;
  saveCustomSoul: (soul: dtos.SoulDTO) => Promise<void>;
  deleteCustomSoul: (soulId: string) => Promise<void>;
  importSoul: (content: string) => Promise<dtos.SoulDTO | null>;
  exportSoul: (soulId: string) => Promise<string>;
  restoreDefaultSouls: () => Promise<void>;
  pendingAskHuman: {
    jobId?: string;
    question: string;
    options?: string[];
    contextSnippet?: string;
    timeoutSeconds?: number;
    createdAt?: number;
  } | null;
  setPendingAskHuman: (data: { jobId?: string; question: string; options?: string[]; contextSnippet?: string; timeoutSeconds?: number; createdAt?: number } | null) => void;
  resolveAskHuman: (jobId: string, answer: string) => Promise<boolean>;

  // Character Graph (L2)
  entities: dtos.EntityDTO[];
  relations: dtos.RelationDTO[];
  isGraphLoading: boolean;
  fetchGraphData: (chapterIndex?: number) => Promise<void>;
  upsertRelation: (req: dtos.UpsertRelationRequest) => Promise<void>;
  toggleLockRelation: (id: string) => Promise<void>;
  deleteRelation: (id: string) => Promise<void>;
  upsertEntity: (req: dtos.UpsertEntityRequest) => Promise<void>;
  deleteEntity: (id: string) => Promise<void>;
  autoScanEntities: (chapterIndex?: number, scanMode?: 'volume' | 'current_chapter' | 'chapter' | 'all' | 'pre_scan', volume?: string) => Promise<dtos.AutoScanResultDTO | null>;
  checkGraphReadiness: (chapterIndex: number) => Promise<dtos.GraphReadinessDTO | null>;

  // Glossary (L3)
  terms: dtos.GlossaryTermDTO[];
  isGlossaryLoading: boolean;
  fetchGlossaryTerms: () => Promise<void>;
  upsertGlossaryTerm: (req: dtos.UpsertGlossaryTermRequest) => Promise<void>;
  deleteGlossaryTerm: (id: string) => Promise<void>;
  importGlossaryTSV: (content: string) => Promise<number>;
  exportGlossaryTSV: () => Promise<string>;
  importGlossaryJSON: (content: string) => Promise<number>;
  exportGlossaryJSON: () => Promise<string>;
  importGlossaryPlainText: (content: string, defaultCategory?: string) => Promise<number>;
  listRemoteGlossarySources: () => Promise<dtos.RemoteGlossarySourceDTO[]>;
  downloadAndImportGlossary: (url: string, format?: string, defaultCategory?: string) => Promise<number>;

  // Benchmark
  benchmarkReport: dtos.BenchmarkReportDTO | null;
  isBenchmarking: boolean;
  runBenchmark: (modes: string[], chapterIndices: number[]) => Promise<dtos.BenchmarkReportDTO | null>;

  // NovelClaw Agentic Chat & Threads
  novelclawThreads: dtos.NovelClawThreadDTO[];
  selectedThread: dtos.NovelClawThreadDTO | null;
  novelclawMessages: dtos.NovelClawMessageDTO[];
  isNovelClawThinking: boolean;
  novelclawThinkingText: string;
  novelclawActiveTool: string | null;
  autoCompactThreshold: number;
  fetchNovelClawThreads: (projectID?: string) => Promise<void>;
  selectNovelClawThread: (thread: dtos.NovelClawThreadDTO) => Promise<void>;
  createSubThread: (req: dtos.CreateSubThreadRequest) => Promise<dtos.NovelClawThreadDTO | null>;
  deleteNovelClawThread: (threadID: string) => Promise<void>;
  fetchNovelClawMessages: (threadID?: string) => Promise<void>;
  sendNovelClawMessage: (content: string) => Promise<void>;
  clearNovelClawThread: () => Promise<void>;
  compactNovelClawThread: () => Promise<void>;
  setAutoCompactThreshold: (tokens: number) => Promise<void>;
  setNovelClawThinkingText: (text: string) => void;
  setNovelClawActiveTool: (tool: string | null) => void;
  setIsNovelClawThinking: (thinking: boolean) => void;

  // Modular Skills & Capabilities (Stage 9)
  skills: dtos.SkillDTO[];
  skillPresets: dtos.SkillPresetDTO[];
  activeSkillsTokenWeight: number;
  isSkillsLoading: boolean;
  fetchSkills: (projectID?: string) => Promise<void>;
  fetchSkillPresets: () => Promise<void>;
  toggleSkill: (skillID: string, enabled: boolean, projectID?: string) => Promise<void>;
  updateSkillCustomRules: (skillID: string, customRules: string, projectID?: string) => Promise<void>;
  resetSkillToDefault: (skillID: string, projectID?: string) => Promise<void>;
  applyCapabilityPreset: (presetName: string, projectID?: string) => Promise<void>;
  fetchActiveSkillsTokenWeight: (projectID?: string) => Promise<void>;

  // Universal World Bible & Lorebook (Stage 10)
  worldCategories: dtos.WorldCategoryDTO[];
  worldEntries: dtos.WorldEntryDTO[];
  selectedWorldCategory: string | null;
  isWorldBibleLoading: boolean;
  isScanningWorld: boolean;
  worldScanReport: dtos.ScanWorldResponse | null;
  fetchWorldCategories: (projectID?: string) => Promise<void>;
  fetchWorldEntries: (projectID?: string, categoryID?: string) => Promise<void>;
  setSelectedWorldCategory: (catID: string | null) => void;
  upsertWorldCategory: (req: dtos.UpsertWorldCategoryRequest) => Promise<dtos.WorldCategoryDTO | null>;
  deleteWorldCategory: (id: string) => Promise<void>;
  upsertWorldEntry: (req: dtos.UpsertWorldEntryRequest) => Promise<dtos.WorldEntryDTO | null>;
  verifyWorldEntry: (id: string, verified: boolean) => Promise<void>;
  deleteWorldEntry: (id: string) => Promise<void>;
  scanWorldBible: (startChap: number, endChap: number, projectID?: string) => Promise<dtos.ScanWorldResponse | null>;
  exportWorldMarkdown: (projectID?: string) => Promise<string>;
  exportWorldJSON: (projectID?: string) => Promise<string>;
  importWorldJSON: (jsonStr: string, projectID?: string) => Promise<[number, number]>;
  resetSkillsToDefault: (projectID?: string) => Promise<void>;
  learnFromEdits: (chapterIndex: number, notes: string, projectID?: string) => Promise<dtos.LearnResultDTO | null>;
}

// The fallback chain is derived from the providers stored in the database
// (see reconcileFallbackChain). It is deliberately NOT seeded from
// localStorage: a stale blob could inject phantom tiers that do not
// correspond to any saved provider.
const getInitialFallbackChain = (): FallbackTier[] => [];

export const useAppStore = create<AppState>((set, get) => ({
  // App Lifecycle
  isAppLoading: true,
  appInfo: null,
  isFirstRun: false,
  loadingMessage: '',
  setAppLoading: (loading) => set({ isAppLoading: loading }),
  setFirstRun: (firstRun) => set({ isFirstRun: firstRun }),
  setLoadingMessage: (msg) => set({ loadingMessage: msg }),
  completeSetupWizard: () => {
    localStorage.setItem('neko_setup_complete', 'true');
    set({ isFirstRun: false });
  },

  // App info & self-update
  fetchAppInfo: async () => {
    try {
      const info = await AppMetaService.GetAppInfo();
      set({ appInfo: info });
    } catch (err) {
      console.error('Failed to fetch app info:', err);
    }
  },
  // Manual check (About tab button): the built-in updater window shows
  // every state — checking, up-to-date, update available and errors.
  checkForUpdates: async () => {
    try {
      await AppMetaService.CheckForUpdates();
    } catch (err) {
      console.error('Check for updates failed:', err);
      throw err;
    }
  },
  // Automatic startup check: silent on error / up-to-date; the updater
  // window only opens when a newer release is actually found.
  checkForUpdatesSilent: async () => {
    try {
      await AppMetaService.CheckSilent();
    } catch (err) {
      // Intentionally silent — startup checks must never disturb the user.
      console.debug('Silent update check skipped:', err);
    }
  },
  initializeApp: async () => {
    set({ isAppLoading: true, loadingMessage: 'loadingScreen.initializing' });
    try {
      // Check if setup was previously completed
      const setupComplete = localStorage.getItem('neko_setup_complete') === 'true';

      set({ loadingMessage: 'loadingScreen.loadingConfigs' });
      await get().fetchLLMConfigs();
      await get().fetchSoul();
      await get().fetchAvailableSouls();
      await get().fetchSkillPresets();

      set({ loadingMessage: 'loadingScreen.loadingProjects' });
      await get().fetchProjects();
      const projId = get().selectedProject?.id;
      if (projId) {
        await get().fetchSkills(projId);
      }

      // Determine if first run: no setup complete flag AND no LLM configs
      const configs = get().llmConfigs;
      const isFirstRun = !setupComplete && (!configs || configs.length === 0);

      // Drop any legacy localStorage fallback blob: the chain is now
      // rebuilt from the DB by fetchLLMConfigs → reconcileFallbackChain.
      try {
        localStorage.removeItem('neko_fallback_chain');
      } catch {
        // ignore
      }

      // Version/platform metadata for the About tab & loading screen footer
      await get().fetchAppInfo();

      set({ loadingMessage: 'loadingScreen.almostReady' });
      // Small delay for visual smoothness
      await new Promise(resolve => setTimeout(resolve, 400));

      set({ isAppLoading: false, isFirstRun });

      // Automatic self-update check once the app has opened: silent on
      // errors and when already up to date; shows the updater window only
      // when a newer release exists.
      void get().checkForUpdatesSilent();
    } catch (err) {
      console.error('App initialization failed:', err);
      set({ isAppLoading: false, isFirstRun: false });
    }
  },

  // Navigation & Modals
  activeTab: 'workspace',
  isStyleScoutOpen: false,
  isExportOpen: false,
  isImportOpen: false,
  isSettingsOpen: false,
  isProjectManagerOpen: false,
  setActiveTab: (tab) => set({ activeTab: tab }),
  setStyleScoutOpen: (open) => set({ isStyleScoutOpen: open }),
  setExportOpen: (open) => set({ isExportOpen: open }),
  setImportOpen: (open) => set({ isImportOpen: open }),
  setSettingsOpen: (open) => set({ isSettingsOpen: open }),
  setProjectManagerOpen: (open) => set({ isProjectManagerOpen: open }),

  // Projects & Chapters
  projects: [],
  selectedProject: null,
  chapters: [],
  selectedChapter: null,

  fetchProjects: async () => {
    try {
      const list = await ProjectService.ListProjects();
      const currentSelected = get().selectedProject;
      if (list && list.length > 0) {
        const stillExists = currentSelected && list.some(p => p.id === currentSelected.id);
        const nextSelected = stillExists ? (list.find(p => p.id === currentSelected.id) || currentSelected) : list[0];
        set({
          projects: list,
          selectedProject: nextSelected,
          styleGuide: nextSelected?.style_guide || '',
        });
        get().fetchChapters();
        get().fetchGraphData();
        get().fetchNovelClawThreads();
        get().fetchSkills(nextSelected.id);
        get().fetchWorldCategories(nextSelected.id);
        get().fetchWorldEntries(nextSelected.id);
      } else {
        set({ projects: [], selectedProject: null, chapters: [], selectedChapter: null, skills: [], worldCategories: [], worldEntries: [], styleGuide: '' });
      }
    } catch (err) {
      console.error('Failed to fetch projects:', err);
    }
  },

  selectProject: (p) => {
    set({
      selectedProject: p,
      selectedChapter: null,
      chapters: [],
      styleGuide: p?.style_guide || '',
    });
    get().fetchChapters();
    get().fetchGraphData();
    get().fetchGlossaryTerms();
    get().fetchNovelClawThreads();
    get().fetchSkills(p.id);
    get().fetchWorldCategories(p.id);
    get().fetchWorldEntries(p.id);
  },

  fetchChapters: async () => {
    const proj = get().selectedProject;
    if (!proj) {
      set({ chapters: [], selectedChapter: null });
      return;
    }
    try {
      const chs = await ProjectService.ListChapters(proj.id);
      const currentCh = get().selectedChapter;
      if (chs && chs.length > 0) {
        const stillSelected = currentCh && chs.find(c => c.id === currentCh.id);
        set({ chapters: chs, selectedChapter: stillSelected || chs[0] });
      } else {
        set({ chapters: [], selectedChapter: null });
      }
    } catch (err) {
      console.error('Failed to fetch chapters:', err);
    }
  },

  selectChapter: (ch) => set({ selectedChapter: ch }),

  updateChapterTranslation: async (id, text) => {
    try {
      await ProjectService.UpdateChapterTranslation({
        id,
        translated_content: text,
        status: 'completed',
      });
      await get().fetchChapters();
    } catch (err) {
      console.error('Failed to update chapter translation:', err);
    }
  },

  updateProjectLanguages: async (sourceLang: string, targetLang: string) => {
    const proj = get().selectedProject;
    if (!proj) return;
    try {
      const updated = await ProjectService.UpdateProjectLanguages({
        id: proj.id,
        source_lang: sourceLang,
        target_lang: targetLang,
      });
      if (updated) {
        set({
          selectedProject: updated,
          projects: get().projects.map(p => p.id === updated.id ? updated : p),
        });
      }
    } catch (err) {
      console.error('Failed to update project languages:', err);
    }
  },

  updateProjectStyle: async (projectId: string, styleName: string, styleGuide: string) => {
    try {
      const updated = await ProjectService.UpdateProjectStyle({
        id: projectId,
        style_name: styleName,
        style_guide: styleGuide,
      });
      if (updated) {
        set((state) => ({
          styleGuide: updated.style_guide || styleGuide,
          selectedProject: state.selectedProject?.id === projectId ? updated : state.selectedProject,
          projects: state.projects.map(p => p.id === projectId ? updated : p),
        }));
      }
    } catch (err) {
      console.error('Failed to update project style:', err);
    }
  },

  deleteProject: async (id: string) => {
    try {
      await ProjectService.DeleteProject(id);
      const remaining = get().projects.filter(p => p.id !== id);
      const wasSelected = get().selectedProject?.id === id;
      const nextSelected = wasSelected ? (remaining.length > 0 ? remaining[0] : null) : get().selectedProject;

      set({
        projects: remaining,
        selectedProject: nextSelected,
      });

      if (nextSelected) {
        await get().fetchChapters();
        await get().fetchGraphData();
      } else {
        set({ chapters: [], selectedChapter: null, entities: [], relations: [] });
      }
    } catch (err) {
      console.error('Failed to delete project:', err);
      throw err;
    }
  },

  renameProject: async (id: string, newTitle: string) => {
    try {
      const updated = await ProjectService.RenameProject({
        id,
        title: newTitle,
      });
      if (updated) {
        set({
          projects: get().projects.map(p => p.id === id ? updated : p),
          selectedProject: get().selectedProject?.id === id ? updated : get().selectedProject,
        });
      }
    } catch (err) {
      console.error('Failed to rename project:', err);
      throw err;
    }
  },

  inheritKnowledge: async (sourceProjectID, targetProjectID) => {
    try {
      const res = await ProjectService.InheritKnowledge({
        source_project_id: sourceProjectID,
        target_project_id: targetProjectID,
      });
      await get().fetchGraphData();
      await get().fetchGlossaryTerms();
      return res;
    } catch (err) {
      console.error('Failed to inherit knowledge:', err);
      throw err;
    }
  },

  rollbackChapter: async () => {
    const proj = get().selectedProject;
    const ch = get().selectedChapter;
    if (!proj || !ch) return;
    try {
      const restored = await ProjectService.RollbackChapter(proj.id, ch.chapter_index);
      if (restored) {
        set({ selectedChapter: restored });
        await get().fetchChapters();
      }
    } catch (err) {
      console.error('Failed to rollback chapter:', err);
      throw err;
    }
  },

  appendBookToProject: async (projectId: string, filePath: string, volName?: string) => {
    try {
      const updated = await ProjectService.AppendBookToProject({
        project_id: projectId,
        file_path: filePath,
        vol_name: volName || '',
      });
      if (updated) {
        await get().fetchProjects();
        if (get().selectedProject?.id === projectId) {
          get().selectProject(updated);
          await get().fetchChapters();
        }
      }
      return updated;
    } catch (err) {
      console.error('Failed to append book to project:', err);
      throw err;
    }
  },

  // Coworker Workspace & Translation
  selectedVolume: 'all',
  setSelectedVolume: (volume) => set({ selectedVolume: volume }),
  isTranslating: false,
  telemetry: null,
  thoughtText: '',
  isThoughtOpen: true,
  activeTool: null,
  translationMode: 'concurrent_dual_agent',
  enableR19: true,
  enableHotPatch: true,
  enableAgenticRAG: true,
  maxToolIterations: 2,
  decisionMode: 'manual',
  styleGuide: '',

  setThoughtOpen: (open) => set({ isThoughtOpen: open }),
  setTranslationMode: (mode) => set({ translationMode: mode }),
  setEnableR19: (enabled) => set({ enableR19: enabled }),
  setEnableHotPatch: (enabled) => set({ enableHotPatch: enabled }),
  setEnableAgenticRAG: (enabled) => set({ enableAgenticRAG: enabled }),
  setMaxToolIterations: (iterations) => set({ maxToolIterations: iterations }),
  setDecisionMode: (mode) => set({ decisionMode: mode }),
  setStyleGuide: (guide) => {
    set({ styleGuide: guide });
    const { selectedProject } = get();
    if (selectedProject) {
      get().updateProjectStyle(selectedProject.id, selectedProject.style_name || 'Project Style', guide);
    }
  },

  batchProgress: null,
  stopBatchRequested: false,
  setStopBatchRequested: (requested) => set({ stopBatchRequested: requested }),

  startTranslation: async (overrideChapter?: dtos.ChapterDTO) => {
    const { selectedProject, translationMode, enableHotPatch, enableR19, enableAgenticRAG, maxToolIterations, decisionMode, styleGuide, defaultLLMConfig } = get();
    const chapterToTranslate = overrideChapter || get().selectedChapter;
    if (!selectedProject || !chapterToTranslate) return null;

    const activeThreadId = get().selectedThread?.id || 'main';
    const startMsg: any = {
      id: `trans_start_${Date.now()}`,
      thread_id: activeThreadId,
      project_id: selectedProject.id,
      sender: 'novelclaw',
      role: 'assistant',
      content: `**${i18n.t('cowork.translationStartingMsg', {
        defaultValue: 'Started translating Chapter {{index}}: {{title}}',
        index: chapterToTranslate.chapter_index,
        title: chapterToTranslate.title || i18n.t('cowork.untitledChapter', 'Untitled Chapter'),
      })}**\n*${i18n.t('cowork.modeLabel', 'Mode')}: ${translationMode} | ${i18n.t('cowork.agenticRagLabel', 'Agentic RAG')}: ${enableAgenticRAG ? i18n.t('common.yes', 'Yes') : i18n.t('common.no', 'No')} | Hot-patching: ${enableHotPatch ? i18n.t('common.on', 'On') : i18n.t('common.off', 'Off')}*`,
      thinking_content: '',
      step_type: 'step_status',
      step_status: 'running',
      action_call_json: '',
      is_collapsed: false,
      is_archived_compact: false,
      token_count: 0,
      created_at: new Date().toISOString(),
    };

    set((state) => ({
      isTranslating: true,
      isNovelClawThinking: true,
      novelclawThinkingText: i18n.t('cowork.preparingContextThought', 'Preparing context and activating Agentic RAG loop...'),
      novelclawActiveTool: 'Agentic RAG Engine',
      thoughtText: i18n.t('cowork.preparingContextThought', 'Preparing context and activating Agentic RAG loop...'),
      selectedChapter: {
        ...chapterToTranslate,
        translated_content: '',
      },
      novelclawMessages: [...state.novelclawMessages, startMsg],
    }));

    try {
      const activeFallbackChain = get().fallbackChain;
      const primaryTier = activeFallbackChain.find((t) => t.enabled);
      const activeModel = primaryTier?.modelName || defaultLLMConfig?.model_name || '';
      const maxRetries = primaryTier?.maxRetries || defaultLLMConfig?.max_retries || 3;
      const timeoutSeconds = primaryTier?.timeoutSeconds || defaultLLMConfig?.timeout_seconds || 120;
      const activeThinkingEffort = primaryTier?.reasoningEffort || (defaultLLMConfig?.reasoning_effort as any) || get().thinkingEffort || 'off';

      const opts: dtos.TranslationOptionsDTO = {
        mode: translationMode,
        concurrency: 3,
        critic_model: activeModel,
        polish_model: activeModel,
        timeout_seconds: timeoutSeconds,
        max_retries: maxRetries,
        enable_hot_patch: enableHotPatch,
        enable_r19: enableR19,
        enable_agentic_rag: enableAgenticRAG,
        max_tool_iterations: maxToolIterations,
        decision_mode: decisionMode || 'manual',
        style_guide: styleGuide,
        thinking_effort: activeThinkingEffort,
      };

      const result = await TranslationService.TranslateChapter({
        project_id: selectedProject.id,
        chapter_index: chapterToTranslate.chapter_index,
        options: opts,
      });

      if (result) {
        const tokPerSec = result.tokens_per_second || (result.duration_ms ? (result.total_tokens / (result.duration_ms / 1000)) : 0);
        const speedText = tokPerSec > 0
          ? `**${tokPerSec.toFixed(1)} tok/s** (${result.runes_per_second?.toFixed(1) || 0} r/s)`
          : `**${result.runes_per_second?.toFixed(1) || 0} r/s**`;

        const summaryMsg: any = {
          id: `trans_done_${Date.now()}`,
          thread_id: activeThreadId,
          project_id: selectedProject.id,
          sender: 'novelclaw',
          role: 'assistant',
          content: `**${i18n.t('cowork.translationCompletedMsg', {
            defaultValue: 'Completed translating Chapter {{index}}!',
            index: chapterToTranslate.chapter_index,
          })}**\n- ${i18n.t('cowork.translationSpeed', 'Speed')}: ${speedText}\n- ${i18n.t('cowork.wordCount', 'Words')}: ${result.total_runes?.toLocaleString() || 0} runes\n- ${i18n.t('cowork.shadowCriticRevisions', 'Shadow Critic Revisions')}: ${result.revised_count || 0}\n- ${i18n.t('cowork.toolsCalled', 'Agentic Tools Called')}: ${result.tool_calls_count || 0}\n- ${i18n.t('cowork.tokensUsed', 'Tokens Used')}: ${result.total_tokens?.toLocaleString() || 0}`,
          step_type: 'translation_summary',
          step_status: 'completed',
          action_call_json: JSON.stringify(result),
          thinking_content: get().novelclawThinkingText || get().thoughtText || '',
          is_collapsed: false,
          is_archived_compact: false,
          token_count: result.total_tokens || 0,
          created_at: new Date().toISOString(),
        };

        set((state) => ({
          telemetry: result,
          isNovelClawThinking: false,
          novelclawActiveTool: null,
          thoughtText: i18n.t('cowork.translationDoneThought', {
            defaultValue: 'Translation completed with {{count}} revisions from Shadow Critic.',
            count: result.revised_count,
          }),
          novelclawMessages: [...state.novelclawMessages, summaryMsg],
        }));
      }

      await get().fetchChapters();
      await get().fetchSoul();
      return result || null;
    } catch (err: any) {
      console.error('Translation execution error:', err);
      const errMsg: any = {
        id: `trans_err_${Date.now()}`,
        thread_id: activeThreadId,
        project_id: selectedProject.id,
        sender: 'novelclaw',
        role: 'assistant',
        content: `**${i18n.t('cowork.translationError', 'Translation error')}:** ${err?.message || err}`,
        step_type: 'step_status',
        step_status: 'error',
        action_call_json: '',
        thinking_content: '',
        is_collapsed: false,
        is_archived_compact: false,
        token_count: 0,
        created_at: new Date().toISOString(),
      };
      set((state) => ({
        thoughtText: `${i18n.t('cowork.translationError', 'Translation error')}: ${err?.message || err}`,
        isNovelClawThinking: false,
        novelclawActiveTool: null,
        novelclawMessages: [...state.novelclawMessages, errMsg],
      }));
      return null;
    } finally {
      set({ isTranslating: false, isNovelClawThinking: false, novelclawActiveTool: null });
    }
  },

  startBatchTranslation: async ({ scope, targetVolume, skipCompleted = true }) => {
    const { selectedProject, chapters } = get();
    if (!selectedProject || !chapters || chapters.length === 0) return;

    let targetChapters = [...chapters].sort((a, b) => a.chapter_index - b.chapter_index);
    if (scope === 'volume') {
      const vol = targetVolume || get().selectedVolume;
      if (vol && vol !== 'all') {
        targetChapters = targetChapters.filter(c => c.title.startsWith(`[${vol}]`));
      }
    }

    if (skipCompleted) {
      targetChapters = targetChapters.filter(c => {
        const isDone = c.status === 'completed' && Boolean(c.translated_content && c.translated_content.trim().length > 0);
        return !isDone;
      });
    }

    if (targetChapters.length === 0) {
      return;
    }

    set({
      stopBatchRequested: false,
      batchProgress: {
        scope,
        totalChapters: targetChapters.length,
        completedChapters: 0,
        currentChapterIndex: targetChapters[0].chapter_index,
        currentChapterTitle: targetChapters[0].title,
        volumeName: targetVolume,
        isStopping: false,
      },
    });

    try {
      for (let i = 0; i < targetChapters.length; i++) {
        if (get().stopBatchRequested) {
          break;
        }
        const ch = targetChapters[i];

        // 1. Automatically switch chapter view for real-time live preview!
        get().selectChapter(ch);

        set(state => ({
          batchProgress: state.batchProgress ? {
            ...state.batchProgress,
            currentChapterIndex: ch.chapter_index,
            currentChapterTitle: ch.title,
            completedChapters: i,
          } : null,
        }));

        // Allow brief interval for UI DOM and refs to sync cleanly
        await new Promise(r => setTimeout(r, 200));

        // 2. Translate current chapter
        const result = await get().startTranslation(ch);

        if (get().stopBatchRequested || !result) {
          break;
        }

        set(state => ({
          batchProgress: state.batchProgress ? {
            ...state.batchProgress,
            completedChapters: i + 1,
          } : null,
        }));
      }
    } finally {
      set({
        batchProgress: null,
        stopBatchRequested: false,
      });
      await get().fetchChapters();
    }
  },

  stopBatchTranslation: async () => {
    set(state => ({
      stopBatchRequested: true,
      batchProgress: state.batchProgress ? { ...state.batchProgress, isStopping: true } : null,
    }));
    await get().triggerSoftStop();
  },

  triggerSoftStop: async () => {
    const { selectedProject, selectedChapter, batchProgress, soul } = get();
    // Always flag batch to stop immediately
    set(state => ({
      stopBatchRequested: true,
      batchProgress: state.batchProgress ? { ...state.batchProgress, isStopping: true } : null,
    }));

    if (!selectedProject) return;
    const currentChapterIdx = batchProgress?.currentChapterIndex ?? selectedChapter?.chapter_index;
    const jobId = currentChapterIdx != null ? `job_${selectedProject.id}_${currentChapterIdx}` : '';

    await SoulService.TriggerSteerAction({
      job_id: jobId,
      project_id: selectedProject.id,
      action: 'soft_stop',
      patch_value: undefined,
    });

    set(state => ({
      novelclawMessages: [
        ...state.novelclawMessages,
        {
          id: `soft_ack_${Date.now()}`,
          thread_id: state.selectedThread?.id || 'main',
          project_id: selectedProject.id,
          sender: 'novelclaw',
          role: 'assistant',
          content: `${soul?.avatar ? soul.avatar + ' ' : ''}${i18n.t('cowork.softStopAck', 'I have sent a safe soft stop signal. The process will complete the current sentence and pause!')}`,
          created_at: new Date().toISOString(),
          token_count: 0,
          is_collapsed: false,
          is_archived_compact: false,
          thinking_content: '',
          step_type: '',
          step_status: '',
          action_call_json: '',
        },
      ],
    }));
  },

  triggerHardAbort: async () => {
    const { selectedProject, selectedChapter, batchProgress, soul } = get();
    // Emergency abort: instantly clear batch progress, flags, and translating states
    set({
      stopBatchRequested: true,
      batchProgress: null,
      isTranslating: false,
      isNovelClawThinking: false,
      novelclawActiveTool: null,
    });

    if (!selectedProject) return;
    const currentChapterIdx = batchProgress?.currentChapterIndex ?? selectedChapter?.chapter_index;
    const jobId = currentChapterIdx != null ? `job_${selectedProject.id}_${currentChapterIdx}` : '';

    await SoulService.TriggerSteerAction({
      job_id: jobId,
      project_id: selectedProject.id,
      action: 'hard_abort',
      patch_value: undefined,
    });

    set(state => ({
      novelclawMessages: [
        ...state.novelclawMessages,
        {
          id: `abort_ack_${Date.now()}`,
          thread_id: state.selectedThread?.id || 'main',
          project_id: selectedProject.id,
          sender: 'novelclaw',
          role: 'assistant',
          content: `${soul?.avatar ? soul.avatar + ' ' : ''}${i18n.t('cowork.hardAbortAck', 'Emergency abort signal received and translation process cancelled!')}`,
          created_at: new Date().toISOString(),
          token_count: 0,
          is_collapsed: false,
          is_archived_compact: false,
          thinking_content: '',
          step_type: '',
          step_status: '',
          action_call_json: '',
        },
      ],
    }));
  },

  // Soul & NovelClaw Companion Chat
  soul: null,
  pendingAskHuman: null,
  setPendingAskHuman: (data) => set({ pendingAskHuman: data }),
  resolveAskHuman: async (jobId: string, answer: string) => {
    try {
      const ok = await TranslationService.ResolveAskHuman(jobId, answer);
      set({ pendingAskHuman: null });
      return ok;
    } catch (err) {
      console.error('Failed to resolve ask human dilemma:', err);
      set({ pendingAskHuman: null });
      return false;
    }
  },
  availableSouls: [],
  fetchSoul: async () => {
    try {
      const s = await SoulService.GetSoul();
      if (s) set({ soul: s });
    } catch (err) {
      console.error('Failed to fetch soul:', err);
    }
  },

  fetchAvailableSouls: async () => {
    try {
      const souls = await SoulService.ListAvailableSouls();
      if (souls) set({ availableSouls: souls });
    } catch (err) {
      console.error('Failed to fetch available souls:', err);
    }
  },

  setActiveSoul: async (soulId: string, projectId?: string) => {
    try {
      const pId = projectId || get().selectedProject?.id || '';
      await SoulService.SetActiveSoul(pId, soulId);
      await get().fetchSoul();
      await get().fetchAvailableSouls();

      const newSoul = get().soul;
      if (newSoul) {
        const greetingText = newSoul.greeting || i18n.t('cowork.switchedSoulMsg', {
          defaultValue: 'Switched to assistant soul {{name}}.',
          name: newSoul.name,
        });
        const activeThreadId = get().selectedThread?.id || 'main';
        const msgId = `soul_switch_${Date.now()}`;
        const ts = new Date().toISOString();

        set((state) => ({
          novelclawMessages: [
            ...state.novelclawMessages,
            {
              id: msgId,
              thread_id: activeThreadId,
              project_id: pId,
              sender: 'novelclaw',
              role: 'assistant',
              content: `${newSoul.avatar ? newSoul.avatar + ' ' : ''}${greetingText}`,
              created_at: ts,
              token_count: 0,
              is_collapsed: false,
              is_archived_compact: false,
              thinking_content: '',
              step_type: '',
              step_status: '',
              action_call_json: '',
            },
          ],
        }));
      }
    } catch (err) {
      console.error('Failed to set active soul:', err);
    }
  },

  saveCustomSoul: async (soul: dtos.SoulDTO) => {
    try {
      await SoulService.SaveCustomSoul(soul);
      await get().fetchAvailableSouls();
      if (get().soul?.id === soul.id) {
        await get().fetchSoul();
      }
    } catch (err) {
      console.error('Failed to save custom soul:', err);
      throw err;
    }
  },

  deleteCustomSoul: async (soulId: string) => {
    try {
      await SoulService.DeleteCustomSoul(soulId);
      await get().fetchAvailableSouls();
      await get().fetchSoul();
    } catch (err) {
      console.error('Failed to delete custom soul:', err);
      throw err;
    }
  },

  restoreDefaultSouls: async () => {
    try {
      await SoulService.RestoreDefaultSouls();
      await get().fetchAvailableSouls();
      await get().fetchSoul();
    } catch (err) {
      console.error('Failed to restore default souls:', err);
      throw err;
    }
  },

  importSoul: async (content: string) => {
    try {
      const imported = await SoulService.ImportSoulMarkdown(content);
      await get().fetchAvailableSouls();
      return imported;
    } catch (err) {
      console.error('Failed to import soul:', err);
      throw err;
    }
  },

  exportSoul: async (soulId: string) => {
    try {
      return await SoulService.ExportSoulMarkdown(soulId);
    } catch (err) {
      console.error('Failed to export soul:', err);
      throw err;
    }
  },

  // Character Graph (L2)
  entities: [],
  relations: [],
  isGraphLoading: false,

  fetchGraphData: async (chapterIndex?: number) => {
    const proj = get().selectedProject;
    if (!proj) {
      set({ entities: [], relations: [] });
      return;
    }
    set({ isGraphLoading: true });
    try {
      const ents = await GraphService.ListEntities(proj.id);
      let rels: dtos.RelationDTO[] | null = null;
      if (chapterIndex && chapterIndex > 0) {
        rels = await GraphService.ListEffectiveRelationsAtChapter(proj.id, chapterIndex);
      } else {
        rels = await GraphService.ListRelations(proj.id);
      }
      set({ entities: ents || [], relations: rels || [] });
    } catch (err) {
      console.error('Failed to fetch graph data:', err);
    } finally {
      set({ isGraphLoading: false });
    }
  },

  upsertRelation: async (req) => {
    try {
      await GraphService.UpsertRelation(req);
      await get().fetchGraphData();
    } catch (err) {
      console.error('Failed to upsert relation:', err);
    }
  },

  toggleLockRelation: async (id) => {
    const rel = get().relations.find(r => r.id === id);
    if (!rel) return;
    try {
      await GraphService.UpsertRelation({
        id: rel.id,
        project_id: rel.project_id,
        from_char: rel.from_char,
        to_char: rel.to_char,
        call_as: rel.call_as,
        self_call_as: rel.self_call_as,
        since_chapter: rel.since_chapter,
        tone: rel.tone,
        is_locked: !rel.is_locked,
      });
      await get().fetchGraphData();
    } catch (err) {
      console.error('Failed to toggle lock relation:', err);
    }
  },

  deleteRelation: async (id) => {
    try {
      await GraphService.DeleteRelation(id);
      await get().fetchGraphData();
    } catch (err) {
      console.error('Failed to delete relation:', err);
      throw err;
    }
  },

  upsertEntity: async (req) => {
    try {
      await GraphService.UpsertEntity(req);
      await get().fetchGraphData();
    } catch (err) {
      console.error('Failed to upsert entity:', err);
    }
  },

  deleteEntity: async (id) => {
    try {
      await GraphService.DeleteEntity(id);
      await get().fetchGraphData();
    } catch (err) {
      console.error('Failed to delete entity:', err);
      throw err;
    }
  },

  autoScanEntities: async (chapterIndex = 0, scanMode = 'volume', volume?: string) => {
    const proj = get().selectedProject;
    if (!proj) return null;
    set({ isGraphLoading: true });
    try {
      const res = await GraphService.AutoScanEntities({
        project_id: proj.id,
        chapter_index: chapterIndex,
        scan_mode: scanMode,
        volume: volume || '',
      });
      await get().fetchGraphData();
      return res;
    } catch (err) {
      console.error('Failed to auto-scan entities:', err);
      throw err;
    } finally {
      set({ isGraphLoading: false });
    }
  },

  checkGraphReadiness: async (chapterIndex: number) => {
    const proj = get().selectedProject;
    if (!proj) return null;
    try {
      return await GraphService.CheckGraphReadiness(proj.id, chapterIndex);
    } catch (err) {
      console.error('Failed to check graph readiness:', err);
      return null;
    }
  },

  // Settings & LLM Configs
  llmConfigs: [],
  defaultLLMConfig: null,
  fallbackChain: getInitialFallbackChain(),
  thinkingEffort: (localStorage.getItem('neko_thinking_effort') as any) || 'off',

  setThinkingEffort: (effort) => {
    try {
      localStorage.setItem('neko_thinking_effort', effort);
    } catch (e) {
      console.error(e);
    }
    set({ thinkingEffort: effort });
  },

  setFallbackChain: (chain) => {
    set({ fallbackChain: chain });
  },

  fetchLLMConfigs: async () => {
    try {
      const configs = await LLMService.ListConfigs();
      const def = await LLMService.GetDefaultConfig();
      set({ llmConfigs: configs || [], defaultLLMConfig: def || null });
      // Reconcile the fallback chain against the providers that ACTUALLY
      // exist in the database: drop tiers whose config was deleted and
      // append tiers for newly added configs. This prevents the chain from
      // showing stale/phantom entries (e.g. from an old localStorage blob).
      get().reconcileFallbackChain(configs || []);
    } catch (err) {
      console.error('Failed to fetch LLM configs:', err);
    }
  },

  reconcileFallbackChain: (configs) => {
    const current = get().fallbackChain || [];
    const byId = new Map(configs.map((c) => [c.id, c]));

    // Keep only tiers whose config still exists.
    const kept = current.filter((tier) => byId.has(tier.configId));

    // Refresh each surviving tier's provider/model from the live config.
    const refreshed = kept.map((tier) => {
      const cfg = byId.get(tier.configId)!;
      return {
        ...tier,
        providerName: cfg.provider_name,
        modelName: cfg.model_name || '',
        timeoutSeconds: cfg.timeout_seconds ?? tier.timeoutSeconds,
        maxRetries: cfg.max_retries ?? tier.maxRetries,
      };
    });

    // Append any config that has no tier yet (newly added providers).
    const present = new Set(refreshed.map((t) => t.configId));
    const appended = [...refreshed];
    for (const cfg of configs) {
      if (present.has(cfg.id)) continue;
      const i = appended.length;
      appended.push({
        id: `tier_${cfg.id}`,
        name: i === 0 ? i18n.t('settings.tierPrimary') : i === 1 ? i18n.t('settings.tierSecondary') : i18n.t('settings.tierTertiary'),
        configId: cfg.id,
        providerName: cfg.provider_name,
        modelName: cfg.model_name || '',
        maxRetries: cfg.max_retries ?? 3,
        timeoutSeconds: cfg.timeout_seconds ?? 120,
        reasoningEffort: (cfg.reasoning_effort as FallbackTier['reasoningEffort']) ?? 'off',
        triggerCondition: '429_5xx_timeout',
        enabled: true,
      });
    }

    const changed =
      appended.length !== current.length ||
      appended.some((t, i) => current[i]?.configId !== t.configId || current[i]?.providerName !== t.providerName || current[i]?.modelName !== t.modelName);

    if (changed) {
      set({ fallbackChain: appended });
    }
  },

  saveLLMConfig: async (req) => {
    try {
      await LLMService.SaveConfig(req);
      await get().fetchLLMConfigs();
    } catch (err) {
      console.error('Failed to save LLM config:', err);
      throw err;
    }
  },

  deleteLLMConfig: async (id) => {
    try {
      await LLMService.DeleteConfig(id);
      await get().fetchLLMConfigs();
    } catch (err) {
      console.error('Failed to delete LLM config:', err);
      throw err;
    }
  },

  setDefaultLLMConfig: async (id) => {
    try {
      await LLMService.SetDefaultConfig(id);
      await get().fetchLLMConfigs();
    } catch (err) {
      console.error('Failed to set default LLM config:', err);
      throw err;
    }
  },

  testLLMConnection: async (req) => {
    try {
      return await LLMService.TestConnection(req);
    } catch (err) {
      console.error('Failed to test LLM connection:', err);
      throw err;
    }
  },

  batchImportBooks: async (req) => {
    try {
      const res = await ProjectService.ImportBooksBatch(req);
      if (res && res.projects && res.projects.length > 0) {
        await get().fetchProjects();
        get().selectProject(res.projects[0]);
      }
      return res;
    } catch (err) {
      console.error('Failed to batch import books:', err);
      throw err;
    }
  },

  previewStyle: async (req) => {
    try {
      return await TranslationService.PreviewStyle(req);
    } catch (err) {
      console.error('Failed to preview style:', err);
      throw err;
    }
  },

  autoScoutStyle: async (req) => {
    try {
      const res = await TranslationService.AutoScoutStyle(req);
      if (res && res.success) {
        set((state) => ({
          styleGuide: res.style_prompt,
          selectedProject: state.selectedProject?.id === req.project_id
            ? { ...state.selectedProject, style_name: res.style_name, style_guide: res.style_prompt }
            : state.selectedProject,
          projects: state.projects.map((p) =>
            p.id === req.project_id ? { ...p, style_name: res.style_name, style_guide: res.style_prompt } : p
          ),
        }));
      }
      return res;
    } catch (err) {
      console.error('Failed to auto scout style:', err);
      throw err;
    }
  },

  // Glossary (L3)
  terms: [],
  isGlossaryLoading: false,

  fetchGlossaryTerms: async () => {
    const proj = get().selectedProject;
    if (!proj) {
      set({ terms: [] });
      return;
    }
    set({ isGlossaryLoading: true });
    try {
      const terms = await GlossaryService.ListTerms(proj.id);
      set({ terms: terms || [] });
    } catch (err) {
      console.error('Failed to fetch glossary terms:', err);
    } finally {
      set({ isGlossaryLoading: false });
    }
  },

  upsertGlossaryTerm: async (req) => {
    try {
      await GlossaryService.UpsertTerm(req);
      await get().fetchGlossaryTerms();
    } catch (err) {
      console.error('Failed to upsert glossary term:', err);
      throw err;
    }
  },

  deleteGlossaryTerm: async (id) => {
    try {
      await GlossaryService.DeleteTerm(id);
      await get().fetchGlossaryTerms();
    } catch (err) {
      console.error('Failed to delete glossary term:', err);
      throw err;
    }
  },

  importGlossaryTSV: async (content) => {
    const proj = get().selectedProject;
    if (!proj) return 0;
    try {
      const count = await GlossaryService.ImportTSV(proj.id, content);
      await get().fetchGlossaryTerms();
      return count || 0;
    } catch (err) {
      console.error('Failed to import glossary TSV:', err);
      throw err;
    }
  },

  exportGlossaryTSV: async () => {
    const proj = get().selectedProject;
    if (!proj) return '';
    try {
      return (await GlossaryService.ExportTSV(proj.id)) || '';
    } catch (err) {
      console.error('Failed to export glossary TSV:', err);
      throw err;
    }
  },

  importGlossaryJSON: async (content) => {
    const proj = get().selectedProject;
    if (!proj) return 0;
    try {
      const count = await GlossaryService.ImportJSON(proj.id, content);
      await get().fetchGlossaryTerms();
      return count || 0;
    } catch (err) {
      console.error('Failed to import glossary JSON:', err);
      throw err;
    }
  },

  exportGlossaryJSON: async () => {
    const proj = get().selectedProject;
    if (!proj) return '';
    try {
      return (await GlossaryService.ExportJSON(proj.id)) || '';
    } catch (err) {
      console.error('Failed to export glossary JSON:', err);
      throw err;
    }
  },

  importGlossaryPlainText: async (content: string, defaultCategory: string = 'other') => {
    const proj = get().selectedProject;
    if (!proj) return 0;
    try {
      const count = await GlossaryService.ImportPlainText(proj.id, content, defaultCategory);
      await get().fetchGlossaryTerms();
      return count || 0;
    } catch (err) {
      console.error('Failed to import glossary plain text:', err);
      throw err;
    }
  },

  listRemoteGlossarySources: async () => {
    try {
      const sources = await GlossaryService.ListRemoteSources();
      return sources || [];
    } catch (err) {
      console.error('Failed to list remote glossary sources:', err);
      return [];
    }
  },

  downloadAndImportGlossary: async (url: string, format: string = 'vietphrase', defaultCategory: string = 'other') => {
    const proj = get().selectedProject;
    if (!proj) return 0;
    try {
      const count = await GlossaryService.DownloadAndImport({
        project_id: proj.id,
        url: url.trim(),
        format: format,
        default_category: defaultCategory,
      });
      await get().fetchGlossaryTerms();
      return count || 0;
    } catch (err) {
      console.error('Failed to download and import remote glossary:', err);
      throw err;
    }
  },

  // Benchmark
  benchmarkReport: null,
  isBenchmarking: false,

  runBenchmark: async (modes, chapterIndices) => {
    const proj = get().selectedProject;
    if (!proj) return null;
    set({ isBenchmarking: true, benchmarkReport: null });
    try {
      const report = await TranslationService.RunBenchmark({
        project_id: proj.id,
        modes: modes,
        chapter_indices: chapterIndices,
      });
      set({ benchmarkReport: report });
      return report;
    } catch (err) {
      console.error('Failed to run benchmark:', err);
      throw err;
    } finally {
      set({ isBenchmarking: false });
    }
  },

  // NovelClaw Agentic Chat & Threads
  novelclawThreads: [],
  selectedThread: null,
  novelclawMessages: [],
  isNovelClawThinking: false,
  novelclawThinkingText: '',
  novelclawActiveTool: null,
  autoCompactThreshold: 250000,

  fetchNovelClawThreads: async (projectID?: string) => {
    const proj = projectID || get().selectedProject?.id;
    if (!proj) return;
    try {
      const list = await NovelClawService.ListThreads(proj);
      const threadList = list ? [...list] : [];
      const currentSelected = get().selectedThread;
      let nextSelected: dtos.NovelClawThreadDTO | null = null;
      if (threadList.length > 0) {
        const stillExists = currentSelected && threadList.find(t => t.id === currentSelected.id);
        nextSelected = stillExists || threadList[0];
      } else {
        const main = await NovelClawService.GetOrCreateMainThread(proj);
        if (main) {
          threadList.push(main);
          nextSelected = main;
        }
      }
      set({ novelclawThreads: threadList, selectedThread: nextSelected });
      if (nextSelected) {
        get().fetchNovelClawMessages(nextSelected.id);
      }
    } catch (err) {
      console.error('Failed to fetch NovelClaw threads:', err);
    }
  },

  selectNovelClawThread: async (thread) => {
    set({ selectedThread: thread });
    await get().fetchNovelClawMessages(thread.id);
  },

  createSubThread: async (req) => {
    try {
      const thread = await NovelClawService.CreateSubThread(req);
      if (thread) {
        await get().fetchNovelClawThreads(req.project_id);
        set({ selectedThread: thread });
        await get().fetchNovelClawMessages(thread.id);
      }
      return thread;
    } catch (err) {
      console.error('Failed to create sub-thread:', err);
      return null;
    }
  },

  deleteNovelClawThread: async (threadID) => {
    try {
      await NovelClawService.DeleteThread(threadID);
      const proj = get().selectedProject?.id;
      if (proj) {
        await get().fetchNovelClawThreads(proj);
      }
    } catch (err) {
      console.error('Failed to delete thread:', err);
    }
  },

  fetchNovelClawMessages: async (threadID?: string) => {
    const targetID = threadID || get().selectedThread?.id;
    if (!targetID) return;
    try {
      const msgs = await NovelClawService.ListMessages(targetID);
      set({ novelclawMessages: msgs || [] });
    } catch (err) {
      console.error('Failed to fetch NovelClaw messages:', err);
    }
  },

  sendNovelClawMessage: async (content: string) => {
    const { selectedProject, selectedThread } = get();
    if (!selectedProject) return;
    let targetThread = selectedThread;
    if (!targetThread) {
      const main = await NovelClawService.GetOrCreateMainThread(selectedProject.id);
      if (main) {
        set({ selectedThread: main });
        targetThread = main;
      }
    }
    if (!targetThread) return;

    set({ isNovelClawThinking: true, novelclawThinkingText: '', novelclawActiveTool: null });

    try {
      await NovelClawService.SendMessage({
        thread_id: targetThread.id,
        project_id: selectedProject.id,
        content: content,
      });
      await get().fetchNovelClawMessages(targetThread.id);
      await get().fetchNovelClawThreads(selectedProject.id);
    } catch (err) {
      console.error('Failed to send NovelClaw message:', err);
    } finally {
      set({ isNovelClawThinking: false, novelclawActiveTool: null });
    }
  },

  clearNovelClawThread: async () => {
    const { selectedThread, selectedProject } = get();
    if (!selectedThread) return;
    try {
      await NovelClawService.ClearThread(selectedThread.id);
      set({ novelclawMessages: [] });
      if (selectedProject) {
        await get().fetchNovelClawThreads(selectedProject.id);
      }
    } catch (err) {
      console.error('Failed to clear thread:', err);
    }
  },

  compactNovelClawThread: async () => {
    const { selectedThread, selectedProject } = get();
    if (!selectedThread) return;
    try {
      await NovelClawService.CompactThread(selectedThread.id);
      await get().fetchNovelClawMessages(selectedThread.id);
      if (selectedProject) {
        await get().fetchNovelClawThreads(selectedProject.id);
      }
    } catch (err) {
      console.error('Failed to compact thread:', err);
    }
  },

  setAutoCompactThreshold: async (tokens: number) => {
    try {
      await NovelClawService.SetAutoCompactThreshold(tokens);
      set({ autoCompactThreshold: tokens });
    } catch (err) {
      console.error('Failed to set auto-compact threshold:', err);
    }
  },

  setNovelClawThinkingText: (text: string) => set({ novelclawThinkingText: text }),
  setNovelClawActiveTool: (tool: string | null) => set({ novelclawActiveTool: tool }),
  setIsNovelClawThinking: (thinking: boolean) => set({ isNovelClawThinking: thinking }),

  // Modular Skills & Capabilities (Stage 9)
  skills: [],
  skillPresets: [],
  activeSkillsTokenWeight: 0,
  isSkillsLoading: false,

  fetchSkills: async (projectID?: string) => {
    const projId = projectID || get().selectedProject?.id || 'default';
    set({ isSkillsLoading: true });
    try {
      const skills = await SkillService.ListSkills(projId);
      set({ skills: skills || [] });
      await get().fetchActiveSkillsTokenWeight(projId);
    } catch (err) {
      console.error('Failed to fetch skills:', err);
    } finally {
      set({ isSkillsLoading: false });
    }
  },

  fetchSkillPresets: async () => {
    try {
      const presets = await SkillService.GetSkillPresets();
      set({ skillPresets: presets || [] });
    } catch (err) {
      console.error('Failed to fetch skill presets:', err);
    }
  },

  toggleSkill: async (skillID: string, enabled: boolean, projectID?: string) => {
    const projId = projectID || get().selectedProject?.id || 'default';
    try {
      await SkillService.ToggleSkill(projId, skillID, enabled);
      await get().fetchSkills(projId);
    } catch (err) {
      console.error('Failed to toggle skill:', err);
    }
  },

  updateSkillCustomRules: async (skillID: string, customRules: string, projectID?: string) => {
    const projId = projectID || get().selectedProject?.id || 'default';
    try {
      await SkillService.UpdateSkillCustomRules(projId, skillID, customRules);
      await get().fetchSkills(projId);
    } catch (err) {
      console.error('Failed to update skill custom rules:', err);
    }
  },

  resetSkillToDefault: async (skillID: string, projectID?: string) => {
    const projId = projectID || get().selectedProject?.id || 'default';
    try {
      await SkillService.ResetSkillToDefault(projId, skillID);
      await get().fetchSkills(projId);
    } catch (err) {
      console.error('Failed to reset skill:', err);
    }
  },

  applyCapabilityPreset: async (presetName: string, projectID?: string) => {
    const projId = projectID || get().selectedProject?.id || 'default';
    try {
      await SkillService.ApplyCapabilityPreset(projId, presetName);
      await get().fetchSkills(projId);
    } catch (err) {
      console.error('Failed to apply capability preset:', err);
    }
  },

  fetchActiveSkillsTokenWeight: async (projectID?: string) => {
    const projId = projectID || get().selectedProject?.id || 'default';
    try {
      const weight = await SkillService.GetActiveSkillsTokenWeight(projId);
      set({ activeSkillsTokenWeight: weight || 0 });
    } catch (err) {
      console.error('Failed to fetch active skills token weight:', err);
    }
  },

  // Universal World Bible & Lorebook (Stage 10)
  worldCategories: [],
  worldEntries: [],
  selectedWorldCategory: null,
  isWorldBibleLoading: false,
  isScanningWorld: false,
  worldScanReport: null,

  setSelectedWorldCategory: (catID: string | null) => {
    set({ selectedWorldCategory: catID });
    get().fetchWorldEntries(undefined, catID || undefined);
  },

  fetchWorldCategories: async (projectID?: string) => {
    const projId = projectID || get().selectedProject?.id;
    if (!projId) {
      set({ worldCategories: [] });
      return;
    }
    try {
      set({ isWorldBibleLoading: true });
      const cats = await WorldBibleService.ListCategories(projId);
      set({ worldCategories: cats || [] });
    } catch (err) {
      console.error('Failed to fetch world categories:', err);
    } finally {
      set({ isWorldBibleLoading: false });
    }
  },

  fetchWorldEntries: async (projectID?: string, categoryID?: string) => {
    const projId = projectID || get().selectedProject?.id;
    if (!projId) {
      set({ worldEntries: [] });
      return;
    }
    try {
      set({ isWorldBibleLoading: true });
      const entries = await WorldBibleService.ListEntries(projId, categoryID || '');
      set({ worldEntries: entries || [] });
    } catch (err) {
      console.error('Failed to fetch world entries:', err);
    } finally {
      set({ isWorldBibleLoading: false });
    }
  },

  upsertWorldCategory: async (req: dtos.UpsertWorldCategoryRequest) => {
    try {
      const cat = await WorldBibleService.UpsertCategory(req);
      await get().fetchWorldCategories(req.project_id);
      return cat;
    } catch (err) {
      console.error('Failed to upsert world category:', err);
      return null;
    }
  },

  deleteWorldCategory: async (id: string) => {
    const projId = get().selectedProject?.id;
    try {
      await WorldBibleService.DeleteCategory(id);
      if (projId) {
        await get().fetchWorldCategories(projId);
        await get().fetchWorldEntries(projId);
      }
    } catch (err) {
      console.error('Failed to delete world category:', err);
    }
  },

  upsertWorldEntry: async (req: dtos.UpsertWorldEntryRequest) => {
    try {
      const entry = await WorldBibleService.UpsertEntry(req);
      await get().fetchWorldEntries(req.project_id, get().selectedWorldCategory || undefined);
      await get().fetchWorldCategories(req.project_id);
      return entry;
    } catch (err) {
      console.error('Failed to upsert world entry:', err);
      return null;
    }
  },

  verifyWorldEntry: async (id: string, verified: boolean) => {
    const projId = get().selectedProject?.id;
    try {
      await WorldBibleService.VerifyEntry(id, verified);
      if (projId) {
        await get().fetchWorldEntries(projId, get().selectedWorldCategory || undefined);
      }
    } catch (err) {
      console.error('Failed to verify world entry:', err);
    }
  },

  deleteWorldEntry: async (id: string) => {
    const projId = get().selectedProject?.id;
    try {
      await WorldBibleService.DeleteEntry(id);
      if (projId) {
        await get().fetchWorldEntries(projId, get().selectedWorldCategory || undefined);
        await get().fetchWorldCategories(projId);
      }
    } catch (err) {
      console.error('Failed to delete world entry:', err);
    }
  },

  scanWorldBible: async (startChap: number, endChap: number, projectID?: string) => {
    const projId = projectID || get().selectedProject?.id;
    if (!projId) return null;
    try {
      set({ isScanningWorld: true, worldScanReport: null });
      const resp = await WorldBibleService.ScanAndBuildWorld(projId, startChap, endChap);
      set({ worldScanReport: resp });
      await get().fetchWorldCategories(projId);
      await get().fetchWorldEntries(projId);
      return resp;
    } catch (err) {
      console.error('Failed to scan world:', err);
      return null;
    } finally {
      set({ isScanningWorld: false });
    }
  },

  exportWorldMarkdown: async (projectID?: string) => {
    const projId = projectID || get().selectedProject?.id;
    if (!projId) return '';
    try {
      return await WorldBibleService.ExportWorldBibleMarkdown(projId);
    } catch (err) {
      console.error('Failed to export world markdown:', err);
      return '';
    }
  },

  exportWorldJSON: async (projectID?: string) => {
    const projId = projectID || get().selectedProject?.id;
    if (!projId) return '';
    try {
      return await WorldBibleService.ExportWorldBibleJSON(projId);
    } catch (err) {
      console.error('Failed to export world JSON:', err);
      return '';
    }
  },

  importWorldJSON: async (jsonStr: string, projectID?: string) => {
    const projId = projectID || get().selectedProject?.id;
    if (!projId) return [0, 0];
    try {
      const counts = await WorldBibleService.ImportWorldBibleJSON(projId, jsonStr);
      await get().fetchWorldCategories(projId);
      await get().fetchWorldEntries(projId);
      return counts;
    } catch (err) {
      console.error('Failed to import world JSON:', err);
      return [0, 0];
    }
  },

  resetSkillsToDefault: async (projectID?: string) => {
    const projId = projectID || get().selectedProject?.id || 'default';
    try {
      await WorldBibleService.ResetSkillsToDefault(projId);
      await get().fetchSkills(projId);
    } catch (err) {
      console.error('Failed to reset skills to default:', err);
    }
  },

  learnFromEdits: async (chapterIndex: number, notes: string, projectID?: string) => {
    const projId = projectID || get().selectedProject?.id;
    if (!projId) return null;
    try {
      const res = await WorldBibleService.LearnFromEdits(projId, chapterIndex, notes);
      await get().fetchSkills(projId);
      return res;
    } catch (err) {
      console.error('Failed to learn from edits:', err);
      return null;
    }
  },
}));
