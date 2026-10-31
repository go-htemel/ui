package main

// groups lists the showcase pages in sidebar order. Add an example by writing a function in an
// examples_*.go file and listing it here with ex.
var groups = []Group{
	{
		Slug: "buttons", Title: "Buttons",
		Intro: "Button, IconButton and ButtonGroup: six variants, five sizes, icon slots and a Datastar loading state.",
		Examples: []Example{
			ex("Variants", buttonVariants),
			ex("Sizes", buttonSizes),
			ex("Icons", buttonIcons),
			ex("Loading state (data-indicator)", buttonLoading),
			ex("Icon buttons", iconButtons),
			ex("Button group", buttonGroup),
		},
	},
	{
		Slug: "display", Title: "Display",
		Intro: "Badge, Avatar, Kbd, icons and dividers.",
		Examples: []Example{
			ex("Badges", badges),
			ex("Avatars", avatars),
			ex("Keyboard keys", keyboard),
			ex("Icons", icons),
			ex("Dividers", dividers),
		},
	},
	{
		Slug: "layout", Title: "Layout",
		Intro: "Stack, Inline and Container flex helpers, and Card with typed header, body and footer slots.",
		Examples: []Example{
			ex("Stack and Inline", stackAndInline),
			ex("Container", container),
			ex("Card", card),
		},
	},
	{
		Slug: "feedback", Title: "Feedback",
		Intro: "Alerts in five tones, a dismissible alert, and toasts pushed by the server over SSE.",
		Examples: []Example{
			ex("Alerts", alerts),
			ex("Dismissible alert", dismissibleAlert),
			ex("Server-pushed toasts", serverToast),
		},
	},
	{
		Slug: "forms", Title: "Forms",
		Intro: "Field with help and error text, Input, Textarea, Select, Checkbox, Radio and Switch, bound to signals with data-bind.",
		Examples: []Example{
			ex("Field with help text", fieldInput),
			ex("Live validation from the server", fieldValidation),
			ex("Server-rendered error", fieldServerError),
			ex("Textarea and Select", textareaAndSelect),
			ex("Checkbox, Radio and Switch", choices),
		},
	},
	{
		Slug: "overlays", Title: "Overlays",
		Intro: "Dialog is a native modal <dialog> and Dropdown a native popover: the browser handles the top layer, Escape, light dismiss and focus.",
		Examples: []Example{
			ex("Dialog", dialog),
			ex("Dropdown menu", dropdown),
		},
	},
	{
		Slug: "disclosure", Title: "Tabs & disclosure",
		Intro: "Tabs, Disclosure and Accordion, switched in the browser by signals with no server round-trip.",
		Examples: []Example{
			ex("Tabs", tabs),
			ex("Disclosure", disclosure),
			ex("Accordion", accordion),
		},
	},
	{
		Slug: "drag-and-drop", Title: "Drag and drop",
		Intro: "Draggable and DropZone: a kanban board whose drops post to the server, which patches the board over SSE.",
		Examples: []Example{
			ex("Kanban board", kanbanBoard),
		},
	},
}

func findGroup(slug string) (Group, bool) {
	for _, g := range groups {
		if g.Slug == slug {
			return g, true
		}
	}
	return Group{}, false
}
