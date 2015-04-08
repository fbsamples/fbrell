// Package og implements the URL based Rell OG abstraction. It allows
// for a human readable and a base64 version API as input.
package og

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"sort"
	"strconv"
	"strings"

	"github.com/daaku/rell/internal/github.com/daaku/go.fburl"
	"github.com/daaku/rell/internal/github.com/daaku/go.static"
	"github.com/daaku/rell/internal/golang.org/x/net/context"
	"github.com/daaku/rell/rellenv"
)

// The representation of of <meta property="{key}" content="{value}">.
type Pair struct {
	Key   string
	Value string
}

// An ordered list of Pairs representing a raw Object.
type Object struct {
	Pairs        []Pair
	context      context.Context
	env          *rellenv.Env
	static       *static.Handler
	skipGenerate []string
}

// Padding is wasteful, but go wants it.
func fixPadding(b64 string) string {
	count := 4 - len(b64)%4
	if count > 0 && count < 4 {
		return b64 + strings.Repeat("=", count)
	}
	return b64
}

// Sorted query string encoding
func sortedEncode(v url.Values) string {
	if v == nil {
		return ""
	}
	parts := make([]string, 0, len(v)) // will be large enough for most uses
	for k, vs := range v {
		prefix := url.QueryEscape(k) + "="
		for _, v := range vs {
			parts = append(parts, prefix+url.QueryEscape(v))
		}
	}
	sort.Strings(parts)
	return strings.Join(parts, "&")
}

// Make a copy of url.Values.
func copyValues(source url.Values) url.Values {
	dest := url.Values{}
	for key, _ := range source {
		switch key {
		case "action_object_map":
		case "action_ref_map":
		case "action_type_map":
		case "fb_action_ids":
		case "fb_action_types":
		case "fb_aggregation_id":
		case "fb_locale":
		case "fb_source":
		case "ref":
		case "refid":
			continue
		default:
			dup := make([]string, len(source[key]))
			copy(dup, source[key])
			dest[key] = dup
		}
	}
	return dest
}

type Parser struct {
	Static *static.Handler
}

// Create a new Object from Base64 JSON encoded data.
func (p *Parser) FromBase64(ctx context.Context, env *rellenv.Env, b64 string) (*Object, error) {
	jsonBytes, err := base64.URLEncoding.DecodeString(fixPadding(b64))
	if err != nil {
		return nil, fmt.Errorf(
			"Failed base64 decode of string \"%s\" with error: %s", b64, err)
	}
	var strSlices [][]interface{}
	err = json.Unmarshal(jsonBytes, &strSlices)
	if err != nil {
		return nil, fmt.Errorf(
			"Failed json unmarshal string %s with error %s", string(jsonBytes), err)
	}

	object := &Object{
		context: ctx,
		env:     env,
		static:  p.Static,
	}
	for _, row := range strSlices {
		if len(row) != 2 {
			return nil, fmt.Errorf("Got more than two elements in pair: %v", row)
		}
		if row[0] == nil {
			return nil, fmt.Errorf("First element in pair is null: %v", row)
		}
		key := fmt.Sprint(row[0])
		val := ""
		switch t := row[1].(type) {
		case nil:
			object.skipGenerate = append(object.skipGenerate, key)
			continue
		case float64:
			val = fmt.Sprint(uint64(t))
		default:
			val = fmt.Sprint(t)
		}
		object.AddPair(key, val)
	}

	if object.shouldGenerate("og:url") {
		url := env.AbsoluteURL("/rog/" + b64).String()
		object.AddPair("og:url", url)
	}

	err = object.generateDefaults()
	if err != nil {
		return nil, err
	}
	return object, nil
}

// Create a new Object from query string data.
func (p *Parser) FromValues(ctx context.Context, env *rellenv.Env, values url.Values) (*Object, error) {
	object := &Object{
		context: ctx,
		env:     env,
		static:  p.Static,
	}
	for key, values := range values {
		if strings.Contains(key, ":") {
			for _, value := range values {
				object.AddPair(key, value)
			}
		}
	}

	if object.shouldGenerate("og:url") {
		copiedValues := copyValues(values)
		copiedValues.Del("og:type")
		copiedValues.Del("og:title")
		url := url.URL{
			Scheme:   env.Scheme,
			Host:     env.Host,
			Path:     "/og/" + object.Type() + "/" + object.Title(),
			RawQuery: sortedEncode(copiedValues),
		}
		object.AddPair("og:url", url.String())
	}

	ogType := object.Type()
	isGlobalOGType := !strings.Contains(ogType, ":")
	isOwnedOGType := strings.HasPrefix(ogType, rellenv.FbApp(ctx).Namespace()+":")
	if object.shouldGenerate("fb:app_id") && (isGlobalOGType || isOwnedOGType) {
		object.AddPair("fb:app_id", strconv.FormatUint(rellenv.FbApp(ctx).ID(), 10))
	}

	err := object.generateDefaults()
	if err != nil {
		return nil, err
	}
	return object, nil
}

