import React, { useState, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'react-hot-toast';
import {
  Globe,
  Plus,
  Search,
  Trash2,
  Edit2,
  Download,
  Upload,
  Sparkles,
  CheckCircle2,
  Clock,
  Cpu,
  Crown,
  Wand2,
  Building2,
  Sword,
  Shield,
  Skull,
  BookOpen,
  Key,
  Layers,
  Boxes,
  X,
  RefreshCw,
  Copy,
  Check,
} from 'lucide-react';
import * as dtos from '@bindings/novelclaw/internal/dtos/models';
import { useAppStore } from '@/store/useAppStore';
import { showConfirmModal } from '@/store/confirmStore';

// Map icon names to Lucide icons
const ICON_MAP: Record<string, React.FC<{ className?: string }>> = {
  Cpu,
  Crown,
  Wand2,
  Building2,
  Sword,
  Shield,
  Skull,
  BookOpen,
  Key,
  Layers,
  Boxes,
  Globe,
  Sparkles,
};

const renderCategoryIcon = (iconName?: string, className = 'w-4 h-4') => {
  const IconComponent = iconName && ICON_MAP[iconName] ? ICON_MAP[iconName] : Boxes;
  return <IconComponent className={className} />;
};

export const WorldBibleView: React.FC = () => {
  const { t } = useTranslation();
  const {
    selectedProject,
    worldCategories,
    worldEntries,
    selectedWorldCategory,
    isWorldBibleLoading,
    isScanningWorld,
    worldScanReport,
    setSelectedWorldCategory,
    upsertWorldCategory,
    deleteWorldCategory,
    upsertWorldEntry,
    verifyWorldEntry,
    deleteWorldEntry,
    scanWorldBible,
    exportWorldMarkdown,
    exportWorldJSON,
    importWorldJSON,
  } = useAppStore();

  const [searchQuery, setSearchQuery] = useState('');
  const [filterVerified, setFilterVerified] = useState<'all' | 'verified' | 'ai_scan'>('all');

  // Modal: Entry
  const [isEntryModalOpen, setIsEntryModalOpen] = useState(false);
  const [editingEntry, setEditingEntry] = useState<dtos.WorldEntryDTO | null>(null);
  const [entryName, setEntryName] = useState('');
  const [entryCategory, setEntryCategory] = useState('');
  const [entryAliases, setEntryAliases] = useState('');
  const [entrySummary, setEntrySummary] = useState('');
  const [entryDescription, setEntryDescription] = useState('');
  const [entryAttrs, setEntryAttrs] = useState<{ key: string; value: string }[]>([]);
  const [entryVerified, setEntryVerified] = useState(true);

  // Modal: Category
  const [isCategoryModalOpen, setIsCategoryModalOpen] = useState(false);
  const [editingCategory, setEditingCategory] = useState<dtos.WorldCategoryDTO | null>(null);
  const [categoryName, setCategoryName] = useState('');
  const [categorySlug, setCategorySlug] = useState('');
  const [categoryIcon, setCategoryIcon] = useState('Boxes');
  const [categoryDesc, setCategoryDesc] = useState('');

  // Modal: Auto-Scan
  const [isScanModalOpen, setIsScanModalOpen] = useState(false);
  const [scanStartChap, setScanStartChap] = useState(1);
  const [scanEndChap, setScanEndChap] = useState(5);

  // Modal: Import / Export
  const [isExportModalOpen, setIsExportModalOpen] = useState(false);
  const [isImportModalOpen, setIsImportModalOpen] = useState(false);
  const [exportFormat, setExportFormat] = useState<'markdown' | 'json'>('markdown');
  const [exportText, setExportText] = useState('');
  const [importText, setImportText] = useState('');
  const [hasCopied, setHasCopied] = useState(false);

  // Filtered entries
  const filteredEntries = useMemo(() => {
    return worldEntries.filter((entry) => {
      // Category filter
      if (selectedWorldCategory && entry.category_id !== selectedWorldCategory) {
        return false;
      }
      // Verification status filter
      if (filterVerified === 'verified' && !entry.is_verified) {
        return false;
      }
      if (filterVerified === 'ai_scan' && entry.is_verified) {
        return false;
      }
      // Search filter
      const q = searchQuery.toLowerCase().trim();
      if (q) {
        const nameMatch = entry.name.toLowerCase().includes(q);
        const aliasMatch = entry.aliases && entry.aliases.some((a) => a.toLowerCase().includes(q));
        const summaryMatch = entry.summary.toLowerCase().includes(q);
        const descMatch = entry.full_description?.toLowerCase().includes(q);
        if (!nameMatch && !aliasMatch && !summaryMatch && !descMatch) {
          return false;
        }
      }
      return true;
    });
  }, [worldEntries, selectedWorldCategory, filterVerified, searchQuery]);

  // Handle open entry modal
  const handleOpenAddEntry = () => {
    if (!selectedProject) {
      toast.error(t('worldBible.projectRequired'));
      return;
    }
    setEditingEntry(null);
    setEntryName('');
    setEntryCategory(selectedWorldCategory || (worldCategories[0]?.id ?? ''));
    setEntryAliases('');
    setEntrySummary('');
    setEntryDescription('');
    setEntryAttrs([]);
    setEntryVerified(true);
    setIsEntryModalOpen(true);
  };

  const handleOpenEditEntry = (entry: dtos.WorldEntryDTO) => {
    setEditingEntry(entry);
    setEntryName(entry.name);
    setEntryCategory(entry.category_id);
    setEntryAliases(entry.aliases ? entry.aliases.join(', ') : '');
    setEntrySummary(entry.summary);
    setEntryDescription(entry.full_description || '');
    const attrs = entry.attributes
      ? Object.entries(entry.attributes).map(([k, v]) => ({ key: k, value: String(v) }))
      : [];
    setEntryAttrs(attrs);
    setEntryVerified(entry.is_verified);
    setIsEntryModalOpen(true);
  };

  const handleSaveEntry = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedProject) return;
    if (!entryName.trim()) {
      toast.error(t('worldBible.entityNameRequired'));
      return;
    }
    if (!entryCategory) {
      toast.error(t('worldBible.categoryRequired'));
      return;
    }

    const attrsMap: Record<string, string> = {};
    entryAttrs.forEach((item) => {
      if (item.key.trim() && item.value.trim()) {
        attrsMap[item.key.trim()] = item.value.trim();
      }
    });

    const aliases = entryAliases
      .split(',')
      .map((a) => a.trim())
      .filter(Boolean);

    const req: dtos.UpsertWorldEntryRequest = {
      id: editingEntry?.id,
      project_id: selectedProject.id,
      category_id: entryCategory,
      name: entryName.trim(),
      aliases,
      summary: entrySummary.trim(),
      full_description: entryDescription.trim(),
      attributes: attrsMap,
      source_chapter_index: editingEntry?.source_chapter_index || 1,
      is_verified: entryVerified,
    };

    const res = await upsertWorldEntry(req);
    if (res) {
      toast.success(editingEntry ? t('worldBible.entryUpdated') : t('worldBible.entryCreated'));
      setIsEntryModalOpen(false);
    } else {
      toast.error(t('worldBible.saveFailed'));
    }
  };

  const handleDeleteEntry = async (id: string, name: string) => {
    const confirmed = await showConfirmModal({
      title: t('worldBible.deleteEntityTitle'),
      message: t('worldBible.deleteEntityMsg', { name }),
      confirmText: t('worldBible.deleteEntityConfirm'),
      cancelText: t('common.cancel'),
      variant: 'danger',
    });
    if (!confirmed) return;
    await deleteWorldEntry(id);
    toast.success(t('worldBible.deleted', { name }));
  };

  const handleToggleVerify = async (entry: dtos.WorldEntryDTO) => {
    await verifyWorldEntry(entry.id, !entry.is_verified);
    toast.success(entry.is_verified ? t('worldBible.toggledPending') : t('worldBible.toggledVerified'));
  };

  // Handle open category modal
  const handleOpenAddCategory = () => {
    setEditingCategory(null);
    setCategoryName('');
    setCategorySlug('');
    setCategoryIcon('Boxes');
    setCategoryDesc('');
    setIsCategoryModalOpen(true);
  };

  const handleOpenEditCategory = (cat: dtos.WorldCategoryDTO) => {
    setEditingCategory(cat);
    setCategoryName(cat.name);
    setCategorySlug(cat.slug);
    setCategoryIcon(cat.icon || 'Boxes');
    setCategoryDesc(cat.description || '');
    setIsCategoryModalOpen(true);
  };

  const handleSaveCategory = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedProject) return;
    if (!categoryName.trim()) {
      toast.error(t('worldBible.categoryNameRequired'));
      return;
    }

    const slug =
      categorySlug.trim() ||
      categoryName
        .toLowerCase()
        .trim()
        .replace(/\s+/g, '_')
        .replace(/[^a-z0-9_]/g, '');

    const req: dtos.UpsertWorldCategoryRequest = {
      id: editingCategory?.id,
      project_id: selectedProject.id,
      name: categoryName.trim(),
      slug,
      icon: categoryIcon,
      description: categoryDesc.trim(),
      display_order: editingCategory?.display_order ?? worldCategories.length,
    };

    const res = await upsertWorldCategory(req);
    if (res) {
      toast.success(editingCategory ? t('worldBible.categoryUpdated') : t('worldBible.categoryCreated'));
      setIsCategoryModalOpen(false);
    } else {
      toast.error(t('worldBible.saveCatFailed'));
    }
  };

  const handleDeleteCategory = async (cat: dtos.WorldCategoryDTO) => {
    const confirmed = await showConfirmModal({
      title: t('worldBible.deleteCatTitle'),
      message: t('worldBible.deleteCatMsg', { name: cat.name }),
      confirmText: t('worldBible.deleteCatConfirm'),
      cancelText: t('common.cancel'),
      variant: 'danger',
    });
    if (!confirmed) return;
    await deleteWorldCategory(cat.id);
    toast.success(t('worldBible.deletedCat', { name: cat.name }));
    if (selectedWorldCategory === cat.id) {
      setSelectedWorldCategory(null);
    }
  };

  // Handle Auto-Scan
  const handleStartScan = async () => {
    if (!selectedProject) return;
    const res = await scanWorldBible(scanStartChap, scanEndChap, selectedProject.id);
    if (res && res.success) {
      toast.success(
        t('worldBible.scanSuccess', { cats: res.categories_added, entries: res.entries_added })
      );
    } else {
      toast.error(t('worldBible.scanFailed'));
    }
  };

  // Export handlers
  const handleOpenExport = async (format: 'markdown' | 'json') => {
    setExportFormat(format);
    setHasCopied(false);
    setIsExportModalOpen(true);
    if (format === 'markdown') {
      const text = await exportWorldMarkdown();
      setExportText(text);
    } else {
      const text = await exportWorldJSON();
      setExportText(text);
    }
  };

  const handleCopyExport = () => {
    navigator.clipboard.writeText(exportText);
    setHasCopied(true);
    toast.success(t('worldBible.copied'));
    setTimeout(() => setHasCopied(false), 2000);
  };

  const handleDownloadExport = () => {
    const filename =
      exportFormat === 'markdown'
        ? `${selectedProject?.title || 'world'}_BIBLE.md`
        : `${selectedProject?.title || 'world'}_BIBLE.json`;
    const blob = new Blob([exportText], {
      type: exportFormat === 'markdown' ? 'text/markdown' : 'application/json',
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    URL.revokeObjectURL(url);
    toast.success(t('worldBible.downloaded', { name: filename }));
  };

  // Import handler
  const handleImport = async () => {
    if (!importText.trim()) {
      toast.error(t('worldBible.jsonRequired'));
      return;
    }
    const [cats, entries] = await importWorldJSON(importText);
    if (cats > 0 || entries > 0) {
      toast.success(t('worldBible.importSuccess', { cats, entries }));
      setIsImportModalOpen(false);
      setImportText('');
    } else {
      toast.error(t('worldBible.importFailed'));
    }
  };

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#090d16] text-slate-200 select-text overflow-hidden">
      {/* Top Header & Action Toolbar */}
      <div className="h-14 border-b border-white/8 px-4 flex items-center justify-between gap-3 bg-[#0c111d]/90 backdrop-blur-md shrink-0">
        {/* Left Search & Filters */}
        <div className="flex items-center gap-3 flex-1 min-w-0">
          <div className="relative w-64 max-w-full">
            <Search className="w-3.5 h-3.5 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2" />
            <input
              type="text"
              placeholder={t('worldBible.searchPlaceholder')}
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              className="w-full pl-8 pr-7 py-1.5 bg-white/4 border border-white/10 rounded-lg text-xs text-slate-200 placeholder-slate-500 focus:outline-none focus:border-indigo-500/50"
            />
            {searchQuery && (
              <button
                onClick={() => setSearchQuery('')}
                className="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-400 hover:text-slate-200"
              >
                <X className="w-3 h-3" />
              </button>
            )}
          </div>

          {/* Verification Segmented Filter */}
          <div className="hidden sm:flex items-center bg-white/4 border border-white/10 rounded-lg p-0.5 text-xs">
            <button
              onClick={() => setFilterVerified('all')}
              className={`px-2.5 py-1 rounded-md transition-colors ${
                filterVerified === 'all'
                  ? 'bg-indigo-600 text-white font-medium shadow-xs'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              {t('worldBible.allTab')} ({worldEntries.length})
            </button>
            <button
              onClick={() => setFilterVerified('verified')}
              className={`px-2.5 py-1 rounded-md transition-colors flex items-center gap-1 ${
                filterVerified === 'verified'
                  ? 'bg-emerald-600 text-white font-medium shadow-xs'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <CheckCircle2 className="w-3 h-3 text-emerald-300" />
              {t('worldBible.verifiedTab')} ({worldEntries.filter((e) => e.is_verified).length})
            </button>
            <button
              onClick={() => setFilterVerified('ai_scan')}
              className={`px-2.5 py-1 rounded-md transition-colors flex items-center gap-1 ${
                filterVerified === 'ai_scan'
                  ? 'bg-amber-600 text-white font-medium shadow-xs'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <Sparkles className="w-3 h-3 text-amber-300" />
              {t('worldBible.proposedTab')} ({worldEntries.filter((e) => !e.is_verified).length})
            </button>
          </div>
        </div>

        {/* Right CTA Actions */}
        <div className="flex items-center gap-2 shrink-0">
          <button
            onClick={() => setIsScanModalOpen(true)}
            disabled={isScanningWorld}
            className="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-medium flex items-center gap-1.5 shadow-sm transition-colors disabled:opacity-50"
            title={t('worldBible.scanEntitiesTooltip')}
          >
            {isScanningWorld ? (
              <RefreshCw className="w-3.5 h-3.5 animate-spin" />
            ) : (
              <Sparkles className="w-3.5 h-3.5 text-amber-300" />
            )}
            <span className="hidden md:inline">{t('worldBible.scanBtn')}</span>
          </button>

          <button
            onClick={handleOpenAddEntry}
            className="px-3 py-1.5 bg-white/4 hover:bg-white/8 text-slate-200 border border-white/10 rounded-lg text-xs font-medium flex items-center gap-1.5 transition-colors"
          >
            <Plus className="w-3.5 h-3.5 text-indigo-400" />
            <span>{t('worldBible.addEntryBtn')}</span>
          </button>

          <div className="h-4 w-px bg-white/10 mx-1 hidden sm:block" />

          {/* Export / Import button */}
          <div className="flex items-center gap-1">
            <button
              onClick={() => handleOpenExport('markdown')}
              className="p-1.5 bg-white/4 hover:bg-white/8 text-slate-300 border border-white/10 rounded-lg text-xs transition-colors"
              title={t('worldBible.exportMdTooltip')}
            >
              <Download className="w-3.5 h-3.5" />
            </button>
            <button
              onClick={() => setIsImportModalOpen(true)}
              className="p-1.5 bg-white/4 hover:bg-white/8 text-slate-300 border border-white/10 rounded-lg text-xs transition-colors"
              title={t('worldBible.importJsonTooltip')}
            >
              <Upload className="w-3.5 h-3.5" />
            </button>
          </div>
        </div>
      </div>

      {/* Main Studio Body: Sidebar Categories & Entries Grid */}
      <div className="flex-1 flex min-h-0 overflow-hidden">
        {/* Left Sidebar: Dynamic Categories */}
        <aside className="w-64 border-r border-white/8 bg-[#0c101d] flex flex-col shrink-0 min-h-0">
          <div className="p-3 border-b border-white/6 flex items-center justify-between">
            <span className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider">
              {t('worldBible.categoriesHeader')}
            </span>
            <button
              onClick={handleOpenAddCategory}
              className="p-1 hover:bg-white/10 rounded text-slate-400 hover:text-slate-200 transition-colors"
              title={t('worldBible.addCategoryTooltip')}
            >
              <Plus className="w-3.5 h-3.5" />
            </button>
          </div>

          <div className="flex-1 overflow-y-auto p-2 space-y-1">
            {/* All Categories Item */}
            <button
              onClick={() => setSelectedWorldCategory(null)}
              className={`w-full flex items-center justify-between px-3 py-2 rounded-lg text-xs transition-colors ${
                selectedWorldCategory === null
                  ? 'bg-indigo-600/20 text-indigo-300 border border-indigo-500/30 font-medium'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-white/4'
              }`}
            >
              <div className="flex items-center gap-2 truncate">
                <Globe className="w-3.5 h-3.5 shrink-0 text-indigo-400" />
                <span className="truncate">{t('worldBible.allItems')}</span>
              </div>
              <span className="text-[10px] px-1.5 py-0.5 rounded-full bg-white/6 font-mono">
                {worldEntries.length}
              </span>
            </button>

            {/* Dynamic Categories List */}
            {worldCategories.map((cat) => {
              const isSelected = selectedWorldCategory === cat.id;
              return (
                <div
                  key={cat.id}
                  className={`group flex items-center justify-between px-3 py-2 rounded-lg text-xs transition-colors cursor-pointer ${
                    isSelected
                      ? 'bg-indigo-600/20 text-indigo-300 border border-indigo-500/30 font-medium'
                      : 'text-slate-400 hover:text-slate-200 hover:bg-white/4'
                  }`}
                  onClick={() => setSelectedWorldCategory(cat.id)}
                >
                  <div className="flex items-center gap-2 truncate min-w-0">
                    <span className="shrink-0 text-indigo-400">
                      {renderCategoryIcon(cat.icon, 'w-3.5 h-3.5')}
                    </span>
                    <span className="truncate" title={cat.description || cat.name}>
                      {cat.name}
                    </span>
                  </div>

                  <div className="flex items-center gap-1.5 shrink-0">
                    <span className="text-[10px] px-1.5 py-0.5 rounded-full bg-white/6 font-mono">
                      {cat.entry_count}
                    </span>
                    {/* Action buttons on hover */}
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleOpenEditCategory(cat);
                      }}
                      className="opacity-0 group-hover:opacity-100 p-0.5 hover:text-slate-100 text-slate-400 transition-opacity"
                    >
                      <Edit2 className="w-3 h-3" />
                    </button>
                    <button
                      onClick={(e) => {
                        e.stopPropagation();
                        handleDeleteCategory(cat);
                      }}
                      className="opacity-0 group-hover:opacity-100 p-0.5 hover:text-rose-400 text-slate-400 transition-opacity"
                    >
                      <Trash2 className="w-3 h-3" />
                    </button>
                  </div>
                </div>
              );
            })}

            {worldCategories.length === 0 && !isWorldBibleLoading && (
              <div className="p-4 text-center text-xs text-slate-500 italic">
                {t('worldBible.noCategoriesNotice')}
              </div>
            )}
          </div>
        </aside>

        {/* Right Main Panel: Entries Cards */}
        <main className="flex-1 flex flex-col min-w-0 bg-[#090d16] overflow-y-auto p-4 md:p-6">
          {/* Active Category Header */}
          <div className="mb-4 flex items-center justify-between pb-3 border-b border-white/6">
            <div>
              <h2 className="text-base font-semibold text-slate-100 flex items-center gap-2">
                {selectedWorldCategory
                  ? worldCategories.find((c) => c.id === selectedWorldCategory)?.name || t('worldBible.categoryLabel')
                  : t('worldBible.defaultTitle')}
              </h2>
              <p className="text-xs text-slate-400 mt-0.5">
                {selectedWorldCategory
                  ? worldCategories.find((c) => c.id === selectedWorldCategory)?.description ||
                    t('worldBible.categorySub')
                  : t('worldBible.defaultSub')}
              </p>
            </div>
            <div className="text-xs text-slate-400 font-mono">
              {t('worldBible.showingEntities')} <span className="text-indigo-300 font-semibold">{filteredEntries.length}</span>{' '}{t('worldBible.entitiesCount')}
            </div>
          </div>

          {/* Entries Grid */}
          {filteredEntries.length > 0 ? (
            <div className="grid grid-cols-1 xl:grid-cols-2 gap-3.5">
              {filteredEntries.map((entry) => {
                return (
                  <div
                    key={entry.id}
                    className="p-4 rounded-xl bg-[#0f1424] border border-white/8 hover:border-white/15 transition-all flex flex-col justify-between group shadow-xs"
                  >
                    <div>
                      {/* Top Badges & Actions */}
                      <div className="flex items-start justify-between gap-2 mb-2">
                        <div className="flex items-center gap-2 flex-wrap">
                          <span className="font-semibold text-slate-100 text-sm">{entry.name}</span>
                          <span className="px-2 py-0.5 rounded-full text-[10px] font-medium bg-indigo-500/10 text-indigo-300 border border-indigo-500/20 flex items-center gap-1">
                            {renderCategoryIcon(
                              worldCategories.find((c) => c.id === entry.category_id)?.icon,
                              'w-3 h-3'
                            )}
                            {entry.category_name || entry.category_slug}
                          </span>
                          {/* Verification Pill */}
                          <button
                            onClick={() => handleToggleVerify(entry)}
                            className={`px-2 py-0.5 rounded-full text-[10px] font-medium transition-colors flex items-center gap-1 ${
                              entry.is_verified
                                ? 'bg-emerald-500/15 text-emerald-300 border border-emerald-500/30'
                                : 'bg-amber-500/15 text-amber-300 border border-amber-500/30 hover:bg-amber-500/25'
                            }`}
                            title={
                              entry.is_verified
                                ? t('worldBible.clickToUnverify') : t('worldBible.clickToVerify')
                            }
                          >
                            {entry.is_verified ? (
                              <>
                                <CheckCircle2 className="w-2.5 h-2.5" />
                                {t('worldBible.verifiedBadge')}
                              </>
                            ) : (
                              <>
                                <Clock className="w-2.5 h-2.5" />
                                {t('worldBible.proposedBadge')}
                              </>
                            )}
                          </button>
                        </div>

                        {/* Card Edit & Delete */}
                        <div className="flex items-center gap-1 opacity-80 group-hover:opacity-100 transition-opacity">
                          <button
                            onClick={() => handleOpenEditEntry(entry)}
                            className="p-1 hover:bg-white/10 rounded text-slate-400 hover:text-slate-200 transition-colors"
                            title={t('common.edit')}
                          >
                            <Edit2 className="w-3.5 h-3.5" />
                          </button>
                          <button
                            onClick={() => handleDeleteEntry(entry.id, entry.name)}
                            className="p-1 hover:bg-rose-500/20 rounded text-slate-400 hover:text-rose-400 transition-colors"
                            title={t('common.delete')}
                          >
                            <Trash2 className="w-3.5 h-3.5" />
                          </button>
                        </div>
                      </div>

                      {/* Aliases */}
                      {entry.aliases && entry.aliases.length > 0 && (
                        <div className="text-[11px] text-slate-400 mb-2 flex items-center gap-1.5 flex-wrap">
                          <span className="text-slate-500">{t('worldBible.aliasesLabel')}</span>
                          {entry.aliases.map((a, idx) => (
                            <span
                              key={idx}
                              className="px-1.5 py-0.2 rounded bg-white/4 text-slate-300 border border-white/5 font-mono text-[10px]"
                            >
                              {a}
                            </span>
                          ))}
                        </div>
                      )}

                      {/* Summary */}
                      <p className="text-xs text-slate-300 leading-relaxed mb-3">
                        {entry.summary}
                      </p>

                      {/* Key-Value Attributes */}
                      {entry.attributes && Object.keys(entry.attributes).length > 0 && (
                        <div className="grid grid-cols-2 gap-1.5 pt-2 border-t border-white/5 text-[11px]">
                          {Object.entries(entry.attributes).map(([key, val]) => (
                            <div
                              key={key}
                              className="flex items-center gap-1.5 truncate text-slate-400"
                            >
                              <span className="text-slate-500 font-medium shrink-0">{key}:</span>
                              <span className="text-slate-300 truncate font-mono text-[10px]">
                                {String(val)}
                              </span>
                            </div>
                          ))}
                        </div>
                      )}
                    </div>

                    <div className="mt-3 pt-2 flex items-center justify-between text-[10px] text-slate-500 border-t border-white/4">
                      <span>{t('worldBible.firstSeen', { chapter: entry.source_chapter_index || 1 })}</span>
                      <span className="capitalize">{entry.discovered_by.replace('_', ' ')}</span>
                    </div>
                  </div>
                );
              })}
            </div>
          ) : (
            <div className="flex-1 flex flex-col items-center justify-center text-center p-8">
              <div className="w-12 h-12 rounded-2xl bg-indigo-500/10 border border-indigo-500/20 flex items-center justify-center text-indigo-400 mb-3">
                <Globe className="w-6 h-6" />
              </div>
              <h3 className="text-sm font-medium text-slate-200 mb-1">{t('worldBible.emptyTitle')}</h3>
              <p className="text-xs text-slate-400 max-w-sm mb-4">
                {t('worldBible.emptyDesc')}
              </p>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => setIsScanModalOpen(true)}
                  className="px-3.5 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-medium flex items-center gap-1.5 shadow-sm transition-colors"
                >
                  <Sparkles className="w-3.5 h-3.5" />
                  {t('worldBible.autoScanFull')}
                </button>
                <button
                  onClick={handleOpenAddEntry}
                  className="px-3.5 py-1.5 bg-white/4 hover:bg-white/8 text-slate-200 border border-white/10 rounded-lg text-xs font-medium flex items-center gap-1.5 transition-colors"
                >
                  <Plus className="w-3.5 h-3.5" />
                  {t('worldBible.addManual')}
                </button>
              </div>
            </div>
          )}
        </main>
      </div>

      {/* MODAL 1: Auto-Scan World */}
      {isScanModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="bg-[#0f1526] border border-white/10 rounded-2xl max-w-md w-full p-5 shadow-2xl">
            <div className="flex items-center justify-between pb-3 border-b border-white/10 mb-4">
              <div className="flex items-center gap-2 text-slate-100 font-semibold text-sm">
                <Sparkles className="w-4 h-4 text-amber-300" />
                <span>Autonomous World Builder Agent</span>
              </div>
              <button
                onClick={() => setIsScanModalOpen(false)}
                className="text-slate-400 hover:text-slate-200"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <p className="text-xs text-slate-300 mb-4 leading-relaxed">
              {t('worldBible.modalScanDesc')}
            </p>

            <div className="grid grid-cols-2 gap-3 mb-4">
              <div>
                <label className="block text-[11px] font-medium text-slate-400 mb-1">
                  {t('worldBible.fromChapter')}
                </label>
                <input
                  type="number"
                  min={1}
                  value={scanStartChap}
                  onChange={(e) => setScanStartChap(Math.max(1, parseInt(e.target.value) || 1))}
                  className="w-full px-3 py-1.5 bg-white/4 border border-white/10 rounded-lg text-xs text-slate-200 focus:outline-none focus:border-indigo-500/50"
                />
              </div>
              <div>
                <label className="block text-[11px] font-medium text-slate-400 mb-1">
                  {t('worldBible.toChapter')}
                </label>
                <input
                  type="number"
                  min={1}
                  value={scanEndChap}
                  onChange={(e) => setScanEndChap(Math.max(1, parseInt(e.target.value) || 1))}
                  className="w-full px-3 py-1.5 bg-white/4 border border-white/10 rounded-lg text-xs text-slate-200 focus:outline-none focus:border-indigo-500/50"
                />
              </div>
            </div>

            {/* Scan Report Display if available */}
            {worldScanReport && (
              <div className="mb-4 p-3 rounded-xl bg-indigo-500/10 border border-indigo-500/20 text-xs">
                <div className="font-semibold text-indigo-300 mb-1">
                  {t('worldBible.inferredGenre', { genre: worldScanReport.inferred_genre })}
                </div>
                <p className="text-slate-300 text-[11px] mb-2">{worldScanReport.summary_report}</p>
                <div className="text-[10px] text-slate-400 flex gap-3">
                  <span>{t('worldBible.catsAdded', { count: worldScanReport.categories_added })}</span>
                  <span>{t('worldBible.entriesAdded', { count: worldScanReport.entries_added })}</span>
                </div>
              </div>
            )}

            <div className="flex items-center justify-end gap-2 pt-3 border-t border-white/10">
              <button
                type="button"
                onClick={() => setIsScanModalOpen(false)}
                className="px-3 py-1.5 bg-white/4 hover:bg-white/8 text-slate-300 rounded-lg text-xs"
              >
                {t('common.close')}
              </button>
              <button
                type="button"
                onClick={handleStartScan}
                disabled={isScanningWorld}
                className="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-medium flex items-center gap-1.5 shadow-sm disabled:opacity-50"
              >
                {isScanningWorld ? (
                  <>
                    <RefreshCw className="w-3.5 h-3.5 animate-spin" />
                    {t('worldBible.analyzingWorld')}
                  </>
                ) : (
                  <>
                    <Sparkles className="w-3.5 h-3.5 text-amber-300" />
                    {t('worldBible.startScan')}
                  </>
                )}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* MODAL 2: Add / Edit Entry */}
      {isEntryModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <form
            onSubmit={handleSaveEntry}
            className="bg-[#0f1526] border border-white/10 rounded-2xl max-w-lg w-full p-5 shadow-2xl flex flex-col max-h-[90vh]"
          >
            <div className="flex items-center justify-between pb-3 border-b border-white/10 mb-4">
              <span className="text-slate-100 font-semibold text-sm">
                {editingEntry ? t('worldBible.editEntityTitle') : t('worldBible.addEntityTitle')}
              </span>
              <button
                type="button"
                onClick={() => setIsEntryModalOpen(false)}
                className="text-slate-400 hover:text-slate-200"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="flex-1 overflow-y-auto space-y-3.5 pr-1">
              <div>
                <label className="block text-[11px] font-medium text-slate-400 mb-1">
                  {t('worldBible.entityNameLabel')} <span className="text-rose-400">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder={t('worldBible.entityNamePlaceholder')}
                  value={entryName}
                  onChange={(e) => setEntryName(e.target.value)}
                  className="w-full px-3 py-1.5 bg-white/4 border border-white/10 rounded-lg text-xs text-slate-200 focus:outline-none focus:border-indigo-500/50"
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-[11px] font-medium text-slate-400 mb-1">
                    {t('worldBible.categorySelectLabel')} <span className="text-rose-400">*</span>
                  </label>
                  <select
                    value={entryCategory}
                    onChange={(e) => setEntryCategory(e.target.value)}
                    className="w-full px-3 py-1.5 bg-[#111728] border border-white/10 rounded-lg text-xs text-slate-200 focus:outline-none focus:border-indigo-500/50"
                  >
                    {worldCategories.map((cat) => (
                      <option key={cat.id} value={cat.id}>
                        {cat.name}
                      </option>
                    ))}
                  </select>
                </div>
                <div>
                  <label className="block text-[11px] font-medium text-slate-400 mb-1">
                    {t('worldBible.aliasesInputLabel')}
                  </label>
                  <input
                    type="text"
                    placeholder="Arasaka, Arasaka Corp"
                    value={entryAliases}
                    onChange={(e) => setEntryAliases(e.target.value)}
                    className="w-full px-3 py-1.5 bg-white/4 border border-white/10 rounded-lg text-xs text-slate-200 focus:outline-none focus:border-indigo-500/50"
                  />
                </div>
              </div>

              <div>
                <label className="block text-[11px] font-medium text-slate-400 mb-1">
                  {t('worldBible.shortSummaryLabel')}
                </label>
                <textarea
                  rows={2}
                  placeholder={t('worldBible.shortSummaryPlaceholder')}
                  value={entrySummary}
                  onChange={(e) => setEntrySummary(e.target.value)}
                  className="w-full px-3 py-1.5 bg-white/4 border border-white/10 rounded-lg text-xs text-slate-200 focus:outline-none focus:border-indigo-500/50"
                />
              </div>

              <div>
                <label className="block text-[11px] font-medium text-slate-400 mb-1">
                  {t('worldBible.fullDescLabel')}
                </label>
                <textarea
                  rows={3}
                  placeholder={t('worldBible.fullDescPlaceholder')}
                  value={entryDescription}
                  onChange={(e) => setEntryDescription(e.target.value)}
                  className="w-full px-3 py-1.5 bg-white/4 border border-white/10 rounded-lg text-xs text-slate-200 focus:outline-none focus:border-indigo-500/50"
                />
              </div>

              {/* Dynamic Key-Value Attributes */}
              <div>
                <div className="flex items-center justify-between mb-1">
                  <label className="text-[11px] font-medium text-slate-400">
                    {t('worldBible.attributesLabel')}
                  </label>
                  <button
                    type="button"
                    onClick={() => setEntryAttrs([...entryAttrs, { key: '', value: '' }])}
                    className="text-[10px] text-indigo-400 hover:text-indigo-300 flex items-center gap-1"
                  >
                    <Plus className="w-3 h-3" /> {t('worldBible.addAttribute')}
                  </button>
                </div>

                <div className="space-y-1.5 max-h-36 overflow-y-auto p-1 bg-black/20 rounded-lg border border-white/5">
                  {entryAttrs.map((attr, idx) => (
                    <div key={idx} className="flex items-center gap-2">
                      <input
                        type="text"
                        placeholder={t('worldBible.attrKeyPlaceholder')}
                        value={attr.key}
                        onChange={(e) => {
                          const updated = [...entryAttrs];
                          updated[idx].key = e.target.value;
                          setEntryAttrs(updated);
                        }}
                        className="w-1/3 px-2 py-1 bg-white/4 border border-white/10 rounded text-[11px] text-slate-200"
                      />
                      <input
                        type="text"
                        placeholder={t('worldBible.attrValPlaceholder')}
                        value={attr.value}
                        onChange={(e) => {
                          const updated = [...entryAttrs];
                          updated[idx].value = e.target.value;
                          setEntryAttrs(updated);
                        }}
                        className="flex-1 px-2 py-1 bg-white/4 border border-white/10 rounded text-[11px] text-slate-200"
                      />
                      <button
                        type="button"
                        onClick={() => setEntryAttrs(entryAttrs.filter((_, i) => i !== idx))}
                        className="text-slate-500 hover:text-rose-400 p-1"
                      >
                        <Trash2 className="w-3 h-3" />
                      </button>
                    </div>
                  ))}
                  {entryAttrs.length === 0 && (
                    <div className="text-[11px] text-slate-500 italic p-1">
                      {t('worldBible.noAttrsNotice')}
                    </div>
                  )}
                </div>
              </div>

              <div className="flex items-center gap-2 pt-1">
                <input
                  type="checkbox"
                  id="entry_verified"
                  checked={entryVerified}
                  onChange={(e) => setEntryVerified(e.target.checked)}
                  className="rounded border-white/20 bg-white/5 text-indigo-600 focus:ring-0"
                />
                <label htmlFor="entry_verified" className="text-xs text-slate-300 select-none">
                  {t('worldBible.verifiedCheckbox')}
                </label>
              </div>
            </div>

            <div className="flex items-center justify-end gap-2 pt-3 border-t border-white/10 mt-4">
              <button
                type="button"
                onClick={() => setIsEntryModalOpen(false)}
                className="px-3 py-1.5 bg-white/4 hover:bg-white/8 text-slate-300 rounded-lg text-xs"
              >
                {t('common.cancel')}
              </button>
              <button
                type="submit"
                className="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-medium shadow-sm"
              >
                {t('worldBible.saveEntity')}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* MODAL 3: Add / Edit Category */}
      {isCategoryModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <form
            onSubmit={handleSaveCategory}
            className="bg-[#0f1526] border border-white/10 rounded-2xl max-w-sm w-full p-5 shadow-2xl"
          >
            <div className="flex items-center justify-between pb-3 border-b border-white/10 mb-4">
              <span className="text-slate-100 font-semibold text-sm">
                {editingCategory ? t('worldBible.editCatTitle') : t('worldBible.addCatTitle')}
              </span>
              <button
                type="button"
                onClick={() => setIsCategoryModalOpen(false)}
                className="text-slate-400 hover:text-slate-200"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="space-y-3">
              <div>
                <label className="block text-[11px] font-medium text-slate-400 mb-1">
                  {t('worldBible.catNameLabel')} <span className="text-rose-400">*</span>
                </label>
                <input
                  type="text"
                  required
                  placeholder={t('worldBible.catNamePlaceholder')}
                  value={categoryName}
                  onChange={(e) => setCategoryName(e.target.value)}
                  className="w-full px-3 py-1.5 bg-white/4 border border-white/10 rounded-lg text-xs text-slate-200 focus:outline-none focus:border-indigo-500/50"
                />
              </div>

              <div>
                <label className="block text-[11px] font-medium text-slate-400 mb-1">
                  {t('worldBible.slugLabel')}
                </label>
                <input
                  type="text"
                  placeholder="factions, magic_ranks, cyberware..."
                  value={categorySlug}
                  onChange={(e) => setCategorySlug(e.target.value)}
                  className="w-full px-3 py-1.5 bg-white/4 border border-white/10 rounded-lg text-xs text-slate-200 focus:outline-none focus:border-indigo-500/50 font-mono"
                />
              </div>

              <div>
                <label className="block text-[11px] font-medium text-slate-400 mb-1">
                  {t('worldBible.iconLabel')}
                </label>
                <select
                  value={categoryIcon}
                  onChange={(e) => setCategoryIcon(e.target.value)}
                  className="w-full px-3 py-1.5 bg-[#111728] border border-white/10 rounded-lg text-xs text-slate-200 focus:outline-none focus:border-indigo-500/50"
                >
                  {Object.keys(ICON_MAP).map((name) => (
                    <option key={name} value={name}>
                      {name}
                    </option>
                  ))}
                </select>
              </div>

              <div>
                <label className="block text-[11px] font-medium text-slate-400 mb-1">{t('worldBible.catDescLabel')}</label>
                <textarea
                  rows={2}
                  placeholder={t('worldBible.catDescPlaceholder')}
                  value={categoryDesc}
                  onChange={(e) => setCategoryDesc(e.target.value)}
                  className="w-full px-3 py-1.5 bg-white/4 border border-white/10 rounded-lg text-xs text-slate-200 focus:outline-none focus:border-indigo-500/50"
                />
              </div>
            </div>

            <div className="flex items-center justify-end gap-2 pt-3 border-t border-white/10 mt-4">
              <button
                type="button"
                onClick={() => setIsCategoryModalOpen(false)}
                className="px-3 py-1.5 bg-white/4 hover:bg-white/8 text-slate-300 rounded-lg text-xs"
              >
                {t('common.cancel')}
              </button>
              <button
                type="submit"
                className="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-medium shadow-sm"
              >
                {t('worldBible.saveCat')}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* MODAL 4: Export Modal */}
      {isExportModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="bg-[#0f1526] border border-white/10 rounded-2xl max-w-2xl w-full p-5 shadow-2xl flex flex-col max-h-[85vh]">
            <div className="flex items-center justify-between pb-3 border-b border-white/10 mb-3">
              <div className="flex items-center gap-2 text-slate-100 font-semibold text-sm">
                <Download className="w-4 h-4 text-indigo-400" />
                <span>{t('worldBible.exportTitle', { format: exportFormat.toUpperCase() })}</span>
              </div>
              <button
                onClick={() => setIsExportModalOpen(false)}
                className="text-slate-400 hover:text-slate-200"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="flex items-center gap-2 mb-3">
              <button
                onClick={() => handleOpenExport('markdown')}
                className={`px-3 py-1 rounded-lg text-xs font-medium transition-colors ${
                  exportFormat === 'markdown'
                    ? 'bg-indigo-600 text-white'
                    : 'bg-white/4 text-slate-400 hover:text-slate-200'
                }`}
              >
                Markdown (WORLD_BIBLE.md)
              </button>
              <button
                onClick={() => handleOpenExport('json')}
                className={`px-3 py-1 rounded-lg text-xs font-medium transition-colors ${
                  exportFormat === 'json'
                    ? 'bg-indigo-600 text-white'
                    : 'bg-white/4 text-slate-400 hover:text-slate-200'
                }`}
              >
                JSON Data
              </button>
            </div>

            <div className="flex-1 overflow-hidden flex flex-col mb-3">
              <textarea
                readOnly
                value={exportText}
                className="flex-1 w-full p-3 bg-black/40 border border-white/10 rounded-xl text-xs text-slate-300 font-mono focus:outline-none resize-none"
              />
            </div>

            <div className="flex items-center justify-between pt-3 border-t border-white/10">
              <div className="text-[11px] text-slate-400">
                {t('worldBible.exportDesc')}
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={handleCopyExport}
                  className="px-3 py-1.5 bg-white/4 hover:bg-white/8 text-slate-200 border border-white/10 rounded-lg text-xs font-medium flex items-center gap-1.5 transition-colors"
                >
                  {hasCopied ? (
                    <>
                      <Check className="w-3.5 h-3.5 text-emerald-400" />
                      {t('common.copied')}
                    </>
                  ) : (
                    <>
                      <Copy className="w-3.5 h-3.5" />
                      {t('worldBible.copyBtn')}
                    </>
                  )}
                </button>
                <button
                  onClick={handleDownloadExport}
                  className="px-3.5 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-medium flex items-center gap-1.5 shadow-sm transition-colors"
                >
                  <Download className="w-3.5 h-3.5" />
                  {t('worldBible.downloadBtn')}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* MODAL 5: Import Modal */}
      {isImportModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="bg-[#0f1526] border border-white/10 rounded-2xl max-w-xl w-full p-5 shadow-2xl flex flex-col max-h-[85vh]">
            <div className="flex items-center justify-between pb-3 border-b border-white/10 mb-3">
              <div className="flex items-center gap-2 text-slate-100 font-semibold text-sm">
                <Upload className="w-4 h-4 text-indigo-400" />
                <span>{t('worldBible.importTitle')}</span>
              </div>
              <button
                onClick={() => setIsImportModalOpen(false)}
                className="text-slate-400 hover:text-slate-200"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <p className="text-xs text-slate-300 mb-2 leading-relaxed">
              {t('worldBible.importDesc')}
            </p>

            <textarea
              rows={10}
              placeholder={t('worldBible.importPlaceholder')}
              value={importText}
              onChange={(e) => setImportText(e.target.value)}
              className="w-full p-3 bg-black/40 border border-white/10 rounded-xl text-xs text-slate-300 font-mono focus:outline-none resize-none mb-3"
            />

            <div className="flex items-center justify-end gap-2 pt-3 border-t border-white/10">
              <button
                type="button"
                onClick={() => setIsImportModalOpen(false)}
                className="px-3 py-1.5 bg-white/4 hover:bg-white/8 text-slate-300 rounded-lg text-xs"
              >
                {t('common.cancel')}
              </button>
              <button
                type="button"
                onClick={handleImport}
                className="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-medium flex items-center gap-1.5 shadow-sm"
              >
                <Upload className="w-3.5 h-3.5" />
                {t('worldBible.startImport')}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
