package ui

import (
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
)

// ToastRegionID is the id of the ToastRegion container. A backend adds a toast by patching it
// into the region, for example with datastar-go:
//
//	sse.PatchElements(html, datastar.WithSelectorID(ui.ToastRegionID), datastar.WithModeAppend())
const ToastRegionID = "toasts"

// ToastSignal is the object signal holding one boolean per visible toast, keyed by toast id:
// $toasts.saved is true while the toast with id "saved" is shown.
const ToastSignal = "toasts"

// ToastRegion returns the fixed, bottom-right container toasts are appended to. Put it once near
// the end of <body>. It is a polite live region, so screen readers announce new toasts.
//
// Datastar has no client-side list rendering, so toasts are elements rather than entries of a
// signal array: the server appends Toast elements, and each toast's visibility is an entry of
// the $toasts object signal, which the close button and the timeout set to false.
//
// The region is aria-live="polite" with aria-relevant="additions", so only new toasts are
// announced, and each toast is role="status".
func ToastRegion() *html.DivElement {
	return html.Div().
		Id(ToastRegionID).
		Aria("live", "polite").
		Aria("relevant", "additions").
		Class(Classes("pointer-events-none fixed inset-x-4 bottom-4 z-50 flex flex-col items-end gap-3 sm:left-auto sm:w-96")).
		Attrs(initially(ToastSignal, "{}"))
}

var toneAccent = map[Tone]string{
	Neutral: "border-zinc-400 dark:border-zinc-500",
	Info:    "border-blue-500",
	Success: "border-green-500",
	Warning: "border-amber-500",
	Danger:  "border-red-500",
}

// Toast returns a notification card for the ToastRegion. id names it in the $toasts signal and
// must be unique on the page; it is reduced to lowercase letters and digits ("Saved-1" becomes
// "saved1"). The toast hides itself after timeout, or stays until dismissed when timeout is 0.
// message may be empty.
//
// The timeout is a countdown that pauses while the pointer is over the toast or focus is inside
// it (a keyboard user tabbing to the close button), and resumes with the time that was left. It
// is driven by data-on-interval ticking every 250ms (or the whole timeout when that is shorter),
// so it is accurate to one tick. The close button is labelled "Dismiss".
func Toast(id string, t Tone, title, message string, timeout time.Duration) *html.DivElement {
	key := toastKey(id)
	sig := "$" + ToastSignal + "." + key
	content := html.Div(
		html.P().Class(Classes("text-sm font-semibold text-zinc-900 dark:text-white")).Text(title),
	).Class(Classes("min-w-0 flex-1"))
	if message != "" {
		content.Children(html.P().Class(Classes("mt-1 text-sm text-zinc-600 dark:text-zinc-300")).Text(message))
	}
	el := html.Div(
		content,
		closeButton("Dismiss", sig+" = false").AddClass(Classes("text-zinc-500 dark:text-zinc-400")),
	).Id("toast-"+key).
		Role("status").
		Class(Classes("pointer-events-auto flex w-full gap-x-3 rounded-md border-l-4 bg-white p-4 shadow-lg ring-1 ring-zinc-950/5",
			"dark:bg-zinc-800 dark:ring-white/10", pick(toneAccent, t, Neutral))).
		Attrs(datastar.Signal(ToastSignal+"."+key, "true"), datastar.Show(sig))
	if timeout > 0 {
		tick := min(toastTick, timeout)
		el.Attrs(datastar.OnInterval(
			"if (el.matches('*:hover, *:focus-within')) return; "+
				"el._left = (el._left ?? "+strconv.FormatInt(timeout.Milliseconds(), 10)+") - "+strconv.FormatInt(tick.Milliseconds(), 10)+"; "+
				"if (el._left <= 0) "+sig+" = false",
			datastar.Duration(tick)))
	}
	return el
}

// toastTick is how often a toast's countdown advances.
const toastTick = 250 * time.Millisecond

// toastKey reduces id to a safe signal path segment: lowercase ASCII letters and digits, never
// starting with a digit.
func toastKey(id string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(id) {
		if 'a' <= r && r <= 'z' || '0' <= r && r <= '9' {
			b.WriteRune(r)
		}
	}
	key := b.String()
	if key == "" || key[0] <= '9' {
		key = "t" + key
	}
	return key
}

// toastSeq numbers toasts for NextToastID.
var toastSeq atomic.Uint64

// NextToastID returns a process-unique toast id ("t1", "t2", ...), for toasts a server pushes.
func NextToastID() string {
	return "t" + strconv.FormatUint(toastSeq.Add(1), 10)
}
