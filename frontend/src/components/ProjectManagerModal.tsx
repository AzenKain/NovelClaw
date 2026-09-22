import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'react-hot-toast';
import {
  FolderKanban,
  Edit2,
  Trash2,
  Check,
  X,
  ExternalLink,
  Plus,
  BookOpen,
  AlertTriangle,
  ArrowRight,
  Share2
} from 'lucide-react';
import * as FSService from '@bindings/novelclaw/internal/services/fsservice';
import { useAppStore } from '@/store/useAppStore';

export const ProjectManagerModal: React.FC = () => {
  const { t } = useTranslation();
  const {
    isProjectManagerOpen,
    setProjectManagerOpen,
    projects,
    selectedProject,
    selectProject,
    deleteProject,
    renameProject,
    inheritKnowledge,
    appendBookToProject,
    setImportOpen,
  } = useAppStore();

  const [inheritTargetId, setInheritTargetId] = useState<string | null>(null);
  const [inheritSourceId, setInheritSourceId] = useState<string>('');
  const [isInheriting, setIsInheriting] = useState(false);
  const [inheritSuccessMsg, setInheritSuccessMsg] = useState<string | null>(null);
  const [editingId, setEditingId] = useState<string | null>(null);
  const [editingTitle, setEditingTitle] = useState('');
  const [confirmDeleteId, setConfirmDeleteId] = useState<string | null>(null);
  const [isDeleting, setIsDeleting] = useState(false);
  const [isRenaming, setIsRenaming] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);

  if (!isProjectManagerOpen) return null;

  const handleStartRename = (id: string, currentTitle: string) => {
    setEditingId(id);
    setEditingTitle(currentTitle);
    setErrorMsg(null);
  };

  const handleSaveRename = async (id: string) => {
    if (!editingTitle.trim()) {
      toast.error(t('projectManager.nameRequired', 'Project title cannot be empty'));
      return;
    }
    setIsRenaming(true);
    setErrorMsg(null);
    try {
      await renameProject(id, editingTitle.trim());
      setEditingId(null);
      toast.success(t('projectManager.renameSuccess', 'Project name updated successfully'));
    } catch (err: any) {
      const msg = err?.message || t('projectManager.renameError', 'Failed to rename project');
      setErrorMsg(msg);
      toast.error(msg);
    } finally {
      setIsRenaming(false);
    }
  };

  const handleDelete = async (id: string) => {
    setIsDeleting(true);
    setErrorMsg(null);
    try {
      await deleteProject(id);
      setConfirmDeleteId(null);
      toast.success(t('projectManager.deleteSuccess', 'Project deleted successfully'));
    } catch (err: any) {
      const msg = err?.message || t('projectManager.deleteError', 'Failed to delete project');
      setErrorMsg(msg);
      toast.error(msg);
    } finally {
      setIsDeleting(false);
    }
  };

  const handleInheritKnowledge = async () => {
    if (!inheritTargetId || !inheritSourceId) return;
    setIsInheriting(true);
    setErrorMsg(null);
    setInheritSuccessMsg(null);
    try {
      const res = await inheritKnowledge(inheritSourceId, inheritTargetId);
      if (res) {
        const msg = t('projectManager.inheritSuccess', {
          defaultValue: `Knowledge inheritance completed! Copied ${res.entities_count} characters, ${res.relations_count} relations and ${res.glossary_count} terms.`,
          entities: res.entities_count,
          relations: res.relations_count,
          glossary: res.glossary_count,
        });
        setInheritSuccessMsg(msg);
        toast.success(msg);
        setTimeout(() => {
          setInheritTargetId(null);
          setInheritSuccessMsg(null);
        }, 2000);
      }
    } catch (err: any) {
      const msg = err?.message || t('projectManager.inheritError', 'Failed to inherit project knowledge');
      setErrorMsg(msg);
      toast.error(msg);
    } finally {
      setIsInheriting(false);
    }
  };

  const handleSelect = (project: typeof projects[0]) => {
    selectProject(project);
    setProjectManagerOpen(false);
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4 bg-black/80">
      <div className="bg-[#0e1320] border border-white/10rounded-2xl w-full max-w-2xl overflow-hidden shadow-2xl flex flex-col max-h-[85vh]">
        {/* Header */}
        <div className="p-4 md:p-5 border-b border-white/8 flex items-center justify-between bg-[#111728] shrink-0 gap-4">
          <div className="flex items-center space-x-3 min-w-0 flex-1">
            <div className="w-9 h-9 rounded-xl bg-indigo-500/15 border border-indigo-500/30 flex items-center justify-center text-indigo-400 shrink-0">
              <FolderKanban className="w-5 h-5" />
            </div>
            <div className="min-w-0 flex-1">
              <h3 className="text-sm md:text-base font-bold text-white truncate">{t('projectManager.title')}</h3>
              <p className="text-xs text-slate-400 leading-relaxed pr-2">{t('projectManager.subtitle')}</p>
            </div>
          </div>

          <div className="flex items-center space-x-2 shrink-0 ml-2">
            <button
              onClick={() => {
                setProjectManagerOpen(false);
                setImportOpen(true);
              }}
              className="flex items-center space-x-1.5 px-3.5 py-1.5 rounded-lg bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-xs transition shrink-0"
            >
              <Plus className="w-3.5 h-3.5" />
              <span>{t('projectManager.importBtn')}</span>
            </button>
            <button
              onClick={() => setProjectManagerOpen(false)}
              className="p-1.5 rounded-lg text-slate-400 hover:text-white hover:bg-white/6 transition shrink-0"
              title={t('common.close')}
            >
              <X className="w-5 h-5" />
            </button>
          </div>
        </div>

        {/* Error Alert */}
        {errorMsg && (
          <div className="m-4 p-3 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-300 text-xs flex items-center space-x-2">
            <AlertTriangle className="w-4 h-4 shrink-0" />
            <span>{errorMsg}</span>
          </div>
        )}

        {/* Project List */}
        <div className="p-4 md:p-5 overflow-y-auto flex-1 space-y-2.5">
          {projects.length === 0 ? (
            <div className="py-12 text-center flex flex-col items-center justify-center">
              <div className="w-14 h-14 rounded-2xl bg-white/3 border border-white/8 flex items-center justify-center text-slate-500 mb-3">
                <BookOpen className="w-7 h-7" />
              </div>
              <p className="text-sm font-semibold text-slate-300 mb-1">{t('projectManager.emptyProjects')}</p>
              <p className="text-xs text-slate-500 max-w-sm mb-4">
                {t('projects.selectProject')}
              </p>
              <button
                onClick={() => {
                  setProjectManagerOpen(false);
                  setImportOpen(true);
                }}
                className="flex items-center space-x-1.5 px-4 py-2 rounded-xl bg-indigo-600 hover:bg-indigo-500 text-white text-xs font-semibold shadow-sm transition"
              >
                <Plus className="w-4 h-4" />
                <span>{t('projectManager.importBtn')}</span>
              </button>
            </div>
          ) : (
            projects.map((p) => {
              const isSelected = selectedProject?.id === p.id;
              const isEditing = editingId === p.id;
              const isConfirmingDelete = confirmDeleteId === p.id;

              return (
                <div
                  key={p.id}
                  className={`p-3.5 rounded-xl border transition flex flex-col sm:flex-row sm:items-center justify-between gap-3 ${
                    isSelected
                      ? 'bg-indigo-500/[0.07] border-indigo-500/30 ring-1 ring-indigo-500/20'
                      : 'bg-[#121826] border-white/6 hover:border-white/15'
                  }`}
                >
                  {/* Left info */}
                  <div className="flex-1 min-w-0">
                    {isEditing ? (
                      <div className="flex items-center space-x-2">
                        <input
                          type="text"
                          value={editingTitle}
                          onChange={(e) => setEditingTitle(e.target.value)}
                          onKeyDown={(e) => {
                            if (e.key === 'Enter') handleSaveRename(p.id);
                            if (e.key === 'Escape') setEditingId(null);
                          }}
                          autoFocus
                          className="flex-1 px-2.5 py-1 text-xs bg-[#0a0d16] border border-indigo-500/50 rounded-lg text-white font-medium focus:outline-none focus:ring-1 focus:ring-indigo-500"
                        />
                        <button
                          onClick={() => handleSaveRename(p.id)}
                          disabled={isRenaming}
                          className="p-1 rounded-md bg-emerald-600 hover:bg-emerald-500 text-white transition shrink-0"
                          title={t('common.save')}
                        >
                          <Check className="w-3.5 h-3.5" />
                        </button>
                        <button
                          onClick={() => setEditingId(null)}
                          className="p-1 rounded-md bg-white/6 hover:bg-white/10 text-slate-300 transition shrink-0"
                          title={t('common.cancel')}
                        >
                          <X className="w-3.5 h-3.5" />
                        </button>
                      </div>
                    ) : (
                      <div className="flex items-center space-x-2">
                        <h4 className="text-xs sm:text-sm font-bold text-slate-100 truncate">
                          {p.title}
                        </h4>
                        {isSelected && (
                          <span className="px-1.5 py-0.2 rounded bg-indigo-500/20 text-indigo-300 text-[10px] font-medium shrink-0 border border-indigo-500/30">
                            {t('settings.activeBadge')}
                          </span>
                        )}
                      </div>
                    )}

                    <div className="flex items-center gap-2 mt-1.5 text-[11px] text-slate-400 flex-wrap">
                      {p.author && (
                        <span className="text-slate-300 font-medium truncate max-w-37.5">
                          {p.author}
                        </span>
                      )}
                      <span className="px-1.5 py-0.2 rounded bg-white/4 text-slate-300 border border-white/8 font-mono">
                        {p.original_format ? p.original_format.toUpperCase() : 'EPUB'}
                      </span>
                      <span className="px-1.5 py-0.2 rounded bg-white/4 text-slate-300 border border-white/8">
                        {p.total_chapters} {t('projectManager.chaptersCount')}
                      </span>
                      <div className="flex items-center space-x-1 text-indigo-300 font-mono font-semibold">
                        <span>{(p.source_lang || 'ja').toUpperCase()}</span>
                        <ArrowRight className="w-2.5 h-2.5 text-slate-500" />
                        <span>{(p.target_lang || 'en').toUpperCase()}</span>
                      </div>
                    </div>
                  </div>

                  {/* Right Actions */}
                  <div className="flex items-center space-x-1.5 shrink-0 self-end sm:self-center">
                    {isConfirmingDelete ? (
                      <div className="flex items-center space-x-1 bg-rose-500/15 border border-rose-500/30 rounded-lg p-1">
                        <span className="text-[11px] text-rose-300 font-medium px-1">{t('projectManager.delete')}?</span>
                        <button
                          onClick={() => handleDelete(p.id)}
                          disabled={isDeleting}
                          className="px-2 py-0.5 rounded bg-rose-600 hover:bg-rose-500 text-white text-xs font-semibold transition shrink-0"
                        >
                          {isDeleting ? '...' : t('projectManager.delete')}
                        </button>
                        <button
                          onClick={() => setConfirmDeleteId(null)}
                          className="p-1 rounded text-slate-300 hover:text-white transition shrink-0"
                          title={t('common.cancel')}
                        >
                          <X className="w-3 h-3" />
                        </button>
                      </div>
                    ) : (
                      <>
                        {!isSelected && (
                          <button
                            onClick={() => handleSelect(p)}
                            className="flex items-center space-x-1 px-2.5 py-1.5 rounded-lg bg-white/4 hover:bg-white/8 text-slate-200 border border-white/10 text-xs font-medium transition shrink-0"
                            title={t('projectManager.openProject')}
                          >
                            <ExternalLink className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
                            <span>{t('common.actions')}</span>
                          </button>
                        )}

                        <button
                          onClick={async () => {
                            try {
                              const path = await FSService.PickBookFile();
                              if (path) {
                                const clean = path.replace(/\\/g, '/');
                                const filename = clean.substring(clean.lastIndexOf('/') + 1);
                                const base = filename.substring(0, filename.lastIndexOf('.')) || filename;
                                await appendBookToProject(p.id, path, base);
                                toast.success(`${t('common.success')}: "${base}"`);
                              }
                            } catch (err: any) {
                              const msg = err?.message || t('common.error', 'Error');
                              setErrorMsg(msg);
                              toast.error(msg);
                            }
                          }}
                          className="flex items-center space-x-1 px-2 py-1.5 rounded-lg bg-indigo-600/15 hover:bg-indigo-600/25 text-indigo-300 border border-indigo-500/30 text-xs font-medium transition shrink-0"
                          title={t('projectManager.seriesTag')}
                        >
                          <Plus className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
                          <span>{t('projectManager.seriesTag')}</span>
                        </button>

                        <button
                          onClick={() => {
                            setInheritTargetId(p.id);
                            const otherProjects = projects.filter(pr => pr.id !== p.id);
                            setInheritSourceId(otherProjects[0]?.id || '');
                            setErrorMsg(null);
                            setInheritSuccessMsg(null);
                          }}
                          className="p-1.5 rounded-lg bg-indigo-500/10 hover:bg-indigo-500/20 text-indigo-300 border border-indigo-500/20 transition shrink-0"
                          title={t('projectManager.inheritTitle')}
                        >
                          <Share2 className="w-3.5 h-3.5" />
                        </button>

                        <button
                          onClick={() => handleStartRename(p.id, p.title)}
                          className="p-1.5 rounded-lg bg-white/4 hover:bg-white/8 text-slate-300 hover:text-white border border-white/10 transition shrink-0"
                          title={t('projectManager.rename')}
                        >
                          <Edit2 className="w-3.5 h-3.5" />
                        </button>

                        <button
                          onClick={() => setConfirmDeleteId(p.id)}
                          className="p-1.5 rounded-lg bg-rose-500/10 hover:bg-rose-500/20 text-rose-300 border border-rose-500/20 transition shrink-0"
                          title={t('projectManager.delete')}
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </button>
                      </>
                    )}
                  </div>
                </div>
              );
            })
          )}
        </div>

        {/* Inherit Knowledge Dialog */}
        {inheritTargetId && (() => {
          const otherProjects = projects.filter(pr => pr.id !== inheritTargetId);
          const targetProj = projects.find(pr => pr.id === inheritTargetId);

          return (
            <div className="p-4 border-t border-indigo-500/30 bg-[#0e1424] text-xs space-y-3">
              <div className="flex items-center justify-between">
                <h4 className="font-bold text-indigo-300 flex items-center gap-1.5">
                  <Share2 className="w-4 h-4" />
                  <span>{t('projectManager.inheritTitle')}</span>
                </h4>
                <button
                  onClick={() => setInheritTargetId(null)}
                  className="text-slate-400 hover:text-white"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>

              <p className="text-slate-300 text-[11px] leading-relaxed">
                {t('projectManager.inheritDesc')} <strong>{targetProj?.title}</strong>.
              </p>

              {otherProjects.length === 0 ? (
                <div className="p-3 rounded-xl bg-amber-500/10 border border-amber-500/30 text-amber-200 text-xs space-y-2">
                  <div className="flex items-center space-x-2 font-semibold text-amber-300">
                    <AlertTriangle className="w-4 h-4 shrink-0" />
                    <span>{t('projectManager.emptyProjects')}</span>
                  </div>
                  <p className="text-[11px] text-amber-200/80 leading-relaxed">
                    {t('projectManager.inheritDescMin2')}
                  </p>
                  <button
                    onClick={() => {
                      setProjectManagerOpen(false);
                      setImportOpen(true);
                    }}
                    className="inline-flex items-center space-x-1.5 px-3 py-1.5 rounded-lg bg-amber-600 hover:bg-amber-500 text-white font-medium text-xs transition mt-1 shadow-sm"
                  >
                    <Plus className="w-3.5 h-3.5" />
                    <span>{t('projectManager.importBtn')}</span>
                  </button>
                </div>
              ) : (
                <div className="flex items-center gap-2">
                  <label className="text-slate-400 whitespace-nowrap">{t('projectManager.selectSourceProj')}</label>
                  <select
                    value={inheritSourceId}
                    onChange={(e) => setInheritSourceId(e.target.value)}
                    className="flex-1 px-3 py-1.5 bg-[#0a0d14] border border-white/10 rounded-lg text-slate-200 text-xs focus:outline-none focus:border-indigo-500"
                  >
                    {otherProjects.map(pr => (
                      <option key={pr.id} value={pr.id}>
                        {pr.title} ({pr.total_chapters} {t('projectManager.chaptersCount')})
                      </option>
                    ))}
                  </select>

                  <button
                    onClick={handleInheritKnowledge}
                    disabled={isInheriting || !inheritSourceId}
                    className="px-3.5 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white font-semibold rounded-lg shadow-sm transition shrink-0 disabled:opacity-50"
                  >
                    {isInheriting ? t('projectManager.inheritingBtn') : t('projectManager.startInheritBtn')}
                  </button>
                </div>
              )}

              {inheritSuccessMsg && (
                <div className="p-2.5 bg-emerald-500/15 border border-emerald-500/30 rounded-lg text-emerald-300 text-xs flex items-center space-x-2">
                  <Check className="w-4 h-4 shrink-0" />
                  <span>{inheritSuccessMsg}</span>
                </div>
              )}
            </div>
          );
        })()}

        {/* Footer */}
        <div className="p-3.5 border-t border-white/8 bg-[#0c101c] flex items-center justify-between text-xs text-slate-400 shrink-0">
          <span>{t('common.all')}: <strong className="text-slate-200">{projects.length}</strong> {t('nav.manageProjects')}</span>
          <button
            onClick={() => setProjectManagerOpen(false)}
            className="px-3 py-1 rounded-lg bg-white/5 hover:bg-white/10 text-slate-300 transition"
          >
            {t('common.close')}
          </button>
        </div>
      </div>
    </div>
  );
};
