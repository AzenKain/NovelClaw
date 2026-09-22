import { toast } from 'react-hot-toast';
import { useAppStore } from './useAppStore';
import i18n from '@/i18n';

export interface AppActionDispatchedEvent {
  action: string;
  category: string;
  description: string;
  data: Record<string, any>;
  timestamp?: string;
}

/**
 * dispatchAppAction updates frontend state immediately in response to NovelClaw Omni-App tools.
 */
export function dispatchAppAction(payload: AppActionDispatchedEvent) {
  if (!payload || !payload.action) return;

  const store = useAppStore.getState();

  // Show live toast for user feedback
  if (payload.description) {
    toast(payload.description, {
      style: {
        borderRadius: '10px',
        background: '#111726',
        color: '#e2e8f0',
        border: '1px solid rgba(99, 102, 241, 0.3)',
        fontSize: '12px',
      },
      duration: 3500,
    });
  }

  const { action, data = {} } = payload;

  switch (action) {
    // -------------------------------------------------------------
    // Group 1: Navigation & UI
    // -------------------------------------------------------------
    case 'switch_tab':
      if (data.tab && ['workspace', 'graph', 'glossary', 'world_bible', 'benchmark'].includes(data.tab)) {
        store.setActiveTab(data.tab as any);
      }
      break;

    case 'select_volume':
      if (typeof data.volume_index === 'number') {
        const volTag = data.volume_tag || (data.volume_index > 0 ? `Vol ${data.volume_index}` : 'all');
        store.setSelectedVolume(volTag);
      } else if (data.volume_tag) {
        store.setSelectedVolume(data.volume_tag);
      }
      break;

    case 'select_chapter':
      if (typeof data.chapter_index === 'number') {
        const target = store.chapters.find(c => c.chapter_index === data.chapter_index);
        if (target) {
          store.selectChapter(target);
        }
      }
      break;

    case 'open_modal':
      switch (data.modal_name) {
        case 'settings':
          store.setSettingsOpen(true);
          break;
        case 'export':
          store.setExportOpen(true);
          break;
        case 'import':
          store.setImportOpen(true);
          break;
        case 'project_manager':
          store.setProjectManagerOpen(true);
          break;
        case 'style_scout':
          store.setStyleScoutOpen(true);
          break;
      }
      break;

    case 'scroll_to_text':
      if (data.keyword) {
        const keyword = String(data.keyword).toLowerCase();
        const readers = document.querySelectorAll('.reader-content, [data-reader-content]');
        let found = false;
        readers.forEach((container) => {
          if (found) return;
          const walker = document.createTreeWalker(container, NodeFilter.SHOW_TEXT);
          let node = walker.nextNode();
          while (node) {
            if (node.textContent && node.textContent.toLowerCase().includes(keyword)) {
              const parentEl = node.parentElement;
              if (parentEl) {
                parentEl.scrollIntoView({ behavior: 'smooth', block: 'center' });
                parentEl.classList.add('bg-indigo-500/20', 'transition-colors', 'duration-500');
                setTimeout(() => {
                  parentEl.classList.remove('bg-indigo-500/20');
                }, 3000);
                found = true;
                break;
              }
            }
            node = walker.nextNode();
          }
        });
      }
      break;

    // -------------------------------------------------------------
    // Group 2 & 3: Character Graph & World Bible
    // -------------------------------------------------------------
    case 'character_created':
    case 'relation_updated':
    case 'world_category_upserted':
    case 'world_entry_upserted':
    case 'world_entry_deleted':
      // Reload graph entities, relations and world bible
      store.fetchGraphData();
      store.fetchWorldCategories();
      store.fetchWorldEntries();
      break;

    case 'graph_scan_triggered':
      if (store.autoScanEntities) {
        store.autoScanEntities(data.start_chapter || 0, data.start_chapter ? 'chapter' : 'all');
      }
      store.fetchGraphData();
      break;

    case 'world_bible_scanned':
      store.fetchWorldCategories();
      store.fetchWorldEntries();
      break;

    // -------------------------------------------------------------
    // Group 4: Glossary & Terminology
    // -------------------------------------------------------------
    case 'glossary_term_added':
    case 'glossary_bulk_imported':
    case 'glossary_term_deleted':
      store.fetchGlossaryTerms();
      break;

    // -------------------------------------------------------------
    // Group 5: Pipeline & Model Routing
    // -------------------------------------------------------------
    case 'update_translation_config':
      if (data.mode) store.setTranslationMode(data.mode);
      if (typeof data.enable_hot_patch === 'boolean') store.setEnableHotPatch(data.enable_hot_patch);
      if (typeof data.enable_r19 === 'boolean') store.setEnableR19(data.enable_r19);
      if (typeof data.enable_agentic_rag === 'boolean') store.setEnableAgenticRAG(data.enable_agentic_rag);
      if (data.style_guide !== undefined) store.setStyleGuide(data.style_guide);
      break;

    case 'llm_configured':
      store.fetchLLMConfigs();
      break;

    case 'soul_changed':
      if (data?.soul_id) {
        store.setActiveSoul(data.soul_id);
      } else {
        store.fetchSoul();
        store.fetchAvailableSouls();
      }
      break;

    // -------------------------------------------------------------
    // Group 6: Translation Execution & Rollback
    // -------------------------------------------------------------
    case 'start_translation':
      if (typeof data.start_chapter === 'number' && data.start_chapter > 0) {
        const target = store.chapters.find(c => c.chapter_index === data.start_chapter);
        if (target && target.chapter_index !== store.selectedChapter?.chapter_index) {
          store.selectChapter(target);
        }
      }
      store.startTranslation();
      break;

    case 'pause_translation':
    case 'soft_stop_translation':
      store.triggerSoftStop();
      break;

    case 'resume_translation':
      store.startTranslation();
      break;

    case 'abort_translation':
      store.triggerHardAbort();
      break;

    case 'rollback_checkpoint':
      if (typeof data.chapter_index === 'number' && data.chapter_index > 0) {
        const target = store.chapters.find(c => c.chapter_index === data.chapter_index);
        if (target && target.chapter_index !== store.selectedChapter?.chapter_index) {
          store.selectChapter(target);
        }
      }
      store.rollbackChapter();
      break;

    // -------------------------------------------------------------
    // Group 7: Export, Evolution & Benchmark
    // -------------------------------------------------------------
    case 'export_book':
      store.setExportOpen(true);
      break;

    case 'trigger_learn':
      toast.success(i18n.t('cowork.learnSuccess', 'Activated self-learning and skill evolution process.'));
      break;

    case 'run_benchmark':
      store.setActiveTab('benchmark');
      break;

    default:
      console.log('[ActionDispatcher] Unhandled app action:', action, data);
  }
}
