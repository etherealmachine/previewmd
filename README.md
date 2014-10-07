previewmd
=========
# What is it?
PreviewMD is a tiny command-line utility to preview markdown files on a local machine. You can use it instead of the web-based github editor. I like it because it fits into my development toolkit (Sublime+iTerm) and lets me organize my thoughts better to have a "clean" commit.

Once installed, you can point previewmd at a markdown file. It starts a web server and watches for changes to the file; Changes automatically refresh the page.

# Install
Installation requires the Go utilities (http://golang.org/doc/install)

`> go install github.com/etherealmachine/previewmd`
# Use
Point it at the file you want to preview.

	> previewmd -f <markdown file>
	Preview file at http://localhost:8080