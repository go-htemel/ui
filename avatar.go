package ui

import "github.com/go-htemel/htemel/html"

var avatarSizes = map[Size]string{
	Xs: "size-6 text-[0.625rem]",
	Sm: "size-8 text-xs",
	Md: "size-10 text-sm",
	Lg: "size-12 text-base",
	Xl: "size-14 text-lg",
}

const avatarBase = "inline-block shrink-0 rounded-full ring-2 ring-white dark:ring-zinc-900"

// Avatar returns a round profile image, <img>. alt describes the person ("Ada Lovelace").
//
// Options: WithSize.
func Avatar(src, alt string, opts ...Option) *html.ImgElement {
	o := collect(opts)
	return html.Img().Src(src).Alt(alt).
		Class(Classes(avatarBase, "bg-zinc-100 object-cover dark:bg-zinc-800", pick(avatarSizes, o.size, Md)))
}

// AvatarInitials returns a round placeholder with initials, for people without a picture. name
// is announced in place of the initials.
//
// Options: WithSize.
func AvatarInitials(initials, name string, opts ...Option) *html.SpanElement {
	o := collect(opts)
	return html.Span().
		Class(Classes(avatarBase, "inline-flex items-center justify-center bg-zinc-600 font-medium text-white",
			"dark:bg-zinc-500", pick(avatarSizes, o.size, Md))).
		Role("img").
		Aria("label", name).
		Text(initials)
}
