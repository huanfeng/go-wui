package res

import (
	"testing"
	"testing/fstest"
)

func TestInflateLinearLayout(t *testing.T) {
	inflater := NewLayoutInflater(nil)
	RegisterBuiltinViews(inflater)

	xmlStr := `<LinearLayout width="match_parent" height="match_parent" orientation="vertical" padding="16dp">
		<TextView id="title" width="match_parent" height="wrap_content" text="Hello" textSize="18dp" />
		<Button id="btn" width="match_parent" height="48dp" text="Click" />
	</LinearLayout>`

	root, err := inflater.InflateFromString(xmlStr)
	if err != nil {
		t.Fatal(err)
	}
	if root.Tag() != "LinearLayout" {
		t.Error("root should be LinearLayout")
	}
	if len(root.Children()) != 2 {
		t.Fatalf("expected 2 children, got %d", len(root.Children()))
	}
	if root.Children()[0].Id() != "title" {
		t.Error("first child id should be title")
	}
	if root.Children()[1].Id() != "btn" {
		t.Error("second child id should be btn")
	}
}

func TestInflateResourceRef(t *testing.T) {
	f := fstest.MapFS{
		"values/strings.xml": &fstest.MapFile{Data: []byte(`<resources><string name="hello">Hello World</string></resources>`)},
	}
	rm := NewResourceManager(f)
	inflater := NewLayoutInflater(rm)
	RegisterBuiltinViews(inflater)

	root, err := inflater.InflateFromString(`<TextView width="wrap_content" height="wrap_content" text="@string/hello" />`)
	if err != nil {
		t.Fatal(err)
	}
	if root.GetDataString("text") != "Hello World" {
		t.Errorf("resource ref not resolved, got %q", root.GetDataString("text"))
	}
}

func TestInflateNestedLayout(t *testing.T) {
	inflater := NewLayoutInflater(nil)
	RegisterBuiltinViews(inflater)

	xmlStr := `<LinearLayout orientation="vertical">
		<LinearLayout orientation="horizontal">
			<Button id="a" text="A" />
			<Button id="b" text="B" />
		</LinearLayout>
	</LinearLayout>`

	root, err := inflater.InflateFromString(xmlStr)
	if err != nil {
		t.Fatal(err)
	}
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
	inner := root.Children()[0]
	if inner.Tag() != "LinearLayout" {
		t.Error("inner should be LinearLayout")
	}
	if len(inner.Children()) != 2 {
		t.Fatalf("inner should have 2 children, got %d", len(inner.Children()))
	}
	if inner.Children()[0].Id() != "a" {
		t.Error("first inner child id should be a")
	}
	if inner.Children()[1].Id() != "b" {
		t.Error("second inner child id should be b")
	}
}

func TestInflateFrameLayout(t *testing.T) {
	inflater := NewLayoutInflater(nil)
	RegisterBuiltinViews(inflater)

	xmlStr := `<FrameLayout width="match_parent" height="match_parent">
		<View id="bg" width="match_parent" height="match_parent" background="#FF0000" />
		<TextView id="label" width="wrap_content" height="wrap_content" text="Overlay" />
	</FrameLayout>`

	root, err := inflater.InflateFromString(xmlStr)
	if err != nil {
		t.Fatal(err)
	}
	if root.Tag() != "FrameLayout" {
		t.Error("root should be FrameLayout")
	}
	if len(root.Children()) != 2 {
		t.Fatalf("expected 2 children, got %d", len(root.Children()))
	}
}

func TestInflateFromLayout(t *testing.T) {
	f := fstest.MapFS{
		"layout/main.xml": &fstest.MapFile{Data: []byte(`<LinearLayout orientation="vertical">
			<TextView id="hello" text="Hello" />
		</LinearLayout>`)},
	}
	rm := NewResourceManager(f)
	inflater := NewLayoutInflater(rm)
	RegisterBuiltinViews(inflater)

	root := inflater.Inflate("@layout/main")
	if root == nil {
		t.Fatal("expected non-nil root")
	}
	if root.Tag() != "LinearLayout" {
		t.Error("root should be LinearLayout")
	}
	if len(root.Children()) != 1 {
		t.Fatalf("expected 1 child, got %d", len(root.Children()))
	}
	if root.Children()[0].Id() != "hello" {
		t.Error("child id should be hello")
	}
}

func TestInflateCommonAttrs(t *testing.T) {
	inflater := NewLayoutInflater(nil)
	RegisterBuiltinViews(inflater)

	xmlStr := `<View id="box" width="100dp" height="50dp" padding="8dp" margin="4dp" background="#FF0000" />`
	root, err := inflater.InflateFromString(xmlStr)
	if err != nil {
		t.Fatal(err)
	}
	if root.Id() != "box" {
		t.Error("id should be box")
	}
	style := root.GetStyle()
	if style == nil {
		t.Fatal("style should not be nil")
	}
	if style.Width.Value != 100 || style.Width.Unit != 1 { // DimensionDp = 1
		t.Errorf("width: %v", style.Width)
	}
	if style.Height.Value != 50 {
		t.Errorf("height: %v", style.Height)
	}
	padding := root.Padding()
	if padding.Left != 8 || padding.Top != 8 || padding.Right != 8 || padding.Bottom != 8 {
		t.Errorf("padding: %+v", padding)
	}
	margin := root.Margin()
	if margin.Left != 4 || margin.Top != 4 || margin.Right != 4 || margin.Bottom != 4 {
		t.Errorf("margin: %+v", margin)
	}
	if style.BackgroundColor.R != 0xFF || style.BackgroundColor.A != 0xFF {
		t.Errorf("background: %+v", style.BackgroundColor)
	}
}

func TestInflateImageView(t *testing.T) {
	inflater := NewLayoutInflater(nil)
	RegisterBuiltinViews(inflater)

	xmlStr := `<ImageView id="img" width="100dp" height="100dp" />`
	root, err := inflater.InflateFromString(xmlStr)
	if err != nil {
		t.Fatal(err)
	}
	if root.Tag() != "ImageView" {
		t.Error("should be ImageView")
	}
	if root.Id() != "img" {
		t.Error("id should be img")
	}
}
