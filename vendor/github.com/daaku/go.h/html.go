package h

import "golang.org/x/net/context"

// XMLNS auto prefixes each attribute key with "xmlns:" in the output.
type XMLNS map[string]string

// Attributes returns the attributes to render.
func (ns XMLNS) Attributes() Attributes {
	attrs := Attributes{}
	for key, val := range ns {
		attrs["xmlns:"+key] = val
	}
	return attrs
}

var _ HTML = (*Document)(nil)

// Document renders a HTML <document> tag.
type Document struct {
	XMLNS XMLNS
	Inner HTML
	ID    string
	Lang  string
}

// HTML renders the content.
func (d *Document) HTML(ctx context.Context) (HTML, error) {
	attrs := d.XMLNS.Attributes()
	if d.ID != "" {
		attrs["id"] = d.ID
	}
	if d.Lang != "" {
		attrs["lang"] = "en"
	}

	return &Frag{
		Unsafe("<!doctype html>"),
		&Node{
			Tag:        "html",
			Attributes: attrs,
			Inner:      d.Inner,
		},
	}, nil
}

var _ HTML = (*A)(nil)

// A renders a HTML <a> tag.
type A struct {
	ID     string                 `h:"attr"`
	Class  string                 `h:"attr"`
	Style  string                 `h:"attr"`
	Title  string                 `h:"attr"`
	HREF   string                 `h:"attr"`
	Target string                 `h:"attr"`
	Rel    string                 `h:"attr"`
	Inner  HTML                   `h:"inner"`
	Data   map[string]interface{} `h:"dict"`
}

// HTML renders the content.
func (a *A) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "a", Node: a}, nil
}

var _ HTML = (*Body)(nil)

// Body renders a HTML <body> tag.
type Body struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (b *Body) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "body", Node: b}, nil
}

var _ HTML = (*Button)(nil)

// Button renders a HTML <button> tag.
type Button struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Type  string                 `h:"attr"`
	Style string                 `h:"attr"`
	Title string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

// HTML renders the content.
func (b *Button) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "button", Node: b}, nil
}

var _ HTML = (*Div)(nil)

// Div renders a HTML <div> tag.
type Div struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

// HTML renders the content.
func (d *Div) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "div", Node: d}, nil
}

var _ HTML = (*P)(nil)

// P renders a HTML <p> tag.
type P struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

// HTML renders the content.
func (p *P) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "p", Node: p}, nil
}

// Method names for use with Forms.
const (
	Post = "post"
	Get  = "get"
)

var _ HTML = (*Form)(nil)

// Form renders a HTML <form> tag.
type Form struct {
	ID      string `h:"attr"`
	Class   string `h:"attr"`
	Style   string `h:"attr"`
	Action  string `h:"attr"`
	Method  string `h:"attr"`
	EncType string `h:"attr"`
	Target  string `h:"attr"`
	Title   string `h:"attr"`
	Inner   HTML   `h:"inner"`
}

// HTML renders the content.
func (f *Form) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "form", Node: f}, nil
}

var _ HTML = (*FieldSet)(nil)

// FieldSet renders a HTML <fieldset> tag.
type FieldSet struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (f *FieldSet) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "fieldset", Node: f}, nil
}

var _ HTML = (*Legend)(nil)

// Legend renders a HTML <legend> tag.
type Legend struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (l *Legend) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "legend", Node: l}, nil
}

var _ HTML = (*Head)(nil)

// Head renders a HTML <head> tag.
type Head struct {
	Inner HTML `h:"inner"`
}

// HTML renders the content.
func (h *Head) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "head", Node: h}, nil
}

var _ HTML = (*H1)(nil)

// H1 renders a HTML <h1> tag.
type H1 struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Title string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (h *H1) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "h1", Node: h}, nil
}

var _ HTML = (*H2)(nil)

// H2 renders a HTML <h2> tag.
type H2 struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Title string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (h *H2) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "h2", Node: h}, nil
}

var _ HTML = (*H3)(nil)

// H3 renders a HTML <h3> tag.
type H3 struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Title string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (h *H3) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "h3", Node: h}, nil
}

var _ HTML = (*H4)(nil)

// H4 renders a HTML <h4> tag.
type H4 struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Title string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (h *H4) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "h4", Node: h}, nil
}

