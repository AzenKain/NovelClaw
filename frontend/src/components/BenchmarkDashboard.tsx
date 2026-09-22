import React, { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Gauge,
  Play,
  CheckCircle2,
  DollarSign,
  Zap,
  ShieldCheck,
  Sparkles,
  Layers,
  Clock,
  Copy,
  Check,
  BookOpen,
} from 'lucide-react';
import { useAppStore } from '@/store/useAppStore';

const getTranslationModes = (t: any) => [
  {
    id: 'concurrent_dual_agent',
    name: t('benchmark.modes.concurrentDualAgent'),
    desc: t('benchmark.modes.concurrentDualAgentDesc'),
    recommended: true,
  },
  {
    id: 'single_pass',
    name: t('benchmark.modes.singlePass'),
    desc: t('benchmark.modes.singlePassDesc'),
  },
  {
    id: 'hierarchical_3pass',
    name: t('benchmark.modes.hierarchical3pass'),
    desc: t('benchmark.modes.hierarchical3passDesc'),
  },
  {
    id: 'swarm_arc_parallel',
    name: t('benchmark.modes.swarmArcParallel'),
    desc: t('benchmark.modes.swarmArcParallelDesc'),
  },
];

export const BenchmarkDashboard: React.FC = () => {
  const { t } = useTranslation();
  const translationModes = getTranslationModes(t);
  const {
    selectedProject,
    chapters,
    benchmarkReport,
    isBenchmarking,
    runBenchmark,
  } = useAppStore();

  const [selectedModes, setSelectedModes] = useState<string[]>([
    'concurrent_dual_agent',
    'single_pass',
  ]);
  const [selectedChapterIndex, setSelectedChapterIndex] = useState<number>(1);
  const [hasCopiedReport, setHasCopiedReport] = useState(false);
  const [benchmarkError, setBenchmarkError] = useState<string | null>(null);

  const toggleMode = (modeId: string) => {
    if (selectedModes.includes(modeId)) {
      if (selectedModes.length > 1) {
        setSelectedModes(selectedModes.filter((m) => m !== modeId));
      }
    } else {
      setSelectedModes([...selectedModes, modeId]);
    }
  };

  const handleStartBenchmark = async () => {
    if (!selectedProject) return;
    setBenchmarkError(null);
    try {
      await runBenchmark(selectedModes, [selectedChapterIndex]);
    } catch (err: any) {
      setBenchmarkError(err?.message || t('benchmark.errorRunning'));
    }
  };

  const handleCopyMarkdown = () => {
    if (!benchmarkReport?.markdown_report) return;
    navigator.clipboard.writeText(benchmarkReport.markdown_report);
    setHasCopiedReport(true);
    setTimeout(() => setHasCopiedReport(false), 2000);
  };

  if (!selectedProject) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center p-8 text-center text-slate-400 bg-[#0a0d14]">
        <Gauge className="w-16 h-16 text-indigo-400/40 mb-4 animate-pulse" />
        <h2 className="text-xl font-semibold text-slate-200 mb-2">{t('projects.selectProject')}</h2>
        <p className="max-w-md text-sm text-slate-500">
          {t('benchmark.subtitle')}
        </p>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col h-full overflow-hidden bg-[#0a0d14] text-slate-200">
      {/* Header */}
      <div className="px-6 py-4 border-b border-white/8 flex flex-wrap items-center justify-between gap-4 bg-[#0d121f]/50">
        <div className="flex items-center gap-2.5">
          <div className="w-8 h-8 rounded-lg bg-emerald-500/20 text-emerald-400 flex items-center justify-center border border-emerald-500/30 shadow-xs">
            <Gauge className="w-4 h-4" />
          </div>
          <div>
            <h1 className="text-base font-bold text-slate-100 flex items-center gap-2">
              {t('benchmark.title')}
              <span className="text-xs px-2 py-0.5 rounded-full bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 font-mono">
                Stage 4 Telemetry
              </span>
            </h1>
            <p className="text-xs text-slate-400">
              {t('benchmark.subtitle')}
            </p>
          </div>
        </div>

        <button
          onClick={handleStartBenchmark}
          disabled={isBenchmarking}
          className={`flex items-center gap-2 px-4 py-2 text-xs font-bold text-white rounded-lg shadow-sm transition ${
            isBenchmarking
              ? 'bg-emerald-600/50 cursor-not-allowed'
              : 'bg-emerald-600 hover:bg-emerald-500 active:scale-98 shadow-emerald-900/30'
          }`}
        >
          {isBenchmarking ? (
            <>
              <span className="animate-spin text-sm">⏳</span>
              <span>{t('benchmark.running')}</span>
            </>
          ) : (
            <>
              <Play className="w-4 h-4 fill-white" />
              <span>{t('benchmark.runBtn')}</span>
            </>
          )}
        </button>
      </div>

      {/* Main Content Area */}
      <div className="flex-1 overflow-auto p-6 space-y-6">
        {benchmarkError && (
          <div className="p-3 bg-rose-500/10 border border-rose-500/20 rounded-xl text-xs text-rose-300">
            {benchmarkError}
          </div>
        )}

        {/* Config Panel */}
        <div className="grid grid-cols-1 lg:grid-cols-3 gap-4">
          {/* Chapter Selector */}
          <div className="p-4 bg-[#0d121f] border border-white/8 rounded-xl space-y-3">
            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-1.5">
              <BookOpen className="w-3.5 h-3.5 text-indigo-400" />
              {t('benchmark.targetChapter')}
            </h3>
            <div>
              <label className="block text-xs text-slate-400 mb-1">
                {t('benchmark.selectChapterLabel')} ({chapters.length} {t('projectManager.chaptersCount')}):
              </label>
              <select
                value={selectedChapterIndex}
                onChange={(e) => setSelectedChapterIndex(Number(e.target.value))}
                className="w-full px-3 py-2 bg-white/3 border border-white/10 rounded-lg text-slate-200 text-xs focus:outline-none focus:border-indigo-500"
              >
                {chapters.map((ch) => (
                  <option key={ch.id} value={ch.chapter_index}>
                    {t('projects.chapters')} {ch.chapter_index}: {ch.title}
                  </option>
                ))}
              </select>
            </div>
          </div>

          {/* Modes Selection */}
          <div className="lg:col-span-2 p-4 bg-[#0d121f] border border-white/8 rounded-xl space-y-3">
            <h3 className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-1.5">
              <Layers className="w-3.5 h-3.5 text-emerald-400" />
              {t('benchmark.selectModesToCompare')}
            </h3>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
              {translationModes.map((mode) => {
                const isSelected = selectedModes.includes(mode.id);
                return (
                  <div
                    key={mode.id}
                    onClick={() => toggleMode(mode.id)}
                    className={`p-3 rounded-lg border cursor-pointer transition select-none ${
                      isSelected
                        ? 'bg-emerald-500/10 border-emerald-500/40 text-slate-100'
                        : 'bg-white/2 border-white/6-slate-400 hover:bg-white/4'
                    }`}
                  >
                    <div className="flex items-center justify-between mb-1">
                      <span className="font-semibold text-xs flex items-center gap-1.5">
                        <CheckCircle2
                          className={`w-3.5 h-3.5 ${
                            isSelected ? 'text-emerald-400' : 'text-slate-600'
                          }`}
                        />
                        {mode.name}
                      </span>
                      {mode.recommended && (
                        <span className="text-[10px] px-1.5 py-0.5 rounded bg-indigo-500/20 text-indigo-300 font-mono">
                          {t('benchmark.recommended')}
                        </span>
                      )}
                    </div>
                    <p className="text-[11px] text-slate-400 line-clamp-2 leading-relaxed">
                      {mode.desc}
                    </p>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* Results Section */}
        {benchmarkReport ? (
          <div className="space-y-6 animate-in fade-in duration-200">
            {/* Summary Highlights */}
            <div className="flex items-center justify-between">
              <h2 className="text-sm font-bold text-slate-100 flex items-center gap-2">
                <Sparkles className="w-4 h-4 text-emerald-400" />
                {t('benchmark.detailedResults')} ({(benchmarkReport.results || []).length} {t('benchmark.modesCount')})
              </h2>
              <button
                onClick={handleCopyMarkdown}
                className="flex items-center gap-1.5 px-3 py-1.5 bg-white/4 hover:bg-white/8 text-slate-300 rounded-lg text-xs border border-white/10 transition"
              >
                {hasCopiedReport ? (
                  <>
                    <Check className="w-3.5 h-3.5 text-emerald-400" />
                    <span className="text-emerald-400">{t('benchmark.copied')}</span>
                  </>
                ) : (
                  <>
                    <Copy className="w-3.5 h-3.5 text-slate-400" />
                    <span>{t('benchmark.copyReport')}</span>
                  </>
                )}
              </button>
            </div>

            {/* Metrics Comparison Cards */}
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4">
              {(benchmarkReport.results || []).map((res) => (
                <div
                  key={res.mode}
                  className="p-4 bg-[#0d121f] border border-white/8 rounded-xl space-y-3"
                >
                  <div className="flex items-center justify-between">
                    <span className="font-bold text-xs text-indigo-300 font-mono">
                      {res.mode}
                    </span>
                    <span className="text-[10px] px-1.5 py-0.5 rounded bg-white/6 text-slate-300 font-mono">
                      {res.duration_ms} ms
                    </span>
                  </div>

                  <div className="space-y-2 text-xs">
                    <div className="flex items-center justify-between text-slate-400">
                      <span className="flex items-center gap-1">
                        <Zap className="w-3 h-3 text-amber-400" /> {t('benchmark.speed')}
                      </span>
                      <span className="font-mono font-bold text-slate-200">
                        {((res.tokens_per_second || (res.duration_ms ? (res.total_tokens / (res.duration_ms / 1000)) : 0)) || res.runes_per_second).toFixed(1)} tok/s
                        {res.runes_per_second > 0 && (
                          <span className="text-[10px] text-slate-400 font-normal ml-1">({res.runes_per_second.toFixed(1)} r/s)</span>
                        )}
                      </span>
                    </div>

                    <div className="flex items-center justify-between text-slate-400">
                      <span className="flex items-center gap-1">
                        <DollarSign className="w-3 h-3 text-emerald-400" /> {t('benchmark.cost')}
                      </span>
                      <span className="font-mono font-bold text-emerald-400">
                        ${res.estimated_cost_usd.toFixed(4)}
                      </span>
                    </div>

                    <div className="flex items-center justify-between text-slate-400">
                      <span className="flex items-center gap-1">
                        <ShieldCheck className="w-3 h-3 text-indigo-400" /> {t('benchmark.pronounConsistency')}
                      </span>
                      <span className="font-mono font-bold text-slate-200">
                        {(res.pronoun_consistency_score * 100).toFixed(0)}%
                      </span>
                    </div>

                    <div className="flex items-center justify-between text-slate-400">
                      <span className="flex items-center gap-1">
                        <Sparkles className="w-3 h-3 text-purple-400" /> Shadow Critic
                      </span>
                      <span className="font-mono font-bold text-purple-300">
                        {res.revised_count} {t('benchmark.pts')}
                      </span>
                    </div>
                  </div>
                </div>
              ))}
            </div>

            {/* Comparison Table */}
            <div className="border border-white/8 rounded-xl overflow-hidden bg-[#0c101c]">
              <table className="w-full text-left text-xs border-collapse">
                <thead>
                  <tr className="bg-white/4 border-b border-white/8 text-slate-400 font-semibold uppercase tracking-wider">
                    <th className="py-3 px-4">{t('benchmark.modeCol')}</th>
                    <th className="py-3 px-4">{t('benchmark.speed')}</th>
                    <th className="py-3 px-4">{t('benchmark.tokens')}</th>
                    <th className="py-3 px-4">{t('benchmark.cost')}</th>
                    <th className="py-3 px-4">{t('benchmark.consistencyScore')}</th>
                    <th className="py-3 px-4">{t('benchmark.revisionScore')}</th>
                    <th className="py-3 px-4">{t('benchmark.duration')}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-white/4">
                  {(benchmarkReport.results || []).map((r) => (
                    <tr key={r.mode} className="hover:bg-white/2">
                      <td className="py-3 px-4 font-mono font-semibold text-indigo-300">
                        {r.mode}
                      </td>
                      <td className="py-3 px-4 font-mono text-slate-200">
                        {((r.tokens_per_second || (r.duration_ms ? (r.total_tokens / (r.duration_ms / 1000)) : 0)) || r.runes_per_second).toFixed(1)} tok/s
                      </td>
                      <td className="py-3 px-4 font-mono text-slate-300">
                        {r.total_tokens.toLocaleString()}
                      </td>
                      <td className="py-3 px-4 font-mono text-emerald-400 font-medium">
                        ${r.estimated_cost_usd.toFixed(4)}
                      </td>
                      <td className="py-3 px-4 font-mono text-indigo-200">
                        {(r.pronoun_consistency_score * 100).toFixed(0)}%
                      </td>
                      <td className="py-3 px-4 font-mono text-purple-300">
                        {r.revised_count}
                      </td>
                      <td className="py-3 px-4 font-mono text-slate-400">
                        {(r.duration_ms / 1000).toFixed(2)}s
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            {/* Raw Markdown Report Accordion */}
            <div className="p-4 bg-[#0d121f] border border-white/8 rounded-xl space-y-3">
              <h3 className="text-xs font-bold uppercase tracking-wider text-slate-400 flex items-center gap-1.5">
                <Clock className="w-3.5 h-3.5 text-indigo-400" />
                {t('benchmark.markdownReportTitle')}
              </h3>
              <pre className="p-3 bg-black/40 border border-white/6 rounded-lg text-slate-300 font-mono text-[11px] leading-relaxed overflow-x-auto select-all max-h-60">
                {benchmarkReport.markdown_report}
              </pre>
            </div>
          </div>
        ) : (
          <div className="flex flex-col items-center justify-center h-64 border border-dashed border-white/10 rounded-xl p-6 text-center text-slate-400">
            <Gauge className="w-12 h-12 text-slate-600 mb-3" />
            <p className="text-sm font-semibold text-slate-300 mb-1">
              {t('benchmark.noDataTitle')}
            </p>
            <p className="text-xs text-slate-500 max-w-sm mb-4">
              {t('benchmark.noDataDesc')}
            </p>
          </div>
        )}
      </div>
    </div>
  );
};
