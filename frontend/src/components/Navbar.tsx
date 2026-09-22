import React from 'react';
import { useTranslation } from 'react-i18next';
import {
  BookOpen,
  Share2,
  Wand2,
  DownloadCloud,
  FolderPlus,
  FolderKanban,
  Settings,
  Languages,
  BookMarked,
  Gauge,
  Globe
} from 'lucide-react';
import { useAppStore } from '@/store/useAppStore';
import { Select } from '@/components/ui/Select';

/**
 * Application header.
 *
 * Layout contract (see .agents/skills/responsive-design):
 *  - `flex-nowrap` everywhere: controls never wrap onto a second row.
 *  - Three clusters with a deliberate shrink budget:
 *      1. Brand + project + Import   → shrink-0 (always fully visible)
 *      2. Navigation tabs + tools    → flex-1 min-w-0 (absorbs all pressure;
 *         tabs scroll horizontally when space runs out)
 *      3. Language + Settings        → shrink-0 (always fully visible)
 *    Giving the middle cluster the flexible budget is what stops the right
 *    cluster from being clipped on narrow windows.
 *  - Labels are progressive: icon-only below `xl`, text from `xl` (1280px),
 *    secondary labels only from `2xl` (1536px). Every icon-only control keeps
 *    a `title` tooltip so it stays discoverable.
 */

type TabId = 'workspace' | 'graph' | 'glossary' | 'world_bible' | 'benchmark';

const NAV_TABS: { id: TabId; icon: React.ReactNode; labelKey: string }[] = [
  { id: 'workspace', icon: <BookOpen className="w-3.5 h-3.5 shrink-0" />, labelKey: 'nav.workspace' },
  { id: 'graph', icon: <Share2 className="w-3.5 h-3.5 shrink-0" />, labelKey: 'nav.graph' },
  { id: 'glossary', icon: <BookMarked className="w-3.5 h-3.5 shrink-0" />, labelKey: 'nav.glossary' },
  { id: 'world_bible', icon: <Globe className="w-3.5 h-3.5 shrink-0 text-indigo-400" />, labelKey: 'nav.worldBible' },
  { id: 'benchmark', icon: <Gauge className="w-3.5 h-3.5 shrink-0" />, labelKey: 'nav.benchmark' },
];

