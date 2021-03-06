package graph

import (
	"crypto/sha256"
	"fmt"
	"io"
	"strings"

	"github.com/storj/archview/arch"
)

// Dot implements .dot encoding.
type Dot struct {
	World *arch.World

	Options
}

// WriteTo writes dot output to w.
func (ctx *Dot) WriteTo(w io.Writer) (n int64, err error) {
	write := func(format string, args ...interface{}) bool {
		if err != nil {
			return false
		}
		var wrote int
		wrote, err = fmt.Fprintf(w, format, args...)
		n += int64(wrote)
		return err == nil
	}

	write("digraph G {\n")
	write("\trankdir=LR;\n")
	write("\tranksep=3;\n")

	if ctx.NoColor {
		write("\tnode [width=3 shape=record target=\"_graphviz\"];\n")
		write("\tedge [];\n")
	} else {
		write("\tnode [penwidth=2 width=3 shape=record target=\"_graphviz\" style=filled fillcolor=white];\n")
		write("\tedge [penwidth=2];\n")
	}

	write("\n")
	defer write("}\n")

	if ctx.Clustering == ClusterByClass {
		byClass := map[string][]*arch.Component{}
		for _, component := range ctx.World.Components {
			byClass[component.Class] = append(byClass[component.Class], component)
		}

		for class, components := range byClass {
			write("\tsubgraph cluster_%v {\n", sanitize(class))
			write("\t\tlabel=%q;\n\n", class)
			write("\t\tbgcolor=gray98; pencolor=gray80; fontsize=10;\n\n")
			for _, component := range components {
				write("\t\t%s %v;\n", ctx.id(component),
					attrs(
						ctx.label(component),
						ctx.href(component),
						ctx.color(component),
						ctx.nodetooltip(component),
					))
			}
			write("\t}\n")
		}
	} else {
		for _, component := range ctx.World.Components {
			write("\t%s %v;\n", ctx.id(component),
				attrs(
					ctx.label(component),
					ctx.href(component),
					ctx.color(component),
				))
		}
	}

	write("\n")

	for _, source := range ctx.World.Components {
		for _, link := range source.Links {
			write("\t%s -> %s %v;\n", ctx.id(source), ctx.id(link.Target),
				attrs(
					ctx.color(link.Target),
					ctx.edgetooltip(source, link),
					ctx.linkStyle(link),
				))
		}
		if len(source.Links) > 0 {
			write("\n")
		}
	}
	return n, err
}

func attrs(list ...string) string {
	xs := list[:0]
	for _, x := range list {
		if x != "" {
			xs = append(xs, x)
		}
	}
	if len(xs) == 0 {
		return ""
	}
	return "[" + strings.Join(xs, ",") + "]"
}

func (ctx *Dot) linkStyle(link *arch.Link) string {
	if link.Implementation {
		return "style=dashed"
	}
	return ""
}

func (ctx *Dot) id(component *arch.Component) string {
	return strings.Map(func(r rune) rune {
		switch {
		case 'a' <= r && r <= 'z':
			return r
		case 'A' <= r && r <= 'Z':
			return r
		case '0' <= r && r <= '9':
			return r
		default:
			return '_'
		}
	}, component.Name())
}

func (ctx *Dot) label(component *arch.Component) string {
	return fmt.Sprintf("label=%q", strings.TrimPrefix(component.Name(), ctx.TrimPrefix))
}

func (ctx *Dot) nodetooltip(component *arch.Component) string {
	return fmt.Sprintf("tooltip=%q", component.Comment)
}

func (ctx *Dot) edgetooltip(source *arch.Component, link *arch.Link) string {
	return fmt.Sprintf("tooltip=%q", link.Path)
}

func (ctx *Dot) href(component *arch.Component) string {
	return fmt.Sprintf("href=%q", "http://godoc.org/"+component.Package()+"#"+component.ShortName())
}

func (ctx *Dot) color(component *arch.Component) string {
	if ctx.NoColor {
		return ""
	}

	hash := sha256.Sum256([]byte(component.Name()))
	hue := float64(uint(hash[0])<<8|uint(hash[1])) / 0xFFFF
	return "color=" + hslahex(hue, 0.9, 0.3, 0.7)
}