func (o *Object) shouldGenerate(key string) bool {
	if len(o.GetAll(key)) > 0 {
		return false
	}
	for _, skipKey := range o.skipGenerate {
		if key == skipKey {
			return false
		}
	}
	return true
}

func (o *Object) generateDefaults() error {
	url := o.URL()
	if o.shouldGenerate("og:image") {
		img, err := o.static.URL("/images/" + hashedPick(url, stockImages))
		if err != nil {
			return err
		}
		o.AddPair("og:image", o.env.AbsoluteURL(img).String())
	}
	if o.shouldGenerate("og:description") {
		o.AddPair("og:description", hashedPick(url, stockDescriptions))
	}
	return nil
}

// Get the first "og:type" value.
func (o *Object) Type() string {
	return o.Get("og:type")
}

// Get the first "og:title" value.
func (o *Object) Title() string {
	return o.Get("og:title")
}

// Get the first "og:url" value.
func (o *Object) URL() string {
	return o.Get("og:url")
}

// Get the first "og:description" value.
func (o *Object) Description() string {
	return o.Get("og:description")
}

// Get the first "og:image" value.
func (o *Object) ImageURL() string {
	return o.Get("og:image")
}

// Get the first "fb:app_id" value.
func (o *Object) AppID() string {
	return o.Get("fb:app_id")
}

// Get a URL to the Lint tool for the Object.
func (o *Object) LintURL() string {
	values := url.Values{}
	values.Set("q", o.URL())
	u := &fburl.URL{
		Scheme:    o.env.Scheme,
		Env:       rellenv.FbEnv(o.context),
		SubDomain: fburl.DDevelopers,
		Path:      "/tools/debug/og/object",
		Values:    values,
	}
	return u.String()
}

// Get a URL to the Like button for the Object.
func (o *Object) LikeURL() string {
	values := url.Values{}
	values.Set("href", o.URL())
	u := &fburl.URL{
		Scheme:    o.env.Scheme,
		Env:       rellenv.FbEnv(o.context),
		SubDomain: fburl.DWww,
		Path:      "/plugins/like",
		Values:    values,
	}
	return u.String()
}

// Get the first value associated with the key.
func (o *Object) Get(key string) string {
	for _, pair := range o.Pairs {
		if pair.Key == key {
			return pair.Value
		}
	}
	return ""
}

// Get all the values associated with the key.
func (o *Object) GetAll(key string) []string {
	results := make([]string, 0)
	for _, pair := range o.Pairs {
		if pair.Key == key {
			results = append(results, pair.Value)
		}
	}
	return results
}

// Add a new Pair to the end of the current list.
func (o *Object) AddPair(key, value string) {
	o.Pairs = append(o.Pairs, Pair{Key: key, Value: value})
}

// Pick an string from the given choices based on a consistent hash of
// the given URL. This allows for "persistant defaults" for fields.
func hashedPick(rawurl string, choices []string) string {
	var key string
	url, err := url.Parse(rawurl)
	if err != nil {
		log.Printf("Failed to parse URL %s in hashed pick: %s", url, err)
		key = ""
	} else {
		key = url.Path
		// TODO figure out if fixing this will break things before removing it
		if url.RawQuery == "" {
			key += "undefined"
		} else {
			key += url.RawQuery
		}
	}
	md5 := md5.New()
	io.WriteString(md5, key)
	hex := fmt.Sprintf("%x", md5.Sum(nil))
	pick, err := strconv.ParseUint(hex[:8], 16, 64)
	if err != nil {
		log.Printf("Failed to parse hex \"%s\" with error: %s", hex[:8], err)
	}
	index := pick % uint64(len(choices))
	return choices[index]
}

var stockDescriptions = []string{
	"You might have seen a housefly, maybe even a super-fly, but I bet you" +
		" ain't never seen a donkey fly!",
	"If I'm not back in five minutes... just wait longer.",
	"I keep forgetting about the goddamn tiger!",
	"I refuse to play your Chinese food mind games!",
	"Everybody remember where we parked.",
	"Oh, my, yes.",
	"Hello there, children.",
	"Yeah, I eat the whole apple. The core, stem, seeds, everything.",
}

var stockImages = []string{
	"beach_skyseeker_3184914.jpg",
	"beetle_gnilenkov_4647458067.jpg",
	"car_damianmorysfotos_5933730674.jpg",
	"circuits_ladyada_5074936971.jpg",
	"dogs_mythicseabass_4662963501.jpg",
	"flower_serrasclimb_3999125500.jpg",
	"jailed_flower_vpolat_3069134052.jpg",
	"stone_house_aamaianos_3040806369.jpg",
	"taxi_rotia_2806339125.jpg",
	"valley_markgee6_90348619.jpg",
}
