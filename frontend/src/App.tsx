import React, { useEffect } from 'react';
import { Toaster, toast } from 'react-hot-toast';
import { useTranslation } from 'react-i18next';
import '@/i18n';
import { Navbar } from '@/components/Navbar';
import { CoworkerWorkspace } from '@/components/CoworkerWorkspace';
import { CharacterGraph } from '@/components/CharacterGraph';
import { GlossaryHub } from '@/components/GlossaryHub';
import { BenchmarkDashboard } from '@/components/BenchmarkDashboard';
import { WorldBibleView } from '@/components/WorldBibleView';
import { StyleScoutModal } from '@/components/StyleScoutModal';
import { ExportModal } from '@/components/ExportModal';
import { ImportModal } from '@/components/ImportModal';
import { SettingsModal } from '@/components/SettingsModal';
import { ProjectManagerModal } from '@/components/ProjectManagerModal';
import { ConfirmModal } from '@/components/ConfirmModal';
import { LoadingScreen } from '@/components/LoadingScreen';
import { SetupWizard } from '@/components/SetupWizard';
import { Events } from '@wailsio/runtime';
import { useAppStore } from '@/store/useAppStore';
import { dispatchAppAction } from '@/store/actionDispatcher';

export const App: React.FC = () => {
  const { t } = useTranslation();
  const { activeTab, isAppLoading, isFirstRun, initializeApp } = useAppStore();

  useEffect(() => {
    initializeApp();

    const unsubs: (() => void)[] = [];
    if (Events && typeof Events.On === 'function') {
      try {
        unsubs.push(
          Events.On('translation:thought', (e: any) => {
            const data = e?.data;
            if (data?.thought) {
              useAppStore.setState({
                thoughtText: data.thought,
                isNovelClawThinking: true,
                novelclawThinkingText: data.thought,
              });
            }
          })
        );
        unsubs.push(
          Events.On('translation:tool_call', (e: any) => {
            const data = e?.data;
            if (data?.tool_name) {
              let label = `Tool: ${data.tool_name}`;
              if (data.tool_name === 'lookup_character_relation') {
                label = `● ${t('tools.charRelation')}`;
              } else if (data.tool_name === 'lookup_world_lore') {
                label = `● ${t('tools.worldLore')}`;
              } else if (data.tool_name === 'search_book_context') {
                label = `● ${t('tools.searchContext')}`;
              } else if (data.tool_name === 'web_lookup') {
                label = `● ${t('tools.webLookup')}`;
              } else if (data.tool_name === 'ask_human_coworker') {
                label = `● ${t('tools.askHuman')}`;
              } else if (data.tool_name === 'shadow_critic_hotpatch' || data.tool_name === 'shadow_critic_review') {
                label = `● ${t('tools.shadowCritic')}`;
              } else if (data.tool_name === 'match_glossary_terms') {
                label = `● ${t('tools.glossaryContext')}`;
              }
              useAppStore.setState({ activeTool: label, novelclawActiveTool: label });

              // Append live tool call card to NovelClaw chat if not duplicate
              const state = useAppStore.getState();
              const activeThreadId = state.selectedThread?.id || 'main';
              const isDup = state.novelclawMessages.some(
                m => m.step_type === 'tool_call' && m.action_call_json?.includes(data.tool_name) && Math.abs(new Date(m.created_at).getTime() - Date.now()) < 800
              );
              if (!isDup) {
                const toolMsg: any = {
                  id: `tool_${Date.now()}_${Math.random().toString(36).substring(2, 7)}`,
                  thread_id: activeThreadId,
                  project_id: data.project_id || state.selectedProject?.id || '',
                  sender: 'novelclaw',
                  role: 'tool',
                  content: t('tools.toolExecution', { toolName: data.tool_name }),
                  thinking_content: '',
                  step_type: 'tool_call',
                  step_status: 'completed',
                  action_call_json: JSON.stringify({
                    tool_name: data.tool_name,
                    input: data.input,
                    output: data.output,
                  }),
                  is_collapsed: true,
                  is_archived_compact: false,
                  token_count: 0,
                  created_at: new Date().toISOString(),
                };
                useAppStore.setState({
                  novelclawMessages: [...state.novelclawMessages, toolMsg],
                });
              }
            }
          })
        );
        unsubs.push(
          Events.On('translation:status', (e: any) => {
            const data = e?.data;
            if (data?.status) {
              useAppStore.setState({ activeTool: data.status, novelclawActiveTool: data.status });

              const text = data.status;
              const isMajor = text.includes('[Pass') || text.includes('[Single-Pass]') || text.includes('[Dual-Agent]') || text.includes('Starting') || text.includes('Shadow Critic') || text.includes('Saving');
              if (isMajor) {
                const state = useAppStore.getState();
                const activeThreadId = state.selectedThread?.id || 'main';
                const isDup = state.novelclawMessages.some(
                  m => m.step_type === 'step_status' && m.content === text && Math.abs(new Date(m.created_at).getTime() - Date.now()) < 1000
                );
                if (!isDup) {
                  const stepMsg: any = {
                    id: `step_${Date.now()}_${Math.random().toString(36).substring(2, 6)}`,
                    thread_id: activeThreadId,
                    project_id: data.project_id || state.selectedProject?.id || '',
                    sender: 'novelclaw',
                    role: 'assistant',
                    content: text,
                    thinking_content: '',
                    step_type: 'step_status',
                    step_status: 'running',
                    action_call_json: '',
                    is_collapsed: false,
                    is_archived_compact: false,
                    token_count: 0,
                    created_at: new Date().toISOString(),
                  };
                  useAppStore.setState({
                    novelclawMessages: [...state.novelclawMessages, stepMsg],
                  });
                }
              }
            }
          })
        );
        unsubs.push(
          Events.On('translation:chunk_content', (e: any) => {
            const data = e?.data ?? e;
            if (data?.content) {
              const currentChapter = useAppStore.getState().selectedChapter;
              if (currentChapter && (currentChapter.chapter_index === data.chapter_index || data.chapter_index === undefined)) {
                const prev = currentChapter.translated_content || '';
                const updated = prev ? `${prev}\n\n${data.content}` : data.content;
                useAppStore.setState({
                  selectedChapter: {
                    ...currentChapter,
                    translated_content: updated,
                  },
                });
              }
            }
          })
        );
        unsubs.push(
          Events.On('app:action_dispatched', (e: any) => {
            const data = e?.data;
            if (data) {
              dispatchAppAction(data);
            }
          })
        );
        unsubs.push(
          Events.On('novelclaw:event', (e: any) => {
            const ev = e?.data;
            if (!ev) return;
            const { type, data } = ev;
            if (type === 'thinking') {
              useAppStore.setState({
                isNovelClawThinking: true,
                novelclawThinkingText: data?.content || '',
              });
            } else if (type === 'step_badge') {
              useAppStore.getState().setNovelClawActiveTool(data?.tool_name || null);
            } else if (type === 'message_created') {
              if (data?.message) {
                const state = useAppStore.getState();
                const alreadyHas = state.novelclawMessages.some(m => m.id === data.message.id);
                if (!alreadyHas) {
                  useAppStore.setState({
                    novelclawMessages: [...state.novelclawMessages, data.message],
                  });
                }
              }
              useAppStore.getState().fetchNovelClawMessages(data?.thread_id);
            } else if (type === 'thread_cleared') {
              useAppStore.setState({ novelclawMessages: [] });
            }
          })
        );
        unsubs.push(
          Events.On('translation:ask_human', (e: any) => {
            const data = e?.data;
            if (data) {
              useAppStore.setState({
                pendingAskHuman: {
                  jobId: data.job_id,
                  question: data.question,
                  options: data.options || [],
                  contextSnippet: data.context_snippet || '',
                  timeoutSeconds: data.timeout_seconds || 60,
                  createdAt: Date.now(),
                },
              });
            }
          })
        );
        unsubs.push(
          Events.On('translation:ask_human_resolved', (e: any) => {
            const data = e?.data;
            const current = useAppStore.getState().pendingAskHuman;
            if (current && (!data?.job_id || current.jobId === data.job_id)) {
              useAppStore.setState({ pendingAskHuman: null });
            }
          })
        );
        unsubs.push(
          Events.On('soul:changed', (e: any) => {
            const data = e?.data;
            if (data?.soul) {
              useAppStore.setState({ soul: data.soul });
            }
          })
        );
        unsubs.push(
          Events.On('skill:updated', (e: any) => {
            const data = e?.data;
            const currentProj = useAppStore.getState().selectedProject?.id || 'default';
            if (!data?.project_id || data.project_id === currentProj) {
              useAppStore.getState().fetchSkills(currentProj);
            }
          })
        );
        unsubs.push(
          Events.On('skill:preset_applied', (e: any) => {
            const data = e?.data;
            const currentProj = useAppStore.getState().selectedProject?.id || 'default';
            if (!data?.project_id || data.project_id === currentProj) {
              useAppStore.getState().fetchSkills(currentProj);
            }
          })
        );
        unsubs.push(
          Events.On('worldbible:category_updated', () => {
            useAppStore.getState().fetchWorldCategories();
          })
        );
        unsubs.push(
          Events.On('worldbible:category_deleted', () => {
            useAppStore.getState().fetchWorldCategories();
            useAppStore.getState().fetchWorldEntries();
          })
        );
        unsubs.push(
          Events.On('worldbible:entry_updated', () => {
            useAppStore.getState().fetchWorldEntries();
          })
        );
        unsubs.push(
          Events.On('worldbible:entry_deleted', () => {
            useAppStore.getState().fetchWorldEntries();
          })
        );
        unsubs.push(
          Events.On('worldbible:entry_verified', () => {
            useAppStore.getState().fetchWorldEntries();
          })
        );
        unsubs.push(
          Events.On('worldbible:scanned', () => {
            useAppStore.getState().fetchWorldCategories();
            useAppStore.getState().fetchWorldEntries();
          })
        );
        unsubs.push(
          Events.On('worldbible:imported', () => {
            useAppStore.getState().fetchWorldCategories();
            useAppStore.getState().fetchWorldEntries();
          })
        );
        unsubs.push(
          Events.On('auditor:plot_warning', (e: any) => {
            const data = e?.data;
            if (data?.warnings && Array.isArray(data.warnings)) {
              data.warnings.forEach((w: any) => {
                toast.error(w.message || t('tools.plotWarning'), {
                  duration: 6000,
                });
              });
            }
          })
        );
      } catch (err) {
        console.warn('Events.On registration error:', err);
      }
    }

    return () => {
      unsubs.forEach((u) => {
        try {
          if (u) u();
        } catch {
          // ignore cleanup errors
        }
      });
    };
  }, [initializeApp, t]);

  // Loading screen
  if (isAppLoading) {
    return <LoadingScreen />;
  }

  // First-time setup wizard
  if (isFirstRun) {
    return (
      <>
        <SetupWizard />
        <Toaster
          position="top-right"
          toastOptions={{
            duration: 3500,
            style: {
              background: '#0e1320',
              color: '#f8fafc',
              border: '1px solid rgba(255, 255, 255, 0.12)',
              boxShadow: '0 10px 25px -5px rgba(0, 0, 0, 0.5), 0 8px 10px -6px rgba(0, 0, 0, 0.5)',
              borderRadius: '0.75rem',
              fontSize: '0.8125rem',
              fontWeight: 500,
            },
            success: { iconTheme: { primary: '#10b981', secondary: '#ffffff' } },
            error: { iconTheme: { primary: '#f43f5e', secondary: '#ffffff' } },
          }}
        />
      </>
    );
  }

  return (
    <div className="flex flex-col h-screen w-screen bg-[#0a0d14] text-slate-100 select-none overflow-hidden font-sans">
      <Navbar />

      <main className="flex-1 flex overflow-hidden min-h-0">
        {activeTab === 'workspace' && <CoworkerWorkspace />}
        {activeTab === 'graph' && <CharacterGraph />}
        {activeTab === 'glossary' && <GlossaryHub />}
        {activeTab === 'world_bible' && <WorldBibleView />}
        {activeTab === 'benchmark' && <BenchmarkDashboard />}
      </main>

      {/* Modals */}
      <StyleScoutModal />
      <ExportModal />
      <ImportModal />
      <SettingsModal />
      <ProjectManagerModal />
      <ConfirmModal />

      {/* Global Notifications */}
      <Toaster
        position="top-right"
        toastOptions={{
          duration: 3500,
          style: {
            background: '#0e1320',
            color: '#f8fafc',
            border: '1px solid rgba(255, 255, 255, 0.12)',
            boxShadow: '0 10px 25px -5px rgba(0, 0, 0, 0.5), 0 8px 10px -6px rgba(0, 0, 0, 0.5)',
            borderRadius: '0.75rem',
            fontSize: '0.8125rem',
            fontWeight: 500,
          },
          success: {
            iconTheme: {
              primary: '#10b981',
              secondary: '#ffffff',
            },
          },
          error: {
            iconTheme: {
              primary: '#f43f5e',
              secondary: '#ffffff',
            },
          },
        }}
      />
    </div>
  );
};
