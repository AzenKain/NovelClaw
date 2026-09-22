import React, { useState, useEffect, useMemo, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { toast } from 'react-hot-toast';
import { 
  Users, 
  Link2, 
  Lock, 
  Unlock, 
  Plus, 
  Edit, 
  X, 
  UserCheck,
  Network,
  LayoutGrid,
  Sparkles,
  Loader2,
  Check,
  AlertTriangle,
  BookOpen,
  Trash2,
  ArrowLeftRight,
  Heart,
  Crown,
  ShieldAlert,
  ChevronRight,
  RotateCcw
} from 'lucide-react';
import * as dtos from '@bindings/novelclaw/internal/dtos/models';
import { useAppStore } from '@/store/useAppStore';
import { Select } from '@/components/ui/Select';
import { showConfirmModal } from '@/store/confirmStore';

// Helper: parse [Nature] Tone string into distinct parts
function parseTone(toneStr?: string): { nature: string; tone: string } {
  if (!toneStr) return { nature: '', tone: '' };
  const match = toneStr.match(/^\[(.*?)\]\s*(.*)$/);
  if (match) {
    return { nature: match[1].trim(), tone: match[2].trim() };
  }
  return { nature: '', tone: toneStr.trim() };
}

function formatTone(nature: string, tone: string): string {
  const cleanNature = nature.trim();
  const cleanTone = tone.trim();
  if (cleanNature && cleanTone) {
    return `[${cleanNature}] ${cleanTone}`;
  }
  if (cleanNature) {
    return `[${cleanNature}]`;
  }
  return cleanTone;
}

// Role configuration helper for visual polish
function getRoleConfig(role?: string, t?: (key: string) => string) {
  const r = (role || '').toLowerCase();
  // Role keywords cover both English roles and legacy Vietnamese role values.
  if (r.includes('protagonist') || r.includes('main') || r.includes('chính')) {
    return {
      label: t ? t('roles.protagonist') : 'Protagonist',
      badgeBg: 'bg-amber-500/15',
      badgeText: 'text-amber-300',
      badgeBorder: 'border-amber-500/30',
      stroke: '#f59e0b',
      fill: '#451a03',
      textFill: '#fef3c7',
      ringColor: 'ring-amber-500/50',
      icon: Crown,
    };
  }
  if (r.includes('heroine') || r.includes('nữ chính') || r.includes('deuteragonist')) {
    return {
      label: t ? t('roles.heroine') : 'Main Heroine / Deuteragonist',
      badgeBg: 'bg-rose-500/15',
      badgeText: 'text-rose-300',
      badgeBorder: 'border-rose-500/30',
      stroke: '#f43f5e',
      fill: '#4c0519',
      textFill: '#ffe4e6',
      ringColor: 'ring-rose-500/50',
      icon: Heart,
    };
  }
  if (r.includes('antagonist') || r.includes('phản diện') || r.includes('villain')) {
    return {
      label: t ? t('roles.antagonist') : 'Antagonist',
      badgeBg: 'bg-red-500/15',
      badgeText: 'text-red-300',
      badgeBorder: 'border-red-500/30',
      stroke: '#ef4444',
      fill: '#450a0a',
      textFill: '#fee2e2',
      ringColor: 'ring-red-500/50',
      icon: ShieldAlert,
    };
  }
  return {
    label: t ? t('roles.supporting') : 'Supporting Character',
    badgeBg: 'bg-indigo-500/15',
    badgeText: 'text-indigo-300',
    badgeBorder: 'border-indigo-500/30',
    stroke: '#6366f1',
    fill: '#1e1b4b',
    textFill: '#e0e7ff',
    ringColor: 'ring-indigo-500/50',
    icon: UserCheck,
  };
}

export const CharacterGraph: React.FC = () => {
  const { t } = useTranslation();

  const {
    selectedProject,
    chapters,
    entities,
    relations,
    upsertRelation,
    toggleLockRelation,
    deleteRelation,
    upsertEntity,
    deleteEntity,
    autoScanEntities,
  } = useAppStore();

  // Navigation & View Mode
  const [viewMode, setViewMode] = useState<'cards' | 'canvas'>('canvas');
  const [selectedCanvasNode, setSelectedCanvasNode] = useState<string | null>(null);
  const [selectedEdgePair, setSelectedEdgePair] = useState<{ from: string; to: string } | null>(null);

  // Extract Volumes from Chapter Titles (e.g. [vol1], [vol2])
  const volumes = useMemo(() => {
    const list: string[] = [];
    const seen = new Set<string>();
    chapters.forEach(c => {
      const m = c.title.match(/^\[(.*?)\]/);
      if (m && m[1] && !seen.has(m[1])) {
        seen.add(m[1]);
        list.push(m[1]);
      }
    });
    return list;
  }, [chapters]);

  // Scope State: 'all' | 'volume' | 'chapter'
  const [scopeType, setScopeType] = useState<'all' | 'volume' | 'chapter'>('volume');
  const [selectedVolume, setSelectedVolume] = useState<string>('');
  const [selectedChapterIndex, setSelectedChapterIndex] = useState<number>(1);

  // Auto-init selectedVolume when volumes load
  useEffect(() => {
    if (volumes.length > 0) {
      if (!selectedVolume || !volumes.includes(selectedVolume)) {
        setSelectedVolume(volumes[0]);
        setScopeType('volume');
      }
    } else {
      setScopeType('all');
    }
  }, [volumes, selectedVolume]);

  // Map of Volume info: start, end, total, story start
  const volumeInfoMap = useMemo(() => {
    const map = new Map<string, { startChapter: number; endChapter: number; totalChapters: number; storyStartChapter: number }>();
    volumes.forEach(vol => {
      const prefix = `[${vol}]`;
      const volChaps = chapters.filter(c => c.title.startsWith(prefix));
      if (volChaps.length > 0) {
        let minIdx = volChaps[0].chapter_index;
        let maxIdx = volChaps[0].chapter_index;
        let storyStart = 0;
        volChaps.forEach(c => {
          if (c.chapter_index < minIdx) minIdx = c.chapter_index;
          if (c.chapter_index > maxIdx) maxIdx = c.chapter_index;
          const lower = (c.title || '').toLowerCase();
          const isNonStory = lower.includes('cover') || lower.includes('insert') || lower.includes('copyright') || lower.includes('toc') || lower.includes('newsletter');
          if (!isNonStory && storyStart === 0) {
            storyStart = c.chapter_index;
          }
        });
        map.set(vol, {
          startChapter: minIdx,
          endChapter: maxIdx,
          totalChapters: volChaps.length,
          storyStartChapter: storyStart > 0 ? storyStart : minIdx,
        });
      }
    });
    return map;
  }, [volumes, chapters]);

  // Effective Chapter for currently active scope
  const effectiveScopeChapter = useMemo(() => {
    if (scopeType === 'all') return 0;
    if (scopeType === 'volume') {
      const info = volumeInfoMap.get(selectedVolume);
      return info ? info.endChapter : 999999;
    }
    return selectedChapterIndex;
  }, [scopeType, selectedVolume, volumeInfoMap, selectedChapterIndex]);

  // Status of the currently selected volume (has been scanned & saved)
  const currentVolumeScanStatus = useMemo(() => {
    if (scopeType !== 'volume' || !selectedVolume) return null;
    const info = volumeInfoMap.get(selectedVolume);
    if (!info) return null;

    const matchingRels = relations.filter(r => r.since_chapter >= info.startChapter && r.since_chapter <= info.endChapter);
    return {
      hasBeenScanned: matchingRels.length > 0,
      relationsCount: matchingRels.length,
      startChapter: info.startChapter,
      endChapter: info.endChapter,
      storyStartChapter: info.storyStartChapter,
      totalChapters: info.totalChapters,
    };
  }, [scopeType, selectedVolume, volumeInfoMap, relations]);

  // Displayed Entities filtered by scope
  const displayedEntities = useMemo(() => {
    if (scopeType === 'all') return entities;
    const targetChap = effectiveScopeChapter;
    return entities.filter(e => (e.first_seen_chapter || 1) <= targetChap);
  }, [entities, scopeType, effectiveScopeChapter]);

  // Displayed Relations filtered by scope (effective relations at target chapter)
  const displayedRelations = useMemo(() => {
    if (scopeType === 'all') return relations;
    const targetChap = effectiveScopeChapter;
    const map = new Map<string, dtos.RelationDTO>();
    relations.forEach(r => {
      if (r.since_chapter <= targetChap) {
        const key = `${r.from_char}-->${r.to_char}`;
        const prev = map.get(key);
        if (!prev || r.since_chapter > prev.since_chapter) {
          map.set(key, r);
        }
      }
    });
    return Array.from(map.values());
  }, [relations, scopeType, effectiveScopeChapter]);

  // Storage key per volume and scope to isolate drag positions
  const scopeStorageKey = useMemo(() => {
    if (!selectedProject) return 'default';
    if (scopeType === 'volume' && selectedVolume) return `${selectedProject.id}_vol_${selectedVolume}`;
    if (scopeType === 'chapter' && selectedChapterIndex) return `${selectedProject.id}_ch_${selectedChapterIndex}`;
    return `${selectedProject.id}_all`;
  }, [selectedProject, scopeType, selectedVolume, selectedChapterIndex]);

  // Canvas Interactive Node Positions & Drag-and-Drop
  const svgRef = useRef<SVGSVGElement | null>(null);
  const [nodePositions, setNodePositions] = useState<Record<string, { x: number; y: number }>>({});
  const [draggingNode, setDraggingNode] = useState<{
    name: string;
    startX: number;
    startY: number;
    origX: number;
    origY: number;
    hasMoved: boolean;
  } | null>(null);

  // Initialize or load saved node positions per scope
  useEffect(() => {
    if (!selectedProject || displayedEntities.length === 0) return;
    let saved: Record<string, { x: number; y: number }> = {};
    try {
      const raw = localStorage.getItem(`neko_graph_pos_${scopeStorageKey}`);
      if (raw) saved = JSON.parse(raw);
    } catch {
      // ignore storage parsing error
    }

    const total = displayedEntities.length;
    const nextPositions: Record<string, { x: number; y: number }> = { ...saved };

    displayedEntities.forEach((ent, idx) => {
      const cur = nextPositions[ent.name];
      if (!cur || cur.y <= 45 || cur.x <= 45) {
        const angle = (idx / total) * 2 * Math.PI - Math.PI / 2;
        const cx = 450, cy = 325, radius = 225;
        nextPositions[ent.name] = {
          x: Math.round(cx + radius * Math.cos(angle)),
          y: Math.round(cy + radius * Math.sin(angle)),
        };
      }
    });

    setNodePositions(nextPositions);
  }, [selectedProject, scopeStorageKey, displayedEntities]);

  // Reset to auto circle layout
  const handleResetLayout = () => {
    const total = displayedEntities.length;
    const nextPositions: Record<string, { x: number; y: number }> = {};
    displayedEntities.forEach((ent, idx) => {
      const angle = (idx / total) * 2 * Math.PI - Math.PI / 2;
      const cx = 450, cy = 325, radius = 225;
      nextPositions[ent.name] = {
        x: Math.round(cx + radius * Math.cos(angle)),
        y: Math.round(cy + radius * Math.sin(angle)),
      };
    });
    setNodePositions(nextPositions);
    if (selectedProject) {
      try {
        localStorage.setItem(`neko_graph_pos_${scopeStorageKey}`, JSON.stringify(nextPositions));
      } catch {
        // ignore storage error
      }
    }
    toast.success(t('graph.toasts.resetCircle'));
  };

  // Convert mouse screen coordinates to SVG viewBox coordinates (0..900, 0..650)
  const getSVGPoint = (clientX: number, clientY: number): { x: number; y: number } => {
    if (!svgRef.current) return { x: 450, y: 325 };
    const svg = svgRef.current;
    const pt = svg.createSVGPoint();
    pt.x = clientX;
    pt.y = clientY;
    const ctm = svg.getScreenCTM();
    if (!ctm) return { x: 450, y: 325 };
    const transformed = pt.matrixTransform(ctm.inverse());
    return { x: transformed.x, y: transformed.y };
  };

  // Window-level mouse move & up listeners while dragging for rock-solid stability
  useEffect(() => {
    if (!draggingNode) return;

    const handleWindowMouseMove = (e: MouseEvent) => {
      const pt = getSVGPoint(e.clientX, e.clientY);
      const dx = pt.x - draggingNode.startX;
      const dy = pt.y - draggingNode.startY;

      if (!draggingNode.hasMoved && Math.hypot(dx, dy) > 3) {
        setDraggingNode(prev => prev ? { ...prev, hasMoved: true } : null);
      }

      const newX = Math.max(50, Math.min(850, Math.round(draggingNode.origX + dx)));
      const newY = Math.max(50, Math.min(600, Math.round(draggingNode.origY + dy)));

      setNodePositions(prev => ({
        ...prev,
        [draggingNode.name]: { x: newX, y: newY },
      }));
    };

    const handleWindowMouseUp = () => {
      if (draggingNode) {
        if (!draggingNode.hasMoved) {
          setSelectedCanvasNode(prev => prev === draggingNode.name ? null : draggingNode.name);
          setSelectedEdgePair(null);
        } else {
          if (selectedProject) {
            try {
              localStorage.setItem(`neko_graph_pos_${scopeStorageKey}`, JSON.stringify(nodePositions));
            } catch {
              // ignore storage write error
            }
          }
        }
        setDraggingNode(null);
      }
    };

    window.addEventListener('mousemove', handleWindowMouseMove);
    window.addEventListener('mouseup', handleWindowMouseUp);
    return () => {
      window.removeEventListener('mousemove', handleWindowMouseMove);
      window.removeEventListener('mouseup', handleWindowMouseUp);
    };
  }, [draggingNode, nodePositions, selectedProject, scopeStorageKey]);

  // Edit / Add Relation Modal
  const [isEditingRelation, setIsEditingRelation] = useState(false);
  const [selectedRelation, setSelectedRelation] = useState<dtos.RelationDTO | null>(null);
  const [editFromChar, setEditFromChar] = useState('');
  const [editToChar, setEditToChar] = useState('');
  const [editRelNature, setEditRelNature] = useState('');
  const [editCallAs, setEditCallAs] = useState('');
  const [editSelfCallAs, setEditSelfCallAs] = useState('');
  const [editSinceChapter, setEditSinceChapter] = useState(1);
  const [editTone, setEditTone] = useState('');
  const [editIsLocked, setEditIsLocked] = useState(false);
  const [createReciprocal, setCreateReciprocal] = useState(false);
  const [reciprocalCallAs, setReciprocalCallAs] = useState('');
  const [reciprocalSelfCallAs, setReciprocalSelfCallAs] = useState('');

  // Edit / Add Character Modal
  const [isEditingEntity, setIsEditingEntity] = useState(false);
  const [editingEntityId, setEditingEntityId] = useState<string | null>(null);
  const [entityFormName, setEntityFormName] = useState('');
  const [entityFormAliases, setEntityFormAliases] = useState('');
  const [entityFormGender, setEntityFormGender] = useState('female');
  const [entityFormRole, setEntityFormRole] = useState('heroine');
  const [entityFormFirstSeen, setEntityFormFirstSeen] = useState(1);
  const [entityFormDescription, setEntityFormDescription] = useState('');
  const [entityFormFacts, setEntityFormFacts] = useState('');

  // Inline Quick Fact input in Dossier
  const [quickFactInput, setQuickFactInput] = useState('');

  // AI Auto-Scan State (Defaults to full volume scan)
  const [isAutoScanOpen, setIsAutoScanOpen] = useState(false);
  const [scanMode, setScanMode] = useState<'volume' | 'current_chapter' | 'all'>('volume');
  const [scanTargetVolume, setScanTargetVolume] = useState<string>('');
  const [selectedScanChapter, setSelectedScanChapter] = useState<number>(1);
  const [isScanning, setIsScanning] = useState(false);
  const [scanResult, setScanResult] = useState<dtos.AutoScanResultDTO | null>(null);
  const [scanError, setScanError] = useState<string | null>(null);

  // Automatically select the first story chapter for single chapter scanning
  useEffect(() => {
    if (chapters && chapters.length > 0) {
      const firstStory = chapters.find(c => {
        const titleLower = (c.title || '').toLowerCase();
        const nonStory = titleLower.includes('cover') || titleLower.includes('insert') || titleLower.includes('title page') || titleLower.includes('copyright') || titleLower.includes('toc') || titleLower.includes('contents') || titleLower.includes('bìa') || titleLower.includes('mục lục');
        const cLen = (c.raw_content?.length || 0) + (c.translated_content?.length || 0);
        return !nonStory && cLen > 200;
      });
      if (firstStory) {
        setSelectedScanChapter(firstStory.chapter_index);
      } else {
        setSelectedScanChapter(chapters[0].chapter_index);
      }
    }
  }, [chapters]);

  // Selected Entity Object for Dossier
  const selectedEntityObj = useMemo(() => {
    if (!selectedCanvasNode) return null;
    return displayedEntities.find(e => e.name === selectedCanvasNode) || null;
  }, [displayedEntities, selectedCanvasNode]);

  // Outgoing and Incoming relations for selected node
  const selectedNodeOutgoingRels = useMemo(() => {
    if (!selectedCanvasNode) return [];
    return displayedRelations.filter(r => r.from_char === selectedCanvasNode);
  }, [displayedRelations, selectedCanvasNode]);

  const selectedNodeIncomingRels = useMemo(() => {
    if (!selectedCanvasNode) return [];
    return displayedRelations.filter(r => r.to_char === selectedCanvasNode);
  }, [displayedRelations, selectedCanvasNode]);

  // Pair relations for Edge Inspector
  const selectedPairRels = useMemo(() => {
    if (!selectedEdgePair) return null;
    const forward = displayedRelations.find(r => r.from_char === selectedEdgePair.from && r.to_char === selectedEdgePair.to) || null;
    const reverse = displayedRelations.find(r => r.from_char === selectedEdgePair.to && r.to_char === selectedEdgePair.from) || null;
    return { forward, reverse };
  }, [displayedRelations, selectedEdgePair]);

  // Open AutoScan modal with pre-selected volume
  const handleOpenAutoScanModal = (defaultVol?: string) => {
    const targetVol = defaultVol || selectedVolume || (volumes[0] || '');
    setScanTargetVolume(targetVol);
    setScanMode(volumes.length > 0 ? 'volume' : 'all');
    setScanError(null);
    setScanResult(null);
    setIsAutoScanOpen(true);
  };

  // AI Auto Scan Handler
  const handleRunAutoScan = async (overrideMode?: 'volume' | 'current_chapter' | 'all', overrideVol?: string) => {
    const targetMode = overrideMode || scanMode;
    const targetVol = overrideVol !== undefined ? overrideVol : scanTargetVolume;
    setIsScanning(true);
    setScanError(null);
    setScanResult(null);
    try {
      const res = await autoScanEntities(selectedScanChapter, targetMode, targetVol);
      if (res) {
        if (!res.success) {
          setScanError(res.message);
          toast.error(res.message || t('graph.toasts.scanError', { err: '' }));
        } else {
          setScanResult(res);
          const timeSec = res.duration_ms ? (res.duration_ms / 1000).toFixed(1) : null;
          toast.success(res.message || t('graph.toasts.scanSuccess', { time: timeSec ? ` (${timeSec}s)` : '', entities: res.entities_found, relations: res.relations_found }));
          if (res.volume_scanned) {
            setScopeType('volume');
            setSelectedVolume(res.volume_scanned);
          }
          setTimeout(() => {
            setIsAutoScanOpen(false);
            setScanResult(null);
          }, 2200);
        }
      }
    } catch (err: any) {
      const msg = err?.message || t('graph.toasts.scanError', { err: '' });
      setScanError(msg);
      toast.error(msg);
    } finally {
      setIsScanning(false);
    }
  };

  // Delete Relation Handler
  const handleDeleteRelation = async (id: string) => {
    const confirmed = await showConfirmModal({
      title: t('graph.toasts.deleteRelTitle'),
      message: t('graph.toasts.deleteRelMsg'),
      confirmText: t('graph.toasts.deleteRelConfirm'),
      cancelText: t('common.cancel'),
      variant: 'danger',
    });
    if (!confirmed) return;
    try {
      await deleteRelation(id);
      toast.success(t('graph.toasts.deleteRelSuccess'));
      if (selectedEdgePair) setSelectedEdgePair(null);
    } catch (err: any) {
      toast.error(err?.message || t('graph.toasts.deleteRelSuccess'));
    }
  };

  // Open Edit/Create Relation
  const handleOpenEditRelation = (r?: dtos.RelationDTO, defaultFrom?: string, defaultTo?: string) => {
    if (r) {
      setSelectedRelation(r);
      setEditFromChar(r.from_char);
      setEditToChar(r.to_char);
      setEditCallAs(r.call_as);
      setEditSelfCallAs(r.self_call_as);
      setEditSinceChapter(r.since_chapter);
      const parsed = parseTone(r.tone);
      setEditRelNature(parsed.nature);
      setEditTone(parsed.tone);
      setEditIsLocked(r.is_locked);
      setCreateReciprocal(false);
    } else {
      setSelectedRelation(null);
      setEditFromChar(defaultFrom || displayedEntities[0]?.name || '');
      setEditToChar(defaultTo || (displayedEntities.length > 1 ? displayedEntities[1]?.name : ''));
      setEditCallAs('');
      setEditSelfCallAs('');
      const defaultSince = scopeType === 'volume' && currentVolumeScanStatus
        ? currentVolumeScanStatus.storyStartChapter
        : (scopeType === 'chapter' ? selectedChapterIndex : 1);
      setEditSinceChapter(defaultSince);
      setEditRelNature('');
      setEditTone('intimate');
      setEditIsLocked(true);
      setCreateReciprocal(false);
      setReciprocalCallAs('');
      setReciprocalSelfCallAs('');
    }
  };

  // Save Relation Handler
  const handleSaveRelation = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedProject || !editFromChar.trim() || !editToChar.trim()) return;

    const fullTone = formatTone(editRelNature, editTone);

    await upsertRelation({
      id: selectedRelation?.id || `rel_${Date.now()}`,
      project_id: selectedProject.id,
      from_char: editFromChar.trim(),
      to_char: editToChar.trim(),
      call_as: editCallAs.trim(),
      self_call_as: editSelfCallAs.trim(),
      since_chapter: editSinceChapter,
      tone: fullTone,
      is_locked: editIsLocked,
    });

    // Optionally create reciprocal relationship
    if (createReciprocal && reciprocalCallAs.trim()) {
      await upsertRelation({
        id: `rel_${Date.now()}_rev`,
        project_id: selectedProject.id,
        from_char: editToChar.trim(),
        to_char: editFromChar.trim(),
        call_as: reciprocalCallAs.trim(),
        self_call_as: reciprocalSelfCallAs.trim() || editCallAs.trim(),
        since_chapter: editSinceChapter,
        tone: fullTone,
        is_locked: editIsLocked,
      });
      toast.success(t('graph.toasts.savedBidirectional', { from: editFromChar, to: editToChar }));
    } else {
      toast.success(t('graph.toasts.savedUnidirectional', { from: editFromChar, to: editToChar }));
    }

    setIsEditingRelation(false);
  };

  // Open Edit/Create Character Modal
  const handleOpenEditEntity = (ent?: dtos.EntityDTO) => {
    if (ent) {
      setEditingEntityId(ent.id);
      setEntityFormName(ent.name);
      setEntityFormAliases(ent.aliases?.join(', ') || '');
      setEntityFormGender(ent.gender || 'female');
      setEntityFormRole(ent.role || 'supporting');
      setEntityFormFirstSeen(ent.first_seen_chapter || 1);
      const desc = ent.metadata?.description || '';
      const facts = Array.isArray(ent.metadata?.facts) ? ent.metadata.facts.join('\n') : '';
      setEntityFormDescription(desc);
      setEntityFormFacts(facts);
    } else {
      setEditingEntityId(null);
      setEntityFormName('');
      setEntityFormAliases('');
      setEntityFormGender('female');
      setEntityFormRole('heroine');
      setEntityFormFirstSeen(1);
      setEntityFormDescription('');
      setEntityFormFacts('');
    }
    setIsEditingEntity(true);
  };

  // Save Character Handler
  const handleSaveEntity = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!selectedProject || !entityFormName.trim()) return;

    const aliasList = entityFormAliases
      .split(',')
      .map(s => s.trim())
      .filter(Boolean);

    const factList = entityFormFacts
      .split('\n')
      .map(s => s.trim())
      .filter(Boolean);

    const metadata: Record<string, any> = {};
    if (entityFormDescription.trim()) {
      metadata.description = entityFormDescription.trim();
    }
    if (factList.length > 0) {
      metadata.facts = factList;
    }

    await upsertEntity({
      id: editingEntityId || `ent_${Date.now()}`,
      project_id: selectedProject.id,
      name: entityFormName.trim(),
      aliases: aliasList,
      category: 'character',
      gender: entityFormGender,
      role: entityFormRole,
      first_seen_chapter: entityFormFirstSeen || 1,
      metadata,
    });

    toast.success(editingEntityId ? t('graph.toasts.charSaved', { name: entityFormName }) : t('graph.toasts.charCreated', { name: entityFormName }));
    setIsEditingEntity(false);
  };

  // Delete Character Handler
  const handleDeleteEntity = async (ent: dtos.EntityDTO) => {
    const confirmed = await showConfirmModal({
      title: t('graph.toasts.deleteCharTitle'),
      message: t('graph.toasts.deleteCharMsg', { name: ent.name }),
      confirmText: t('graph.toasts.deleteCharConfirm'),
      cancelText: t('common.cancel'),
      variant: 'danger',
    });
    if (!confirmed) return;
    try {
      await deleteEntity(ent.id);
      if (selectedCanvasNode === ent.name) {
        setSelectedCanvasNode(null);
      }
      toast.success(t('graph.toasts.charDeleted', { name: ent.name }));
    } catch (err: any) {
      toast.error(err?.message || t('common.error'));
    }
  };

  // Quick Inline Add Fact for Selected Entity
  const handleAddQuickFact = async () => {
    if (!selectedEntityObj || !quickFactInput.trim() || !selectedProject) return;

    const currentFacts: string[] = Array.isArray(selectedEntityObj.metadata?.facts) 
      ? [...selectedEntityObj.metadata.facts] 
      : [];

    currentFacts.push(quickFactInput.trim());

    await upsertEntity({
      id: selectedEntityObj.id,
      project_id: selectedProject.id,
      name: selectedEntityObj.name,
      aliases: selectedEntityObj.aliases || [],
      category: selectedEntityObj.category || 'character',
      gender: selectedEntityObj.gender || 'unknown',
      role: selectedEntityObj.role || 'supporting',
      first_seen_chapter: selectedEntityObj.first_seen_chapter || 1,
      metadata: {
        ...selectedEntityObj.metadata,
        facts: currentFacts,
      },
    });

    toast.success(t('graph.toasts.factAdded', { name: selectedEntityObj.name }));
    setQuickFactInput('');
  };

  // Quick Remove Fact for Selected Entity
  const handleRemoveFact = async (factIndex: number) => {
    if (!selectedEntityObj || !selectedProject) return;

    const currentFacts: string[] = Array.isArray(selectedEntityObj.metadata?.facts) 
      ? [...selectedEntityObj.metadata.facts] 
      : [];

    currentFacts.splice(factIndex, 1);

    await upsertEntity({
      id: selectedEntityObj.id,
      project_id: selectedProject.id,
      name: selectedEntityObj.name,
      aliases: selectedEntityObj.aliases || [],
      category: selectedEntityObj.category || 'character',
      gender: selectedEntityObj.gender || 'unknown',
      role: selectedEntityObj.role || 'supporting',
      first_seen_chapter: selectedEntityObj.first_seen_chapter || 1,
      metadata: {
        ...selectedEntityObj.metadata,
        facts: currentFacts,
      },
    });

    toast.success(t('graph.toasts.factDeleted'));
  };

  if (!selectedProject) {
    return (
      <div className="flex-1 flex flex-col items-center justify-center p-8 text-slate-400">
        <Users className="w-12 h-12 text-slate-600 mb-3" />
        <h2 className="text-base font-semibold text-slate-200">{t('projects.selectProject')}</h2>
      </div>
    );
  }

  return (
    <div className="flex-1 flex flex-col h-full min-h-0 bg-[#090d16] overflow-hidden">
      {/* Top Bar - Rigid Studio Layout */}
      <div className="h-12 px-2.5 sm:px-4 border-b border-white/8 flex items-center justify-between bg-[#0a0e17] shrink-0 gap-2 select-none">
        {/* Left: Brand / Title + Scope Switcher */}
        <div className="flex items-center space-x-2 sm:space-x-2.5 min-w-0 flex-1 overflow-x-auto no-scrollbar py-1">
          <div className="flex items-center space-x-1.5 shrink-0">
            <Users className="w-4 h-4 text-indigo-400 shrink-0" />
            <span className="text-xs font-bold text-slate-100 uppercase tracking-wider hidden xl:inline">{t('graph.titleShort', t('graph.title'))}</span>
            <span className="text-xs font-bold text-slate-100 uppercase tracking-wider xl:hidden hidden sm:inline">L2 Graph</span>
          </div>

          <div className="h-4 w-px bg-white/10 shrink-0" />

          {/* Scope Pill Switcher */}
          <div className="flex items-center p-0.5 bg-white/4 border border-white/10 rounded-lg shrink-0 gap-0.5 text-xs">
            <button
              onClick={() => {
                setScopeType('volume');
                if (volumes.length > 0 && !selectedVolume) setSelectedVolume(volumes[0]);
              }}
              className={`px-2 py-0.5 rounded-md font-medium transition ${
                scopeType === 'volume'
                  ? 'bg-indigo-600 text-white shadow-xs font-semibold'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <span className="hidden sm:inline">{t('graph.scopeVolume')}</span>
              <span className="sm:hidden">Vol</span>
            </button>
            <button
              onClick={() => setScopeType('chapter')}
              className={`px-2 py-0.5 rounded-md font-medium transition ${
                scopeType === 'chapter'
                  ? 'bg-indigo-600 text-white shadow-xs font-semibold'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <span className="hidden sm:inline">{t('graph.scopeChapter')}</span>
              <span className="sm:hidden">Ch</span>
            </button>
            <button
              onClick={() => setScopeType('all')}
              className={`px-2 py-0.5 rounded-md font-medium transition ${
                scopeType === 'all'
                  ? 'bg-indigo-600 text-white shadow-xs font-semibold'
                  : 'text-slate-400 hover:text-slate-200'
              }`}
            >
              <span className="hidden sm:inline">{t('graph.scopeAll')}</span>
              <span className="sm:hidden">All</span>
            </button>
          </div>

          {/* Scope Active Selector Dropdown & Status */}
          {scopeType === 'volume' && volumes.length > 0 && (
            <div className="flex items-center space-x-1.5 shrink-0">
              <Select
                value={selectedVolume}
                onChange={(e) => setSelectedVolume(e.target.value)}
                className="h-7 text-xs bg-[#111726] border-white/10 text-slate-200 py-0 px-2 rounded-lg max-w-32.5 sm:max-w-42.5 truncate"
              >
                {volumes.map(vol => {
                  const info = volumeInfoMap.get(vol);
                  return (
                    <option key={vol} value={vol}>
                      {t('graph.volumeLabel', { vol })} {info ? `(${info.totalChapters} ${t('projectManager.chaptersCount')})` : ''}
                    </option>
                  );
                })}
              </Select>

              {/* Status Badge */}
              {currentVolumeScanStatus?.hasBeenScanned ? (
                <span className="hidden xl:inline-flex items-center space-x-1 px-2 py-0.5 rounded-full text-[10.5px] font-medium bg-emerald-500/15 text-emerald-300 border border-emerald-500/30 whitespace-nowrap">
                  <Check className="w-3 h-3 text-emerald-400 shrink-0" />
                  <span>{t('graph.volumeSaved', { vol: selectedVolume, count: currentVolumeScanStatus.relationsCount })}</span>
                </span>
              ) : (
                <button
                  onClick={() => handleOpenAutoScanModal(selectedVolume)}
                  className="hidden xl:inline-flex items-center space-x-1 px-2 py-0.5 rounded-full text-[10.5px] font-medium bg-amber-500/15 hover:bg-amber-500/25 text-amber-300 border border-amber-500/40 whitespace-nowrap transition cursor-pointer"
                  title={t('graph.volumeUnscannedTooltip')}
                >
                  <AlertTriangle className="w-3 h-3 text-amber-400 shrink-0" />
                  <span>{t('graph.volumeUnscannedBtn', { vol: selectedVolume })}</span>
                </button>
              )}
            </div>
          )}

          {scopeType === 'chapter' && chapters.length > 0 && (
            <div className="flex items-center space-x-1.5 shrink-0 max-w-xs">
              <Select
                value={String(selectedChapterIndex)}
                onChange={(e) => setSelectedChapterIndex(Number(e.target.value))}
                className="h-7 text-xs bg-[#111726] border-white/10 text-slate-200 py-0 px-2 rounded-lg truncate max-w-35 sm:max-w-45"
              >
                {chapters.map(c => (
                  <option key={c.id} value={c.chapter_index}>
                    Ch.{c.chapter_index}: {c.title}
                  </option>
                ))}
              </Select>
              <span className="text-[10.5px] text-slate-400 whitespace-nowrap hidden 2xl:inline">
                {t('graph.effectiveUntil', { ch: selectedChapterIndex })}
              </span>
            </div>
          )}

          {scopeType === 'all' && (
            <span className="text-[11px] text-slate-400 font-medium whitespace-nowrap hidden xl:inline">
              {t('graph.storyOverview', { entities: displayedEntities.length, relations: displayedRelations.length })}
            </span>
          )}
        </div>

        {/* Right: Actions */}
        <div className="flex items-center space-x-1.5 sm:space-x-2 shrink-0 ml-auto">
          {/* View Mode Switcher */}
          <div className="flex items-center p-0.5 bg-white/4 border border-white/10 rounded-lg shrink-0 gap-0.5 text-xs">
            <button
              onClick={() => setViewMode('canvas')}
              className={`flex items-center gap-1.5 px-2.5 py-1 rounded-md transition ${
                viewMode === 'canvas'
                  ? 'bg-indigo-600 text-white font-semibold shadow-xs'
                  : 'text-slate-400 hover:text-white'
              }`}
              title={t('graph.diagramViewTooltip')}
            >
              <Network className="w-3.5 h-3.5 shrink-0" />
              <span className="hidden 2xl:inline">{t('graph.diagramView')}</span>
            </button>
            <button
              onClick={() => setViewMode('cards')}
              className={`flex items-center gap-1.5 px-2.5 py-1 rounded-md transition ${
                viewMode === 'cards'
                  ? 'bg-indigo-600 text-white font-semibold shadow-xs'
                  : 'text-slate-400 hover:text-white'
              }`}
              title={t('graph.listViewTooltip')}
            >
              <LayoutGrid className="w-3.5 h-3.5 shrink-0" />
              <span className="hidden 2xl:inline">{t('graph.listView')}</span>
            </button>
          </div>

          <div className="h-4 w-px bg-white/10 shrink-0 hidden sm:block" />

          {viewMode === 'canvas' && (
            <button
              onClick={handleResetLayout}
              className="p-1.5 bg-white/4 hover:bg-white/8 text-slate-300 hover:text-white border border-white/10 rounded-lg text-xs transition shrink-0"
              title={t('graph.resetCircleTooltip')}
            >
              <RotateCcw className="w-3.5 h-3.5 shrink-0" />
            </button>
          )}

          <button
            onClick={() => handleOpenEditEntity()}
            className="flex items-center space-x-1 px-2.5 py-1.5 bg-white/4 hover:bg-white/8 text-slate-200 border border-white/10 rounded-lg text-xs font-medium whitespace-nowrap shrink-0 transition"
            title={t('graph.addCharacterBtn')}
          >
            <Plus className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
            <span className="hidden 2xl:inline">{t('graph.addCharacterBtn')}</span>
          </button>

          <button
            onClick={() => handleOpenEditRelation()}
            className="flex items-center space-x-1 px-2.5 py-1.5 bg-white/4 hover:bg-white/8 text-slate-200 border border-white/10 rounded-lg text-xs font-medium whitespace-nowrap shrink-0 transition"
            title={t('graph.addRelationBtn')}
          >
            <Link2 className="w-3.5 h-3.5 text-indigo-400 shrink-0" />
            <span className="hidden 2xl:inline">{t('graph.addRelationBtn')}</span>
          </button>

          {/* AI Auto-Scan Primary Button */}
          <button
            onClick={() => handleOpenAutoScanModal(selectedVolume)}
            className="flex items-center space-x-1.5 px-2.5 sm:px-3 py-1.5 bg-linear-to-r from-indigo-600 via-purple-600 to-pink-600 hover:from-indigo-500 hover:to-purple-500 text-white rounded-lg text-xs font-semibold shadow-md border border-indigo-400/30 whitespace-nowrap shrink-0 transition active:scale-[0.98]"
            title={t('graph.aiVolumeScanTooltip')}
          >
            <Sparkles className="w-3.5 h-3.5 text-amber-300 animate-pulse shrink-0" />
            <span className="hidden sm:inline">{t('graph.aiVolumeScan')}</span>
            <span className="sm:hidden">Scan</span>
          </button>
        </div>
      </div>

      {/* Main Container */}
      {viewMode === 'canvas' ? (
        <div className="flex-1 flex overflow-hidden relative bg-[#090d16]">
          {/* Canvas Area */}
          <div className="flex-1 flex flex-col items-center justify-center p-2 relative select-none min-w-0 overflow-hidden">
            {displayedEntities.length === 0 ? (
              <div className="text-center text-slate-400 max-w-md mx-auto p-6 bg-[#0c101b] border border-white/8 rounded-2xl shadow-xl">
                <div className="w-14 h-14 mx-auto mb-3 rounded-2xl bg-indigo-500/15 border border-indigo-500/30 flex items-center justify-center text-indigo-400">
                  <Sparkles className="w-7 h-7 text-indigo-300 animate-pulse" />
                </div>
                <h4 className="text-sm font-bold text-white mb-1.5">
                  {scopeType === 'volume' && selectedVolume
                    ? t('graph.emptyVolTitle', { vol: selectedVolume })
                    : scopeType === 'chapter'
                    ? t('graph.emptyChapterTitle', { ch: selectedChapterIndex })
                    : t('graph.emptyAllTitle')}
                </h4>
                <p className="text-xs text-slate-400 leading-relaxed mb-4">
                  {scopeType === 'volume' && selectedVolume
                    ? t('graph.emptyVolDesc', { vol: selectedVolume })
                    : t('graph.emptyAllDesc')}
                </p>
                <div className="flex items-center justify-center gap-2 flex-wrap">
                  <button
                    onClick={() => handleRunAutoScan('volume', selectedVolume || undefined)}
                    disabled={isScanning}
                    className="px-4 py-2 bg-linear-to-r from-indigo-600 to-purple-600 hover:from-indigo-500 hover:to-purple-500 text-white rounded-xl text-xs font-semibold shadow-md flex items-center space-x-2 transition disabled:opacity-50"
                  >
                    {isScanning ? (
                      <>
                        <Loader2 className="w-4 h-4 animate-spin text-white" />
                        <span>{t('graph.scanningVol', { vol: selectedVolume || '1' })}</span>
                      </>
                    ) : (
                      <>
                        <Sparkles className="w-4 h-4 text-amber-300" />
                        <span>{t('graph.scanVolBtn', { vol: selectedVolume || '1' })}</span>
                      </>
                    )}
                  </button>
                  <button
                    onClick={() => handleOpenEditEntity()}
                    className="px-3 py-2 bg-white/5 hover:bg-white/10 text-slate-300 rounded-xl text-xs font-medium border border-white/10 transition"
                  >
                    {t('graph.manualAddBtn')}
                  </button>
                </div>
              </div>
            ) : (
              <div className="relative w-full h-full flex items-center justify-center">
                <svg
                  ref={svgRef}
                  viewBox="0 0 900 650"
                  className={`w-full h-full max-h-[82vh] drop-shadow-md select-none ${draggingNode ? 'cursor-grabbing' : ''}`}
                >
                  <defs>
                    <marker
                      id="arrowhead-normal"
                      markerWidth="9"
                      markerHeight="7"
                      refX="10"
                      refY="3.5"
                      orient="auto"
                    >
                      <polygon points="0 0, 9 3.5, 0 7" fill="#6366f1" />
                    </marker>
                    <marker
                      id="arrowhead-highlight"
                      markerWidth="9"
                      markerHeight="7"
                      refX="10"
                      refY="3.5"
                      orient="auto"
                    >
                      <polygon points="0 0, 9 3.5, 0 7" fill="#ec4899" />
                    </marker>
                    <filter id="badge-glow" x="-20%" y="-20%" width="140%" height="140%">
                      <feDropShadow dx="0" dy="2" stdDeviation="3" floodColor="#000000" floodOpacity="0.6" />
                    </filter>
                  </defs>

                  {/* Curving Edges with Bézier curvature */}
                  {displayedRelations.map((rel) => {
                    const pos1 = nodePositions[rel.from_char] || { x: 450, y: 325 };
                    const pos2 = nodePositions[rel.to_char] || { x: 450, y: 325 };
                    const x1 = pos1.x;
                    const y1 = pos1.y;
                    const x2 = pos2.x;
                    const y2 = pos2.y;

                    // Check bidirectional pair
                    const isBidirectional = displayedRelations.some(
                      r => r.from_char === rel.to_char && r.to_char === rel.from_char
                    );

                    const dx = x2 - x1;
                    const dy = y2 - y1;
                    const dist = Math.hypot(dx, dy);
                    if (dist === 0) return null;

                    // Normal perpendicular vector
                    const nx = -dy / dist;
                    const ny = dx / dist;

                    // Node radius offset
                    const nodeRadius = 26;
                    const startX = x1 + (dx / dist) * nodeRadius;
                    const startY = y1 + (dy / dist) * nodeRadius;
                    const endX = x2 - (dx / dist) * (nodeRadius + 6);
                    const endY = y2 - (dy / dist) * (nodeRadius + 6);

                    let pathD: string;
                    let apexX: number;
                    let apexY: number;

                    if (isBidirectional) {
                      // Curvature offset ensures both directions curve on opposite sides
                      const curvature = Math.min(58, Math.max(40, dist * 0.18));
                      const midX = (startX + endX) / 2;
                      const midY = (startY + endY) / 2;
                      const ctrlX = midX + nx * curvature;
                      const ctrlY = midY + ny * curvature;

                      pathD = `M ${startX} ${startY} Q ${ctrlX} ${ctrlY}, ${endX} ${endY}`;
                      
                      // Position badge at t = 0.38 along curve (closer to speaker/originator node)
                      // This staggers bidirectional badges so they NEVER collide with each other!
                      const t = 0.38;
                      const oneMinusT = 1 - t;
                      apexX = oneMinusT * oneMinusT * startX + 2 * oneMinusT * t * ctrlX + t * t * endX;
                      apexY = oneMinusT * oneMinusT * startY + 2 * oneMinusT * t * ctrlY + t * t * endY;
                    } else {
                      pathD = `M ${startX} ${startY} L ${endX} ${endY}`;
                      apexX = 0.5 * startX + 0.5 * endX;
                      apexY = 0.5 * startY + 0.5 * endY;
                    }

                    const isNodeFocus = selectedCanvasNode && (selectedCanvasNode === rel.from_char || selectedCanvasNode === rel.to_char);
                    const isEdgeFocus = selectedEdgePair && 
                      ((selectedEdgePair.from === rel.from_char && selectedEdgePair.to === rel.to_char) ||
                       (selectedEdgePair.from === rel.to_char && selectedEdgePair.to === rel.from_char));

                    const isHighlighted = isNodeFocus || isEdgeFocus || (!selectedCanvasNode && !selectedEdgePair);

                    const { nature } = parseTone(rel.tone);

                    // Clean & truncate nature string so text never spills out of frame
                    const rawNature = (nature || '').trim();
                    let cleanNature = rawNature;
                    if (cleanNature.length > 15 && cleanNature.includes('(')) {
                      const noParen = cleanNature.replace(/\s*\(.*?\)\s*/g, '').trim();
                      if (noParen.length >= 3 && noParen.length <= 15) {
                        cleanNature = noParen;
                      }
                    }
                    const displayNature = cleanNature.length > 17 ? `${cleanNature.slice(0, 16)}…` : cleanNature;

                    // Clean & truncate call_as string
                    const rawCall = rel.call_as ? `"${rel.call_as.trim()}"` : '';
                    const displayCall = rawCall.length > 17 ? `${rawCall.slice(0, 16)}…"` : rawCall;

                    // Accurately compute dynamic box width to guarantee zero overflow
                    const maxCharLen = Math.max(displayNature.length, displayCall.length);
                    // In 9px font, characters are ~6.5px average width. 7.4px + 22px gives generous breathing room
                    const boxWidth = Math.max(72, Math.min(160, Math.round(maxCharLen * 7.4 + 22)));
                    const boxHeight = displayNature && displayCall ? 32 : 22;

                    return (
                      <g
                        key={rel.id}
                        className="cursor-pointer group"
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedEdgePair({ from: rel.from_char, to: rel.to_char });
                        }}
                      >
                        {/* Curved Path */}
                        <path
                          d={pathD}
                          fill="none"
                          stroke={isEdgeFocus ? '#ec4899' : (isHighlighted ? (rel.is_locked ? '#818cf8' : '#6366f1') : '#1e293b')}
                          strokeWidth={isEdgeFocus ? 3.2 : (isHighlighted ? (selectedCanvasNode ? 2.5 : 1.8) : 1)}
                          strokeDasharray={rel.is_locked ? undefined : '5 3'}
                          opacity={isHighlighted ? 0.9 : 0.15}
                          markerEnd={isHighlighted ? (isEdgeFocus ? 'url(#arrowhead-highlight)' : 'url(#arrowhead-normal)') : undefined}
                          className="transition-all duration-200"
                        />

                        {/* Separate Curved Apex Badge */}
                        {isHighlighted && (
                          <g
                            transform={`translate(${apexX}, ${apexY})`}
                            filter="url(#badge-glow)"
                          >
                            <title>{`${rawNature ? `[${rawNature}] ` : ''}${rel.call_as ? `${t('graph.callsThem')} "${rel.call_as}"` : ''}${rel.self_call_as ? ` | ${t('graph.callsSelf')} "${rel.self_call_as}"` : ''}`}</title>
                            <rect
                              x={-boxWidth / 2}
                              y={-boxHeight / 2}
                              width={boxWidth}
                              height={boxHeight}
                              rx={7}
                              fill="#0d1222"
                              stroke={isEdgeFocus ? '#f43f5e' : (rel.is_locked ? '#818cf8' : '#4f46e5')}
                              strokeWidth={isEdgeFocus ? 2 : 1.2}
                              opacity={0.96}
                            />

                            {displayNature && displayCall ? (
                              <>
                                <text
                                  x={0}
                                  y={-3}
                                  fill="#c7d2fe"
                                  fontSize="9"
                                  fontWeight="bold"
                                  textAnchor="middle"
                                  className="pointer-events-none tracking-wide"
                                >
                                  {displayNature}
                                </text>
                                <text
                                  x={0}
                                  y={9.5}
                                  fill="#fde047"
                                  fontSize="9"
                                  fontWeight="600"
                                  textAnchor="middle"
                                  className="pointer-events-none font-mono"
                                >
                                  {displayCall}
                                </text>
                              </>
                            ) : (
                              <text
                                x={0}
                                y={3.5}
                                fill={displayNature ? '#c7d2fe' : '#fde047'}
                                fontSize="9"
                                fontWeight="600"
                                textAnchor="middle"
                                className="pointer-events-none font-mono"
                              >
                                {displayNature ? `[${displayNature}]` : (displayCall || t('graph.notSet'))}
                              </text>
                            )}

                            {rel.is_locked && (
                              <circle
                                cx={boxWidth / 2 - 5}
                                cy={-boxHeight / 2 + 5}
                                r={3}
                                fill="#818cf8"
                              />
                            )}
                          </g>
                        )}
                      </g>
                    );
                  })}

                  {/* Character Nodes with Drag-and-Drop */}
                  {displayedEntities.map((ent) => {
                    const pos = nodePositions[ent.name] || { x: 450, y: 325 };
                    const x = pos.x;
                    const y = pos.y;
                    const isBeingDragged = draggingNode?.name === ent.name;

                    const isSelected = selectedCanvasNode === ent.name;
                    const roleCfg = getRoleConfig(ent.role, t);
                    const factsCount = Array.isArray(ent.metadata?.facts) ? ent.metadata.facts.length : 0;
                    const relCount = displayedRelations.filter(r => r.from_char === ent.name || r.to_char === ent.name).length;

                    return (
                      <g
                        key={ent.id}
                        transform={`translate(${x}, ${y})`}
                        className={`cursor-grab select-none group ${isBeingDragged ? 'cursor-grabbing' : ''}`}
                        onMouseDown={(e) => {
                          e.preventDefault();
                          e.stopPropagation();
                          const pt = getSVGPoint(e.clientX, e.clientY);
                          const curPos = nodePositions[ent.name] || { x: 450, y: 325 };
                          setDraggingNode({
                            name: ent.name,
                            startX: pt.x,
                            startY: pt.y,
                            origX: curPos.x,
                            origY: curPos.y,
                            hasMoved: false,
                          });
                        }}
                      >
                        {/* Ripple Aura when Selected */}
                        {isSelected && (
                          <circle
                            r={36}
                            fill="none"
                            stroke={roleCfg.stroke}
                            strokeWidth={2}
                            className="animate-ping opacity-40"
                          />
                        )}

                        {/* Node Circle - Grows smoothly on drag/select without CSS transform bugs */}
                        <circle
                          r={isBeingDragged ? 28 : (isSelected ? 26 : 24)}
                          fill={roleCfg.fill}
                          stroke={isSelected || isBeingDragged ? '#ffffff' : roleCfg.stroke}
                          strokeWidth={isSelected || isBeingDragged ? 3 : 2}
                          className="group-hover:stroke-white group-hover:brightness-125 transition-all shadow-lg"
                        />

                        {/* Node Initial */}
                        <text
                          x={0}
                          y={6}
                          fill={roleCfg.textFill}
                          fontSize="15"
                          fontWeight="bold"
                          textAnchor="middle"
                          className="pointer-events-none select-none"
                        >
                          {ent.name.charAt(0)}
                        </text>

                        {/* Unified Character Label Capsule */}
                        <g transform="translate(0, 39)">
                          <title>{`${ent.name} (${roleCfg.label})`}</title>
                          <rect
                            x={-58}
                            y={-14}
                            width={116}
                            height={28}
                            rx={7}
                            fill="#0c101c"
                            stroke={isSelected ? roleCfg.stroke : '#253248'}
                            strokeWidth={isSelected ? 1.5 : 1}
                            opacity={0.95}
                          />
                          <text
                            x={0}
                            y={-1}
                            fill={isSelected ? '#ffffff' : '#f1f5f9'}
                            fontSize="10"
                            fontWeight="600"
                            textAnchor="middle"
                            className="pointer-events-none select-none drop-shadow"
                          >
                            {ent.name.length > 13 ? `${ent.name.slice(0, 12)}…` : ent.name}
                          </text>
                          <text
                            x={0}
                            y={9.5}
                            fill="#94a3b8"
                            fontSize="8"
                            textAnchor="middle"
                            className="pointer-events-none select-none font-medium capitalize"
                          >
                            {roleCfg.label} • {factsCount > 0 ? `${factsCount} facts` : `${relCount} rels`}
                          </text>
                        </g>
                      </g>
                    );
                  })}
                </svg>

                {/* Bottom Canvas Toolbar & Legend */}
                <div className="absolute bottom-2.5 left-3 right-3 flex items-center justify-between text-[11px] text-slate-400 bg-[#0c101d]/90 border border-white/10 rounded-xl px-3.5 py-2 backdrop-blur-md shadow-lg flex-wrap gap-2">
                  <div className="flex items-center space-x-3 text-xs">
                    <span className="font-semibold text-slate-200">{t('graph.legendFilter')}</span>
                    <span className="flex items-center space-x-1">
                      <span className="w-2.5 h-2.5 rounded-full bg-amber-500 inline-block" />
                      <span className="text-slate-300 text-[11px]">{t('graph.legendProtagonist')}</span>
                    </span>
                    <span className="flex items-center space-x-1">
                      <span className="w-2.5 h-2.5 rounded-full bg-rose-500 inline-block" />
                      <span className="text-slate-300 text-[11px]">{t('graph.legendHeroine')}</span>
                    </span>
                    <span className="flex items-center space-x-1">
                      <span className="w-2.5 h-2.5 rounded-full bg-indigo-500 inline-block" />
                      <span className="text-slate-300 text-[11px]">{t('graph.legendSupporting')}</span>
                    </span>
                    <span className="flex items-center space-x-1">
                      <span className="w-2.5 h-2.5 rounded-full bg-red-500 inline-block" />
                      <span className="text-slate-300 text-[11px]">{t('graph.legendAntagonist')}</span>
                    </span>
                  </div>

                  <div className="flex items-center space-x-2.5 text-xs flex-wrap">
                    <button
                      onClick={handleResetLayout}
                      className="flex items-center space-x-1.5 px-2.5 py-1 bg-white/5 hover:bg-white/10 text-slate-300 hover:text-white rounded-lg text-[11px] border border-white/10 transition shadow-xs"
                      title={t('graph.resetCircleTooltip')}
                    >
                      <RotateCcw className="w-3 h-3 text-indigo-400" />
                      <span>{t('graph.resetCircle')}</span>
                    </button>
                    <span className="text-slate-400">
                      {t('graph.dragTip')}
                    </span>
                    <span className="font-mono text-indigo-400 font-semibold">
                      {t('graph.summaryCount', { entities: displayedEntities.length, relations: displayedRelations.length })}
                    </span>
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Docked Character Dossier Panel (When Node is Selected) */}
          {selectedEntityObj && !selectedEdgePair && (
            <div className="w-96 max-w-full border-l border-white/8 bg-[#0c101c] flex flex-col h-full z-20 shadow-2xl overflow-hidden animate-in slide-in-from-right-8 duration-150 shrink-0">
              {/* Dossier Header */}
              <div className="p-4 border-b border-white/8 bg-[#0e1424] flex items-start justify-between shrink-0">
                <div className="flex items-center space-x-3 min-w-0">
                  <div className={`w-11 h-11 rounded-xl flex items-center justify-center font-bold text-lg border ${getRoleConfig(selectedEntityObj.role, t).badgeBorder} ${getRoleConfig(selectedEntityObj.role, t).badgeBg} text-white shrink-0`}>
                    {selectedEntityObj.name.charAt(0)}
                  </div>
                  <div className="min-w-0">
                    <h3 className="text-sm font-bold text-slate-100 truncate flex items-center gap-1.5">
                      <span className="truncate">{selectedEntityObj.name}</span>
                    </h3>
                    <div className="flex items-center space-x-1.5 mt-0.5 flex-wrap gap-1">
                      <span className={`text-[10px] px-1.5 py-0.5 rounded font-medium border ${getRoleConfig(selectedEntityObj.role, t).badgeBorder} ${getRoleConfig(selectedEntityObj.role, t).badgeBg} ${getRoleConfig(selectedEntityObj.role, t).badgeText}`}>
                        {getRoleConfig(selectedEntityObj.role, t).label}
                      </span>
                      <span className="text-[10px] px-1.5 py-0.5 rounded bg-white/4 text-slate-300 border border-white/10 capitalize">
                        {selectedEntityObj.gender === 'male' ? t('roles.male') : selectedEntityObj.gender === 'female' ? t('roles.female') : t('roles.other')}
                      </span>
                      <span className="text-[10px] px-1.5 py-0.5 rounded bg-white/4 text-slate-400 border border-white/10 font-mono">
                        Ch.{selectedEntityObj.first_seen_chapter}
                      </span>
                    </div>
                  </div>
                </div>

                <div className="flex items-center space-x-1">
                  <button
                    onClick={() => handleOpenEditEntity(selectedEntityObj)}
                    className="p-1.5 text-slate-400 hover:text-white rounded-lg hover:bg-white/6 transition"
                    title={t('graph.editCharTooltip')}
                  >
                    <Edit className="w-4 h-4" />
                  </button>
                  <button
                    onClick={() => handleDeleteEntity(selectedEntityObj)}
                    className="p-1.5 text-slate-400 hover:text-rose-400 rounded-lg hover:bg-rose-500/10 transition"
                    title={t('graph.deleteCharTooltip')}
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                  <button
                    onClick={() => setSelectedCanvasNode(null)}
                    className="p-1.5 text-slate-400 hover:text-white rounded-lg hover:bg-white/6 transition"
                  >
                    <X className="w-4 h-4" />
                  </button>
                </div>
              </div>

              {/* Dossier Content (Scrollable) */}
              <div className="flex-1 overflow-y-auto p-4 space-y-5 text-xs">
                {/* Aliases */}
                {selectedEntityObj.aliases && selectedEntityObj.aliases.length > 0 && (
                  <div>
                    <div className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider mb-1.5">
                      {t('graph.aliasesHeader')}
                    </div>
                    <div className="flex flex-wrap gap-1">
                      {selectedEntityObj.aliases.map((a, i) => (
                        <span key={i} className="px-2 py-0.5 rounded-md bg-white/4 border border-white/8 text-slate-300 text-[11px]">
                          {a}
                        </span>
                      ))}
                    </div>
                  </div>
                )}

                {/* Description */}
                <div>
                  <div className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider mb-1.5 flex items-center justify-between">
                    <span>{t('graph.bioHeader')}</span>
                  </div>
                  <div className="p-3 rounded-xl bg-[#0e1422] border border-white/8 text-slate-300 leading-relaxed text-[11.5px]">
                    {selectedEntityObj.metadata?.description ? (
                      selectedEntityObj.metadata.description
                    ) : (
                      <span className="text-slate-500 italic">{t('graph.noBio')}</span>
                    )}
                  </div>
                </div>

                {/* Canonical Facts */}
                <div>
                  <div className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider mb-1.5 flex items-center justify-between">
                    <span className="flex items-center space-x-1.5 text-amber-300">
                      <Sparkles className="w-3.5 h-3.5" />
                      <span>{t('graph.factsHeader')}</span>
                    </span>
                    <span className="text-[10px] text-slate-500 font-mono">
                      {Array.isArray(selectedEntityObj.metadata?.facts) ? selectedEntityObj.metadata.facts.length : 0} facts
                    </span>
                  </div>

                  <div className="space-y-1.5 mb-2">
                    {Array.isArray(selectedEntityObj.metadata?.facts) && selectedEntityObj.metadata.facts.length > 0 ? (
                      selectedEntityObj.metadata.facts.map((fact: string, idx: number) => (
                        <div
                          key={idx}
                          className="p-2.5 rounded-lg bg-[#0e1424] border border-white/6 text-slate-200 flex items-start justify-between group hover:border-indigo-500/30 transition"
                        >
                          <div className="flex items-start space-x-2">
                            <span className="text-indigo-400 font-bold mt-0.5">•</span>
                            <span className="leading-snug text-[11.5px]">{fact}</span>
                          </div>
                          <button
                            onClick={() => handleRemoveFact(idx)}
                            className="text-slate-500 hover:text-rose-400 p-1 rounded opacity-0 group-hover:opacity-100 transition shrink-0 ml-1"
                            title={t('graph.deleteFactTooltip')}
                          >
                            <X className="w-3 h-3" />
                          </button>
                        </div>
                      ))
                    ) : (
                      <div className="p-2.5 rounded-lg bg-white/2 border border-dashed border-white/10 text-slate-500 text-center italic">
                        {t('graph.noFacts')}
                      </div>
                    )}
                  </div>

                  {/* Inline Quick Add Fact */}
                  <div className="flex items-center space-x-1.5 mt-2">
                    <input
                      type="text"
                      placeholder={t('graph.addFactPlaceholder')}
                      className="flex-1 bg-[#111728] border border-white/10 rounded-lg px-2.5 py-1.5 text-slate-100 text-xs focus:outline-none focus:border-indigo-500"
                      value={quickFactInput}
                      onChange={e => setQuickFactInput(e.target.value)}
                      onKeyDown={e => {
                        if (e.key === 'Enter') {
                          e.preventDefault();
                          handleAddQuickFact();
                        }
                      }}
                    />
                    <button
                      onClick={handleAddQuickFact}
                      disabled={!quickFactInput.trim()}
                      className="px-2.5 py-1.5 bg-indigo-600 hover:bg-indigo-500 disabled:opacity-40 text-white rounded-lg font-medium text-xs transition shrink-0"
                    >
                      {t('graph.addFactBtn')}
                    </button>
                  </div>
                </div>

                {/* Interpersonal Pronouns Matrix */}
                <div>
                  <div className="text-[11px] font-semibold text-slate-400 uppercase tracking-wider mb-2 flex items-center justify-between">
                    <span className="flex items-center space-x-1.5 text-indigo-300">
                      <ArrowLeftRight className="w-3.5 h-3.5" />
                      <span>{t('graph.relationsHeader')}</span>
                    </span>
                    <button
                      onClick={() => handleOpenEditRelation(undefined, selectedEntityObj.name)}
                      className="text-indigo-400 hover:text-indigo-300 text-[11px] flex items-center space-x-1"
                    >
                      <Plus className="w-3 h-3" />
                      <span>{t('graph.addLink')}</span>
                    </button>
                  </div>

                  {/* Outgoing (This character calls others) */}
                  <div className="space-y-2 mb-3">
                    <div className="text-[10px] text-slate-500 font-semibold uppercase">
                      {t('graph.outgoingTitle', { name: selectedEntityObj.name, count: selectedNodeOutgoingRels.length })}
                    </div>
                    {selectedNodeOutgoingRels.length > 0 ? (
                      selectedNodeOutgoingRels.map(rel => {
                        const { nature, tone } = parseTone(rel.tone);
                        return (
                          <div
                            key={rel.id}
                            className="p-2.5 rounded-xl bg-[#0e1422] border border-white/8 hover:border-indigo-500/30 transition space-y-1.5"
                          >
                            <div className="flex items-center justify-between">
                              <div className="flex items-center space-x-1.5 font-semibold text-slate-200">
                                <span className="text-slate-400">&rarr;</span>
                                <span className="text-indigo-300 text-[12px]">{rel.to_char}</span>
                                {nature && (
                                  <span className="px-1.5 py-0.2 rounded bg-indigo-500/15 border border-indigo-500/30 text-indigo-300 text-[10px]">
                                    {nature}
                                  </span>
                                )}
                              </div>
                              <div className="flex items-center space-x-1">
                                <button
                                  onClick={() => toggleLockRelation(rel.id)}
                                  className={`p-1 rounded transition ${rel.is_locked ? 'text-indigo-400' : 'text-slate-500 hover:text-slate-300'}`}
                                  title={rel.is_locked ? t('graph.lockTooltipLocked') : t('graph.lockTooltipUnlocked')}
                                >
                                  {rel.is_locked ? <Lock className="w-3.5 h-3.5" /> : <Unlock className="w-3.5 h-3.5" />}
                                </button>
                                <button
                                  onClick={() => handleOpenEditRelation(rel)}
                                  className="p-1 text-slate-400 hover:text-white rounded transition"
                                  title={t('common.edit')}
                                >
                                  <Edit className="w-3.5 h-3.5" />
                                </button>
                                <button
                                  onClick={() => handleDeleteRelation(rel.id)}
                                  className="p-1 text-slate-500 hover:text-rose-400 rounded transition"
                                  title={t('graph.deleteRelTitle')}
                                >
                                  <Trash2 className="w-3.5 h-3.5" />
                                </button>
                              </div>
                            </div>

                            <div className="grid grid-cols-2 gap-1.5 text-[11px] bg-[#0a0d16] p-2 rounded-lg border border-white/4">
                              <div>
                                <span className="text-slate-500 block text-[10px]">{t('graph.callsThem')}</span>
                                <span className="font-semibold text-amber-300">{rel.call_as || t('graph.notSet')}</span>
                              </div>
                              <div>
                                <span className="text-slate-500 block text-[10px]">{t('graph.callsSelf')}</span>
                                <span className="font-semibold text-emerald-300">{rel.self_call_as || t('graph.notSet')}</span>
                              </div>
                            </div>

                            <div className="flex items-center justify-between text-[10.5px] text-slate-500 pt-0.5">
                              <span>{t('graph.toneLabel')} <span className="text-slate-300">{tone || t('graph.toneNatural')}</span></span>
                              <span>{t('graph.fromChapterLabel', { ch: rel.since_chapter })}</span>
                            </div>
                          </div>
                        );
                      })
                    ) : (
                      <div className="p-2 rounded-lg bg-white/2 border border-white/6 text-slate-500 text-center italic text-[11px]">
                        {t('graph.noOutgoing')}
                      </div>
                    )}
                  </div>

                  {/* Incoming (Others call this character) */}
                  <div className="space-y-2">
                    <div className="text-[10px] text-slate-500 font-semibold uppercase">
                      {t('graph.incomingTitle', { name: selectedEntityObj.name, count: selectedNodeIncomingRels.length })}
                    </div>
                    {selectedNodeIncomingRels.length > 0 ? (
                      selectedNodeIncomingRels.map(rel => {
                        const { nature, tone } = parseTone(rel.tone);
                        return (
                          <div
                            key={rel.id}
                            className="p-2.5 rounded-xl bg-[#0e1422] border border-white/8 hover:border-indigo-500/30 transition space-y-1.5"
                          >
                            <div className="flex items-center justify-between">
                              <div className="flex items-center space-x-1.5 font-semibold text-slate-200">
                                <span className="text-amber-300 text-[12px]">{rel.from_char}</span>
                                <span className="text-slate-400">&rarr;</span>
                                {nature && (
                                  <span className="px-1.5 py-0.2 rounded bg-indigo-500/15 border border-indigo-500/30 text-indigo-300 text-[10px]">
                                    {nature}
                                  </span>
                                )}
                              </div>
                              <div className="flex items-center space-x-1">
                                <button
                                  onClick={() => toggleLockRelation(rel.id)}
                                  className={`p-1 rounded transition ${rel.is_locked ? 'text-indigo-400' : 'text-slate-500 hover:text-slate-300'}`}
                                  title={rel.is_locked ? t('graph.lockTooltipLocked') : t('graph.lockTooltipUnlocked')}
                                >
                                  {rel.is_locked ? <Lock className="w-3.5 h-3.5" /> : <Unlock className="w-3.5 h-3.5" />}
                                </button>
                                <button
                                  onClick={() => handleOpenEditRelation(rel)}
                                  className="p-1 text-slate-400 hover:text-white rounded transition"
                                  title={t('common.edit')}
                                >
                                  <Edit className="w-3.5 h-3.5" />
                                </button>
                                <button
                                  onClick={() => handleDeleteRelation(rel.id)}
                                  className="p-1 text-slate-500 hover:text-rose-400 rounded transition"
                                  title={t('graph.deleteRelTitle')}
                                >
                                  <Trash2 className="w-3.5 h-3.5" />
                                </button>
                              </div>
                            </div>

                            <div className="grid grid-cols-2 gap-1.5 text-[11px] bg-[#0a0d16] p-2 rounded-lg border border-white/4">
                              <div>
                                <span className="text-slate-500 block text-[10px]">{t('graph.callsThem')}</span>
                                <span className="font-semibold text-amber-300">{rel.call_as || t('graph.notSet')}</span>
                              </div>
                              <div>
                                <span className="text-slate-500 block text-[10px]">{t('graph.callsSelf')}</span>
                                <span className="font-semibold text-emerald-300">{rel.self_call_as || t('graph.notSet')}</span>
                              </div>
                            </div>

                            <div className="flex items-center justify-between text-[10.5px] text-slate-500 pt-0.5">
                              <span>{t('graph.toneLabel')} <span className="text-slate-300">{tone || t('graph.toneNatural')}</span></span>
                              <span>{t('graph.fromChapterLabel', { ch: rel.since_chapter })}</span>
                            </div>
                          </div>
                        );
                      })
                    ) : (
                      <div className="p-2 rounded-lg bg-white/2 border border-white/6 text-slate-500 text-center italic text-[11px]">
                        {t('graph.noIncoming')}
                      </div>
                    )}
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Docked Edge Inspector Panel (When Edge is Clicked) */}
          {selectedPairRels && (
            <div className="w-96 max-w-full border-l border-white/8 bg-[#0c101c] flex flex-col h-full z-20 shadow-2xl overflow-hidden animate-in slide-in-from-right-8 duration-150 shrink-0">
              <div className="p-4 border-b border-white/8 bg-[#0e1424] flex items-center justify-between shrink-0">
                <div className="min-w-0">
                  <div className="text-[10px] text-indigo-400 font-semibold uppercase tracking-wider flex items-center space-x-1">
                    <ArrowLeftRight className="w-3.5 h-3.5" />
                    <span>{t('graph.pairDrawerTitle')}</span>
                  </div>
                  <h3 className="text-sm font-bold text-white truncate mt-0.5">
                    {selectedEdgePair?.from} ⮂ {selectedEdgePair?.to}
                  </h3>
                </div>
                <button
                  onClick={() => setSelectedEdgePair(null)}
                  className="p-1.5 text-slate-400 hover:text-white rounded-lg hover:bg-white/6 transition"
                >
                  <X className="w-4 h-4" />
                </button>
              </div>

              <div className="flex-1 overflow-y-auto p-4 space-y-4 text-xs">
                {/* Direction 1: From ➔ To */}
                <div className="p-3.5 rounded-xl bg-[#0e1422] border border-white/8 space-y-2">
                  <div className="flex items-center justify-between border-b border-white/6 pb-2">
                    <div className="flex items-center space-x-2 font-semibold text-slate-100">
                      <span className="text-amber-300 font-bold">{selectedEdgePair?.from}</span>
                      <span className="text-slate-500">&rarr;</span>
                      <span className="text-indigo-300 font-bold">{selectedEdgePair?.to}</span>
                    </div>
                    {selectedPairRels.forward && (
                      <div className="flex items-center space-x-1">
                        <button
                          onClick={() => toggleLockRelation(selectedPairRels.forward!.id)}
                          className={`p-1 rounded transition ${selectedPairRels.forward.is_locked ? 'text-indigo-400' : 'text-slate-500 hover:text-slate-300'}`}
                          title={selectedPairRels.forward.is_locked ? t('graph.lockTooltipLocked') : t('graph.lockTooltipUnlocked')}
                        >
                          {selectedPairRels.forward.is_locked ? <Lock className="w-3.5 h-3.5" /> : <Unlock className="w-3.5 h-3.5" />}
                        </button>
                        <button
                          onClick={() => handleOpenEditRelation(selectedPairRels.forward!)}
                          className="p-1 text-slate-400 hover:text-white rounded transition"
                          title={t('common.edit')}
                        >
                          <Edit className="w-3.5 h-3.5" />
                        </button>
                        <button
                          onClick={() => handleDeleteRelation(selectedPairRels.forward!.id)}
                          className="p-1 text-slate-500 hover:text-rose-400 rounded transition"
                          title={t('graph.deleteRelTitle')}
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </button>
                      </div>
                    )}
                  </div>

                  {selectedPairRels.forward ? (
                    <>
                      {(() => {
                        const { nature, tone } = parseTone(selectedPairRels.forward.tone);
                        return (
                          <>
                            {nature && (
                              <div className="inline-block px-2 py-0.5 rounded bg-indigo-500/15 border border-indigo-500/30 text-indigo-300 text-[11px] font-semibold">
                                {t('graph.pairNature', { nature })}
                              </div>
                            )}
                            <div className="bg-[#090d16] p-2.5 rounded-lg border border-white/6 space-y-1.5">
                              <div className="flex justify-between">
                                <span className="text-slate-400">{t('graph.callsThem')}</span>
                                <strong className="text-amber-300 font-medium">{selectedPairRels.forward.call_as || t('graph.notSet')}</strong>
                              </div>
                              <div className="flex justify-between">
                                <span className="text-slate-400">{t('graph.callsSelf')}</span>
                                <strong className="text-emerald-300 font-medium">{selectedPairRels.forward.self_call_as || t('graph.notSet')}</strong>
                              </div>
                              <div className="flex justify-between text-[11px] text-slate-500 pt-1 border-t border-white/4">
                                <span>{t('graph.toneLabel')} <span className="text-slate-300">{tone || t('graph.toneNatural')}</span></span>
                                <span>{t('projects.chapters')}: <span className="text-slate-300 font-mono">{selectedPairRels.forward.since_chapter}</span></span>
                              </div>
                            </div>
                          </>
                        );
                      })()}
                    </>
                  ) : (
                    <div className="text-center py-3">
                      <p className="text-slate-400 mb-2">{t('graph.noForwardRel')}</p>
                      <button
                        onClick={() => handleOpenEditRelation(undefined, selectedEdgePair?.from, selectedEdgePair?.to)}
                        className="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold transition"
                      >
                        {t('graph.createForwardRel', { from: selectedEdgePair?.from, to: selectedEdgePair?.to })}
                      </button>
                    </div>
                  )}
                </div>

                {/* Direction 2: To ➔ From */}
                <div className="p-3.5 rounded-xl bg-[#0e1422] border border-white/8 space-y-2">
                  <div className="flex items-center justify-between border-b border-white/6 pb-2">
                    <div className="flex items-center space-x-2 font-semibold text-slate-100">
                      <span className="text-indigo-300 font-bold">{selectedEdgePair?.to}</span>
                      <span className="text-slate-500">&rarr;</span>
                      <span className="text-amber-300 font-bold">{selectedEdgePair?.from}</span>
                    </div>
                    {selectedPairRels.reverse && (
                      <div className="flex items-center space-x-1">
                        <button
                          onClick={() => toggleLockRelation(selectedPairRels.reverse!.id)}
                          className={`p-1 rounded transition ${selectedPairRels.reverse.is_locked ? 'text-indigo-400' : 'text-slate-500 hover:text-slate-300'}`}
                          title={selectedPairRels.reverse.is_locked ? t('graph.lockTooltipLocked') : t('graph.lockTooltipUnlocked')}
                        >
                          {selectedPairRels.reverse.is_locked ? <Lock className="w-3.5 h-3.5" /> : <Unlock className="w-3.5 h-3.5" />}
                        </button>
                        <button
                          onClick={() => handleOpenEditRelation(selectedPairRels.reverse!)}
                          className="p-1 text-slate-400 hover:text-white rounded transition"
                          title={t('common.edit')}
                        >
                          <Edit className="w-3.5 h-3.5" />
                        </button>
                        <button
                          onClick={() => handleDeleteRelation(selectedPairRels.reverse!.id)}
                          className="p-1 text-slate-500 hover:text-rose-400 rounded transition"
                          title={t('graph.deleteRelTitle')}
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </button>
                      </div>
                    )}
                  </div>

                  {selectedPairRels.reverse ? (
                    <>
                      {(() => {
                        const { nature, tone } = parseTone(selectedPairRels.reverse.tone);
                        return (
                          <>
                            {nature && (
                              <div className="inline-block px-2 py-0.5 rounded bg-indigo-500/15 border border-indigo-500/30 text-indigo-300 text-[11px] font-semibold">
                                {t('graph.pairNature', { nature })}
                              </div>
                            )}
                            <div className="bg-[#090d16] p-2.5 rounded-lg border border-white/6 space-y-1.5">
                              <div className="flex justify-between">
                                <span className="text-slate-400">{t('graph.callsThem')}</span>
                                <strong className="text-amber-300 font-medium">{selectedPairRels.reverse.call_as || t('graph.notSet')}</strong>
                              </div>
                              <div className="flex justify-between">
                                <span className="text-slate-400">{t('graph.callsSelf')}</span>
                                <strong className="text-emerald-300 font-medium">{selectedPairRels.reverse.self_call_as || t('graph.notSet')}</strong>
                              </div>
                              <div className="flex justify-between text-[11px] text-slate-500 pt-1 border-t border-white/4">
                                <span>{t('graph.toneLabel')} <span className="text-slate-300">{tone || t('graph.toneNatural')}</span></span>
                                <span>{t('projects.chapters')}: <span className="text-slate-300 font-mono">{selectedPairRels.reverse.since_chapter}</span></span>
                              </div>
                            </div>
                          </>
                        );
                      })()}
                    </>
                  ) : (
                    <div className="text-center py-3">
                      <p className="text-slate-400 mb-2">{t('graph.noReverseRel')}</p>
                      <button
                        onClick={() => handleOpenEditRelation(undefined, selectedEdgePair?.to, selectedEdgePair?.from)}
                        className="px-3 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg text-xs font-semibold transition"
                      >
                        {t('graph.createReverseRel', { from: selectedEdgePair?.to, to: selectedEdgePair?.from })}
                      </button>
                    </div>
                  )}
                </div>
              </div>
            </div>
          )}
        </div>
      ) : (
        /* Cards View: Full Rich Grid */
        <div className="flex-1 p-4 sm:p-6 overflow-y-auto space-y-6">
          {/* Character Nodes Section */}
          <div>
            <div className="text-xs font-semibold uppercase tracking-wider text-slate-400 mb-3 flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <UserCheck className="w-4 h-4 text-indigo-400" />
                <span>{t('graph.charListTitle', { count: displayedEntities.length })}</span>
              </div>
              <button
                onClick={() => handleOpenEditEntity()}
                className="text-indigo-400 hover:text-indigo-300 text-xs font-medium flex items-center space-x-1"
              >
                <Plus className="w-3.5 h-3.5" />
                <span>{t('graph.addNewChar')}</span>
              </button>
            </div>

            <div className="grid grid-cols-1 sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-3.5">
              {displayedEntities.map(ent => {
                const roleCfg = getRoleConfig(ent.role, t);
                const facts = Array.isArray(ent.metadata?.facts) ? ent.metadata.facts : [];

                return (
                  <div
                    key={ent.id}
                    className="bg-[#0e1320] border border-white/8 hover:border-indigo-500/30 p-3.5 rounded-xl shadow-xs flex flex-col justify-between transition group"
                  >
                    <div>
                      <div className="flex items-start justify-between">
                        <div className="flex items-center space-x-2.5 min-w-0">
                          <div className={`w-9 h-9 rounded-xl ${roleCfg.badgeBg} border ${roleCfg.badgeBorder} flex items-center justify-center text-sm font-bold ${roleCfg.badgeText} shrink-0`}>
                            {ent.name.charAt(0)}
                          </div>
                          <div className="min-w-0">
                            <div className="text-xs font-bold text-slate-100 truncate">{ent.name}</div>
                            <div className="text-[10px] text-slate-400 capitalize flex items-center space-x-1">
                              <span>{roleCfg.label}</span>
                              <span>•</span>
                              <span>{ent.gender === 'male' ? t('roles.male') : ent.gender === 'female' ? t('roles.female') : t('roles.other')}</span>
                            </div>
                          </div>
                        </div>

                        <div className="flex items-center space-x-1">
                          <button
                            onClick={() => handleOpenEditEntity(ent)}
                            className="p-1 text-slate-400 hover:text-white rounded hover:bg-white/6 transition"
                            title={t('common.edit')}
                          >
                            <Edit className="w-3 h-3" />
                          </button>
                          <button
                            onClick={() => handleDeleteEntity(ent)}
                            className="p-1 text-slate-400 hover:text-rose-400 rounded hover:bg-rose-500/10 transition"
                            title={t('common.delete')}
                          >
                            <Trash2 className="w-3 h-3" />
                          </button>
                        </div>
                      </div>

                      {/* Aliases */}
                      {ent.aliases && ent.aliases.length > 0 && (
                        <div className="mt-2 text-[11px] text-slate-400 flex flex-wrap gap-1">
                          {ent.aliases.map((a, i) => (
                            <span key={i} className="px-1.5 py-0.2 rounded bg-white/3 border border-white/6 text-[10px] text-slate-300">
                              {a}
                            </span>
                          ))}
                        </div>
                      )}

                      {/* Description Preview */}
                      {ent.metadata?.description && (
                        <p className="mt-2 text-[11px] text-slate-300 line-clamp-2 leading-relaxed bg-[#0a0d16] p-1.5 rounded-lg border border-white/4">
                          {ent.metadata.description}
                        </p>
                      )}

                      {/* Facts Preview */}
                      {facts.length > 0 && (
                        <div className="mt-2 space-y-1">
                          {facts.slice(0, 2).map((fact: string, fi: number) => (
                            <div key={fi} className="text-[10.5px] text-slate-300 flex items-start space-x-1.5 truncate">
                              <span className="text-amber-400 font-bold">•</span>
                              <span className="truncate">{fact}</span>
                            </div>
                          ))}
                          {facts.length > 2 && (
                            <div className="text-[10px] text-slate-500 italic">
                              {t('graph.moreFacts', { count: facts.length - 2 })}
                            </div>
                          )}
                        </div>
                      )}
                    </div>

                    {/* Footer */}
                    <div className="mt-3 pt-2 border-t border-white/6 flex items-center justify-between text-[10.5px] text-slate-500">
                      <span>{t('graph.firstSeenAt', { ch: ent.first_seen_chapter })}</span>
                      <button
                        onClick={() => {
                          setSelectedCanvasNode(ent.name);
                          setViewMode('canvas');
                        }}
                        className="text-indigo-400 hover:text-indigo-300 font-semibold flex items-center space-x-1"
                      >
                        <span>{t('graph.viewInGraph')}</span>
                        <ChevronRight className="w-3 h-3" />
                      </button>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>

          {/* Relations Section */}
          <div>
            <div className="text-xs font-semibold uppercase tracking-wider text-slate-400 mb-3 flex items-center justify-between">
              <div className="flex items-center space-x-2">
                <Link2 className="w-4 h-4 text-indigo-400" />
                <span>{t('graph.matrixTitle', { count: displayedRelations.length })}</span>
              </div>
              <button
                onClick={() => handleOpenEditRelation()}
                className="text-indigo-400 hover:text-indigo-300 text-xs font-medium flex items-center space-x-1"
              >
                <Plus className="w-3.5 h-3.5" />
                <span>{t('graph.addNewRel')}</span>
              </button>
            </div>

            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3.5">
              {displayedRelations.map(rel => {
                const { nature, tone } = parseTone(rel.tone);
                return (
                  <div
                    key={rel.id}
                    className={`p-3.5 rounded-xl border transition ${
                      rel.is_locked
                        ? 'bg-indigo-500/6 border-indigo-500/30 shadow-xs'
                        : 'bg-[#0e1320] border-white/8 hover:border-white/20'
                    }`}
                  >
                    {/* Connection Header */}
                    <div className="flex items-center justify-between mb-2">
                      <div className="flex items-center space-x-1.5 text-xs font-semibold text-slate-200">
                        <span className="text-amber-300 font-bold">{rel.from_char}</span>
                        <span className="text-slate-500">&rarr;</span>
                        <span className="text-indigo-300 font-bold">{rel.to_char}</span>
                      </div>

                      <div className="flex items-center space-x-1.5">
                        <button
                          onClick={() => toggleLockRelation(rel.id)}
                          className={`px-2 py-0.5 rounded text-[10px] font-medium flex items-center space-x-1 border transition ${
                            rel.is_locked
                              ? 'bg-indigo-500/15 text-indigo-300 border-indigo-500/30'
                              : 'bg-white/4 text-slate-400 border-white/10'
                          }`}
                          title={rel.is_locked ? t('graph.lockTooltipLocked') : t('graph.lockTooltipUnlocked')}
                        >
                          {rel.is_locked ? <Lock className="w-3 h-3" /> : <Unlock className="w-3 h-3" />}
                          <span>{rel.is_locked ? t('graph.lockedStatus') : t('graph.unlockedStatus')}</span>
                        </button>

                        <button
                          onClick={() => handleOpenEditRelation(rel)}
                          className="p-1 hover:bg-white/6 text-slate-400 hover:text-slate-200 rounded-md transition"
                          title={t('common.edit')}
                        >
                          <Edit className="w-3.5 h-3.5" />
                        </button>

                        <button
                          onClick={() => handleDeleteRelation(rel.id)}
                          className="p-1 hover:bg-rose-500/10 text-slate-400 hover:text-rose-400 rounded-md transition"
                          title={t('graph.deleteRelTitle')}
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </button>
                      </div>
                    </div>

                    {/* Nature Tag */}
                    {nature && (
                      <div className="mb-2">
                        <span className="px-2 py-0.5 rounded bg-indigo-500/15 border border-indigo-500/30 text-indigo-300 text-[10.5px] font-semibold">
                          {nature}
                        </span>
                      </div>
                    )}

                    {/* Pronouns Pair Details */}
                    <div className="bg-[#090d16] rounded-lg p-2.5 text-xs space-y-1.5 border border-white/6">
                      <div className="flex justify-between">
                        <span className="text-slate-400">{t('graph.callsThem')}</span>
                        <strong className="text-amber-300 font-semibold">{rel.call_as || t('graph.notSet')}</strong>
                      </div>
                      <div className="flex justify-between">
                        <span className="text-slate-400">{t('graph.callsSelf')}</span>
                        <strong className="text-emerald-300 font-semibold">{rel.self_call_as || t('graph.notSet')}</strong>
                      </div>
                      <div className="flex justify-between text-[11px] text-slate-500 pt-1.5 border-t border-white/6">
                        <span>{t('graph.toneLabel')} <span className="text-slate-300">{tone || t('graph.toneNatural')}</span></span>
                        <span>{t('graph.sinceChapterLabel')} <span className="text-slate-300 font-mono">{rel.since_chapter}</span></span>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>
      )}

      {/* Edit / Add Relation Modal */}
      {isEditingRelation && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
          <div className="bg-[#0c101a] border border-white/10rounded-2xl max-w-md w-full p-5 shadow-2xl animate-in fade-in zoom-in-95 duration-200">
            <div className="flex items-center justify-between pb-3 border-b border-white/8">
              <h3 className="text-sm font-semibold text-slate-100 flex items-center space-x-2">
                <Link2 className="w-4 h-4 text-indigo-400" />
                <span>{selectedRelation ? t('graph.editRelTitle') : t('graph.addRelTitle')}</span>
              </h3>
              <button onClick={() => setIsEditingRelation(false)} className="text-slate-400 hover:text-slate-200 p-1 rounded-lg hover:bg-white/6 transition">
                <X className="w-4 h-4" />
              </button>
            </div>

            <form onSubmit={handleSaveRelation} className="mt-4 space-y-3 text-xs">
              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-slate-400 mb-1 font-medium">{t('graph.speakerLabel')}</label>
                  <input
                    type="text"
                    className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-2.5 py-1.5 text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                    value={editFromChar}
                    onChange={e => setEditFromChar(e.target.value)}
                    required
                  />
                </div>
                <div>
                  <label className="block text-slate-400 mb-1 font-medium">{t('graph.listenerLabel')}</label>
                  <input
                    type="text"
                    className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-2.5 py-1.5 text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                    value={editToChar}
                    onChange={e => setEditToChar(e.target.value)}
                    required
                  />
                </div>
              </div>

              <div>
                <label className="block text-slate-400 mb-1 font-medium">{t('graph.relNatureLabel')}</label>
                <input
                  type="text"
                  placeholder={t('graph.relNaturePlaceholder')}
                  className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-2.5 py-1.5 text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                  value={editRelNature}
                  onChange={e => setEditRelNature(e.target.value)}
                />
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-slate-400 mb-1 font-medium">{t('graph.callAsLabel')}</label>
                  <input
                    type="text"
                    placeholder={t('graph.callAsPlaceholder')}
                    className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-2.5 py-1.5 text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                    value={editCallAs}
                    onChange={e => setEditCallAs(e.target.value)}
                    required
                  />
                </div>
                <div>
                  <label className="block text-slate-400 mb-1 font-medium">{t('graph.selfCallAsLabel')}</label>
                  <input
                    type="text"
                    placeholder={t('graph.selfCallAsPlaceholder')}
                    className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-2.5 py-1.5 text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                    value={editSelfCallAs}
                    onChange={e => setEditSelfCallAs(e.target.value)}
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div>
                  <label className="block text-slate-400 mb-1 font-medium">{t('graph.sinceChapterLabel')}</label>
                  <input
                    type="number"
                    min="1"
                    className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-2.5 py-1.5 text-slate-100 font-mono focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                    value={editSinceChapter}
                    onChange={e => setEditSinceChapter(parseInt(e.target.value) || 1)}
                  />
                </div>
                <div>
                  <label className="block text-slate-400 mb-1 font-medium">{t('graph.relToneLabel')}</label>
                  <input
                    type="text"
                    placeholder={t('graph.relTonePlaceholder')}
                    className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-2.5 py-1.5 text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                    value={editTone}
                    onChange={e => setEditTone(e.target.value)}
                  />
                </div>
              </div>

              {/* Option to create reciprocal relationship */}
              {!selectedRelation && (
                <div className="p-2.5 rounded-lg bg-indigo-500/5 border border-indigo-500/20 space-y-2">
                  <div className="flex items-center space-x-2">
                    <input
                      type="checkbox"
                      id="reciprocalCheckbox"
                      checked={createReciprocal}
                      onChange={e => setCreateReciprocal(e.target.checked)}
                      className="w-3.5 h-3.5 rounded border-white/20 text-indigo-600 focus:ring-indigo-500"
                    />
                    <label htmlFor="reciprocalCheckbox" className="text-slate-200 font-medium cursor-pointer">
                      {t('graph.createReverseCheckbox', { to: editToChar || 'B', from: editFromChar || 'A' })}
                    </label>
                  </div>

                  {createReciprocal && (
                    <div className="grid grid-cols-2 gap-2 pt-1 border-t border-indigo-500/10">
                      <div>
                        <label className="block text-[10.5px] text-slate-400 mb-0.5">{t('graph.reverseCallAsLabel')}</label>
                        <input
                          type="text"
                          placeholder={t('graph.reverseCallAsPlaceholder')}
                          className="w-full bg-[#0d1222] border border-white/10 rounded-lg px-2 py-1 text-slate-100 text-xs"
                          value={reciprocalCallAs}
                          onChange={e => setReciprocalCallAs(e.target.value)}
                        />
                      </div>
                      <div>
                        <label className="block text-[10.5px] text-slate-400 mb-0.5">{t('graph.reverseSelfCallAsLabel')}</label>
                        <input
                          type="text"
                          placeholder={t('graph.reverseSelfCallAsPlaceholder')}
                          className="w-full bg-[#0d1222] border border-white/10 rounded-lg px-2 py-1 text-slate-100 text-xs"
                          value={reciprocalSelfCallAs}
                          onChange={e => setReciprocalSelfCallAs(e.target.value)}
                        />
                      </div>
                    </div>
                  )}
                </div>
              )}

              <div className="flex items-center space-x-2 pt-1">
                <input
                  type="checkbox"
                  id="lockCheckbox"
                  checked={editIsLocked}
                  onChange={e => setEditIsLocked(e.target.checked)}
                  className="w-3.5 h-3.5 rounded border-white/20 text-indigo-600 focus:ring-indigo-500"
                />
                <label htmlFor="lockCheckbox" className="text-slate-300 font-medium cursor-pointer">
                  {t('graph.hardLockCheckbox')}
                </label>
              </div>

              <div className="flex items-center justify-end space-x-2 pt-3 border-t border-white/8">
                <button
                  type="button"
                  onClick={() => setIsEditingRelation(false)}
                  className="px-3.5 py-1.5 bg-white/4 hover:bg-white/8 text-slate-300 rounded-lg border border-white/10 transition"
                >
                  {t('common.cancel')}
                </button>
                <button
                  type="submit"
                  className="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg font-semibold shadow-sm border border-indigo-400/30 transition"
                >
                  {t('graph.saveRelBtn')}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Edit / Add Character Modal */}
      {isEditingEntity && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
          <div className="bg-[#0c101a] border border-white/10rounded-2xl max-w-md w-full p-5 shadow-2xl animate-in fade-in zoom-in-95 duration-200">
            <div className="flex items-center justify-between pb-3 border-b border-white/8">
              <h3 className="text-sm font-semibold text-slate-100 flex items-center space-x-2">
                <Users className="w-4 h-4 text-indigo-400" />
                <span>{editingEntityId ? t('graph.editCharModalTitle') : t('graph.addCharModalTitle')}</span>
              </h3>
              <button onClick={() => setIsEditingEntity(false)} className="text-slate-400 hover:text-slate-200 p-1 rounded-lg hover:bg-white/6 transition">
                <X className="w-4 h-4" />
              </button>
            </div>

            <form onSubmit={handleSaveEntity} className="mt-4 space-y-3 text-xs">
              <div>
                <label className="block text-slate-400 mb-1 font-medium">{t('graph.nameLabel')}</label>
                <input
                  type="text"
                  className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-2.5 py-1.5 text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                  value={entityFormName}
                  onChange={e => setEntityFormName(e.target.value)}
                  placeholder={t('graph.namePlaceholder')}
                  required
                />
              </div>

              <div>
                <label className="block text-slate-400 mb-1 font-medium">{t('graph.aliasesLabel')}</label>
                <input
                  type="text"
                  className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-2.5 py-1.5 text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                  value={entityFormAliases}
                  onChange={e => setEntityFormAliases(e.target.value)}
                  placeholder={t('graph.aliasesPlaceholder')}
                />
              </div>

              <div className="grid grid-cols-3 gap-2.5">
                <div>
                  <label className="block text-slate-400 mb-1 font-medium">{t('graph.genderLabel')}</label>
                  <Select
                    value={entityFormGender}
                    onChange={e => setEntityFormGender(e.target.value)}
                    className="w-full py-1.5 text-xs"
                    containerClassName="w-full"
                  >
                    <option value="male">{t('roles.male')}</option>
                    <option value="female">{t('roles.female')}</option>
                    <option value="other">{t('roles.other')}</option>
                  </Select>
                </div>

                <div>
                  <label className="block text-slate-400 mb-1 font-medium">{t('graph.roleLabel')}</label>
                  <Select
                    value={entityFormRole}
                    onChange={e => setEntityFormRole(e.target.value)}
                    className="w-full py-1.5 text-xs"
                    containerClassName="w-full"
                  >
                    <option value="protagonist">{t('roles.protagonist')}</option>
                    <option value="heroine">{t('roles.heroine')}</option>
                    <option value="supporting">{t('roles.supporting')}</option>
                    <option value="antagonist">{t('roles.antagonist')}</option>
                  </Select>
                </div>

                <div>
                  <label className="block text-slate-400 mb-1 font-medium">{t('graph.firstSeenLabel')}</label>
                  <input
                    type="number"
                    min="1"
                    className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-2.5 py-1.5 text-slate-100 font-mono text-xs focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition"
                    value={entityFormFirstSeen}
                    onChange={e => setEntityFormFirstSeen(parseInt(e.target.value) || 1)}
                  />
                </div>
              </div>

              <div>
                <label className="block text-slate-400 mb-1 font-medium">{t('graph.bioLabel')}</label>
                <textarea
                  rows={2}
                  className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-2.5 py-1.5 text-slate-100 focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition resize-none"
                  value={entityFormDescription}
                  onChange={e => setEntityFormDescription(e.target.value)}
                  placeholder={t('graph.bioPlaceholder')}
                />
              </div>

              <div>
                <label className="block text-slate-400 mb-1 font-medium">{t('graph.factsLabel')}</label>
                <textarea
                  rows={3}
                  className="w-full bg-[#111726] border border-white/10 hover:border-white/20 focus:border-indigo-500 rounded-lg px-2.5 py-1.5 text-slate-100 font-mono focus:outline-none focus:ring-1 focus:ring-indigo-500/30 transition resize-none"
                  value={entityFormFacts}
                  onChange={e => setEntityFormFacts(e.target.value)}
                  placeholder={t('graph.factsPlaceholder')}
                />
              </div>

              <div className="flex items-center justify-end space-x-2 pt-3 border-t border-white/8">
                <button
                  type="button"
                  onClick={() => setIsEditingEntity(false)}
                  className="px-3.5 py-1.5 bg-white/4 hover:bg-white/8 text-slate-300 rounded-lg border border-white/10 transition"
                >
                  {t('common.cancel')}
                </button>
                <button
                  type="submit"
                  className="px-4 py-1.5 bg-indigo-600 hover:bg-indigo-500 text-white rounded-lg font-semibold shadow-sm border border-indigo-400/30 transition"
                >
                  {editingEntityId ? t('graph.updateProfileBtn') : t('graph.addCharSubmitBtn')}
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* AI Auto-Scan Modal Dialog */}
      {isAutoScanOpen && (
        <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
          <div className="bg-[#0e1320] border border-white/10rounded-2xl max-w-lg w-full p-5 md:p-6 shadow-2xl animate-in fade-in zoom-in-95 duration-150">
            <div className="flex items-center justify-between pb-3.5 border-b border-white/8">
              <h3 className="text-sm font-semibold text-slate-100 flex items-center space-x-2">
                <Sparkles className="w-4 h-4 text-amber-300 animate-pulse" />
                <span>{t('graph.autoScanModalTitle')}</span>
              </h3>
              <button
                onClick={() => setIsAutoScanOpen(false)}
                className="text-slate-400 hover:text-white p-1 rounded-lg hover:bg-white/6 transition"
              >
                <X className="w-4 h-4" />
              </button>
            </div>

            {scanError && (
              <div className="mt-4 p-3 rounded-lg bg-rose-500/10 border border-rose-500/20 text-rose-300 text-xs flex items-center space-x-2">
                <AlertTriangle className="w-4 h-4 shrink-0" />
                <span>{scanError}</span>
              </div>
            )}

            {scanResult && (
              <div className="mt-4 p-3 rounded-lg bg-emerald-500/10 border border-emerald-500/20 text-emerald-300 text-xs flex items-center justify-between">
                <div className="flex items-center space-x-2">
                  <Check className="w-4 h-4 shrink-0 text-emerald-400" />
                  <span>{scanResult.message}</span>
                </div>
                {scanResult.duration_ms && (
                  <span className="font-mono text-emerald-400 font-semibold text-[11px] shrink-0 ml-2">
                    {(scanResult.duration_ms / 1000).toFixed(1)}s
                  </span>
                )}
              </div>
            )}

            <div className="mt-4 space-y-4 text-xs">
              <div>
                <label className="text-slate-400 block mb-2 font-medium">{t('graph.scanModeLabel')}</label>
                <div className="grid grid-cols-1 gap-2.5">
                  <button
                    type="button"
                    onClick={() => setScanMode('volume')}
                    className={`p-3 rounded-xl border text-left transition flex items-start space-x-3 ${
                      scanMode === 'volume'
                        ? 'bg-indigo-600/15 border-indigo-500/40 text-white shadow-xs'
                        : 'bg-[#111726] border-white/6 text-slate-300 hover:border-white/20'
                    }`}
                  >
                    <Sparkles className="w-4 h-4 text-amber-300 shrink-0 mt-0.5 animate-pulse" />
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center space-x-2 mb-0.5">
                        <span className="font-semibold text-slate-100">{t('graph.modeVolumeTitle')}</span>
                        <span className="px-1.5 py-0.2 bg-linear-to-r from-amber-500 to-indigo-500 text-white text-[9px] font-bold rounded uppercase">
                          {t('graph.modeVolumeBadge')}
                        </span>
                      </div>
                      <div className="text-[11px] text-slate-400 leading-relaxed">
                        {t('graph.modeVolumeDesc')}
                      </div>
                      <div className="mt-1 text-[10.5px] text-emerald-400 font-medium">
                        {t('graph.modeVolumeBenchmark')}
                      </div>
                    </div>
                  </button>

                  <button
                    type="button"
                    onClick={() => setScanMode('current_chapter')}
                    className={`p-3 rounded-xl border text-left transition flex items-start space-x-3 ${
                      scanMode === 'current_chapter'
                        ? 'bg-indigo-600/15 border-indigo-500/40 text-white shadow-xs'
                        : 'bg-[#111726] border-white/6 text-slate-300 hover:border-white/20'
                    }`}
                  >
                    <BookOpen className="w-4 h-4 text-indigo-400 shrink-0 mt-0.5" />
                    <div>
                      <div className="font-semibold text-slate-100 mb-0.5">{t('graph.modeChapterTitle')}</div>
                      <div className="text-[11px] text-slate-400">
                        {t('graph.modeChapterDesc')}
                      </div>
                    </div>
                  </button>

                  <button
                    type="button"
                    onClick={() => setScanMode('all')}
                    className={`p-3 rounded-xl border text-left transition flex items-start space-x-3 ${
                      scanMode === 'all'
                        ? 'bg-indigo-600/15 border-indigo-500/40 text-white shadow-xs'
                        : 'bg-[#111726] border-white/6 text-slate-300 hover:border-white/20'
                    }`}
                  >
                    <Network className="w-4 h-4 text-emerald-400 shrink-0 mt-0.5" />
                    <div>
                      <div className="font-semibold text-slate-100 mb-0.5">{t('graph.modeFullTitle')}</div>
                      <div className="text-[11px] text-slate-400">
                        {t('graph.modeFullDesc')}
                      </div>
                    </div>
                  </button>
                </div>
              </div>

              {scanMode === 'volume' && volumes.length > 0 && (
                <div className="space-y-1.5 p-3 rounded-xl bg-[#111726] border border-white/8">
                  <label className="text-slate-300 block font-medium">{t('graph.selectVolLabel')}</label>
                  <Select
                    value={scanTargetVolume || selectedVolume || volumes[0]}
                    onChange={(e) => setScanTargetVolume(e.target.value)}
                    className="w-full text-xs"
                  >
                    {volumes.map((vol) => {
                      const info = volumeInfoMap.get(vol);
                      return (
                        <option key={vol} value={vol}>
                          {t('graph.volOption', { vol, count: info ? info.totalChapters : 0, start: info ? info.startChapter : 1 })}
                        </option>
                      );
                    })}
                  </Select>
                  <p className="text-[10.5px] text-slate-400">
                    {t('graph.filterNotice')}
                  </p>
                </div>
              )}

              {scanMode === 'current_chapter' && chapters && chapters.length > 0 && (() => {
                const currentCh = chapters.find(c => c.chapter_index === selectedScanChapter);
                const currentTitle = (currentCh?.title || '').toLowerCase();
                const curLen = (currentCh?.raw_content?.length || 0) + (currentCh?.translated_content?.length || 0);
                const isCurrentMeta = currentTitle.includes('cover') || currentTitle.includes('insert') || currentTitle.includes('title page') || currentTitle.includes('copyright') || currentTitle.includes('toc') || currentTitle.includes('contents') || currentTitle.includes('bìa') || currentTitle.includes('mục lục') || (curLen < 150);

                return (
                  <div className="space-y-1.5">
                    <label className="text-slate-400 block font-medium">{t('graph.selectChapterLabel')}</label>
                    <Select
                      value={String(selectedScanChapter)}
                      onChange={(e) => setSelectedScanChapter(Number(e.target.value))}
                      className="w-full text-xs"
                    >
                      {chapters.map((c) => {
                        const titleLower = (c.title || '').toLowerCase();
                        const cLen = (c.raw_content?.length || 0) + (c.translated_content?.length || 0);
                        const isMeta = titleLower.includes('cover') || titleLower.includes('insert') || titleLower.includes('title page') || titleLower.includes('copyright') || titleLower.includes('toc') || titleLower.includes('contents') || titleLower.includes('bìa') || titleLower.includes('mục lục') || (cLen < 150);
                        return (
                          <option key={c.id} value={c.chapter_index}>
                            Ch.{c.chapter_index}: {c.title} {isMeta ? t('graph.appendixTag') : ''}
                          </option>
                        );
                      })}
                    </Select>
                    {isCurrentMeta && (
                      <div className="p-2 rounded-lg bg-amber-500/10 border border-amber-500/20 text-amber-300 text-[11px] flex items-center space-x-1.5 mt-1">
                        <AlertTriangle className="w-3.5 h-3.5 shrink-0" />
                        <span>{t('graph.appendixNotice')}</span>
                      </div>
                    )}
                  </div>
                );
              })()}

              <div className="flex items-center justify-end space-x-2 pt-3 border-t border-white/8">
                <button
                  type="button"
                  onClick={() => setIsAutoScanOpen(false)}
                  disabled={isScanning}
                  className="px-3.5 py-1.5 bg-white/4 hover:bg-white/8 text-slate-300 rounded-lg border border-white/10 transition"
                >
                  {t('common.close')}
                </button>
                <button
                  type="button"
                  onClick={() => handleRunAutoScan()}
                  disabled={isScanning}
                  className="px-5 py-2 bg-linear-to-r from-indigo-600 via-purple-600 to-pink-600 hover:from-indigo-500 hover:to-purple-500 text-white rounded-lg font-semibold shadow-md flex items-center space-x-2 transition disabled:opacity-50"
                >
                  {isScanning ? (
                    <>
                      <Loader2 className="w-4 h-4 animate-spin" />
                      <span>{t('graph.analyzingLLM')}</span>
                    </>
                  ) : (
                    <>
                      <Sparkles className="w-4 h-4 text-amber-300" />
                      <span>{t('graph.startScanBtn')}</span>
                    </>
                  )}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
