package xml2json_test

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/ashep/goxml2json"
)

func TestXMLErrEOF(t *testing.T) {
	_, err := xml2json.Convert([]byte(`
<foo>
`))

	assert.EqualError(t, err, "XML syntax error on line 3: unexpected EOF")
}

func TestXMLErrElementExpected(t *testing.T) {
	_, err := xml2json.Convert([]byte(`
<foo></
`))

	assert.EqualError(t, err, "XML syntax error on line 2: expected element name after </")
}

func TestXMLErrInvalidElement(t *testing.T) {
	_, err := xml2json.Convert([]byte(`
<foo></bar>
`))

	assert.EqualError(t, err, "XML syntax error on line 2: element <foo> closed by </bar>")
}

func TestTypes(t *testing.T) {
	d, err := xml2json.Convert([]byte(`
<int>42</int>
<spacedInt>  42   </spacedInt>
<intAsString>"42"</intAsString>

<float>3.14</float>
<spacedFloat>  3.14   </spacedFloat>
<floatAsString>"3.14"</floatAsString>

<string>Foo Bar</string>
<spacedString>  Foo Bar  </spacedString>
<quotedString>"Foo Bar"</quotedString>
<quotedSpacedString>  "  Foo Bar  "  </quotedSpacedString>
`))

	require.NoError(t, err)

	v := struct {
		Int                int     `json:"int"`
		SpacedInt          int     `json:"spacedInt"`
		IntAsString        string  `json:"intAsString"`
		Float              float64 `json:"float"`
		SpacedFloat        float64 `json:"spacedFloat"`
		FloatAsString      string  `json:"floatAsString"`
		String             string  `json:"string"`
		SpacedString       string  `json:"spacedString"`
		QuotedString       string  `json:"quotedString"`
		QuotedSpacedString string  `json:"quotedSpacedString"`
	}{}
	require.NoError(t, json.Unmarshal(d, &v))

	assert.Equal(t, 42, v.Int)
	assert.Equal(t, 42, v.SpacedInt)
	assert.Equal(t, "42", v.IntAsString)

	assert.Equal(t, 3.14, v.Float)
	assert.Equal(t, 3.14, v.SpacedFloat)
	assert.Equal(t, "3.14", v.FloatAsString)

	assert.Equal(t, "Foo Bar", v.String)
	assert.Equal(t, "Foo Bar", v.SpacedString)
	assert.Equal(t, "Foo Bar", v.QuotedString)
	assert.Equal(t, "  Foo Bar  ", v.QuotedSpacedString)
}

func TestUnexpectedCharData(t *testing.T) {
	d, err := xml2json.Convert([]byte(`abc<foo>bar</foo>def`))

	require.NoError(t, err)
	assert.Equal(t, []byte(`{"foo":"bar"}`), d)
}

func TestStructure(t *testing.T) {
	d, err := xml2json.Convert([]byte(`
<foo>
	<one>1</one>
	<two>2</two>
</foo>
<foo>
	<one>3</one>
	<two>4</two>
</foo>
<bar>
	<one>5</one>
	<two>6</two>
	<foo>
		<one>7</one>
		<two>8</two>
		<baz>
			<one>9</one>
			<two>10</two>
		</baz>
	</foo>
</bar>
<baz>
	<empty1></empty1>
	<empty2>""</empty2>
	<empty3>" "</empty3>
</baz>
`))

	require.NoError(t, err)
	assert.Equal(t, []byte(`{"bar":{"foo":{"baz":{"one":9,"two":10},"one":7,"two":8},"one":5,"two":6},"baz":{"empty1":"","empty2":"","empty3":" "},"foo":[{"one":1,"two":2},{"one":3,"two":4}]}`), d)
}

func TestStructureNamespaced(t *testing.T) {
	d, err := xml2json.Convert([]byte(`
<NS1:foo xmlns:NS1="https://foo.com"><NS1:bar>baz</NS1:bar></NS1:foo>
`))

	require.NoError(t, err)
	assert.Equal(t, []byte(`{"foo":{"bar":"baz"}}`), d)
}