var _ HTML = (*H5)(nil)

// H5 renders a HTML <h5> tag.
type H5 struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Title string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (h *H5) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "h5", Node: h}, nil
}

var _ HTML = (*H6)(nil)

// H6 renders a HTML <h6> tag.
type H6 struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Title string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (h *H6) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "h6", Node: h}, nil
}

var _ HTML = (*Iframe)(nil)

// Iframe renders a HTML <iframe> tag.
type Iframe struct {
	ID              string `h:"attr"`
	Class           string `h:"attr"`
	Style           string `h:"attr"`
	Src             string `h:"attr"`
	Width           int    `h:"attr"`
	Height          int    `h:"attr"`
	FrameBorder     int    `h:"attr"`
	AllowFullScreen bool   `h:"attr"`
}

// HTML renders the content.
func (t *Iframe) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "iframe", Node: t}, nil
}

var _ HTML = (*I)(nil)

// I renders a HTML <i> tag.
type I struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Inner HTML                   `h:"inner"`
	Data  map[string]interface{} `h:"dict"`
}

// HTML renders the content.
func (i *I) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "i", Node: i}, nil
}

var _ HTML = (*Img)(nil)

// Img renders a HTML <img> tag.
type Img struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Src   string                 `h:"attr"`
	Alt   string                 `h:"attr"`
	Inner HTML                   `h:"inner"`
	Data  map[string]interface{} `h:"dict"`
}

// HTML renders the content.
func (t *Img) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{
		Tag:         "img",
		Node:        t,
		SelfClosing: true,
	}, nil
}

var _ HTML = (*Input)(nil)

// Input renders a HTML <input> tag.
type Input struct {
	ID          string                 `h:"attr"`
	Class       string                 `h:"attr"`
	Name        string                 `h:"attr"`
	Style       string                 `h:"attr"`
	Type        string                 `h:"attr"`
	Value       string                 `h:"attr"`
	Placeholder string                 `h:"attr"`
	Checked     bool                   `h:"attr"`
	Multiple    bool                   `h:"attr"`
	Data        map[string]interface{} `h:"dict"`
	Inner       HTML                   `h:"inner"`
}

// HTML renders the content.
func (i *Input) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{
		Tag:         "input",
		Node:        i,
		SelfClosing: true,
	}, nil
}

var _ HTML = (*Label)(nil)

// Label renders a HTML <label> tag.
type Label struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Title string `h:"attr"`
	For   string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (l *Label) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "label", Node: l}, nil
}

var _ HTML = (*Li)(nil)

// Li renders a HTML <li> tag.
type Li struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Title string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (l *Li) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "li", Node: l}, nil
}

var _ HTML = (*Link)(nil)

// Link renders a HTML <link> tag.
type Link struct {
	HREF string `h:"attr"`
	Type string `h:"attr"`
	Rel  string `h:"attr"`
}

// HTML renders the content.
func (l *Link) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{"link", l, true}, nil
}

var _ HTML = (*Meta)(nil)

// Meta renders a HTML <meta> tag.
type Meta struct {
	Charset  string `h:"attr"`
	Name     string `h:"attr"`
	Property string `h:"attr"`
	Content  string `h:"attr"`
}

// HTML renders the content.
func (m *Meta) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{"meta", m, true}, nil
}

var _ HTML = (*Option)(nil)

// Option renders a HTML <option> tag.
type Option struct {
	ID       string                 `h:"attr"`
	Class    string                 `h:"attr"`
	Style    string                 `h:"attr"`
	Title    string                 `h:"attr"`
	Value    string                 `h:"attr"`
	Selected bool                   `h:"attr"`
	Data     map[string]interface{} `h:"dict"`
	Inner    HTML                   `h:"inner"`
}

// HTML renders the content.
func (o *Option) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "option", Node: o}, nil
}

var _ HTML = (*Pre)(nil)

// Pre renders a HTML <pre> tag.
type Pre struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (d *Pre) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "pre", Node: d}, nil
}

var _ HTML = (*Script)(nil)

