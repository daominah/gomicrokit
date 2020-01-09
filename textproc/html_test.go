package textproc

import (
	"strings"
	"testing"

	"golang.org/x/net/html"
)

func TestHtmlUtils(t *testing.T) {
	htmlStr := `<html>

<head><title>A Simple HTML Document</title></head>

<body>
    <div id="main">
        The main div
        <p>This is a very simple HTML document</p>
        <p>It only has two paragraphs</p>
        <a href="https://www.w3schools.com">This is a link</a>
        <a href="https://www.google.com.vn#frag2" id="xxx2">Another link</a>
        <a href="https://www.google.com.vn" class="c1">Link3</a>
        <a href="https://www.google.com.vn#frag3" class="c2">Link4</a>
        <img src="https://www.google.com.vn/images/branding/googlelogo/2x/googlelogo_color_272x92dp.png">
        
        <div id="dev_page_content_wrap" class=" ">
            <h3>This is dev_page_content_wrap_h3</h3>
            <a href="https://www.google.com.vn" class="c1">Link5</a>
            <a href="https://www.google.com.vn" class="c2">Link6</a>
            <div class="blog_wide_image">
                <a href="/file/811140722/2/3cY8obIVMwk.102339/19827a3ef0493d9816" target="_blank">
                    <img src="/file/811140722/2/3cY8obIVMwk.102339/19827a3ef0493d9816" title="Themes for Android" />
                </a>
            </div>
        </div>
    </div>
    
    <p id="demo"></p>
    <script>
        document.getElementById("demo").innerHTML = "Hello JavaScript!";
    </script>
</body>

</html>`
	root, err := html.Parse(strings.NewReader(htmlStr))
	if err != nil {
		t.Fatal(err)
	}

	elems, err := HtmlXpath(root, "//a[@class='c1']")
	if err != nil || len(elems) != 2 {
		t.Fatal(err)
	}

	if HtmlGetText(elems[0]) != "Link3" ||
		HtmlGetText(elems[1]) != "Link5" {
		t.Error()
	}

	urls := HtmlGetHrefs("http://daominah.com/index", root)
	if len(urls) != 3 {
		t.Error("len(urls)", len(urls))
	}

	text := HtmlGetText(root)
	if strings.Index(text, "getElementById") != -1 {
		t.Error(text)
	}
}
