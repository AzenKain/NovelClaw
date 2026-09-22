import React, { useState, useMemo, useEffect } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'react-hot-toast';
import {
  BookMarked,
  Plus,
  Search,
  Trash2,
  Edit2,
  Download,
  Upload,
  FileSpreadsheet,
  FileCode,
  FileText,
  Copy,
  Check,
  X,
  AlertCircle,
  Layers,
  Globe,
  CloudDownload,
  ExternalLink,
  Link as LinkIcon,
} from 'lucide-react';
import * as dtos from '@bindings/novelclaw/internal/dtos/models';
import { useAppStore } from '@/store/useAppStore';
import { Select } from '@/components/ui/Select';
import { showConfirmModal } from '@/store/confirmStore';

// Multi-genre perspective definitions
interface GenreTaxonomy {
  id: string;
  name: string;
  badgeColor: string;
  categoryLabels: Record<string, string>;
}

const getGenreTaxonomies = (t: any): GenreTaxonomy[] => [
  {
    id: 'universal',
    name: t('glossary.taxonomies.universal'),
    badgeColor: 'bg-indigo-500/20 text-indigo-300 border-indigo-500/30',
    categoryLabels: {
      proper_name: t('glossary.taxonomies.u_proper_name'),
      faction: t('glossary.taxonomies.u_faction'),
      rank: t('glossary.taxonomies.u_rank'),
      ability: t('glossary.taxonomies.u_ability'),
      item: t('glossary.taxonomies.u_item'),
      location: t('glossary.taxonomies.u_location'),
      other: t('glossary.taxonomies.u_other'),
    },
  },
  {
    id: 'romcom',
    name: t('glossary.taxonomies.romcom'),
    badgeColor: 'bg-pink-500/20 text-pink-300 border-pink-500/30',
    categoryLabels: {
      proper_name: t('glossary.taxonomies.u_proper_name'),
      faction: t('glossary.taxonomies.u_faction'),
      rank: t('glossary.taxonomies.u_rank'),
      ability: t('glossary.taxonomies.u_ability'),
      item: t('glossary.taxonomies.u_item'),
      location: t('glossary.taxonomies.u_location'),
      other: t('glossary.taxonomies.u_other'),
    },
  },
  {
    id: 'xianxia',
    name: t('glossary.taxonomies.xianxia'),
    badgeColor: 'bg-emerald-500/20 text-emerald-300 border-emerald-500/30',
    categoryLabels: {
      proper_name: t('glossary.taxonomies.x_proper_name'),
      faction: t('glossary.taxonomies.x_faction'),
      rank: t('glossary.taxonomies.x_rank'),
      ability: t('glossary.taxonomies.x_ability'),
      item: t('glossary.taxonomies.x_item'),
      location: t('glossary.taxonomies.x_location'),
      other: t('glossary.taxonomies.x_other'),
    },
  },
  {
    id: 'scifi',
    name: t('glossary.taxonomies.scifi'),
    badgeColor: 'bg-cyan-500/20 text-cyan-300 border-cyan-500/30',
    categoryLabels: {
      proper_name: t('glossary.taxonomies.s_proper_name'),
      faction: t('glossary.taxonomies.s_faction'),
      rank: t('glossary.taxonomies.s_rank'),
      ability: t('glossary.taxonomies.s_ability'),
      item: t('glossary.taxonomies.s_item'),
      location: t('glossary.taxonomies.s_location'),
      other: t('glossary.taxonomies.s_other'),
    },
  },
  {
    id: 'fantasy',
    name: t('glossary.taxonomies.fantasy'),
    badgeColor: 'bg-purple-500/20 text-purple-300 border-purple-500/30',
    categoryLabels: {
      proper_name: t('glossary.taxonomies.f_proper_name'),
      faction: t('glossary.taxonomies.f_faction'),
      rank: t('glossary.taxonomies.f_rank'),
      ability: t('glossary.taxonomies.f_ability'),
      item: t('glossary.taxonomies.f_item'),
      location: t('glossary.taxonomies.f_location'),
      other: t('glossary.taxonomies.f_other'),
    },
  },
  {
    id: 'history',
    name: t('glossary.taxonomies.history'),
    badgeColor: 'bg-amber-500/20 text-amber-300 border-amber-500/30',
    categoryLabels: {
      proper_name: t('glossary.taxonomies.h_proper_name'),
      faction: t('glossary.taxonomies.h_faction'),
      rank: t('glossary.taxonomies.h_rank'),
      ability: t('glossary.taxonomies.h_ability'),
      item: t('glossary.taxonomies.h_item'),
      location: t('glossary.taxonomies.h_location'),
      other: t('glossary.taxonomies.h_other'),
    },
  },
  {
    id: 'apocalypse',
    name: t('glossary.taxonomies.apocalypse'),
    badgeColor: 'bg-rose-500/20 text-rose-300 border-rose-500/30',
    categoryLabels: {
      proper_name: t('glossary.taxonomies.a_proper_name'),
      faction: t('glossary.taxonomies.a_faction'),
      rank: t('glossary.taxonomies.a_rank'),
      ability: t('glossary.taxonomies.a_ability'),
      item: t('glossary.taxonomies.a_item'),
      location: t('glossary.taxonomies.a_location'),
      other: t('glossary.taxonomies.a_other'),
    },
  },
];

// Base categories with colors
const BASE_CATEGORIES = [
  { key: 'all', color: 'bg-slate-700 text-slate-200' },
  { key: 'proper_name', color: 'bg-indigo-500/20 text-indigo-300 border-indigo-500/30' },
  { key: 'faction', color: 'bg-emerald-500/20 text-emerald-300 border-emerald-500/30' },
  { key: 'rank', color: 'bg-amber-500/20 text-amber-300 border-amber-500/30' },
  { key: 'ability', color: 'bg-rose-500/20 text-rose-300 border-rose-500/30' },
  { key: 'item', color: 'bg-blue-500/20 text-blue-300 border-blue-500/30' },
  { key: 'location', color: 'bg-cyan-500/20 text-cyan-300 border-cyan-500/30' },
  { key: 'other', color: 'bg-purple-500/20 text-purple-300 border-purple-500/30' },
];

