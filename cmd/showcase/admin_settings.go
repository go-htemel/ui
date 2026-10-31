package main

import (
	"net/http"
	"strings"
	"sync"

	"github.com/go-chi/chi/v5"
	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/datastar"
	"github.com/go-htemel/htemel/html"
	dsse "github.com/starfederation/datastar-go/datastar"

	"github.com/go-htemel/ui"
	"github.com/go-htemel/ui/markdown"
)

// settingsValues are the site settings, kept in memory like the posts (and reset with them).
type settingsValues struct {
	SiteTitle     string `json:"siteTitle"`
	Tagline       string `json:"tagline"`
	Timezone      string `json:"timezone"`
	DefaultAuthor string `json:"defaultAuthor"`
	DefaultStatus string `json:"defaultStatus"`
	Autosave      bool   `json:"autosave"`
}

type settingsStore struct {
	mu sync.Mutex
	v  settingsValues
}

func defaultSettings() settingsValues {
	return settingsValues{
		SiteTitle: "Acme Journal", Tagline: "Notes on building for the web.", Timezone: "Europe/London",
		DefaultAuthor: adminAuthors[0], DefaultStatus: statusDraft, Autosave: true,
	}
}

var adminSettingsData = &settingsStore{v: defaultSettings()}

func (s *settingsStore) get() settingsValues {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.v
}

func (s *settingsStore) set(f func(*settingsValues)) {
	s.mu.Lock()
	defer s.mu.Unlock()
	f(&s.v)
}

func (s *settingsStore) reset() { s.set(func(v *settingsValues) { *v = defaultSettings() }) }

var timezones = []string{"Europe/London", "Europe/Berlin", "America/New_York", "America/Los_Angeles", "Asia/Tokyo", "Australia/Sydney"}

const (
	settingsGeneral    = "general"
	settingsWriting    = "writing"
	settingsAppearance = "appearance"
)

var settingsTabs = []struct{ key, label, href string }{
	{settingsGeneral, "General", "/admin/settings"},
	{settingsWriting, "Writing", "/admin/settings/writing"},
	{settingsAppearance, "Appearance", "/admin/settings/appearance"},
}

const generalHelp = `## About these settings

The **site title** appears in the browser tab and in feeds. The tagline is a *short* line under it.

- Times use the timezone you choose here
- Scheduled posts publish at that local time

Need a fresh start? **Reset demo data** below restores the posts, notes and settings of the seed.`

const writingHelp = `## Markdown cheat sheet

| You type | You get |
| --- | --- |
| ` + "`**bold**`" + ` | **bold** |
| ` + "`[text](url)`" + ` | a link |
| ` + "`- item`" + ` | a bulleted list |

Fenced code is highlighted on the server:

` + "```go\nfmt.Println(\"hello\")\n```" + `
`

// settingsPage is GET /admin/settings and /admin/settings/{writing,appearance}.
func settingsPage(r *http.Request, tab string) htemel.Node {
	nav := make([]ui.NavItem, len(settingsTabs))
	label := ""
	for i, t := range settingsTabs {
		nav[i] = ui.NavLink(t.label, t.href, t.key == tab)
		if t.key == tab {
			label = t.label
		}
	}
	v := adminSettingsData.get()
	var body htemel.FlowContent
	switch tab {
	case settingsGeneral:
		body = settingsGeneralForm(v)
	case settingsWriting:
		body = settingsWritingForm(v)
	default:
		body = settingsAppearanceForm()
	}
	return adminPage("Settings: "+label, navSettings, false, r.URL.Query().Get("notice"),
		adminBody(
			ui.PageHeader("Settings",
				ui.WithBreadcrumbs(ui.Breadcrumbs(ui.Crumb{Label: "Dashboard", Href: "/admin"}, ui.Crumb{Label: "Settings", Href: "/admin/settings"}, ui.Crumb{Label: label})),
				ui.WithDescription("Site-wide options for "+v.SiteTitle+".")),
			ui.SettingsLayout(nav, body),
		))
}

func adminSettingsForm(tab string, signals map[string]any, children ...htemel.FlowContent) *html.FormElement {
	signals["errors"] = map[string]any{"siteTitle": ""}
	return html.Form(
		ui.Stack(ui.Lg, children...),
		html.Div(ui.Button(ui.Primary, "Save changes", ui.WithLoading("saving")).Type(html.ButtonTypeSubmit)).Class("mt-6 flex justify-end"),
	).BoolAttr("novalidate", true).Attrs(
		datastar.Signals(signals),
		datastar.OnSubmit(afterRequest(datastar.Post("/admin/settings/"+tab), "document.querySelector('[aria-invalid=true]')?.focus()"), datastar.Prevent()))
}

func settingsGeneralForm(v settingsValues) htemel.FlowContent {
	zones := []html.SelectChild{}
	for _, z := range timezones {
		zones = append(zones, ui.Choice(z, z).Selected(z == v.Timezone))
	}
	return ui.Stack(ui.Xl,
		adminSettingsForm(settingsGeneral, map[string]any{"siteTitle": v.SiteTitle, "tagline": v.Tagline, "timezone": v.Timezone},
			ui.SectionHeader("General"),
			ui.Field("set-title", "Site title", ui.Input(html.InputTypeText, "siteTitle").Attrs(datastar.Bind("siteTitle")), ui.ErrorSignal("errors.siteTitle")),
			ui.Field("set-tagline", "Tagline", ui.Input(html.InputTypeText, "tagline").Attrs(datastar.Bind("tagline")), ui.HelpText("Shown under the title in feeds.")),
			ui.Field("set-timezone", "Timezone", ui.Select("timezone", zones...).Attrs(datastar.Bind("timezone"))),
		),
		html.Section(
			ui.SectionHeader("Help"),
			html.Div(markdown.Render(generalHelp, markdown.WithSize(ui.Sm), markdown.WithHeadingOffset(1), markdown.WithIDPrefix("help-general-"))).Class("mt-4"),
		),
		html.Section(
			ui.SectionHeader("Demo data"),
			html.P().Class("mt-4 text-sm text-neutral-600 dark:text-neutral-400").Text("Everything in this demo lives in server memory and is shared by every visitor."),
			html.Div(ui.Button(ui.Secondary, "Reset demo data", ui.WithLoading("resetting")).
				Attrs(datastar.OnClick(datastar.Post("/admin/reset")))).Class("mt-4"),
		),
	)
}

