package h

import "golang.org/x/net/context"

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

func (a *A) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "a", Node: a}, nil
}

type Body struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Inner HTML   `h:"inner"`
}

func (b *Body) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "body", Node: b}, nil
}

type Button struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Type  string                 `h:"attr"`
	Style string                 `h:"attr"`
	Title string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

func (b *Button) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "button", Node: b}, nil
}

type Div struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

func (d *Div) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "div", Node: d}, nil
}

type P struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

func (p *P) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "p", Node: p}, nil
}

const (
	Post = "post"
	Get  = "get"
)

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

func (f *Form) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "form", Node: f}, nil
}

type FieldSet struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Inner HTML   `h:"inner"`
}

func (f *FieldSet) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "fieldset", Node: f}, nil
}

type Legend struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Inner HTML   `h:"inner"`
}

func (l *Legend) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "legend", Node: l}, nil
}

type Head struct {
	Inner HTML `h:"inner"`
}

func (h *Head) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "head", Node: h}, nil
}

type H1 struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Title string `h:"attr"`
	Inner HTML   `h:"inner"`
}

func (h *H1) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "h1", Node: h}, nil
}

type H2 struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Title string `h:"attr"`
	Inner HTML   `h:"inner"`
}

func (h *H2) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "h2", Node: h}, nil
}

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

func (t *Iframe) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "iframe", Node: t}, nil
}

type I struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Inner HTML                   `h:"inner"`
	Data  map[string]interface{} `h:"dict"`
}

func (i *I) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "i", Node: i}, nil
}

type Img struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Src   string `h:"attr"`
	Alt   string `h:"attr"`
	Inner HTML   `h:"inner"`
}

func (t *Img) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{
		Tag:         "img",
		Node:        t,
		SelfClosing: true,
	}, nil
}

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

func (i *Input) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{
		Tag:         "input",
		Node:        i,
		SelfClosing: true,
	}, nil
}

type Label struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Title string `h:"attr"`
	For   string `h:"attr"`
	Inner HTML   `h:"inner"`
}

func (l *Label) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "label", Node: l}, nil
}

type Li struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Title string `h:"attr"`
	Inner HTML   `h:"inner"`
}

func (l *Li) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "li", Node: l}, nil
}

type Link struct {
	HREF string `h:"attr"`
	Type string `h:"attr"`
	Rel  string `h:"attr"`
}

func (l *Link) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{"link", l, true}, nil
}

type Meta struct {
	Charset  string `h:"attr"`
	Name     string `h:"attr"`
	Property string `h:"attr"`
	Content  string `h:"attr"`
}

func (m *Meta) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{"meta", m, true}, nil
}

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

func (o *Option) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "option", Node: o}, nil
}

type Pre struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Inner HTML   `h:"inner"`
}

func (d *Pre) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "pre", Node: d}, nil
}

type Script struct {
	Src   string `h:"attr"`
	Async bool   `h:"attr"`
	Inner HTML   `h:"inner"`
}

func (s *Script) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "script", Node: s}, nil
}

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

func (s *Select) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "select", Node: s}, nil
}

type Span struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Inner HTML   `h:"inner"`
}

func (s *Span) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "span", Node: s}, nil
}

type Strong struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Title string `h:"attr"`
	Inner HTML   `h:"inner"`
}

func (s *Strong) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "strong", Node: s}, nil
}

type Table struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Inner HTML   `h:"inner"`
}

func (t *Table) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "table", Node: t}, nil
}

type Tr struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

func (t *Tr) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "tr", Node: t}, nil
}

type Th struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Inner HTML   `h:"inner"`
}

func (t *Th) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "th", Node: t}, nil
}

type Td struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Inner HTML   `h:"inner"`
}

func (t *Td) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "td", Node: t}, nil
}

type Thead struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Inner HTML   `h:"inner"`
}

func (t *Thead) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "thead", Node: t}, nil
}

type Tbody struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Inner HTML   `h:"inner"`
}

func (t *Tbody) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "tbody", Node: t}, nil
}

type Textarea struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Name  string `h:"attr"`
	Inner HTML   `h:"inner"`
}

func (t *Textarea) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "textarea", Node: t}, nil
}

type Title Frag

func (t *Title) HTML(ctx context.Context) (HTML, error) {
	f := Frag(*t)
	return &Node{
		Tag:   "title",
		Inner: &f,
	}, nil
}

type Ul struct {
	ID    string `h:"attr"`
	Class string `h:"attr"`
	Style string `h:"attr"`
	Title string `h:"attr"`
	Inner HTML   `h:"inner"`
}

func (ul *Ul) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "ul", Node: ul}, nil
}

type Style struct {
	ID    string `h:"attr"`
	Inner HTML   `h:"inner"`
}

func (s *Style) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "style", Node: s}, nil
}

type Header struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

func (d *Header) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "header", Node: d}, nil
}

type Footer struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

func (d *Footer) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "footer", Node: d}, nil
}

type Main struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

func (d *Main) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "main", Node: d}, nil
}

type Nav struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

func (d *Nav) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "nav", Node: d}, nil
}

type Aside struct {
	ID    string                 `h:"attr"`
	Class string                 `h:"attr"`
	Style string                 `h:"attr"`
	Data  map[string]interface{} `h:"dict"`
	Inner HTML                   `h:"inner"`
}

func (d *Aside) HTML(ctx context.Context) (HTML, error) {
	return &ReflectNode{Tag: "aside", Node: d}, nil
}