export const Navbar: React.FC = () => {
  const { t, i18n } = useTranslation();

  const {
    activeTab,
    setActiveTab,
    setStyleScoutOpen,
    setExportOpen,
    setImportOpen,
    setSettingsOpen,
    setProjectManagerOpen,
    projects,
    selectedProject,
    selectProject,
  } = useAppStore();

  const handleLanguageChange = (lang: string) => {
    i18n.changeLanguage(lang);
    try {
      localStorage.setItem('neko_app_lang', lang);
      localStorage.setItem('i18nextLng', lang);
    } catch {
      // ignore
    }
  };

  return (
    <header className="h-12 w-full bg-[#090d16] border-b border-white/8 flex flex-nowrap items-center gap-1.5 sm:gap-2 px-2 sm:px-3 md:px-4 select-none shrink-0 z-30 overflow-hidden">
      {/* ── Cluster 1: Brand & Project Switcher (fixed budget) ── */}
      <div className="flex items-center gap-1.5 sm:gap-2 shrink-0 min-w-0">
        <div className="flex items-center gap-1.5 shrink-0">
          <img
            src="/novelclaw-logo.png"
            alt="NovelClaw Logo"
            className="w-7 h-7 rounded-lg object-cover border border-white/10 shrink-0 shadow-xs"
          />
          <span className="font-bold text-slate-100 text-xs tracking-tight whitespace-nowrap hidden md:inline">
            Novel<span className="text-indigo-400">Claw</span>
          </span>
        </div>

        <div className="h-4 w-px bg-white/10 shrink-0 hidden sm:block" />

        <div className="flex items-center gap-1 min-w-0">
          <Select
            value={selectedProject?.id || ''}
            onChange={(e) => {
              if (e.target.value === '__manage__') {
                setProjectManagerOpen(true);
                return;
              }
              const proj = projects.find(p => p.id === e.target.value);
              if (proj) selectProject(proj);
            }}
            className="w-full text-xs font-medium"
            containerClassName="shrink-0 w-[6.5rem] sm:w-32 md:w-36 xl:w-44 2xl:w-52"
            title={selectedProject?.title || t('projects.selectProject')}
          >
            {projects.length === 0 && <option value="">{t('nav.noProjects')}</option>}
            {projects.map((p) => (
              <option key={p.id} value={p.id}>
                {p.title}
              </option>
            ))}
            {projects.length > 0 && (
              <option value="__manage__">{t('nav.manageList')}</option>
            )}
          </Select>

          <button
            onClick={() => setProjectManagerOpen(true)}
            className="h-7 px-2 text-xs font-medium bg-[#111726] hover:bg-white/8 text-slate-300 hover:text-white border border-white/10 rounded-md transition whitespace-nowrap shrink-0 flex items-center gap-1"
            title={t('nav.manageTitle')}
          >
            <FolderKanban className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
            <span className="hidden 2xl:inline">{t('nav.manageProjects')}</span>
          </button>

          <button
            onClick={() => setImportOpen(true)}
            className="h-7 px-2 text-xs font-medium bg-[#111726] hover:bg-white/8 text-slate-200 border border-white/10 rounded-md transition whitespace-nowrap shrink-0 flex items-center gap-1"
            title={t('projects.importBtn')}
          >
            <FolderPlus className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
            <span className="hidden 2xl:inline">{t('nav.import')}</span>
          </button>
        </div>
      </div>

      {/* ── Cluster 2: Navigation Tabs & Studio Tools (flexible budget) ──
          This cluster absorbs every bit of layout pressure. The tab strip
          scrolls horizontally rather than pushing the right cluster off
          screen, and the scrollbar is hidden so it stays visually clean. */}
      <div className="flex-1 flex items-center justify-center gap-1.5 min-w-0 overflow-hidden">
        <nav className="h-7 flex items-center p-0.5 bg-[#111726] border border-white/10 rounded-md gap-0.5 min-w-0 max-w-full overflow-x-auto no-scrollbar">
          {NAV_TABS.map((tab) => (
            <button
              key={tab.id}
              onClick={() => setActiveTab(tab.id)}
              className={`h-full flex items-center gap-1 px-2 xl:px-2.5 text-xs rounded transition whitespace-nowrap shrink-0 ${
                activeTab === tab.id
                  ? 'bg-white/10 text-white font-semibold shadow-xs border border-white/10'
                  : 'text-slate-400 hover:text-slate-200 hover:bg-white/3 font-medium'
              }`}
              title={t(tab.labelKey)}
            >
              {tab.icon}
              <span className="hidden xl:inline">{t(tab.labelKey)}</span>
            </button>
          ))}
        </nav>

        <button
          onClick={() => setStyleScoutOpen(true)}
          className="h-7 hidden xl:flex items-center gap-1 px-2.5 text-xs font-medium text-slate-300 hover:text-white bg-[#111726] hover:bg-white/8 border border-white/10 rounded-md whitespace-nowrap shrink-0 transition"
          title={t('nav.styleScout')}
        >
          <Wand2 className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
          <span className="hidden 2xl:inline">{t('nav.styleScout')}</span>
        </button>

        <button
          onClick={() => setExportOpen(true)}
          className="h-7 hidden xl:flex items-center gap-1 px-2.5 text-xs font-medium text-slate-300 hover:text-white bg-[#111726] hover:bg-white/8 border border-white/10 rounded-md whitespace-nowrap shrink-0 transition"
          title={t('nav.export')}
        >
          <DownloadCloud className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
          <span className="hidden 2xl:inline">{t('nav.export')}</span>
        </button>
      </div>

      {/* ── Cluster 3: Language Switcher & Settings (fixed budget) ── */}
      <div className="flex items-center gap-1.5 shrink-0">
        <Select
          leftIcon={<Languages className="w-3.5 h-3.5 text-slate-400" />}
          value={i18n.language}
          onChange={(e) => handleLanguageChange(e.target.value)}
          className="pr-5 text-xs uppercase font-mono font-semibold"
          containerClassName="shrink-0"
          title={t('nav.interfaceLang')}
        >
          <option value="en">EN</option>
          <option value="vi">VI</option>
          <option value="ja">JA</option>
          <option value="zh">ZH</option>
          <option value="ko">KO</option>
        </Select>

        <button
          onClick={() => setSettingsOpen(true)}
          className="h-7 flex items-center gap-1 px-2.5 text-xs font-medium text-slate-300 hover:text-white bg-[#111726] hover:bg-white/8 border border-white/10 rounded-md transition whitespace-nowrap shrink-0"
          title={t('nav.settings')}
        >
          <Settings className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
          <span className="hidden xl:inline">{t('nav.settings')}</span>
        </button>
      </div>
    </header>
  );
};