func settingsWritingForm(v settingsValues) htemel.FlowContent {
	authors := []html.SelectChild{}
	for _, a := range adminAuthors {
		authors = append(authors, ui.Choice(a, a).Selected(a == v.DefaultAuthor))
	}
	radios := []*html.LabelElement{}
	for _, s := range postStatuses {
		attrs := []htemel.Attribute{datastar.Bind("defaultStatus")}
		if s == v.DefaultStatus {
			attrs = append(attrs, htemel.Attr("checked", ""))
		}
		radios = append(radios, ui.Radio("defaultStatus", s, statusLabel(s), attrs...))
	}
	autosave := []htemel.Attribute{datastar.Bind("autosave")}
	if v.Autosave {
		autosave = append(autosave, htemel.Attr("checked", ""))
	}
	return ui.Stack(ui.Xl,
		adminSettingsForm(settingsWriting, map[string]any{"defaultAuthor": v.DefaultAuthor, "defaultStatus": v.DefaultStatus, "autosave": v.Autosave},
			ui.SectionHeader("Writing"),
			ui.Field("set-author", "Default author", ui.Select("defaultAuthor", authors...).Attrs(datastar.Bind("defaultAuthor"))),
			ui.RadioGroup("Status of new posts", radios...),
			ui.Switch("Autosave drafts while writing", autosave...),
		),
		html.Section(
			ui.SectionHeader("Help"),
			html.Div(markdown.Render(writingHelp, markdown.WithSize(ui.Sm), markdown.WithHeadingOffset(1), markdown.WithIDPrefix("help-writing-"))).Class("mt-4"),
		),
	)
}

// settingsAppearanceForm exposes the data-theme switch of the Theming page. The choice is the
// $theme signal that adminDocument applies to <html data-theme> on every admin page, and it is
// kept in localStorage so it survives navigation.
func settingsAppearanceForm() htemel.FlowContent {
	options := []html.SelectChild{ui.Choice("", "Default (indigo, zinc, md corners)")}
	for _, t := range demoThemes {
		options = append(options, ui.Choice(t.name, strings.ToUpper(t.name[:1])+t.name[1:]))
	}
	save := "try { localStorage.setItem('" + adminThemeKey + "', $theme) } catch (e) {}"
	return ui.Stack(ui.Xl,
		ui.Stack(ui.Lg,
			ui.SectionHeader("Appearance"),
			ui.Field("set-theme", "Theme", ui.Select("theme", options...).Attrs(datastar.Bind("theme"), datastar.OnChange(save)),
				ui.HelpText("Changes the color tokens and corners of every admin page. Remembered in this browser.")),
			ui.Inline(ui.Sm,
				ui.Button(ui.Primary, "Primary"), ui.Button(ui.Secondary, "Secondary"), ui.Button(ui.Outline, "Outline"),
				ui.Badge(ui.Brand, "Brand"), ui.Badge(ui.Success, "Success"), ui.Badge(ui.Neutral, "Neutral"),
			),
			ui.Alert(ui.Info, "Light and dark", htemel.Text("The color scheme toggle is in the sidebar footer; themes work in both.")),
		),
	)
}

// settingsSave is POST /admin/settings/{general,writing}.
func settingsSave(w http.ResponseWriter, r *http.Request) {
	tab := chi.URLParam(r, "tab")
	if tab != settingsGeneral && tab != settingsWriting {
		http.NotFound(w, r)
		return
	}
	var in settingsValues
	if err := dsse.ReadSignals(r, &in); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	sse := dsse.NewSSE(w, r)
	switch tab {
	case settingsGeneral:
		title := strings.TrimSpace(in.SiteTitle)
		if len([]rune(title)) < 2 {
			patchSignals(sse, map[string]any{"errors": map[string]any{"siteTitle": "Enter a site title of at least 2 characters."}})
			toast(sse, ui.Danger, "Settings not saved", "Fix the highlighted field.")
			return
		}
		adminSettingsData.set(func(v *settingsValues) {
			v.SiteTitle, v.Tagline, v.Timezone = title, strings.TrimSpace(in.Tagline), in.Timezone
		})
	case settingsWriting:
		adminSettingsData.set(func(v *settingsValues) {
			v.DefaultAuthor, v.DefaultStatus, v.Autosave = in.DefaultAuthor, in.DefaultStatus, in.Autosave
		})
	}
	patchSignals(sse, map[string]any{"errors": map[string]any{"siteTitle": ""}})
	toast(sse, ui.Success, "Settings saved", "Your changes were stored in server memory.")
}

// adminReset is POST /admin/reset.
func adminReset(w http.ResponseWriter, r *http.Request) {
	adminData.reset()
	adminNotes.reset()
	adminSettingsData.reset()
	sse := dsse.NewSSE(w, r)
	v := adminSettingsData.get()
	patchSignals(sse, map[string]any{"siteTitle": v.SiteTitle, "tagline": v.Tagline, "timezone": v.Timezone})
	toast(sse, ui.Info, "Demo data reset", "Posts, notes and settings are back to the seed.")
}
