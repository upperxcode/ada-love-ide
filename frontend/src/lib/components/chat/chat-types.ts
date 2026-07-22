export type ActionTool = 'exec' | 'read' | 'write' | 'search' | 'plan' | 'unknown';

export interface ActionLog {
  id: string;
  tool: ActionTool;
  status: 'pending' | 'done' | 'error' | 'expanded';
  label: string;
  detail?: string;
  diffStats?: string;
  resultCount?: number;
  command?: string;
  filePath?: string;
}

export interface ThinkingSection {
  type: 'plan' | 'explore' | 'exec' | 'read' | 'diff' | 'text';
  content: string;
}

export interface ChatMessage {
  role: 'user' | 'assistant';
  content: string;
  thinkingDuration?: number;
  thinkingContent?: string;
  thinkingSections?: ThinkingSection[];
  actions?: ActionLog[];
}

/**
 * Parses a raw action payload from the backend into a typed ActionLog.
 *
 * Backend emits payloads like:
 *   - "🖥️  <command>"             (pending exec)
 *   - "📖  Reading <file>"        (pending read)
 *   - "✏️  Writing <file>"        (pending write)
 *   - "🔍  Explore \"<query>\""   (pending search)
 *   - "📋  Planning: <task>"      (pending plan)
 *   - "🖥️  <cmd>  ✅ <firstline>" (done exec)
 *   - "✏️  Edited 📤 <file> +5 -4" (done write)
 *   - "🔍  Explore \"<q>\"  N results" (done search)
 *   - "📖  Read <file>"           (done read)
 *   - "📋  Planned"               (done plan)
 */
export function parseActionPayload(payload: string, id: string, meta: string, status: 'pending' | 'done'): ActionLog {
  const base: ActionLog = { id, tool: 'unknown', status, label: payload };

  // Detect tool type from emojis/prefixes in the payload
  if (payload.includes('Explore') || payload.includes('\u{1F50D}')) {
    base.tool = 'search';
    // Extract query between quotes
    const qMatch = payload.match(/"([^"]+)"/);
    if (qMatch) base.label = qMatch[1];
    else base.label = stripEmojis(payload).trim();
    // Extract result count: "N results"
    const countMatch = payload.match(/(\d+)\s+results?/);
    if (countMatch) base.resultCount = parseInt(countMatch[1], 10);
  } else if (payload.includes('Edited') || payload.includes('Writing') || payload.includes('\u270F')) {
    base.tool = 'write';
    // Extract file path (after "📤 " or "Writing " or "Edited ")
    const pathMatch = payload.match(/(?:\u{1F4E4}\s*|(?:Writing|Edited)\s+)(.+)/);
    if (pathMatch) {
      const pathPart = pathMatch[1].replace(/\s*\+.*$/, '').trim();
      base.filePath = pathPart;
      base.label = pathPart;
    } else {
      base.label = stripEmojis(payload).replace(/^(Edited|Writing)\s*/, '').trim();
    }
    // Extract diff stats: "+5 -4"
    const diffMatch = payload.match(/(\+\d+\s*-\d+)/);
    if (diffMatch) base.diffStats = diffMatch[1];
  } else if (payload.includes('Read') || payload.includes('Reading') || payload.includes('\u{1F4D6}')) {
    base.tool = 'read';
    const pathMatch = payload.match(/(?:Reading|Read)\s+(.+)/);
    if (pathMatch) {
      base.filePath = pathMatch[1].trim();
      base.label = pathMatch[1].trim();
    } else {
      base.label = stripEmojis(payload).replace(/^(Reading|Read)\s*/, '').trim();
    }
  } else if (payload.includes('Planning') || payload.includes('Planned') || payload.includes('\u{1F4CB}')) {
    base.tool = 'plan';
    base.label = status === 'done' ? 'Planned' : 'Planning...';
  } else if (payload.includes('\u{1F5A5}') || payload.includes('exec')) {
    base.tool = 'exec';
    // The command is the main payload content (after emoji, before ✅/❌)
    const cmdPart = stripEmojis(payload).replace(/\s*[✅❌].*$/, '').trim();
    base.command = cmdPart;
    base.label = cmdPart;
    if (payload.includes('\u274C')) base.status = 'error';
  } else {
    base.label = stripEmojis(payload).trim();
  }

  // Detail is the raw tool output (meta field) for done status
  if (status === 'done' && meta && meta !== 'pending') {
    base.detail = meta;
  }

  return base;
}

function stripEmojis(text: string): string {
  // Remove common emoji + variation selectors + zero-width joiners
  return text
    .replace(/[\u{1F300}-\u{1FAFF}]\u{FE0F}?/gu, '')
    .replace(/[\u{2600}-\u{26FF}]\u{FE0F}?/gu, '')
    .replace(/[\u{2700}-\u{27BF}]\u{FE0F}?/gu, '')
    .replace(/\s{2,}/g, ' ')
    .trim();
}

/** Icon name from icon-map for a given tool type */
export function toolIcon(tool: ActionTool): string {
  switch (tool) {
    case 'search': return 'search';
    case 'write': return 'pencil';
    case 'read': return 'eye';
    case 'exec': return 'terminal';
    case 'plan': return 'layers';
    default: return 'bot';
  }
}