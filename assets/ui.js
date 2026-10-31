// go-htemel/ui progressive enhancement. Optional: without this script every component still
// works (native drag and drop for a mouse, the Pick up / Move here buttons for touch and keys;
// a MarkdownEditor without its toolbar; a TagInput that submits its server-rendered tags; a
// FileUpload that is a plain file input).
//
// <ui-sortable> adds touch (and pen) dragging to the Draggable and DropZone components inside it.
// Mouse pointers are left to the browser's native HTML5 drag and drop.
//
// Contract (light DOM, no shadow root, no dependencies):
//   [data-ui-drag]      a draggable element; data-ui-payload is its payload
//   [data-ui-dropzone]  a drop target; gets data-over while a drag is above it
//   ui:dragstart        on the dragged element, bubbles, detail {payload}
//   ui:drop             on the zone, bubbles, detail {payload}
//   ui:dragend          on the dragged element, bubbles, detail {payload, dropped}
'use strict';

(() => {
  const HOLD_MS = 150; // long press before a touch drag starts, so a swipe still scrolls
  const SLOP = 8; // px of movement during the hold that turns the gesture into a scroll

  class UISortable extends HTMLElement {
    constructor() {
      super();
      this._s = null;
      this._down = this._down.bind(this);
      this._move = this._move.bind(this);
      this._up = this._up.bind(this);
      this._cancel = this._cancel.bind(this);
      this._block = this._block.bind(this);
    }

    connectedCallback() {
      this.addEventListener('pointerdown', this._down);
      this.addEventListener('pointermove', this._move);
      this.addEventListener('pointerup', this._up);
      this.addEventListener('pointercancel', this._cancel);
      // Non-passive, so a drag in progress can stop the page from scrolling under the finger.
      this.addEventListener('touchmove', this._block, { passive: false });
      this.addEventListener('contextmenu', this._block);
    }

    disconnectedCallback() {
      this._end(false);
      this.removeEventListener('pointerdown', this._down);
      this.removeEventListener('pointermove', this._move);
      this.removeEventListener('pointerup', this._up);
      this.removeEventListener('pointercancel', this._cancel);
      this.removeEventListener('touchmove', this._block);
      this.removeEventListener('contextmenu', this._block);
    }

    _block(evt) {
      if (this._s && (this._s.active || evt.type === 'contextmenu')) evt.preventDefault();
    }

    _down(evt) {
      if (this._s || evt.pointerType === 'mouse' || !evt.isPrimary) return;
      const src = evt.target.closest('[data-ui-drag], [draggable="true"]');
      if (!src || !this.contains(src)) return;
      const s = (this._s = {
        src, id: evt.pointerId, x: evt.clientX, y: evt.clientY, dx: 0, dy: 0,
        active: false, ghost: null, zone: null,
        payload: src.getAttribute('data-ui-payload') ?? '',
      });
      s.timer = setTimeout(() => this._start(), HOLD_MS);
    }

    _start() {
      const s = this._s;
      if (!s) return;
      s.timer = 0;
      s.active = true;
      const r = s.src.getBoundingClientRect();
      s.dx = s.x - r.left;
      s.dy = s.y - r.top;
      const g = s.src.cloneNode(true);
      g.removeAttribute('id');
      g.querySelectorAll('[id]').forEach((n) => n.removeAttribute('id'));
      g.setAttribute('data-ignore', ''); // Datastar must not bind the copy
      g.setAttribute('aria-hidden', 'true');
      g.removeAttribute('data-ui-drag');
      Object.assign(g.style, {
        position: 'fixed', left: '0', top: '0', width: r.width + 'px', margin: '0',
        pointerEvents: 'none', zIndex: '2147483647', opacity: '0.9',
        boxShadow: '0 8px 24px rgb(0 0 0 / 0.25)', transform: this._at(r.left, r.top),
      });
      document.body.appendChild(g);
      s.ghost = g;
      try { s.src.setPointerCapture(s.id); } catch { /* pointer already gone */ }
      s.src.toggleAttribute('data-dragging', true);
      s.src.setAttribute('aria-grabbed', 'true');
      this._fire(s.src, 'ui:dragstart', { payload: s.payload });
    }

    _at(x, y) { return 'translate(' + Math.round(x) + 'px,' + Math.round(y) + 'px)'; }

    _move(evt) {
      const s = this._s;
      if (!s || evt.pointerId !== s.id) return;
      if (!s.active) {
        // Moved before the hold elapsed: the user is scrolling, not dragging.
        if (Math.hypot(evt.clientX - s.x, evt.clientY - s.y) > SLOP) this._end(false);
        return;
      }
      s.ghost.style.transform = this._at(evt.clientX - s.dx, evt.clientY - s.dy);
      const zone = this._zoneAt(evt.clientX, evt.clientY);
      if (zone !== s.zone) {
        s.zone?.removeAttribute('data-over');
        zone?.setAttribute('data-over', '');
        s.zone = zone;
      }
    }

    _zoneAt(x, y) {
      for (const el of document.elementsFromPoint(x, y)) {
        const z = el.closest('[data-ui-dropzone]');
        if (z) return z;
      }
      return null;
    }

    _up(evt) {
      const s = this._s;
      if (s && evt.pointerId === s.id) this._end(s.active);
    }

    _cancel(evt) {
      const s = this._s;
      if (s && evt.pointerId === s.id) this._end(false);
    }

    _fire(el, type, detail) {
      el.dispatchEvent(new CustomEvent(type, { bubbles: true, detail }));
    }

    // _end finishes the gesture; drop is true when the pointer was released normally.
    _end(drop) {
      const s = this._s;
      if (!s) return;
      this._s = null;
      clearTimeout(s.timer);
      if (!s.active) return;
      s.ghost.remove();
      s.zone?.removeAttribute('data-over');
      s.src.toggleAttribute('data-dragging', false);
      s.src.setAttribute('aria-grabbed', 'false');
      try { s.src.releasePointerCapture(s.id); } catch { /* already released */ }
      // The browser may still synthesize a click on release; it must not press "Pick up".
      const eat = (e) => e.stopPropagation();
      s.src.addEventListener('click', eat, { capture: true, once: true });
      setTimeout(() => s.src.removeEventListener('click', eat, { capture: true }), 400);
      const dropped = drop && s.zone !== null;
      if (dropped) this._fire(s.zone, 'ui:drop', { payload: s.payload });
      this._fire(s.src, 'ui:dragend', { payload: s.payload, dropped });
    }
  }

  if (!customElements.get('ui-sortable')) customElements.define('ui-sortable', UISortable);
})();