// Script renders a HTML <script> tag.
type Script struct {
	Src   string `h:"attr"`
	Async bool   `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (s *Script) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "script", Node: s}, nil
}

var _ HTML = (*Select)(nil)

// Select renders a HTML <select> tag.
type Select struct {
	ID       string `h:"attr"`
	Class    string `h:"attr"`
	Name     string `h:"attr"`
	Style    string `h:"attr"`
	Title    string `h:"attr"`
	Multiple bool   `h:"attr"`
	Size     int    `h:"size"`
	Inner    HTML   `h:"inner"`
}

// HTML renders the content.
func (s *Select) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "select", Node: s}, nil
}

var _ HTML = (*Span)(nil)

// Span renders a HTML <span> tag.
type Span struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (s *Span) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "span", Node: s}, nil
}

var _ HTML = (*Strong)(nil)

// Strong renders a HTML <strong> tag.
type Strong struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Title string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (s *Strong) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "strong", Node: s}, nil
}

var _ HTML = (*Table)(nil)

// Table renders a HTML <table> tag.
type Table struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (t *Table) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "table", Node: t}, nil
}

var _ HTML = (*Tr)(nil)

// Tr renders a HTML <tr> tag.
type Tr struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

// HTML renders the content.
func (t *Tr) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "tr", Node: t}, nil
}

var _ HTML = (*Th)(nil)

// Th renders a HTML <th> tag.
type Th struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (t *Th) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "th", Node: t}, nil
}

var _ HTML = (*Td)(nil)

// Td renders a HTML <td> tag.
type Td struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (t *Td) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "td", Node: t}, nil
}

var _ HTML = (*Thead)(nil)

// Thead renders a HTML <thead> tag.
type Thead struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (t *Thead) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "thead", Node: t}, nil
}

var _ HTML = (*Tbody)(nil)

// Tbody renders a HTML <tbody> tag.
type Tbody struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (t *Tbody) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "tbody", Node: t}, nil
}

var _ HTML = (*Textarea)(nil)

// Textarea renders a HTML <textarea> tag.
type Textarea struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Name  string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (t *Textarea) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "textarea", Node: t}, nil
}

// Title renders a HTML <title> tag.
type Title Frag

// HTML renders the content.
func (t *Title) HTML(ctx context.Context) (HTML, error) {
	f := Frag(*t)
	return &Node{
		Tag:   "title",
		Inner: &f,
	}, nil
}

var _ HTML = (*Ul)(nil)

// Ul renders a HTML <ul> tag.
type Ul struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Title string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (ul *Ul) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "ul", Node: ul}, nil
}

var _ HTML = (*Style)(nil)

// Style renders a HTML <style> tag.
type Style struct {
	ID    string `h:"attr"`
	Inner HTML   `h:"inner"`
}

// HTML renders the content.
func (s *Style) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "style", Node: s}, nil
}

var _ HTML = (*Header)(nil)

// Header renders a HTML <header> tag.
type Header struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

// HTML renders the content.
func (d *Header) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "header", Node: d}, nil
}

var _ HTML = (*Footer)(nil)

// Footer renders a HTML <footer> tag.
type Footer struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

// HTML renders the content.
func (d *Footer) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "footer", Node: d}, nil
}

var _ HTML = (*Main)(nil)

// Main renders a HTML <main> tag.
type Main struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

// HTML renders the content.
func (d *Main) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "main", Node: d}, nil
}

var _ HTML = (*Nav)(nil)

// Nav renders a HTML <nav> tag.
type Nav struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

// HTML renders the content.
func (d *Nav) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "nav", Node: d}, nil
}

var _ HTML = (*Aside)(nil)

// Aside renders a HTML <aside> tag.
type Aside struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

// HTML renders the content.
func (d *Aside) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "aside", Node: d}, nil
}

var _ HTML = (*Small)(nil)

// Small renders a HTML <small> tag.
type Small struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

// HTML renders the content.
func (d *Small) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "small", Node: d}, nil
}

var _ HTML = (*Figure)(nil)

// Figure renders a HTML <figure> tag.
type Figure struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

// HTML renders the content.
func (p *Figure) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "figure", Node: p}, nil
}

var _ HTML = (*FigCaption)(nil)

// FigCaption renders a HTML <figcaption> tag.
type FigCaption struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

// HTML renders the content.
func (p *FigCaption) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "figcaption", Node: p}, nil
}