export const GlossaryHub: React.FC = () => {
  const { t } = useTranslation();
  const genreTaxonomies = useMemo(() => getGenreTaxonomies(t), [t]);
  const {
    selectedProject,
    terms,
    isGlossaryLoading,
    upsertGlossaryTerm,
    deleteGlossaryTerm,
    importGlossaryTSV,
    exportGlossaryTSV,
    importGlossaryJSON,
    exportGlossaryJSON,
    importGlossaryPlainText,
    listRemoteGlossarySources,
    downloadAndImportGlossary,
  } = useAppStore();

  const [searchQuery, setSearchQuery] = useState('');
  const [selectedCategory, setSelectedCategory] = useState('all');
  const [activeGenreTaxonomy, setActiveGenreTaxonomy] = useState('universal');

  // Edit / Add modal
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingTerm, setEditingTerm] = useState<dtos.GlossaryTermDTO | null>(null);
  const [sourceTerm, setSourceTerm] = useState('');
  const [targetTerm, setTargetTerm] = useState('');
  const [category, setCategory] = useState('proper_name');
  const [_notes, setNotes] = useState('');
  const [formError, setFormError] = useState('');

  // Import / Export modal
  const [isImportModalOpen, setIsImportModalOpen] = useState(false);
  const [importFormat, setImportFormat] = useState<'vietphrase' | 'tsv' | 'json'>('vietphrase');
  const [importCategory, setImportCategory] = useState('other');
  const [importContent, setImportContent] = useState('');
  const [importStatus, setImportStatus] = useState<string | null>(null);
  const [isImportError, setIsImportError] = useState(false);

  const [isExportModalOpen, setIsExportModalOpen] = useState(false);
  const [exportFormat, setExportFormat] = useState<'tsv' | 'json'>('tsv');
  const [exportContent, setExportContent] = useState('');
  const [hasCopied, setHasCopied] = useState(false);

  // Remote Online Dictionary Modal
  const [isRemoteModalOpen, setIsRemoteModalOpen] = useState(false);
  const [remoteSources, setRemoteSources] = useState<dtos.RemoteGlossarySourceDTO[]>([]);
  const [remoteTab, setRemoteTab] = useState<'community' | 'custom'>('custom');
  const [customUrl, setCustomUrl] = useState('');
  const [customFormat, setCustomFormat] = useState<'vietphrase' | 'tsv' | 'json'>('vietphrase');
  const [customCategory, setCustomCategory] = useState('proper_name');
  const [isDownloading, setIsDownloading] = useState<string | null>(null);

  // Active taxonomy helper
  const currentTaxonomy = useMemo(() => {
    return genreTaxonomies.find((g) => g.id === activeGenreTaxonomy) || genreTaxonomies[0];
  }, [genreTaxonomies, activeGenreTaxonomy]);

  // Load remote sources on demand
  useEffect(() => {
    if (isRemoteModalOpen) {
      listRemoteGlossarySources().then((list) => {
        setRemoteSources(list || []);
      });
    }
  }, [isRemoteModalOpen, listRemoteGlossarySources]);

  // Backward compatibility normalizer
  const normalizeTermCategory = (rawCat: string): string => {
    if (!rawCat) return 'other';
    const c = rawCat.toLowerCase();
    if (c === 'proper_name') return 'proper_name';
    if (c === 'faction' || c === 'sect') return 'faction';
    if (c === 'rank' || c === 'realm') return 'rank';
    if (c === 'ability' || c === 'skill') return 'ability';
    if (c === 'item' || c === 'artifact') return 'item';
    if (c === 'location') return 'location';
    return 'other';
  };

  // Filtered terms
  const filteredTerms = useMemo(() => {
    return terms.filter((term) => {
      const normalizedCat = normalizeTermCategory(term.category);
      const matchCat = selectedCategory === 'all' || normalizedCat === selectedCategory;
      const q = searchQuery.toLowerCase().trim();
      const matchSearch =
        !q ||
        term.source_term.toLowerCase().includes(q) ||
        term.target_term.toLowerCase().includes(q) ||
        (term.notes && term.notes.toLowerCase().includes(q));
      return matchCat && matchSearch;
    });
  }, [terms, selectedCategory, searchQuery]);

  // Open modal for add
  const handleOpenAdd = () => {
    setEditingTerm(null);
    setSourceTerm('');
    setTargetTerm('');
    setCategory('proper_name');
    setNotes('');
    setFormError('');
    setIsModalOpen(true);
  };

  // Open modal for edit
  const handleOpenEdit = (term: dtos.GlossaryTermDTO) => {
    setEditingTerm(term);
    setSourceTerm(term.source_term);
    setTargetTerm(term.target_term);
    setCategory(normalizeTermCategory(term.category));
    setNotes(term.notes || '');
    setFormError('');
    setIsModalOpen(true);
  };

  // Submit term
  const handleSubmitTerm = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedProject) return;
    if (!sourceTerm.trim() || !targetTerm.trim()) {
      setFormError(t('glossary.formErrorRequired'));
      return;
    }

    try {
      await upsertGlossaryTerm({
        id: editingTerm ? editingTerm.id : '',
        project_id: selectedProject.id,
        source_term: sourceTerm.trim(),
        target_term: targetTerm.trim(),
        category: category,
      });
      setIsModalOpen(false);
      toast.success(editingTerm ? t('glossary.toasts.termUpdated') : t('glossary.toasts.termCreated'));
    } catch (err: any) {
      const msg = err?.message || t('glossary.toasts.saveError', { err: '' });
      setFormError(msg);
      toast.error(msg);
    }
  };

  // Delete term
  const handleDeleteTerm = async (id: string) => {
    const confirmed = await showConfirmModal({
      title: t('glossary.deleteModalTitle'),
      message: t('glossary.confirmDelete'),
      confirmText: t('common.delete'),
      cancelText: t('common.cancel'),
      variant: 'danger',
    });
    if (!confirmed) return;
    try {
      await deleteGlossaryTerm(id);
      toast.success(t('glossary.toasts.deleted'));
    } catch (err: any) {
      toast.error(t('glossary.toasts.deleteError', { err: err?.message || err }));
    }
  };

  // Export handlers
  const handleOpenExport = async (format: 'tsv' | 'json') => {
    setExportFormat(format);
    setHasCopied(false);
    try {
      const content = format === 'tsv' ? await exportGlossaryTSV() : await exportGlossaryJSON();
      setExportContent(content);
      setIsExportModalOpen(true);
    } catch (err) {
      console.error('Export error:', err);
    }
  };

  const handleCopyExport = () => {
    navigator.clipboard.writeText(exportContent);
    setHasCopied(true);
    toast.success(t('glossary.toasts.copied'));
    setTimeout(() => setHasCopied(false), 2000);
  };

  const handleDownloadExport = () => {
    const filename = `${selectedProject?.title || 'glossary'}_${exportFormat}.${exportFormat}`;
    const blob = new Blob([exportContent], {
      type: exportFormat === 'json' ? 'application/json' : 'text/tab-separated-values',
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = filename;
    a.click();
    URL.revokeObjectURL(url);
    toast.success(t('glossary.toasts.downloaded'));
  };

  // Import handler
  const handleDoImport = async () => {
    if (!importContent.trim()) {
      setIsImportError(true);
      setImportStatus(t('glossary.toasts.pasteRequired'));
      toast.error(t('glossary.toasts.pasteRequired'));
      return;
    }
    setIsImportError(false);
    setImportStatus(t('glossary.toasts.processingImport'));
    try {
      let count = 0;
      if (importFormat === 'vietphrase') {
        count = await importGlossaryPlainText(importContent, importCategory);
      } else if (importFormat === 'tsv') {
        count = await importGlossaryTSV(importContent);
      } else {
        count = await importGlossaryJSON(importContent);
      }
      setIsImportError(false);
      setImportStatus(t('glossary.toasts.importSuccessCount', { count }));
      toast.success(t('glossary.toasts.importSuccessToast', { count }));
      setTimeout(() => {
        setIsImportModalOpen(false);
        setImportContent('');
        setImportStatus(null);
      }, 1200);
    } catch (err: any) {
      setIsImportError(true);
      const msg = t('glossary.toasts.importError', { err: err?.message || err });
      setImportStatus(msg);
      toast.error(msg);
    }
  };

  // Handle file upload in import modal
  const handleFileUpload = (e: React.ChangeEvent<HTMLInputElement>) => {
    const file = e.target.files?.[0];
    if (!file) return;
    const reader = new FileReader();
    reader.onload = (event) => {
      const text = event.target?.result as string;
      if (text) {
        setImportContent(text);
        if (file.name.endsWith('.json')) {
          setImportFormat('json');
        } else if (file.name.endsWith('.tsv')) {
          setImportFormat('tsv');
        } else {
          setImportFormat('vietphrase');
        }
        toast.success(t('glossary.toasts.fileLoaded', { name: file.name, size: (file.size / 1024).toFixed(1) }));
      }
    };
    reader.readAsText(file);
  };

  // Remote Online Dictionary Download handler
  const handleDownloadRemote = async (url: string, title: string, format: string = 'vietphrase', defaultCategory: string = 'proper_name') => {
    if (!url.trim()) {
      toast.error(t('glossary.toasts.emptyUrl'));
      return;
    }
    setIsDownloading(url);
    try {
      const count = await downloadAndImportGlossary(url, format, defaultCategory);
      toast.success(t('glossary.toasts.downloadSuccess', { count, title }));
      if (remoteTab === 'custom') {
        setIsRemoteModalOpen(false);
        setCustomUrl('');
      }
    } catch (err: any) {
      toast.error(t('glossary.toasts.downloadError', { err: err?.message || err }));
    } finally {
      setIsDownloading(null);
    }
  };

  if (!selectedProject) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center p-8 text-center text-slate-400 bg-[#0a0d14]">
        <BookMarked className="w-16 h-16 text-indigo-400/40 mb-4 animate-pulse" />
        <h2 className="text-xl font-semibold text-slate-200 mb-2">{t('glossary.noProjectTitle')}</h2>
        <p className="max-w-md text-sm text-slate-500">{t('glossary.noProjectSubtitle')}</p>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col h-full overflow-hidden bg-[#0a0d14] text-slate-200">
      {/* Top Header Bar */}
      <div className="px-6 py-4 border-b border-white/8 flex flex-wrap items-center justify-between gap-4 bg-[#0d121f]/50">
        <div>
          <div className="flex items-center gap-2.5">
            <div className="w-8 h-8 rounded-lg bg-indigo-500/20 text-indigo-400 flex items-center justify-center border border-indigo-500/30 shadow-xs">
              <BookMarked className="w-4 h-4" />
            </div>
            <div>
              <h1 className="text-base font-bold text-slate-100 flex items-center gap-2">
                {t('glossary.title')}
                <span className="text-xs px-2 py-0.5 rounded-full bg-indigo-500/10 text-indigo-400 border border-indigo-500/20 font-mono">
                  {terms.length} {t('glossary.termsCount')}
                </span>
              </h1>
              <p className="text-xs text-slate-400">
                {t('glossary.subtitle')}
              </p>
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="flex items-center gap-2 flex-wrap">
          {/* Remote Online Dictionary Downloader */}
          <button
            onClick={() => setIsRemoteModalOpen(true)}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold text-sky-300 hover:text-sky-200 bg-sky-500/10 hover:bg-sky-500/20 border border-sky-500/30 rounded-lg transition shadow-xs"
            title={t('glossary.remoteDownloadTooltip')}
          >
            <CloudDownload className="w-3.5 h-3.5 text-sky-400" />
            <span>{t('glossary.remoteDownloadTitle')}</span>
          </button>

          {/* Import Button */}
          <button
            onClick={() => {
              setImportContent('');
              setImportStatus(null);
              setIsImportModalOpen(true);
            }}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-slate-300 hover:text-white bg-white/4 hover:bg-white/8 border border-white/10 rounded-lg transition"
            title={t('glossary.importTooltip')}
          >
            <Upload className="w-3.5 h-3.5 text-indigo-400" />
            <span>{t('glossary.importBtn')}</span>
          </button>

          {/* Export Button */}
          <button
            onClick={() => handleOpenExport('tsv')}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-medium text-slate-300 hover:text-white bg-white/4 hover:bg-white/8 border border-white/10 rounded-lg transition"
            title={t('glossary.exportTooltip')}
          >
            <Download className="w-3.5 h-3.5 text-indigo-400" />
            <span>{t('glossary.exportBtn')}</span>
          </button>

          {/* Add Term Button */}
          <button
            onClick={handleOpenAdd}
            className="flex items-center gap-1.5 px-3 py-1.5 text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-500 rounded-lg shadow-sm transition"
          >
            <Plus className="w-4 h-4" />
            <span>{t('glossary.addTerm')}</span>
          </button>
        </div>
      </div>

      {/* Sub-toolbar Tier 1: Genre Perspective Selector + Search Input */}
      <div className="px-6 py-2.5 border-b border-white/6 flex items-center justify-between gap-4 bg-[#0a0e17] flex-nowrap">
        {/* Left: Genre Perspective Selector */}
        <div className="flex items-center gap-2 bg-white/3 hover:bg-white/5 border border-white/10 px-3 py-1.5 rounded-lg text-xs transition shrink-0">
          <Layers className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
          <span className="text-slate-400 text-xs font-medium shrink-0">{t('glossary.genrePerspective')}</span>
          <select
            value={activeGenreTaxonomy}
            onChange={(e) => setActiveGenreTaxonomy(e.target.value)}
            className="bg-transparent text-slate-200 text-xs font-medium focus:outline-none cursor-pointer pr-1"
          >
            {genreTaxonomies.map((g) => (
              <option key={g.id} value={g.id} className="bg-[#111625] text-slate-200">
                {g.name}
              </option>
            ))}
          </select>
        </div>

        {/* Right: Search input on the exact same row */}
        <div className="relative flex-1 max-w-xs sm:max-w-sm ml-auto shrink-0">
          <Search className="w-3.5 h-3.5 text-slate-400 absolute left-3 top-1/2 -translate-y-1/2 pointer-events-none" />
          <input
            type="text"
            value={searchQuery}
            onChange={(e) => setSearchQuery(e.target.value)}
            placeholder={t('glossary.searchPlaceholder')}
            className="w-full pl-9 pr-8 py-1.5 text-xs bg-white/3 hover:bg-white/5 focus:bg-white/8 text-slate-200 placeholder-slate-500 border border-white/10 rounded-lg outline-none focus:border-indigo-500/50 transition"
          />
          {searchQuery && (
            <button
              onClick={() => setSearchQuery('')}
              className="absolute right-2.5 top-1/2 -translate-y-1/2 text-slate-500 hover:text-slate-300"
            >
              <X className="w-3.5 h-3.5" />
            </button>
          )}
        </div>
      </div>

      {/* Sub-toolbar Tier 2: Category Pills */}
      <div className="px-6 py-2 border-b border-white/4 flex items-center gap-1.5 overflow-x-auto scrollbar-none bg-[#090d16] flex-nowrap">
        {BASE_CATEGORIES.map((cat) => {
          const isActive = selectedCategory === cat.key;
          const label =
            cat.key === 'all'
              ? t('glossary.filterCategoryAll')
              : (currentTaxonomy.categoryLabels[cat.key] || t('glossary.categories.' + cat.key, cat.key));

          return (
            <button
              key={cat.key}
              onClick={() => setSelectedCategory(cat.key)}
              className={`px-3 py-1 rounded-md text-xs font-medium transition whitespace-nowrap border shrink-0 ${
                isActive
                  ? 'bg-indigo-600 text-white border-indigo-400 shadow-xs'
                  : 'bg-white/3 text-slate-400 hover:text-slate-200 border-white/6 hover:bg-white/6'
              }`}
              title={label}
            >
              {label}
            </button>
          );
        })}
      </div>

      {/* Terms Table */}
      <div className="flex-1 overflow-auto p-6">
        {isGlossaryLoading ? (
          <div className="flex items-center justify-center h-48 text-slate-400 text-sm">
            <span className="animate-spin mr-2">⏳</span> {t('glossary.loading')}
          </div>
        ) : filteredTerms.length === 0 ? (
          <div className="flex flex-col items-center justify-center h-64 border border-dashed border-white/10 rounded-xl p-6 text-center">
            <BookMarked className="w-10 h-10 text-slate-600 mb-3" />
            <p className="text-sm font-medium text-slate-300 mb-1">
              {searchQuery || selectedCategory !== 'all'
                ? t('glossary.notFound')
                : t('glossary.empty')}
            </p>
            <p className="text-xs text-slate-500 max-w-md mb-4">
              {t('glossary.emptyDesc')}
            </p>
            <div className="flex items-center gap-3">
              <button
                onClick={() => setIsRemoteModalOpen(true)}
                className="flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-semibold text-sky-300 bg-sky-500/10 hover:bg-sky-500/20 border border-sky-500/30 rounded-lg transition"
              >
                <CloudDownload className="w-3.5 h-3.5 text-sky-400" />
                <span>{t('glossary.remoteDownloadTitle')}</span>
              </button>
              <button
                onClick={handleOpenAdd}
                className="flex items-center gap-1.5 px-3.5 py-1.5 text-xs font-semibold text-white bg-indigo-600 hover:bg-indigo-500 rounded-lg transition"
              >
                <Plus className="w-3.5 h-3.5" />
                <span>{t('glossary.addManually')}</span>
              </button>
            </div>
          </div>
        ) : (
          <div className="border border-white/8 rounded-xl overflow-hidden shadow-xs bg-[#0c101c]">
            <table className="w-full text-left text-xs border-collapse">
              <thead>
                <tr className="bg-white/4 border-b border-white/8 text-slate-400 font-semibold uppercase tracking-wider">
                  <th className="py-3 px-4 w-1/4">{t('glossary.colSource')}</th>
                  <th className="py-3 px-4 w-1/4">{t('glossary.colTarget')}</th>
                  <th className="py-3 px-4 w-1/6">{t('glossary.colCategory')}</th>
                  <th className="py-3 px-4 w-1/4">{t('glossary.colNotes')}</th>
                  <th className="py-3 px-4 w-24 text-right">{t('glossary.colActions')}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-white/4">
                {filteredTerms.map((term) => {
                  const normalizedCat = normalizeTermCategory(term.category);
                  const baseCatConfig =
                    BASE_CATEGORIES.find((c) => c.key === normalizedCat) ||
                    BASE_CATEGORIES[BASE_CATEGORIES.length - 1];
                  const dynamicLabel =
                    currentTaxonomy.categoryLabels[normalizedCat] ||
                    currentTaxonomy.categoryLabels['other'] ||
                    t('glossary.otherCategory');

                  return (
                    <tr key={term.id} className="hover:bg-white/2 transition-colors group">
                      <td className="py-3 px-4 font-mono font-medium text-slate-200">
                        {term.source_term}
                      </td>
                      <td className="py-3 px-4 font-medium text-indigo-300">
                        {term.target_term}
                      </td>
                      <td className="py-3 px-4">
                        <span
                          className={`inline-flex items-center px-2 py-0.5 rounded-full text-[10px] font-medium border ${baseCatConfig.color}`}
                        >
                          {dynamicLabel}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-slate-400 truncate max-w-xs">
                        {term.notes || <span className="text-slate-600 italic">—</span>}
                      </td>
                      <td className="py-3 px-4 text-right">
                        <div className="flex items-center justify-end gap-1.5 opacity-80 group-hover:opacity-100 transition-opacity">
                          <button
                            onClick={() => handleOpenEdit(term)}
                            className="p-1.5 text-slate-400 hover:text-white hover:bg-white/8 rounded-md transition"
                            title={t('glossary.editAction')}
                          >
                            <Edit2 className="w-3.5 h-3.5" />
                          </button>
                          <button
                            onClick={() => handleDeleteTerm(term.id)}
                            className="p-1.5 text-slate-400 hover:text-rose-400 hover:bg-rose-500/10 rounded-md transition"
                            title={t('glossary.deleteAction')}
                          >
                            <Trash2 className="w-3.5 h-3.5" />
                          </button>
                        </div>
                      </td>
                    </tr>
                  );
                })}
              </tbody>
            </table>
          </div>
        )}
      </div>

      {/* Remote Online Dictionary Downloader Modal */}
      {isRemoteModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="bg-[#111625] border border-white/10 rounded-2xl w-full max-w-2xl p-6 shadow-2xl animate-in fade-in zoom-in duration-150 flex flex-col max-h-[85vh]">
            {/* Modal Header */}
            <div className="flex items-center justify-between pb-4 border-b border-white/10">
              <div className="flex items-center gap-2.5">
                <div className="w-8 h-8 rounded-lg bg-sky-500/20 text-sky-400 flex items-center justify-center border border-sky-500/30 shadow-xs">
                  <CloudDownload className="w-4 h-4" />
                </div>
                <div>
                  <h3 className="text-sm font-bold text-slate-100 flex items-center gap-2">
                    {t('glossary.remoteModalTitle')}
                  </h3>
                  <p className="text-xs text-slate-400">
                    {t('glossary.remoteModalSubtitle')}
                  </p>
                </div>
              </div>
              <button
                onClick={() => setIsRemoteModalOpen(false)}
                className="text-slate-400 hover:text-white p-1 rounded-md transition"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {/* Tab Navigation */}
            <div className="flex items-center gap-2 mt-4 border-b border-white/8 pb-2">
              <button
                onClick={() => setRemoteTab('custom')}
                className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-semibold transition ${
                  remoteTab === 'custom'
                    ? 'bg-sky-500/20 text-sky-300 border border-sky-500/30'
                    : 'text-slate-400 hover:text-slate-200 hover:bg-white/4'
                }`}
              >
                <LinkIcon className="w-3.5 h-3.5" />
                <span>{t('glossary.tabCustomUrl')}</span>
              </button>
              {remoteSources.length > 0 && (
                <button
                  onClick={() => setRemoteTab('community')}
                  className={`flex items-center gap-2 px-3 py-1.5 rounded-lg text-xs font-semibold transition ${
                    remoteTab === 'community'
                      ? 'bg-sky-500/20 text-sky-300 border border-sky-500/30'
                      : 'text-slate-400 hover:text-slate-200 hover:bg-white/4'
                  }`}
                >
                  <Globe className="w-3.5 h-3.5" />
                  <span>{t('glossary.tabCommunity')} ({remoteSources.length})</span>
                </button>
              )}
            </div>

            {/* Tab 1: Community Sources */}
            {remoteTab === 'community' && (
              <div className="mt-4 overflow-y-auto space-y-3 pr-1 flex-1">
                {remoteSources.length === 0 ? (
                  <div className="text-center py-8 text-slate-400 text-xs">
                    {t('glossary.noRemoteCatalogs')}
                  </div>
                ) : (
                  remoteSources.map((src) => {
                    const isBusy = isDownloading === src.url;
                    return (
                      <div
                        key={src.id}
                        className="p-4 rounded-xl bg-white/2 hover:bg-white/4 border border-white/10 hover:border-sky-500/30 transition flex flex-col sm:flex-row items-start sm:items-center justify-between gap-4"
                      >
                        <div className="space-y-1.5 flex-1 min-w-0">
                          <div className="flex items-center gap-2 flex-wrap">
                            <h4 className="text-xs font-bold text-slate-100">{src.title}</h4>
                            <span className="text-[10px] px-2 py-0.5 rounded-full bg-sky-500/10 text-sky-300 border border-sky-500/20 font-mono uppercase">
                              {src.format}
                            </span>
                            <span className="text-[10px] px-2 py-0.5 rounded-full bg-indigo-500/10 text-indigo-300 border border-indigo-500/20 font-mono">
                              {src.default_category}
                            </span>
                          </div>
                          <p className="text-xs text-slate-400 leading-relaxed">
                            {src.description}
                          </p>
                          <div className="pt-1">
                            <a
                              href={src.url}
                              target="_blank"
                              rel="noreferrer"
                              className="text-[11px] text-sky-400/80 hover:text-sky-300 flex items-center gap-1 font-mono truncate max-w-md hover:underline"
                            >
                              <ExternalLink className="w-3 h-3 shrink-0" />
                              <span className="truncate">{src.url}</span>
                            </a>
                          </div>
                        </div>

                        <button
                          disabled={isBusy}
                          onClick={() => handleDownloadRemote(src.url, src.title, src.format, src.default_category)}
                          className="shrink-0 flex items-center gap-1.5 px-3.5 py-2 text-xs font-semibold text-white bg-sky-600 hover:bg-sky-500 disabled:opacity-50 rounded-lg shadow-sm transition"
                        >
                          {isBusy ? (
                            <>
                              <span className="animate-spin text-xs">⏳</span>
                              <span>{t('glossary.downloading')}</span>
                            </>
                          ) : (
                            <>
                              <CloudDownload className="w-3.5 h-3.5 text-sky-200" />
                              <span>{t('glossary.downloadAndLoad')}</span>
                            </>
                          )}
                        </button>
                      </div>
                    );
                  })
                )}
              </div>
            )}

            {/* Tab 2: Custom URL Downloader */}
            {remoteTab === 'custom' && (
              <div className="mt-4 space-y-4 flex-1">
                <div>
                  <label className="block text-xs font-medium text-slate-300 mb-1.5">
                    {t('glossary.urlLabel')}
                  </label>
                  <input
                    type="url"
                    value={customUrl}
                    onChange={(e) => setCustomUrl(e.target.value)}
                    placeholder="https://raw.githubusercontent.com/<owner>/<repo>/main/Names.txt"
                    className="w-full px-3 py-2 bg-white/4 border border-white/10 rounded-lg text-slate-200 text-xs focus:outline-none focus:border-sky-500 transition font-mono"
                  />
                  <p className="text-[11px] text-slate-400 mt-1">
                    {t('glossary.urlHelp')}
                  </p>
                </div>

                <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
                  <div>
                    <label className="block text-xs font-medium text-slate-300 mb-1.5">
                      {t('glossary.formatLabel')}
                    </label>
                    <select
                      value={customFormat}
                      onChange={(e) => setCustomFormat(e.target.value as any)}
                      className="w-full px-3 py-2 bg-[#111625] border border-white/10 rounded-lg text-slate-200 text-xs focus:outline-none focus:border-sky-500 transition"
                    >
                      <option value="vietphrase">{t('glossary.vietphraseFormat')}</option>
                      <option value="tsv">{t('glossary.tsvFormat')}</option>
                      <option value="json">{t('glossary.jsonFormat')}</option>
                    </select>
                  </div>

                  <div>
                    <label className="block text-xs font-medium text-slate-300 mb-1.5">
                      {t('glossary.defaultCategoryLabel')}
                    </label>
                    <select
                      value={customCategory}
                      onChange={(e) => setCustomCategory(e.target.value)}
                      className="w-full px-3 py-2 bg-[#111625] border border-white/10 rounded-lg text-slate-200 text-xs focus:outline-none focus:border-sky-500 transition"
                    >
                      <option value="proper_name">{t('glossary.categories.proper_name')}</option>
                      <option value="faction">{t('glossary.categories.faction')}</option>
                      <option value="rank">{t('glossary.categories.rank')}</option>
                      <option value="ability">{t('glossary.categories.ability')}</option>
                      <option value="item">{t('glossary.categories.item')}</option>
                      <option value="location">{t('glossary.categories.location')}</option>
                      <option value="other">{t('glossary.categories.other')}</option>
                    </select>
                  </div>
                </div>

                <div className="p-3 bg-sky-500/10 border border-sky-500/20 rounded-xl text-[11px] text-sky-300 leading-relaxed">
                  {t('glossary.urlTip')}
                </div>

                <button
                  disabled={!customUrl.trim() || isDownloading === customUrl}
                  onClick={() => handleDownloadRemote(customUrl, t('glossary.customUrlTitle'), customFormat, customCategory)}
                  className="w-full py-2.5 px-4 bg-sky-600 hover:bg-sky-500 disabled:opacity-50 text-white font-semibold text-xs rounded-lg shadow-sm transition flex items-center justify-center gap-2"
                >
                  {isDownloading === customUrl ? (
                    <>
                      <span className="animate-spin">⏳</span>
                      <span>{t('glossary.loadingIntoDb')}</span>
                    </>
                  ) : (
                    <>
                      <CloudDownload className="w-4 h-4" />
                      <span>{t('glossary.startDownloadBtn')}</span>
                    </>
                  )}
                </button>
              </div>
            )}

            {/* Modal Footer */}
            <div className="pt-4 mt-3 border-t border-white/10 flex items-center justify-between text-xs text-slate-400">
              <span>{t('glossary.upsertNotice')}</span>
              <button
                onClick={() => setIsRemoteModalOpen(false)}
                className="px-4 py-1.5 bg-white/4 hover:bg-white/8 text-slate-300 rounded-lg transition"
              >
                {t('glossary.close')}
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Add / Edit Modal */}
      {isModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="bg-[#111625] border border-white/10 rounded-2xl w-full max-w-md p-6 shadow-2xl animate-in fade-in zoom-in duration-150">
            <div className="flex items-center justify-between pb-4 border-b border-white/10">
              <h3 className="text-sm font-bold text-slate-100 flex items-center gap-2">
                <BookMarked className="w-4 h-4 text-indigo-400" />
                {editingTerm ? t('glossary.editTermTitle') : t('glossary.addTermTitle')}
              </h3>
              <button
                onClick={() => setIsModalOpen(false)}
                className="text-slate-400 hover:text-white p-1 rounded-md"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {formError && (
              <div className="mt-4 p-3 bg-rose-500/10 border border-rose-500/20 rounded-lg flex items-center gap-2 text-rose-400 text-xs">
                <AlertCircle className="w-4 h-4 shrink-0" />
                <span>{formError}</span>
              </div>
            )}

            <form onSubmit={handleSubmitTerm} className="mt-4 space-y-4 text-xs">
              <div>
                <label className="block text-slate-300 font-medium mb-1.5">
                  {t('glossary.sourceTermLabel')}
                </label>
                <input
                  type="text"
                  value={sourceTerm}
                  onChange={(e) => setSourceTerm(e.target.value)}
                  placeholder={t('glossary.sourceTermPlaceholder')}
                  required
                  className="w-full px-3 py-2 bg-white/4 border border-white/10 rounded-lg text-slate-100 placeholder-slate-500 focus:outline-none focus:border-indigo-500/60 font-mono text-xs"
                />
              </div>

              <div>
                <label className="block text-slate-300 font-medium mb-1.5">
                  {t('glossary.targetTermLabel')}
                </label>
                <input
                  type="text"
                  value={targetTerm}
                  onChange={(e) => setTargetTerm(e.target.value)}
                  placeholder={t('glossary.targetTermPlaceholder')}
                  required
                  className="w-full px-3 py-2 bg-white/4 border border-white/10 rounded-lg text-slate-100 placeholder-slate-500 focus:outline-none focus:border-indigo-500/60 text-xs"
                />
              </div>

              <div>
                <label className="block text-slate-300 font-medium mb-1.5">
                  {t('glossary.categoryLabel')}
                </label>
                <Select
                  value={category}
                  onChange={(e) => setCategory(e.target.value)}
                  className="w-full text-xs"
                >
                  <option value="proper_name">{currentTaxonomy.categoryLabels.proper_name}</option>
                  <option value="faction">{currentTaxonomy.categoryLabels.faction}</option>
                  <option value="rank">{currentTaxonomy.categoryLabels.rank}</option>
                  <option value="ability">{currentTaxonomy.categoryLabels.ability}</option>
                  <option value="item">{currentTaxonomy.categoryLabels.item}</option>
                  <option value="location">{currentTaxonomy.categoryLabels.location}</option>
                  <option value="other">{currentTaxonomy.categoryLabels.other}</option>
                </Select>
              </div>

              <div className="flex items-center justify-end gap-2 pt-3 border-t border-white/10">
                <button
                  type="button"
                  onClick={() => setIsModalOpen(false)}
                  className="px-3 py-1.5 bg-white/4 hover:bg-white/8 text-slate-300 rounded-lg transition"
                >
                  {t('glossary.cancel')}
                </button>
                <button
                  type="submit"
                  className="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white font-semibold rounded-lg shadow-sm transition"
                >
                  {editingTerm ? t('glossary.update') : t('glossary.create')}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Import Modal */}
      {isImportModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="bg-[#111625] border border-white/10 rounded-2xl w-full max-w-lg p-6 shadow-2xl animate-in fade-in zoom-in duration-150">
            <div className="flex items-center justify-between pb-4 border-b border-white/10">
              <h3 className="text-sm font-bold text-slate-100 flex items-center gap-2">
                <Upload className="w-4 h-4 text-indigo-400" />
                {t('glossary.importModalTitle')}
              </h3>
              <button
                onClick={() => setIsImportModalOpen(false)}
                className="text-slate-400 hover:text-white p-1 rounded-md"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="mt-4 space-y-4 text-xs">
              {/* Format selection */}
              <div className="flex items-center gap-2 flex-wrap">
                <button
                  type="button"
                  onClick={() => setImportFormat('vietphrase')}
                  className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg border text-xs font-semibold transition ${
                    importFormat === 'vietphrase'
                      ? 'bg-amber-500/20 border-amber-400 text-amber-200'
                      : 'bg-white/3 border-white/10 text-slate-400'
                  }`}
                >
                  <FileText className="w-3.5 h-3.5 text-amber-400" />
                  <span>{t('glossary.vietphraseTab')}</span>
                </button>

                <button
                  type="button"
                  onClick={() => setImportFormat('tsv')}
                  className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg border text-xs font-medium transition ${
                    importFormat === 'tsv'
                      ? 'bg-indigo-600/30 border-indigo-400 text-indigo-200'
                      : 'bg-white/3 border-white/10 text-slate-400'
                  }`}
                >
                  <FileSpreadsheet className="w-3.5 h-3.5" />
                  <span>TSV (Tab-Separated)</span>
                </button>

                <button
                  type="button"
                  onClick={() => setImportFormat('json')}
                  className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg border text-xs font-medium transition ${
                    importFormat === 'json'
                      ? 'bg-indigo-600/30 border-indigo-400 text-indigo-200'
                      : 'bg-white/3 border-white/10 text-slate-400'
                  }`}
                >
                  <FileCode className="w-3.5 h-3.5" />
                  <span>JSON Format</span>
                </button>
              </div>

              {/* Vietphrase category dropdown */}
              {importFormat === 'vietphrase' && (
                <div className="p-3 bg-white/2 border border-white/10 rounded-lg space-y-2">
                  <div className="flex items-center justify-between">
                    <label className="text-slate-300 font-medium">{t('glossary.assignDefaultCategory')}</label>
                    <select
                      value={importCategory}
                      onChange={(e) => setImportCategory(e.target.value)}
                      className="bg-[#0e1322] border border-white/10 text-slate-200 px-2 py-1 rounded text-xs"
                    >
                      <option value="proper_name">{currentTaxonomy.categoryLabels.proper_name}</option>
                      <option value="faction">{currentTaxonomy.categoryLabels.faction}</option>
                      <option value="rank">{currentTaxonomy.categoryLabels.rank}</option>
                      <option value="ability">{currentTaxonomy.categoryLabels.ability}</option>
                      <option value="item">{currentTaxonomy.categoryLabels.item}</option>
                      <option value="location">{currentTaxonomy.categoryLabels.location}</option>
                      <option value="other">{currentTaxonomy.categoryLabels.other}</option>
                    </select>
                  </div>
                  <p className="text-[11px] text-slate-400">
                    {t('glossary.vietphraseSupportNotice')}
                  </p>
                </div>
              )}

              {/* Textarea + file upload */}
              <div>
                <div className="flex items-center justify-between mb-1">
                  <label className="text-slate-300 font-medium">
                    {importFormat === 'vietphrase'
                      ? t('glossary.pastePromptVietphrase')
                      : importFormat === 'tsv'
                      ? t('glossary.pastePromptTsv')
                      : t('glossary.pastePromptJson')}
                  </label>
                  <label className="text-[11px] text-indigo-400 hover:text-indigo-300 cursor-pointer flex items-center gap-1">
                    <Upload className="w-3 h-3" />
                    <span>{t('glossary.chooseFile')}</span>
                    <input
                      type="file"
                      accept=".txt,.tsv,.json"
                      onChange={handleFileUpload}
                      className="hidden"
                    />
                  </label>
                </div>

                <textarea
                  value={importContent}
                  onChange={(e) => setImportContent(e.target.value)}
                  placeholder={
                    importFormat === 'vietphrase'
                      ? '筑基=Foundation\n金丹=Golden Core\nArasaka=Arasaka Corp\nExcalibur=Excalibur'
                      : importFormat === 'tsv'
                      ? '降龍十八掌\tEighteen Dragon Palms\tskill\n郭靖\tGuo Jing\tproper_name'
                      : '[{"source_term": "郭靖", "target_term": "Guo Jing", "category": "proper_name"}]'
                  }
                  rows={8}
                  className="w-full px-3 py-2 bg-white/3 border border-white/10 rounded-lg text-slate-200 font-mono text-[11px] leading-relaxed placeholder-slate-600 focus:outline-none focus:border-indigo-500/60"
                />
              </div>

              {importStatus && (
                <div
                  className={`p-2.5 rounded-lg text-xs flex items-center gap-2 ${
                    isImportError
                      ? 'bg-rose-500/10 text-rose-300 border border-rose-500/20'
                      : 'bg-emerald-500/10 text-emerald-300 border border-emerald-500/20'
                  }`}
                >
                  <AlertCircle className="w-3.5 h-3.5 shrink-0" />
                  <span>{importStatus}</span>
                </div>
              )}

              <div className="flex items-center justify-end gap-2 pt-3 border-t border-white/10">
                <button
                  type="button"
                  onClick={() => setIsImportModalOpen(false)}
                  className="px-3 py-1.5 bg-white/4 hover:bg-white/8 text-slate-300 rounded-lg transition"
                >
                  {t('glossary.close')}
                </button>
                <button
                  type="button"
                  onClick={handleDoImport}
                  className="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white font-semibold rounded-lg shadow-sm transition flex items-center gap-1.5"
                >
                  <Upload className="w-3.5 h-3.5" />
                  <span>{t('glossary.confirmImport')}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Export Modal */}
      {isExportModalOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4">
          <div className="bg-[#111625] border border-white/10 rounded-2xl w-full max-w-lg p-6 shadow-2xl animate-in fade-in zoom-in duration-150">
            <div className="flex items-center justify-between pb-4 border-b border-white/10">
              <h3 className="text-sm font-bold text-slate-100 flex items-center gap-2">
                <Download className="w-4 h-4 text-indigo-400" />
                {t('glossary.exportModalTitle')} ({exportFormat.toUpperCase()})
              </h3>
              <button
                onClick={() => setIsExportModalOpen(false)}
                className="text-slate-400 hover:text-white p-1 rounded-md"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            <div className="mt-4 space-y-4 text-xs">
              <div className="flex items-center gap-2">
                <button
                  type="button"
                  onClick={() => handleOpenExport('tsv')}
                  className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg border text-xs font-medium transition ${
                    exportFormat === 'tsv'
                      ? 'bg-indigo-600/30 border-indigo-400 text-indigo-200'
                      : 'bg-white/3 border-white/10 text-slate-400'
                  }`}
                >
                  <FileSpreadsheet className="w-3.5 h-3.5" />
                  <span>TSV</span>
                </button>
                <button
                  type="button"
                  onClick={() => handleOpenExport('json')}
                  className={`flex items-center gap-1.5 px-3 py-1.5 rounded-lg border text-xs font-medium transition ${
                    exportFormat === 'json'
                      ? 'bg-indigo-600/30 border-indigo-400 text-indigo-200'
                      : 'bg-white/3 border-white/10 text-slate-400'
                  }`}
                >
                  <FileCode className="w-3.5 h-3.5" />
                  <span>JSON</span>
                </button>
              </div>

              <div>
                <textarea
                  readOnly
                  value={exportContent}
                  rows={9}
                  className="w-full px-3 py-2 bg-white/3 border border-white/10 rounded-lg text-slate-200 font-mono text-[11px] leading-relaxed select-all focus:outline-none"
                />
              </div>

              <div className="flex items-center justify-between pt-3 border-t border-white/10">
                <button
                  type="button"
                  onClick={handleCopyExport}
                  className="flex items-center gap-1.5 px-3 py-1.5 bg-white/4 hover:bg-white/8 text-slate-200 border border-white/10 rounded-lg transition"
                >
                  {hasCopied ? (
                    <>
                      <Check className="w-3.5 h-3.5 text-emerald-400" />
                      <span className="text-emerald-400">{t('glossary.copiedClipboard')}</span>
                    </>
                  ) : (
                    <>
                      <Copy className="w-3.5 h-3.5 text-indigo-400" />
                      <span>{t('glossary.copyAll')}</span>
                    </>
                  )}
                </button>

                <button
                  type="button"
                  onClick={handleDownloadExport}
                  className="flex items-center gap-1.5 px-4 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white font-semibold rounded-lg shadow-sm transition"
                >
                  <Download className="w-3.5 h-3.5" />
                  <span>{t('glossary.downloadFile')}</span>
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