// <ui-mdeditor> runs the toolbar and the shortcuts of MarkdownEditor. The markup only says what
// each button does; this element does the editing, so no attribute holds script.
//
// Contract (light DOM, no shadow root, no dependencies):
//   textarea                         the text (the first one inside the element)
//   [role=toolbar]                   roving tabindex: ArrowLeft/ArrowRight/Home/End move focus
//   [data-ui-md-action]              a button: bold italic strike code link image heading quote
//                                    ul ol task table
//   Ctrl/Cmd+B I E K, Shift+X        bold, italic, code, link, strikethrough
//   Enter in a list item             continues the list; on an empty item, ends it
//   data-ui-upload (on the element)  pasted or dropped files are not inserted: each one dispatches
//   ui:md-upload                     on the textarea, bubbles, detail {file, insert(text)}
// Every edit goes through execCommand('insertText') where it works (so undo works, and the
// browser fires input), else setRangeText plus a dispatched input event: data-bind sees it.
(() => {
  const WRAP = { bold: ['**', '**', 'bold text'], italic: ['_', '_', 'italic text'], strike: ['~~', '~~', 'struck text'], code: ['`', '`', 'code'] };
  const PREFIX = { heading: '## ', quote: '> ', ul: '- ', task: '- [ ] ' };
  const KEYS = { b: 'bold', i: 'italic', e: 'code', k: 'link' };
  const LIST = /^(\s*)([-*+] \[[ xX]\] |[-*+] |(\d+)([.)]) )(.*)$/;

  class UIMDEditor extends HTMLElement {
    constructor() {
      super();
      this._click = this._click.bind(this);
      this._key = this._key.bind(this);
      this._files = this._files.bind(this);
    }

    connectedCallback() {
      this.addEventListener('click', this._click);
      this.addEventListener('keydown', this._key);
      this.addEventListener('paste', this._files);
      this.addEventListener('drop', this._files);
    }

    disconnectedCallback() {
      this.removeEventListener('click', this._click);
      this.removeEventListener('keydown', this._key);
      this.removeEventListener('paste', this._files);
      this.removeEventListener('drop', this._files);
    }

    get textarea() { return this.querySelector('textarea'); }

    _click(evt) {
      const b = evt.target.closest('[data-ui-md-action]');
      if (!b || !this.contains(b)) return;
      this.querySelectorAll('[role=toolbar] [data-ui-md-action]').forEach((x) => { x.tabIndex = x === b ? 0 : -1; });
      this.run(b.getAttribute('data-ui-md-action'));
    }

    _key(evt) {
      const ta = this.textarea;
      if (evt.target.closest('[role=toolbar]')) return this._rove(evt);
      if (evt.target !== ta || evt.isComposing || evt.altKey) return;
      const k = evt.key.toLowerCase();
      if ((evt.ctrlKey || evt.metaKey) && (KEYS[k] && !evt.shiftKey || k === 'x' && evt.shiftKey)) {
        evt.preventDefault();
        this.run(k === 'x' ? 'strike' : KEYS[k]);
      } else if (evt.key === 'Enter' && !evt.shiftKey && !evt.ctrlKey && !evt.metaKey && this._continueList()) {
        evt.preventDefault();
      }
    }

    _rove(evt) {
      const items = [...this.querySelectorAll('[role=toolbar] [data-ui-md-action]')];
      const i = items.indexOf(document.activeElement);
      if (i < 0) return;
      const n = { ArrowRight: (i + 1) % items.length, ArrowLeft: (i + items.length - 1) % items.length, Home: 0, End: items.length - 1 }[evt.key];
      if (n === undefined) return;
      evt.preventDefault();
      items.forEach((b, j) => { b.tabIndex = j === n ? 0 : -1; });
      items[n].focus();
    }

    // run applies a toolbar action to the selection.
    run(action) {
      const ta = this.textarea;
      if (!ta || ta.disabled || ta.readOnly) return;
      const v = ta.value, s = ta.selectionStart, e = ta.selectionEnd, sel = v.slice(s, e);
      if (action === 'code' && sel.includes('\n')) return this._block(s, e, '```\n' + sel + '\n```');
      if (WRAP[action]) {
        const [l, r, ph] = WRAP[action];
        if (v.slice(s - l.length, s) === l && v.slice(e, e + r.length) === r) {
          return this._replace(s - l.length, e + r.length, sel, 0, sel.length); // toggle off
        }
        const t = sel || ph;
        return this._replace(s, e, l + t + r, l.length, l.length + t.length);
      }
      if (action === 'link' || action === 'image') {
        const t = sel || (action === 'link' ? 'link text' : 'alt text');
        const head = (action === 'image' ? '!' : '') + '[' + t + '](';
        return this._replace(s, e, head + 'https://)', head.length, head.length + 8);
      }
      if (action === 'table') return this._block(s, e, '| Column | Column |\n| ------ | ------ |\n| Cell   | Cell   |');
      if (PREFIX[action] || action === 'ol') this._lines(action);
    }

    // _lines toggles a prefix on every line the selection touches.
    _lines(action) {
      const ta = this.textarea, v = ta.value;
      const a = v.lastIndexOf('\n', ta.selectionStart - 1) + 1;
      let b = v.indexOf('\n', ta.selectionEnd - (ta.selectionEnd > ta.selectionStart && v[ta.selectionEnd - 1] === '\n' ? 1 : 0));
      if (b < 0) b = v.length;
      const lines = v.slice(a, b).split('\n');
      const has = (l) => (action === 'ol' ? /^\d+\. /.test(l) : l.startsWith(PREFIX[action]));
      const off = lines.every(has);
      const out = lines.map((l, i) => {
        if (off) return action === 'ol' ? l.replace(/^\d+\. /, '') : l.slice(PREFIX[action].length);
        const bare = l.replace(/^(#{1,6} |> |[-*+] \[[ xX]\] |[-*+] |\d+\. )/, '');
        return (action === 'ol' ? (i + 1) + '. ' : PREFIX[action]) + bare;
      }).join('\n');
      this._replace(a, b, out, 0, out.length);
    }

    // _block puts text on lines of its own.
    _block(s, e, text) {
      const v = this.textarea.value;
      const before = s > 0 && v[s - 1] !== '\n' ? '\n\n' : '', after = e < v.length && v[e] !== '\n' ? '\n\n' : '';
      this._replace(s, e, before + text + after, before.length, before.length + text.length);
    }

    _continueList() {
      const ta = this.textarea, v = ta.value, s = ta.selectionStart;
      if (s !== ta.selectionEnd) return false;
      const a = v.lastIndexOf('\n', s - 1) + 1;
      const m = LIST.exec(v.slice(a, s));
      if (!m) return false;
      if (!m[5].trim() && s === (v.indexOf('\n', s) < 0 ? v.length : v.indexOf('\n', s))) {
        this._replace(a, s, '', 0, 0); // an empty item ends the list
        return true;
      }
      const marker = m[3] ? (+m[3] + 1) + m[4] + ' ' : m[2].replace(/\[[xX]\]/, '[ ]');
      const t = '\n' + m[1] + marker;
      this._replace(s, s, t, t.length, t.length);
      return true;
    }

    // _replace swaps [start, end) for text and selects [start+a, start+b) of it.
    _replace(start, end, text, a, b) {
      const ta = this.textarea;
      ta.focus();
      ta.setSelectionRange(start, end);
      let ok = false;
      try { ok = document.execCommand('insertText', false, text); } catch { ok = false; }
      if (!ok || ta.value.slice(start, start + text.length) !== text) {
        ta.setRangeText(text, start, end, 'end');
        ta.dispatchEvent(new Event('input', { bubbles: true }));
      }
      ta.setSelectionRange(start + a, start + b);
    }

    _files(evt) {
      if (!this.hasAttribute('data-ui-upload') || evt.target !== this.textarea) return;
      const files = [...((evt.clipboardData || evt.dataTransfer || {}).files || [])];
      if (!files.length) return;
      evt.preventDefault();
      const ta = this.textarea;
      for (const file of files) {
        const insert = (text) => this._replace(ta.selectionStart, ta.selectionEnd, text, text.length, text.length);
        ta.dispatchEvent(new CustomEvent('ui:md-upload', { bubbles: true, detail: { file, insert } }));
      }
    }
  }

  if (!customElements.get('ui-mdeditor')) customElements.define('ui-mdeditor', UIMDEditor);
})();

// <ui-taginput> edits the chips of TagInput. Each chip carries its hidden <input name>, so the
// form value is always the chips on screen.
//
// Contract (light DOM, no shadow root, no dependencies):
//   data-ui-name, data-ui-max          the field name; the tag limit (optional)
//   [data-ui-tag="go"]                 a chip, with a [data-ui-tag-remove] button and its hidden input
//   [data-ui-tag-input]                the text field; Enter or "," adds, Backspace when empty removes
//   template[data-ui-tag-template]     the chip cloned for new tags
//   [data-ui-tag-status]               polite live region for "Added go", "Removed go", refusals
//   ui:tags                            on the element, bubbles, detail {tags} after every change
//   add(text), remove(chip), tags      the same operations for scripts
// Text left in the field when the form is submitted is appended to the form data as one more tag.
(() => {
  class UITagInput extends HTMLElement {
    constructor() {
      super();
      for (const m of ['_key', '_input', '_click', '_formdata']) this[m] = this[m].bind(this);
    }

    connectedCallback() {
      this.addEventListener('keydown', this._key);
      this.addEventListener('input', this._input);
      this.addEventListener('click', this._click);
      this._form = this.closest('form');
      this._form?.addEventListener('formdata', this._formdata);
    }

    disconnectedCallback() {
      this.removeEventListener('keydown', this._key);
      this.removeEventListener('input', this._input);
      this.removeEventListener('click', this._click);
      this._form?.removeEventListener('formdata', this._formdata);
    }

    get field() { return this.querySelector('[data-ui-tag-input]'); }
    get chips() { return [...this.querySelectorAll('[data-ui-tag]')]; }
    get tags() { return this.chips.map((c) => c.getAttribute('data-ui-tag')); }

    add(text) {
      const tag = String(text).trim().replace(/\s+/g, ' ');
      if (!tag) return false;
      const max = +this.dataset.uiMax || 0;
      if (this.tags.some((t) => t.toLowerCase() === tag.toLowerCase())) return this._say(tag + ' is already added'), false;
      if (max && this.chips.length >= max) return this._say('Up to ' + max + ' tags'), false;
      const chip = this.querySelector('template[data-ui-tag-template]').content.firstElementChild.cloneNode(true);
      chip.setAttribute('data-ui-tag', tag);
      chip.querySelector('[data-ui-tag-text]').textContent = tag;
      chip.querySelector('[data-ui-tag-remove]').setAttribute('aria-label', 'Remove ' + tag);
      Object.assign(chip.querySelector('input[type=hidden]'), { name: this.dataset.uiName, value: tag });
      this.querySelector('[role=list]').appendChild(chip);
      this._changed('Added ' + tag);
      return true;
    }

    remove(chip) {
      const tag = chip.getAttribute('data-ui-tag');
      chip.remove();
      this._changed('Removed ' + tag);
      this.field?.focus();
    }

    _key(evt) {
      const f = this.field;
      if (evt.target !== f || evt.isComposing) return;
      if (evt.key === ',' || (evt.key === 'Enter' && f.value.trim())) {
        evt.preventDefault();
        if (this.add(f.value)) f.value = '';
      } else if (evt.key === 'Backspace' && f.value === '' && this.chips.length) {
        evt.preventDefault();
        this.remove(this.chips[this.chips.length - 1]);
      }
    }

    _input(evt) {
      const f = this.field;
      if (evt.target !== f) return;
      if (f.value.includes(',')) {
        const parts = f.value.split(',');
        const rest = parts.pop();
        parts.forEach((p) => this.add(p));
        f.value = rest.trimStart();
        return;
      }
      // Picking a <datalist> suggestion fires input without an inputType (or insertReplacementText).
      const picked = !(evt instanceof InputEvent) || evt.inputType === 'insertReplacementText';
      const list = f.list;
      if (picked && list && [...list.options].some((o) => o.value === f.value) && this.add(f.value)) f.value = '';
    }

    _click(evt) {
      const b = evt.target.closest('[data-ui-tag-remove]');
      if (b && this.contains(b)) return this.remove(b.closest('[data-ui-tag]'));
      if (evt.target === this) this.field?.focus();
    }

    _formdata(evt) {
      const v = this.field?.value.trim();
      if (v && !this.tags.some((t) => t.toLowerCase() === v.toLowerCase())) evt.formData.append(this.dataset.uiName, v);
    }

    _changed(msg) {
      this._say(msg);
      this.dispatchEvent(new CustomEvent('ui:tags', { bubbles: true, detail: { tags: this.tags } }));
    }

    _say(msg) {
      const s = this.querySelector('[data-ui-tag-status]');
      if (s) s.textContent = msg;
    }
  }

  if (!customElements.get('ui-taginput')) customElements.define('ui-taginput', UITagInput);
})();

// <ui-fileupload> runs the drop area and the file list of FileUpload around a native file input
// (which works without it).
//
// Contract (light DOM, no shadow root, no dependencies):
//   input[type=file]                   the input; its FileList is the selection
//   [data-ui-dropzone]                 accepts dropped files; gets data-over during a drag of files
//   [data-ui-files]                    the list, rebuilt from template[data-ui-file-template] rows:
//                                      [data-ui-file-name] [data-ui-file-size] [data-ui-file-preview]
//                                      (an <img>, shown for images) [data-ui-file-icon] [data-ui-file-remove]
//   [data-ui-file-status]              polite live region ("2 files selected", "Removed a.png")
//   data-ui-upload (on the element)    each newly selected file dispatches
//   ui:upload                          on the element, bubbles, detail {file, files}
(() => {
  class UIFileUpload extends HTMLElement {
    constructor() {
      super();
      this._urls = [];
      this._sent = new WeakSet();
      for (const m of ['_change', '_drag', '_drop', '_click']) this[m] = this[m].bind(this);
    }

    connectedCallback() {
      this.addEventListener('change', this._change);
      this.addEventListener('click', this._click);
      for (const t of ['dragenter', 'dragover', 'dragleave']) this.addEventListener(t, this._drag);
      this.addEventListener('drop', this._drop);
    }

    disconnectedCallback() {
      this.removeEventListener('change', this._change);
      this.removeEventListener('click', this._click);
      for (const t of ['dragenter', 'dragover', 'dragleave']) this.removeEventListener(t, this._drag);
      this.removeEventListener('drop', this._drop);
      this._revoke();
    }

    get input() { return this.querySelector('input[type=file]'); }
    get zone() { return this.querySelector('[data-ui-dropzone]'); }

    _change(evt) {
      if (evt.target !== this.input) return;
      const files = [...this.input.files];
      this._render(files);
      if (!this.hasAttribute('data-ui-upload')) return;
      for (const file of files) {
        if (this._sent.has(file)) continue;
        this._sent.add(file);
        this.dispatchEvent(new CustomEvent('ui:upload', { bubbles: true, detail: { file, files } }));
      }
    }

    _drag(evt) {
      if (!evt.dataTransfer || ![...evt.dataTransfer.types].includes('Files')) return;
      evt.preventDefault();
      const z = this.zone;
      if (evt.type === 'dragleave') {
        if (!z.contains(evt.relatedTarget)) z.removeAttribute('data-over');
        return;
      }
      z.setAttribute('data-over', '');
      evt.dataTransfer.dropEffect = 'copy';
    }

    _drop(evt) {
      const dropped = [...(evt.dataTransfer?.files || [])];
      if (!dropped.length) return;
      evt.preventDefault();
      this.zone.removeAttribute('data-over');
      const ok = dropped.filter((f) => this._accepts(f));
      if (!ok.length) return this._say('That file type is not accepted');
      this._set(this.input.multiple ? [...this.input.files, ...ok] : ok.slice(0, 1));
    }

    _accepts(file) {
      const accept = this.input.accept;
      if (!accept) return true;
      const name = file.name.toLowerCase(), type = file.type.toLowerCase();
      return accept.split(',').some((a) => {
        a = a.trim().toLowerCase();
        return a.startsWith('.') ? name.endsWith(a) : a.endsWith('/*') ? type.startsWith(a.slice(0, -1)) : type === a;
      });
    }

    // _set replaces the selection and lets everyone (this element, Datastar, the app) see a change.
    _set(files) {
      const dt = new DataTransfer();
      files.forEach((f) => dt.items.add(f));
      this.input.files = dt.files;
      this.input.dispatchEvent(new Event('change', { bubbles: true }));
    }

    _click(evt) {
      const b = evt.target.closest('[data-ui-file-remove]');
      if (!b || !this.contains(b)) return;
      const files = [...this.input.files];
      const [gone] = files.splice(+b.closest('[data-ui-index]').dataset.uiIndex, 1);
      this._set(files);
      if (gone) this._say('Removed ' + gone.name);
      this.input.focus();
    }

    _render(files) {
      this._revoke();
      const list = this.querySelector('[data-ui-files]'), tpl = this.querySelector('template[data-ui-file-template]');
      list.replaceChildren(...files.map((f, i) => {
        const row = tpl.content.firstElementChild.cloneNode(true);
        row.dataset.uiIndex = i;
        row.querySelector('[data-ui-file-name]').textContent = f.name;
        row.querySelector('[data-ui-file-size]').textContent = this._size(f.size);
        const b = row.querySelector('[data-ui-file-remove]');
        b.setAttribute('aria-label', 'Remove ' + f.name);
        b.title = 'Remove ' + f.name;
        if (f.type.startsWith('image/')) {
          const url = URL.createObjectURL(f);
          this._urls.push(url);
          const img = row.querySelector('[data-ui-file-preview]');
          img.src = url;
          img.hidden = false;
          row.querySelector('[data-ui-file-icon]')?.remove();
        }
        return row;
      }));
      this._say(files.length === 1 ? '1 file selected' : files.length + ' files selected');
    }

    _size(n) {
      return n < 1024 ? n + ' B' : n < 1048576 ? (n / 1024).toFixed(1) + ' KB' : (n / 1048576).toFixed(1) + ' MB';
    }

    _revoke() {
      this._urls.forEach((u) => URL.revokeObjectURL(u));
      this._urls = [];
    }

    _say(msg) {
      const s = this.querySelector('[data-ui-file-status]');
      if (s) s.textContent = msg;
    }
  }

  if (!customElements.get('ui-fileupload')) customElements.define('ui-fileupload', UIFileUpload);
})();
