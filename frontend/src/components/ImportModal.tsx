import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  FolderPlus,
  Layers,
  FileText,
  CheckCircle2,
  AlertTriangle,
  X,
  UploadCloud,
  Library,
  FolderOpen,
  FileUp,
  Trash2,
  Plus,
  ChevronDown,
  ChevronRight,
  BookOpen,
} from 'lucide-react';
import * as ProjectService from '@bindings/novelclaw/internal/services/projectservice';
import * as FSService from '@bindings/novelclaw/internal/services/fsservice';
import * as dtos from '@bindings/novelclaw/internal/dtos/models';
import { useAppStore } from '@/store/useAppStore';
import { Select } from '@/components/ui/Select';

export const ImportModal: React.FC = () => {
  const { t } = useTranslation();

  const {
    isImportOpen,
    setImportOpen,
    fetchProjects,
    selectProject,
    batchImportBooks,
  } = useAppStore();

  const [tab, setTab] = useState<'single' | 'batch' | 'manual'>('single');

  // Single file state
  const [filePath, setFilePath] = useState('');
  const [showManualPathInput, setShowManualPathInput] = useState(false);

  // Batch multi-file state
  const [batchFiles, setBatchFiles] = useState<string[]>([]);
  const [asSeries, setAsSeries] = useState(true);
  const [seriesName, setSeriesName] = useState('');

  // Manual project state
  const [title, setTitle] = useState('');
  const [author, setAuthor] = useState('');

  // Shared language state
  const [sourceLang, setSourceLang] = useState('ja');
  const [targetLang, setTargetLang] = useState('en');

  // Execution & Status
  const [isLoading, setIsLoading] = useState(false);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [batchResult, setBatchResult] = useState<dtos.BatchImportResultDTO | null>(null);

  if (!isImportOpen) return null;

  const getFileName = (path: string) => {
    if (!path) return '';
    const clean = path.replace(/\\/g, '/');
    return clean.substring(clean.lastIndexOf('/') + 1) || path;
  };

  const getFileExtension = (path: string) => {
    const name = getFileName(path);
    const dot = name.lastIndexOf('.');
    return dot !== -1 ? name.substring(dot + 1).toUpperCase() : 'BOOK';
  };

  const handleClose = () => {
    setImportOpen(false);
    setErrorMsg(null);
    setBatchResult(null);
  };

  // Native Single File Picker
  const handlePickSingleFile = async () => {
    try {
      const picked = await FSService.PickBookFile();
      if (picked) {
        setFilePath(picked);
        setErrorMsg(null);
      }
    } catch (err: any) {
      console.error('PickBookFile failed:', err);
    }
  };

  // Native Batch Files Picker
  const handlePickBatchFiles = async () => {
    try {
      const paths = await FSService.PickBookFiles();
      if (paths && paths.length > 0) {
        setBatchFiles((prev) => {
          const set = new Set([...prev, ...paths]);
          const newArr = Array.from(set);
          if (!seriesName && newArr.length > 0) {
            const first = newArr[0].replace(/\\/g, '/');
            const fname = first.substring(first.lastIndexOf('/') + 1);
            const base = fname.substring(0, fname.lastIndexOf('.')) || fname;
            // Strips English/Vietnamese volume prefixes from the file name.
            const cleanTitle = base.replace(/[\s\-_]*(vol(ume)?|tập)[\s\-_]*\d+/i, '').trim();
            setSeriesName(cleanTitle || base);
          }
          return newArr;
        });
        setErrorMsg(null);
      }
    } catch (err: any) {
      console.error('PickBookFiles failed:', err);
    }
  };

  // Native Folder Picker + Auto Scan
  const handlePickFolder = async () => {
    try {
      const folder = await FSService.PickFolder();
      if (folder) {
        const books = await FSService.ScanFolderForBooks(folder);
        if (books && books.length > 0) {
          setBatchFiles((prev) => {
            const set = new Set([...prev, ...books]);
            const newArr = Array.from(set);
            if (!seriesName && newArr.length > 0) {
              const first = newArr[0].replace(/\\/g, '/');
              const fname = first.substring(first.lastIndexOf('/') + 1);
              const base = fname.substring(0, fname.lastIndexOf('.')) || fname;
              // Strips English/Vietnamese volume prefixes from the file name.
              const cleanTitle = base.replace(/[\s\-_]*(vol(ume)?|tập)[\s\-_]*\d+/i, '').trim();
              setSeriesName(cleanTitle || base);
            }
            return newArr;
          });
          setErrorMsg(null);
        } else {
          setErrorMsg(t('batchImport.errFolderEmpty', { folder }));
        }
      }
    } catch (err: any) {
      console.error('PickFolder failed:', err);
    }
  };

  const handleRemoveBatchFile = (indexToRemove: number) => {
    setBatchFiles((prev) => prev.filter((_, idx) => idx !== indexToRemove));
  };

  const handleSingleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!filePath.trim()) {
      setErrorMsg(t('batchImport.errSelectFile'));
      return;
    }
    setIsLoading(true);
    setErrorMsg(null);
    try {
      const proj = await ProjectService.ImportBook({
        file_path: filePath.trim(),
        project_id: '',
        source_lang: sourceLang,
        target_lang: targetLang,
      });
      if (proj) {
        await fetchProjects();
        selectProject(proj);
        handleClose();
      }
    } catch (err: any) {
      console.error('Import failed:', err);
      setErrorMsg(err?.message || t('batchImport.errImportFailed'));
    } finally {
      setIsLoading(false);
    }
  };

  const handleBatchSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (batchFiles.length === 0) {
      setErrorMsg(t('batchImport.errSelectBatch'));
      return;
    }

    setIsLoading(true);
    setErrorMsg(null);
    setBatchResult(null);

    try {
      const res = await batchImportBooks({
        file_paths: batchFiles,
        source_lang: sourceLang,
        target_lang: targetLang,
        as_series: asSeries,
        series_name: seriesName.trim() || undefined,
      });

      if (res) {
        setBatchResult(res);
        await fetchProjects();
      }
    } catch (err: any) {
      console.error('Batch import failed:', err);
      setErrorMsg(err?.message || t('batchImport.errBatchFailed'));
    } finally {
      setIsLoading(false);
    }
  };

  const handleManualSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!title.trim()) {
      setErrorMsg(t('batchImport.errEnterTitle'));
      return;
    }
    setIsLoading(true);
    setErrorMsg(null);
    try {
      const id = `proj_${Date.now()}`;
      await ProjectService.CreateProject({
        id,
        title: title.trim(),
        author: author.trim(),
        source_lang: sourceLang,
        target_lang: targetLang,
        original_format: 'manual',
        total_chapters: 0,
      });
      const created = await ProjectService.GetProject(id);
      if (created) {
        await fetchProjects();
        selectProject(created);
        handleClose();
      }
    } catch (err: any) {
      console.error('Manual project creation failed:', err);
      setErrorMsg(err?.message || t('batchImport.errCreateManualFailed'));
    } finally {
      setIsLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
      <div className="bg-[#0c101a] border border-white/10rounded-2xl max-w-xl w-full p-6 shadow-2xl flex flex-col max-h-[90vh] overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        {/* Header */}
        <div className="flex items-center justify-between pb-3.5 border-b border-white/8 shrink-0">
          <h3 className="text-sm font-semibold text-slate-100 flex items-center space-x-2">
            <FolderPlus className="w-4 h-4 text-indigo-400" />
            <span>{t('batchImport.title')}</span>
          </h3>
          <button
            onClick={handleClose}
            className="text-slate-400 hover:text-slate-200 p-1 rounded-lg hover:bg-white/6 transition"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Tab Switcher */}
        <div className="grid grid-cols-3 bg-white/3 border border-white/8 p-0.5 rounded-lg mt-4 text-xs shrink-0 gap-0.5">
          <button
            type="button"
            onClick={() => {
              setTab('single');
              setErrorMsg(null);
              setBatchResult(null);
            }}
            className={`py-2 px-1.5 font-medium rounded-md flex items-center justify-center gap-1.5 transition whitespace-nowrap ${
              tab === 'single'
                ? 'bg-white/10 text-white font-semibold shadow-xs border border-white/10'
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            }`}
          >
            <FileText className="w-3.5 h-3.5 shrink-0" />
            <span className="truncate">{t('batchImport.singleTab')}</span>
          </button>

          <button
            type="button"
            onClick={() => {
              setTab('batch');
              setErrorMsg(null);
              setBatchResult(null);
            }}
            className={`py-2 px-1.5 font-medium rounded-md flex items-center justify-center gap-1.5 transition whitespace-nowrap ${
              tab === 'batch'
                ? 'bg-white/10 text-white font-semibold shadow-xs border border-white/10'
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            }`}
          >
            <Layers className="w-3.5 h-3.5 shrink-0" />
            <span className="truncate">{t('batchImport.batchTab')}</span>
          </button>

          <button
            type="button"
            onClick={() => {
              setTab('manual');
              setErrorMsg(null);
              setBatchResult(null);
            }}
            className={`py-2 px-1.5 font-medium rounded-md flex items-center justify-center gap-1.5 transition whitespace-nowrap ${
              tab === 'manual'
                ? 'bg-white/10 text-white font-semibold shadow-xs border border-white/10'
                : 'text-slate-400 hover:text-slate-200 border border-transparent'
            }`}
          >
            <Library className="w-3.5 h-3.5 shrink-0" />
            <span className="truncate">{t('batchImport.manualTab')}</span>
          </button>
        </div>

        {/* Form Body - Scrollable */}
        <div className="mt-4 overflow-y-auto pr-1">
          {tab === 'single' && (
            <form onSubmit={handleSingleSubmit} className="space-y-4 text-xs">
              {/* Native File Selector Card */}
              <div>
                <label className="block text-slate-300 mb-2 font-medium">
                  {t('batchImport.bookFileOnDevice')}
                </label>

                {!filePath ? (
                  <div
                    onClick={handlePickSingleFile}
                    className="border-2 border-dashed border-white/15 hover:border-indigo-500/60 bg-white/2 hover:bg-indigo-500/4 rounded-xl p-6 flex flex-col items-center justify-center cursor-pointer transition group text-center"
                  >
                    <div className="w-12 h-12 rounded-full bg-indigo-500/10 border border-indigo-500/20 flex items-center justify-center text-indigo-400 group-hover:scale-110 group-hover:bg-indigo-500/20 transition duration-200 mb-3">
                      <FileUp className="w-6 h-6" />
                    </div>
                    <span className="text-sm font-semibold text-slate-200 group-hover:text-white">
                      {t('batchImport.clickToPick')}
                    </span>
                    <span className="text-[11px] text-slate-400 mt-1">
                      {t('batchImport.supportedFormats')}
                    </span>
                    <button
                      type="button"
                      onClick={(e) => {
                        e.stopPropagation();
                        handlePickSingleFile();
                      }}
                      className="mt-3 px-3.5 py-1.5 bg-indigo-600/80 hover:bg-indigo-600 text-white rounded-lg text-xs font-medium border border-indigo-400/30 transition shadow-sm"
                    >
                      {t('batchImport.browseFile')}
                    </button>
                  </div>
                ) : (
                  <div className="p-3.5 bg-white/4 border border-white/15 rounded-xl flex items-center justify-between space-x-3">
                    <div className="flex items-center space-x-3 min-w-0">
                      <div className="w-10 h-10 rounded-lg bg-indigo-500/20 border border-indigo-500/30 flex items-center justify-center text-indigo-300 shrink-0">
                        <BookOpen className="w-5 h-5" />
                      </div>
                      <div className="min-w-0">
                        <div className="flex items-center space-x-2">
                          <span className="font-semibold text-slate-100 truncate text-xs">
                            {getFileName(filePath)}
                          </span>
                          <span className="px-1.5 py-0.5 rounded bg-indigo-500/20 text-indigo-300 font-mono text-[10px] font-bold shrink-0">
                            {getFileExtension(filePath)}
                          </span>
                        </div>
                        <p className="text-[11px] text-slate-400 font-mono truncate mt-0.5">
                          {filePath}
                        </p>
                      </div>
                    </div>

                    <div className="flex items-center space-x-2 shrink-0">
                      <button
                        type="button"
                        onClick={handlePickSingleFile}
                        className="px-2.5 py-1 bg-white/6 hover:bg-white/12 text-slate-200 rounded-lg text-[11px] font-medium border border-white/10 transition"
                      >
                        {t('batchImport.changeFile')}
                      </button>
                      <button
                        type="button"
                        onClick={() => setFilePath('')}
                        className="p-1 text-slate-400 hover:text-rose-400 rounded-lg hover:bg-rose-500/10 transition"
                        title={t('batchImport.removeFile')}
                      >
                        <Trash2 className="w-4 h-4" />
                      </button>
                    </div>
                  </div>
                )}
              </div>

              {/* Optional Advanced Path Input Toggle */}
              <div>
                <button
                  type="button"
                  onClick={() => setShowManualPathInput(!showManualPathInput)}
                  className="flex items-center space-x-1 text-[11px] text-slate-400 hover:text-slate-300 transition"
                >
                  {showManualPathInput ? (
                    <ChevronDown className="w-3.5 h-3.5" />
                  ) : (
                    <ChevronRight className="w-3.5 h-3.5" />
                  )}
                  <span>{t('batchImport.manualPathToggle')}</span>
                </button>

                {showManualPathInput && (
                  <div className="mt-2 pt-2 border-t border-white/5">
                    <input
                      type="text"
                      placeholder={t('batchImport.manualPathPlaceholder')}
                      className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-3 py-1.5 text-slate-100 font-mono text-xs focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                      value={filePath}
                      onChange={(e) => setFilePath(e.target.value)}
                    />
                  </div>
                )}
              </div>

              {/* Language Selection */}
              <div className="grid grid-cols-2 gap-3 pt-1">
                <div>
                  <label className="block text-slate-400 mb-1 font-medium">{t('batchImport.sourceLangLabel')}</label>
                  <Select
                    value={sourceLang}
                    onChange={(e) => setSourceLang(e.target.value)}
                    className="w-full py-2"
                    containerClassName="w-full"
                  >
                    <option value="ja">{t('languages.ja')}</option>
                    <option value="en">{t('languages.en')}</option>
                    <option value="zh">{t('languages.zh')}</option>
                    <option value="ko">{t('languages.ko')}</option>
                    <option value="vi">{t('languages.vi')}</option>
                  </Select>
                </div>

                <div>
                  <label className="block text-slate-400 mb-1 font-medium">{t('batchImport.targetLangLabel')}</label>
                  <Select
                    value={targetLang}
                    onChange={(e) => setTargetLang(e.target.value)}
                    className="w-full py-2"
                    containerClassName="w-full"
                  >
                    <option value="en">{t('languages.en')}</option>
                    <option value="vi">{t('languages.vi')}</option>
                    <option value="zh">{t('languages.zh')}</option>
                    <option value="ja">{t('languages.ja')}</option>
                    <option value="ko">{t('languages.ko')}</option>
                  </Select>
                </div>
              </div>

              {errorMsg && (
                <div className="p-2.5 bg-rose-950/40 border border-rose-500/30 rounded-lg text-rose-300 text-xs flex items-center space-x-1.5">
                  <AlertTriangle className="w-3.5 h-3.5 shrink-0" />
                  <span>{errorMsg}</span>
                </div>
              )}

              <div className="flex items-center justify-end space-x-2 pt-3 border-t border-white/8">
                <button
                  type="button"
                  onClick={handleClose}
                  className="px-3.5 py-1.5 bg-white/4 hover:bg-white/8 text-slate-300 rounded-lg border border-white/10 transition"
                >
                  {t('common.cancel')}
                </button>
                <button
                  type="submit"
                  disabled={isLoading || !filePath.trim()}
                  className="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg font-semibold shadow-sm border border-indigo-400/30 transition disabled:opacity-50"
                >
                  {isLoading ? t('batchImport.analyzing') : t('batchImport.singleImportBtn')}
                </button>
              </div>
            </form>
          )}

          {tab === 'batch' && (
            <form onSubmit={handleBatchSubmit} className="space-y-4 text-xs">
              {/* Batch Action Buttons */}
              <div className="grid grid-cols-2 gap-2.5">
                <button
                  type="button"
                  onClick={handlePickBatchFiles}
                  className="flex items-center justify-center space-x-2 p-3 rounded-xl border border-white/10 bg-white/3 hover:bg-indigo-600/10 hover:border-indigo-500/40 text-slate-200 hover:text-white transition group"
                >
                  <FileUp className="w-4 h-4 text-indigo-400 group-hover:scale-110 transition" />
                  <span className="font-medium text-xs">{t('batchImport.pickBatchFilesBtn')}</span>
                </button>

                <button
                  type="button"
                  onClick={handlePickFolder}
                  className="flex items-center justify-center space-x-2 p-3 rounded-xl border border-white/10 bg-white/3 hover:bg-indigo-600/10 hover:border-indigo-500/40 text-slate-200 hover:text-white transition group"
                >
                  <FolderOpen className="w-4 h-4 text-indigo-400 group-hover:scale-110 transition" />
                  <span className="font-medium text-xs">{t('batchImport.pickFolderBtn')}</span>
                </button>
              </div>

              {/* Selected Files List or Empty State */}
              <div>
                <div className="flex items-center justify-between mb-2">
                  <label className="text-slate-300 font-medium flex items-center space-x-1.5">
                    <span>{t('batchImport.selectedFilesList')}</span>
                    <span className="px-2 py-0.5 rounded-full bg-indigo-500/20 text-indigo-300 text-[10px] font-bold">
                      {batchFiles.length}
                    </span>
                  </label>
                  {batchFiles.length > 0 && (
                    <div className="flex items-center space-x-2">
                      <button
                        type="button"
                        onClick={handlePickBatchFiles}
                        className="text-[11px] text-indigo-400 hover:text-indigo-300 flex items-center space-x-1"
                      >
                        <Plus className="w-3 h-3" />
                        <span>{t('batchImport.addMore')}</span>
                      </button>
                      <button
                        type="button"
                        onClick={() => setBatchFiles([])}
                        className="text-[11px] text-slate-400 hover:text-rose-400"
                      >
                        {t('batchImport.clearAll')}
                      </button>
                    </div>
                  )}
                </div>

                {batchFiles.length === 0 ? (
                  <div className="border border-white/10 rounded-xl p-6 text-center bg-white/1">
                    <Layers className="w-8 h-8 text-slate-500 mx-auto mb-2 opacity-50" />
                    <p className="text-slate-400 font-medium">{t('batchImport.noFilesSelected')}</p>
                    <p className="text-[11px] text-slate-500 mt-1">
                      {t('batchImport.noFilesDesc')}
                    </p>
                  </div>
                ) : (
                  <div className="max-h-44 overflow-y-auto space-y-1.5 border border-white/10 rounded-xl p-2 bg-[#080c14]">
                    {batchFiles.map((file, idx) => (
                      <div
                        key={idx}
                        className="flex items-center justify-between p-2 rounded-lg bg-white/3 hover:bg-white/6 border border-white/5 transition text-xs"
                      >
                        <div className="flex items-center space-x-2.5 min-w-0 pr-2">
                          <span className="w-5 h-5 rounded-full bg-white/5 border border-white/10 flex items-center justify-center text-[10px] font-mono text-slate-400 shrink-0">
                            {idx + 1}
                          </span>
                          <span className="px-1.5 py-0.5 rounded bg-indigo-500/20 text-indigo-300 font-mono text-[9px] font-bold shrink-0">
                            {getFileExtension(file)}
                          </span>
                          <div className="min-w-0">
                            <p className="font-medium text-slate-200 truncate">{getFileName(file)}</p>
                            <p className="text-[10px] text-slate-400 font-mono truncate">{file}</p>
                          </div>
                        </div>
                        <button
                          type="button"
                          onClick={() => handleRemoveBatchFile(idx)}
                          className="p-1 text-slate-400 hover:text-rose-400 rounded hover:bg-rose-500/10 transition shrink-0"
                          title={t('batchImport.removeFile')}
                        >
                          <X className="w-3.5 h-3.5" />
                        </button>
                      </div>
                    ))}
                  </div>
                )}
              </div>

              {/* Aggregation Option */}
              <div className="p-3 bg-[#101522] rounded-xl border border-white/8 space-y-2">
                <label className="flex items-start space-x-2.5 cursor-pointer">
                  <input
                    type="checkbox"
                    checked={asSeries}
                    onChange={(e) => setAsSeries(e.target.checked)}
                    className="w-4 h-4 mt-0.5 rounded border-white/20 text-indigo-600 focus:ring-indigo-500"
                  />
                  <div>
                    <span className="font-semibold text-slate-200">
                      {t('batchImport.asSeriesLabel')}
                    </span>
                    <p className="text-[11px] text-slate-400 mt-0.5">
                      {asSeries ? t('batchImport.asSeriesDesc') : t('batchImport.separateDesc')}
                    </p>
                  </div>
                </label>

                {asSeries && (
                  <div className="pt-2">
                    <label className="block text-slate-400 text-[11px] font-medium mb-1">
                      {t('batchImport.seriesNameLabel')}
                    </label>
                    <input
                      type="text"
                      placeholder={t('batchImport.seriesPlaceholder')}
                      className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-3 py-1.5 text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500/30 text-xs transition"
                      value={seriesName}
                      onChange={(e) => setSeriesName(e.target.value)}
                    />
                  </div>
                )}
              </div>

              {/* Language Selection */}
              <div className="grid grid-cols-2 gap-3 pt-1">
                <div>
                  <label className="block text-slate-400 mb-1 font-medium">{t('batchImport.sourceLangLabel')}</label>
                  <Select
                    value={sourceLang}
                    onChange={(e) => setSourceLang(e.target.value)}
                    className="w-full py-2"
                    containerClassName="w-full"
                  >
                    <option value="ja">{t('languages.ja')}</option>
                    <option value="en">{t('languages.en')}</option>
                    <option value="zh">{t('languages.zh')}</option>
                    <option value="ko">{t('languages.ko')}</option>
                    <option value="vi">{t('languages.vi')}</option>
                  </Select>
                </div>

                <div>
                  <label className="block text-slate-400 mb-1 font-medium">{t('batchImport.targetLangLabel')}</label>
                  <Select
                    value={targetLang}
                    onChange={(e) => setTargetLang(e.target.value)}
                    className="w-full py-2"
                    containerClassName="w-full"
                  >
                    <option value="en">{t('languages.en')}</option>
                    <option value="vi">{t('languages.vi')}</option>
                    <option value="zh">{t('languages.zh')}</option>
                    <option value="ja">{t('languages.ja')}</option>
                    <option value="ko">{t('languages.ko')}</option>
                  </Select>
                </div>
              </div>

              {/* Result Banner */}
              {batchResult && (
                <div className="p-3 bg-emerald-950/40 border border-emerald-500/30 rounded-xl space-y-1.5 text-xs text-emerald-300">
                  <div className="font-semibold flex items-center space-x-1.5">
                    <CheckCircle2 className="w-4 h-4 text-emerald-400" />
                    <span>{t('batchImport.resultSummary')}</span>
                  </div>
                  <div className="text-[11px] text-slate-300 flex items-center space-x-3">
                    <span>
                      {t('batchImport.successCount')}{' '}
                      <strong className="text-emerald-400 font-mono">{batchResult.success_count}</strong>
                    </span>
                    <span>
                      {t('batchImport.failedCount')}{' '}
                      <strong className="text-rose-400 font-mono">{batchResult.failed_count}</strong>
                    </span>
                  </div>
                  {batchResult.projects && batchResult.projects.length > 0 && (
                    <div className="text-[11px] pt-1 text-slate-200">
                      {t('batchImport.importedProjects')}{' '}
                      <span className="font-semibold text-indigo-300">
                        {batchResult.projects.map((p) => p.title).join(', ')}
                      </span>
                    </div>
                  )}
                </div>
              )}

              {errorMsg && (
                <div className="p-2.5 bg-rose-950/40 border border-rose-500/30 rounded-lg text-rose-300 text-xs flex items-center space-x-1.5">
                  <AlertTriangle className="w-3.5 h-3.5 shrink-0" />
                  <span>{errorMsg}</span>
                </div>
              )}

              <div className="flex items-center justify-end space-x-2 pt-3 border-t border-white/8">
                <button
                  type="button"
                  onClick={handleClose}
                  className="px-3.5 py-1.5 bg-white/4 hover:bg-white/8 text-slate-300 rounded-lg border border-white/10 transition"
                >
                  {batchResult ? t('batchImport.done') : t('common.cancel')}
                </button>
                <button
                  type="submit"
                  disabled={isLoading || batchFiles.length === 0}
                  className="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg font-semibold shadow-sm border border-indigo-400/30 transition disabled:opacity-50 flex items-center space-x-1.5"
                >
                  <UploadCloud className="w-4 h-4" />
                  <span>{isLoading ? t('batchImport.batchProcessing') : t('batchImport.batchImportBtn')}</span>
                </button>
              </div>
            </form>
          )}

          {tab === 'manual' && (
            <form onSubmit={handleManualSubmit} className="space-y-3.5 text-xs">
              <div>
                <label className="block text-slate-400 mb-1 font-medium">{t('batchImport.titleLabel')}</label>
                <input
                  type="text"
                  placeholder={t('batchImport.titlePlaceholder')}
                  className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-3 py-2 text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                  value={title}
                  onChange={(e) => setTitle(e.target.value)}
                  required
                />
              </div>

              <div>
                <label className="block text-slate-400 mb-1 font-medium">{t('batchImport.authorLabel')}</label>
                <input
                  type="text"
                  placeholder={t('batchImport.authorPlaceholder')}
                  className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-3 py-2 text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                  value={author}
                  onChange={(e) => setAuthor(e.target.value)}
                />
              </div>

              <div className="grid grid-cols-2 gap-3 pt-1">
                <div>
                  <label className="block text-slate-400 mb-1 font-medium">{t('batchImport.sourceLangLabel')}</label>
                  <Select
                    value={sourceLang}
                    onChange={(e) => setSourceLang(e.target.value)}
                    className="w-full py-2"
                    containerClassName="w-full"
                  >
                    <option value="ja">{t('languages.ja')}</option>
                    <option value="en">{t('languages.en')}</option>
                    <option value="zh">{t('languages.zh')}</option>
                    <option value="ko">{t('languages.ko')}</option>
                    <option value="vi">{t('languages.vi')}</option>
                  </Select>
                </div>

                <div>
                  <label className="block text-slate-400 mb-1 font-medium">{t('batchImport.targetLangLabel')}</label>
                  <Select
                    value={targetLang}
                    onChange={(e) => setTargetLang(e.target.value)}
                    className="w-full py-2"
                    containerClassName="w-full"
                  >
                    <option value="en">{t('languages.en')}</option>
                    <option value="vi">{t('languages.vi')}</option>
                    <option value="zh">{t('languages.zh')}</option>
                    <option value="ja">{t('languages.ja')}</option>
                    <option value="ko">{t('languages.ko')}</option>
                  </Select>
                </div>
              </div>

              {errorMsg && (
                <div className="p-2.5 bg-rose-950/40 border border-rose-500/30 rounded-lg text-rose-300 text-xs flex items-center space-x-1.5">
                  <AlertTriangle className="w-3.5 h-3.5 shrink-0" />
                  <span>{errorMsg}</span>
                </div>
              )}

              <div className="flex items-center justify-end space-x-2 pt-3 border-t border-white/8">
                <button
                  type="button"
                  onClick={handleClose}
                  className="px-3.5 py-1.5 bg-white/4 hover:bg-white/8 text-slate-300 rounded-lg border border-white/10 transition"
                >
                  {t('common.cancel')}
                </button>
                <button
                  type="submit"
                  disabled={isLoading || !title.trim()}
                  className="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg font-semibold shadow-sm border border-indigo-400/30 transition disabled:opacity-50"
                >
                  {isLoading ? t('batchImport.creatingProject') : t('batchImport.createProjectBtn')}
                </button>
              </div>
            </form>
          )}
        </div>
      </div>
    </div>
  );
};
