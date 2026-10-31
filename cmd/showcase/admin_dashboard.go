package main

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-htemel/htemel"
	"github.com/go-htemel/htemel/html"

	"github.com/go-htemel/ui"
)

// dashboardPage is GET /admin: the numbers, what changed lately, and the storage meter.
func dashboardPage(r *http.Request) htemel.Node {
	c := adminData.counts()
	delta := ""
	if d := c.PublishedThisMonth - c.PublishedLastMonth; d != 0 {
		delta = fmt.Sprintf("%+d vs last month", d)
	}
	stats := ui.StatGrid(
		ui.Stat("Posts", strconv.Itoa(c.Total), fmt.Sprintf("%d scheduled", c.Scheduled)),
		ui.Stat("Published this month", strconv.Itoa(c.PublishedThisMonth), delta),
		ui.Stat("Drafts", strconv.Itoa(c.Drafts)),
		ui.Stat("Comments", strconv.Itoa(c.Comments)),
	)

	body := ui.TableBody()
	for _, p := range adminData.recent(6) {
		event, tone := "Edited", ui.Neutral
		switch p.Status {
		case statusPublished:
			event, tone = "Published", ui.Success
		case statusScheduled:
			event, tone = "Scheduled", ui.Brand
		}
		body.Children(html.Tr(
			ui.Td(html.A().Href("/admin/posts/"+strconv.Itoa(p.ID)+"/edit").
				Class("font-medium text-neutral-900 hover:text-primary-700 focus-visible:rounded-xs focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-primary-600 dark:text-white dark:hover:text-primary-300 dark:focus-visible:outline-primary-500").
				Text(p.Title)),
			ui.Td(ui.Badge(tone, event)),
			ui.Td(htemel.Text(p.Author)),
			ui.Td(adminTime(p.UpdatedAt)),
		))
	}
	activity := html.Section(
		ui.SectionHeader("Recent activity", linkButton(false, "All posts", "/admin/posts", nil)),
		html.Div(ui.Table(
			ui.TableHead(ui.Th(htemel.Text("Post")), ui.Th(htemel.Text("Event")), ui.Th(htemel.Text("Author")), ui.Th(htemel.Text("When"))),
			body, ui.TableCaption("Recent activity"),
		)).Class("mt-4"),
	)

	side := ui.Stack(ui.Lg,
		ui.Card(
			ui.CardHeader("Storage"),
			ui.CardBody(
				ui.Progress(62, 100, ui.ProgressLabel("Storage used"), ui.ProgressTone(ui.Warning)),
				html.P().Class("mt-3 text-sm text-neutral-600 dark:text-neutral-400").Text("3.1 GB of 5 GB used. Media is the biggest share."),
			),
		),
		ui.Card(
			ui.CardHeader("Up next"),
			ui.CardBody(htemel.Group(
				html.P().Text(fmt.Sprintf("%d scheduled and %d draft posts are waiting.", c.Scheduled, c.Drafts)),
				html.Div(linkButton(true, "New post", "/admin/posts/new", ui.IconPlus())).Class("mt-4"),
			)),
		),
	)

	return adminPage("Dashboard", navDashboard, false, r.URL.Query().Get("notice"),
		adminBody(
			ui.PageHeader("Dashboard", ui.WithDescription("How the site is doing.")),
			stats,
			html.Div(activity, side).Class("grid gap-8 @3xl:grid-cols-[minmax(0,2fr)_minmax(0,1fr)]"),
		))
}

// adminTime renders t as a <time> element in a short, absolute form ("Oct 29, 14:20").
func adminTime(t time.Time) htemel.FlowContent {
	return html.Time().Datetime(t.Format("2006-01-02T15:04Z07:00")).Text(t.Format("Jan 2, 15:04"))
}
