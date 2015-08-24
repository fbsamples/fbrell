package h

import (
	"fmt"
	"io"
	"reflect"
	"strings"
)

type ReflectNode struct {
	Tag         string
	Node        interface{}
	SelfClosing bool
}

func (n *ReflectNode) HTML() (HTML, error) {
	return n, fmt.Errorf("Called HTML for ReflectNode: %+v", n)
}

// Write the generated markup for a ReflectNode.
func (n *ReflectNode) Write(w io.Writer) (int, error) {
	written := 0
	i := 0
	var err error

	i, err = fmt.Fprint(w, "<", n.Tag)
	written += i
	if err != nil {
		return written, err
	}

	i, err = n.writeAttributes(w)
	written += i
	if err != nil {
		return written, err
	}

	i, err = fmt.Fprint(w, ">")
	written += i
	if err != nil {
		return written, err
	}

	i, err = n.writeInner(w)
	written += i
	if err != nil {
		return written, err
	}

	if !n.SelfClosing {
		i, err = fmt.Fprint(w, "</", n.Tag, ">")
		written += i
		if err != nil {
			return written, err
		}
	}

	return written, nil
}

// Use reflection to write attributes.
func (n *ReflectNode) writeAttributes(w io.Writer) (int, error) {
	value := reflect.ValueOf(n.Node).Elem()
	typeOf := value.Type()
	written := 0
	tmp := 0
	var err error
	for i := 0; i < typeOf.NumField(); i++ {
		field := typeOf.Field(i)
		switch field.Tag.Get("h") {
		case "attr":
			tmp, err = writeKeyValue(
				w, strings.ToLower(field.Name), value.Field(i).Interface())
			written += tmp
			if err != nil {
				return written, err
			}
		case "dict":
			val := value.Field(i).Interface()
			rawAttrs, ok := val.(map[string]interface{})
			if !ok {
				return written, fmt.Errorf(
					"Invalid dict2 value: %+v of type %T", val, val)
			}
			attrs := Attributes(rawAttrs)
			tmp, err = attrs.Write(w, strings.ToLower(field.Name)+"-")
			written += tmp
			if err != nil {
				return written, err
			}
		}
	}
	return written, nil
}

// Use reflection to write inner HTML.
func (n *ReflectNode) writeInner(w io.Writer) (int, error) {
	value := reflect.ValueOf(n.Node).Elem()
	typeOf := value.Type()
	written := 0
	tmp := 0
	var err error
	for i := 0; i < typeOf.NumField(); i++ {
		field := typeOf.Field(i)
		if field.Tag.Get("h") != "inner" {
			continue
		}

		fieldValue := value.Field(i).Interface()
		if fieldValue == nil {
			continue
		}
		html, ok := fieldValue.(HTML)
		if !ok {
			return written, fmt.Errorf(
				"Field %s was marked as inner but does not satisfy the HTML interface",
				field.Name)
		}
		tmp, err = Write(w, html)
		written += tmp
		if err != nil {
			return written, err
		}
	}
	return written, nil
}
