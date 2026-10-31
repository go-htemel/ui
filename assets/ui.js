// go-htemel/ui progressive enhancement. Optional: without this script every component still
// works (native drag and drop for a mouse, the Pick up / Move here buttons for touch and keys).
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
