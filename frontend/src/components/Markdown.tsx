// Minimal, dependency-free Markdown renderer covering the subset used in the
// lesson content: headings (##, ###), paragraphs, unordered/ordered lists,
// blockquotes, fenced code blocks, inline code, bold, and links.
// Not a general-purpose Markdown parser.

import { useMemo, type JSX } from 'react';

function escapeHtml(s: string): string {
  return s
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;');
}

function renderInline(text: string): string {
  let out = escapeHtml(text);
  // inline code
  out = out.replace(/`([^`]+)`/g, '<code>$1</code>');
  // bold
  out = out.replace(/\*\*([^*]+)\*\*/g, '<strong>$1</strong>');
  // links [text](url)
  out = out.replace(/\[([^\]]+)\]\(([^)]+)\)/g, '<a href="$2" target="_blank" rel="noopener noreferrer" class="text-brand-600 underline hover:text-brand-700">$1</a>');
  return out;
}

interface Block {
  type: 'h2' | 'h3' | 'p' | 'ul' | 'ol' | 'blockquote' | 'pre';
  content?: string;
  items?: string[];
  lang?: string;
}

function parseBlocks(md: string): Block[] {
  const lines = md.replace(/\r\n/g, '\n').split('\n');
  const blocks: Block[] = [];
  let i = 0;

  while (i < lines.length) {
    const line = lines[i];

    // blank line
    if (/^\s*$/.test(line)) { i++; continue; }

    // fenced code block
    const fence = line.match(/^```(\w*)\s*$/);
    if (fence) {
      const lang = fence[1] || '';
      const code: string[] = [];
      i++;
      while (i < lines.length && !/^```\s*$/.test(lines[i])) {
        code.push(lines[i]);
        i++;
      }
      i++; // skip closing fence
      blocks.push({ type: 'pre', content: code.join('\n'), lang });
      continue;
    }

    // heading
    const h3 = line.match(/^###\s+(.*)$/);
    if (h3) { blocks.push({ type: 'h3', content: h3[1] }); i++; continue; }
    const h2 = line.match(/^##\s+(.*)$/);
    if (h2) { blocks.push({ type: 'h2', content: h2[1] }); i++; continue; }

    // blockquote
    if (/^>\s?/.test(line)) {
      const quote: string[] = [];
      while (i < lines.length && /^>\s?/.test(lines[i])) {
        quote.push(lines[i].replace(/^>\s?/, ''));
        i++;
      }
      blocks.push({ type: 'blockquote', content: quote.join(' ') });
      continue;
    }

    // unordered list
    if (/^[-*]\s+/.test(line)) {
      const items: string[] = [];
      while (i < lines.length && /^[-*]\s+/.test(lines[i])) {
        items.push(lines[i].replace(/^[-*]\s+/, ''));
        i++;
      }
      blocks.push({ type: 'ul', items });
      continue;
    }

    // ordered list
    if (/^\d+\.\s+/.test(line)) {
      const items: string[] = [];
      while (i < lines.length && /^\d+\.\s+/.test(lines[i])) {
        items.push(lines[i].replace(/^\d+\.\s+/, ''));
        i++;
      }
      blocks.push({ type: 'ol', items });
      continue;
    }

    // paragraph: gather consecutive non-blank, non-special lines
    const para: string[] = [];
    while (
      i < lines.length &&
      !/^\s*$/.test(lines[i]) &&
      !/^```/.test(lines[i]) &&
      !/^#{1,3}\s+/.test(lines[i]) &&
      !/^>\s?/.test(lines[i]) &&
      !/^[-*]\s+/.test(lines[i]) &&
      !/^\d+\.\s+/.test(lines[i])
    ) {
      para.push(lines[i]);
      i++;
    }
    blocks.push({ type: 'p', content: para.join(' ') });
  }

  return blocks;
}

export function Markdown({ source }: { source: string }): JSX.Element {
  const blocks = useMemo(() => parseBlocks(source), [source]);

  return (
    <div className="prose-lesson">
      {blocks.map((b, idx) => {
        switch (b.type) {
          case 'h2': return <h2 key={idx} dangerouslySetInnerHTML={{ __html: renderInline(b.content || '') }} />;
          case 'h3': return <h3 key={idx} dangerouslySetInnerHTML={{ __html: renderInline(b.content || '') }} />;
          case 'p': return <p key={idx} dangerouslySetInnerHTML={{ __html: renderInline(b.content || '') }} />;
          case 'blockquote': return <blockquote key={idx} dangerouslySetInnerHTML={{ __html: renderInline(b.content || '') }} />;
          case 'ul':
            return <ul key={idx}>{(b.items || []).map((it, j) => <li key={j} dangerouslySetInnerHTML={{ __html: renderInline(it) }} />)}</ul>;
          case 'ol':
            return <ol key={idx}>{(b.items || []).map((it, j) => <li key={j} dangerouslySetInnerHTML={{ __html: renderInline(it) }} />)}</ol>;
          case 'pre':
            return (
              <pre key={idx}><code dangerouslySetInnerHTML={{ __html: escapeHtml(b.content || '') }} /></pre>
            );
          default: return null;
        }
      })}
    </div>
  );
}
