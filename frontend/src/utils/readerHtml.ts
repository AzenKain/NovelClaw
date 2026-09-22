/**
 * Utility functions for auto-detecting and rendering clean book reader HTML.
 * Adapted and optimized from the reading engine architecture of NovelHub.
 */

export const isHtmlContent = (content?: string | null): boolean => {
  if (!content) return false;
  return /<\/?(?:p|div|section|article|span|h[1-6]|html|body|head|br|img|table|em|strong|blockquote)\b|<!DOCTYPE|<\?xml/i.test(
    content
  );
};

const escapeHtml = (text: string): string => {
  return text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')
    .replace(/"/g, '&quot;')
    .replace(/'/g, '&#039;');
};

export const sanitizeReaderHtml = (
  raw?: string | null,
  projectId?: string,
  contentPath?: string
): string => {
  if (!raw) return '';

  const normalized = raw.normalize('NFC');

  // If content is plain text (no HTML tags), format paragraphs naturally
  if (!isHtmlContent(normalized)) {
    return normalized
      .split(/\r?\n\r?\n+/)
      .map(para => {
        const trimmed = para.trim();
        if (!trimmed) return '';
        const lines = trimmed.split(/\r?\n/).map(line => escapeHtml(line)).join('<br/>');
        return `<p class="mb-4 leading-relaxed">${lines}</p>`;
      })
      .filter(Boolean)
      .join('\n');
  }

  // Strip XML declarations, doctypes, head/meta/link/title/scripts/styles
  const stripped = normalized
    .replace(/<\?xml\b[^>]*\?>/gi, '')
    .replace(/<!DOCTYPE\b[^>]*>/gi, '')
    .replace(/<head\b[^>]*>[\s\S]*?<\/head>/gi, '')
    .replace(/<title\b[^>]*>[\s\S]*?<\/title>/gi, '')
    .replace(/<meta\b[^>]*>/gi, '')
    .replace(/<link\b[^>]*>/gi, '')
    .replace(/<style\b[^>]*>[\s\S]*?<\/style>/gi, '')
    .replace(/<script\b[^>]*>[\s\S]*?<\/script>/gi, '')
    .replace(/<\/?(?:html|head|body)\b[^>]*>/gi, '')
    .replace(/<img\b(?![^>]*\bsrc=["'][^"']+)[\s\S]*?>/gi, '')
    .replace(/<img\b[^>]*\bsrc=["']\s*(?:#|about:blank)?\s*["'][^>]*>/gi, '')
    .replace(/<a\b[^>]*>\s*<\/a>/gi, '')
    .replace(/<p\b[^>]*>\s*<\/p>/gi, '');

  if (typeof document !== 'undefined') {
    try {
      const doc = new DOMParser().parseFromString(`<body>${stripped}</body>`, 'text/html');

      // Remove explicitly hidden elements
      doc
        .querySelectorAll(
          '[hidden], [style*="display: none"], [style*="display:none"], [style*="visibility: hidden"], [style*="visibility:hidden"], .d-none, .hidden, .invisible'
        )
        .forEach(el => el.remove());

      // Remove empty paragraphs without images
      doc.querySelectorAll('p').forEach(p => {
        if (!p.textContent?.trim() && !p.querySelector('img, svg')) {
          p.remove();
        }
      });

      // Remove any lingering script or style tags
      doc.querySelectorAll('script, style, link, meta').forEach(el => el.remove());

      // Resolve and display images cleanly via reader asset server
      doc.querySelectorAll('img').forEach(img => {
        let src = (img.getAttribute('src') || '').trim();

        if (!src || src === '#' || src === 'about:blank') {
          img.remove();
          return;
        }

        // If it's a relative path (e.g. "../Images/cover.jpg") and we have projectId
        if (
          projectId &&
          !src.startsWith('/api/') &&
          !src.startsWith('http://') &&
          !src.startsWith('https://') &&
          !src.startsWith('data:') &&
          !src.startsWith('blob:')
        ) {
          const baseDir = contentPath && contentPath.includes('/')
            ? contentPath.substring(0, contentPath.lastIndexOf('/'))
            : '';
          const joined = baseDir ? `${baseDir}/${src}` : src;
          const parts = joined.split('/');
          const resolvedParts: string[] = [];
          for (const p of parts) {
            if (p === '.' || !p) continue;
            if (p === '..') {
              if (resolvedParts.length > 0) resolvedParts.pop();
            } else {
              resolvedParts.push(p);
            }
          }
          const cleanPath = resolvedParts.join('/');
          src = `/api/v1/reader/${encodeURIComponent(projectId)}/asset/${cleanPath}`;
          img.setAttribute('src', src);
        }

        img.setAttribute('loading', 'lazy');
        img.setAttribute('decoding', 'async');
        img.setAttribute('class', 'reader-image max-w-full h-auto mx-auto my-4 block rounded-lg shadow-md');
        img.setAttribute('onerror', "this.style.display='none';");
      });

      // Wrap orphan text nodes directly under body or div into proper <p> tags
      // to eliminate whitespace collapse where dialogues/paragraphs get squashed into a single block
      const wrapOrphanNodes = (parent: HTMLElement) => {
        const childNodes = Array.from(parent.childNodes);
        for (const child of childNodes) {
          if (child.nodeType === Node.TEXT_NODE) {
            const rawText = child.textContent || '';
            const trimmed = rawText.trim();
            if (trimmed) {
              const lines = rawText.split(/\r?\n+/).map(l => l.trim()).filter(Boolean);
              if (lines.length > 0) {
                const fragment = doc.createDocumentFragment();
                for (const line of lines) {
                  const p = doc.createElement('p');
                  p.className = 'calibre2 leading-relaxed mb-3';
                  p.textContent = line;
                  fragment.appendChild(p);
                }
                parent.replaceChild(fragment, child);
              } else {
                parent.removeChild(child);
              }
            }
          } else if (child.nodeType === Node.ELEMENT_NODE) {
            const el = child as HTMLElement;
            const tag = el.tagName.toLowerCase();
            if (tag === 'div' || tag === 'section' || tag === 'article') {
              wrapOrphanNodes(el);
            }
          }
        }
      };
      wrapOrphanNodes(doc.body);

      return doc.body.innerHTML;
    } catch {
      return stripped;
    }
  }

  return stripped;
};

export const extractCleanNarrativeText = (content?: string | null): string => {
  if (!content) return '';
  if (!/<\/?(?:p|div|section|article|span|h[1-6]|html|body|head|br|img|table|em|strong|blockquote)\b|<!DOCTYPE|<\?xml/i.test(content)) {
    return content.trim();
  }
  try {
    const doc = new DOMParser().parseFromString(content, 'text/html');
    doc.querySelectorAll('script, style, head, meta, link, title, nav, [hidden], svg, img').forEach(el => el.remove());
    const paragraphs: string[] = [];
    doc.querySelectorAll('p, div, blockquote, h1, h2, h3, h4, h5, h6, li').forEach(el => {
      if (el.children.length === 0 || Array.from(el.children).every(c => ['SPAN', 'EM', 'STRONG', 'B', 'I', 'A', 'FONT', 'SMALL'].includes(c.tagName))) {
        const t = el.textContent?.trim();
        if (t && t.length > 0) {
          paragraphs.push(t);
        }
      }
    });
    if (paragraphs.length > 0) {
      return paragraphs.join('\n\n');
    }
    return doc.body.textContent?.trim() || '';
  } catch {
    return content.replace(/<[^>]+>/g, ' ').replace(/\s+/g, ' ').trim();
  }
};

