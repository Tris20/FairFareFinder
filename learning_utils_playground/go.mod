module github.com/Tris20/FairFareFinder/learning_utils_playground

go 1.23.1

require (
	github.com/Tris20/FairFareFinder/src/backend/model v0.0.1
	// github.com/Tris20/FairFareFinder/config/handlers v0.0.1
	github.com/chromedp/chromedp v0.11.2
	github.com/mattn/go-sqlite3 v1.14.24
	github.com/rwcarlsen/goexif v0.0.0-20190401172101-9e8deecbddbd
	github.com/schollz/progressbar/v3 v3.17.1
	github.com/tdewolff/parse/v2 v2.7.19
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/chromedp/cdproto v0.0.0-20241022234722-4d5d5faf59fb // indirect
	github.com/chromedp/sysutil v1.1.0 // indirect
	github.com/gobwas/httphead v0.1.0 // indirect
	github.com/gobwas/pool v0.2.1 // indirect
	github.com/gobwas/ws v1.4.0 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	golang.org/x/sys v0.27.0 // indirect
	golang.org/x/term v0.26.0 // indirect
)

replace github.com/Tris20/FairFareFinder/src/backend/model => ../src/backend/model

// replace github.com/Tris20/FairFareFinder/config/handlers => ../config/handlers
