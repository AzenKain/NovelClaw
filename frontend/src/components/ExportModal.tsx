import React, { useState, useEffect, useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { 
  DownloadCloud, 
  BookOpen, 
  CheckCircle2, 
  X,
  Layers,
  Sparkles,
  Check
} from 'lucide-react';
import * as dtos from '@bindings/novelclaw/internal/dtos/models';
import * as ExportService from '@bindings/novelclaw/internal/services/exportservice';
import { useAppStore } from '@/store/useAppStore';

export const ExportModal: React.FC = () => {
  const { t } = useTranslation();

  const {
    isExportOpen,
    setExportOpen,
    selectedProject,
    selectedVolume,
    chapters,
    selectedChapter,
  } = useAppStore();

  const [formats, setFormats] = useState<dtos.FormatDescriptorDTO[]>([]);
  const [selectedFormat, setSelectedFormat] = useState<string>('epub');
  const [useTranslatedOnly, setUseTranslatedOnly] = useState<boolean>(true);

  // Extract all volumes in project
  const volumes = useMemo(() => {
    const set = new Set<string>();
    chapters.forEach(ch => {
      const m = ch.title.match(/^\[(.*?)\]/);
      if (m && m[1]) set.add(m[1]);
    });
    return Array.from(set);
  }, [chapters]);

  // Determine active/detected volume from workspace context
  const detectedVolume = useMemo(() => {
    if (selectedVolume && selectedVolume !== 'all' && volumes.includes(selectedVolume)) {
      return selectedVolume;
    }
    if (selectedChapter) {
      const m = selectedChapter.title.match(/^\[(.*?)\]/);
      if (m && m[1] && volumes.includes(m[1])) {
        return m[1];
      }
    }
    return volumes.length > 0 ? volumes[0] : 'all';
  }, [selectedVolume, selectedChapter, volumes]);

  const [targetVolume, setTargetVolume] = useState<string>(detectedVolume);
  const [customTitle, setCustomTitle] = useState<string>('');
  const [customAuthor, setCustomAuthor] = useState<string>('');
  const [outputPath, setOutputPath] = useState<string>('');

  const [isExporting, setIsExporting] = useState<boolean>(false);
  const [exportResult, setExportResult] = useState<dtos.ExportResultDTO | null>(null);
  const [errorMsg, setErrorMsg] = useState<string | null>(null);
  const [imageError, setImageError] = useState<boolean>(false);

  // Sync target volume when modal opens or detected volume changes
  useEffect(() => {
    if (isExportOpen) {
      setTargetVolume(detectedVolume);
      setExportResult(null);
      setErrorMsg(null);
      setImageError(false);
    }
  }, [isExportOpen, detectedVolume]);

  // Dynamically compute default Title and OutputPath based on target volume and format
  useEffect(() => {
    if (!selectedProject) return;
    const isSeries = volumes.length > 0;
    const safe = selectedProject.title.replace(/\s+/g, '_');
    const ext = selectedFormat === 'kepub.epub' ? 'kepub.epub' : selectedFormat;

    if (isSeries && targetVolume !== 'all') {
      const volSuffix = targetVolume.replace(/^vol/i, 'Vol ');
      setCustomTitle(`${selectedProject.title} - ${volSuffix}`);
      setOutputPath(`./dist/${safe}_${targetVolume}.${ext}`);
    } else {
      setCustomTitle(selectedProject.title || '');
      setOutputPath(`./dist/${safe}.${ext}`);
    }
    setCustomAuthor(selectedProject.author || '');
  }, [selectedProject, targetVolume, selectedFormat, volumes.length]);

  useEffect(() => {
    if (!isExportOpen) return;
    const loadFormats = async () => {
      try {
        const fmts = await ExportService.ListSupportedFormats();
        if (fmts) setFormats(fmts);
      } catch (err) {
        console.error('Failed to load export formats:', err);
      }
    };
    loadFormats();
  }, [isExportOpen]);

  const handleFormatChange = (fmt: string) => {
    setSelectedFormat(fmt);
  };

  // Chapter statistics for currently selected volume scope
  const volumeStats = useMemo(() => {
    const chs = targetVolume === 'all'
      ? chapters
      : chapters.filter(c => c.title.toLowerCase().startsWith(`[${targetVolume.toLowerCase()}]`));
    const total = chs.length;
    const completed = chs.filter(c => c.status === 'completed' || (c.translated_content && c.translated_content.trim().length > 0)).length;
    return { total, completed };
  }, [chapters, targetVolume]);

  // Extract cover image URL for selected volume
  const coverUrl = useMemo(() => {
    if (!selectedProject) return '';
    const chs = targetVolume === 'all'
      ? chapters
      : chapters.filter(c => c.title.toLowerCase().startsWith(`[${targetVolume.toLowerCase()}]`));

    const extractImgUrl = (text?: string): string => {
      if (!text) return '';
      // Support <img src="..."> as well as SVG <image xlink:href="..."> or <image href="...">
      const m = text.match(/(?:src|xlink:href|href)=["']([^"']+\.(?:jpe?g|png|webp|gif|svg|avif)[^"']*)["']/i) ||
                text.match(/(?:src|xlink:href|href)=["']([^"']+)["']/i);
      if (m && m[1]) {
        let url = m[1].trim();
        // If relative URL, convert to API reader asset URL
        if (url && !url.startsWith('/api/') && !url.startsWith('http') && !url.startsWith('data:')) {
          const volPrefix = targetVolume !== 'all' ? `${targetVolume}/` : '';
          const cleanRel = url.replace(/^(\.\/|\.\.\/)+/, '');
          url = `/api/v1/reader/${encodeURIComponent(selectedProject.id)}/asset/${volPrefix}${cleanRel}`;
        }
        return url;
      }
      return '';
    };

    // 1. Try finding a chapter explicitly marked with 'cover' or 'bìa'
    const coverCh = chs.find(c => {
      const t = c.title.toLowerCase();
      return t.includes('cover') || t.includes('bìa');
    });
    if (coverCh) {
      const url = extractImgUrl(coverCh.translated_content) || extractImgUrl(coverCh.raw_content);
      if (url) return url;
    }

    // 2. Try the first chapter in the volume (many books put the cover in Chapter 1 without 'cover' in title)
    if (chs.length > 0) {
      const url = extractImgUrl(chs[0].translated_content) || extractImgUrl(chs[0].raw_content);
      if (url) return url;
    }

    // 3. Fallback: check any chapter in volume with an image
    for (const c of chs) {
      const url = extractImgUrl(c.translated_content) || extractImgUrl(c.raw_content);
      if (url) return url;
    }

    // 4. Default asset location fallback if project has extracted assets
    const vol = targetVolume !== 'all' ? targetVolume : 'vol1';
    return `/api/v1/reader/${encodeURIComponent(selectedProject.id)}/asset/${vol}/cover.jpeg`;
  }, [selectedProject, chapters, targetVolume]);

  const handleExport = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedProject) return;
    setIsExporting(true);
    setErrorMsg(null);
    setExportResult(null);

    try {
      if (selectedFormat === 'neko') {
        const res = await ExportService.ExportProjectBundle({
          project_id: selectedProject.id,
          output_path: outputPath,
        });
        if (res) setExportResult(res);
      } else {
        const coverDiskPath = coverUrl.startsWith('/api/v1/reader/')
          ? decodeURIComponent(coverUrl.replace(new RegExp(`^/api/v1/reader/[^/]+/asset/`), `data/assets/${selectedProject.id}/`))
          : '';

        const res = await ExportService.ExportEbook({
          project_id: selectedProject.id,
          target_format: selectedFormat,
          output_path: outputPath,
          custom_title: customTitle,
          custom_author: customAuthor,
          cover_image_path: coverDiskPath,
          use_translated_only: useTranslatedOnly,
          volume: targetVolume === 'all' ? '' : targetVolume,
        });
        if (res) setExportResult(res);
      }
    } catch (err: any) {
      console.error('Export failed:', err);
      setErrorMsg(err?.message || t('export.errorGeneric'));
    } finally {
      setIsExporting(false);
    }
  };

  if (!isExportOpen || !selectedProject) return null;

  return (
    <div className="fixed inset-0 bg-black/80 flex items-center justify-center z-50 p-4">
      <div className="bg-[#0c101a] border border-white/10 rounded-2xl max-w-3xl w-full p-4 sm:p-6 shadow-2xl flex flex-col max-h-[90vh] overflow-hidden animate-in fade-in zoom-in-95 duration-200">
        {/* Modal Header */}
        <div className="flex items-center justify-between pb-4 border-b border-white/8 shrink-0">
          <div>
            <h2 className="text-sm font-bold text-slate-100 flex items-center space-x-2">
              <DownloadCloud className="w-4 h-4 text-indigo-400" />
              <span>{t('export.title')}</span>
            </h2>
            <p className="text-xs text-slate-400 mt-0.5">{t('export.subtitle')}</p>
          </div>
          <button
            onClick={() => setExportOpen(false)}
            className="text-slate-400 hover:text-slate-200 p-1 rounded-lg hover:bg-white/6 transition"
          >
            <X className="w-4 h-4" />
          </button>
        </div>

        {/* Modal Body */}
        <div className="flex-1 min-h-0 overflow-y-auto py-4 space-y-5">
          {/* Series Volume Scope Selector (when project has multiple volumes) */}
          {volumes.length > 0 && (
            <div className="p-3.5 rounded-xl bg-[#090d18] border border-indigo-500/20 space-y-2.5">
              <div className="flex items-center justify-between flex-wrap gap-2">
                <div className="flex items-center space-x-2">
                  <Layers className="w-4 h-4 text-indigo-400" />
                  <span className="text-xs font-semibold text-slate-200">
                    {t('export.scopeLabel')}
                  </span>
                </div>
                {targetVolume !== 'all' && (
                  <span className="inline-flex items-center gap-1 px-2 py-0.5 rounded-full text-[10px] font-medium bg-indigo-500/10 text-indigo-300 border border-indigo-500/20">
                    <Sparkles className="w-3 h-3 text-amber-400" />
                    {t('export.autoDetectNotice', {
                      volume: targetVolume.toUpperCase(),
                      completed: volumeStats.completed,
                      total: volumeStats.total
                    })}
                  </span>
                )}
                {targetVolume === 'all' && (
                  <span className="text-[11px] text-slate-400">
                    {t('export.allVolumesNotice', { total: volumeStats.total })}
                  </span>
                )}
              </div>

              {/* Volume Pills */}
              <div className="flex items-center gap-1.5 flex-wrap">
                <button
                  type="button"
                  onClick={() => setTargetVolume('all')}
                  className={`px-3 py-1.5 rounded-lg text-xs font-medium transition flex items-center gap-1.5 ${
                    targetVolume === 'all'
                      ? 'bg-indigo-600 text-white shadow-xs font-semibold'
                      : 'bg-white/4 text-slate-300 hover:bg-white/8 border border-white/10'
                  }`}
                >
                  {targetVolume === 'all' && <Check className="w-3 h-3" />}
                  <span>{t('export.allVolumes')}</span>
                </button>

                {volumes.map(vol => {
                  const isSelected = targetVolume === vol;
                  const isDetected = detectedVolume === vol;
                  const volLabel = vol.replace(/^vol/i, 'Vol ');

                  return (
                    <button
                      key={vol}
                      type="button"
                      onClick={() => setTargetVolume(vol)}
                      className={`px-3 py-1.5 rounded-lg text-xs font-medium transition flex items-center gap-1.5 relative ${
                        isSelected
                          ? 'bg-indigo-600 text-white shadow-xs font-semibold'
                          : 'bg-white/4 text-slate-300 hover:bg-white/8 border border-white/10'
                      }`}
                    >
                      {isSelected && <Check className="w-3 h-3" />}
                      <span>{volLabel}</span>
                      {isDetected && (
                        <span className={`text-[9px] px-1 rounded font-mono ${
                          isSelected ? 'bg-white/20 text-white' : 'bg-indigo-500/20 text-indigo-300'
                        }`}>
                          {t('export.detectedBadge')}
                        </span>
                      )}
                    </button>
                  );
                })}
              </div>
            </div>
          )}

          {/* Format Selector Grid */}
          <div>
            <label className="block text-[11px] font-semibold uppercase tracking-wider text-slate-400 mb-2">
              {t('export.formats')}
            </label>
            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-4 gap-2.5">
              {formats.map(f => {
                const isSelected = selectedFormat === f.format;
                return (
                  <button
                    key={f.format}
                    type="button"
                    onClick={() => handleFormatChange(f.format)}
                    className={`p-3 rounded-xl border text-left transition flex flex-col justify-between ${
                      isSelected
                        ? 'border-indigo-500/50 bg-indigo-500/10 ring-1 ring-indigo-500/40 shadow-xs'
                        : 'border-white/8 bg-white/2 hover:border-white/20'
                    }`}
                  >
                    <div className="flex items-center justify-between mb-1">
                      <span className="text-xs font-semibold text-slate-200">{f.name}</span>
                      <span className="text-[10px] text-indigo-400 font-mono font-semibold">{f.extension}</span>
                    </div>
                    <p className="text-[10px] text-slate-400 line-clamp-2">{f.recommended_for}</p>
                  </button>
                );
              })}
            </div>
          </div>

          {/* Book Metadata & Cover Preview */}
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4 pt-1">
            {/* Cover Preview Card */}
            <div className="bg-[#090d15] border border-white/6 rounded-xl p-4 flex flex-col items-center justify-center text-center">
              <div className="w-28 h-40 bg-linear-to-b from-indigo-950/60 to-slate-900 rounded-lg border border-indigo-500/30 shadow-md flex flex-col items-center justify-center p-2.5 relative overflow-hidden">
                {coverUrl && !imageError ? (
                  <img
                    src={coverUrl}
                    alt="Cover Preview"
                    onError={() => setImageError(true)}
                    className="w-full h-full object-cover rounded shadow-inner"
                  />
                ) : (
                  <>
                    <div className="absolute top-1.5 right-1.5 text-[8px] px-1 rounded bg-indigo-500/20 text-indigo-300 font-mono">
                      NEKO
                    </div>
                    <BookOpen className="w-7 h-7 text-indigo-400 mb-2 opacity-80" />
                    <div className="text-[10px] font-semibold text-slate-200 line-clamp-2">{customTitle || t('export.defaultBookTitle')}</div>
                    <div className="text-[8px] text-slate-400 mt-1 line-clamp-1">{customAuthor || t('export.defaultAuthor')}</div>
                  </>
                )}
              </div>
              <span className="text-[10px] text-slate-500 mt-2 font-medium">{t('export.coverPreview')}</span>
            </div>

            {/* Form Fields */}
            <form onSubmit={handleExport} className="md:col-span-2 space-y-3 text-xs">
              <div>
                <label className="block text-slate-400 mb-1 font-medium">{t('export.bookTitle')}</label>
                <input
                  type="text"
                  className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-3 py-2 text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                  value={customTitle}
                  onChange={e => setCustomTitle(e.target.value)}
                  required
                />
              </div>

              <div>
                <label className="block text-slate-400 mb-1 font-medium">{t('export.author')}</label>
                <input
                  type="text"
                  className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-3 py-2 text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                  value={customAuthor}
                  onChange={e => setCustomAuthor(e.target.value)}
                />
              </div>

              <div>
                <label className="block text-slate-400 mb-1 font-medium">{t('export.outputPath')}</label>
                <input
                  type="text"
                  className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-3 py-2 text-slate-100 font-mono focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                  value={outputPath}
                  onChange={e => setOutputPath(e.target.value)}
                  required
                />
              </div>

              <div className="flex items-center space-x-2 pt-1">
                <input
                  type="checkbox"
                  id="transOnlyCheck"
                  checked={useTranslatedOnly}
                  onChange={e => setUseTranslatedOnly(e.target.checked)}
                  className="w-3.5 h-3.5 rounded border-white/20 text-indigo-600 focus:ring-indigo-500"
                />
                <label htmlFor="transOnlyCheck" className="text-slate-300 font-medium cursor-pointer">
                  {t('export.translatedOnly')}
                </label>
              </div>
            </form>
          </div>

          {/* Success or Error Report */}
          {exportResult && (
            <div className="p-3.5 bg-emerald-950/40 border border-emerald-500/30 rounded-xl text-xs space-y-1">
              <div className="flex items-center space-x-2 text-emerald-400 font-semibold">
                <CheckCircle2 className="w-4 h-4" />
                <span>{t('export.exportSuccess')}</span>
              </div>
              <div className="text-slate-300">
                <span>{t('export.fileLocation')} </span>
                <strong className="text-slate-100 font-mono">{exportResult.output_path}</strong>
              </div>
              <div className="flex items-center space-x-4 text-slate-400 text-[11px] pt-1">
                <span>{t('export.fileSize')} <strong className="text-slate-200 font-mono">{(exportResult.file_size_bytes / 1024).toFixed(1)} KB</strong></span>
                <span>{t('export.chapterCount')} <strong className="text-slate-200 font-mono">{exportResult.chapter_count}</strong></span>
                <span>{t('export.duration')} <strong className="text-slate-200 font-mono">{exportResult.duration_ms} ms</strong></span>
              </div>
            </div>
          )}

          {errorMsg && (
            <div className="p-3 bg-rose-950/40 border border-rose-500/30 rounded-xl text-xs text-rose-300">
              {errorMsg}
            </div>
          )}
        </div>

        {/* Modal Footer */}
        <div className="flex items-center justify-end space-x-2.5 pt-4 border-t border-white/8 shrink-0">
          <button
            type="button"
            onClick={() => setExportOpen(false)}
            className="px-4 py-2 bg-white/4 hover:bg-white/8 text-slate-300 rounded-lg text-xs font-medium border border-white/10 transition"
          >
            {t('common.close')}
          </button>
          <button
            type="button"
            onClick={handleExport}
            disabled={isExporting}
            className={`flex items-center space-x-1.5 px-5 py-2 rounded-lg text-xs font-semibold transition shadow-sm ${
              isExporting
                ? 'bg-white/5 text-slate-400 border border-white/10 cursor-not-allowed'
                : 'bg-indigo-600 hover:bg-indigo-500 text-white border border-indigo-400/30'
            }`}
          >
            <DownloadCloud className="w-4 h-4" />
            <span>{isExporting ? t('export.exporting') : t('export.exportBtn')}</span>
          </button>
        </div>
      </div>
    </div>
  );
};
