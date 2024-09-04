module github.com/Tris20/FairFareFinder

go 1.18

require (
	github.com/Tris20/FairFareFinder/src/go_files/timeutils v0.0.0-20240902195925-3b765470afe5
	github.com/mattn/go-sqlite3 v1.14.22
	github.com/schollz/progressbar/v3 v3.14.2
	gopkg.in/yaml.v2 v2.4.0
)

require (
	github.com/mitchellh/colorstring v0.0.0-20190213212951-d06e56a500db // indirect
	github.com/rivo/uniseg v0.4.7 // indirect
	golang.org/x/sys v0.17.0 // indirect
	golang.org/x/term v0.17.0 // indirect
)

replace github.com/Tris20/FairFareFinder/src/go_files/timeutils => ./src/go_files/timeutils
